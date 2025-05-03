package sale

import (
	"go.uber.org/zap"
)

// Service provides high-level user management operations on a LocalStorage backend.
type Service struct {
	// storage is the underlying persistence for User entities.
	storage Storage

	// logger is our observability component to log.
	logger *zap.Logger
}

// NewService crea un nuevo servicio de ventas
func NewService(storage Storage, logger *zap.Logger) *Service {
	if logger == nil {
		logger, _ = zap.NewProduction()
		defer logger.Sync()
	}

	return &Service{
		storage: storage,
		logger:  logger,
	}
}
