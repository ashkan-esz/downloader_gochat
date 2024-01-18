package service

import (
	"downloader_gochat/internal/repository"
	"downloader_gochat/model"
	"log"
	"time"

	"github.com/fasthttp/websocket"
	"github.com/valyala/fasthttp"
)

type IWsService interface {
	CreateRoom(roomId string, roomName string) error
	JoinRoom(ctx *fasthttp.RequestCtx, roomId string, clientId string, username string) error
	GetRooms() (*[]model.RoomRes, error)
	GetRoomClient(roomId string) (*[]model.ClientRes, error)
}

type WsService struct {
	wsRepo  repository.IWsRepository
	timeout time.Duration
	hube    *Hube
}

func NewWsService(WsRepo repository.IWsRepository) *WsService {
	wsSvc := WsService{
		wsRepo:  WsRepo,
		timeout: time.Duration(2) * time.Second,
		hube:    NewHub(),
	}
	go wsSvc.hube.Run()
	return &wsSvc
}

type Hube struct {
	Rooms      map[string]*Room
	Register   chan *Client
	UnRegister chan *Client
	Broadcast  chan *Message
}

type Room struct {
	ID      string             `json:"id"`
	Name    string             `json:"name"`
	Clients map[string]*Client `json:"clients"`
}

type Client struct {
	Conn     *websocket.Conn
	Message  chan *Message
	ID       string `json:"id"`
	RoomId   string `json:"roomId"`
	Username string `json:"username"`
}

type Message struct {
	Content  string `json:"content"`
	RoomId   string `json:"roomId"`
	Username string `json:"username"`
}

func NewHub() *Hube {
	return &Hube{
		Rooms:      make(map[string]*Room),
		Register:   make(chan *Client),
		UnRegister: make(chan *Client),
		Broadcast:  make(chan *Message, 5),
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

func (h *Hube) Run() {
	// run in separate goroutine
	for {
		select {
		case cl := <-h.Register:
			if _, ok := h.Rooms[cl.RoomId]; ok {
				r := h.Rooms[cl.RoomId]
				if _, ok := r.Clients[cl.ID]; !ok {
					r.Clients[cl.ID] = cl
				}
			}
		case cl := <-h.UnRegister:
			if _, ok := h.Rooms[cl.RoomId]; ok {
				if _, ok := h.Rooms[cl.RoomId].Clients[cl.ID]; ok {
					// Broadcast a message saying that the client left the room
					if len(h.Rooms[cl.RoomId].Clients) != 0 {
						h.Broadcast <- &Message{
							Content:  "user left the chat",
							RoomId:   cl.RoomId,
							Username: cl.Username,
						}
					}

					delete(h.Rooms[cl.RoomId].Clients, cl.ID)
					close(cl.Message)
				}
			}
		case m := <-h.Broadcast:
			if _, ok := h.Rooms[m.RoomId]; ok {
				for _, cl := range h.Rooms[m.RoomId].Clients {
					cl.Message <- m
				}
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

func (c *Client) ReadMessage(hube *Hube) {
	defer func() {
		hube.UnRegister <- c
		c.Conn.Close()
	}()
	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		_, m, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		msg := &Message{
			Content:  string(m),
			RoomId:   c.RoomId,
			Username: c.Username,
		}

		hube.Broadcast <- msg
	}
}

//------------------------------------------
//------------------------------------------

func (w *WsService) CreateRoom(roomId string, roomName string) error {
	w.hube.Rooms[roomId] = &Room{
		ID:      roomId,
		Name:    roomName,
		Clients: make(map[string]*Client),
	}

	return nil
}

func (w *WsService) JoinRoom(ctx *fasthttp.RequestCtx, roomId string, clientId string, username string) error {
	err := upgrader.Upgrade(ctx, func(conn *websocket.Conn) {
		cl := &Client{
			Conn:     conn,
			Message:  make(chan *Message, 10),
			ID:       clientId,
			RoomId:   roomId,
			Username: username,
		}

		m := &Message{
			Content:  "A new user has joined the room",
			Username: username,
			RoomId:   roomId,
		}

		// Register a new client through the register channel
		w.hube.Register <- cl
		// Broadcast that message
		w.hube.Broadcast <- m

		go cl.WriteMessage()
		cl.ReadMessage(w.hube)
	})

	return err
}

func (w *WsService) GetRooms() (*[]model.RoomRes, error) {
	rooms := make([]model.RoomRes, 0)

	for _, r := range w.hube.Rooms {
		rooms = append(rooms, model.RoomRes{ID: r.ID, Name: r.Name})
	}

	return &rooms, nil
}

func (w *WsService) GetRoomClient(roomId string) (*[]model.ClientRes, error) {
	var clients []model.ClientRes
	if _, ok := w.hube.Rooms[roomId]; !ok {
		clients = make([]model.ClientRes, 0)
		return &clients, nil
	}

	for _, cl := range w.hube.Rooms[roomId].Clients {
		clients = append(clients, model.ClientRes{ID: cl.ID, Username: cl.Username})
	}

	return &clients, nil
}
