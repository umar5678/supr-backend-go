package wallet

import (
	"github.com/gin-gonic/gin"
	"github.com/umar5678/go-backend/internal/modules/wallet/dto"
	"github.com/umar5678/go-backend/internal/utils/response"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// GetWallet godoc
// @Summary Get wallet details
// @Tags wallet
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response{data=dto.WalletResponse}
// @Router /wallet [get]
func (h *Handler) GetWallet(c *gin.Context) {
	userID, _ := c.Get("userID")

	wallet, err := h.service.GetWallet(c.Request.Context(), userID.(string))
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, wallet, "Wallet retrieved successfully")
}

// GetBalance godoc
// @Summary Get wallet balance
// @Tags wallet
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response{data=float64}
// @Router /wallet/balance [get]
func (h *Handler) GetBalance(c *gin.Context) {
	userID, _ := c.Get("userID")

	balance, err := h.service.GetBalance(c.Request.Context(), userID.(string))
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, map[string]interface{}{
		"balance": balance,
	}, "Balance retrieved successfully")
}

// AddFunds godoc
// @Summary Add funds to wallet
// @Tags wallet
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.AddFundsRequest true "Add funds data"
// @Success 200 {object} response.Response{data=dto.TransactionResponse}
// @Router /wallet/add-funds [post]
func (h *Handler) AddFunds(c *gin.Context) {
	userID, _ := c.Get("userID")

	var req dto.AddFundsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body"))
		return
	}

	transaction, err := h.service.AddFunds(c.Request.Context(), userID.(string), req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, transaction, "Funds added successfully")
}

// WithdrawFunds godoc
// @Summary Withdraw funds from wallet
// @Tags wallet
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.WithdrawFundsRequest true "Withdraw funds data"
// @Success 200 {object} response.Response{data=dto.TransactionResponse}
// @Router /wallet/withdraw [post]
func (h *Handler) WithdrawFunds(c *gin.Context) {
	userID, _ := c.Get("userID")

	var req dto.WithdrawFundsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body"))
		return
	}

	transaction, err := h.service.WithdrawFunds(c.Request.Context(), userID.(string), req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, transaction, "Funds withdrawn successfully")
}

// TransferFunds godoc
// @Summary Transfer funds to another user
// @Tags wallet
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.TransferFundsRequest true "Transfer data"
// @Success 200 {object} response.Response{data=dto.TransactionResponse}
// @Router /wallet/transfer [post]
func (h *Handler) TransferFunds(c *gin.Context) {
	userID, _ := c.Get("userID")

	var req dto.TransferFundsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body"))
		return
	}

	transaction, err := h.service.TransferFunds(c.Request.Context(), userID.(string), req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, transaction, "Funds transferred successfully")
}

// ListTransactions godoc
// @Summary List wallet transactions
// @Tags wallet
// @Security BearerAuth
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Param type query string false "Transaction type"
// @Param status query string false "Transaction status"
// @Success 200 {object} response.Response{data=[]dto.TransactionResponse}
// @Router /wallet/transactions [get]
func (h *Handler) ListTransactions(c *gin.Context) {
	userID, _ := c.Get("userID")

	var req dto.ListTransactionsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.Error(response.BadRequest("Invalid query parameters"))
		return
	}

	transactions, total, err := h.service.ListTransactions(c.Request.Context(), userID.(string), req)
	if err != nil {
		c.Error(err)
		return
	}

	pagination := response.NewPaginationMeta(total, req.Page, req.Limit)
	response.Paginated(c, transactions, pagination, "Transactions retrieved successfully")
}

// GetTransaction godoc
// @Summary Get transaction details
// @Tags wallet
// @Security BearerAuth
// @Produce json
// @Param id path string true "Transaction ID"
// @Success 200 {object} response.Response{data=dto.TransactionResponse}
// @Router /wallet/transactions/{id} [get]
func (h *Handler) GetTransaction(c *gin.Context) {
	userID, _ := c.Get("userID")
	txID := c.Param("id")

	transaction, err := h.service.GetTransaction(c.Request.Context(), userID.(string), txID)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, transaction, "Transaction retrieved successfully")
}

// HoldFunds godoc
// @Summary Place a hold on funds
// @Tags wallet
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.HoldFundsRequest true "Hold funds data"
// @Success 200 {object} response.Response{data=dto.HoldResponse}
// @Router /wallet/hold [post]
func (h *Handler) HoldFunds(c *gin.Context) {
	userID, _ := c.Get("userID")

	var req dto.HoldFundsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body"))
		return
	}

	hold, err := h.service.HoldFunds(c.Request.Context(), userID.(string), req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, hold, "Funds held successfully")
}

// ReleaseHold godoc
// @Summary Release a hold
// @Tags wallet
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.ReleaseHoldRequest true "Hold ID"
// @Success 200 {object} response.Response
// @Router /wallet/hold/release [post]
func (h *Handler) ReleaseHold(c *gin.Context) {
	userID, _ := c.Get("userID")

	var req dto.ReleaseHoldRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body"))
		return
	}

	if err := h.service.ReleaseHold(c.Request.Context(), userID.(string), req); err != nil {
		c.Error(err)
		return
	}

	response.Success(c, nil, "Hold released successfully")
}

// CaptureHold godoc
// @Summary Capture a hold
// @Tags wallet
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.CaptureHoldRequest true "Capture hold data"
// @Success 200 {object} response.Response{data=dto.TransactionResponse}
// @Router /wallet/hold/capture [post]
func (h *Handler) CaptureHold(c *gin.Context) {
	userID, _ := c.Get("userID")

	var req dto.CaptureHoldRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body"))
		return
	}

	transaction, err := h.service.CaptureHold(c.Request.Context(), userID.(string), req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, transaction, "Hold captured successfully")
}
