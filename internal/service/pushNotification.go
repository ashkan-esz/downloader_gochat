package service

import (
	"context"
	"downloader_gochat/configs"
	errorHandler "downloader_gochat/pkg/error"
	"encoding/base64"
	"fmt"
	"sync"
	"time"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"github.com/gofiber/fiber/v2/log"
	"google.golang.org/api/option"
)

type IPushNotificationService interface {
	PushNotificationBufferHandler()
	SendBufferedPushNotifications(messages []*messaging.Message)
	AddPushNotificationToBuffer(deviceToken string, title string, body string, imageUrl string, collapseKey string)
	StopPushNotificationBuffer()
	RunPushNotificationBuffer()
	sendPushNotification(deviceToken string, title string, body string, imageUrl string) (string, error)
	sendPushNotificationToMultiDevice(deviceTokens []string, title string, body string, imageUrl string) (*messaging.BatchResponse, error)
}

type PushNotificationService struct {
	FcmClient        *messaging.Client
	dispatchInterval time.Duration
	batchCh          chan *messaging.Message
	wg               sync.WaitGroup
}

func NewPushNotificationService() *PushNotificationService {
	decodedKey, err := getDecodedFireBaseKey()
	if err != nil {
		log.Fatalf("Error in initializing firebase app: %s", err)
		return nil
	}

	opts := []option.ClientOption{option.WithCredentialsJSON(decodedKey)}

	// Initialize firebase app
	app, err := firebase.NewApp(context.Background(), nil, opts...)
	if err != nil {
		log.Fatalf("Error in initializing firebase app: %s", err)
		return nil
	}

	fcmClient, err := app.Messaging(context.Background())
	if err != nil {
		log.Fatalf("Error in initializing firebase cloud messaging app: %s", err)
		return nil
	}

	svc := &PushNotificationService{
		FcmClient:        fcmClient,
		batchCh:          make(chan *messaging.Message, 500),
		dispatchInterval: 500 * time.Millisecond,
		wg:               sync.WaitGroup{},
	}
	svc.RunPushNotificationBuffer()
	return svc
}

func getDecodedFireBaseKey() ([]byte, error) {
	fireBaseAuthKey := configs.GetConfigs().FirebaseAuthKey
	decodedKey, err := base64.StdEncoding.DecodeString(fireBaseAuthKey)

	if err != nil {
		return nil, err
	}
	return decodedKey, nil
}

//------------------------------------------
//------------------------------------------

func (p *PushNotificationService) PushNotificationBufferHandler() {
	defer p.wg.Done()

	// set your interval
	t := time.NewTicker(p.dispatchInterval)

	// we can send up to 500 messages per call to Firebase
	messages := make([]*messaging.Message, 0, 500)

	defer func() {
		t.Stop()
		// send all buffered messages before quit
		p.SendBufferedPushNotifications(messages)
		//fmt.Println("notification batch sender finished")
	}()

	for {
		select {
		case m, ok := <-p.batchCh:
			if !ok {
				return
			}

			messages = append(messages, m)
		case <-t.C:
			p.SendBufferedPushNotifications(messages)
			messages = messages[:0]
		}
	}
}

func (p *PushNotificationService) SendBufferedPushNotifications(messages []*messaging.Message) {
	if len(messages) == 0 {
		return
	}

	batchResp, err := p.FcmClient.SendEach(context.TODO(), messages)
	if err != nil {
		errorMessage := fmt.Sprintf("firebase batch response: %+v, err: %s", batchResp, err)
		errorHandler.SaveError(errorMessage, err)
	}
}

func (p *PushNotificationService) AddPushNotificationToBuffer(deviceToken string, title string, body string, imageUrl string, collapseKey string) {
	message := &messaging.Message{
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
			//ImageURL: imageUrl,
		},
		Token: deviceToken,
		Android: &messaging.AndroidConfig{
			CollapseKey: collapseKey,
			Notification: &messaging.AndroidNotification{
				Title:    title,
				Body:     body,
				ImageURL: imageUrl,
			},
		},
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					MutableContent: true,
				},
			},
			FCMOptions: &messaging.APNSFCMOptions{
				ImageURL: imageUrl,
			},
		},
		Webpush: &messaging.WebpushConfig{
			Headers: map[string]string{
				"image": imageUrl,
			},
		},
	}
	p.batchCh <- message
}

func (p *PushNotificationService) StopPushNotificationBuffer() {
	close(p.batchCh)
	p.wg.Wait()
}

func (p *PushNotificationService) RunPushNotificationBuffer() {
	p.wg.Add(1)
	go p.PushNotificationBufferHandler()
}

//------------------------------------------
//------------------------------------------

func (p *PushNotificationService) sendPushNotification(deviceToken string, title string, body string, imageUrl string) (string, error) {
	response, err := p.FcmClient.Send(context.Background(), &messaging.Message{
		Notification: &messaging.Notification{
			Title:    title,
			Body:     body,
			ImageURL: imageUrl,
		},
		Token: deviceToken,
	})
	return response, err
}

func (p *PushNotificationService) sendPushNotificationToMultiDevice(deviceTokens []string, title string, body string, imageUrl string) (*messaging.BatchResponse, error) {
	response, err := p.FcmClient.SendEachForMulticast(context.Background(), &messaging.MulticastMessage{
		Notification: &messaging.Notification{
			Title:    title,
			Body:     body,
			ImageURL: imageUrl,
		},
		Tokens: deviceTokens,
	})

	return response, err
}
