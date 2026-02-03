package entity

import "time"

type Event struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Title       string    `gorm:"type:varchar(255);not null" json:"title"`
	Description string    `gorm:"type:text" json:"description"`
	Location    string    `gorm:"type:varchar(255)" json:"location"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`

	// Has many Tickets (types/categories)
	Tickets []Ticket `gorm:"foreignKey:EventID" json:"tickets,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
