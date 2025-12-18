package shared

// Order Status Constants
const (
	// Initial statuses
	OrderStatusPending           = "pending"
	OrderStatusSearchingProvider = "searching_provider"

	// Assignment statuses
	OrderStatusAssigned = "assigned"
	OrderStatusAccepted = "accepted"

	// In-progress statuses
	OrderStatusInProgress = "in_progress"

	// Final statuses
	OrderStatusCompleted = "completed"
	OrderStatusCancelled = "cancelled"
	OrderStatusExpired   = "expired"
)

// AllOrderStatuses returns all valid order statuses
func AllOrderStatuses() []string {
	return []string{
		OrderStatusPending,
		OrderStatusSearchingProvider,
		OrderStatusAssigned,
		OrderStatusAccepted,
		OrderStatusInProgress,
		OrderStatusCompleted,
		OrderStatusCancelled,
		OrderStatusExpired,
	}
}

// IsValidOrderStatus checks if status is valid
func IsValidOrderStatus(status string) bool {
	for _, s := range AllOrderStatuses() {
		if s == status {
			return true
		}
	}
	return false
}

// ActiveOrderStatuses returns statuses that are considered "active"
func ActiveOrderStatuses() []string {
	return []string{
		OrderStatusPending,
		OrderStatusSearchingProvider,
		OrderStatusAssigned,
		OrderStatusAccepted,
		OrderStatusInProgress,
	}
}

// CancelableOrderStatuses returns statuses that allow cancellation
func CancelableOrderStatuses() []string {
	return []string{
		OrderStatusPending,
		OrderStatusSearchingProvider,
		OrderStatusAssigned,
		OrderStatusAccepted,
	}
}

// Payment Status Constants
const (
	PaymentStatusPending   = "pending"
	PaymentStatusCompleted = "completed"
	PaymentStatusFailed    = "failed"
	PaymentStatusRefunded  = "refunded"
)

// Payment Method Constants
const (
	PaymentMethodWallet = "wallet"
	PaymentMethodCash   = "cash"
	PaymentMethodCard   = "card"
)

// Cancellation Actor Constants
const (
	CancelledByCustomer = "customer"
	CancelledByProvider = "provider"
	CancelledByAdmin    = "admin"
	CancelledBySystem   = "system"
)

// Actor Role Constants
const (
	RoleCustomer = "customer"
	RoleProvider = "provider"
	RoleAdmin    = "admin"
	RoleSystem   = "system"
)

// Expertise Level Constants
const (
	ExpertiseBeginner     = "beginner"
	ExpertiseIntermediate = "intermediate"
	ExpertiseExpert       = "expert"
)

// Preferred Time Constants
const (
	PreferredTimeMorning   = "morning"   // 6:00 - 12:00
	PreferredTimeAfternoon = "afternoon" // 12:00 - 17:00
	PreferredTimeEvening   = "evening"   // 17:00 - 21:00
)

// Commission Constants
const (
	PlatformCommissionRate = 0.10 // 10%
)

// Cancellation Fee Constants
const (
	CancellationFeeBeforeAcceptance = 0.10 // 10% before provider accepts
	CancellationFeeAfterAcceptance  = 0.50 // 50% after provider accepts
	CancellationFeeAfterStart       = 1.00 // 100% after service starts (no refund)
)

// Order Expiration Constants
const (
	OrderExpirationMinutes = 30 // Orders expire after 30 minutes if no provider found
)

// Pagination Defaults
const (
	DefaultPage  = 1
	DefaultLimit = 20
	MaxLimit     = 100
)
