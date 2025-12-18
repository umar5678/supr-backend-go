package shared

import (
	"fmt"
	"math"
	"time"
)

// CalculatePlatformCommission calculates the platform commission
func CalculatePlatformCommission(total float64) float64 {
	return RoundToTwoDecimals(total * PlatformCommissionRate)
}

// CalculateProviderEarnings calculates what provider earns after commission
func CalculateProviderEarnings(total float64) float64 {
	commission := CalculatePlatformCommission(total)
	return RoundToTwoDecimals(total - commission)
}

// CalculateCancellationFee calculates the cancellation fee based on order status
func CalculateCancellationFee(status string, totalPrice float64) (cancellationFee, refundAmount float64) {
	var feeRate float64

	switch status {
	case OrderStatusPending, OrderStatusSearchingProvider:
		feeRate = CancellationFeeBeforeAcceptance
	case OrderStatusAssigned, OrderStatusAccepted:
		feeRate = CancellationFeeAfterAcceptance
	case OrderStatusInProgress:
		feeRate = CancellationFeeAfterStart
	default:
		// Cannot cancel
		return 0, 0
	}

	cancellationFee = RoundToTwoDecimals(totalPrice * feeRate)
	refundAmount = RoundToTwoDecimals(totalPrice - cancellationFee)

	return cancellationFee, refundAmount
}

// RoundToTwoDecimals rounds a float to two decimal places
func RoundToTwoDecimals(value float64) float64 {
	return math.Round(value*100) / 100
}

// CalculateServicesTotal calculates total price for selected services
func CalculateServicesTotal(services []ServiceItem) float64 {
	var total float64
	for _, s := range services {
		total += s.Price * float64(s.Quantity)
	}
	return RoundToTwoDecimals(total)
}

// CalculateAddonsTotal calculates total price for selected addons
func CalculateAddonsTotal(addons []AddonItem) float64 {
	var total float64
	for _, a := range addons {
		total += a.Price * float64(a.Quantity)
	}
	return RoundToTwoDecimals(total)
}

// ServiceItem represents a service for calculation
type ServiceItem struct {
	Price    float64
	Quantity int
}

// AddonItem represents an addon for calculation
type AddonItem struct {
	Price    float64
	Quantity int
}

// CalculateOrderExpiration calculates when an order should expire
func CalculateOrderExpiration() time.Time {
	return time.Now().Add(time.Duration(OrderExpirationMinutes) * time.Minute)
}

// IsOrderExpired checks if an order has expired
func IsOrderExpired(expiresAt *time.Time) bool {
	if expiresAt == nil {
		return false
	}
	return time.Now().After(*expiresAt)
}

// GenerateOrderNumber generates a unique order number
// Format: HS-YYYY-XXXXXX (handled by database trigger, this is a fallback)
func GenerateOrderNumber() string {
	year := time.Now().Year()
	timestamp := time.Now().UnixNano() / 1000000 // milliseconds
	return fmt.Sprintf("HS-%d-%06d", year, timestamp%1000000)
}

// ParseBookingDateTime parses booking date and time into a time.Time
func ParseBookingDateTime(date, timeStr string) (time.Time, error) {
	dateTimeStr := fmt.Sprintf("%s %s", date, timeStr)
	return time.Parse("2006-01-02 15:04", dateTimeStr)
}

// IsBookingTimeValid checks if the booking time is in the future
func IsBookingTimeValid(date, timeStr string) bool {
	bookingTime, err := ParseBookingDateTime(date, timeStr)
	if err != nil {
		return false
	}
	// Must be at least 2 hours in the future
	minBookingTime := time.Now().Add(2 * time.Hour)
	return bookingTime.After(minBookingTime)
}

// GetDayOfWeek returns the day of week for a date string
func GetDayOfWeek(dateStr string) (string, error) {
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return "", err
	}
	return date.Weekday().String(), nil
}

// ValidatePreferredTime validates preferred time value
func ValidatePreferredTime(preferredTime string) bool {
	validTimes := []string{
		PreferredTimeMorning,
		PreferredTimeAfternoon,
		PreferredTimeEvening,
	}
	for _, t := range validTimes {
		if t == preferredTime {
			return true
		}
	}
	return false
}

// Haversine calculates distance between two coordinates in kilometers
func Haversine(lat1, lng1, lat2, lng2 float64) float64 {
	const earthRadius = 6371 // km

	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	deltaLat := (lat2 - lat1) * math.Pi / 180
	deltaLng := (lng2 - lng1) * math.Pi / 180

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLng/2)*math.Sin(deltaLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}

// CalculateProviderScore calculates a score for provider matching
func CalculateProviderScore(
	rating float64,
	completedJobs int,
	totalJobs int,
	distance float64, // in km
	avgResponseTime float64, // in minutes
) float64 {
	score := 0.0

	// Rating weight: 40% (rating is 1-5, normalize to 0-40)
	score += rating * 8.0

	// Completion rate weight: 20%
	if totalJobs > 0 {
		completionRate := float64(completedJobs) / float64(totalJobs)
		score += completionRate * 20.0
	}

	// Distance weight: 30% (closer is better, max 50km)
	distanceScore := math.Max(0, 30-distance*0.6)
	score += distanceScore

	// Response time weight: 10% (faster is better)
	responseScore := math.Max(0, 10-avgResponseTime/6) // 60 min = 0 points
	score += responseScore

	return score
}

// StringPtr returns a pointer to a string
func StringPtr(s string) *string {
	return &s
}

// IntPtr returns a pointer to an int
func IntPtr(i int) *int {
	return &i
}

// Float64Ptr returns a pointer to a float64
func Float64Ptr(f float64) *float64 {
	return &f
}

// BoolPtr returns a pointer to a bool
func BoolPtr(b bool) *bool {
	return &b
}

// TimePtr returns a pointer to a time.Time
func TimePtr(t time.Time) *time.Time {
	return &t
}
