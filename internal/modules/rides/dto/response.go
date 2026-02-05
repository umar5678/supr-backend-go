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

	DriverFare *float64 `json:"driverFare,omitempty"`
	RiderFare  *float64 `json:"riderFare,omitempty"` 

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

type AvailableCarResponse struct {
	ID                 string    `json:"id"`          
	DriverID           string    `json:"driverId"`    
	DriverName         string    `json:"driverName"`  
	DriverRating       float64   `json:"driverRating"`
	DriverImage        *string   `json:"driverImage,omitempty"`
	VehicleTypeID      string    `json:"vehicleTypeId"`
	VehicleType        string    `json:"vehicleType"`
	VehicleDisplayName string    `json:"vehicleDisplayName"`
	Make               string    `json:"make"`           
	Model              string    `json:"model"`          
	Color              string    `json:"color"`          
	LicensePlate       string    `json:"licensePlate"`   
	Capacity           int       `json:"capacity"`       
	CurrentLatitude    float64   `json:"currentLatitude"`
	CurrentLongitude   float64   `json:"currentLongitude"`
	Heading            int       `json:"heading"`          
	DistanceKm         float64   `json:"distanceKm"`       
	ETASeconds         int       `json:"etaSeconds"`       
	ETAMinutes         int       `json:"etaMinutes"`       
	EstimatedFare      float64   `json:"estimatedFare"`    
	SurgeMultiplier    float64   `json:"surgeMultiplier"`  
	AcceptanceRate     float64   `json:"acceptanceRate"`   
	CancellationRate   float64   `json:"cancellationRate"` 
	TotalTrips         int       `json:"totalTrips"`       
	Status             string    `json:"status"`           
	IsVerified         bool      `json:"isVerified"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

type AvailableCarsListResponse struct {
	TotalCount int                     `json:"totalCount"` 
	CarsCount  int                     `json:"carsCount"`  
	RiderLat   float64                 `json:"riderLat"`
	RiderLon   float64                 `json:"riderLon"`
	RadiusKm   float64                 `json:"radiusKm"`
	Cars       []*AvailableCarResponse `json:"cars"`
	Timestamp  time.Time               `json:"timestamp"`
}

type WebSocketAvailableCarsMessage struct {
	Type      string                     `json:"type"`
	Data      *AvailableCarsListResponse `json:"data,omitempty"`
	Error     string                     `json:"error,omitempty"`
	Timestamp time.Time                  `json:"timestamp"`
}

type VehicleWithDetailsResponse struct {
	ID               string  `json:"id"`          
	DriverID         string  `json:"driverId"`    
	DriverName       string  `json:"driverName"`  
	DriverRating     float64 `json:"driverRating"`
	DriverImage      *string `json:"driverImage,omitempty"`
	AcceptanceRate   float64 `json:"acceptanceRate"`   
	CancellationRate float64 `json:"cancellationRate"` 
	TotalTrips       int     `json:"totalTrips"`       
	IsVerified       bool    `json:"isVerified"`       
	Status           string  `json:"status"`           

	VehicleTypeID      string `json:"vehicleTypeId"`      
	VehicleType        string `json:"vehicleType"`        
	VehicleDisplayName string `json:"vehicleDisplayName"` 
	Make               string `json:"make"`               
	Model              string `json:"model"`              
	Year               int    `json:"year"`               
	Color              string `json:"color"`              
	LicensePlate       string `json:"licensePlate"`       
	Capacity           int    `json:"capacity"`           

	CurrentLatitude  float64 `json:"currentLatitude"` 
	CurrentLongitude float64 `json:"currentLongitude"`
	Heading          int     `json:"heading"`         
	DistanceKm       float64 `json:"distanceKm"`      

	ETASeconds   int    `json:"etaSeconds"`  
	ETAMinutes   int    `json:"etaMinutes"`  
	ETAFormatted string `json:"etaFormatted"`

	BaseFare              float64 `json:"baseFare"`              
	PerKmRate             float64 `json:"perKmRate"`             
	PerMinRate            float64 `json:"perMinRate"`            
	EstimatedFare         float64 `json:"estimatedFare"`         
	EstimatedDistance     float64 `json:"estimatedDistance"`     
	EstimatedDuration     int     `json:"estimatedDuration"`     
	EstimatedDurationMins int     `json:"estimatedDurationMins"` 

	SurgeMultiplier float64 `json:"surgeMultiplier"` 
	SurgeReason     string  `json:"surgeReason"`     

	PendingRequests  int    `json:"pendingRequests"` 
	AvailableDrivers int    `json:"availableDrivers"`
	Demand           string `json:"demand"`          

	UpdatedAt time.Time `json:"updatedAt"` 
	Timestamp time.Time `json:"timestamp"` 
}

type VehiclesWithDetailsListResponse struct {
	PickupLat      float64 `json:"pickupLat"`
	PickupLon      float64 `json:"pickupLon"`
	PickupAddress  string  `json:"pickupAddress"`
	DropoffLat     float64 `json:"dropoffLat"`
	DropoffLon     float64 `json:"dropoffLon"`
	DropoffAddress string  `json:"dropoffAddress"`
	RadiusKm       float64 `json:"radiusKm"`

	TripDistance     float64 `json:"tripDistance"`     
	TripDuration     int     `json:"tripDuration"`     
	TripDurationMins int     `json:"tripDurationMins"` 

	TotalCount int                           `json:"totalCount"`
	CarsCount  int                           `json:"carsCount"` 
	Vehicles   []*VehicleWithDetailsResponse `json:"vehicles"`  

	Timestamp time.Time `json:"timestamp"`
}
