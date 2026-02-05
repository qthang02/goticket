package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/qthang02/goticket/internal/usecase"
)

type TicketHandler struct {
	usecase usecase.TicketUsecase
}

func NewTicketHandler(usecase usecase.TicketUsecase) *TicketHandler {
	return &TicketHandler{usecase: usecase}
}

type bookTicketRequest struct {
	UserID   uint64 `json:"user_id" binding:"required"`
	TicketID uint64 `json:"ticket_id" binding:"required"`
	Quantity int    `json:"quantity" binding:"required,min=1"`
}

func (h *TicketHandler) BookTicket(c *gin.Context) {
	var req bookTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	idempotencyKey := c.GetHeader("X-Idempotency-Key")
	if idempotencyKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "X-Idempotency-Key header is required"})
		return
	}

	order, err := h.usecase.BookTicket(c.Request.Context(), req.UserID, req.TicketID, req.Quantity, idempotencyKey)
	if err != nil {
		// Map domain errors to HTTP Status codes
		// For simplicity, we return 400 or 500 based on error message text or define custom errors.
		// In a real app, use typed errors.
		switch err.Error() {
		case "sold out":
			c.JSON(http.StatusBadRequest, gin.H{"error": "sold_out", "message": "Ticket is out of stock"})
		case "system busy, please try again", "duplicate request: processing or completed":
			c.JSON(http.StatusConflict, gin.H{"error": "concurrency_limit", "message": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "ticket booked successfully",
		"order_id": order.ID,
		"status":   order.Status,
	})
}
