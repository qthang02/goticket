package repository

import (
	"context"

	"github.com/qthang02/goticket/internal/entity"
	"gorm.io/gorm"
)

type OutboxRepository interface {
	Create(ctx context.Context, tx *gorm.DB, outbox *entity.Outbox) error
	GetPending(ctx context.Context, limit int) ([]*entity.Outbox, error)
	Update(ctx context.Context, outbox *entity.Outbox) error
}

type outboxRepository struct {
	db *gorm.DB
}

func NewOutboxRepository(db *gorm.DB) OutboxRepository {
	return &outboxRepository{db: db}
}

// Create uses an existing transaction (tx) to insert the event
func (r *outboxRepository) Create(ctx context.Context, tx *gorm.DB, outbox *entity.Outbox) error {
	if tx == nil {
		tx = r.db
	}
	return tx.WithContext(ctx).Create(outbox).Error
}

func (r *outboxRepository) GetPending(ctx context.Context, limit int) ([]*entity.Outbox, error) {
	var events []*entity.Outbox
	err := r.db.WithContext(ctx).
		Where("status = ?", entity.OutboxStatusPending).
		Order("created_at asc").
		Limit(limit).
		Find(&events).Error
	return events, err
}

func (r *outboxRepository) Update(ctx context.Context, outbox *entity.Outbox) error {
	return r.db.WithContext(ctx).Save(outbox).Error
}
