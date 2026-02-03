package main

import (
	"log"
	"time"

	"github.com/qthang02/goticket/config"
	"github.com/qthang02/goticket/internal/entity"
	"github.com/qthang02/goticket/internal/infrastructure/mysql"
	"github.com/qthang02/goticket/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	// 1. Load Config
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	// 2. Init Logger
	logger.InitLogger()

	// 3. Connect to MySQL
	db, err := mysql.NewMySQLDatabase(cfg)
	if err != nil {
		logger.Error("Database init failed", zap.Error(err))
		return
	}

	// 4. Auto Migrate (Quick schema creation for dev)
	logger.Info("Running AutoMigrate...")
	err = db.AutoMigrate(&entity.User{}, &entity.Event{}, &entity.Ticket{}, &entity.Order{})
	if err != nil {
		logger.Error("Migration failed", zap.Error(err))
		return
	}

	// 5. Seed Data
	logger.Info("Seeding data...")

	// Create Event
	event := entity.Event{
		Title:       "Taylor Swift - The Eras Tour",
		Description: "The biggest concert of the year.",
		Location:    "My Dinh Stadium, Hanoi",
		StartTime:   time.Now().Add(24 * time.Hour),
		EndTime:     time.Now().Add(28 * time.Hour),
	}
	if err := db.Create(&event).Error; err != nil {
		logger.Error("Failed to seed event", zap.Error(err))
		// Don't return, maybe it already exists
	}

	// Create Tickets
	tickets := []entity.Ticket{
		{
			EventID:        event.ID,
			Name:           "VIP Zone A",
			Price:          500.00,
			TotalStock:     100,
			RemainingStock: 100, // Important for testing concurrency
		},
		{
			EventID:        event.ID,
			Name:           "General Admission",
			Price:          100.00,
			TotalStock:     1000,
			RemainingStock: 1000,
		},
	}
	if err := db.Create(&tickets).Error; err != nil {
		logger.Error("Failed to seed tickets", zap.Error(err))
	}

	// Create Dummy Users
	users := []entity.User{
		{Name: "Alice", Email: "alice@example.com", Password: "hashed_password"},
		{Name: "Bob", Email: "bob@example.com", Password: "hashed_password"},
		{Name: "Charlie", Email: "charlie@example.com", Password: "hashed_password"},
	}
	if err := db.Create(&users).Error; err != nil {
		logger.Error("Failed to seed users", zap.Error(err))
	}

	logger.Info("Seeding completed successfully!", zap.Uint64("event_id", event.ID))
}
