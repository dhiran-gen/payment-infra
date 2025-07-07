package internal

import (
	"context"
	"log"

	"github.com/IBM/sarama"
)

type KafkaConsumer struct {
	Consumer sarama.ConsumerGroup
}

type paymentHandler struct{}

func (h *paymentHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (h *paymentHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (h *paymentHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		log.Printf("Received message: %s", string(msg.Value))
		sess.MarkMessage(msg, "")
	}
	return nil
}

func NewKafkaConsumer(brokers []string, groupID string) (*KafkaConsumer, error) {
	config := sarama.NewConfig()
	config.Version = sarama.V2_1_0_0
	consumer, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, err
	}
	return &KafkaConsumer{Consumer: consumer}, nil
}

func (kc *KafkaConsumer) Start(topic string, ctx context.Context) {
	h := &paymentHandler{}
	for {
		if err := kc.Consumer.Consume(ctx, []string{topic}, h); err != nil {
			log.Printf("Error from consumer: %v", err)
		}
		if ctx.Err() != nil {
			return
		}
	}
}

func (kc *KafkaConsumer) Close() {
	if err := kc.Consumer.Close(); err != nil {
		log.Printf("failed to close consumer: %v", err)
	}
}
