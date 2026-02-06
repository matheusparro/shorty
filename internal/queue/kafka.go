// internal/queue/kafka.go
package queue

import (
	"fmt"
	"log"

	"github.com/IBM/sarama"
)

type KafkaClient struct {
	brokers []string
	config  *sarama.Config
}

func NewKafkaClient(brokers []string) (*KafkaClient, error) {
	config := sarama.NewConfig()
	
	// Producer configs
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForLocal
	config.Producer.Compression = sarama.CompressionSnappy
	config.Producer.Retry.Max = 3
	
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	
	config.Metadata.Retry.Max = 3
	config.Version = sarama.V3_0_0_0

	client, err := sarama.NewClient(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka client: %w", err)
	}
	defer client.Close()

	log.Printf("✅ Kafka connected to brokers: %v", brokers)

	return &KafkaClient{
		brokers: brokers,
		config:  config,
	}, nil
}

func (k *KafkaClient) Config() *sarama.Config {
	return k.config
}

func (k *KafkaClient) Brokers() []string {
	return k.brokers
}