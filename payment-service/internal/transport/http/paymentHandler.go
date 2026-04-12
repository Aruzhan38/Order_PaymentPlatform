package http

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	nethttp "net/http"
	"payment-service/internal/usecase"
)

type PaymentHandler struct {
	uc *usecase.PaymentUsecase
}

func NewPaymentHandler(uc *usecase.PaymentUsecase) *PaymentHandler {
	return &PaymentHandler{uc: uc}
}

type createPaymentRequest struct {
	OrderID string `json:"order_id"`
	Amount  int64  `json:"amount"`
}

func (h *PaymentHandler) CreatePayment(c *gin.Context) {
	var req createPaymentRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(nethttp.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	payment, err := h.uc.CreatePayment(c.Request.Context(), req.OrderID, req.Amount)
	if err != nil {
		c.JSON(nethttp.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(nethttp.StatusCreated, gin.H{
		"id":             payment.ID,
		"order_id":       payment.OrderID,
		"transaction_id": payment.TransactionID,
		"amount":         payment.Amount,
		"status":         payment.Status,
	})
}

func (h *PaymentHandler) GetPayment(c *gin.Context) {
	orderID := c.Param("order_id")

	payment, err := h.uc.GetPaymentByOrderID(c.Request.Context(), orderID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(nethttp.StatusNotFound, gin.H{"error": "payment not found"})
			return
		}
		c.JSON(nethttp.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(nethttp.StatusOK, payment)
}
