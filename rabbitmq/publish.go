package rabbitmq

import (
	"context"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

// IPublish is the interface for publishing messages to an exchange
type IPublish interface {
	Publish(ctx context.Context, myStruct interface{}, config ConfigPublish, id int64) (err error)
}

// Publish publishes body to exchange with routing key
func (r *rabbit) Publish(ctx context.Context, myStruct interface{}, config ConfigPublish, id int64) (err error) {
	body, err := json.Marshal(myStruct)
	if err != nil {
		// handle error
		return err
	}

	if r.chProducer == nil {
		return amqp.ErrClosed
	}
	r.wg.Add(1)
	defer r.wg.Done()

	//get producer channel from pool
	producerChan := r.producerChannelPool[id%int64(len(r.producerChannelPool))]

	err = producerChan.PublishWithContext(
		ctx,
		config.Exchange,
		config.RoutingKey,
		config.Mandatory,
		config.Immediate,
		amqp.Publishing{
			Headers:         config.Headers,
			ContentType:     config.ContentType,
			ContentEncoding: config.ContentEncoding,
			Priority:        config.Priority,
			CorrelationId:   config.CorrelationID,
			MessageId:       config.MessageID,
			Body:            body,
		},
	)
	return
}
