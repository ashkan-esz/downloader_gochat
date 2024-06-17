package service

import (
	"context"
	"downloader_gochat/internal/repository"
	"downloader_gochat/model"
	errorHandler "downloader_gochat/pkg/error"
	"downloader_gochat/rabbitmq"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strconv"

	amqp "github.com/rabbitmq/amqp091-go"
	"gorm.io/gorm"
)

type INotificationService interface {
	GetUserNotifications(userId int64, skip int, limit int, entityTypeId int, status int, autoUpdateStatus bool) (*[]model.NotificationDataModel, error)
	BatchUpdateUserNotificationStatus(userId int64, id int64, entityTypeId int, status int) error
}

type NotificationService struct {
	notifRepo    repository.INotificationRepository
	userRep      repository.IUserRepository
	movieRep     repository.IMovieRepository
	rabbitmq     rabbitmq.RabbitMQ
	pushNotifSvc IPushNotificationService
}

const (
	notificationConsumerCount = 4
)

func NewNotificationService(notifRepo repository.INotificationRepository, userRep repository.IUserRepository, movieRep repository.IMovieRepository, rabbit rabbitmq.RabbitMQ, pushNotifSvc IPushNotificationService) *NotificationService {
	notifSvc := NotificationService{
		notifRepo:    notifRepo,
		userRep:      userRep,
		movieRep:     movieRep,
		rabbitmq:     rabbit,
		pushNotifSvc: pushNotifSvc,
	}

	notificationConfig := rabbitmq.NewConfigConsume(rabbitmq.NotificationQueue, "")
	for i := 0; i < notificationConsumerCount; i++ {
		ctx, _ := context.WithCancel(context.Background())
		go func() {
			openConChan := make(chan struct{})
			rabbitmq.NotifySetupDone(openConChan)
			<-openConChan
			if err := rabbit.Consume(ctx, notificationConfig, &notifSvc, NotificationConsumer); err != nil {
				errorMessage := fmt.Sprintf("error consuming from queue %s: %s", rabbitmq.NotificationQueue, err)
				errorHandler.SaveError(errorMessage, err)
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
	case model.FollowNotifAction, model.MovieNotifAction:
		// need to save the notification, show notification in app, send push-notification to followed user
		err = notifSvc.notifRepo.SaveUserNotification(channelMessage.NotificationData)
		if err != nil && !errors.Is(err, gorm.ErrDuplicatedKey) {
			if err = d.Nack(false, true); err != nil {
				errorMessage := fmt.Sprintf("error nacking [notification] message: %s", err)
				errorHandler.SaveError(errorMessage, err)
			}
			return
		} else {
			notifSvc.handleNotification(channelMessage.NotificationData)

			receiverUser, ok := getClientFromHub(channelMessage.NotificationData.ReceiverId)
			if ok {
				receiverUser.Message <- channelMessage
			}
		}
	case model.NewMessageNotifAction:
		// don't need to save this notification, show notification in app, send push-notification (only if user is offline)
		// in app notification in handled by newMessage action, just send push-notification
		_, ok := getClientFromHub(channelMessage.NotificationData.ReceiverId)
		if !ok {
			notifSvc.handleNotification(channelMessage.NotificationData)
		}
	}

	if err = d.Ack(false); err != nil {
		errorMessage := fmt.Sprintf("error acking [notification] message: %s", err)
		errorHandler.SaveError(errorMessage, err)
	}
}

//------------------------------------------
//------------------------------------------

func (n *NotificationService) GetUserNotifications(userId int64, skip int, limit int, entityTypeId int, status int, autoUpdateStatus bool) (*[]model.NotificationDataModel, error) {
	result, err := n.notifRepo.GetUserNotifications(userId, skip, limit, entityTypeId, status)
	if err != nil {
		return nil, err
	}

	userIds := []int64{}
	movieIds := []string{}
	for i := range result {
		if result[i].EntityTypeId == model.FollowNotificationTypeId {
			userIds = append(userIds, result[i].CreatorId)
		} else if result[i].EntityTypeId == model.MoviesNotificationTypeId {
			movieIds = append(movieIds, result[i].EntityId)
		}
	}
	userIds = slices.Compact(userIds)
	movieIds = slices.Compact(movieIds)

	//--------------------------------------

	// follow,new_message notification
	if len(userIds) > 0 {
		misCacheUserIds := []int64{}

		cachedData, _ := getCachedMultiUserData(userIds)
		if cachedData != nil && len(cachedData) > 0 {
			for i := range result {
				found := false
				for i2 := range cachedData {
					if cachedData[i2].UserId == result[i].CreatorId {
						result[i].Message = generateNotificationMessage(&result[i], cachedData[i2].Username)
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
			if err != nil {
				return nil, err
			}

			if users != nil {
				for i := range result {
					if result[i].Message == "" {
						for i2 := range users {
							if users[i2].UserId == result[i].CreatorId {
								result[i].Message = generateNotificationMessage(&result[i], users[i2].Username)
								result[i].CreatorImage = addCreatorImageToNotification(users[i2].ProfileImages)
								break
							}
						}
					}
				}
			}
		}
	}

	//--------------------------------------

	// movies notification
	if len(movieIds) > 0 {
		misCacheMovieIds := []string{}

		cachedData, _ := getCachedMultiMovieData(movieIds)
		if cachedData != nil && len(cachedData) > 0 {
			for i := range result {
				found := false
				for i2 := range cachedData {
					if cachedData[i2].MovieId == result[i].EntityId {
						result[i].Message = generateNotificationMessage(&result[i], cachedData[i2].RawTitle)
						result[i].CreatorImage = addMoviePosterToNotification(cachedData[i2].Posters)
						found = true
						break
					}
				}
				if !found {
					misCacheMovieIds = append(misCacheMovieIds, result[i].EntityId)
				}
			}
			misCacheMovieIds = slices.Compact(misCacheMovieIds)
		} else {
			misCacheMovieIds = movieIds
		}

		if len(misCacheMovieIds) > 0 {
			movies, err := n.movieRep.GetBatchMovieBriefData(misCacheMovieIds)
			if err != nil {
				return nil, err
			}

			if movies != nil {
				for i := range result {
					for i2 := range movies {
						if movies[i2].MovieId == result[i].EntityId {
							result[i].Message = generateNotificationMessage(&result[i], movies[i2].RawTitle)
							result[i].CreatorImage = addMoviePosterToNotification(movies[i2].Posters)
							break
						}
					}
				}
			}
		}
	}

	//--------------------------------------

	if len(result) > 0 && skip == 0 && limit > 0 && entityTypeId != 0 && status == 0 && autoUpdateStatus {
		_ = n.notifRepo.BatchUpdateNotificationStatusByDate(result[0].Date, userId, entityTypeId, 2)
	}
	return &result, nil
}

func (n *NotificationService) BatchUpdateUserNotificationStatus(userId int64, id int64, entityTypeId int, status int) error {
	err := n.notifRepo.BatchUpdateNotificationStatusById(userId, id, entityTypeId, status)
	return err
}

//------------------------------------------
//------------------------------------------

func (n *NotificationService) handleNotification(notificationData *model.NotificationDataModel) {
	if notificationData.EntityTypeId == model.MoviesNotificationTypeId {
		cacheData, _ := getCachedMovieData(notificationData.EntityId)
		if cacheData != nil {
			notificationData.Message = generateNotificationMessage(notificationData, cacheData.RawTitle)
			notificationData.CreatorImage = addMoviePosterToNotification(cacheData.Posters)
		} else {
			movieData, err := n.movieRep.GetMovieBriefData(notificationData.EntityId)
			if err == nil && movieData != nil {
				notificationData.Message = generateNotificationMessage(notificationData, movieData.RawTitle)
				notificationData.CreatorImage = addMoviePosterToNotification(movieData.Posters)
				movieCacheData := &model.CachedMovieData{
					MovieId:  movieData.MovieId,
					RawTitle: movieData.RawTitle,
					Type:     movieData.Type,
					Year:     movieData.Year,
					Posters:  movieData.Posters,
				}
				_ = setMovieDataCache(notificationData.EntityId, movieCacheData)
			}
		}
	} else {
		cacheData, _ := getCachedUserData(notificationData.CreatorId)
		if cacheData != nil {
			notificationData.Message = generateNotificationMessage(notificationData, cacheData.Username)
			notificationData.CreatorImage = addCreatorImageToNotification(cacheData.ProfileImages)
		} else {
			userData, err := n.notifRepo.GetUserMetaDataWithImage(notificationData.CreatorId, 1)
			if err == nil && userData != nil {
				notificationData.Message = generateNotificationMessage(notificationData, userData.Username)
				notificationData.CreatorImage = addCreatorImageToNotification(userData.ProfileImages)
			}
		}
	}

	pushNotificationTitle := ""
	switch notificationData.EntityTypeId {
	case model.FollowNotificationTypeId:
		pushNotificationTitle = "New Follower"
	case model.NewMessageNotificationTypeId:
		pushNotificationTitle = "New Message"
	case model.MoviesNotificationTypeId:
		pushNotificationTitle = "Movie Update"
	}

	receiverCacheData, _ := getCachedUserData(notificationData.ReceiverId)
	if receiverCacheData != nil {
		if (notificationData.EntityTypeId == model.FollowNotificationTypeId && !receiverCacheData.NotificationSettings.NewFollower) ||
			(notificationData.EntityTypeId == model.NewMessageNotificationTypeId && !receiverCacheData.NotificationSettings.NewMessage) ||
			checkMoviePushNotifDisabled(notificationData, receiverCacheData.NotificationSettings) {
			//push-notification is disabled
			return
		}
		for i := range receiverCacheData.NotifTokens {
			if receiverCacheData.NotifTokens[i] != "" {
				n.pushNotifSvc.AddPushNotificationToBuffer(
					receiverCacheData.NotifTokens[i],
					pushNotificationTitle,
					notificationData.Message,
					notificationData.CreatorImage,
					strconv.FormatInt(notificationData.CreatorId, 10),
				)
			}
		}
	} else {
		receiverUserData, err := n.userRep.GetUserMetaDataAndNotificationSettings(notificationData.ReceiverId, 1)
		if err == nil && receiverUserData != nil {
			if (notificationData.EntityTypeId == model.FollowNotificationTypeId && !receiverUserData.NewFollower) ||
				(notificationData.EntityTypeId == model.NewMessageNotificationTypeId && !receiverUserData.NewMessage) ||
				checkMoviePushNotifDisabled2(notificationData, receiverUserData) {
				//push-notification is disabled
				return
			}
			for i := range receiverUserData.ActiveSessions {
				if receiverUserData.ActiveSessions[i].NotifToken != "" {
					n.pushNotifSvc.AddPushNotificationToBuffer(
						receiverUserData.ActiveSessions[i].NotifToken,
						pushNotificationTitle,
						notificationData.Message,
						notificationData.CreatorImage,
						strconv.FormatInt(notificationData.CreatorId, 10),
					)
				}
			}
		}
	}
}

func generateNotificationMessage(notificationData *model.NotificationDataModel, username string) string {
	message := ""
	switch notificationData.EntityTypeId {
	case model.FollowNotificationTypeId:
		//new follower
		message = fmt.Sprintf("%v Started Following You", username)
	case model.NewMessageNotificationTypeId:
		//new message
		message = fmt.Sprintf("%v: %v", username, notificationData.Message)
	case model.MoviesNotificationTypeId:
		switch notificationData.SubEntityTypeId {
		case model.FinishedListSpinOffSequel:
			message = fmt.Sprintf("%v: released (related to your watched-list)", username)
		case model.FollowingMovie:
			if notificationData.Message != "" {
				message = fmt.Sprintf("%v: updated (%v)", username, notificationData.Message)
			} else {
				message = fmt.Sprintf("%v: updated", username)
			}
		case model.FollowMovieBetterQuality:
			if notificationData.Message != "" {
				message = fmt.Sprintf("%v: updated (%v)", username, notificationData.Message)
			} else {
				message = fmt.Sprintf("%v: better quality released", username)
			}
		case model.FollowMovieSubtitle:
			if notificationData.Message != "" {
				message = fmt.Sprintf("%v: subtitle released (%v)", username, notificationData.Message)
			} else {
				message = fmt.Sprintf("%v: subtitle released", username)
			}
		case model.FutureList:
			if notificationData.Message != "" {
				message = fmt.Sprintf("%v (from watch-list): updated (%v)", username, notificationData.Message)
			} else {
				message = fmt.Sprintf("%v (from watch-list): updated", username)
			}
		case model.FutureListSerialSeasonEnd:
			if notificationData.Message != "" {
				message = fmt.Sprintf("%v (from watch-list): season ended (%v)", username, notificationData.Message)
			} else {
				message = fmt.Sprintf("%v (from watch-list): season ended, time to watch", username)
			}
		case model.FutureListSubtitle:
			if notificationData.Message != "" {
				message = fmt.Sprintf("%v (from watch-list): subtitle released (%v)", username, notificationData.Message)
			} else {
				message = fmt.Sprintf("%v (from watch-list): subtitle released", username)
			}
		}
	}
	return message
}

func addCreatorImageToNotification(profileImages []model.ProfileImageDataModel) string {
	if len(profileImages) > 0 {
		if profileImages[0].Url != "" {
			return profileImages[0].Url
		} else if profileImages[0].Thumbnail != "" {
			return profileImages[0].Thumbnail
		}
	}
	return ""
}

func addMoviePosterToNotification(posters []model.MoviePoster) string {
	if len(posters) > 0 {
		if posters[0].Url != "" {
			return posters[0].Url
		} else if posters[0].Thumbnail != "" {
			return posters[0].Thumbnail
		}
	}
	return ""
}

func checkMoviePushNotifDisabled(notificationData *model.NotificationDataModel, notificationSettings model.NotificationSettings) bool {
	if notificationData.EntityTypeId != model.MoviesNotificationTypeId {
		return false
	}

	return (notificationData.SubEntityTypeId == model.FinishedListSpinOffSequel && !notificationSettings.FutureListSerialSeasonEnd) ||
		(notificationData.SubEntityTypeId == model.FollowingMovie && !notificationSettings.FollowMovie) ||
		(notificationData.SubEntityTypeId == model.FollowMovieBetterQuality && !notificationSettings.FollowMovieBetterQuality) ||
		(notificationData.SubEntityTypeId == model.FollowMovieSubtitle && !notificationSettings.FollowMovieSubtitle) ||
		(notificationData.SubEntityTypeId == model.FutureList && !notificationSettings.FutureList) ||
		(notificationData.SubEntityTypeId == model.FutureListSerialSeasonEnd && !notificationSettings.FutureListSerialSeasonEnd) ||
		(notificationData.SubEntityTypeId == model.FutureListSubtitle && !notificationSettings.FutureListSubtitle)
}

func checkMoviePushNotifDisabled2(notificationData *model.NotificationDataModel, notificationSettings *model.UserMetaWithNotificationSettings) bool {
	if notificationData.EntityTypeId != model.MoviesNotificationTypeId {
		return false
	}

	return (notificationData.SubEntityTypeId == model.FinishedListSpinOffSequel && !notificationSettings.FutureListSerialSeasonEnd) ||
		(notificationData.SubEntityTypeId == model.FollowingMovie && !notificationSettings.FollowMovie) ||
		(notificationData.SubEntityTypeId == model.FollowMovieBetterQuality && !notificationSettings.FollowMovieBetterQuality) ||
		(notificationData.SubEntityTypeId == model.FollowMovieSubtitle && !notificationSettings.FollowMovieSubtitle) ||
		(notificationData.SubEntityTypeId == model.FutureList && !notificationSettings.FutureList) ||
		(notificationData.SubEntityTypeId == model.FutureListSerialSeasonEnd && !notificationSettings.FutureListSerialSeasonEnd) ||
		(notificationData.SubEntityTypeId == model.FutureListSubtitle && !notificationSettings.FutureListSubtitle)
}
