package http

import (
	"github.com/gin-gonic/gin"
	"github.com/qthang02/goticket/internal/transport/http/handler"
)

func NewRouter(r *gin.Engine, ticketHandler *handler.TicketHandler) {
	api := r.Group("/api/v1")
	{
		// Events
		events := api.Group("/events")
		{
			// events.GET("", ...)
			events.POST("/:id/book", ticketHandler.BookTicket)
		}
	}
}
