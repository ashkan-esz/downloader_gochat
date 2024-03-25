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
	userRep  repository.IUserRepository
	rabbitmq rabbitmq.RabbitMQ
	timeout  time.Duration
	hub      *Hub
}

const (
	userMessageConsumerCount  = 10
	groupMessageConsumerCount = 1
	messageStateConsumerCount = 3
)

var globalHub *Hub

func NewWsService(WsRepo repository.IWsRepository, userRep repository.IUserRepository, rabbit rabbitmq.RabbitMQ) *WsService {
	wsSvc := WsService{
		wsRepo:   WsRepo,
		userRep:  userRep,
		rabbitmq: rabbit,
		timeout:  time.Duration(2) * time.Second,
		hub:      NewHub(),
	}
	globalHub = wsSvc.hub

	config := rabbitmq.NewConfigConsume(rabbitmq.SingleChatQueue, "")
	for i := 0; i < userMessageConsumerCount; i++ {
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
	for i := 0; i < groupMessageConsumerCount; i++ {
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
	for i := 0; i < messageStateConsumerCount; i++ {
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

func getClientFromHub(userId int64) (*Client, bool) {
	globalHub.ClientsRwLock.RLock()
	defer globalHub.ClientsRwLock.RUnlock()
	cl, ok := globalHub.Clients[userId]
	return cl, ok
}

//------------------------------------------
//------------------------------------------

func (h *Hub) RunGroupHandler() {
	//Note : use rabbitmq

	for {
		select {
		case chd := <-h.Register:
			if room, ok := h.getRoom(chd.ReceiveNewMessage.RoomId); ok {
				if _, ok := h.getRoomClient(room, chd.ReceiveNewMessage.UserId); !ok {
					client, _, _ := h.getClient(chd.ReceiveNewMessage.UserId)
					h.addClientToRoom(room, chd.ReceiveNewMessage.UserId, client)
				}
			}
		case chd := <-h.UnRegister:
			if room, ok := h.getRoom(chd.ReceiveNewMessage.RoomId); ok {
				if _, ok := h.getRoomClient(room, chd.ReceiveNewMessage.UserId); ok {
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

					h.removeClientFromRoom(room, chd.ReceiveNewMessage.UserId)
					if client, ok, _ := h.getClient(chd.ReceiveNewMessage.UserId); ok {
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
			//if sender, ok, _ := wsSvc.hub.getClient(channelMessage.ChatMessagesReq.UserId); ok {
			//	errorData := model.CreateActionError(500, err.Error(), model.SingleChatMessagesAction, channelMessage.ChatMessagesReq)
			//	sender.Message <- errorData
			//}
		} else {
			m := model.CreateReturnChatMessagesAction(chatMessages)
			if sender, ok, _ := wsSvc.hub.getClient(channelMessage.ChatMessagesReq.UserId); ok {
				sender.Message <- m
			}
		}
	case model.SingleChatsListAction:
		if _, ok, _ := wsSvc.hub.getClient(channelMessage.ChatsListReq.UserId); ok {
			chatMessages, err := wsSvc.GetSingleChatList(channelMessage.ChatsListReq)
			if err != nil {
				if err = d.Nack(false, true); err != nil {
					log.Printf("error nacking message: %s\n", err)
				}
				//if sender, ok, _ := wsSvc.hub.getClient(channelMessage.ChatsListReq.UserId); ok {
				//	errorData := model.CreateActionError(500, err.Error(), model.SingleChatsListAction, channelMessage.ChatsListReq)
				//	sender.Message <- errorData
				//}
			} else {
				m := model.CreateReturnChatListAction(chatMessages)
				if sender, ok, _ := wsSvc.hub.getClient(channelMessage.ChatsListReq.UserId); ok {
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
		err := wsSvc.wsRepo.BatchUpdateMessageState(
			channelMessage.MessageRead.Id,
			channelMessage.MessageRead.RoomId,
			channelMessage.MessageRead.UserId,
			channelMessage.MessageRead.ReceiverId,
			channelMessage.MessageRead.State)
		if err != nil {
			if messageReceiver, ok, _ := wsSvc.hub.getClient(channelMessage.MessageRead.ReceiverId); ok {
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
			if messageCreator, ok, _ := wsSvc.hub.getClient(channelMessage.MessageRead.UserId); ok {
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
	case model.UserStatusAction:
		req := channelMessage.UserStatusReq
		res := model.UserStatusRes{
			Type:          model.UserStatusOnlineUsers,
			OnlineUserIds: []int64{},
		}
		for _, id := range req.UserIds {
			cl, ok, _ := wsSvc.hub.getClient(id)
			if ok && cl != nil {
				res.OnlineUserIds = append(res.OnlineUserIds, id)
			}
		}
		m := model.CreateSendUserStatusAction(&res)
		if user, ok, _ := wsSvc.hub.getClient(req.UserId); ok {
			user.Message <- m
		}
	case model.UserIsTypingAction:
		req := channelMessage.UserStatusReq
		for _, id := range req.UserIds {
			cl, ok, _ := wsSvc.hub.getClient(id)
			if ok && cl != nil {
				res := model.UserStatusRes{
					Type:            req.Type,
					IsTypingUserIds: []int64{},
				}
				res.IsTypingUserIds = []int64{req.UserId}
				m := model.CreateSendUserStatusAction(&res)
				cl.Message <- m
			}
		}
	}

	if err = d.Ack(false); err != nil {
		log.Printf("error acking [messageState] message: %s\n", err)
	}
}

func HandleSingleChatMessage(receiveNewMessage *model.ReceiveNewMessage, wsSvc *WsService) error {
	sender, senderExist, _ := wsSvc.hub.getClient(receiveNewMessage.UserId)
	mid, err := wsSvc.wsRepo.SaveMessage(receiveNewMessage)
	if err != nil {
		if errors.Is(err, gorm.ErrForeignKeyViolated) {
			// receiver user not found
			if senderExist {
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
			if senderExist {
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
		cl, receiverExist, _ := wsSvc.hub.getClient(receiveNewMessage.ReceiverId)
		if receiverExist {
			//receiver is online
			receiveNewMessage.Id = mid

			// add creator profileImage, read from cache only
			userCacheData, _ := getCachedUserData(receiveNewMessage.UserId)
			if userCacheData != nil && len(userCacheData.ProfileImages) > 0 {
				receiveNewMessage.CreatorImage = userCacheData.ProfileImages[0].Thumbnail
			}

			receiveMessage := model.CreateReceiveNewMessageAction(receiveNewMessage)
			cl.Message <- receiveMessage
		}

		if senderExist {
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

		if !receiverExist {
			//receiver is offline
			// don't need to save this notification, show notification in app, send push-notification (only if user is offline)
			ctx, _ := context.WithCancel(context.Background())
			//defer cancel()
			notifQueueConf := rabbitmq.NewConfigPublish(rabbitmq.NotificationExchange, rabbitmq.NotificationBindingKey)
			notifMessage := model.CreateNewMessageNotificationAction(receiveNewMessage)
			wsSvc.rabbitmq.Publish(ctx, notifMessage, notifQueueConf, receiveNewMessage.ReceiverId)
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

func (c *ClientConnection) ReadMessage(cc *Client, hub *Hub, rabbit rabbitmq.RabbitMQ, wsRepo repository.IWsRepository) {
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
			_ = wsRepo.UpdateUserLastSeenTime(cc.UserId, time.Now().UTC())
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
				//consider end of typing
				if _, ok, _ := hub.getClient(clientMessage.NewMessage.ReceiverId); ok {
					readQueueConf := rabbitmq.NewConfigPublish(rabbitmq.MessageStateExchange, rabbitmq.MessageStateBindingKey)
					userStatusReq := model.UserStatusReq{
						Type:    model.UserStatusStopTyping,
						UserId:  cc.UserId,
						UserIds: []int64{clientMessage.NewMessage.ReceiverId},
					}
					message2 := model.CreateSendUserIsTypingAction(&userStatusReq)
					rabbit.Publish(ctx, message2, readQueueConf, cc.UserId)
				}
			} else {
				//group/channel message
				hub.Broadcast <- receiveMessage
			}
		case model.SingleChatMessagesAction:
			clientMessage.ChatMessagesReq.UserId = cc.UserId
			message := model.CreateGetChatMessagesAction(&clientMessage.ChatMessagesReq)
			rabbit.Publish(ctx, message, conf, cc.UserId)
		case model.SingleChatsListAction:
			clientMessage.ChatsListReq.UserId = cc.UserId
			message := model.CreateGetChatListAction(&clientMessage.ChatsListReq)
			rabbit.Publish(ctx, message, conf, cc.UserId)
		case model.MessageReadAction:
			clientMessage.MessageRead.ReceiverId = cc.UserId
			readQueueConf := rabbitmq.NewConfigPublish(rabbitmq.MessageStateExchange, rabbitmq.MessageStateBindingKey)
			message := model.CreateMessageReadAction(
				clientMessage.MessageRead.Id,
				clientMessage.MessageRead.RoomId,
				clientMessage.MessageRead.UserId,
				clientMessage.MessageRead.ReceiverId,
				clientMessage.MessageRead.Date,
				2, false)
			rabbit.Publish(ctx, message, readQueueConf, cc.UserId)
		case model.UserStatusAction:
			clientMessage.UserStatusReq.UserId = cc.UserId
			readQueueConf := rabbitmq.NewConfigPublish(rabbitmq.MessageStateExchange, rabbitmq.MessageStateBindingKey)
			message := model.CreateGetUserStatusAction(clientMessage.UserStatusReq)
			rabbit.Publish(ctx, message, readQueueConf, cc.UserId)
		case model.UserIsTypingAction:
			clientMessage.UserStatusReq.UserId = cc.UserId
			readQueueConf := rabbitmq.NewConfigPublish(rabbitmq.MessageStateExchange, rabbitmq.MessageStateBindingKey)
			message := model.CreateSendUserIsTypingAction(clientMessage.UserStatusReq)
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

		client, ok, clientRwLock := w.hub.getClient(userId)
		if ok {
			clientRwLock.Lock()
			client.Connections = append(client.Connections, connection)
			clientRwLock.Unlock()
		} else {
			cl := &Client{
				Connections: []*ClientConnection{connection},
				Message:     make(chan *model.ChannelMessage, 10),
				UserId:      userId,
				Username:    username,
			}

			w.hub.addClientToHub(userId, cl)
			client = cl

			go connection.WriteMessage(cl)
		}

		_ = w.wsRepo.UpdateUserLastSeenTime(userId, time.Now().UTC())

		chatsListReq := &model.GetSingleChatListReq{
			UserId:               userId,
			MessagePerChatLimit:  2,
			ChatsLimit:           20,
			IncludeProfileImages: true,
		}
		message := model.CreateGetChatListAction(chatsListReq)
		conf := rabbitmq.NewConfigPublish(rabbitmq.ChatExchange, rabbitmq.SingleChatBindingKey)
		w.rabbitmq.Publish(ctx, message, conf, userId)

		// load profileImage and notification settings from db, save to redis for cache
		notificationSettings, _ := w.userRep.GetUserMetaDataAndNotificationSettings(userId, 1)
		if notificationSettings != nil {
			cacheData := model.CachedUserData{
				UserId:        notificationSettings.UserId,
				Username:      notificationSettings.Username,
				PublicName:    notificationSettings.PublicName,
				ProfileImages: notificationSettings.ProfileImages,
				NotificationSettings: model.NotificationSettings{
					UserId:                    userId,
					NewFollower:               notificationSettings.NewFollower,
					NewMessage:                notificationSettings.NewMessage,
					FinishedListSpinOffSequel: notificationSettings.FinishedListSpinOffSequel,
					FollowMovie:               notificationSettings.FollowMovie,
					FollowMovieBetterQuality:  notificationSettings.FollowMovieBetterQuality,
					FollowMovieSubtitle:       notificationSettings.FollowMovieSubtitle,
					FutureList:                notificationSettings.FutureList,
					FutureListSerialSeasonEnd: notificationSettings.FutureListSerialSeasonEnd,
					FutureListSubtitle:        notificationSettings.FutureListSubtitle,
				},
				NotifTokens: []string{},
			}

			for i := range notificationSettings.ActiveSessions {
				cacheData.NotifTokens = append(cacheData.NotifTokens, notificationSettings.ActiveSessions[i].NotifToken)
			}
			cacheData.NotifTokens = slices.Compact(cacheData.NotifTokens)

			_ = setUserDataCache(userId, &cacheData)
			client.Message <- model.CreateNotificationSettingsAction(&cacheData.NotificationSettings)
		}

		connection.ReadMessage(client, w.hub, w.rabbitmq, w.wsRepo)
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

	creatorIds := make([]int64, 0)
	for i := range chats {
		creatorIds = append(creatorIds, chats[i].UserId)
	}
	creatorIds = slices.Compact(creatorIds)
	counts, err := w.wsRepo.GetSingleChatsMessageCount(creatorIds, params.UserId, 1)
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
			Medias: []model.MediaFile{
				{
					Id:        chat.MediaFileId,
					MessageId: chat.Id,
					Date:      chat.MediaFileDate,
					Url:       chat.Url,
					Type:      chat.Type,
					Size:      chat.Size,
					Thumbnail: chat.Thumbnail,
					BlurHash:  chat.BlurHash,
				},
			},
		}
		if chat.MediaFileId == 0 {
			m.Medias = nil
		}
		exist := false
		for i := range compressedChats {
			if chat.UserId == compressedChats[i].UserId {
				exist = true
				messageExist := false
				if m.Medias != nil {
					for i2, message := range compressedChats[i].Messages {
						if message.Id == m.Id {
							messageExist = true
							compressedChats[i].Messages[i2].Medias = append(compressedChats[i].Messages[i2].Medias, m.Medias[0])
							break
						}
					}
				}
				if !messageExist {
					compressedChats[i].Messages = append(compressedChats[i].Messages, m)
				}
				break
			}
		}
		if !exist {
			chatUserId := chat.UserId
			if chatUserId == params.UserId {
				chatUserId = m.ReceiverId
			}
			cChat := model.ChatsCompressedDataModel{
				UserId:        chatUserId,
				Username:      chat.Username,
				PublicName:    chat.PublicName,
				Role:          chat.Role,
				LastSeenDate:  chat.LastSeenDate,
				ProfileImages: filterProfileImages(profileImages, chat.UserId),
				Messages:      []model.MessageDataModel{m},
				IsOnline:      false,
			}
			for i := range counts {
				if counts[i].CreatorId == cChat.UserId {
					cChat.UnreadMessagesCount = counts[i].Count
					break
				}
			}
			compressedChats = append(compressedChats, cChat)
		}
	}

	for i := range compressedChats {
		slices.SortFunc(compressedChats[i].Messages, func(a, b model.MessageDataModel) int {
			return b.Date.Compare(a.Date)
		})
		cl, ok, _ := w.hub.getClient(compressedChats[i].UserId)
		compressedChats[i].IsOnline = ok && cl != nil

		for i2 := range compressedChats[i].Messages {
			slices.SortFunc(compressedChats[i].Messages[i2].Medias, func(a, b model.MediaFile) int {
				return a.Date.Compare(b.Date)
			})
		}
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
