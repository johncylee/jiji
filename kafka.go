//go:build cgo && (((darwin || linux) && amd64) || dynamic)

// build constraint introduced by confluent-kafka-go

package jiji

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

// "Kafka" transport needs update. Obsolete for now.
type Kafka struct {
	Brokers        string
	Topic          string
	producer       *kafka.Producer
	topicPartition kafka.TopicPartition
}

func (t *Kafka) Connect() (err error) {
	t.producer, err = kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": t.Brokers,
	})
	t.topicPartition = kafka.TopicPartition{
		Topic:     &t.Topic,
		Partition: kafka.PartitionAny,
	}
	return
}

func (t *Kafka) Close() {
	t.producer.Close()
}

func (t *Kafka) Send(msg []byte) (err error) {
	err = t.producer.Produce(&kafka.Message{
		TopicPartition: t.topicPartition,
		Value:          msg,
	}, nil)
	if err != nil {
		return
	}
FOR:
	for e := range t.producer.Events() {
		switch ev := e.(type) {
		case *kafka.Message:
			err = ev.TopicPartition.Error
			if err != nil {
				Logger.Error("Delivery failed", "error", err)
				return
			}
			Logger.Debug("Delivered to",
				"topic", *ev.TopicPartition.Topic,
				"partition", ev.TopicPartition.Partition,
				"offset", ev.TopicPartition.Offset)
			break FOR
		case kafka.Error:
			err = ev
			break FOR
		default:
			Logger.Warn("Ignored", "event", e)
		}
	}
	return
}
