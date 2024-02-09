package service

import (
	"context"
	"downloader_gochat/internal/repository"
	"downloader_gochat/model"
	"downloader_gochat/rabbitmq"
	"encoding/json"
	"errors"
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
			if err := rabbit.Consume(ctx, config, &wsSvc, HandleSingleChatMessage); err != nil {
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
	Register   chan *ChannelData
	UnRegister chan *ChannelData
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

type ChannelData struct {
	Client  *Client
	Message *model.ChannelMessage
}

//todo : need to limit parallel db operations

func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[int64]*Client, avgClients),
		Rooms:      make(map[int64]*Room),
		Register:   make(chan *ChannelData),
		UnRegister: make(chan *ChannelData),
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
			if room, ok := h.Rooms[chd.Message.RoomId]; ok {
				if _, ok := room.Clients[chd.Client.UserId]; !ok {
					room.Clients[chd.Client.UserId] = chd.Client
				}
			}
		case chd := <-h.UnRegister:
			if room, ok := h.Rooms[chd.Message.RoomId]; ok {
				if _, ok := room.Clients[chd.Client.UserId]; ok {
					// Broadcast a message saying that the client left the room
					if len(room.Clients) != 0 {
						h.Broadcast <- &model.ChannelMessage{
							Content:  "user left the chat",
							RoomId:   chd.Message.RoomId,
							UserId:   chd.Client.UserId,
							Username: chd.Client.Username,
						}
					}

					delete(room.Clients, chd.Client.UserId)
					close(chd.Client.Message)
				}
			}
		case m := <-h.Broadcast:
			if room, ok := h.Rooms[m.RoomId]; ok {
				for _, cl := range room.Clients {
					cl.Message <- m
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

func HandleSingleChatMessage(d *amqp.Delivery, extraConsumerData interface{}) {
	// run as rabbitmq consumer
	wsSvc := extraConsumerData.(*WsService)
	var m *model.ChannelMessage
	err := json.Unmarshal(d.Body, &m)
	if err != nil {
		return
	}

	sender, ok := wsSvc.hub.Clients[m.UserId]
	err = wsSvc.wsRepo.SaveMessage(m)
	if err != nil {
		if errors.Is(err, gorm.ErrForeignKeyViolated) {
			// receiver user not found
			if ok {
				m.State = -1
				m.Code = 404
				m.ErrorMessage = "Receiver User Not Found"
				sender.Message <- m
			} else {
				// maybe save error
			}
		} else {
			if ok {
				m.State = -1
				m.Code = 500
				m.ErrorMessage = err.Error()
				sender.Message <- m
				// maybe save error
			} else {
				// maybe save error
			}
		}
	} else {
		cl, ok := wsSvc.hub.Clients[m.ReceiverId]
		if ok {
			//receiver is online
			cl.Message <- m
			m.Code = 200
			sender.Message <- m
		}
		err = wsSvc.wsRepo.UpdateUserReceivedMessageTime(m.ReceiverId)
	}

	if err = d.Ack(false); err != nil {
		log.Printf("error acking message: %s\n", err)
	}
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
	for {
		var clientMessage model.ClientMessage
		err := c.Conn.ReadJSON(&clientMessage)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		msg := &model.ChannelMessage{
			Content:    clientMessage.Content,
			RoomId:     clientMessage.RoomId,
			ReceiverId: clientMessage.ReceiverId,
			State:      1,
			UserId:     c.UserId,
			Username:   c.Username,
		}

		ctx, _ := context.WithCancel(context.Background())
		if clientMessage.RoomId == -1 {
			//one to one message
			conf := rabbitmq.NewConfigPublish(rabbitmq.ChatExchange, rabbitmq.SingleChatBindingKey)
			rabbit.Publish(ctx, msg, conf, c.UserId)
		} else {
			//group/channel message
			hub.Broadcast <- msg
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

		m := &model.ChannelMessage{
			Content:  "A new user has joined the room",
			UserId:   clientId,
			Username: username,
			RoomId:   roomId,
		}

		// Register a new client through the register channel
		w.hub.Register <- &ChannelData{Client: cl, Message: m}
		// Broadcast that message
		w.hub.Broadcast <- m

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
