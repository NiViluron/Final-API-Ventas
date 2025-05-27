package sale

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"resty.dev/v3"
)

var ErrInvalidTransition = errors.New("actual status is not pending")
var ErrEmptyStatus = errors.New("new status is empty")
var ErrInvalidAmount = errors.New("amount cannot be zero")
var ErrInvalidUser = errors.New("user not found")
var ErrInvalidStatus = errors.New("status must be 'pending', 'approved' or 'rejected'")

// Service provides high-level user management operations on a LocalStorage backend.
type Service struct {
	// storage is the underlying persistence for User entities.
	storage Storage
	// logger is our observability component to log.
	logger *zap.Logger
	// userClient is a Resty client to interact with the user service.
	userClient *resty.Client
	urlUser    string
}

// NewService crea un nuevo servicio de ventas
func NewService(storage Storage, logger *zap.Logger, urlUser string) *Service {
	if logger == nil {
		logger, _ = zap.NewProduction()
		defer logger.Sync()
	}

	restyClient := resty.New()
	defer restyClient.Close()

	return &Service{
		storage:    storage,
		logger:     logger,
		userClient: restyClient,
		urlUser:    urlUser,
	}
}

func (s *Service) Create(sale *Sale) error {
	sale.ID = uuid.NewString()
	now := time.Now()
	sale.CreateAt = now
	sale.UpdateAt = now
	sale.Version = 1

	// Validar que user exista (utilizamos API user)
	userID := sale.UserID

	res, err := s.userClient.R().
		EnableTrace().
		Get(s.urlUser + "/users/" + userID)

	if err != nil {
		s.logger.Error("error trying to get user", zap.Error(err))
		return err
	}

	if res.IsError() {
		s.logger.Warn("user not found", zap.String("id", userID))
		return ErrInvalidUser
	}

	// Validar que amount no sea cero
	if sale.Amount == 0 {
		s.logger.Warn("amount cannot be zero", zap.Float64("amount", sale.Amount))
		return ErrInvalidAmount
	}

	// Asignar estado aleatorio
	estados := []string{"pending", "approved", "rejected"}
	randomIndex := time.Now().UnixNano() % int64(len(estados))
	status := estados[randomIndex]

	sale.Status = status

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
	if status != "" && status != "pending" && status != "approved" && status != "rejected" {
		s.logger.Warn("invalid status value", zap.String("status", status))
		return nil, ErrInvalidStatus
	}

	res, err := s.userClient.R().
		EnableTrace().
		Get(s.urlUser + "/users/" + user_id)

	if err != nil {
		s.logger.Error("error trying to get user", zap.Error(err))
		return nil, err
	}

	if res.IsError() {
		s.logger.Warn("user not found", zap.String("id", user_id))
		return nil, ErrInvalidUser
	}

	return s.storage.GetSales(user_id, status)
}
