package response

import "github.com/gin-gonic/gin"

type PaginationMeta struct {
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalPages int   `json:"totalPages"`
	HasNext    bool  `json:"hasNext"`
	HasPrev    bool  `json:"hasPrev"`
}

func NewPaginationMeta(total int64, page, limit int) PaginationMeta {
	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	return PaginationMeta{
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}
}

func Paginated(c *gin.Context, data interface{}, pagination PaginationMeta, message string) {
	resp := Response{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    extractMeta(c),
	}

	c.Set("pagination", pagination)

	c.JSON(200, gin.H{
		"success":    resp.Success,
		"message":    resp.Message,
		"data":       resp.Data,
		"meta":       resp.Meta,
		"pagination": pagination,
	})
}
