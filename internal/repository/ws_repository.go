package repository

import (
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

type IWsRepository interface {
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
