package ws

import (
	"github.com/fasthttp/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"
)

type Handler struct {
	hube *Hube
}

func NewHandler(h *Hube) *Handler {
	return &Handler{
		hube: h,
	}
}

type CreateRoomReq struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (h *Handler) CreateRoom(c *fiber.Ctx) error {
	var req CreateRoomReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(map[string]string{"error": err.Error()})
	}

	h.hube.Rooms[req.ID] = &Room{
		ID:      req.ID,
		Name:    req.Name,
		Clients: make(map[string]*Client),
	}

	return c.Status(fiber.StatusOK).JSON(req)
}

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

func (h *Handler) JoinRoom(c *fiber.Ctx) error {
	err := upgrader.Upgrade(c.Context(), func(conn *websocket.Conn) {
		roomId := c.Params("roomId")
		clientId := c.Query("userId")
		username := c.Query("username")

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
		h.hube.Register <- cl
		// Broadcast that message
		h.hube.Broadcast <- m

		go cl.WriteMessage()
		cl.ReadMessage(h.hube)
	})
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(map[string]string{"error": err.Error()})
	}

	return err
}

type RoomRes struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (h *Handler) GetRooms(c *fiber.Ctx) error {
	rooms := make([]RoomRes, 0)

	for _, r := range h.hube.Rooms {
		rooms = append(rooms, RoomRes{ID: r.ID, Name: r.Name})
	}

	return c.Status(fiber.StatusOK).JSON(rooms)
}

type ClientRes struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

func (h *Handler) GetClients(c *fiber.Ctx) error {
	var clients []ClientRes
	roomId := c.Params("roomId")

	if _, ok := h.hube.Rooms[roomId]; !ok {
		clients = make([]ClientRes, 0)
		return c.Status(fiber.StatusOK).JSON(clients)
	}

	for _, cl := range h.hube.Rooms[roomId].Clients {
		clients = append(clients, ClientRes{ID: cl.ID, Username: cl.Username})
	}

	return c.Status(fiber.StatusOK).JSON(clients)
}
