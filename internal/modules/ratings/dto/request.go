package dto

type CreateRatingRequest struct {
	OrderID string  `json:"orderId" binding:"required"`
	Score   int     `json:"score" binding:"required,min=1,max=5"`
	Comment *string `json:"comment" binding:"omitempty,max=500"`
}
