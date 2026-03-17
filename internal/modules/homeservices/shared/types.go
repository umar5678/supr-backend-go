package shared

import "time"

type StatusTransition struct {
	From []string
	To   string
}

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

type PaginationParams struct {
	Page  int `form:"page" binding:"omitempty,min=1"`
	Limit int `form:"limit" binding:"omitempty,min=1,max=100"`
}

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

func (p *PaginationParams) GetOffset() int {
	return (p.Page - 1) * p.Limit
}

type DateRange struct {
	StartDate *time.Time `form:"startDate" time_format:"2006-01-02"`
	EndDate   *time.Time `form:"endDate" time_format:"2006-01-02"`
}

type ServiceCategory struct {
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	SortOrder   int    `json:"sortOrder"`
	IsActive    bool   `json:"isActive"`
}

type SortParams struct {
	SortBy   string `form:"sortBy"`
	SortDesc bool   `form:"sortDesc"`
}
