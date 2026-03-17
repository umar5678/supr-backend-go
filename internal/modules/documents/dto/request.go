package documentdto

type UploadDocumentRequest struct {
	DocumentType string `form:"documentType" binding:"required"`
	FileName     string `form:"fileName" binding:"required"`
	FileSize     int64  `form:"fileSize" binding:"required"`
	IsFront      bool   `form:"isFront"`
}

type VerifyDocumentRequest struct {
	DocumentID      string `json:"documentId" binding:"required"`
	Status          string `json:"status" binding:"required"`
	RejectionReason string `json:"rejectionReason"`          
}

type ListDocumentsRequest struct {
	UserID            string `query:"userId"`
	DriverID          string `query:"driverId"`
	ServiceProviderID string `query:"serviceProviderId"`
	DocumentType      string `query:"documentType"`
	Status            string `query:"status"`
	Page              int    `query:"page" binding:"min=1"`
	Limit             int    `query:"limit" binding:"min=1,max=100"`
}

func (r *ListDocumentsRequest) SetDefaults() {
	if r.Page == 0 {
		r.Page = 1
	}
	if r.Limit == 0 {
		r.Limit = 20
	}
}
