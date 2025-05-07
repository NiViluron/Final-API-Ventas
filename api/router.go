package api

import (
	"Final-API-Ventas/internal/sale"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// InitRoutes registers all user CRUD endpoints on the given GIN engine.
// It initializes the storage, service, and handler, then binds each HTTP
// method and path to the appropiate handler function.
func InitRoutes(e *gin.Engine) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	storageSale := sale.NewLocalStorage()
	serviceSale := sale.NewService(storageSale, logger)

	h := handler{
		saleService: serviceSale,
		logger:      logger,
	}

	e.POST("/sales", h.handleCreateSale)
	e.PATCH("/sales/:id", h.handlePatchSales)
	e.GET("/sales", h.handleReadSale)
}
