package rabbitmq

import (
	"log"

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
		log.Printf("error creating exchange %v: %s\n", ChatExchange, err)
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
		log.Printf("error creating exchange %v: %s\n", MessageStateExchange, err)
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
		log.Printf("error creating exchange %v: %s\n", NotificationExchange, err)
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
