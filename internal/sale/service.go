package sale

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

var ErrInvalidTransition = errors.New("actual status is not pending")
var ErrEmptyStatus = errors.New("new status is empty")

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

func (s *Service) Create(sale *Sale) error {
	sale.ID = uuid.NewString()
	now := time.Now()
	sale.CreateAt = now
	sale.UpdateAt = now
	sale.Version = 1

	return s.storage.Set(sale)
}

func (s *Service) Update(id string, newStatus string) (*Sale, error) {
	if newStatus == "" {
		return nil, ErrEmptyStatus
	}
	existing, err := s.storage.Read(id)
	if err != nil {
		return nil, err
	}
	if existing.Status != "pending" {
		return nil, ErrInvalidTransition
	}

	existing.Status = newStatus
	existing.UpdateAt = time.Now()
	existing.Version++
	return existing, nil
}
func (s *Service) Get(user_id string, status string) ([]*Sale, error) {
	return s.storage.GetSales(user_id, status)
}
