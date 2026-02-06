// internal/queue/consumer.go
package queue

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/IBM/sarama"
)

type EventHandler func(ctx context.Context, message []byte) error

type Consumer struct {
	consumerGroup sarama.ConsumerGroup
	handlers      map[string]EventHandler
}

func NewConsumer(client *KafkaClient, groupID string) (*Consumer, error) {
	consumerGroup, err := sarama.NewConsumerGroup(client.Brokers(), groupID, client.Config())
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer group: %w", err)
	}

	return &Consumer{
		consumerGroup: consumerGroup,
		handlers:      make(map[string]EventHandler),
	}, nil
}

// RegisterHandler registra um handler para um tópico
func (c *Consumer) RegisterHandler(topic string, handler EventHandler) {
	c.handlers[topic] = handler
	log.Printf("📥 handler registered for topic: %s", topic)
}

// Start inicia o consumer (blocking) com retry/backoff e leitura de Errors()
func (c *Consumer) Start(ctx context.Context, topics []string) error {
	handler := &consumerGroupHandler{handlers: c.handlers}

	// ✅ drena erros do consumer group
	go func() {
		for err := range c.consumerGroup.Errors() {
			log.Printf("❌ consumer group error: %v", err)
		}
	}()

	backoff := 500 * time.Millisecond
	maxBackoff := 5 * time.Second

	for {
		if err := c.consumerGroup.Consume(ctx, topics, handler); err != nil {
			// ✅ retry com backoff (evita loop de CPU / crash por oscilação)
			log.Printf("❌ consume error: %v (retry in %s)", err, backoff)
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return ctx.Err()
			}

			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
			continue
		}

		// se consumiu sem erro, reseta backoff
		backoff = 500 * time.Millisecond

		// contexto cancelado => encerra
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
}

func (c *Consumer) Close() error {
	return c.consumerGroup.Close()
}

// consumerGroupHandler implementa sarama.ConsumerGroupHandler
type consumerGroupHandler struct {
	handlers map[string]EventHandler
}

func (h *consumerGroupHandler) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (h *consumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (h *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		topic := message.Topic

		handler, exists := h.handlers[topic]
		if !exists {
			log.Printf("⚠️ no handler for topic: %s", topic)
			session.MarkMessage(message, "")
			continue
		}

		if err := handler(session.Context(), message.Value); err != nil {
			log.Printf("❌ error processing message topic=%s offset=%d: %v", topic, message.Offset, err)
			// não marca -> reprocessa
			continue
		}

		session.MarkMessage(message, "")
		log.Printf("✅ processed topic=%s offset=%d", topic, message.Offset)
	}

	return nil
}
