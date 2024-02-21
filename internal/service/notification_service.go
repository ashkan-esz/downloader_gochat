package service

import (
	"context"
	"downloader_gochat/internal/repository"
	"downloader_gochat/model"
	"downloader_gochat/rabbitmq"
	"encoding/json"
	"fmt"
	"log"
	"slices"

	amqp "github.com/rabbitmq/amqp091-go"
)

type INotificationService interface {
	GetUserNotifications(userId int64, skip int, limit int, entityTypeId int, status int) (*[]model.NotificationDataModel, error)
}

type NotificationService struct {
	notifRepo repository.INotificationRepository
	userRep   repository.IUserRepository
	rabbitmq  rabbitmq.RabbitMQ
}

const (
	notificationConsumerCount = 3
)

func NewNotificationService(notifRepo repository.INotificationRepository, userRep repository.IUserRepository, rabbit rabbitmq.RabbitMQ) *NotificationService {
	notifSvc := NotificationService{
		notifRepo: notifRepo,
		userRep:   userRep,
		rabbitmq:  rabbit,
	}

	notificationConfig := rabbitmq.NewConfigConsume(rabbitmq.NotificationQueue, "")
	for i := 0; i < notificationConsumerCount; i++ {
		ctx, _ := context.WithCancel(context.Background())
		go func() {
			openConChan := make(chan struct{})
			rabbitmq.NotifySetupDone(openConChan)
			<-openConChan
			if err := rabbit.Consume(ctx, notificationConfig, &notifSvc, NotificationConsumer); err != nil {
				log.Printf("error consuming from queue %s: %s\n", rabbitmq.NotificationQueue, err)
			}
		}()
	}

	return &notifSvc
}

//------------------------------------------
//------------------------------------------

func NotificationConsumer(d *amqp.Delivery, extraConsumerData interface{}) {
	// run as rabbitmq consumer
	notifSvc := extraConsumerData.(*NotificationService)
	var channelMessage *model.ChannelMessage
	err := json.Unmarshal(d.Body, &channelMessage)
	if err != nil {
		return
	}

	switch channelMessage.Action {
	case model.FollowNotifAction:
		// need to save the notification, show notification in app, send push-notification to followed user
		err = notifSvc.notifRepo.SaveUserNotification(channelMessage.NotificationData)
		if err != nil {
			if err = d.Nack(false, true); err != nil {
				log.Printf("error nacking [notification] message: %s\n", err)
			}
		} else {
			//todo : send push-notification to channelMessage.NotificationData.ReceiverId
			notifSvc.handleNotificationMessageAndImage(channelMessage.NotificationData)
			receiverUser, ok := getClientFromHub(channelMessage.NotificationData.ReceiverId)
			if ok {
				receiverUser.Message <- channelMessage
			}
		}
	case model.NewMessageNotifAction:
		// don't need to save this notification, show notification in app, send push-notification (only if user is offline)
		// in app notification in handled by newMessage action, just send push-notification
		//todo : send push-notification to channelMessage.NotificationData.ReceiverId
	}

	if err = d.Ack(false); err != nil {
		log.Printf("error acking [notification] message: %s\n", err)
	}
}

//------------------------------------------
//------------------------------------------

func (n *NotificationService) GetUserNotifications(userId int64, skip int, limit int, entityTypeId int, status int) (*[]model.NotificationDataModel, error) {
	result, err := n.notifRepo.GetUserNotifications(userId, skip, limit, entityTypeId, status)
	userIds := []int64{}
	for i := range result {
		if result[i].EntityTypeId == model.FollowNotificationTypeId {
			userIds = append(userIds, result[i].CreatorId)
		}
	}
	userIds = slices.Compact(userIds)
	misCacheUserIds := []int64{}

	cachedData, _ := getCachedMultiUserData(userIds)
	if cachedData != nil && len(cachedData) > 0 {
		for i := range result {
			found := false
			for i2 := range cachedData {
				if cachedData[i2].UserId == result[i].CreatorId {
					result[i].Message = generateNotificationMessage(result[i].EntityTypeId, cachedData[i2].Username)
					result[i].CreatorImage = addCreatorImageToNotification(cachedData[i2].ProfileImages)
					found = true
					break
				}
			}
			if !found {
				misCacheUserIds = append(misCacheUserIds, result[i].CreatorId)
			}
		}
		misCacheUserIds = slices.Compact(misCacheUserIds)
	} else {
		misCacheUserIds = userIds
	}

	if len(misCacheUserIds) > 0 {
		users, err := n.notifRepo.GetBatchUserMetaDataWithImage(misCacheUserIds)
		if users != nil {
			for i := range result {
				if result[i].Message == "" {
					for i2 := range users {
						if users[i2].UserId == result[i].CreatorId {
							result[i].Message = generateNotificationMessage(result[i].EntityTypeId, users[i2].Username)
							result[i].CreatorImage = addCreatorImageToNotification(users[i2].ProfileImages)
							break
						}
					}
				}
			}
		}
		return &result, err
	}

	return &result, err
}

//------------------------------------------
//------------------------------------------

func (n *NotificationService) handleNotificationMessageAndImage(notification *model.NotificationDataModel) *model.NotificationDataModel {
	switch notification.EntityTypeId {
	case 1:
		//follow notification
		cacheData, _ := getCachedUserData(notification.CreatorId)
		if cacheData != nil {
			notification.Message = generateNotificationMessage(notification.EntityTypeId, cacheData.Username)
			notification.CreatorImage = addCreatorImageToNotification(cacheData.ProfileImages)
		} else {
			user, err := n.notifRepo.GetUserMetaDataWithImage(notification.CreatorId, 1)
			if err == nil && user != nil {
				notification.Message = generateNotificationMessage(notification.EntityTypeId, user.Username)
				notification.CreatorImage = addCreatorImageToNotification(user.ProfileImages)
			}
		}
	case 2:
	}
	return notification
}

func generateNotificationMessage(entityTypeId int, username string) string {
	message := ""
	switch entityTypeId {
	case 1:
		//new follower
		message = fmt.Sprintf("User %v Started Following You", username)
	case 2:
		//new message
	}
	return message
}

func addCreatorImageToNotification(profileImages []model.ProfileImage) string {
	if len(profileImages) > 0 {
		if profileImages[0].Thumbnail != "" {
			return profileImages[0].Thumbnail
		} else {
			return profileImages[0].Url
		}
	}
	return ""
}
