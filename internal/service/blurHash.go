package service

import (
	"bytes"
	"context"
	"downloader_gochat/configs"
	"downloader_gochat/internal/repository"
	errorHandler "downloader_gochat/pkg/error"
	"downloader_gochat/rabbitmq"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"io"
	"log"
	"net/http"
	"os"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/buckket/go-blurhash"
	"github.com/disintegration/imaging"
	"github.com/getsentry/sentry-go"
	"github.com/h2non/bimg"
	amqp "github.com/rabbitmq/amqp091-go"
)

type IBlurHashService interface {
	CreateBlurHash(url string) (string, error)
}

type BlurHashService struct {
	movieRepo repository.IMovieRepository
	castRepo  repository.ICastRepository
	rabbitmq  rabbitmq.RabbitMQ
}

func NewBlurHashService(movieRepo repository.IMovieRepository, castRepo repository.ICastRepository, rabbit rabbitmq.RabbitMQ) *BlurHashService {
	blurHashSvc := BlurHashService{
		movieRepo: movieRepo,
		castRepo:  castRepo,
		rabbitmq:  rabbit,
	}

	blurHashConfig := rabbitmq.NewConfigConsume(rabbitmq.BlurHashQueue, "")
	for i := 0; i < configs.GetConfigs().BlurHashConsumerCount; i++ {
		ctx, _ := context.WithCancel(context.Background())
		go func() {
			openConChan := make(chan struct{})
			rabbitmq.NotifySetupDone(openConChan)
			<-openConChan
			if err := rabbit.Consume(ctx, blurHashConfig, &blurHashSvc, BlurHashConsumer); err != nil {
				errorMessage := fmt.Sprintf("error consuming from queue %s: %s", rabbitmq.BlurHashQueue, err)
				errorHandler.SaveError(errorMessage, err)
			}
		}()
	}

	return &blurHashSvc
}

//------------------------------------------
//------------------------------------------

type blurHashType string

const (
	staff             blurHashType = "staff"
	character         blurHashType = "character"
	moviePoster       blurHashType = "movie"
	movieS3Poster     blurHashType = "movieS3"
	movieWideS3Poster blurHashType = "movieWideS3"
)

type blurHashQueueModel struct {
	Type blurHashType `json:"type"`
	Id   string       `json:"id"`
	Url  string       `json:"url"`
}

//------------------------------------------
//------------------------------------------

func BlurHashConsumer(d *amqp.Delivery, extraConsumerData interface{}) {
	defer revive()
	// run as rabbitmq consumer
	blurSvc := extraConsumerData.(*BlurHashService)
	var channelMessage *blurHashQueueModel
	err := json.Unmarshal(d.Body, &channelMessage)
	if err != nil {
		return
	}

	hashStr := ""
	if channelMessage.Type != moviePoster {
		hashStr, err = blurSvc.CreateBlurHash(channelMessage.Url)
		if err != nil {
			if err = d.Nack(false, true); err != nil {
				errorMessage := fmt.Sprintf("error nacking [blurHash] message: %s", err)
				errorHandler.SaveError(errorMessage, err)
			}
			return
		}
	}

	switch channelMessage.Type {
	case staff, character:
		id, err := strconv.ParseInt(channelMessage.Id, 10, 64)
		if err != nil {
			if err = d.Nack(false, true); err != nil {
				errorMessage := fmt.Sprintf("error nacking [blurHash] message: %s", err)
				errorHandler.SaveError(errorMessage, err)
			}
			return
		}
		if channelMessage.Type == staff {
			err = blurSvc.castRepo.SaveStaffCastImageBlurHash(id, hashStr)
		} else {
			err = blurSvc.castRepo.SaveCharacterCastImageBlurHash(id, hashStr)
		}
		if err != nil {
			errorMessage := fmt.Sprintf("error saving [%s] blurHash: %s", channelMessage.Type, err)
			errorHandler.SaveError(errorMessage, err)
			if err = d.Nack(false, true); err != nil {
				errorMessage := fmt.Sprintf("error nacking [blurHash] message: %s", err)
				errorHandler.SaveError(errorMessage, err)
			}
			return
		}
	case moviePoster:
		posters, err := blurSvc.movieRepo.GetPosters(channelMessage.Id)
		if err != nil {
			errorMessage := fmt.Sprintf("error on getting posters to generate [blurHash]: %s", err)
			errorHandler.SaveError(errorMessage, err)
		}
		if posters != nil && len(posters) > 0 {
			for i := range posters {
				if posters[i].BlurHash == "" {
					hashStr, _ := blurSvc.CreateBlurHash(posters[i].Url)
					posters[i].BlurHash = hashStr
				}
			}

			err = blurSvc.movieRepo.SavePosters(channelMessage.Id, posters)
			if err != nil {
				errorMessage := fmt.Sprintf("error saving [%s] blurHash: %s", channelMessage.Type, err)
				errorHandler.SaveError(errorMessage, err)
				if err = d.Nack(false, true); err != nil {
					errorMessage := fmt.Sprintf("error nacking [blurHash] message: %s", err)
					errorHandler.SaveError(errorMessage, err)
				}
				return
			}
		}
	case movieS3Poster, movieWideS3Poster:
		if channelMessage.Type == movieS3Poster {
			err = blurSvc.movieRepo.SavePosterS3BlurHash(channelMessage.Id, channelMessage.Url, hashStr)
		} else {
			err = blurSvc.movieRepo.SavePosterWideS3BlurHash(channelMessage.Id, channelMessage.Url, hashStr)
		}
		if err != nil {
			errorMessage := fmt.Sprintf("error saving [%s] blurHash: %s", channelMessage.Type, err)
			errorHandler.SaveError(errorMessage, err)
			if err = d.Nack(false, true); err != nil {
				errorMessage := fmt.Sprintf("error nacking [blurHash] message: %s", err)
				errorHandler.SaveError(errorMessage, err)
			}
			return
		}
	}

	if err = d.Ack(false); err != nil {
		errorMessage := fmt.Sprintf("error acking [blurHash] message: %s", err)
		errorHandler.SaveError(errorMessage, err)
	}
}

//------------------------------------------
//------------------------------------------

func (b *BlurHashService) CreateBlurHash(url string) (string, error) {
	// download
	resp, err := http.Get(url)
	if errors.Is(err, os.ErrDeadlineExceeded) || os.IsTimeout(err) {
		return "", errors.New("image download timeout")
	}
	if err != nil {
		errorMessage := fmt.Sprintf("Error on downloading image: %s", err)
		errorHandler.SaveError(errorMessage, err)
		return "", err
	}
	if resp.StatusCode == http.StatusGatewayTimeout || resp.StatusCode == http.StatusRequestTimeout {
		return "", errors.New("image download timeout")
	}
	if resp.StatusCode == http.StatusNotFound {
		return "", errors.New("image not found")
	}
	if resp.StatusCode != http.StatusOK {
		errorMessage := fmt.Sprintf("Error on downloading image: %v", fmt.Errorf("bad status: %s", resp.Status))
		errorHandler.SaveError(errorMessage, err)
		return "", err
	}
	defer resp.Body.Close()

	// decode
	var img image.Image
	if strings.HasSuffix(url, ".png") {
		img, err = imaging.Decode(resp.Body)
		if err != nil {
			errorMessage := fmt.Sprintf("Error on decoding downloaded image: %v", err)
			errorHandler.SaveError(errorMessage, err)
			return "", err
		}
	} else {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			errorMessage := fmt.Sprintf("Error reading response body: %v", err)
			errorHandler.SaveError(errorMessage, err)
			return "", err
		}
		// Create a new reader from the body content
		reader := bytes.NewReader(body)

		img, err = imaging.Decode(reader)
		if err != nil {
			if err.Error() == "unsupported JPEG feature: luma/chroma subsampling ratio" {
				// https://static-cdn.sr.se/images/2519/4ed3bde6-46a3-469b-af8e-5c2bde6d1749.jpg?preset=256x256
				//img2, err := bimg.NewImage(body).Convert(bimg.PNG)
				img2, err := bimg.NewImage(body).Convert(bimg.JPEG)
				if err != nil {
					errorMessage := fmt.Sprintf("Error on coverting image to png: %v", err)
					errorHandler.SaveError(errorMessage, err)
					return "", err
				}
				img, err = imaging.Decode(bytes.NewReader(img2))
				if err != nil {
					errorMessage := fmt.Sprintf("Error on decoding coverted image: %v", err)
					errorHandler.SaveError(errorMessage, err)
					return "", err
				}
			} else {
				errorMessage := fmt.Sprintf("Error on decoding downloaded image: %v", err)
				errorHandler.SaveError(errorMessage, err)
				return "", err
			}
		}
	}

	//creating blurHash
	str, err := blurhash.Encode(4, 3, img)
	if err != nil {
		errorMessage := fmt.Sprintf("Error on creating blurHash from downloaded image: %v", err)
		errorHandler.SaveError(errorMessage, err)
		return "", err
	}
	return str, nil

}

//------------------------------------------
//------------------------------------------

func revive() {
	if err := recover(); err != nil {
		sentry.CurrentHub().Recover(err)
		if os.Getenv("LOG_PANIC_TRACE") == "true" {
			log.Println(
				"level:", "error",
				"err: ", err,
				"trace", string(debug.Stack()),
			)
		} else {
			log.Println(
				"level", "error",
				"err", err,
			)
		}
	}
}
