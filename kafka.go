//go:build cgo && (((darwin || linux) && amd64) || dynamic)

// build constraint introduced by confluent-kafka-go

package jiji

import (
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
)

const (
	KAFKA_TIMEOUT_MS = 30000 // 30 seconds
)

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
	return
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
				Logger.Println("Delivery failed:", err)
				return
			}
			debugf("Delivered to topic %s [%d] at offset %v\n",
				*ev.TopicPartition.Topic,
				ev.TopicPartition.Partition,
				ev.TopicPartition.Offset)
			break FOR
		case kafka.Error:
			err = ev
			break FOR
		default:
			Logger.Println("Ignored event:", e)
		}
	}
	return
}
