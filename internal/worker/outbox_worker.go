package worker

import (
	"context"
	"time"

	"github.com/qthang02/goticket/internal/entity"
	"github.com/qthang02/goticket/internal/infrastructure/kafka"
	"github.com/qthang02/goticket/internal/repository"
	"github.com/qthang02/goticket/pkg/logger"
	"go.uber.org/zap"
)

type OutboxWorker struct {
	repo     repository.OutboxRepository
	producer *kafka.Producer
}

func NewOutboxWorker(repo repository.OutboxRepository, producer *kafka.Producer) *OutboxWorker {
	return &OutboxWorker{
		repo:     repo,
		producer: producer,
	}
}

func (w *OutboxWorker) Start(ctx context.Context) {
	logger.Info("Outbox worker started")
	ticker := time.NewTicker(2 * time.Second) // Poll every 2 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("Outbox worker stopped")
			return
		case <-ticker.C:
			w.processOutbox(ctx)
		}
	}
}

func (w *OutboxWorker) processOutbox(ctx context.Context) {
	// 1. Fetch Pending Events
	events, err := w.repo.GetPending(ctx, 10)
	if err != nil {
		logger.Error("Failed to fetch pending outbox events", zap.Error(err))
		return
	}

	if len(events) == 0 {
		return
	}

	logger.Info("Processing outbox events", zap.Int("count", len(events)))

	// 2. Publish to Kafka
	for _, event := range events {
		err := w.producer.Publish(ctx, event.Topic, nil, []byte(event.Payload))
		if err != nil {
			logger.Error("Failed to publish to Kafka", zap.Uint64("id", event.ID), zap.Error(err))
			// Retry Logic: Increment retry count or just leave it PENDING to pick up later.
			// For simplicity: Leave it PENDING. Ideally use Exponential Backoff.
			continue
		}

		// 3. Mark as Processed
		event.Status = entity.OutboxStatusProcessed
		event.UpdatedAt = time.Now()
		if err := w.repo.Update(ctx, event); err != nil {
			logger.Error("Failed to update outbox status", zap.Uint64("id", event.ID), zap.Error(err))
		}

		logger.Info("Event published", zap.Uint64("id", event.ID), zap.String("topic", event.Topic))
	}
}
