package service

import (
	"downloader_gochat/model"
	errorHandler "downloader_gochat/pkg/error"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type ITelegramMessageService interface {
	TelegramMessageQueueHandler()
	AddTelegramMessageToQueue(botToken string, chatId string, message string)
	StopTelegramMessageQueue()
	RunTelegramMessageQueue()
}

type TelegramMessageService struct {
	queue            []QueueItem
	queueMux         *sync.Mutex
	dispatchInterval time.Duration
	size             int
	perChatLimit     int
	wg               sync.WaitGroup
	taskInfo         *model.TaskInfo
}

type QueueItem struct {
	botToken string
	chatId   string
	messages []string
}

const (
	telegramMessageConsumerCount = 10
)

func NewTelegramMessageService() *TelegramMessageService {
	svc := &TelegramMessageService{
		queue:            make([]QueueItem, 0),
		queueMux:         &sync.Mutex{},
		size:             2000,
		perChatLimit:     10,
		dispatchInterval: 3 * time.Second,
		wg:               sync.WaitGroup{},
	}

	for i := 0; i < telegramMessageConsumerCount; i++ {
		svc.RunTelegramMessageQueue()
	}

	taskInfo := &model.TaskInfo{
		ConsumerCount: telegramMessageConsumerCount,
		Links:         make([]string, 0),
		Mux:           &sync.Mutex{},
	}
	svc.taskInfo = taskInfo
	AdminSvc.status.Tasks.TelegramMessage = taskInfo

	return svc
}

//------------------------------------------
//------------------------------------------

func (t *TelegramMessageService) TelegramMessageQueueHandler() {
	defer t.wg.Done()

	defer func() {
		if r := recover(); r != nil {
			// Convert the panic to an error
			fmt.Printf("recovered from panic: %v\n", r)
			t.RunTelegramMessageQueue()
		}
	}()

	for {
		time.Sleep(t.dispatchInterval)
		t.queueMux.Lock()
		if len(t.queue) == 0 {
			t.queueMux.Unlock()
			time.Sleep(t.dispatchInterval)
			continue
		}

		queueItem := t.queue[0]
		t.queue = t.queue[1:]
		t.taskInfo.RunningCount = len(t.queue)
		t.queueMux.Unlock()

		findMessage := url.PathEscape(strings.ReplaceAll("---------------------- UPDATES/NEWS ----------------------", "-", "\\-"))
		lineSeparator := url.PathEscape("\n\n")
		dot := url.PathEscape("\\. ")
		for i, m := range queueItem.messages {
			findMessage += lineSeparator + strconv.Itoa(i+1) + dot + m
		}

		apiUrl := fmt.Sprintf("https://api.telegram.org/bot%v/sendMessage?chat_id=%v&text=%v&parse_mode=MarkdownV2",
			queueItem.botToken, queueItem.chatId, findMessage)

		needReQueue := false
		resp, err := http.Get(apiUrl)

		if errors.Is(err, os.ErrDeadlineExceeded) || os.IsTimeout(err) {
			errorMessage := "Error on calling telegram api: timeout"
			errorHandler.SaveError(errorMessage, err)
			needReQueue = true
		}
		if err != nil {
			errorMessage := fmt.Sprintf("Error on calling telegram api: %s", err)
			errorHandler.SaveError(errorMessage, err)
		}
		if resp.StatusCode == http.StatusGatewayTimeout || resp.StatusCode == http.StatusRequestTimeout {
			errorMessage := "Error on calling telegram api: timeout"
			errorHandler.SaveError(errorMessage, err)
			needReQueue = true
		}
		if resp.StatusCode != http.StatusOK {
			if resp.Status == "400 Bad Request" {
				errorMessage := fmt.Sprintf("Error on calling telegram api: Bad-Request : %v", findMessage)
				errorHandler.SaveError(errorMessage, err)
			} else {
				errorMessage := fmt.Sprintf("Error on calling telegram api: %v", fmt.Errorf("bad status: %s", resp.Status))
				errorHandler.SaveError(errorMessage, err)
			}
		}

		if needReQueue {
			t.queueMux.Lock()
			t.queue = append(t.queue, queueItem)
			t.queueMux.Unlock()
		}
	}
}

func (t *TelegramMessageService) AddTelegramMessageToQueue(botToken string, chatId string, message string) {
	t.queueMux.Lock()
	defer t.queueMux.Unlock()

	added := false
	for i := range t.queue {
		if t.queue[i].botToken == botToken && t.queue[i].chatId == chatId {
			if len(t.queue[i].messages) < t.perChatLimit {
				added = true
				t.queue[i].messages = append(t.queue[i].messages, message)
				break
			}
		}
	}

	if !added {
		newItem := QueueItem{
			botToken: botToken,
			chatId:   chatId,
			messages: []string{message},
		}
		t.queue = append(t.queue, newItem)
	}
}

func (t *TelegramMessageService) StopTelegramMessageQueue() {
	t.wg.Wait()
}

func (t *TelegramMessageService) RunTelegramMessageQueue() {
	t.wg.Add(1)
	go t.TelegramMessageQueueHandler()
}

//------------------------------------------
//------------------------------------------
