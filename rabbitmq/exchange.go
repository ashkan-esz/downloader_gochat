package rabbitmq

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type ICreateExchange interface {
	CreateExchange(config ConfigExchange) (err error)
}

const (
	chatExchange     = "chatExchange"
	chatExchangeType = "topic"
)

func (r *rabbit) createExchanges() {
	config := ConfigExchange{
		Name:       chatExchange,
		Type:       chatExchangeType,
		Durable:    true,
		AutoDelete: false,
		Internal:   false,
		NoWait:     false,
		Args:       nil,
	}
	err := r.CreateExchange(config)
	if err != nil {
		log.Printf("error creating queue: %s\n", err)
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
