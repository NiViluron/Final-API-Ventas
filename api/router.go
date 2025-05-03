package api

import (
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

	storage := user.NewLocalStorage()
	service := user.NewService(storage, logger)

	hUser := handlerUser{
		userService: service,
		logger:      logger,
	}

	e.POST("/users", hUser.handleCreate)
	e.GET("/users/:id", hUser.handleRead)
	e.PATCH("/users/:id", hUser.handleUpdate)
	e.DELETE("/users/:id", hUser.handleDelete)
}
