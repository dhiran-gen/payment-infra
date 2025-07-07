package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/payment-infra/payment-worker/internal"
)

func main() {
	log.Println("payment-worker started (Kafka consumer)")
	brokers := []string{"localhost:9092"}
	groupID := "payment-worker-group"
	consumer, err := internal.NewKafkaConsumer(brokers, groupID)
	if err != nil {
		log.Fatalf("failed to create kafka consumer: %v", err)
	}
	defer consumer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		cancel()
	}()

	consumer.Start("payments.commands", ctx)
}
