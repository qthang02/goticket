package entity

import "time"

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "PENDING"
	OrderStatusPaid      OrderStatus = "PAID"
	OrderStatusCancelled OrderStatus = "CANCELLED"
	OrderStatusFailed    OrderStatus = "FAILED"
)

type Order struct {
	ID          uint64      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      uint64      `gorm:"index;not null" json:"user_id"`
	TicketID    uint64      `gorm:"index;not null" json:"ticket_id"`
	Quantity    int         `gorm:"not null" json:"quantity"`
	TotalAmount float64     `gorm:"type:decimal(10,2);not null" json:"total_amount"`
	Status      OrderStatus `gorm:"type:varchar(20);default:'PENDING'" json:"status"`

	// Idempotency Key logic often uses a separate table or field,
	// but adding a reference here can be useful for deduplication.
	IdempotencyKey string `gorm:"type:varchar(255);uniqueIndex" json:"idempotency_key,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
