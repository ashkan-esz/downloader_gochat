package service

import (
	"context"
	"downloader_gochat/internal/repository"
	"downloader_gochat/rabbitmq"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime/debug"
	"strconv"

	"github.com/buckket/go-blurhash"
	"github.com/disintegration/imaging"
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
	for i := 0; i < blurHashConsumerCount; i++ {
		ctx, _ := context.WithCancel(context.Background())
		go func() {
			openConChan := make(chan struct{})
			rabbitmq.NotifySetupDone(openConChan)
			<-openConChan
			if err := rabbit.Consume(ctx, blurHashConfig, &blurHashSvc, BlurHashConsumer); err != nil {
				log.Printf("error consuming from queue %s: %s\n", rabbitmq.BlurHashQueue, err)
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
				log.Printf("error nacking [blurHash] message: %s\n", err)
			}
			return
		}
	}

	switch channelMessage.Type {
	case staff, character:
		id, err := strconv.ParseInt(channelMessage.Id, 10, 64)
		if err != nil {
			if err = d.Nack(false, true); err != nil {
				log.Printf("error nacking [blurHash] message: %s\n", err)
			}
			return
		}
		if channelMessage.Type == staff {
			err = blurSvc.castRepo.SaveStaffCastImageBlurHash(id, hashStr)
		} else {
			err = blurSvc.castRepo.SaveCharacterCastImageBlurHash(id, hashStr)
		}
		if err != nil {
			log.Printf("error saving [%s] blurHash: %s\n", channelMessage.Type, err)
			if err = d.Nack(false, true); err != nil {
				log.Printf("error nacking [blurHash] message: %s\n", err)
			}
		}
	case moviePoster:
		posters, err := blurSvc.movieRepo.GetPosters(channelMessage.Id)
		if err != nil {
			log.Printf("error on getting posters to generate [blurHash]: %s\n", err)
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
				log.Printf("error saving [%s] blurHash: %s\n", channelMessage.Type, err)
				if err = d.Nack(false, true); err != nil {
					log.Printf("error nacking [blurHash] message: %s\n", err)
				}
			}
		}
	case movieS3Poster, movieWideS3Poster:
		if channelMessage.Type == movieS3Poster {
			err = blurSvc.movieRepo.SavePosterS3BlurHash(channelMessage.Id, channelMessage.Url, hashStr)
		} else {
			err = blurSvc.movieRepo.SavePosterWideS3BlurHash(channelMessage.Id, channelMessage.Url, hashStr)
		}
		if err != nil {
			log.Printf("error saving [%s] blurHash: %s\n", channelMessage.Type, err)
			if err = d.Nack(false, true); err != nil {
				log.Printf("error nacking [blurHash] message: %s\n", err)
			}
		}
	}

	if err = d.Ack(false); err != nil {
		log.Printf("error acking [blurHash] message: %s\n", err)
	}
}

//------------------------------------------
//------------------------------------------

func (b *BlurHashService) CreateBlurHash(url string) (string, error) {
	// download
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error on downloading image: ", err)
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Println("Error on downloading image: ", fmt.Errorf("bad status: %s", resp.Status))
		return "", err
	}
	defer resp.Body.Close()

	// decode
	img, err := imaging.Decode(resp.Body)
	if err != nil {
		fmt.Println("Error on decoding downloaded image: ", err)
		return "", err
	}

	//creating blurHash
	str, err := blurhash.Encode(4, 3, img)
	if err != nil {
		fmt.Println("Error on creating blurHash from downloaded image: ", err)
		return "", err
	}
	return str, nil

}

//------------------------------------------
//------------------------------------------

func revive() {
	if err := recover(); err != nil {
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
