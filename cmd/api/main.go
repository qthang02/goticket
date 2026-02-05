package main

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/qthang02/goticket/config"
	"github.com/qthang02/goticket/internal/infrastructure/kafka"
	"github.com/qthang02/goticket/internal/infrastructure/mysql"
	"github.com/qthang02/goticket/internal/infrastructure/redis"
	"github.com/qthang02/goticket/internal/repository"
	transportHttp "github.com/qthang02/goticket/internal/transport/http"
	"github.com/qthang02/goticket/internal/transport/http/handler"
	"github.com/qthang02/goticket/internal/usecase"
	"github.com/qthang02/goticket/internal/worker"
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

	// 4. Connect to Redis
	rdb, err := redis.NewRedisClient(cfg)
	if err != nil {
		logger.Error("Redis init failed", zap.Error(err))
		return
	}

	// 5. Init Kafka Producer
	kafkaProducer := kafka.NewProducer(cfg)
	defer kafkaProducer.Close()

	// Ensure Topic Exists
	if err := kafkaProducer.EnsureTopic(cfg.Kafka.Topic); err != nil {
		logger.Error("Failed to ensure kafka topic", zap.Error(err))
		// Continue anyway, maybe it exists or auto-create works
	} else {
		logger.Info("Kafka topic ensured", zap.String("topic", cfg.Kafka.Topic))
	}

	// 6. Init Repositories
	eventRepo := repository.NewEventRepository(db)
	ticketRepo := repository.NewTicketRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	outboxRepo := repository.NewOutboxRepository(db)
	// userRepo := repository.NewUserRepository(db)

	// 7. Init Usecases
	ticketUsecase := usecase.NewTicketUsecase(eventRepo, ticketRepo, orderRepo, outboxRepo, rdb, db)

	// 8. Init Handlers
	ticketHandler := handler.NewTicketHandler(ticketUsecase)

	// 9. Init & Start Worker
	outboxWorker := worker.NewOutboxWorker(outboxRepo, kafkaProducer)
	go outboxWorker.Start(context.Background())

	// 10. Setup Router
	if cfg.App.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger()) // Add logger middleware

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "UP",
			"db":     "connected",
			"redis":  "connected",
			"kafka":  "connected",
		})
	})

	// Register Routes
	transportHttp.NewRouter(r, ticketHandler)

	// 9. Start Server
	logger.Info("Server listening", zap.String("port", cfg.App.Port))
	if err := r.Run(":" + cfg.App.Port); err != nil {
		logger.Error("Server failed to start", zap.Error(err))
	}
}
