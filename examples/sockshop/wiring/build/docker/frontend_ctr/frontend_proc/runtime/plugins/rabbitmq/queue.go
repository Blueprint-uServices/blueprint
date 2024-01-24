// Package rabbitmq provides a client-wrapper implementation of the [backend.Queue] interface for a rabbitmq server.
package rabbitmq

import (
	"context"
	"encoding/json"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	amqp "github.com/rabbitmq/amqp091-go"
)

// Implements a Queue that uses the rabbitmq package
type RabbitMQ struct {
	name  string
	queue amqp.Queue
	ch    *amqp.Channel
	conn  *amqp.Connection
	msgs  <-chan amqp.Delivery
}

// Instantiates a new [Queue] instances that provides a queue interface via a RabbitMQ instance
func NewRabbitMQ(ctx context.Context, addr string, queue_name string) (*RabbitMQ, error) {
	conn, err := amqp.Dial("amqp://guest:guest@" + addr + "/")
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	q, err := ch.QueueDeclare(queue_name, false, false, false, false, nil)
	if err != nil {
		return nil, err
	}
	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		return nil, err
	}
	return &RabbitMQ{name: queue_name, conn: conn, ch: ch, queue: q, msgs: msgs}, nil
}

func getBytes(key interface{}) ([]byte, error) {
	return json.Marshal(key)
}

func decodeBytes(val []byte) (interface{}, error) {
	var res interface{}
	err := json.Unmarshal(val, &res)
	return res, err
}

// Push implements backend.Queue
func (q *RabbitMQ) Push(ctx context.Context, item interface{}) (bool, error) {
	raw_bytes, err := getBytes(item)
	if err != nil {
		return false, err
	}
	publish_msg := amqp.Publishing{ContentType: "text/plain", Body: raw_bytes}
	return true, q.ch.Publish("", q.queue.Name, false, false, publish_msg)
}

// Pop implements backend.Queue
func (q *RabbitMQ) Pop(ctx context.Context, dst interface{}) (bool, error) {
	select {
	case v := <-q.msgs:
		val, err := decodeBytes(v.Body)
		if err != nil {
			return true, err
		}
		return true, backend.CopyResult(val, dst)
	default:
		{
			select {
			case v := <-q.msgs:
				val, err := decodeBytes(v.Body)
				if err != nil {
					return true, err
				}
				return true, backend.CopyResult(val, dst)
			case <-ctx.Done():
				return false, nil
			}
		}
	}
}
