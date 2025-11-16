package location

import (
	"math"
)

const (
	earthRadiusKm = 6371 // Earth's radius in kilometers
)

// Point represents a geographic coordinate
type Point struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// HaversineDistance calculates the distance between two points using the Haversine formula
// Returns distance in kilometers
func HaversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	// Convert degrees to radians
	lat1Rad := degreesToRadians(lat1)
	lat2Rad := degreesToRadians(lat2)
	deltaLat := degreesToRadians(lat2 - lat1)
	deltaLon := degreesToRadians(lon2 - lon1)

	// Haversine formula
	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusKm * c
}

// CalculateDistance is a convenience wrapper
func CalculateDistance(from, to Point) float64 {
	return HaversineDistance(from.Latitude, from.Longitude, to.Latitude, to.Longitude)
}

// degreesToRadians converts degrees to radians
func degreesToRadians(degrees float64) float64 {
	return degrees * math.Pi / 180
}

// radiansToDegrees converts radians to degrees
func radiansToDegrees(radians float64) float64 {
	return radians * 180 / math.Pi
}

// CalculateBearing calculates the bearing (heading) from point1 to point2
// Returns bearing in degrees (0-360)
func CalculateBearing(lat1, lon1, lat2, lon2 float64) float64 {
	lat1Rad := degreesToRadians(lat1)
	lat2Rad := degreesToRadians(lat2)
	deltaLon := degreesToRadians(lon2 - lon1)

	y := math.Sin(deltaLon) * math.Cos(lat2Rad)
	x := math.Cos(lat1Rad)*math.Sin(lat2Rad) -
		math.Sin(lat1Rad)*math.Cos(lat2Rad)*math.Cos(deltaLon)

	bearing := math.Atan2(y, x)
	bearingDegrees := radiansToDegrees(bearing)

	// Normalize to 0-360
	return math.Mod(bearingDegrees+360, 360)
}

// CalculateETA calculates estimated time of arrival
// distance in km, speed in km/h
// Returns ETA in seconds
func CalculateETA(distanceKm, speedKmh float64) int {
	if speedKmh <= 0 {
		speedKmh = 40 // Default average speed: 40 km/h
	}

	hours := distanceKm / speedKmh
	seconds := hours * 3600

	return int(math.Ceil(seconds))
}

// CalculateETAWithTraffic calculates ETA with traffic multiplier
func CalculateETAWithTraffic(distanceKm, speedKmh, trafficMultiplier float64) int {
	baseETA := CalculateETA(distanceKm, speedKmh)
	return int(float64(baseETA) * trafficMultiplier)
}

// IsWithinRadius checks if a point is within a given radius (in km) of another point
func IsWithinRadius(centerLat, centerLon, pointLat, pointLon, radiusKm float64) bool {
	distance := HaversineDistance(centerLat, centerLon, pointLat, pointLon)
	return distance <= radiusKm
}

// CalculateMidpoint calculates the midpoint between two coordinates
func CalculateMidpoint(lat1, lon1, lat2, lon2 float64) Point {
	lat1Rad := degreesToRadians(lat1)
	lat2Rad := degreesToRadians(lat2)
	lon1Rad := degreesToRadians(lon1)
	deltaLon := degreesToRadians(lon2 - lon1)

	bx := math.Cos(lat2Rad) * math.Cos(deltaLon)
	by := math.Cos(lat2Rad) * math.Sin(deltaLon)

	midLat := math.Atan2(
		math.Sin(lat1Rad)+math.Sin(lat2Rad),
		math.Sqrt((math.Cos(lat1Rad)+bx)*(math.Cos(lat1Rad)+bx)+by*by),
	)

	midLon := lon1Rad + math.Atan2(by, math.Cos(lat1Rad)+bx)

	return Point{
		Latitude:  radiansToDegrees(midLat),
		Longitude: radiansToDegrees(midLon),
	}
}
