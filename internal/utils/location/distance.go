package location

import (
	"math"
)

const (
	earthRadiusKm = 6371
)

type Point struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

func HaversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	lat1Rad := degreesToRadians(lat1)
	lat2Rad := degreesToRadians(lat2)
	deltaLat := degreesToRadians(lat2 - lat1)
	deltaLon := degreesToRadians(lon2 - lon1)

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusKm * c
}

func CalculateDistance(from, to Point) float64 {
	return HaversineDistance(from.Latitude, from.Longitude, to.Latitude, to.Longitude)
}

func degreesToRadians(degrees float64) float64 {
	return degrees * math.Pi / 180
}

func radiansToDegrees(radians float64) float64 {
	return radians * 180 / math.Pi
}

func CalculateBearing(lat1, lon1, lat2, lon2 float64) float64 {
	lat1Rad := degreesToRadians(lat1)
	lat2Rad := degreesToRadians(lat2)
	deltaLon := degreesToRadians(lon2 - lon1)

	y := math.Sin(deltaLon) * math.Cos(lat2Rad)
	x := math.Cos(lat1Rad)*math.Sin(lat2Rad) -
		math.Sin(lat1Rad)*math.Cos(lat2Rad)*math.Cos(deltaLon)

	bearing := math.Atan2(y, x)
	bearingDegrees := radiansToDegrees(bearing)

	return math.Mod(bearingDegrees+360, 360)
}

func CalculateETAWithTraffic(distanceKm, speedKmh, trafficMultiplier float64) int {
	baseETA := CalculateETA(distanceKm, speedKmh)
	return int(float64(baseETA) * trafficMultiplier)
}

func IsWithinRadius(centerLat, centerLon, pointLat, pointLon, radiusKm float64) bool {
	distance := HaversineDistance(centerLat, centerLon, pointLat, pointLon)
	return distance <= radiusKm
}

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

func CalculateETA(distance, speed float64) int {
	if speed <= 0 {
		speed = 40 
	}
	hours := distance / speed
	seconds := hours * 3600
	return int(math.Round(seconds))
}

func toRadians(degrees float64) float64 {
	return degrees * math.Pi / 180
}

func EncodePolyline(points []Point) string {
	if len(points) == 0 {
		return ""
	}

	var encoded []byte
	prevLat := 0
	prevLon := 0

	for _, point := range points {
		lat := int(math.Round(point.Latitude * 1e5))
		lon := int(math.Round(point.Longitude * 1e5))

		encoded = append(encoded, encodeValue(lat-prevLat)...)
		encoded = append(encoded, encodeValue(lon-prevLon)...)

		prevLat = lat
		prevLon = lon
	}

	return string(encoded)
}

func DecodePolyline(encoded string) []Point {
	if len(encoded) == 0 {
		return nil
	}

	var points []Point
	index := 0
	lat := 0
	lon := 0

	for index < len(encoded) {
		result := 1
		shift := 0
		var b int

		for {
			b = int(encoded[index]) - 63 - 1
			index++
			result += b << shift
			shift += 5
			if b < 0x1f {
				break
			}
		}

		lat += (result & 1) ^ (-(result >> 1))

		result = 1
		shift = 0

		for {
			b = int(encoded[index]) - 63 - 1
			index++
			result += b << shift
			shift += 5
			if b < 0x1f {
				break
			}
		}

		lon += (result & 1) ^ (-(result >> 1))

		points = append(points, Point{
			Latitude:  float64(lat) / 1e5,
			Longitude: float64(lon) / 1e5,
		})
	}

	return points
}

func encodeValue(value int) []byte {
	value = value << 1

	if value < 0 {
		value = ^value
	}

	var encoded []byte

	for value >= 0x20 {
		encoded = append(encoded, byte((0x20|(value&0x1f))+63))
		value >>= 5
	}

	encoded = append(encoded, byte(value+63))
	return encoded
}

func SimplifyPolyline(points []Point, tolerance float64) []Point {
	if len(points) <= 2 {
		return points
	}

	maxDistance := 0.0
	maxIndex := 0

	for i := 1; i < len(points)-1; i++ {
		distance := perpendicularDistance(points[i], points[0], points[len(points)-1])
		if distance > maxDistance {
			maxDistance = distance
			maxIndex = i
		}
	}

	if maxDistance > tolerance {
		left := SimplifyPolyline(points[:maxIndex+1], tolerance)
		right := SimplifyPolyline(points[maxIndex:], tolerance)

		return append(left[:len(left)-1], right...)
	}

	return []Point{points[0], points[len(points)-1]}
}

func perpendicularDistance(point, lineStart, lineEnd Point) float64 {
	if lineStart.Latitude == lineEnd.Latitude && lineStart.Longitude == lineEnd.Longitude {
		return CalculateDistance(point, lineStart)
	}

	dx := lineEnd.Longitude - lineStart.Longitude
	dy := lineEnd.Latitude - lineStart.Latitude

	num := math.Abs(dy*point.Longitude - dx*point.Latitude + lineEnd.Longitude*lineStart.Latitude - lineEnd.Latitude*lineStart.Longitude)
	den := math.Sqrt(dy*dy + dx*dx)

	return num / den * 111.32
}
