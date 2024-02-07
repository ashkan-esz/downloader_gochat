package repository

import (
	"downloader_gochat/model"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type IWsRepository interface {
	GetReceiverUser(userId int64) (*model.UserDataModel, error)
	CreateRoom(senderId int64, receiverId int64) (int64, error)
	SaveMessage(message *model.ChannelMessage) error
	UpdateUserReadMessageTime(userId int64) error
	GetSingleChatMessages(params *model.GetSingleMessagesReq) (*[]model.MessageDataModel, error)
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

func (w *WsRepository) GetReceiverUser(userId int64) (*model.UserDataModel, error) {
	var userDataModel model.UserDataModel
	err := w.db.Where("\"userId\" = ?", userId).
		Model(&model.User{}).
		Limit(1).
		Find(&userDataModel).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &userDataModel, nil
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

func (w *WsRepository) UpdateUserReadMessageTime(userId int64) error {
	result := w.db.Model(&model.UserMessageRead{}).
		Where("\"userId\" = ?", userId).
		UpdateColumn("\"lastMessageReceived\"", time.Now().UTC())
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil
		}
		return result.Error
	}

	return nil
}

//------------------------------------------
//------------------------------------------

func (w *WsRepository) GetSingleChatMessages(params *model.GetSingleMessagesReq) (*[]model.MessageDataModel, error) {
	var messages []model.MessageDataModel

	query := "\"roomId\" IS NULL AND date > ? AND ((\"creatorId\" = ? AND \"receiverId\" = ?) OR (\"creatorId\" = ? AND \"receiverId\" = ?))"
	if params.ReverseOrder {
		query = "\"roomId\" IS NULL AND date < ? AND ((\"creatorId\" = ? AND \"receiverId\" = ?) OR (\"creatorId\" = ? AND \"receiverId\" = ?))"
	}

	err := w.db.Debug().Model(&model.Message{}).
		Where(query, params.Date, params.UserId, params.ReceiverId, params.ReceiverId, params.UserId).
		Order(clause.OrderByColumn{Column: clause.Column{Name: "date"}, Desc: params.ReverseOrder}).
		Offset(params.Skip).
		Limit(params.Limit).
		Find(&messages).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			t := make([]model.MessageDataModel, 0)
			return &t, nil
		}
		return nil, err
	}
	return &messages, nil
}
