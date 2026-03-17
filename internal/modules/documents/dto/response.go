package documentdto

type DocumentResponse struct {
	ID                string  `json:"id"`
	UserID            string  `json:"userId"`
	DriverID          *string `json:"driverId,omitempty"`
	ServiceProviderID *string `json:"serviceProviderId,omitempty"`
	DocumentType      string  `json:"documentType"`
	FileName          string  `json:"fileName"`
	FileURL           string  `json:"fileUrl"`
	FileSize          int64   `json:"fileSize"`
	MimeType          string  `json:"mimeType"`
	Status            string  `json:"status"` 
	VerifiedBy        *string `json:"verifiedBy,omitempty"`
	VerifiedAt        *string `json:"verifiedAt,omitempty"`
	RejectionReason   string  `json:"rejectionReason,omitempty"`
	ExpiryDate        *string `json:"expiryDate,omitempty"`
	IsFront           bool    `json:"isFront,omitempty"`
	UploadedAt        string  `json:"uploadedAt"`
}

type DocumentListResponse struct {
	Documents []*DocumentResponse `json:"documents"`
	Total     int64               `json:"total"`
	Page      int                 `json:"page"`
	Limit     int                 `json:"limit"`
}

type VerifyDocumentResponse struct {
	DocumentID      string `json:"documentId"`
	Status          string `json:"status"`
	RejectionReason string `json:"rejectionReason,omitempty"`
	VerifiedAt      string `json:"verifiedAt"`
	Message         string `json:"message"`
}

type DocumentUploadTokenResponse struct {
	Token       string `json:"token"`
	Expire      int64  `json:"expire"`
	PublicKey   string `json:"publicKey"`
	URLEndpoint string `json:"urlEndpoint"`
	Folder      string `json:"folder"`
}
