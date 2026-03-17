package shared

import (
	"fmt"
	"math"
	"time"
)

func CalculatePlatformCommission(total float64) float64 {
	return RoundToTwoDecimals(total * PlatformCommissionRate)
}

func CalculateProviderEarnings(total float64) float64 {
	commission := CalculatePlatformCommission(total)
	return RoundToTwoDecimals(total - commission)
}

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
		return 0, 0
	}

	cancellationFee = RoundToTwoDecimals(totalPrice * feeRate)
	refundAmount = RoundToTwoDecimals(totalPrice - cancellationFee)

	return cancellationFee, refundAmount
}

func RoundToTwoDecimals(value float64) float64 {
	return math.Round(value*100) / 100
}
func CalculateServicesTotal(services []ServiceItem) float64 {
	var total float64
	for _, s := range services {
		total += s.Price * float64(s.Quantity)
	}
	return RoundToTwoDecimals(total)
}

func CalculateAddonsTotal(addons []AddonItem) float64 {
	var total float64
	for _, a := range addons {
		total += a.Price * float64(a.Quantity)
	}
	return RoundToTwoDecimals(total)
}
type ServiceItem struct {
	Price    float64
	Quantity int
}

type AddonItem struct {
	Price    float64
	Quantity int
}

func CalculateOrderExpiration() time.Time {
	return time.Now().Add(time.Duration(OrderExpirationMinutes) * time.Minute)
}
func IsOrderExpired(expiresAt *time.Time) bool {
	if expiresAt == nil {
		return false
	}
	return time.Now().After(*expiresAt)
}

func GenerateOrderNumber() string {
	year := time.Now().Year()
	timestamp := time.Now().UnixNano() / 1000000
	return fmt.Sprintf("HS-%d-%06d", year, timestamp%1000000)
}
func ParseBookingDateTime(date, timeStr string) (time.Time, error) {
	dateTimeStr := fmt.Sprintf("%s %s", date, timeStr)
	return time.Parse("2006-01-02 15:04", dateTimeStr)
}
func IsBookingTimeValid(date, timeStr string) bool {
	bookingTime, err := ParseBookingDateTime(date, timeStr)
	if err != nil {
		return false
	}
	minBookingTime := time.Now().Add(2 * time.Hour)
	return bookingTime.After(minBookingTime)
}

func GetDayOfWeek(dateStr string) (string, error) {
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return "", err
	}
	return date.Weekday().String(), nil
}
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

func CalculateProviderScore(
	rating float64,
	completedJobs int,
	totalJobs int,
	distance float64, 
	avgResponseTime float64,
) float64 {
	score := 0.0

	score += rating * 8.0

	if totalJobs > 0 {
		completionRate := float64(completedJobs) / float64(totalJobs)
		score += completionRate * 20.0
	}

	distanceScore := math.Max(0, 30-distance*0.6)
	score += distanceScore

	responseScore := math.Max(0, 10-avgResponseTime/6) 
	score += responseScore

	return score
}

func StringPtr(s string) *string {
	return &s
}

func IntPtr(i int) *int {
	return &i
}

func Float64Ptr(f float64) *float64 {
	return &f
}

func BoolPtr(b bool) *bool {
	return &b
}

func TimePtr(t time.Time) *time.Time {
	return &t
}
