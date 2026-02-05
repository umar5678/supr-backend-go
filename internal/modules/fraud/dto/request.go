package dto

type ListFraudPatternsRequest struct {
	PatternType  string `form:"patternType" binding:"omitempty"`
	Status       string `form:"status" binding:"omitempty,oneof=flagged investigating confirmed dismissed"`
	MinRiskScore int    `form:"minRiskScore" binding:"omitempty,min=0,max=100"`
	Page         int    `form:"page" binding:"omitempty,min=1"`
	Limit        int    `form:"limit" binding:"omitempty,min=1,max=100"`
}

func (r *ListFraudPatternsRequest) SetDefaults() {
	if r.Page == 0 {
		r.Page = 1
	}
	if r.Limit == 0 {
		r.Limit = 20
	}
	if r.MinRiskScore == 0 {
		r.MinRiskScore = 50 // Default: show medium+ risk
	}
}

type ReviewFraudPatternRequest struct {
	Status      string `json:"status" binding:"required,oneof=investigating confirmed dismissed"`
	ReviewNotes string `json:"reviewNotes" binding:"omitempty,max=1000"`
}
