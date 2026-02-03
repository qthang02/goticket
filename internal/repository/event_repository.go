package repository

import (
	"context"

	"github.com/qthang02/goticket/internal/entity"
	"gorm.io/gorm"
)

type EventRepository interface {
	GetByID(ctx context.Context, id uint64) (*entity.Event, error)
	List(ctx context.Context, limit, offset int) ([]*entity.Event, error)
	Create(ctx context.Context, event *entity.Event) error
	Update(ctx context.Context, event *entity.Event) error
	Delete(ctx context.Context, id uint64) error
}

type eventRepository struct {
	db *gorm.DB
}

func NewEventRepository(db *gorm.DB) EventRepository {
	return &eventRepository{db: db}
}

func (r *eventRepository) GetByID(ctx context.Context, id uint64) (*entity.Event, error) {
	var event entity.Event
	if err := r.db.WithContext(ctx).Preload("Tickets").First(&event, id).Error; err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *eventRepository) List(ctx context.Context, limit, offset int) ([]*entity.Event, error) {
	var events []*entity.Event
	if err := r.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}

func (r *eventRepository) Create(ctx context.Context, event *entity.Event) error {
	return r.db.WithContext(ctx).Create(event).Error
}

func (r *eventRepository) Update(ctx context.Context, event *entity.Event) error {
	return r.db.WithContext(ctx).Save(event).Error
}

func (r *eventRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&entity.Event{}, id).Error
}
