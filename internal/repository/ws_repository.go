package repository

import (
	"downloader_gochat/model"
	"errors"
	"slices"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type IWsRepository interface {
	GetReceiverUser(userId int64) (*model.UserDataModel, error)
	CreateRoom(senderId int64, receiverId int64) (int64, error)
	SaveMessage(message *model.ReceiveNewMessage) error
	UpdateUserReceivedMessageTime(userId int64) error
	UpdateUserReadMessageTime(userId int64, readTime time.Time) error
	GetSingleChatMessages(params *model.GetSingleMessagesReq) (*[]model.MessageDataModel, error)
	GetSingleChatList(params *model.GetSingleChatListReq) ([]model.ChatsDataModel, []model.ProfileImageDataModel, error)
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

func (w *WsRepository) SaveMessage(message *model.ReceiveNewMessage) error {
	m := model.Message{
		CreatorId:  message.UserId,
		ReceiverId: message.ReceiverId,
		Content:    message.Content,
		RoomId:     &message.RoomId,
		Date:       time.Now().UTC(),
		State:      message.State,
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

func (w *WsRepository) UpdateUserReceivedMessageTime(userId int64) error {
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

func (w *WsRepository) UpdateUserReadMessageTime(userId int64, readTime time.Time) error {
	result := w.db.Model(&model.UserMessageRead{}).
		Where("\"userId\" = ?", userId).
		UpdateColumn("\"lastTimeRead\"", readTime)
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

	query := "\"roomId\" IS NULL AND date > @date AND ((\"creatorId\" = @userid AND \"receiverId\" = @receiverid) OR (\"creatorId\" = @receiverid AND \"receiverId\" = @userid))"
	if params.ReverseOrder {
		query = "\"roomId\" IS NULL AND date < @date AND ((\"creatorId\" = @userid AND \"receiverId\" = @receiverid) OR (\"creatorId\" = @receiverid AND \"receiverId\" = @userid))"
	}
	if params.MessageState != 0 {
		query = query + " AND state = @messagestate"
	}

	err := w.db.Model(&model.Message{}).
		Where(query, map[string]interface{}{
			"date":         params.Date,
			"userid":       params.UserId,
			"receiverid":   params.ReceiverId,
			"messagestate": params.MessageState,
		}).
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

func (w *WsRepository) GetSingleChatList(params *model.GetSingleChatListReq) ([]model.ChatsDataModel, []model.ProfileImageDataModel, error) {
	var chats []model.ChatsDataModel

	//SELECT t_limited.*, "User"."publicName", "User".username, "User"."userId", "User".role
	//FROM (
	//    (
	//         SELECT DISTINCT "creatorId"
	//         FROM "Message"
	//        offset 0
	//         limit 20
	//     ) as t_groups
	//    JOIN LATERAL (
	//        SELECT *
	//        FROM "Message" t_all
	//        WHERE t_all."creatorId" = t_groups."creatorId" and t_all."receiverId" = 4
	//        ORDER BY t_all.date desc
	//        offset 0
	//        LIMIT 2
	//    ) as t_limited ON t_limited.state = 1 and t_limited."roomId" IS NULL
	//) join "User" on t_limited."creatorId" = "User"."userId";

	queryStr := "SELECT t_limited.*, \"User\".\"publicName\", \"User\".username, \"User\".\"userId\", \"User\".role " +
		"FROM ( ( SELECT DISTINCT \"creatorId\" FROM \"Message\" offset @chatskip limit @chatlimit) as t_groups " +
		"JOIN LATERAL (SELECT * FROM \"Message\" t_all WHERE t_all.\"creatorId\" = t_groups.\"creatorId\" and t_all.\"receiverId\" = @receiverid " +
		"ORDER BY t_all.date desc Offset @messageskip LIMIT @messagelimit) " +
		"as t_limited ON t_limited.state = @messagestate and t_limited.\"roomId\" IS NULL) join \"User\" on t_limited.\"creatorId\" = \"User\".\"userId\";"
	if params.MessageState == 0 {
		queryStr = strings.Replace(queryStr, "t_limited.state = @messagestate and ", "", 1)
	}

	err := w.db.Raw(queryStr,
		map[string]interface{}{
			"chatskip":     params.ChatsSkip,
			"chatlimit":    params.ChatsLimit,
			"receiverid":   params.UserId,
			"messageskip":  params.MessagePerChatSkip,
			"messagelimit": params.MessagePerChatLimit,
			"messagestate": params.MessageState,
		}).
		Scan(&chats).Error

	var profileImages []model.ProfileImageDataModel
	if params.IncludeProfileImages {
		userIds := make([]int64, 0)
		for _, ch := range chats {
			if !slices.Contains(userIds, ch.UserId) {
				userIds = append(userIds, ch.UserId)
			}
		}
		err = w.db.Model(&model.ProfileImage{}).Where("\"userId\" In ?", userIds).Find(&profileImages).Error
		if err != nil {
			profileImages = make([]model.ProfileImageDataModel, 0)
		}
	} else {
		profileImages = make([]model.ProfileImageDataModel, 0)
	}

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			t := make([]model.ChatsDataModel, 0)
			return t, profileImages, nil
		}
		return nil, profileImages, err
	}
	return chats, profileImages, nil
}
