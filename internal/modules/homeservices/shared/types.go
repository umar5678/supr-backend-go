package shared

import "time"

// StatusTransition defines valid status transitions
type StatusTransition struct {
	From []string
	To   string
}

// ValidTransitions defines all valid status transitions
var ValidTransitions = map[string]StatusTransition{
	OrderStatusSearchingProvider: {
		From: []string{OrderStatusPending},
		To:   OrderStatusSearchingProvider,
	},
	OrderStatusAssigned: {
		From: []string{OrderStatusSearchingProvider},
		To:   OrderStatusAssigned,
	},
	OrderStatusAccepted: {
		From: []string{OrderStatusAssigned},
		To:   OrderStatusAccepted,
	},
	OrderStatusInProgress: {
		From: []string{OrderStatusAccepted},
		To:   OrderStatusInProgress,
	},
	OrderStatusCompleted: {
		From: []string{OrderStatusInProgress},
		To:   OrderStatusCompleted,
	},
	OrderStatusCancelled: {
		From: []string{
			OrderStatusPending,
			OrderStatusSearchingProvider,
			OrderStatusAssigned,
			OrderStatusAccepted,
		},
		To: OrderStatusCancelled,
	},
	OrderStatusExpired: {
		From: []string{
			OrderStatusPending,
			OrderStatusSearchingProvider,
		},
		To: OrderStatusExpired,
	},
}

// CanTransition checks if a status transition is valid
func CanTransition(fromStatus, toStatus string) bool {
	transition, exists := ValidTransitions[toStatus]
	if !exists {
		return false
	}
	for _, allowedFrom := range transition.From {
		if allowedFrom == fromStatus {
			return true
		}
	}
	return false
}

// PaginationParams holds pagination parameters
type PaginationParams struct {
	Page  int `form:"page" binding:"omitempty,min=1"`
	Limit int `form:"limit" binding:"omitempty,min=1,max=100"`
}

// SetDefaults sets default pagination values
func (p *PaginationParams) SetDefaults() {
	if p.Page <= 0 {
		p.Page = DefaultPage
	}
	if p.Limit <= 0 {
		p.Limit = DefaultLimit
	}
	if p.Limit > MaxLimit {
		p.Limit = MaxLimit
	}
}

// GetOffset calculates the offset for database queries
func (p *PaginationParams) GetOffset() int {
	return (p.Page - 1) * p.Limit
}

// DateRange represents a date range filter
type DateRange struct {
	StartDate *time.Time `form:"startDate" time_format:"2006-01-02"`
	EndDate   *time.Time `form:"endDate" time_format:"2006-01-02"`
}

// ServiceCategory represents a service category (from external config/DB)
type ServiceCategory struct {
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	SortOrder   int    `json:"sortOrder"`
	IsActive    bool   `json:"isActive"`
}

// SortParams holds sorting parameters
type SortParams struct {
	SortBy   string `form:"sortBy"`
	SortDesc bool   `form:"sortDesc"`
}
