package api

import (
	"Final-API-Ventas/internal/sale"
	"Final-API-Ventas/internal/user"
	"errors"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
)

// handler holds the user service and implements HTTP handlers for user CRUD.
type handler struct {
	saleService *sale.Service
	userService *user.Service
	logger      *zap.Logger
}

// handleCreate handles POST /users
func (h *handler) handleCreate(ctx *gin.Context) {
	// request payload
	var req struct {
		Name     string `json:"name"`
		Address  string `json:"address"`
		NickName string `json:"nickname"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	u := &user.User{
		Name:     req.Name,
		Address:  req.Address,
		NickName: req.NickName,
	}
	if err := h.userService.Create(u); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("user created", zap.Any("user", u))
	ctx.JSON(http.StatusCreated, u)
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

	// Validar que user exista (utilizamos userService)
	userID := req.UserID
	_, err := h.userService.Get(userID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Usuario no encontrado"})
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

// handleRead handles GET /users/:id
func (h *handler) handleRead(ctx *gin.Context) {
	id := ctx.Param("id")

	u, err := h.userService.Get(id)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			h.logger.Warn("user not found", zap.String("id", id))
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		h.logger.Error("error trying to get user", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("get user succeed", zap.Any("user", u))
	ctx.JSON(http.StatusOK, u)
}

// handleUpdate handles PUT /users/:id
func (h *handler) handleUpdate(ctx *gin.Context) {
	id := ctx.Param("id")

	// bind partial update fields
	var fields *user.UpdateFields
	if err := ctx.ShouldBindJSON(&fields); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	u, err := h.userService.Update(id, fields)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, u)
}

// handleDelete handles DELETE /users/:id
func (h *handler) handleDelete(ctx *gin.Context) {
	id := ctx.Param("id")

	if err := h.userService.Delete(id); err != nil {
		if errors.Is(err, user.ErrNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (h *handler) handleReadSale(ctx *gin.Context) {
	userID := ctx.Query("user_id")
	status := ctx.Query("status")

	if status != "" && status != "pending" && status != "approved " && status != "rejected" {
		h.logger.Warn("invalid status value", zap.String("status", status))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status value"})
		return
	}

	u, err := h.userService.Get(userID)

	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			h.logger.Warn("user not found", zap.String("id", userID))
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		h.logger.Error("error trying to get user", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	s, err := h.saleService.Get(u.ID, status)

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
