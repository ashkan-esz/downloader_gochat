package rabbitmq

import (
	"context"
	"downloader_gochat/configs"
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitMQ combines all the interfaces of the package
type RabbitMQ interface {
	ICreateQueue
	IQueueBinder
	ICreateExchange
	IConsumer
	IPublish
	Setup(ctx context.Context, config ConfigConnection)
	Connect(config ConfigConnection) (notify chan *amqp.Error, err error)
	Close(ctx context.Context) (done chan struct{})
}

type rabbit struct {
	conn       *amqp.Connection
	chConsumer *amqp.Channel
	chProducer *amqp.Channel
	wg         *sync.WaitGroup
}

//------------------------------------
//------------------------------------

var rabbitmq *rabbit

func Start(ctx context.Context) RabbitMQ {
	rabbitmq = &rabbit{
		wg: &sync.WaitGroup{},
	}
	conf := ConfigConnection{URI: configs.GetConfigs().RabbitMqUrl, PrefetchCount: 1}
	rabbitmq.Setup(ctx, conf)
	return rabbitmq
}

// Setup starts a goroutine to keep the connection open and everytime the connection is open,
// it will call the setupRabbit function.
// It is important to pass a context with cancel so the goroutine can be closed when the context is done.
// Otherwise, it will run until the program ends.
func (r *rabbit) Setup(ctx context.Context, config ConfigConnection) {
	go func() {
		for {
			notifyClose, err := r.Connect(config)
			if err != nil {
				log.Printf("error connecting to rabbitmq: [%s]\n", err)
				time.Sleep(time.Second * 5)
				continue
			}
			r.createExchanges()
			r.createQueuesAndBind()
			notifySetupIsDone()
			select {
			case <-notifyClose:
				continue
			case <-ctx.Done():
				return
			}
		}
	}()
}

// Connect connects to the rabbitMQ server and also creates the channels to produce and consume messages.
// It can also notify the connection is open to other goroutines if the function NotifyOpenConnection
// is called before connecting.
func (r *rabbit) Connect(config ConfigConnection) (notify chan *amqp.Error, err error) {
	//todo : how to use multi consumer goroutine
	//todo : its ok to use only one channel
	r.conn, err = amqp.Dial(config.URI)
	if err != nil {
		return
	}
	r.chProducer, err = r.conn.Channel()
	if err != nil {
		return
	}
	r.chConsumer, err = r.conn.Channel()
	if err != nil {
		return
	}
	if config.PrefetchCount > 0 {
		//todo : check this
		err = r.chConsumer.Qos(config.PrefetchCount, 0, false)
		if err != nil {
			return
		}
	}
	notifyOpenConnections()
	notify = make(chan *amqp.Error)
	r.conn.NotifyClose(notify)
	return
}

//---------------------------------------
//---------------------------------------

func (r *rabbit) Close(ctx context.Context) (done chan struct{}) {
	done = make(chan struct{})

	doneWaiting := make(chan struct{})
	go func() {
		r.wg.Wait()
		close(doneWaiting)
	}()

	go func() {
		defer close(done)
		select { // either waits for the messages to process or timeout from context
		case <-doneWaiting:
		case <-ctx.Done():
		}
		closeConnections(r)
	}()
	return
}

func closeConnections(r *rabbit) {
	var err error
	if r.chConsumer != nil {
		err = r.chConsumer.Close()
		if err != nil {
			log.Printf("Error closing consumer channel: [%s]\n", err)
		}
	}
	if r.chProducer != nil {
		err = r.chProducer.Close()
		if err != nil {
			log.Printf("Error closing producer channel: [%s]\n", err)
		}
	}
	if r.conn != nil {
		err = r.conn.Close()
		if err != nil {
			log.Printf("Error closing connection: [%s]\n", err)
		}
	}
}
