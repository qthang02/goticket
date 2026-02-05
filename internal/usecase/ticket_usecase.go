package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/qthang02/goticket/internal/entity"
	"github.com/qthang02/goticket/internal/repository"
	"github.com/qthang02/goticket/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Lua Script for Atomic Inventory Check & Decrement
// Keys: [1] ticket_stock_key (e.g., ticket:{id}:stock)
// Args: [1] quantity
const luaCheckAndDecr = `
local stock = tonumber(redis.call("GET", KEYS[1]))
if not stock then return -1 end -- Cache Miss
if stock < tonumber(ARGV[1]) then return 0 end -- Sold Out
redis.call("DECRBY", KEYS[1], ARGV[1]) -- Decrement
return 1 -- Success
`

type TicketUsecase interface {
	BookTicket(ctx context.Context, userID, ticketID uint64, quantity int, idempotencyKey string) (*entity.Order, error)
}

type ticketUsecase struct {
	eventRepo  repository.EventRepository
	ticketRepo repository.TicketRepository
	orderRepo  repository.OrderRepository
	outboxRepo repository.OutboxRepository
	redis      *redis.Client
	db         *gorm.DB
}

func NewTicketUsecase(
	eventRepo repository.EventRepository,
	ticketRepo repository.TicketRepository,
	orderRepo repository.OrderRepository,
	outboxRepo repository.OutboxRepository,
	redis *redis.Client,
	db *gorm.DB,
) TicketUsecase {
	return &ticketUsecase{
		eventRepo:  eventRepo,
		ticketRepo: ticketRepo,
		orderRepo:  orderRepo,
		outboxRepo: outboxRepo,
		redis:      redis,
		db:         db,
	}
}

func (u *ticketUsecase) BookTicket(ctx context.Context, userID, ticketID uint64, quantity int, idempotencyKey string) (*entity.Order, error) {
	// 1. Input Validation
	if quantity <= 0 {
		return nil, errors.New("quantity must be greater than 0")
	}
	if idempotencyKey == "" {
		return nil, errors.New("idempotency key is required")
	}

	// 2. Idempotency Check (Redis Layer)
	idemKey := fmt.Sprintf("idempotency:%s", idempotencyKey)
	processed, err := u.redis.SetNX(ctx, idemKey, "processing", 1*time.Minute).Result()
	if err != nil {
		logger.Error("Redis idempotency check failed", zap.Error(err))
		return nil, err
	}
	if !processed {
		return nil, errors.New("duplicate request: processing or completed")
	}

	// 3. Distributed Lock (Mutex for TicketID)
	// We lock the specific TicketID to prevent too many concurrent requests race condition
	// although Lua script handles atomicity, lock helps reduce contention on Redis key.
	lockKey := fmt.Sprintf("lock:ticket:%d", ticketID)
	locked, err := u.redis.SetNX(ctx, lockKey, "locked", 5*time.Second).Result()
	if err != nil {
		logger.Error("Redis lock failed", zap.Error(err))
		return nil, err
	}
	if !locked {
		return nil, errors.New("system busy, please try again")
	}
	// Defer Release Lock
	defer u.redis.Del(ctx, lockKey)

	// 4. Atomic Inventory Check (Lua Script)
	stockKey := fmt.Sprintf("ticket:%d:stock", ticketID)

	// Execute Lua Script
	res, err := u.redis.Eval(ctx, luaCheckAndDecr, []string{stockKey}, quantity).Int()
	if err != nil {
		logger.Error("Lua script execution failed", zap.Error(err))
		return nil, err
	}

	// Handle Result
	if res == -1 {
		// Cache Miss: Lazy Load from DB
		ticket, err := u.ticketRepo.GetByID(ctx, ticketID)
		if err != nil {
			return nil, err
		}
		// Set stock to Redis (add TTL to prevent old data sticking forever)
		if err := u.redis.Set(ctx, stockKey, ticket.RemainingStock, 1*time.Hour).Err(); err != nil {
			logger.Error("Failed to warm up redis stock", zap.Error(err))
			return nil, err
		}
		// Retry Lua Script (Recursive call or strict linear retry?)
		// For simplicity, we just fail and ask user to retry, or we could recurse once.
		// Let's recurse once or just return busy.
		// To follow the flow safely: Return Busy.
		return nil, errors.New("system warming up, please retry")
	}

	if res == 0 {
		return nil, errors.New("sold out")
	}

	// If res == 1: Success decrement in Redis -> Proceed to DB

	// 5. Database Transaction (Atomic Order + Outbox)
	// 5. Database Transaction (Atomic Order + Outbox)
	var order *entity.Order

	err = u.db.Transaction(func(tx *gorm.DB) error {
		// 5.1 Create Order
		order = &entity.Order{
			UserID:         userID,
			TicketID:       ticketID,
			Quantity:       quantity,
			TotalAmount:    0,
			Status:         entity.OrderStatusPaid,
			IdempotencyKey: idempotencyKey,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		// Calculate Price (Safeish to read ticket here or use cached info)
		ticketInfo, err := u.ticketRepo.GetByID(ctx, ticketID)
		if err == nil {
			order.TotalAmount = ticketInfo.Price * float64(quantity)
		}

		// Note: We need a way to pass `tx` to repository or use it directly.
		// For this MVP, we assume repositories are tied to global DB connection inject, so they don't support TX easy without refactor.
		// So we use tx.Create directly for consistency.
		if err := tx.Create(order).Error; err != nil {
			return err
		}

		// 5.2 Create Outbox Event
		payload, _ := json.Marshal(map[string]interface{}{
			"order_id":  order.ID,
			"user_id":   order.UserID,
			"ticket_id": order.TicketID,
			"quantity":  order.Quantity,
		})

		outbox := &entity.Outbox{
			Topic:     "TicketBooked",
			Payload:   string(payload),
			Status:    entity.OutboxStatusPending,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := u.outboxRepo.Create(ctx, tx, outbox); err != nil {
			return err
		}

		// 5.3 Decrement Stock in DB (Consistent with Order)
		if err := tx.Model(&entity.Ticket{}).Where("id = ?", ticketID).UpdateColumn("remaining_stock", gorm.Expr("remaining_stock - ?", quantity)).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		logger.Error("Transaction failed", zap.Error(err))

		// Rollback Redis Stock (Compensating Transaction)
		u.redis.IncrBy(ctx, stockKey, int64(quantity))

		// Check for duplicate entry (MySQL Error 1062)
		if strings.Contains(err.Error(), "Duplicate entry") {
			return nil, errors.New("duplicate request: processing or completed")
		}

		return nil, errors.New("failed to process order")
	}

	// 6. Success
	logger.Info("Ticket booked successfully", zap.Uint64("order_id", order.ID))
	return order, nil
}
