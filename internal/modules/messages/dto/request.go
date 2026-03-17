package dto


type SendMessageRequest struct {
	RideID   string                 `json:"rideId" binding:"required"`
	Content  string                 `json:"content" binding:"required"`
	Metadata map[string]interface{} `json:"metadata"`
}