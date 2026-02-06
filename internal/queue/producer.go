// internal/queue/producer.go
package queue

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/IBM/sarama"
)

type Producer struct {
	producer sarama.SyncProducer
}

func NewProducer(client *KafkaClient) (*Producer, error) {
	producer, err := sarama.NewSyncProducer(client.Brokers(), client.Config())
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}
	return &Producer{producer: producer}, nil
}

// PublishEvent envia evento com key (particiona por shortCode, por exemplo)
func (p *Producer) PublishEvent(topic string, key string, event any) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(data),
	}

	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	log.Printf("📤 event published topic=%s partition=%d offset=%d key=%s", topic, partition, offset, key)
	return nil
}

func (p *Producer) Close() error {
	return p.producer.Close()
}
