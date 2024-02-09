package service

import (
	"time"

	"github.com/fasthttp/websocket"
	"github.com/valyala/fasthttp"
)

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

const (
	avgClients    = 512
	dbBufSize     = 64
	flushInterval = time.Second * 5
)
