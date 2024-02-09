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
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/valyala/fasthttp"
	"gorm.io/gorm"
)

type IWsService interface {
	AddClient(ctx *fasthttp.RequestCtx, userId int64, username string) error
	CreateRoom(senderId int64, receiverId int64) (int64, error)
	JoinRoom(ctx *fasthttp.RequestCtx, roomId int64, clientId int64, username string) error
	GetRooms() (*[]model.RoomRes, error)
	GetRoomClient(roomId int64) (*[]model.ClientRes, error)
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
			time.Sleep(3 * time.Second)
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
			time.Sleep(3 * time.Second)
			if err := rabbit.Consume(ctx, groupConfig, &wsSvc, HandleGroupChatMessage); err != nil {
				log.Printf("error consuming from queue %s: %s\n", rabbitmq.GroupChatQueue, err)
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
	Conn     *websocket.Conn
	Message  chan *model.ChannelMessage
	UserId   int64  `json:"userId"`
	Username string `json:"username"`
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
							Content:    "user left the chat",
							RoomId:     chd.ReceiveNewMessage.RoomId,
							ReceiverId: 0,
							State:      0,
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
			//todo : return error
			fmt.Println(err.Error())
		} else {
			m := model.CreateReturnChatMessagesAction(chatMessages)
			if sender, ok := wsSvc.hub.Clients[channelMessage.ChatMessagesReq.UserId]; ok {
				sender.Message <- m
			}
		}
	case model.SingleChatsListAction:
		chatMessages, err := wsSvc.GetSingleChatList(channelMessage.ChatsListReq)
		if err != nil {
			//todo : return error
			fmt.Println(err.Error())
		} else {
			m := model.CreateReturnChatListAction(chatMessages)
			if sender, ok := wsSvc.hub.Clients[channelMessage.ChatsListReq.UserId]; ok {
				sender.Message <- m
			}
		}
		//case model.MessageReadAction:  //todo :
	}

	if err = d.Ack(false); err != nil {
		log.Printf("error acking message: %s\n", err)
	}
}

func HandleSingleChatMessage(receiveNewMessage *model.ReceiveNewMessage, wsSvc *WsService) error {
	sender, ok := wsSvc.hub.Clients[receiveNewMessage.UserId]
	err := wsSvc.wsRepo.SaveMessage(receiveNewMessage)
	if err != nil {
		if errors.Is(err, gorm.ErrForeignKeyViolated) {
			// receiver user not found
			if ok {
				messageSendResult := model.CreateNewMessageSendResult(receiveNewMessage.RoomId, receiveNewMessage.ReceiverId, -1, 404, "Receiver User Not Found")
				sender.Message <- messageSendResult
			} else {
				// maybe save error
			}
		} else {
			if ok {
				messageSendResult := model.CreateNewMessageSendResult(receiveNewMessage.RoomId, receiveNewMessage.ReceiverId, -1, 500, err.Error())
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
			receiveMessage := model.CreateReceiveNewMessageAction(receiveNewMessage)
			cl.Message <- receiveMessage
			messageSendResult := model.CreateNewMessageSendResult(receiveNewMessage.RoomId, receiveNewMessage.ReceiverId, receiveNewMessage.State, 200, "")
			sender.Message <- messageSendResult
		}
		err = wsSvc.wsRepo.UpdateUserReceivedMessageTime(receiveNewMessage.ReceiverId)
	}

	return err
}

func (c *Client) WriteMessage() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Message:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			err := c.Conn.WriteJSON(message)
			if err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) ReadMessage(hub *Hub, rabbit rabbitmq.RabbitMQ) {
	defer func() {
		//hub.UnRegister <- c  //it just offline, didnt left
		c.Conn.Close()
		delete(hub.Clients, c.UserId)
		for _, room := range hub.Rooms {
			delete(room.Clients, c.UserId)
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
				Content:    clientMessage.NewMessage.Content,
				RoomId:     clientMessage.NewMessage.RoomId,
				ReceiverId: clientMessage.NewMessage.ReceiverId,
				State:      1,
				UserId:     c.UserId,
				Username:   c.Username,
			}
			receiveMessage := model.CreateReceiveNewMessageAction(message)

			if clientMessage.NewMessage.RoomId == -1 {
				//one to one message
				rabbit.Publish(ctx, receiveMessage, conf, c.UserId)
			} else {
				//group/channel message
				hub.Broadcast <- receiveMessage
			}
		case model.SingleChatMessagesAction:
			message := model.CreateGetChatMessagesAction(&clientMessage.ChatMessagesReq)
			rabbit.Publish(ctx, message, conf, c.UserId)
		case model.SingleChatsListAction:
			message := model.CreateGetChatListAction(&clientMessage.ChatsListReq)
			rabbit.Publish(ctx, message, conf, c.UserId)
			//case model.MessageReadAction: //todo :
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

func (w *WsService) AddClient(ctx *fasthttp.RequestCtx, userId int64, username string) error {
	err := upgrader.Upgrade(ctx, func(conn *websocket.Conn) {
		cl := &Client{
			Conn:     conn,
			Message:  make(chan *model.ChannelMessage, 10),
			UserId:   userId,
			Username: username,
		}

		w.hub.Clients[userId] = cl

		//todo : get messages through websocket or rest api call?
		//todo : handle unread messages

		go cl.WriteMessage()
		cl.ReadMessage(w.hub, w.rabbitmq)
	})

	return err
}

func (w *WsService) CreateRoom(senderId int64, receiverId int64) (int64, error) {
	roomId, err := w.wsRepo.CreateRoom(senderId, receiverId)
	if err != nil {
		return 0, err
	}

	room := &Room{
		ID:      roomId,
		Clients: make(map[int64]*Client),
	}
	room.Clients[senderId] = w.hub.Clients[senderId]
	if cl, ok := w.hub.Clients[receiverId]; ok {
		room.Clients[receiverId] = cl
	}
	w.hub.Rooms[roomId] = room

	return roomId, nil
}

func (w *WsService) JoinRoom(ctx *fasthttp.RequestCtx, roomId int64, clientId int64, username string) error {
	//this func if for group/channel which is not going to implement in this time.
	if _, ok := w.hub.Rooms[roomId]; !ok {
		return errors.New("not found")
	}
	err := upgrader.Upgrade(ctx, func(conn *websocket.Conn) {
		cl := &Client{
			Conn:     conn,
			Message:  make(chan *model.ChannelMessage, 10),
			UserId:   clientId,
			Username: username,
		}

		message := &model.ReceiveNewMessage{
			Content:    "A new user has joined the room",
			RoomId:     roomId,
			ReceiverId: 0,
			State:      0,
			UserId:     clientId,
			Username:   username,
		}
		receiveMessage := model.CreateReceiveNewMessageAction(message)

		// Register a new client through the register channel
		w.hub.Register <- receiveMessage
		// Broadcast that message
		w.hub.Broadcast <- receiveMessage

		go cl.WriteMessage()
		cl.ReadMessage(w.hub, w.rabbitmq)
	})

	return err
}

func (w *WsService) GetRooms() (*[]model.RoomRes, error) {
	rooms := make([]model.RoomRes, 0)

	for _, r := range w.hub.Rooms {
		rooms = append(rooms, model.RoomRes{ID: r.ID})
	}

	return &rooms, nil
}

func (w *WsService) GetRoomClient(roomId int64) (*[]model.ClientRes, error) {
	var clients []model.ClientRes
	if _, ok := w.hub.Rooms[roomId]; !ok {
		clients = make([]model.ClientRes, 0)
		return &clients, nil
	}

	for _, cl := range w.hub.Rooms[roomId].Clients {
		clients = append(clients, model.ClientRes{ID: cl.UserId, Username: cl.Username})
	}

	return &clients, nil
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
