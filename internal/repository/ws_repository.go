package repository

import (
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

type IWsRepository interface {
	CreateRoom(senderId string, receiverId string) (string, error)
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

func (w *WsRepository) CreateRoom(senderId string, receiverId string) (string, error) {
	//todo : handle when room already exist
	//todo : check db, room exist, create room, return roomId
	return "sampleRoomId", nil
}
