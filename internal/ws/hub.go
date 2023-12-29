package ws

type Room struct {
	ID      string             `json:"id"`
	Name    string             `json:"name"`
	Clients map[string]*Client `json:"clients"`
}

type Hube struct {
	Rooms      map[string]*Room
	Register   chan *Client
	UnRegister chan *Client
	Broadcast  chan *Message
}

func NewHub() *Hube {
	return &Hube{
		Rooms:      make(map[string]*Room),
		Register:   make(chan *Client),
		UnRegister: make(chan *Client),
		Broadcast:  make(chan *Message, 5),
	}
}

func (h *Hube) Run() {
	//run in separate goroutine
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
					//Broadcast a message saying that the client left the room
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
