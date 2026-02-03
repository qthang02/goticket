package main

import (
	"github.com/gin-gonic/gin"
	"github.com/qthang02/goticket/config"
	"github.com/qthang02/goticket/internal/infrastructure/mysql"
	"github.com/qthang02/goticket/internal/infrastructure/redis"
	"github.com/qthang02/goticket/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	// 1. Load Config
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	// 2. Init Logger
	logger.InitLogger()
	logger.Info("Starting GoTicket API", zap.String("env", cfg.App.Mode))

	// 3. Connect to Database
	db, err := mysql.NewMySQLDatabase(cfg)
	if err != nil {
		logger.Error("Database init failed", zap.Error(err))
		return
	}
	_ = db // In future, pass this to repository

	// 4. Connect to Redis
	rdb, err := redis.NewRedisClient(cfg)
	if err != nil {
		logger.Error("Redis init failed", zap.Error(err))
		return
	}
	_ = rdb // In future, pass this to repository

	// 5. Setup Router
	if cfg.App.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "UP",
			"db":     "connected", // basic check
			"redis":  "connected", // basic check
		})
	})

	// 6. Start Server
	logger.Info("Server listening", zap.String("port", cfg.App.Port))
	if err := r.Run(":" + cfg.App.Port); err != nil {
		logger.Error("Server failed to start", zap.Error(err))
	}
}
