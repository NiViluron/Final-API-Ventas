package api

import (
	"Final-API-Ventas/internal/sale"
	"errors"
	"net/http"

	"go.uber.org/zap"

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

	s := &sale.Sale{
		UserID: req.UserID,
		Amount: req.Amount,
	}

	if err := h.saleService.Create(s); err != nil {
		if errors.Is(err, sale.ErrInvalidAmount) || errors.Is(err, sale.ErrInvalidUser) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

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

	s, err := h.saleService.Get(userID, status)

	if err != nil {
		if errors.Is(err, sale.ErrInvalidStatus) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
