package entity

import "time"

type OutboxStatus string

const (
	OutboxStatusPending   OutboxStatus = "PENDING"
	OutboxStatusProcessed OutboxStatus = "PROCESSED"
	OutboxStatusFailed    OutboxStatus = "FAILED"
)

type Outbox struct {
	ID         uint64       `gorm:"primaryKey;autoIncrement" json:"id"`
	Topic      string       `gorm:"type:varchar(255);not null" json:"topic"`
	Payload    string       `gorm:"type:text;not null" json:"payload"` // JSON string
	Status     OutboxStatus `gorm:"type:varchar(50);default:'PENDING';index" json:"status"`
	RetryCount int          `gorm:"default:0" json:"retry_count"`
	CreatedAt  time.Time    `json:"created_at"`
	UpdatedAt  time.Time    `json:"updated_at"`
}
