package documents

import (
	"github.com/gin-gonic/gin"
	documentdto "github.com/umar5678/go-backend/internal/modules/documents/dto"
	"github.com/umar5678/go-backend/internal/utils/response"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// UploadDocument godoc
// @Summary Upload verification document
// @Tags documents
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param documentType formData string true "Document type (license, aadhaar, registration, insurance, etc.)"
// @Param fileName formData string true "File name"
// @Param file formData file true "Document file (PDF, JPG, PNG, WebP)"
// @Param isFront formData boolean false "Is front side (for dual documents)"
// @Success 200 {object} response.Response{data=documentdto.DocumentResponse}
// @Description Users can reupload documents after rejection. Each upload creates a new document record. Admin can verify the latest document, and once all required documents are verified, the profile is marked as verified.
// @Router /documents/upload [post]
func (h *Handler) UploadDocument(c *gin.Context) {
	userID, _ := c.Get("userID")

	documentType := c.PostForm("documentType")
	if documentType == "" {
		c.Error(response.BadRequest("Document type is required"))
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.Error(response.BadRequest("File is required"))
		return
	}

	doc, err := h.service.UploadDocument(c.Request.Context(), userID.(string), documentType, file)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, doc, "Document uploaded successfully")
}

// GetDocuments godoc
// @Summary Get user documents
// @Tags documents
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response{data=[]documentdto.DocumentResponse}
// @Router /documents [get]
func (h *Handler) GetDocuments(c *gin.Context) {
	userID, _ := c.Get("userID")

	docs, err := h.service.GetDocuments(c.Request.Context(), userID.(string))
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, docs, "Documents retrieved successfully")
}

// GetDocumentsByDriver godoc
// @Summary Get driver documents (admin only)
// @Tags documents
// @Security BearerAuth
// @Produce json
// @Param driverId query string true "Driver ID"
// @Success 200 {object} response.Response{data=[]documentdto.DocumentResponse}
// @Router /documents/driver [get]
func (h *Handler) GetDocumentsByDriver(c *gin.Context) {
	driverID := c.Query("driverId")
	if driverID == "" {
		c.Error(response.BadRequest("Driver ID is required"))
		return
	}

	docs, err := h.service.GetDocumentsByDriver(c.Request.Context(), driverID)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, docs, "Driver documents retrieved successfully")
}

// GetDocumentsByServiceProvider godoc
// @Summary Get service provider documents (admin only)
// @Tags documents
// @Security BearerAuth
// @Produce json
// @Param serviceProviderId query string true "Service Provider ID"
// @Success 200 {object} response.Response{data=[]documentdto.DocumentResponse}
// @Router /documents/service-provider [get]
func (h *Handler) GetDocumentsByServiceProvider(c *gin.Context) {
	providerID := c.Query("serviceProviderId")
	if providerID == "" {
		c.Error(response.BadRequest("Service Provider ID is required"))
		return
	}

	docs, err := h.service.GetDocumentsByServiceProvider(c.Request.Context(), providerID)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, docs, "Service provider documents retrieved successfully")
}

// GetDocumentsAdmin godoc
// @Summary Get all documents with filters (admin only)
// @Tags documents
// @Security BearerAuth
// @Produce json
// @Param userId query string false "User ID"
// @Param driverId query string false "Driver ID"
// @Param documentType query string false "Document type"
// @Param status query string false "Status (pending, verified, rejected)"
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Success 200 {object} response.Response{data=documentdto.DocumentListResponse}
// @Router /documents/admin [get]
func (h *Handler) GetDocumentsAdmin(c *gin.Context) {
	req := &documentdto.ListDocumentsRequest{}
	if err := c.ShouldBindQuery(req); err != nil {
		c.Error(response.BadRequest("Invalid query parameters"))
		return
	}

	req.SetDefaults()

	filters := map[string]interface{}{}
	if req.UserID != "" {
		filters["user_id"] = req.UserID
	}
	if req.DriverID != "" {
		filters["driver_id"] = req.DriverID
	}
	if req.ServiceProviderID != "" {
		filters["service_provider_id"] = req.ServiceProviderID
	}
	if req.DocumentType != "" {
		filters["document_type"] = req.DocumentType
	}
	if req.Status != "" {
		filters["status"] = req.Status
	}

	listResp, err := h.service.GetDocumentsPaginated(c.Request.Context(), filters, req.Page, req.Limit)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, listResp, "Documents retrieved successfully")
}

// VerifyDocument godoc
// @Summary Verify/reject document (admin only)
// @Tags documents
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body documentdto.VerifyDocumentRequest true "Verification request"
// @Success 200 {object} response.Response{data=documentdto.VerifyDocumentResponse}
// @Description When a document is verified, the system checks if all required documents are verified. If so, the user's profile (driver or service provider) is marked as verified. When a document is rejected, the profile is marked as not verified and the user can reupload a new document.
// @Router /documents/verify [post]
func (h *Handler) VerifyDocument(c *gin.Context) {
	adminID, _ := c.Get("userID")

	var req documentdto.VerifyDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body"))
		return
	}

	verifyResp, err := h.service.VerifyDocument(c.Request.Context(), adminID.(string), req.DocumentID, req.Status, req.RejectionReason)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, verifyResp, "Document verified successfully")
}

// DeleteDocument godoc
// @Summary Delete document
// @Tags documents
// @Security BearerAuth
// @Produce json
// @Param documentId query string true "Document ID"
// @Success 200 {object} response.Response
// @Router /documents/:id [delete]
func (h *Handler) DeleteDocument(c *gin.Context) {
	docID := c.Param("id")
	if docID == "" {
		c.Error(response.BadRequest("Document ID is required"))
		return
	}

	if err := h.service.DeleteDocument(c.Request.Context(), docID); err != nil {
		c.Error(err)
		return
	}

	response.Success(c, nil, "Document deleted successfully")
}

// GetPendingDocuments godoc
// @Summary Get pending documents awaiting verification (admin only)
// @Tags documents
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response{data=[]documentdto.DocumentResponse}
// @Router /documents/pending [get]
func (h *Handler) GetPendingDocuments(c *gin.Context) {
	docs, err := h.service.GetPendingDocuments(c.Request.Context())
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, docs, "Pending documents retrieved successfully")
}
