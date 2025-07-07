package internal

import (
	"log"

	"github.com/IBM/sarama"
)

type KafkaProducer struct {
	Producer sarama.SyncProducer
}

func NewKafkaProducer(brokers []string) (*KafkaProducer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}
	return &KafkaProducer{Producer: producer}, nil
}

func (kp *KafkaProducer) SendMessage(topic string, key, value []byte) error {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.ByteEncoder(key),
		Value: sarama.ByteEncoder(value),
	}
	_, _, err := kp.Producer.SendMessage(msg)
	return err
}

func (kp *KafkaProducer) Close() {
	if err := kp.Producer.Close(); err != nil {
		log.Printf("failed to close producer: %v", err)
	}
}
