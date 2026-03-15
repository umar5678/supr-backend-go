package notifications

import (
	"fmt"
	"sync"
)

type EventType string

const (
	EventFraudPatternDetected EventType = "fraud.pattern.detected"
	EventFraudAlertCreated    EventType = "fraud.alert.created"

	EventRideRequested                EventType = "ride.requested"
	EventRideAccepted                 EventType = "ride.accepted"
	EventRideStarted                  EventType = "ride.started"
	EventRideCompleted                EventType = "ride.completed"
	EventRideCancelled                EventType = "ride.cancelled"
	EventRideRequestSent              EventType = "ride.request.sent"
	EventRideRequestCancelledBySystem EventType = "ride.request.cancelled_by_system"
	EventRideRequestAlreadyAccepted   EventType = "ride.request.already_accepted"
	EventRideRequestAccepted          EventType = "ride.request.accepted"
	EventRideRequestRejected          EventType = "ride.request.rejected"
	EventRideRequestExpired           EventType = "ride.request.expired"
	EventHighRiskRider                EventType = "ride.high_risk_rider"
	EventDriverArrived                EventType = "ride.driver.arrived"

	EventPaymentProcessed EventType = "payment.processed"
	EventPaymentFailed    EventType = "payment.failed"
	EventRefundIssued     EventType = "payment.refund.issued"

	EventVehicleRegistered EventType = "vehicle.registered"
	EventVehicleUpdated    EventType = "vehicle.updated"
	EventVehicleDeleted    EventType = "vehicle.deleted"

	EventUserRegistered EventType = "user.registered"
	EventUserVerified   EventType = "user.verified"
	EventUserSuspended  EventType = "user.suspended"

	EventSOSAlert     EventType = "sos.alert"
	EventSOSTriggered EventType = "sos.triggered"
	EventSOSResolved  EventType = "sos.resolved"

	EventPromoCodeApplied EventType = "promo.applied"
	EventPromoCodeExpired EventType = "promo.expired"

	EventMessageReceived             EventType = "message.received"
	EventMessageRead                 EventType = "message.read"
	EventMessageUnreadCountRetrieved EventType = "message.unread_count.retrieved"

	EventReferralCodeGenerated   EventType = "referral.code.generated"
	EventReferralCodeApplied     EventType = "referral.code.applied"
	EventKYCSubmitted            EventType = "kyc.submitted"
	EventLocationSaved           EventType = "location.saved"
	EventUserVerificationPending EventType = "user.verification.pending"

	EventPhoneSignup EventType = "auth.phone.signup"
	EventPhoneLogin  EventType = "auth.phone.login"
	EventEmailSignup EventType = "auth.email.signup"
	EventEmailLogin  EventType = "auth.email.login"

	EventRiderProfileCreated EventType = "rider.profile.created"
	EventRiderProfileUpdated EventType = "rider.profile.updated"
	EventRiderRatingUpdated  EventType = "rider.rating.updated"

	EventSurgeZoneCreated    EventType = "pricing.surge.zone.created"
	EventSurgePricingUpdated EventType = "pricing.surge.pricing.updated"
	EventFareEstimated       EventType = "pricing.fare.estimated"

	EventServiceProviderApproved EventType = "admin.provider.approved"
	EventUserStatusChanged       EventType = "admin.user.status.changed"

	EventRidePINGenerated   EventType = "ridepin.generated"
	EventRidePINRegenerated EventType = "ridepin.regenerated"
	EventRidePINVerified    EventType = "ridepin.verified"

	EventDriverLocationUpdated EventType = "tracking.location.updated"
	EventDriverOnline          EventType = "tracking.driver.online"
	EventDriverOffline         EventType = "tracking.driver.offline"

	EventUserConversationsRetrieved EventType = "admin.user.conversations.retrieved"
	EventUserConversationResolved   EventType = "admin.user.conversation.resolved"

	EventDocumentStatusUpdated EventType = "admin.document.status.updated"

	EventUserProfileUpdated EventType = "admin.user.profile.updated"

	EventRideDestinationChanged EventType = "ride.destination.changed"
	EventRideRouteUpdated       EventType = "ride.route.updated"
	EventRideUpdated            EventType = "ride.updated"
	EventRideAssigned           EventType = "ride.assigned"
	EventInvalidRidePINAttempt  EventType = "ride.pin.invalid_attempt"

	EventOrderPlaced       EventType = "food:order:placed"
	EventOrderAccepted     EventType = "food:order:accepted"
	EventOrderPickedUp     EventType = "food:order:picked_up"
	EventOrderDelivered    EventType = "food:order:delivered"
	EventOrderCancelled    EventType = "food:order:cancelled"
	EventOrderFailed       EventType = "food:order:failed"
	EventNewOrderAvailable EventType = "food:order:available"
	EventDeliveryAssigned  EventType = "food:delivery:assigned"
	EventDealCreated       EventType = "food:deal:created"
	EventDealExpired       EventType = "food:deal:expired"
	EventProductOOS        EventType = "food:product:out_of_stock"
)

type EventSchema struct {
	EventType   EventType
	Topic       string
	Module      string
	Description string
	Version     string
}

type EventRegistry struct {
	mu      sync.RWMutex
	schemas map[EventType]*EventSchema
}

func NewEventRegistry() *EventRegistry {
	registry := &EventRegistry{
		schemas: make(map[EventType]*EventSchema),
	}
	registry.registerDefaultSchemas()
	return registry
}

func (r *EventRegistry) registerDefaultSchemas() {
	defaultSchemas := []*EventSchema{

		{EventFraudPatternDetected, "fraud-events", "fraud", "Fraud pattern detected", "v1"},
		{EventFraudAlertCreated, "fraud-events", "fraud", "Fraud alert created", "v1"},

		{EventRideRequested, "ride-events", "rides", "New ride requested", "v1"},
		{EventRideAccepted, "ride-events", "rides", "Ride accepted by driver", "v1"},
		{EventRideStarted, "ride-events", "rides", "Ride started", "v1"},
		{EventRideCompleted, "ride-events", "rides", "Ride completed", "v1"},
		{EventRideCancelled, "ride-events", "rides", "Ride cancelled", "v1"},
		{EventRideRequestSent, "ride-events", "rides", "Ride request sent to driver", "v1"},
		{EventRideRequestCancelledBySystem, "ride-events", "rides", "Ride request cancelled by system", "v1"},
		{EventRideRequestAlreadyAccepted, "ride-events", "rides", "Ride request already accepted", "v1"},
		{EventRideRequestAccepted, "ride-events", "rides", "Ride request accepted by driver", "v1"},
		{EventRideRequestRejected, "ride-events", "rides", "Ride request rejected by driver", "v1"},
		{EventRideRequestExpired, "ride-events", "rides", "Ride request expired", "v1"},
		{EventHighRiskRider, "ride-events", "rides", "High risk rider detected", "v1"},
		{EventDriverArrived, "ride-events", "rides", "Driver arrived at pickup location", "v1"},
		{EventInvalidRidePINAttempt, "ride-events", "rides", "Invalid ride PIN attempt", "v1"},
		{EventRideAssigned, "ride-events", "rides", "Ride assigned to driver", "v1"},

		{EventPaymentProcessed, "payment-events", "payments", "Payment processed", "v1"},
		{EventPaymentFailed, "payment-events", "payments", "Payment failed", "v1"},
		{EventRefundIssued, "payment-events", "payments", "Refund issued", "v1"},

		{EventVehicleRegistered, "vehicle-events", "vehicles", "Vehicle registered", "v1"},
		{EventVehicleUpdated, "vehicle-events", "vehicles", "Vehicle updated", "v1"},
		{EventVehicleDeleted, "vehicle-events", "vehicles", "Vehicle deleted", "v1"},

		{EventUserRegistered, "user-events", "auth", "User registered", "v1"},
		{EventUserVerified, "user-events", "auth", "User verified", "v1"},
		{EventUserSuspended, "user-events", "auth", "User suspended", "v1"},
		{EventUserVerificationPending, "user-events", "profile", "User verification pending", "v1"},

		{EventSOSAlert, "sos-events", "sos", "SOS alert triggered", "v1"},
		{EventSOSTriggered, "sos-events", "sos", "SOS triggered", "v1"},
		{EventSOSResolved, "sos-events", "sos", "SOS alert resolved", "v1"},

		{EventPromoCodeApplied, "promotion-events", "promotions", "Promo code applied", "v1"},
		{EventPromoCodeExpired, "promotion-events", "promotions", "Promo code expired", "v1"},

		{EventMessageReceived, "message-events", "messages", "Message received", "v1"},
		{EventMessageRead, "message-events", "messages", "Message read", "v1"},

		{EventReferralCodeGenerated, "profile-events", "profile", "Referral code generated", "v1"},
		{EventReferralCodeApplied, "profile-events", "profile", "Referral code applied", "v1"},
		{EventKYCSubmitted, "profile-events", "profile", "KYC submitted", "v1"},
		{EventLocationSaved, "profile-events", "profile", "Location saved", "v1"},

		{EventPhoneSignup, "auth-events", "auth", "Phone signup", "v1"},
		{EventPhoneLogin, "auth-events", "auth", "Phone login", "v1"},
		{EventEmailSignup, "auth-events", "auth", "Email signup", "v1"},
		{EventEmailLogin, "auth-events", "auth", "Email login", "v1"},

		{EventRiderProfileCreated, "rider-events", "riders", "Rider profile created", "v1"},
		{EventRiderProfileUpdated, "rider-events", "riders", "Rider profile updated", "v1"},
		{EventRiderRatingUpdated, "rider-events", "riders", "Rider rating updated", "v1"},

		{EventSurgeZoneCreated, "pricing-events", "pricing", "Surge zone created", "v1"},
		{EventSurgePricingUpdated, "pricing-events", "pricing", "Surge pricing updated", "v1"},
		{EventFareEstimated, "pricing-events", "pricing", "Fare estimated", "v1"},

		{EventServiceProviderApproved, "admin-events", "admin", "Service provider approved", "v1"},
		{EventUserStatusChanged, "admin-events", "admin", "User status changed", "v1"},

		{EventRidePINGenerated, "ridepin-events", "ridepin", "RidePin generated", "v1"},
		{EventRidePINRegenerated, "ridepin-events", "ridepin", "RidePin regenerated", "v1"},
		{EventRidePINVerified, "ridepin-events", "ridepin", "RidePin verified", "v1"},

		{EventDriverLocationUpdated, "tracking-events", "tracking", "Driver location updated", "v1"},
		{EventDriverOnline, "tracking-events", "tracking", "Driver online", "v1"},
		{EventDriverOffline, "tracking-events", "tracking", "Driver offline", "v1"},

		{EventMessageReceived, "message-events", "messages", "Message received", "v1"},
		{EventMessageRead, "message-events", "messages", "Message read", "v1"},
		{EventMessageUnreadCountRetrieved, "message-events", "messages", "Unread message count retrieved", "v1"},
		{EventUserConversationResolved, "message-events", "messages", "User conversation resolved", "v1"},
		{EventUserConversationsRetrieved, "message-events", "messages", "User conversations retrieved", "v1"},

		{EventDocumentStatusUpdated, "document-events", "documents", "Document status updated", "v1"},

		{EventUserProfileUpdated, "profile-events", "profile", "User profile updated", "v1"},

		{EventRideDestinationChanged, "ride-events", "ride", "Ride destination changed", "v1"},

		{EventOrderPlaced, "food-order-events", "food", "Order placed", "v1"},
		{EventOrderAccepted, "food-order-events", "food", "Order accepted by delivery person", "v1"},
		{EventOrderPickedUp, "food-order-events", "food", "Order picked up from restaurant", "v1"},
		{EventOrderDelivered, "food-order-events", "food", "Order delivered to customer", "v1"},
		{EventOrderCancelled, "food-order-events", "food", "Order cancelled", "v1"},
		{EventOrderFailed, "food-order-events", "food", "Order delivery failed", "v1"},
		{EventNewOrderAvailable, "food-order-events", "food", "New order available for delivery persons", "v1"},
		{EventDeliveryAssigned, "food-order-events", "food", "Delivery assigned to a delivery person", "v1"},
		{EventDealCreated, "food-deals-events", "food", "New deal created", "v1"},
		{EventDealExpired, "food-deals-events", "food", "Deal expired", "v1"},
		{EventProductOOS, "food-product-events", "food", "Product out of stock", "v1"},
	}

	for _, schema := range defaultSchemas {
		r.schemas[schema.EventType] = schema
	}
}

func (r *EventRegistry) Register(schema *EventSchema) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.schemas[schema.EventType]; exists {
		return fmt.Errorf("event type %s already registered", schema.EventType)
	}

	r.schemas[schema.EventType] = schema
	return nil
}

func (r *EventRegistry) Get(eventType EventType) (*EventSchema, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	schema, exists := r.schemas[eventType]
	if !exists {
		return nil, fmt.Errorf("event type %s not registered", eventType)
	}

	return schema, nil
}

func (r *EventRegistry) GetTopic(eventType EventType) (string, error) {
	schema, err := r.Get(eventType)
	if err != nil {
		return "", err
	}
	return schema.Topic, nil
}

func (r *EventRegistry) ListAll() []*EventSchema {
	r.mu.RLock()
	defer r.mu.RUnlock()

	schemas := make([]*EventSchema, 0, len(r.schemas))
	for _, schema := range r.schemas {
		schemas = append(schemas, schema)
	}
	return schemas
}
