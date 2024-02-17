package repository

import (
	"downloader_gochat/model"
	"errors"

	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

type INotificationRepository interface {
	SaveUserNotification(notifData *model.NotificationDataModel) error
	GetUserNotifications(userId int64, skip int, limit int, entityTypeId int, status int) ([]model.NotificationDataModel, error)
	GetUserMetaData(id int64) (*model.UserMetaDataModel, error)
	GetBatchUserMetaData(ids []int64) ([]model.UserMetaDataModel, error)
}

type NotificationRepository struct {
	db      *gorm.DB
	mongodb *mongo.Database
}

func NewNotificationRepository(db *gorm.DB, mongodb *mongo.Database) *NotificationRepository {
	return &NotificationRepository{db: db, mongodb: mongodb}
}

//------------------------------------------
//------------------------------------------

//todo : handle flag notification as seen

func (n *NotificationRepository) SaveUserNotification(notifData *model.NotificationDataModel) error {
	notif := model.Notification{
		CreatorId:    notifData.CreatorId,
		ReceiverId:   notifData.ReceiverId,
		Date:         notifData.Date,
		Status:       notifData.Status,
		EntityId:     notifData.EntityId,
		EntityTypeId: notifData.EntityTypeId,
	}
	err := n.db.Create(&notif).Error
	notifData.Id = notif.Id
	return err
}

func (n *NotificationRepository) GetUserNotifications(userId int64, skip int, limit int, entityTypeId int, status int) ([]model.NotificationDataModel, error) {
	var notifications []model.NotificationDataModel
	queryStr := "\"receiverId\" = @receiverid "
	if entityTypeId > 0 {
		queryStr = "\"receiverId\" = @receiverid AND \"entityTypeId\" = @entitytypeid "
	}
	if status > 0 {
		queryStr = queryStr + " AND status = @status"
	}
	err := n.db.Where(queryStr, map[string]interface{}{
		"receiverid":   userId,
		"entitytypeid": entityTypeId,
		"status":       status,
	}).
		Model(&model.Notification{}).
		Omit("message").
		Order("date desc").
		Offset(skip).
		Limit(limit).
		Find(&notifications).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []model.NotificationDataModel{}, nil
		}
		return nil, err
	}

	return notifications, nil
}

//------------------------------------------
//------------------------------------------

func (n *NotificationRepository) GetUserMetaData(id int64) (*model.UserMetaDataModel, error) {
	var userDataModel model.UserMetaDataModel
	err := n.db.Where("\"userId\" = ?", id).
		Model(&model.User{}).
		Limit(1).
		Find(&userDataModel).
		Error
	if err != nil {
		return nil, err
	}

	return &userDataModel, nil
}

func (n *NotificationRepository) GetBatchUserMetaData(ids []int64) ([]model.UserMetaDataModel, error) {
	var userDataModel []model.UserMetaDataModel
	err := n.db.Where("\"userId\" IN ?", ids).
		Model(&model.User{}).
		Find(&userDataModel).
		Error
	if err != nil {
		return nil, err
	}

	return userDataModel, nil
}
