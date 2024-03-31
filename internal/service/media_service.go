package service

import (
	"bytes"
	"context"
	"downloader_gochat/cloudStorage"
	"downloader_gochat/internal/repository"
	"downloader_gochat/model"
	errorHandler "downloader_gochat/pkg/error"
	"downloader_gochat/rabbitmq"
	"encoding/base64"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/buckket/go-blurhash"
	"github.com/disintegration/imaging"
	"github.com/google/uuid"
	"github.com/kolesa-team/go-webp/encoder"
	"github.com/kolesa-team/go-webp/webp"
)

// sudo apt-get install libwebp-dev

type IMediaService interface {
	UploadFile(userId int64, messageData *model.UploadMediaReq, contentType string, fileSize int64, fileName string, fileBuffer multipart.File) (*model.MediaFile, error)
}

type MediaService struct {
	mediaRepo    repository.IMediaRepository
	userRep      repository.IUserRepository
	wsRep        repository.IWsRepository
	rabbitmq     rabbitmq.RabbitMQ
	cloudStorage cloudStorage.IS3Storage
}

func NewMediaService(mediaRepo repository.IMediaRepository, userRep repository.IUserRepository, wsRep repository.IWsRepository, rabbit rabbitmq.RabbitMQ, cloudStorage cloudStorage.IS3Storage) *MediaService {
	return &MediaService{
		mediaRepo:    mediaRepo,
		userRep:      userRep,
		wsRep:        wsRep,
		rabbitmq:     rabbit,
		cloudStorage: cloudStorage,
	}
}

//------------------------------------------
//------------------------------------------

func (m *MediaService) UploadFile(userId int64, messageData *model.UploadMediaReq, contentType string, fileSize int64, fileName string, fileBuffer multipart.File) (*model.MediaFile, error) {
	newMessage := model.ReceiveNewMessage{
		Uuid:       messageData.Uuid,
		Content:    messageData.Content,
		RoomId:     messageData.RoomId,
		ReceiverId: messageData.ReceiverId,
		State:      1,
		Date:       time.Now().UTC(),
		UserId:     userId,
	}
	messageId, err := m.wsRep.SaveMessage(&newMessage)
	newMessage.Id = messageId
	if err != nil {
		return nil, err
	}

	savingFileName := uuid.NewString() + filepath.Ext(fileName)
	result, err := m.cloudStorage.UploadLargeFile(cloudStorage.MediaFileBucketName, savingFileName, fileBuffer)
	if err != nil {
		return nil, err
	}

	mediaFile := model.MediaFile{
		Id:        0,
		MessageId: messageId,
		Date:      time.Now().UTC(),
		Url:       result.Location,
		Type:      contentType,
		Size:      fileSize,
		Thumbnail: "",
		BlurHash:  "",
	}
	if strings.Contains(contentType, "image") {
		mediaFile.Thumbnail, mediaFile.BlurHash = createThumbnailAndBlurHash(contentType, fileBuffer)
	}

	err = m.mediaRepo.SaveMediaData(&mediaFile)
	if err != nil {
		return nil, err
	}

	newMessage.Medias = []model.MediaFile{mediaFile}

	//----------------------------------------------------
	//----------------------------------------------------
	sender, senderExist := getClientFromHub(userId)
	cl, ok := getClientFromHub(messageData.ReceiverId)
	if ok {
		// receiver is online
		// add creator profileImage, read from cache only
		userCacheData, _ := getCachedUserData(newMessage.UserId)
		if userCacheData != nil {
			if len(userCacheData.ProfileImages) > 0 {
				newMessage.CreatorImage = userCacheData.ProfileImages[0].Thumbnail
			}
			newMessage.Username = userCacheData.Username
		}

		receiveMessage := model.CreateReceiveNewMessageAction(&newMessage)
		cl.Message <- receiveMessage
	}

	if senderExist {
		messageSendResult := model.CreateNewMessageSendResult(
			newMessage.Id,
			newMessage.Uuid,
			newMessage.RoomId,
			newMessage.ReceiverId,
			newMessage.Date,
			newMessage.State,
			200, "")
		sender.Message <- messageSendResult
	}

	if !ok {
		// receiver is offline
		// don't need to save this notification, show notification in app, send push-notification (only if user is offline)
		ctx, _ := context.WithCancel(context.Background())
		//defer cancel()
		notifQueueConf := rabbitmq.NewConfigPublish(rabbitmq.NotificationExchange, rabbitmq.NotificationBindingKey)
		notifMessage := model.CreateNewMessageNotificationAction(&newMessage)
		m.rabbitmq.Publish(ctx, notifMessage, notifQueueConf, newMessage.ReceiverId)
	}
	_ = m.wsRep.UpdateUserReceivedMessageTime(newMessage.ReceiverId)

	return &mediaFile, err
}

//------------------------------------------
//------------------------------------------

func createThumbnailAndBlurHash(contentType string, fileBuffer multipart.File) (string, string) {
	img, err := imaging.Decode(fileBuffer)
	if err != nil {
		errorMessage := fmt.Sprintf("Error on decoding uploaded image: %v", err)
		errorHandler.SaveError(errorMessage, err)
		return "", ""
	}

	thumbnailStr := ""
	dstImage := imaging.Thumbnail(img, 128, 128, imaging.Linear)

	switch contentType {
	case "image/jpeg", "image/jpg", "image/png":
		newbuf := new(bytes.Buffer)
		//err = jpeg.Encode(newbuf, dstImage, &jpeg.Options{Quality: 20})
		options, err := encoder.NewLossyEncoderOptions(encoder.PresetDefault, 30)
		if err != nil {
			errorMessage := fmt.Sprintf("Error on creating webp options: %v", err)
			errorHandler.SaveError(errorMessage, err)
		} else {
			err = webp.Encode(newbuf, dstImage, options)
			if err != nil {
				errorMessage := fmt.Sprintf("Error on encoding webp: %v", err)
				errorHandler.SaveError(errorMessage, err)
			} else {
				str := base64.StdEncoding.EncodeToString(newbuf.Bytes())
				//thumbnailStr = "data:image/jpeg;base64," + str
				thumbnailStr = "data:image/webp;base64," + str
			}
		}
	}

	if thumbnailStr != "" {
		//creating blurHash from thumbnail image
		str, err := blurhash.Encode(4, 3, dstImage)
		if err != nil {
			errorMessage := fmt.Sprintf("Error on creating blurHash from thumbnail image: %v", err)
			errorHandler.SaveError(errorMessage, err)
			return thumbnailStr, ""
		}
		return thumbnailStr, str
	} else {
		//creating blurHash from original image
		str, err := blurhash.Encode(4, 3, img)
		if err != nil {
			errorMessage := fmt.Sprintf("Error on creating blurHash from original image: %v", err)
			errorHandler.SaveError(errorMessage, err)
			return thumbnailStr, ""
		}
		return thumbnailStr, str
	}
}
