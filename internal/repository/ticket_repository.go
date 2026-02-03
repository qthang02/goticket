package repository

import (
	"context"

	"github.com/qthang02/goticket/internal/entity"
	"gorm.io/gorm"
)

type TicketRepository interface {
	Create(ctx context.Context, ticket *entity.Ticket) error
	GetByID(ctx context.Context, id uint64) (*entity.Ticket, error)
	GetByEventID(ctx context.Context, eventID uint64) ([]*entity.Ticket, error)
	Update(ctx context.Context, ticket *entity.Ticket) error
	Delete(ctx context.Context, id uint64) error
	// Atomic update for database-only locking (pessimistic) - optional benchmark reference
	DecrementStock(ctx context.Context, ticketID uint64, quantity int) error
}

type ticketRepository struct {
	db *gorm.DB
}

func NewTicketRepository(db *gorm.DB) TicketRepository {
	return &ticketRepository{db: db}
}

func (r *ticketRepository) Create(ctx context.Context, ticket *entity.Ticket) error {
	return r.db.WithContext(ctx).Create(ticket).Error
}

func (r *ticketRepository) GetByID(ctx context.Context, id uint64) (*entity.Ticket, error) {
	var ticket entity.Ticket
	if err := r.db.WithContext(ctx).First(&ticket, id).Error; err != nil {
		return nil, err
	}
	return &ticket, nil
}

func (r *ticketRepository) GetByEventID(ctx context.Context, eventID uint64) ([]*entity.Ticket, error) {
	var tickets []*entity.Ticket
	if err := r.db.WithContext(ctx).Where("event_id = ?", eventID).Find(&tickets).Error; err != nil {
		return nil, err
	}
	return tickets, nil
}

func (r *ticketRepository) Update(ctx context.Context, ticket *entity.Ticket) error {
	return r.db.WithContext(ctx).Save(ticket).Error
}

func (r *ticketRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&entity.Ticket{}, id).Error
}

func (r *ticketRepository) DecrementStock(ctx context.Context, ticketID uint64, quantity int) error {
	// Simple SQL decrement, not the main high-concurrency solution (which uses Redis),
	// but good for reliability/fallback.
	return r.db.WithContext(ctx).Model(&entity.Ticket{}).
		Where("id = ? AND remaining_stock >= ?", ticketID, quantity).
		UpdateColumn("remaining_stock", gorm.Expr("remaining_stock - ?", quantity)).Error
}
