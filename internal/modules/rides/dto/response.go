package dto

import (
	"time"

	"github.com/umar5678/go-backend/internal/models"
	authdto "github.com/umar5678/go-backend/internal/modules/auth/dto"
	vehicledto "github.com/umar5678/go-backend/internal/modules/vehicles/dto"
)

type RideResponse struct {
	ID            string                          `json:"id"`
	RiderID       string                          `json:"riderId"`
	Rider         *authdto.UserResponse           `json:"rider,omitempty"`
	DriverID      *string                         `json:"driverId"`
	Driver        *authdto.DriverResponse         `json:"driver,omitempty"`
	VehicleTypeID string                          `json:"vehicleTypeId"`
	VehicleType   *vehicledto.VehicleTypeResponse `json:"vehicleType,omitempty"`
	Status        string                          `json:"status"`

	PickupLat      float64 `json:"pickupLat"`
	PickupLon      float64 `json:"pickupLon"`
	PickupAddress  string  `json:"pickupAddress"`
	DropoffLat     float64 `json:"dropoffLat"`
	DropoffLon     float64 `json:"dropoffLon"`
	DropoffAddress string  `json:"dropoffAddress"`

	EstimatedDistance float64 `json:"estimatedDistance"`
	EstimatedDuration int     `json:"estimatedDuration"`
	EstimatedFare     float64 `json:"estimatedFare"`

	ActualDistance *float64 `json:"actualDistance,omitempty"`
	ActualDuration *int     `json:"actualDuration,omitempty"`
	ActualFare     *float64 `json:"actualFare,omitempty"`
	PromoDiscount  *float64 `json:"promoDiscount,omitempty"`
	WaitTimeCharge *float64 `json:"waitTimeCharge,omitempty"`

	// Driver and Rider Fares - Both see the same amount (with promo discount applied if used)
	DriverFare *float64 `json:"driverFare,omitempty"` // What driver earns (with discount applied)
	RiderFare  *float64 `json:"riderFare,omitempty"`  // What rider pays (with discount applied)

	SurgeMultiplier    float64 `json:"surgeMultiplier"`
	RiderNotes         string  `json:"riderNotes,omitempty"`
	CancellationReason string  `json:"cancellationReason,omitempty"`
	CancelledBy        *string `json:"cancelledBy,omitempty"`

	IsScheduled bool       `json:"isScheduled"`
	ScheduledAt *time.Time `json:"scheduledAt,omitempty"`

	RequestedAt time.Time  `json:"requestedAt"`
	AcceptedAt  *time.Time `json:"acceptedAt,omitempty"`
	ArrivedAt   *time.Time `json:"arrivedAt,omitempty"`
	StartedAt   *time.Time `json:"startedAt,omitempty"`
	CompletedAt *time.Time `json:"completedAt,omitempty"`
	CancelledAt *time.Time `json:"cancelledAt,omitempty"`

	HasActiveSOS bool    `json:"hasActiveSos"`
	SOSAlertID   *string `json:"sosAlertId,omitempty"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	DriverLocation *LocationDTO `json:"driverLocation,omitempty"`
}

type RideListResponse struct {
	ID             string    `json:"id"`
	Status         string    `json:"status"`
	PickupAddress  string    `json:"pickupAddress"`
	DropoffAddress string    `json:"dropoffAddress"`
	EstimatedFare  float64   `json:"estimatedFare"`
	RequestedAt    time.Time `json:"requestedAt"`
}

func ToRideResponse(ride *models.Ride) *RideResponse {
	// ✅ CRITICAL: Handle nil ride to prevent panic
	if ride == nil {
		return nil
	}

	resp := &RideResponse{
		ID:                 ride.ID,
		RiderID:            ride.RiderID,
		DriverID:           ride.DriverID,
		VehicleTypeID:      ride.VehicleTypeID,
		Status:             ride.Status,
		PickupLat:          ride.PickupLat,
		PickupLon:          ride.PickupLon,
		PickupAddress:      ride.PickupAddress,
		DropoffLat:         ride.DropoffLat,
		DropoffLon:         ride.DropoffLon,
		DropoffAddress:     ride.DropoffAddress,
		EstimatedDistance:  ride.EstimatedDistance,
		EstimatedDuration:  ride.EstimatedDuration,
		EstimatedFare:      ride.EstimatedFare,
		ActualDistance:     ride.ActualDistance,
		ActualDuration:     ride.ActualDuration,
		ActualFare:         ride.ActualFare,
		SurgeMultiplier:    ride.SurgeMultiplier,
		RiderNotes:         ride.RiderNotes,
		CancellationReason: ride.CancellationReason,
		CancelledBy:        ride.CancelledBy,
		IsScheduled:        ride.IsScheduled,
		ScheduledAt:        ride.ScheduledAt,
		PromoDiscount:      ride.PromoDiscount,
		WaitTimeCharge:     ride.WaitTimeCharge,
		DriverFare:         ride.DriverFare,
		RiderFare:          ride.RiderFare,
		RequestedAt:        ride.RequestedAt,
		AcceptedAt:         ride.AcceptedAt,
		ArrivedAt:          ride.ArrivedAt,
		StartedAt:          ride.StartedAt,
		CompletedAt:        ride.CompletedAt,
		CancelledAt:        ride.CancelledAt,
		CreatedAt:          ride.CreatedAt,
		UpdatedAt:          ride.UpdatedAt,
	}

	if ride.Rider.ID != "" {
		resp.Rider = authdto.ToUserResponse(&ride.Rider)
	}
	if ride.Driver != nil && ride.Driver.ID != "" {
		resp.Driver = authdto.ToDriverResponse(ride.Driver, ride.DriverProfile)
	}
	if ride.VehicleType.ID != "" {
		resp.VehicleType = vehicledto.ToVehicleTypeResponse(&ride.VehicleType)
	}

	return resp
}

func ToRideListResponse(ride *models.Ride) *RideListResponse {
	return &RideListResponse{
		ID:             ride.ID,
		Status:         ride.Status,
		PickupAddress:  ride.PickupAddress,
		DropoffAddress: ride.DropoffAddress,
		EstimatedFare:  ride.EstimatedFare,
		RequestedAt:    ride.RequestedAt,
	}
}

type LocationDTO struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// AvailableCarResponse represents a single available car
type AvailableCarResponse struct {
	ID                 string    `json:"id"`           // Car/Vehicle ID
	DriverID           string    `json:"driverId"`     // Driver ID
	DriverName         string    `json:"driverName"`   // Driver name
	DriverRating       float64   `json:"driverRating"` // Driver rating (0-5)
	DriverImage        *string   `json:"driverImage,omitempty"`
	VehicleTypeID      string    `json:"vehicleTypeId"`
	VehicleType        string    `json:"vehicleType"` // e.g., "economy", "comfort", "premium", "xl"
	VehicleDisplayName string    `json:"vehicleDisplayName"`
	Make               string    `json:"make"`            // e.g., "Toyota"
	Model              string    `json:"model"`           // e.g., "Corolla"
	Color              string    `json:"color"`           // e.g., "Silver"
	LicensePlate       string    `json:"licensePlate"`    // e.g., "DHA-1234"
	Capacity           int       `json:"capacity"`        // Passenger capacity
	CurrentLatitude    float64   `json:"currentLatitude"` // Current location
	CurrentLongitude   float64   `json:"currentLongitude"`
	Heading            int       `json:"heading"`          // Direction (0-360)
	DistanceKm         float64   `json:"distanceKm"`       // Distance from rider
	ETASeconds         int       `json:"etaSeconds"`       // ETA in seconds
	ETAMinutes         int       `json:"etaMinutes"`       // ETA in minutes
	EstimatedFare      float64   `json:"estimatedFare"`    // Estimated ride fare
	SurgeMultiplier    float64   `json:"surgeMultiplier"`  // Current surge pricing multiplier
	AcceptanceRate     float64   `json:"acceptanceRate"`   // Driver acceptance rate
	CancellationRate   float64   `json:"cancellationRate"` // Driver cancellation rate
	TotalTrips         int       `json:"totalTrips"`       // Driver total trips
	Status             string    `json:"status"`           // "online", "busy", etc.
	IsVerified         bool      `json:"isVerified"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

// AvailableCarsListResponse represents a list of available cars
type AvailableCarsListResponse struct {
	TotalCount int                     `json:"totalCount"` // Total available cars found
	CarsCount  int                     `json:"carsCount"`  // Number of cars in response
	RiderLat   float64                 `json:"riderLat"`
	RiderLon   float64                 `json:"riderLon"`
	RadiusKm   float64                 `json:"radiusKm"`
	Cars       []*AvailableCarResponse `json:"cars"`
	Timestamp  time.Time               `json:"timestamp"`
}

// WebSocketAvailableCarsMessage represents a WebSocket message for streaming available cars
type WebSocketAvailableCarsMessage struct {
	Type      string                     `json:"type"` // "cars_update", "error", "end"
	Data      *AvailableCarsListResponse `json:"data,omitempty"`
	Error     string                     `json:"error,omitempty"`
	Timestamp time.Time                  `json:"timestamp"`
}

// VehicleWithDetailsResponse represents a vehicle with all pricing and availability details
type VehicleWithDetailsResponse struct {
	// Vehicle Information
	ID               string  `json:"id"`           // Vehicle ID
	DriverID         string  `json:"driverId"`     // Driver Profile ID
	DriverName       string  `json:"driverName"`   // Driver name
	DriverRating     float64 `json:"driverRating"` // Driver rating (0-5)
	DriverImage      *string `json:"driverImage,omitempty"`
	AcceptanceRate   float64 `json:"acceptanceRate"`   // Driver acceptance rate %
	CancellationRate float64 `json:"cancellationRate"` // Driver cancellation rate %
	TotalTrips       int     `json:"totalTrips"`       // Driver total trips
	IsVerified       bool    `json:"isVerified"`       // Is driver verified
	Status           string  `json:"status"`           // Driver status: online, busy

	// Vehicle Details
	VehicleTypeID      string `json:"vehicleTypeId"`      // Vehicle type ID
	VehicleType        string `json:"vehicleType"`        // e.g., "economy", "comfort"
	VehicleDisplayName string `json:"vehicleDisplayName"` // e.g., "Economy"
	Make               string `json:"make"`               // e.g., "Toyota"
	Model              string `json:"model"`              // e.g., "Corolla"
	Year               int    `json:"year"`               // Year of manufacture
	Color              string `json:"color"`              // e.g., "Silver"
	LicensePlate       string `json:"licensePlate"`       // e.g., "DHA-1234"
	Capacity           int    `json:"capacity"`           // Passenger capacity

	// Location & Distance
	CurrentLatitude  float64 `json:"currentLatitude"`  // Driver current latitude
	CurrentLongitude float64 `json:"currentLongitude"` // Driver current longitude
	Heading          int     `json:"heading"`          // Direction 0-360
	DistanceKm       float64 `json:"distanceKm"`       // Distance from pickup

	// ETA Information
	ETASeconds   int    `json:"etaSeconds"`   // ETA in seconds (to pickup)
	ETAMinutes   int    `json:"etaMinutes"`   // ETA in minutes
	ETAFormatted string `json:"etaFormatted"` // Formatted ETA string e.g., "4 min"

	// Pricing Information
	BaseFare              float64 `json:"baseFare"`              // Base fare for this vehicle type
	PerKmRate             float64 `json:"perKmRate"`             // Rate per km
	PerMinRate            float64 `json:"perMinRate"`            // Rate per minute
	EstimatedFare         float64 `json:"estimatedFare"`         // Estimated fare for trip
	EstimatedDistance     float64 `json:"estimatedDistance"`     // Estimated trip distance
	EstimatedDuration     int     `json:"estimatedDuration"`     // Estimated trip duration in seconds
	EstimatedDurationMins int     `json:"estimatedDurationMins"` // Estimated duration in minutes

	// Surge Pricing
	SurgeMultiplier float64 `json:"surgeMultiplier"` // Current surge multiplier
	SurgeReason     string  `json:"surgeReason"`     // Why surge is active: "peak_hours", "high_demand", etc.

	// Demand Information
	PendingRequests  int    `json:"pendingRequests"`  // Pending ride requests in zone
	AvailableDrivers int    `json:"availableDrivers"` // Available drivers in zone
	Demand           string `json:"demand"`           // Demand level: "low", "normal", "high", "extreme"

	// Timestamps
	UpdatedAt time.Time `json:"updatedAt"` // When driver info was last updated
	Timestamp time.Time `json:"timestamp"` // Response timestamp
}

// VehiclesWithDetailsListResponse represents response with multiple vehicles and pricing summary
type VehiclesWithDetailsListResponse struct {
	// Trip Information
	PickupLat      float64 `json:"pickupLat"`
	PickupLon      float64 `json:"pickupLon"`
	PickupAddress  string  `json:"pickupAddress"`
	DropoffLat     float64 `json:"dropoffLat"`
	DropoffLon     float64 `json:"dropoffLon"`
	DropoffAddress string  `json:"dropoffAddress"`
	RadiusKm       float64 `json:"radiusKm"`

	// Trip Estimates (applies to all vehicles)
	TripDistance     float64 `json:"tripDistance"`     // Total trip distance in km
	TripDuration     int     `json:"tripDuration"`     // Total trip duration in seconds
	TripDurationMins int     `json:"tripDurationMins"` // Trip duration in minutes

	// Vehicles List
	TotalCount int                           `json:"totalCount"` // Total vehicles found
	CarsCount  int                           `json:"carsCount"`  // Number of vehicles in response
	Vehicles   []*VehicleWithDetailsResponse `json:"vehicles"`   // List of vehicles

	// Metadata
	Timestamp time.Time `json:"timestamp"`
}
