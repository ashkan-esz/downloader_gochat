package service

import (
	"downloader_gochat/model"
	"sync"
)

type IHub interface {
	getClient(userId int64) (*Client, bool, *sync.RWMutex)
	addClientToHub(userId int64, client *Client)
	getRoom(roomId int64) (*Room, bool)
	getRoomClient(room *Room, userId int64) (*Client, bool)
	addClientToRoom(room *Room, userId int64, client *Client)
	removeClientFromRoom(room *Room, userId int64)
}

type Hub struct {
	Clients       map[int64]*Client
	ClientsRwLock *sync.RWMutex
	Rooms         map[int64]*Room
	RoomsRwLock   *sync.RWMutex
	Register      chan *model.ChannelMessage
	UnRegister    chan *model.ChannelMessage
	Broadcast     chan *model.ChannelMessage
}

func NewHub() *Hub {
	return &Hub{
		Clients:       make(map[int64]*Client, avgClients),
		ClientsRwLock: &sync.RWMutex{},
		Rooms:         make(map[int64]*Room),
		RoomsRwLock:   &sync.RWMutex{},
		Register:      make(chan *model.ChannelMessage),
		UnRegister:    make(chan *model.ChannelMessage),
		Broadcast:     make(chan *model.ChannelMessage, 5),
	}
}

//------------------------------------------
//------------------------------------------

func (h *Hub) getClient(userId int64) (*Client, bool, *sync.RWMutex) {
	h.ClientsRwLock.RLock()
	defer h.ClientsRwLock.RUnlock()
	cl, ok := h.Clients[userId]
	return cl, ok, h.ClientsRwLock
}

func (h *Hub) addClientToHub(userId int64, client *Client) {
	h.ClientsRwLock.Lock()
	defer h.ClientsRwLock.Unlock()
	h.Clients[userId] = client
}

//------------------------------------------
//------------------------------------------

func (h *Hub) getRoom(roomId int64) (*Room, bool) {
	h.RoomsRwLock.RLock()
	defer h.RoomsRwLock.RUnlock()
	room, ok := h.Rooms[roomId]
	return room, ok
}

func (h *Hub) getRoomClient(room *Room, userId int64) (*Client, bool) {
	h.RoomsRwLock.RLock()
	defer h.RoomsRwLock.RUnlock()
	cl, ok := room.Clients[userId]
	return cl, ok
}

func (h *Hub) addClientToRoom(room *Room, userId int64, client *Client) {
	h.RoomsRwLock.Lock()
	defer h.RoomsRwLock.Unlock()
	room.Clients[userId] = client
}

func (h *Hub) removeClientFromRoom(room *Room, userId int64) {
	h.RoomsRwLock.Lock()
	defer h.RoomsRwLock.Unlock()
	delete(room.Clients, userId)
}
