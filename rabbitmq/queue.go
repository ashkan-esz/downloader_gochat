package rabbitmq

import (
	errorHandler "downloader_gochat/pkg/error"
	"fmt"

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
	SingleChatQueue        = "singleChat"
	SingleChatBindingKey   = "chat.single"
	GroupChatQueue         = "groupChat"
	GroupChatBindingKey    = "chat.group"
	MessageStateQueue      = "messageState"
	MessageStateBindingKey = "message.state"
	NotificationQueue      = "notification"
	NotificationBindingKey = "notification"
	BlurHashQueue          = "blurHash"
	BlurHashBindingKey     = "blurHash"
	EmailQueue             = "email"
	EmailBindingKey        = "email"
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
		errorMessage := fmt.Sprintf("error creating queue %s: %s", SingleChatQueue, err)
		errorHandler.SaveError(errorMessage, err)
	}

	bindConfig := ConfigBindQueue{
		QueueName:  SingleChatQueue,
		Exchange:   ChatExchange,
		RoutingKey: SingleChatBindingKey,
		NoWait:     false,
	}
	err = r.BindQueueExchange(bindConfig)
	if err != nil {
		errorMessage := fmt.Sprintf("error binding queue %s: %s", SingleChatQueue, err)
		errorHandler.SaveError(errorMessage, err)
	}

	//------------------------------------
	//------------------------------------

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
		errorMessage := fmt.Sprintf("error creating queue %s: %s", GroupChatQueue, err)
		errorHandler.SaveError(errorMessage, err)
	}

	bindConfig = ConfigBindQueue{
		QueueName:  GroupChatQueue,
		Exchange:   ChatExchange,
		RoutingKey: GroupChatBindingKey,
		NoWait:     false,
	}
	err = r.BindQueueExchange(bindConfig)
	if err != nil {
		errorMessage := fmt.Sprintf("error binding queue %s: %s", GroupChatQueue, err)
		errorHandler.SaveError(errorMessage, err)
	}

	//------------------------------------
	//------------------------------------

	MessageStateConfig := ConfigQueue{
		Name:       MessageStateQueue,
		Durable:    true,
		AutoDelete: false,
		Exclusive:  false,
		NoWait:     false,
		Args:       nil,
	}
	_, err = r.CreateQueue(MessageStateConfig)
	if err != nil {
		errorMessage := fmt.Sprintf("error creating queue %s: %s", MessageStateQueue, err)
		errorHandler.SaveError(errorMessage, err)
	}

	MessageStateBindConfig := ConfigBindQueue{
		QueueName:  MessageStateQueue,
		Exchange:   MessageStateExchange,
		RoutingKey: MessageStateBindingKey,
		NoWait:     false,
	}
	err = r.BindQueueExchange(MessageStateBindConfig)
	if err != nil {
		errorMessage := fmt.Sprintf("error binding queue %s: %s", MessageStateQueue, err)
		errorHandler.SaveError(errorMessage, err)
	}

	//------------------------------------
	//------------------------------------

	NotificationConfig := ConfigQueue{
		Name:       NotificationQueue,
		Durable:    true,
		AutoDelete: false,
		Exclusive:  false,
		NoWait:     false,
		Args:       nil,
	}
	_, err = r.CreateQueue(NotificationConfig)
	if err != nil {
		errorMessage := fmt.Sprintf("error creating queue %s: %s", NotificationQueue, err)
		errorHandler.SaveError(errorMessage, err)
	}

	NotificationBindConfig := ConfigBindQueue{
		QueueName:  NotificationQueue,
		Exchange:   NotificationExchange,
		RoutingKey: NotificationBindingKey,
		NoWait:     false,
	}
	err = r.BindQueueExchange(NotificationBindConfig)
	if err != nil {
		errorMessage := fmt.Sprintf("error binding queue %s: %s", NotificationQueue, err)
		errorHandler.SaveError(errorMessage, err)
	}

	//------------------------------------
	//------------------------------------

	blurHashConfig := ConfigQueue{
		Name:       BlurHashQueue,
		Durable:    true,
		AutoDelete: false,
		Exclusive:  false,
		NoWait:     false,
		Args:       nil,
	}
	_, err = r.CreateQueue(blurHashConfig)
	if err != nil {
		errorMessage := fmt.Sprintf("error creating queue %s: %s", BlurHashQueue, err)
		errorHandler.SaveError(errorMessage, err)
	}

	blurHashConfigBindConfig := ConfigBindQueue{
		QueueName:  BlurHashQueue,
		Exchange:   BlurHashExchange,
		RoutingKey: BlurHashBindingKey,
		NoWait:     false,
	}
	err = r.BindQueueExchange(blurHashConfigBindConfig)
	if err != nil {
		errorMessage := fmt.Sprintf("error binding queue %s: %s", BlurHashQueue, err)
		errorHandler.SaveError(errorMessage, err)
	}

	//------------------------------------
	//------------------------------------

	emailConfig := ConfigQueue{
		Name:       EmailQueue,
		Durable:    true,
		AutoDelete: false,
		Exclusive:  false,
		NoWait:     false,
		Args:       nil,
	}
	_, err = r.CreateQueue(emailConfig)
	if err != nil {
		errorMessage := fmt.Sprintf("error creating queue %s: %s", EmailQueue, err)
		errorHandler.SaveError(errorMessage, err)
	}

	emailConfigBindConfig := ConfigBindQueue{
		QueueName:  EmailQueue,
		Exchange:   EmailExchange,
		RoutingKey: EmailBindingKey,
		NoWait:     false,
	}
	err = r.BindQueueExchange(emailConfigBindConfig)
	if err != nil {
		errorMessage := fmt.Sprintf("error binding queue %s: %s", EmailQueue, err)
		errorHandler.SaveError(errorMessage, err)
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
