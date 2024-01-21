package repository

import (
	"downloader_gochat/model"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

type IWsRepository interface {
	CreateRoom(senderId int64, receiverId int64) (int64, error)
	SaveMessage(message *model.ChannelMessage) error
}

type WsRepository struct {
	db      *gorm.DB
	mongodb *mongo.Database
}

func NewWsRepository(db *gorm.DB, mongodb *mongo.Database) *WsRepository {
	return &WsRepository{db: db, mongodb: mongodb}
}

//------------------------------------------
//------------------------------------------

func (w *WsRepository) CreateRoom(senderId int64, receiverId int64) (int64, error) {
	//todo : handle when room already exist
	//todo : check db, room exist, create room, return roomId
	return 55, nil
}

func (w *WsRepository) SaveMessage(message *model.ChannelMessage) error {
	m := model.Message{
		CreatorId:  message.UserId,
		ReceiverId: message.ReceiverId,
		Content:    message.Content,
		RoomId:     &message.RoomId,
		Date:       time.Now().UTC(),
		State:      1,
	}
	if *m.RoomId == -1 {
		m.RoomId = nil
	}
	err := w.db.Create(&m).Error
	if err != nil {
		return err
	}
	return nil
}
