package rabbitmq

import (
	"context"
	"downloader_gochat/configs"
	errorHandler "downloader_gochat/pkg/error"
	"fmt"
	"log"
	"runtime"
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
	consumerConn        *amqp.Connection
	producerConn        *amqp.Connection
	chConsumer          *amqp.Channel
	chProducer          *amqp.Channel
	consumerChannelPool []*amqp.Channel
	producerChannelPool []*amqp.Channel
	consumerPoolMux     *sync.Mutex
	producerPoolMux     *sync.Mutex
	wg                  *sync.WaitGroup
}

//------------------------------------
//------------------------------------

var rabbitmq *rabbit

func Start(ctx context.Context) RabbitMQ {
	rabbitmq = &rabbit{
		wg:                  &sync.WaitGroup{},
		consumerPoolMux:     &sync.Mutex{},
		producerPoolMux:     &sync.Mutex{},
		consumerChannelPool: make([]*amqp.Channel, 0),
		producerChannelPool: make([]*amqp.Channel, 0),
	}
	conf := ConfigConnection{URI: configs.GetConfigs().RabbitMqUrl, PrefetchCount: 3, PublishChannelPoolCount: 5}
	rabbitmq.Setup(ctx, conf)
	return rabbitmq
}

// Setup starts a goroutine to keep the connection open and everytime the connection is open,
// it will call the setupRabbit function.
// It is important to pass a context with cancel so the goroutine can be closed when the context is done.
// Otherwise, it will run until the program ends.
func (r *rabbit) Setup(ctx context.Context, config ConfigConnection) {
	go func() {
		counter := 0
		for {
			notifyClose, err := r.Connect(config)
			if err != nil {
				counter = counter + 1
				if counter > 10 {
					log.Printf("error connecting to rabbitmq: [%s]\n", err)
				}
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
	r.producerConn, err = amqp.Dial(config.URI)
	if err != nil {
		return
	}
	r.consumerConn, err = amqp.Dial(config.URI)
	if err != nil {
		return
	}

	r.chConsumer, err = r.consumerConn.Channel()
	if err != nil {
		return
	}
	if config.PrefetchCount > 0 {
		err = r.chConsumer.Qos(config.PrefetchCount, 0, false)
		if err != nil {
			return
		}
	}

	r.chProducer, err = r.producerConn.Channel()
	if err != nil {
		return
	}
	for i := 0; i < config.PublishChannelPoolCount; i++ {
		_, err = r.createProducerChannel()
		if err != nil {
			return
		}
	}

	notifyOpenConnections()
	notify = make(chan *amqp.Error)
	r.producerConn.NotifyClose(notify)
	r.consumerConn.NotifyClose(notify)
	return
}

//---------------------------------------
//---------------------------------------

func (r *rabbit) createConsumerChannel(prefetchCount int) (*amqp.Channel, error) {
	r.consumerPoolMux.Lock()
	defer r.consumerPoolMux.Unlock()
	consumerChan, err := r.consumerConn.Channel()
	if err != nil {
		return nil, err
	}
	err = consumerChan.Qos(prefetchCount, 0, false)
	if err != nil {
		return nil, err
	}
	r.consumerChannelPool = append(r.consumerChannelPool, consumerChan)
	return consumerChan, nil
}

func (r *rabbit) createProducerChannel() (*amqp.Channel, error) {
	r.producerPoolMux.Lock()
	defer r.producerPoolMux.Unlock()
	producerChan, err := r.producerConn.Channel()
	if err != nil {
		return nil, err
	}
	r.producerChannelPool = append(r.producerChannelPool, producerChan)
	return producerChan, nil
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
			errorMessage := fmt.Sprintf("Error closing consumer channel: [%s]", err)
			errorHandler.SaveError(errorMessage, err)
		}
	}
	for i, consumeChan := range r.consumerChannelPool {
		if consumeChan != nil {
			err = consumeChan.Close()
			if err != nil {
				errorMessage := fmt.Sprintf("Error closing consumer channel %v: [%s]", i, err)
				errorHandler.SaveError(errorMessage, err)
			}
		}
	}

	if r.chProducer != nil {
		err = r.chProducer.Close()
		if err != nil {
			errorMessage := fmt.Sprintf("Error closing producer channel: [%s]", err)
			errorHandler.SaveError(errorMessage, err)
		}
	}
	for i, produceChan := range r.producerChannelPool {
		if produceChan != nil {
			err = produceChan.Close()
			if err != nil {
				errorMessage := fmt.Sprintf("Error closing producer channel %v: [%s]", i, err)
				errorHandler.SaveError(errorMessage, err)
			}
		}
	}

	if r.producerConn != nil {
		err = r.producerConn.Close()
		if err != nil {
			errorMessage := fmt.Sprintf("Error closing connection: [%s]", err)
			errorHandler.SaveError(errorMessage, err)
		}
	}
	if r.consumerConn != nil {
		err = r.consumerConn.Close()
		if err != nil {
			errorMessage := fmt.Sprintf("Error closing connection: [%s]", err)
			errorHandler.SaveError(errorMessage, err)
		}
	}
}

//---------------------------------------
//---------------------------------------

func getMaxParallelism() int {
	maxProcs := runtime.GOMAXPROCS(0)
	numCPU := runtime.NumCPU()
	if maxProcs < numCPU {
		return maxProcs
	}
	return numCPU
}
