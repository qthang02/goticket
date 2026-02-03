package mysql

import (
	"time"

	"github.com/qthang02/goticket/config"
	"github.com/qthang02/goticket/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func NewMySQLDatabase(cfg *config.Config) (*gorm.DB, error) {

	logger.Info("Connecting to MySQL database", zap.String("dsn", cfg.Database.Dsn))

	db, err := gorm.Open(mysql.Open(cfg.Database.Dsn), &gorm.Config{})
	if err != nil {
		logger.Error("Failed to connect to database", zap.Error(err))
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdle)
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpen)
	sqlDB.SetConnMaxLifetime(time.Hour)

	logger.Info("Connected to MySQL database")
	return db, nil
}
