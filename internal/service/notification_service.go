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
	rabbitmq  rabbitmq.RabbitMQ
}

const (
	notificationConsumerCount = 3
)

func NewNotificationService(notifRepo repository.INotificationRepository, rabbit rabbitmq.RabbitMQ) *NotificationService {
	notifSvc := NotificationService{
		notifRepo: notifRepo,
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

	//todo : also need profileImage for push-notification
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
			receiverUser, ok := getClientFromHub(channelMessage.NotificationData.ReceiverId)
			if ok {
				notifSvc.generateNotificationMessage(channelMessage.NotificationData)
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
	users, err := n.notifRepo.GetBatchUserMetaData(userIds)
	if err == nil {
		for i := range result {
			//n.generateNotificationMessage(&result[i])
			for i2 := range users {
				if users[i2].UserId == result[i].CreatorId {
					message := fmt.Sprintf("User %v Started Following You", users[i2].Username)
					result[i].Message = message
					break
				}
			}
		}
	}
	return &result, err
}

//------------------------------------------
//------------------------------------------

func (n *NotificationService) generateNotificationMessage(notification *model.NotificationDataModel) *model.NotificationDataModel {
	switch notification.EntityTypeId {
	case 1:
		//follow notification
		creatorUser, ok := getClientFromHub(notification.CreatorId)
		if ok {
			message := fmt.Sprintf("User %v Started Following You", creatorUser.Username)
			notification.Message = message
		} else {
			user, err := n.notifRepo.GetUserMetaData(notification.CreatorId)
			if err == nil && user != nil {
				message := fmt.Sprintf("User %v Started Following You", user.Username)
				notification.Message = message
			}
		}
	case 2:
	}
	return notification
}
