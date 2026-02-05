package kafka

import (
	"context"
	"net"
	"strconv"

	"github.com/qthang02/goticket/config"
	"github.com/qthang02/goticket/pkg/logger"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type Producer struct {
	Writer *kafka.Writer
}

func NewProducer(cfg *config.Config) *Producer {
	// brokers := []string{net.JoinHostPort(cfg.Kafka.Broker, strconv.Itoa(cfg.Kafka.Port))}
	brokers := cfg.Kafka.Brokers
	if len(brokers) == 0 {
		// Fallback for localhost if empty (or panic)
		brokers = []string{"localhost:9092"}
	}

	logger.Info("Connecting to Kafka...", zap.Strings("brokers", brokers))

	w := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Balancer: &kafka.LeastBytes{},
	}

	return &Producer{Writer: w}
}

func (p *Producer) EnsureTopic(topic string) error {
	// For simplicity, execute against the first broker
	addr := p.Writer.Addr.String()

	// Create connection to Kafka Controller (or any broker in KRaft)
	conn, err := kafka.Dial("tcp", addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return err
	}

	controllerConn, err := kafka.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		return err
	}
	defer controllerConn.Close()

	topicConfigs := []kafka.TopicConfig{
		{
			Topic:             topic,
			NumPartitions:     3,
			ReplicationFactor: 1,
		},
	}

	return controllerConn.CreateTopics(topicConfigs...)
}

func (p *Producer) Publish(ctx context.Context, topic string, key, value []byte) error {
	msg := kafka.Message{
		Topic: topic,
		Key:   key,
		Value: value,
	}
	return p.Writer.WriteMessages(ctx, msg)
}

func (p *Producer) Close() error {
	return p.Writer.Close()
}
