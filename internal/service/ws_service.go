package service

import (
	"downloader_gochat/internal/repository"
	"downloader_gochat/model"
	"errors"
	"log"
	"time"

	"github.com/fasthttp/websocket"
	"github.com/valyala/fasthttp"
)

type IWsService interface {
	AddClient(ctx *fasthttp.RequestCtx, userId int64, username string) error
	CreateRoom(senderId int64, receiverId int64) (int64, error)
	JoinRoom(ctx *fasthttp.RequestCtx, roomId int64, clientId int64, username string) error
	GetRooms() (*[]model.RoomRes, error)
	GetRoomClient(roomId int64) (*[]model.ClientRes, error)
}

type WsService struct {
	wsRepo  repository.IWsRepository
	timeout time.Duration
	hub     *Hub
}

func NewWsService(WsRepo repository.IWsRepository) *WsService {
	wsSvc := WsService{
		wsRepo:  WsRepo,
		timeout: time.Duration(2) * time.Second,
		hub:     NewHub(),
	}
	go wsSvc.hub.Run()
	go wsSvc.hub.SingleChatRun(&wsSvc)
	return &wsSvc
}

type Hub struct {
	Clients         map[int64]*Client
	Rooms           map[int64]*Room
	Register        chan *ChannelData
	UnRegister      chan *ChannelData
	Broadcast       chan *model.ChannelMessage
	SingleBroadcast chan *model.ChannelMessage
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

func NewHub() *Hub {
	return &Hub{
		Clients:         make(map[int64]*Client),
		Rooms:           make(map[int64]*Room),
		Register:        make(chan *ChannelData),
		UnRegister:      make(chan *ChannelData),
		Broadcast:       make(chan *model.ChannelMessage, 5),
		SingleBroadcast: make(chan *model.ChannelMessage, 10),
	}
}

//------------------------------------------
//------------------------------------------

var upgrader = websocket.FastHTTPUpgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	//todo :
	CheckOrigin: func(r *fasthttp.RequestCtx) bool {
		//origin := r.Header.Get("Origin")
		//return origin == "http://localhost:3000"
		return true
	},
}

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second
	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second
	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

//------------------------------------------
//------------------------------------------

func (h *Hub) Run() {
	// run in separate goroutine
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
				}
			}
		}
	}
}

func (h *Hub) SingleChatRun(wsSvc *WsService) {
	// run in separate goroutine
	for {
		select {
		case m := <-h.SingleBroadcast:
			if cl, ok := h.Clients[m.ReceiverId]; ok {
				// receiver is online
				err := wsSvc.wsRepo.SaveMessage(m)
				if err != nil {
					//todo : notify sender

				} else {
					cl.Message <- m
				}
			} else {
				// receiver is offline
			}
		}
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

func (c *Client) ReadMessage(hub *Hub) {
	// user A wants to chat B
	// -- without room implementation --
	// 1. receiver is online, exist in hub.clients:
	//		1.1. send message to user, wright in its socket
	//		1.2. save message into db
	// 2. load receiver user from db
	//		2.1. if user not found, return error
	// 3. save message into some king of queue to send to user later
	// 4. save message into db

	// -- with room implementation --
	// 1. first time chatting::
	// `1.1. client requests to create room
	//	1.2. save room data into db
	//	1.3. add room to hub.rooms, add receiver user to room if exist in hub.clients
	//	1.4. send roomId to client
	// 2. room already exist, roomId is provided in the readMessage
	//	2.1. roomId exists in hub.rooms::
	//		2.1.1. receiver is online, exist in hub.room.clients:
	//			2.1.1.1. send message to user, wright in its socket
	//			2.1.1.2. save message into db
	//		2.1.2. receiver is offline
	//			2.1.2.1. save message into some king of queue to send to user later
	//			2.1.2.2. save message into db
	//			2.1.2.3. add client to room.clients after receiver login
	//	2.2. roomId doesn't exist in hub.rooms
	//	2.3. load room data from db::
	//		2.3.1. room doesn't exist, return error
	//		2.3.2. load receiver user from db
	//		2.3.2. add room to hub.rooms, add receiver user to room if exist in hub.clients::
	//			2.3.2.1. receiver is online, exist in hub.room.clients::
	//				2.3.2.1.1. send message to user, wright in its socket
	//				2.3.2.1.2. save message into db
	//			2.3.2.2. receiver is offline
	//				2.3.2.1.1. save message into some king of queue to send to user later
	//				2.3.2.1.2. save message into db
	//				2.3.2.1.3. add client to room.clients after receiver login
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

		if clientMessage.RoomId == -1 {
			//one to one message
			hub.SingleBroadcast <- msg
		} else {
			//group/channel message
			hub.Broadcast <- msg
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

		go cl.WriteMessage()
		cl.ReadMessage(w.hub)
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
		cl.ReadMessage(w.hub)
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
