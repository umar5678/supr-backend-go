package documentdto

type UploadDocumentRequest struct {
	DocumentType string `form:"documentType" binding:"required"` // license, aadhaar, registration, insurance, trade-license, profile-photo
	FileName     string `form:"fileName" binding:"required"`
	FileSize     int64  `form:"fileSize" binding:"required"`
	IsFront      bool   `form:"isFront"` // For dual documents (license front/back)
	// File is uploaded as multipart form
}

type VerifyDocumentRequest struct {
	DocumentID      string `json:"documentId" binding:"required"`
	Status          string `json:"status" binding:"required"`      // verified, rejected
	RejectionReason string `json:"rejectionReason"`               // Required if status is rejected
}

type ListDocumentsRequest struct {
	UserID           string `query:"userId"`
	DriverID         string `query:"driverId"`
	ServiceProviderID string `query:"serviceProviderId"`
	DocumentType     string `query:"documentType"`
	Status           string `query:"status"` // pending, verified, rejected, expired
	Page             int    `query:"page" binding:"min=1"`
	Limit            int    `query:"limit" binding:"min=1,max=100"`
}

func (r *ListDocumentsRequest) SetDefaults() {
	if r.Page == 0 {
		r.Page = 1
	}
	if r.Limit == 0 {
		r.Limit = 20
	}
}
