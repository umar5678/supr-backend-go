package dto

import (
	"errors"
	"fmt"
)

// CreateOrderRequest represents a customer's service booking request
type CreateOrderRequest struct {
	Items       []CreateOrderItemRequest `json:"items" binding:"required,min=1,dive"`
	Address     string                   `json:"address" binding:"required,min=5,max=500"`
	Latitude    float64                  `json:"latitude" binding:"required,latitude"`
	Longitude   float64                  `json:"longitude" binding:"required,longitude"`
	ServiceDate string                   `json:"serviceDate" binding:"required"` // RFC3339 format
	Frequency   string                   `json:"frequency" binding:"omitempty,oneof=once daily weekly monthly"`
	Notes       *string                  `json:"notes" binding:"omitempty,max=500"`
	CouponCode  *string                  `json:"couponCode" binding:"omitempty,max=50"`
}

type CreateOrderItemRequest struct {
	ServiceID       uint                    `json:"serviceId" binding:"required,min=1"`
	SelectedOptions []SelectedOptionRequest `json:"selectedOptions" binding:"omitempty,dive"`
}

type SelectedOptionRequest struct {
	OptionID uint    `json:"optionId" binding:"required,min=1"`
	ChoiceID *uint   `json:"choiceId" binding:"omitempty,min=1"` // For select_single/select_multiple
	Value    *string `json:"value" binding:"omitempty"`          // For text/quantity types
}

func (r *CreateOrderRequest) Validate() error {
	if len(r.Items) == 0 {
		return errors.New("at least one service item is required")
	}
	if r.Frequency == "" {
		r.Frequency = "once"
	}
	return nil
}

func (r *CreateOrderRequest) SetDefaults() {
	if r.Frequency == "" {
		r.Frequency = "once"
	}
}

// UpdateOrderStatusRequest for provider actions
type UpdateOrderStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=accepted rejected in_progress completed"`
}

// ListServicesQuery for browsing services
type ListServicesQuery struct {
	Page       int      `form:"page" binding:"omitempty,min=1"`
	Limit      int      `form:"limit" binding:"omitempty,min=1,max=100"`
	CategoryID *uint    `form:"categoryId" binding:"omitempty,min=1"`
	Search     string   `form:"search" binding:"omitempty,max=100"`
	MinPrice   *float64 `form:"minPrice" binding:"omitempty,gte=0"`
	MaxPrice   *float64 `form:"maxPrice" binding:"omitempty,gte=0"`
	IsActive   *bool    `form:"isActive"`
}

func (q *ListServicesQuery) SetDefaults() {
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.Limit <= 0 {
		q.Limit = 20
	}
}

func (q *ListServicesQuery) GetOffset() int {
	return (q.Page - 1) * q.Limit
}

// ListOrdersQuery for order history
type ListOrdersQuery struct {
	Page   int     `form:"page" binding:"omitempty,min=1"`
	Limit  int     `form:"limit" binding:"omitempty,min=1,max=100"`
	Status *string `form:"status"`
}

func (q *ListOrdersQuery) SetDefaults() {
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.Limit <= 0 {
		q.Limit = 20
	}
}

func (q *ListOrdersQuery) GetOffset() int {
	return (q.Page - 1) * q.Limit
}

// Admin requests
type CreateServiceRequest struct {
	CategoryID          uint    `json:"categoryId" binding:"required,min=1"`
	Name                string  `json:"name" binding:"required,min=3,max=150"`
	Description         string  `json:"description" binding:"required,max=2000"`
	ImageURL            string  `json:"imageUrl" binding:"omitempty,url"`
	BasePrice           float64 `json:"basePrice" binding:"required,gt=0"`
	PricingModel        string  `json:"pricingModel" binding:"required,oneof=fixed hourly per_unit"`
	BaseDurationMinutes int     `json:"baseDurationMinutes" binding:"required,min=1"`
}

type UpdateServiceRequest struct {
	CategoryID          *uint    `json:"categoryId" binding:"omitempty,min=1"`
	Name                *string  `json:"name" binding:"omitempty,min=3,max=150"`
	Description         *string  `json:"description" binding:"omitempty,max=2000"`
	ImageURL            *string  `json:"imageUrl" binding:"omitempty,url"`
	BasePrice           *float64 `json:"basePrice" binding:"omitempty,gt=0"`
	PricingModel        *string  `json:"pricingModel" binding:"omitempty,oneof=fixed hourly per_unit"`
	BaseDurationMinutes *int     `json:"baseDurationMinutes" binding:"omitempty,min=1"`
	IsActive            *bool    `json:"isActive"`
}

func (r *UpdateServiceRequest) Validate() error {
	hasUpdate := r.CategoryID != nil || r.Name != nil || r.Description != nil ||
		r.ImageURL != nil || r.BasePrice != nil || r.PricingModel != nil ||
		r.BaseDurationMinutes != nil || r.IsActive != nil

	if !hasUpdate {
		return fmt.Errorf("at least one field must be provided for update")
	}
	return nil
}

type CreateServiceOptionRequest struct {
	ServiceID  uint   `json:"serviceId" binding:"required,min=1"`
	Name       string `json:"name" binding:"required,min=2,max=100"`
	Type       string `json:"type" binding:"required,oneof=select_single select_multiple quantity text"`
	IsRequired bool   `json:"isRequired"`
}

type CreateOptionChoiceRequest struct {
	OptionID                uint    `json:"optionId" binding:"required,min=1"`
	Label                   string  `json:"label" binding:"required,min=1,max=100"`
	PriceModifier           float64 `json:"priceModifier"`
	DurationModifierMinutes int     `json:"durationModifierMinutes"`
}
