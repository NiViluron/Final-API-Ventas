package api

import (
	"Final-API-Ventas/internal/sale"
	"Final-API-Ventas/internal/user"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// InitRoutes registers all user CRUD endpoints on the given GIN engine.
// It initializes the storage, service, and handler, then binds each HTTP
// method and path to the appropiate handler function.
func InitRoutes(e *gin.Engine) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	storageUser := user.NewLocalStorage()
	serviceUser := user.NewService(storageUser, logger)

	storageSale := sale.NewLocalStorage()
	serviceSale := sale.NewService(storageSale, logger)

	h := handler{
		userService: serviceUser,
		saleService: serviceSale,
		logger:      logger,
	}

	e.POST("/users", h.handleCreate)
	e.GET("/users/:id", h.handleRead)
	e.PATCH("/users/:id", h.handleUpdate)
	e.DELETE("/users/:id", h.handleDelete)
	e.POST("/sales", h.handleCreateSale)
	e.GET("/sales", h.handleReadSale)
}
