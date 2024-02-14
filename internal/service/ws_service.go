package service

import (
	"context"
	"downloader_gochat/internal/repository"
	"downloader_gochat/model"
	"downloader_gochat/rabbitmq"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"slices"
	"time"

	"github.com/fasthttp/websocket"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/valyala/fasthttp"
	"gorm.io/gorm"
)

type IWsService interface {
	AddClient(ctx *fasthttp.RequestCtx, userId int64, username string, deviceId string) error
	GetSingleChatMessages(params *model.GetSingleMessagesReq) (*[]model.MessageDataModel, error)
	GetSingleChatList(params *model.GetSingleChatListReq) (*[]model.ChatsCompressedDataModel, error)
}

type WsService struct {
	wsRepo   repository.IWsRepository
	rabbitmq rabbitmq.RabbitMQ
	timeout  time.Duration
	hub      *Hub
}

func NewWsService(WsRepo repository.IWsRepository, rabbit rabbitmq.RabbitMQ) *WsService {
	wsSvc := WsService{
		wsRepo:   WsRepo,
		rabbitmq: rabbit,
		timeout:  time.Duration(2) * time.Second,
		hub:      NewHub(),
	}

	config := rabbitmq.NewConfigConsume(rabbitmq.SingleChatQueue, "")
	for i := 0; i < 10; i++ {
		ctx, _ := context.WithCancel(context.Background())
		go func() {
			openConChan := make(chan struct{})
			rabbitmq.NotifySetupDone(openConChan)
			<-openConChan
			if err := rabbit.Consume(ctx, config, &wsSvc, UserMessageConsumer); err != nil {
				log.Printf("error consuming from queue %s: %s\n", rabbitmq.SingleChatQueue, err)
			}
		}()
	}

	for i := 0; i < 1; i++ {
		go wsSvc.hub.RunGroupHandler()
	}

	groupConfig := rabbitmq.NewConfigConsume(rabbitmq.GroupChatQueue, "")
	for i := 0; i < 1; i++ {
		ctx, _ := context.WithCancel(context.Background())
		go func() {
			openConChan := make(chan struct{})
			rabbitmq.NotifySetupDone(openConChan)
			<-openConChan
			if err := rabbit.Consume(ctx, groupConfig, &wsSvc, HandleGroupChatMessage); err != nil {
				log.Printf("error consuming from queue %s: %s\n", rabbitmq.GroupChatQueue, err)
			}
		}()
	}

	messageStateConfig := rabbitmq.NewConfigConsume(rabbitmq.MessageStateQueue, "")
	for i := 0; i < 3; i++ {
		ctx, _ := context.WithCancel(context.Background())
		go func() {
			openConChan := make(chan struct{})
			rabbitmq.NotifySetupDone(openConChan)
			<-openConChan
			if err := rabbit.Consume(ctx, messageStateConfig, &wsSvc, MessageStateConsumer); err != nil {
				log.Printf("error consuming from queue %s: %s\n", rabbitmq.MessageStateQueue, err)
			}
		}()
	}

	return &wsSvc
}

type Hub struct {
	//todo : handle concurrency sharing maps
	Clients    map[int64]*Client
	Rooms      map[int64]*Room
	Register   chan *model.ChannelMessage
	UnRegister chan *model.ChannelMessage
	Broadcast  chan *model.ChannelMessage
}

type Room struct {
	ID      int64             `json:"id"`
	Clients map[int64]*Client `json:"clients"`
}

type Client struct {
	Connections []*ClientConnection
	Message     chan *model.ChannelMessage
	UserId      int64  `json:"userId"`
	Username    string `json:"username"`
}

type ClientConnection struct {
	Conn     *websocket.Conn
	DeviceId string
}

//todo : need to limit parallel db operations

func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[int64]*Client, avgClients),
		Rooms:      make(map[int64]*Room),
		Register:   make(chan *model.ChannelMessage),
		UnRegister: make(chan *model.ChannelMessage),
		Broadcast:  make(chan *model.ChannelMessage, 5),
	}
}

//------------------------------------------
//------------------------------------------

func (h *Hub) RunGroupHandler() {
	//todo : use rabbitmq

	for {
		select {
		case chd := <-h.Register:
			if room, ok := h.Rooms[chd.ReceiveNewMessage.RoomId]; ok {
				if _, ok := room.Clients[chd.ReceiveNewMessage.UserId]; !ok {
					client, _ := h.Clients[chd.ReceiveNewMessage.UserId]
					room.Clients[chd.ReceiveNewMessage.UserId] = client
				}
			}
		case chd := <-h.UnRegister:
			if room, ok := h.Rooms[chd.ReceiveNewMessage.RoomId]; ok {
				if _, ok := room.Clients[chd.ReceiveNewMessage.UserId]; ok {
					// Broadcast a message saying that the client left the room
					if len(room.Clients) != 0 {
						message := &model.ReceiveNewMessage{
							Id:         0,
							Uuid:       uuid.NewString(),
							Content:    "user left the chat",
							RoomId:     chd.ReceiveNewMessage.RoomId,
							ReceiverId: 0,
							State:      0,
							Date:       time.Now(),
							UserId:     chd.ReceiveNewMessage.UserId,
							Username:   chd.ReceiveNewMessage.Username,
						}
						receiveMessage := model.CreateReceiveNewMessageAction(message)
						h.Broadcast <- receiveMessage
					}

					delete(room.Clients, chd.ReceiveNewMessage.UserId)
					if client, ok := h.Clients[chd.ReceiveNewMessage.UserId]; ok {
						close(client.Message)
					}
				}
			}
		case m := <-h.Broadcast:
			if room, ok := h.Rooms[m.ReceiveNewMessage.RoomId]; ok {
				for _, cl := range room.Clients {
					receiveMessage := model.CreateReceiveNewMessageAction(m.ReceiveNewMessage)
					cl.Message <- receiveMessage
					//conf := rabbitmq.NewConfigPublish(rabbitmq.ChatExchange, rabbitmq.GroupChatBindingKey)
					//rabbit.Publish(ctx, msg, conf)
				}
			}
		}
	}
}

func HandleGroupChatMessage(d *amqp.Delivery, extraConsumerData interface{}) {
	//placeholder function for handling messaging in groups

	////run as rabbitmq consumer
	//wsSvc := extraConsumerData.(*WsService)
	//var m *model.ChannelMessage
	//err := json.Unmarshal(d.Body, &m)
	//if err != nil {
	//	return
	//}
	//
	//if err := d.Ack(false); err != nil {
	//	log.Printf("error acking message: %s\n", err)
	//}
}

func UserMessageConsumer(d *amqp.Delivery, extraConsumerData interface{}) {
	// run as rabbitmq consumer
	wsSvc := extraConsumerData.(*WsService)
	var channelMessage *model.ChannelMessage
	err := json.Unmarshal(d.Body, &channelMessage)
	if err != nil {
		return
	}

	switch channelMessage.Action {
	case model.ReceiveNewMessageAction:
		err = HandleSingleChatMessage(channelMessage.ReceiveNewMessage, wsSvc)
	case model.SingleChatMessagesAction:
		chatMessages, err := wsSvc.wsRepo.GetSingleChatMessages(channelMessage.ChatMessagesReq)
		if err != nil {
			if err = d.Nack(false, true); err != nil {
				log.Printf("error nacking message: %s\n", err)
			}
			//if sender, ok := wsSvc.hub.Clients[channelMessage.ChatMessagesReq.UserId]; ok {
			//	errorData := model.CreateActionError(500, err.Error(), model.SingleChatMessagesAction, channelMessage.ChatMessagesReq)
			//	sender.Message <- errorData
			//}
		} else {
			m := model.CreateReturnChatMessagesAction(chatMessages)
			if sender, ok := wsSvc.hub.Clients[channelMessage.ChatMessagesReq.UserId]; ok {
				sender.Message <- m
			}
		}
	case model.SingleChatsListAction:
		if _, ok := wsSvc.hub.Clients[channelMessage.ChatsListReq.UserId]; ok {
			chatMessages, err := wsSvc.GetSingleChatList(channelMessage.ChatsListReq)
			if err != nil {
				if err = d.Nack(false, true); err != nil {
					log.Printf("error nacking message: %s\n", err)
				}
				//if sender, ok := wsSvc.hub.Clients[channelMessage.ChatsListReq.UserId]; ok {
				//	errorData := model.CreateActionError(500, err.Error(), model.SingleChatsListAction, channelMessage.ChatsListReq)
				//	sender.Message <- errorData
				//}
			} else {
				m := model.CreateReturnChatListAction(chatMessages)
				if sender, ok := wsSvc.hub.Clients[channelMessage.ChatsListReq.UserId]; ok {
					sender.Message <- m
				}
			}
		}
	}

	if err = d.Ack(false); err != nil {
		log.Printf("error acking message: %s\n", err)
	}
}

func MessageStateConsumer(d *amqp.Delivery, extraConsumerData interface{}) {
	// run as rabbitmq consumer
	wsSvc := extraConsumerData.(*WsService)
	var channelMessage *model.ChannelMessage
	err := json.Unmarshal(d.Body, &channelMessage)
	if err != nil {
		return
	}

	switch channelMessage.Action {
	case model.MessageReadAction:
		//todo : validation on provided ReceiverId
		err := wsSvc.wsRepo.BatchUpdateMessageState(
			channelMessage.MessageRead.Id,
			channelMessage.MessageRead.RoomId,
			channelMessage.MessageRead.UserId,
			channelMessage.MessageRead.ReceiverId,
			channelMessage.MessageRead.State)
		if err != nil {
			if messageReceiver, ok := wsSvc.hub.Clients[channelMessage.MessageRead.ReceiverId]; ok {
				if err.Error() == "notfound" {
					errorData := model.CreateActionError(404, "message not found", model.MessageReadAction, channelMessage.MessageRead)
					messageReceiver.Message <- errorData
				} else {
					if err = d.Nack(false, true); err != nil {
						log.Printf("error nacking message: %s\n", err)
					}
					//errorData := model.CreateActionError(500, err.Error(), model.MessageReadAction, channelMessage.MessageRead)
					//messageReceiver.Message <- errorData
				}
			}
		} else {
			if messageCreator, ok := wsSvc.hub.Clients[channelMessage.MessageRead.UserId]; ok {
				message := model.CreateMessageReadAction(
					channelMessage.MessageRead.Id,
					channelMessage.MessageRead.RoomId,
					channelMessage.MessageRead.UserId,
					channelMessage.MessageRead.ReceiverId,
					channelMessage.MessageRead.Date,
					channelMessage.MessageRead.State, true)
				messageCreator.Message <- message
			}
		}
		//case model.ReceiveMessageStateAction:
	}

	if err = d.Ack(false); err != nil {
		log.Printf("error acking [messageState] message: %s\n", err)
	}
}

func HandleSingleChatMessage(receiveNewMessage *model.ReceiveNewMessage, wsSvc *WsService) error {
	sender, ok := wsSvc.hub.Clients[receiveNewMessage.UserId]
	mid, err := wsSvc.wsRepo.SaveMessage(receiveNewMessage)
	if err != nil {
		if errors.Is(err, gorm.ErrForeignKeyViolated) {
			// receiver user not found
			if ok {
				messageSendResult := model.CreateNewMessageSendResult(
					-1,
					receiveNewMessage.Uuid,
					receiveNewMessage.RoomId,
					receiveNewMessage.ReceiverId,
					receiveNewMessage.Date,
					-1, 404, "Receiver User Not Found")
				sender.Message <- messageSendResult
			} else {
				// maybe save error
			}
		} else {
			if ok {
				messageSendResult := model.CreateNewMessageSendResult(
					-1,
					receiveNewMessage.Uuid,
					receiveNewMessage.RoomId,
					receiveNewMessage.ReceiverId,
					receiveNewMessage.Date,
					-1, 500, err.Error())
				sender.Message <- messageSendResult
				// maybe save error
			} else {
				// maybe save error
			}
		}
	} else {
		cl, ok := wsSvc.hub.Clients[receiveNewMessage.ReceiverId]
		if ok {
			//receiver is online
			receiveNewMessage.Id = mid
			receiveMessage := model.CreateReceiveNewMessageAction(receiveNewMessage)
			cl.Message <- receiveMessage
			messageSendResult := model.CreateNewMessageSendResult(
				mid,
				receiveNewMessage.Uuid,
				receiveNewMessage.RoomId,
				receiveNewMessage.ReceiverId,
				receiveNewMessage.Date,
				receiveNewMessage.State,
				200, "")
			sender.Message <- messageSendResult
		}
		err = wsSvc.wsRepo.UpdateUserReceivedMessageTime(receiveNewMessage.ReceiverId)
	}

	return err
}

func (c *ClientConnection) WriteMessage(cc *Client) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
		cc.Connections = slices.DeleteFunc(cc.Connections, func(item *ClientConnection) bool {
			return item.DeviceId == c.DeviceId
		})
	}()

	for {
		select {
		case message, ok := <-cc.Message:
			if !ok {
				c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
				// The hub closed the channel.
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			errorFlag := false
			for i := range cc.Connections {
				cc.Connections[i].Conn.SetWriteDeadline(time.Now().Add(writeWait))
				err := cc.Connections[i].Conn.WriteJSON(message)
				if err != nil {
					fmt.Printf("error on sending json to client: %v\n", err)
					//return
					errorFlag = true
				}
			}
			if errorFlag {
				return
			}
		case <-ticker.C:
			errorFlag := false
			for i := range cc.Connections {
				cc.Connections[i].Conn.SetWriteDeadline(time.Now().Add(writeWait))
				if err := cc.Connections[i].Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}
			}
			if errorFlag {
				return
			}
		}
	}
}

func (c *ClientConnection) ReadMessage(cc *Client, hub *Hub, rabbit rabbitmq.RabbitMQ) {
	defer func() {
		//hub.UnRegister <- c  //it just offline, didnt left
		c.Conn.Close()
		cc.Connections = slices.DeleteFunc(cc.Connections, func(item *ClientConnection) bool {
			return item.DeviceId == c.DeviceId
		})
		if len(cc.Connections) == 0 {
			delete(hub.Clients, cc.UserId)
			for _, room := range hub.Rooms {
				delete(room.Clients, cc.UserId)
			}
		}
	}()
	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for {
		var clientMessage model.ClientMessage
		err := c.Conn.ReadJSON(&clientMessage)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		conf := rabbitmq.NewConfigPublish(rabbitmq.ChatExchange, rabbitmq.SingleChatBindingKey)
		switch clientMessage.Action {
		case model.SendNewMessageAction:
			message := &model.ReceiveNewMessage{
				Id:         0,
				Uuid:       clientMessage.NewMessage.Uuid,
				Content:    clientMessage.NewMessage.Content,
				RoomId:     clientMessage.NewMessage.RoomId,
				ReceiverId: clientMessage.NewMessage.ReceiverId,
				Date:       time.Now(),
				State:      1,
				UserId:     cc.UserId,
				Username:   cc.Username,
			}
			receiveMessage := model.CreateReceiveNewMessageAction(message)

			if clientMessage.NewMessage.RoomId == -1 {
				//one to one message
				rabbit.Publish(ctx, receiveMessage, conf, cc.UserId)
			} else {
				//group/channel message
				hub.Broadcast <- receiveMessage
			}
		case model.SingleChatMessagesAction:
			message := model.CreateGetChatMessagesAction(&clientMessage.ChatMessagesReq)
			rabbit.Publish(ctx, message, conf, cc.UserId)
		case model.SingleChatsListAction:
			message := model.CreateGetChatListAction(&clientMessage.ChatsListReq)
			rabbit.Publish(ctx, message, conf, cc.UserId)
		case model.MessageReadAction:
			readQueueConf := rabbitmq.NewConfigPublish(rabbitmq.MessageStateExchange, rabbitmq.MessageStateBindingKey)
			message := model.CreateMessageReadAction(
				clientMessage.MessageRead.Id,
				clientMessage.MessageRead.RoomId,
				clientMessage.MessageRead.UserId,
				clientMessage.MessageRead.ReceiverId,
				clientMessage.MessageRead.Date,
				2, false)
			rabbit.Publish(ctx, message, readQueueConf, cc.UserId)
		}

	}
}

func (h *Hub) ReviveWebsocket(wsSvc *WsService) {
	//todo :
	if err := recover(); err != nil {
		if os.Getenv("LOG_PANIC_TRACE") == "true" {
			log.Println(
				"level:", "error",
				"err: ", err,
				"trace", string(debug.Stack()),
			)
		} else {
			log.Println(
				"level", "error",
				"err", err,
			)
		}
	}
}

//------------------------------------------
//------------------------------------------

func (w *WsService) AddClient(ctx *fasthttp.RequestCtx, userId int64, username string, deviceId string) error {
	err := upgrader.Upgrade(ctx, func(conn *websocket.Conn) {
		connection := &ClientConnection{
			Conn:     conn,
			DeviceId: deviceId,
		}

		client, ok := w.hub.Clients[userId]
		if ok {
			client.Connections = append(client.Connections, connection)
		} else {
			cl := &Client{
				Connections: []*ClientConnection{connection},
				Message:     make(chan *model.ChannelMessage, 10),
				UserId:      userId,
				Username:    username,
			}

			w.hub.Clients[userId] = cl
			client = cl

			go connection.WriteMessage(cl)
		}

		chatsListReq := &model.GetSingleChatListReq{
			UserId:               userId,
			MessagePerChatLimit:  3,
			ChatsLimit:           20,
			IncludeProfileImages: true,
		}
		message := model.CreateGetChatListAction(chatsListReq)
		conf := rabbitmq.NewConfigPublish(rabbitmq.ChatExchange, rabbitmq.SingleChatBindingKey)
		w.rabbitmq.Publish(ctx, message, conf, userId)

		connection.ReadMessage(client, w.hub, w.rabbitmq)
	})

	return err
}

func (w *WsService) GetSingleChatMessages(params *model.GetSingleMessagesReq) (*[]model.MessageDataModel, error) {
	messages, err := w.wsRepo.GetSingleChatMessages(params)
	return messages, err
}

func (w *WsService) GetSingleChatList(params *model.GetSingleChatListReq) (*[]model.ChatsCompressedDataModel, error) {
	readTime := time.Now().UTC()
	chats, profileImages, err := w.wsRepo.GetSingleChatList(params)

	if err != nil {
		return nil, err
	}

	if params.MessageState != 2 {
		err = w.wsRepo.UpdateUserReadMessageTime(params.UserId, readTime)
		if err != nil {
			return nil, err
		}
	}

	compressedChats := make([]model.ChatsCompressedDataModel, 0)
	for _, chat := range chats {
		m := model.MessageDataModel{
			Id:         chat.Id,
			ReceiverId: chat.ReceiverId,
			CreatorId:  chat.CreatorId,
			RoomId:     chat.RoomId,
			Date:       chat.Date,
			State:      chat.State,
			Content:    chat.Content,
		}
		exist := false
		for i := range compressedChats {
			if chat.UserId == compressedChats[i].UserId {
				exist = true
				compressedChats[i].Messages = append(compressedChats[i].Messages, m)
				break
			}
		}
		if !exist {
			cChat := model.ChatsCompressedDataModel{
				UserId:        chat.UserId,
				Username:      chat.Username,
				PublicName:    chat.PublicName,
				Role:          chat.Role,
				ProfileImages: filterProfileImages(profileImages, chat.UserId),
				Messages:      []model.MessageDataModel{m},
			}
			compressedChats = append(compressedChats, cChat)
		}
	}

	for i := range compressedChats {
		slices.SortFunc(compressedChats[i].Messages, func(a, b model.MessageDataModel) int {
			return b.Date.Compare(a.Date)
		})
	}
	slices.SortFunc(compressedChats, func(a, b model.ChatsCompressedDataModel) int {
		return b.Messages[0].Date.Compare(a.Messages[0].Date)
	})

	return &compressedChats, err
}

func filterProfileImages(profileImages []model.ProfileImageDataModel, userId int64) []model.ProfileImageDataModel {
	var images = make([]model.ProfileImageDataModel, 0)
	for i := range profileImages {
		if profileImages[i].UserId == userId {
			images = append(images, profileImages[i])
		}
	}
	slices.SortFunc(images, func(a, b model.ProfileImageDataModel) int {
		return b.AddDate.Compare(a.AddDate)
	})
	return images
}
