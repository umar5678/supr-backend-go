package shared

const (
	OrderStatusPending           = "pending"
	OrderStatusSearchingProvider = "searching_provider"

	OrderStatusAssigned = "assigned"
	OrderStatusAccepted = "accepted"

	OrderStatusInProgress = "in_progress"

	OrderStatusCompleted = "completed"
	OrderStatusCancelled = "cancelled"
	OrderStatusExpired   = "expired"
)

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

func IsValidOrderStatus(status string) bool {
	for _, s := range AllOrderStatuses() {
		if s == status {
			return true
		}
	}
	return false
}

func ActiveOrderStatuses() []string {
	return []string{
		OrderStatusPending,
		OrderStatusSearchingProvider,
		OrderStatusAssigned,
		OrderStatusAccepted,
		OrderStatusInProgress,
	}
}

func CancelableOrderStatuses() []string {
	return []string{
		OrderStatusPending,
		OrderStatusSearchingProvider,
		OrderStatusAssigned,
		OrderStatusAccepted,
	}
}

const (
	PaymentStatusPending   = "pending"
	PaymentStatusCompleted = "completed"
	PaymentStatusFailed    = "failed"
	PaymentStatusRefunded  = "refunded"
)

const (
	PaymentMethodWallet = "wallet"
	PaymentMethodCash   = "cash"
	PaymentMethodCard   = "card"
)

const (
	CancelledByCustomer = "customer"
	CancelledByProvider = "provider"
	CancelledByAdmin    = "admin"
	CancelledBySystem   = "system"
)

const (
	RoleCustomer = "customer"
	RoleProvider = "provider"
	RoleAdmin    = "admin"
	RoleSystem   = "system"
)

const (
	ExpertiseBeginner     = "beginner"
	ExpertiseIntermediate = "intermediate"
	ExpertiseExpert       = "expert"
)

const (
	PreferredTimeMorning   = "morning"  
	PreferredTimeAfternoon = "afternoon"
	PreferredTimeEvening   = "evening"  
)

const (
	PlatformCommissionRate = 0.10 
)

const (
	CancellationFeeBeforeAcceptance = 0.10 
	CancellationFeeAfterAcceptance  = 0.50 
	CancellationFeeAfterStart       = 1.00 
)

const (
	OrderExpirationMinutes = 10 
)

const (
	DefaultPage  = 1
	DefaultLimit = 20
	MaxLimit     = 100
)
