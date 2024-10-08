package error

import (
	"downloader_gochat/configs"
	"errors"
	"log"

	"github.com/getsentry/sentry-go"
)

func SaveError(message string, err error) {
	if configs.GetConfigs().PrintErrors {
		log.Println(message)
	}

	if err == nil {
		sentry.CaptureMessage(message)
	} else {
		sentry.CaptureException(err)
	}
}

var (
	ImageNotFoundError = errors.New("image not found")
)
