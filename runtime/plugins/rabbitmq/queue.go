package rabbitmq

import (
	"bytes"
	"context"
	"encoding/gob"

	amqp "github.com/rabbitmq/amqp091-go"
	"gitlab.mpi-sws.org/cld/blueprint/runtime/core/backend"
)

type RabbitMQ struct {
	name  string
	queue amqp.Queue
	ch    *amqp.Channel
	conn  *amqp.Connection
	msgs  <-chan amqp.Delivery
}

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

// Encoding requires that the real type of the interface is already registered with gob.
func getBytes(key interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(key)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func decodeBytes(val []byte) (interface{}, error) {
	encoded := bytes.NewBuffer(val)
	dec := gob.NewDecoder(encoded)
	var res interface{}
	err := dec.Decode(&res)
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
