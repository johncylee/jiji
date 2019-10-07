package jiji

import (
	"testing"
)

const (
	KAFKA_BROKERS = "kafka.test.server"
	KAFKA_TOPIC   = "kafka.test.topic"
)

func TestKafka(t *testing.T) {
	transport := Kafka{
		Brokers: KAFKA_BROKERS,
		Topic:   KAFKA_TOPIC,
	}
	testDelivery(&transport, t)
}
