package dto

import (
	"fmt"
)

// CreateLaundryOrderRequest - Create new laundry order with products
type CreateLaundryOrderRequest struct {
	ServiceSlug  string             `json:"serviceSlug" binding:"required"`
	Items        []OrderItemRequest `json:"items" binding:"required,dive"`
	PickupDate   string             `json:"pickupDate" binding:"required"`
	PickupTime   string             `json:"pickupTime" binding:"required"`
	IsExpress    bool               `json:"isExpress"`
	SpecialNotes string             `json:"specialNotes"`
	Address      string             `json:"address" binding:"required"`
	Lat          float64            `json:"lat" binding:"required"`
	Lng          float64            `json:"lng" binding:"required"`
	Tip          *float64           `json:"tip,omitempty"` // Optional tip for delivery person
}

// OrderServiceRequest represents a service with its items in a multi-service order
type OrderServiceRequest struct {
	ServiceSlug string             `json:"serviceSlug" binding:"required"`
	Items       []OrderItemRequest `json:"items" binding:"required,dive"`
}

type OrderItemRequest struct {
	ProductSlug string   `json:"productSlug" binding:"required"`
	Quantity    int      `json:"quantity" binding:"required,min=1"`
	Weight      *float64 `json:"weight,omitempty"` // Optional, can be calculated from product
	Notes       string   `json:"notes"`
}

// Validate validates the CreateLaundryOrderRequest
func (r *CreateLaundryOrderRequest) Validate() error {
	if r.ServiceSlug == "" {
		return fmt.Errorf("serviceSlug is required")
	}
	if r.PickupDate == "" {
		return fmt.Errorf("pickupDate is required")
	}
	if r.PickupTime == "" {
		return fmt.Errorf("pickupTime is required")
	}
	if r.Address == "" {
		return fmt.Errorf("address is required")
	}
	if r.Lat == 0 || r.Lng == 0 {
		return fmt.Errorf("valid coordinates are required")
	}

	// Validate each item
	for i, item := range r.Items {
		if item.ProductSlug == "" {
			return fmt.Errorf("productSlug is required for item %d", i+1)
		}
		if item.Quantity <= 0 {
			return fmt.Errorf("quantity must be greater than 0 for item %d", i+1)
		}
	}

	return nil
}

// CompletePickupRequest represents a pickup completion request
type CompletePickupRequest struct {
	BagCount int     `json:"bagCount" binding:"required,gt=0"`
	Notes    string  `json:"notes"`
	PhotoURL *string `json:"photoUrl"`
}

// Validate validates the CompletePickupRequest
func (r *CompletePickupRequest) Validate() error {
	if r.BagCount <= 0 {
		return fmt.Errorf("bagCount must be greater than 0")
	}
	return nil
}

// AddLaundryItemsRequest represents a request to add items to an order
type AddLaundryItemsRequest struct {
	Items []AddItemDTO `json:"items" binding:"required,min=1"`
}

// AddItemDTO represents a single item being added
type AddItemDTO struct {
	ProductSlug string   `json:"productSlug" binding:"required"` // Changed from ItemType
	ItemType    string   `json:"itemType" binding:"required"`
	Quantity    int      `json:"quantity" binding:"required,gt=0"`
	ServiceSlug string   `json:"serviceSlug" binding:"required"`
	Weight      *float64 `json:"weight"`
	Price       float64  `json:"price" binding:"required,gt=0"`
}

// Validate validates the AddLaundryItemsRequest
func (r *AddLaundryItemsRequest) Validate() error {
	if len(r.Items) == 0 {
		return fmt.Errorf("at least one item is required")
	}
	for i, item := range r.Items {
		if item.ProductSlug == "" {
			return fmt.Errorf("productSlug is required for item %d", i+1)
		}
		if item.ItemType == "" {
			return fmt.Errorf("itemType is required for item %d", i+1)
		}
		if item.Quantity <= 0 {
			return fmt.Errorf("quantity must be greater than 0 for item %d", i+1)
		}
		if item.ServiceSlug == "" {
			return fmt.Errorf("serviceSlug is required for item %d", i+1)
		}
		if item.Price <= 0 {
			return fmt.Errorf("price must be greater than 0 for item %d", i+1)
		}
	}
	return nil
}

// UpdateItemStatusRequest represents a request to update an item's status
type UpdateItemStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=pending received washing drying pressing packed delivered"`
}

// Validate validates the UpdateItemStatusRequest
func (r *UpdateItemStatusRequest) Validate() error {
	if r.Status == "" {
		return fmt.Errorf("status is required")
	}
	validStatuses := map[string]bool{
		"pending": true, "received": true, "washing": true, "drying": true,
		"pressing": true, "packed": true, "delivered": true,
	}
	if !validStatuses[r.Status] {
		return fmt.Errorf("invalid status: %s", r.Status)
	}
	return nil
}

// Validate validates the CompleteDeliveryRequest
func (r *CompleteDeliveryRequest) Validate() error {
	if r.RecipientName == "" {
		return fmt.Errorf("recipientName is required")
	}
	return nil
}

// Validate validates the ReportIssueRequest
func (r *ReportIssueRequest) Validate() error {
	if r.IssueType == "" {
		return fmt.Errorf("issueType is required")
	}
	if r.Description == "" {
		return fmt.Errorf("description is required")
	}
	if r.Priority == "" {
		r.Priority = "medium"
	}

	validIssueTypes := map[string]bool{
		"missing_item": true, "damage": true, "poor_cleaning": true,
		"late_delivery": true, "wrong_item": true, "stain_not_removed": true,
		"color_bleeding": true, "shrinkage": true, "other": true,
	}
	if !validIssueTypes[r.IssueType] {
		return fmt.Errorf("invalid issueType: %s", r.IssueType)
	}

	return nil
}

// Validate validates the ResolveIssueRequest
func (r *ResolveIssueRequest) Validate() error {
	if r.Resolution == "" {
		return fmt.Errorf("resolution is required")
	}

	if r.CompensationType != "" {
		validTypes := map[string]bool{
			"refund": true, "discount": true, "re_clean": true,
			"replacement": true, "voucher": true,
		}
		if !validTypes[r.CompensationType] {
			return fmt.Errorf("invalid compensationType: %s", r.CompensationType)
		}
	}

	if r.RefundAmount != nil && *r.RefundAmount < 0 {
		return fmt.Errorf("refundAmount cannot be negative")
	}

	return nil
}

// CompleteDeliveryRequest represents a delivery completion request
type CompleteDeliveryRequest struct {
	RecipientName      string  `json:"recipientName" binding:"required"`
	RecipientSignature *string `json:"recipientSignature"`
	Notes              string  `json:"notes"`
	PhotoURL           *string `json:"photoUrl"`
}

// ReportIssueRequest represents a request to report an issue
type ReportIssueRequest struct {
	IssueType   string `json:"issueType" binding:"required,oneof=missing_item damage poor_cleaning late_delivery wrong_item stain_not_removed color_bleeding shrinkage other"`
	Description string `json:"description" binding:"required"`
	Priority    string `json:"priority" binding:"omitempty,oneof=low medium high urgent"`
}

// ResolveIssueRequest represents a request to resolve an issue
type ResolveIssueRequest struct {
	Resolution       string   `json:"resolution" binding:"required"`
	RefundAmount     *float64 `json:"refundAmount"`
	CompensationType string   `json:"compensationType" binding:"omitempty,oneof=refund discount re_clean replacement voucher"`
}

// OrderService represents a service being ordered (legacy support)
type OrderService struct {
	ServiceSlug string  `json:"serviceSlug" binding:"required"`
	Quantity    int     `json:"quantity" binding:"required,gt=0"`
	Price       float64 `json:"price" binding:"required,gt=0"`
}
