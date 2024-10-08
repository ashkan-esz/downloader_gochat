package repository

import (
	"downloader_gochat/model"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

type INotificationRepository interface {
	SaveUserNotification(notifData *model.NotificationDataModel) error
	GetUserNotifications(userId int64, skip int, limit int, entityTypeId int, status int) ([]model.NotificationDataModel, error)
	GetUserMetaData(id int64) (*model.UserMetaDataModel, error)
	GetBatchUserMetaData(ids []int64) ([]model.UserMetaDataModel, error)
	GetUserMetaDataWithImage(id int64, imageLimit int) (*model.UserMetaWithImageDataModel, error)
	GetBatchUserMetaDataWithImage(ids []int64) ([]model.UserMetaWithImageDataModel, error)
	BatchUpdateNotificationStatusByDate(date time.Time, receiverId int64, entityTypeId int, status int) error
	BatchUpdateNotificationStatusById(receiverId int64, nid int64, entityTypeId int, status int) error
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

func (n *NotificationRepository) SaveUserNotification(notifData *model.NotificationDataModel) error {
	notif := model.Notification{
		CreatorId:       notifData.CreatorId,
		ReceiverId:      notifData.ReceiverId,
		Date:            notifData.Date,
		Status:          notifData.Status,
		Message:         notifData.Message,
		EntityId:        notifData.EntityId,
		EntityTypeId:    notifData.EntityTypeId,
		SubEntityTypeId: notifData.SubEntityTypeId,
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
		Omit("creatorImage").
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

func (n *NotificationRepository) BatchUpdateNotificationStatusByDate(date time.Time, receiverId int64, entityTypeId int, status int) error {
	result := n.db.
		Model(&model.Notification{}).
		Where("\"receiverId\" = ? and \"entityTypeId\" = ? and \"date\" <= (?) ", receiverId, entityTypeId, date).
		UpdateColumn("\"status\"", status)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil
		}
		return result.Error
	}
	return nil
}

func (n *NotificationRepository) BatchUpdateNotificationStatusById(receiverId int64, nid int64, entityTypeId int, status int) error {
	var result *gorm.DB
	if entityTypeId == 0 {
		subQuery := n.db.
			Model(&model.Notification{}).
			Where("\"id\" = ? and \"receiverId\" = ? and status < ?", nid, receiverId, status).
			Select("date")

		result = n.db.
			Model(&model.Notification{}).
			Where("\"receiverId\" = ? and \"date\" <= (?) ", receiverId, subQuery).
			UpdateColumn("\"status\"", status)
	} else {
		subQuery := n.db.
			Model(&model.Notification{}).
			Where("\"id\" = ? and \"receiverId\" = ? and \"entityTypeId\" = ? and status < ?", nid, receiverId, entityTypeId, status).
			Select("date")

		result = n.db.
			Model(&model.Notification{}).
			Where("\"receiverId\" = ? and \"entityTypeId\" = ? and \"date\" <= (?) ", receiverId, entityTypeId, subQuery).
			UpdateColumn("\"status\"", status)
	}

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

func (n *NotificationRepository) GetUserMetaDataWithImage(id int64, imageLimit int) (*model.UserMetaWithImageDataModel, error) {
	var result model.UserMetaWithImageDataModel
	err := n.db.
		Model(&model.User{}).
		Where("\"userId\" = ?", id).
		Limit(1).
		Preload("ProfileImages", func(db *gorm.DB) *gorm.DB {
			return db.Order("\"addDate\" DESC").Limit(imageLimit)
		}).
		Find(&result).
		Error

	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (n *NotificationRepository) GetBatchUserMetaDataWithImage(ids []int64) ([]model.UserMetaWithImageDataModel, error) {
	var userDataModel []model.UserMetaWithImageDataModel
	err := n.db.Where("\"userId\" IN ?", ids).
		Model(&model.User{}).
		Preload("ProfileImages", func(db *gorm.DB) *gorm.DB {
			return db.Order("\"addDate\" DESC")
		}).
		Find(&userDataModel).
		Error
	if err != nil {
		return nil, err
	}

	return userDataModel, nil
}
