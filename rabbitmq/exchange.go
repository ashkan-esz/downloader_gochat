package rabbitmq

import (
	errorHandler "downloader_gochat/pkg/error"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type ICreateExchange interface {
	CreateExchange(config ConfigExchange) (err error)
}

const (
	ChatExchange             = "ChatExchange"
	ChatExchangeType         = "topic"
	MessageStateExchange     = "MessageStateExchange"
	MessageStateExchangeType = "direct"
	NotificationExchange     = "NotificationExchange"
	NotificationExchangeType = "direct"
	BlurHashExchange         = "BlurHashExchange"
	BlurHashExchangeType     = "direct"
	EmailExchange            = "EmailExchange"
	EmailExchangeType        = "direct"
)

func (r *rabbit) createExchanges() {
	config := ConfigExchange{
		Name:       ChatExchange,
		Type:       ChatExchangeType,
		Durable:    true,
		AutoDelete: false,
		Internal:   false,
		NoWait:     false,
		Args:       nil,
	}
	err := r.CreateExchange(config)
	if err != nil {
		errorMessage := fmt.Sprintf("error creating exchange %v: %s", ChatExchange, err)
		errorHandler.SaveError(errorMessage, err)
	}

	messageStateConfig := ConfigExchange{
		Name:       MessageStateExchange,
		Type:       MessageStateExchangeType,
		Durable:    true,
		AutoDelete: false,
		Internal:   false,
		NoWait:     false,
		Args:       nil,
	}
	err = r.CreateExchange(messageStateConfig)
	if err != nil {
		errorMessage := fmt.Sprintf("error creating exchange %v: %s", MessageStateExchange, err)
		errorHandler.SaveError(errorMessage, err)
	}

	notificationConfig := ConfigExchange{
		Name:       NotificationExchange,
		Type:       NotificationExchangeType,
		Durable:    true,
		AutoDelete: false,
		Internal:   false,
		NoWait:     false,
		Args:       nil,
	}
	err = r.CreateExchange(notificationConfig)
	if err != nil {
		errorMessage := fmt.Sprintf("error creating exchange %v: %s", NotificationExchange, err)
		errorHandler.SaveError(errorMessage, err)
	}

	blurHashConfig := ConfigExchange{
		Name:       BlurHashExchange,
		Type:       BlurHashExchangeType,
		Durable:    true,
		AutoDelete: false,
		Internal:   false,
		NoWait:     false,
		Args:       nil,
	}
	err = r.CreateExchange(blurHashConfig)
	if err != nil {
		errorMessage := fmt.Sprintf("error creating exchange %v: %s", BlurHashExchange, err)
		errorHandler.SaveError(errorMessage, err)
	}

	emailConfig := ConfigExchange{
		Name:       EmailExchange,
		Type:       EmailExchangeType,
		Durable:    true,
		AutoDelete: false,
		Internal:   false,
		NoWait:     false,
		Args:       nil,
	}
	err = r.CreateExchange(emailConfig)
	if err != nil {
		errorMessage := fmt.Sprintf("error creating exchange %v: %s", EmailExchange, err)
		errorHandler.SaveError(errorMessage, err)
	}
}

// CreateExchange creates an exchange
func (r *rabbit) CreateExchange(config ConfigExchange) (err error) {
	if r.chConsumer == nil {
		return amqp.ErrClosed
	}
	err = r.chConsumer.ExchangeDeclare(
		config.Name,
		config.Type,
		config.Durable,
		config.AutoDelete,
		config.Internal,
		config.NoWait,
		config.Args,
	)
	return
}
