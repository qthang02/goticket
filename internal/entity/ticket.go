package entity

import "time"

// Ticket represents a ticket category/type for an event (e.g., "VIP", "General Admission")
// It holds the inventory/stock information.
type Ticket struct {
	ID      uint64  `gorm:"primaryKey;autoIncrement" json:"id"`
	EventID uint64  `gorm:"index;not null" json:"event_id"`
	Name    string  `gorm:"type:varchar(100);not null" json:"name"` // e.g. "VIP", "Regular"
	Price   float64 `gorm:"type:decimal(10,2);not null" json:"price"`

	// Inventory Management
	TotalStock     int `gorm:"not null" json:"total_stock"`
	RemainingStock int `gorm:"not null" json:"remaining_stock"` // This will be decremented

	Version int `gorm:"default:0" json:"version"` // For Optimistic Locking (if needed later)

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
