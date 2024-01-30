package rabbitmq

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

// ICreateQueue is the interface for creating
type ICreateQueue interface {
	CreateQueue(config ConfigQueue) (queue amqp.Queue, err error)
}

// IQueueBinder is the interface for binding and unbinding queues
type IQueueBinder interface {
	BindQueueExchange(config ConfigBindQueue) (err error)
	UnbindQueueExchange(config ConfigBindQueue) (err error)
}

const (
	SingleChatQueue      = "singleChat"
	SingleChatBindingKey = "chat.single"
	GroupChatQueue       = "groupChat"
	GroupChatBindingKey  = "chat.group"
)

func (r *rabbit) createQueuesAndBind() {
	config := ConfigQueue{
		Name:       SingleChatQueue,
		Durable:    true,
		AutoDelete: false,
		Exclusive:  false,
		NoWait:     false,
		Args:       nil,
	}
	_, err := r.CreateQueue(config)
	if err != nil {
		log.Printf("error creating queue %s: %s\n", SingleChatQueue, err)
	}

	bindConfig := ConfigBindQueue{
		QueueName:  SingleChatQueue,
		Exchange:   ChatExchange,
		RoutingKey: SingleChatBindingKey,
		NoWait:     false,
	}
	err = r.BindQueueExchange(bindConfig)
	if err != nil {
		log.Printf("error binding queue %s: %s\n", SingleChatQueue, err)
	}

	config = ConfigQueue{
		Name:       GroupChatQueue,
		Durable:    true,
		AutoDelete: false,
		Exclusive:  false,
		NoWait:     false,
		Args:       nil,
	}
	_, err = r.CreateQueue(config)
	if err != nil {
		log.Printf("error creating queue %s: %s\n", GroupChatQueue, err)
	}

	bindConfig = ConfigBindQueue{
		QueueName:  GroupChatQueue,
		Exchange:   ChatExchange,
		RoutingKey: GroupChatBindingKey,
		NoWait:     false,
	}
	err = r.BindQueueExchange(bindConfig)
	if err != nil {
		log.Printf("error binding queue %s: %s\n", GroupChatQueue, err)
	}

}

// CreateQueue creates a queue
func (r *rabbit) CreateQueue(config ConfigQueue) (queue amqp.Queue, err error) {
	if r.chConsumer == nil {
		err = amqp.ErrClosed
		return
	}
	queue, err = r.chConsumer.QueueDeclare(
		config.Name,
		config.Durable,
		config.AutoDelete,
		config.Exclusive,
		config.NoWait,
		config.Args,
	)
	return
}

// BindQueueExchange binds a queue to an exchange
func (r *rabbit) BindQueueExchange(config ConfigBindQueue) (err error) {
	if r.chConsumer == nil {
		err = amqp.ErrClosed
		return
	}
	err = r.chConsumer.QueueBind(
		config.QueueName,
		config.RoutingKey,
		config.Exchange,
		config.NoWait,
		config.Args,
	)
	return
}

// UnbindQueueExchange unbinds a queue from an exchange
func (r *rabbit) UnbindQueueExchange(config ConfigBindQueue) (err error) {
	if r.chConsumer == nil {
		err = amqp.ErrClosed
		return
	}
	err = r.chConsumer.QueueUnbind(
		config.QueueName,
		config.RoutingKey,
		config.Exchange,
		config.Args,
	)
	return
}
