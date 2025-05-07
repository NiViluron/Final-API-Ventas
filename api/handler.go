package api

import (
	"Final-API-Ventas/internal/sale"
	"errors"
	"net/http"
	"time"

	"go.uber.org/zap"
	"resty.dev/v3"

	"github.com/gin-gonic/gin"
)

// handler holds the user service and implements HTTP handlers for user CRUD.
type handler struct {
	saleService *sale.Service
	logger      *zap.Logger
}

// handleCreate handles POST /sales
func (h *handler) handleCreateSale(ctx *gin.Context) {
	// request payload
	var req struct {
		UserID string  `json:"user_id"`
		Amount float64 `json:"amount"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validar que amount no sea cero
	if req.Amount == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "El monto no puede ser cero"})
		return
	}

	// Validar que user exista (utilizamos API user)
	userID := req.UserID
	client := resty.New()
	defer client.Close()

	res, err := client.R().
		EnableTrace().
		Get("http://localhost:8080/users/" + userID)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error en la consulta de usuario"})
		h.logger.Error("error trying to get user", zap.Error(err))
		return
	}

	if res.IsError() {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Usuario no encontrado"})
		h.logger.Warn("user not found", zap.String("id", userID))
		return
	}

	// Asignar estado aleatorio
	estados := []string{"pending", "approved", "rejected"}
	randomIndex := time.Now().UnixNano() % int64(len(estados))
	status := estados[randomIndex]

	s := &sale.Sale{
		UserID: req.UserID,
		Amount: req.Amount,
		Status: status,
	}

	if err := h.saleService.Create(s); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("sale created", zap.Any("sale", s))
	ctx.JSON(http.StatusCreated, s)
}

// handlePatchSales handles /sales/:id
func (h *handler) handlePatchSales(ctx *gin.Context) {
	id := ctx.Param("id")

	var fields *sale.UpdateFields
	if err := ctx.ShouldBindJSON(&fields); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if *fields.Status != "approved" && *fields.Status != "rejected" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": errors.New("the new status must be 'approved' or 'rejected'").Error()})
		return
	}

	s, err := h.saleService.Update(id, *fields.Status)
	if err != nil {
		if errors.Is(err, sale.ErrNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, sale.ErrInvalidTransition) {
			ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
	}
	ctx.JSON(http.StatusOK, s)
}

func (h *handler) handleReadSale(ctx *gin.Context) {
	userID := ctx.Query("user_id")
	status := ctx.Query("status")

	if status != "" && status != "pending" && status != "approved " && status != "rejected" {
		h.logger.Warn("invalid status value", zap.String("status", status))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status value"})
		return
	}

	client := resty.New()
	defer client.Close()

	res, err := client.R().
		EnableTrace().
		Get("http://localhost:8080/users/" + userID)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error en la consulta de usuario"})
		h.logger.Error("error trying to get user", zap.Error(err))
		return
	}

	if res.IsError() {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Usuario no encontrado"})
		h.logger.Warn("user not found", zap.String("id", userID))
		return
	}

	s, err := h.saleService.Get(userID, status)

	if err != nil {
		if errors.Is(err, sale.ErrNotFound) {
			h.logger.Warn("sale not found", zap.String("id", userID))
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		h.logger.Error("error trying to get sale", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("get sale succeed", zap.Any("sale", s))

	// Calculate total amount and count sales by status
	totalAmount := 0.0
	approvedCount := 0
	rejectedCount := 0
	pendingCount := 0

	for _, sale := range s {
		switch sale.Status {
		case "approved":
			approvedCount++
		case "rejected":
			rejectedCount++
		case "pending":
			pendingCount++
		}
		totalAmount += sale.Amount
	}

	finalRes := map[string]interface{}{
		"metadata": map[string]interface{}{
			"quantity":     len(s),
			"approved":     approvedCount,
			"rejected":     rejectedCount,
			"pending":      pendingCount,
			"total_amount": totalAmount,
		},
		"results": s,
	}

	ctx.JSON(http.StatusOK, finalRes)
}
