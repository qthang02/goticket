package repository

import (
	"context"

	"github.com/qthang02/goticket/internal/entity"
	"gorm.io/gorm"
)

type OrderRepository interface {
	Create(ctx context.Context, order *entity.Order) error
	GetByID(ctx context.Context, id uint64) (*entity.Order, error)
	List(ctx context.Context, limit, offset int) ([]*entity.Order, error)
	Update(ctx context.Context, order *entity.Order) error
	Delete(ctx context.Context, id uint64) error
}

type orderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) Create(ctx context.Context, order *entity.Order) error {
	return r.db.WithContext(ctx).Create(order).Error
}

func (r *orderRepository) GetByID(ctx context.Context, id uint64) (*entity.Order, error) {
	var order entity.Order
	if err := r.db.WithContext(ctx).First(&order, id).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) List(ctx context.Context, limit, offset int) ([]*entity.Order, error) {
	var orders []*entity.Order
	if err := r.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

func (r *orderRepository) Update(ctx context.Context, order *entity.Order) error {
	return r.db.WithContext(ctx).Save(order).Error
}

func (r *orderRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&entity.Order{}, id).Error
}
