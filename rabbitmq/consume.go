package rabbitmq

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
)

// IConsumer is the interface for consuming messages from a queue
type IConsumer interface {
	Consume(ctx context.Context, config ConfigConsume, f func(*amqp.Delivery)) (err error)
}

// Consume starts consuming messages from a queue until the context is canceled
func (r *rabbit) Consume(ctx context.Context, config ConfigConsume, f func(*amqp.Delivery)) (err error) {
	if r.chConsumer == nil {
		return amqp.ErrClosed
	}
	r.wg.Add(1)
	defer r.wg.Done()
	var msgs <-chan amqp.Delivery
	//todo : whats consumer string
	msgs, err = r.chConsumer.Consume(
		config.QueueName,
		config.Consumer,
		config.AutoAck,
		config.Exclusive,
		config.NoLocal,
		config.NoWait,
		config.Args,
	)
	if err != nil {
		return
	}
	var allCanceled bool
	for {
		select {
		case msg, ok := <-msgs:
			if !ok {
				return
			}
			r.wg.Add(1)
			if config.ExecuteConcurrent {
				go func() {
					f(&msg)
					r.wg.Done()
				}()
			} else {
				f(&msg)
				r.wg.Done()
			}
		case <-ctx.Done():
			if allCanceled {
				continue
			}
			err = r.chConsumer.Cancel(config.Consumer, false)
			allCanceled = true
			continue
		}
	}
}
