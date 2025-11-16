package location

import (
	"strings"
)

const (
	base32 = "0123456789bcdefghjkmnpqrstuvwxyz"
)

// Encode generates a geohash for given coordinates
// precision: length of geohash string (1-12, default 9)
func Encode(latitude, longitude float64, precision int) string {
	if precision <= 0 || precision > 12 {
		precision = 9
	}

	var geohash strings.Builder
	var bits uint
	var bit uint
	var even = true

	latMin, latMax := -90.0, 90.0
	lonMin, lonMax := -180.0, 180.0

	for geohash.Len() < precision {
		if even {
			mid := (lonMin + lonMax) / 2
			if longitude > mid {
				bit = 1
				lonMin = mid
			} else {
				bit = 0
				lonMax = mid
			}
		} else {
			mid := (latMin + latMax) / 2
			if latitude > mid {
				bit = 1
				latMin = mid
			} else {
				bit = 0
				latMax = mid
			}
		}

		bits = (bits << 1) | bit
		if (even && bits >= 16) || (!even && bits >= 16) {
			geohash.WriteByte(base32[bits])
			bits = 0
		}
		even = !even
	}

	return geohash.String()
}

// GetGeohashPrecision returns appropriate geohash precision for distance
// Returns precision level (1-12) based on search radius
func GetGeohashPrecision(radiusKm float64) int {
	switch {
	case radiusKm <= 0.019: // ~19m
		return 9
	case radiusKm <= 0.076: // ~76m
		return 8
	case radiusKm <= 0.61: // ~610m
		return 7
	case radiusKm <= 2.4: // ~2.4km
		return 6
	case radiusKm <= 20: // ~20km
		return 5
	case radiusKm <= 78: // ~78km
		return 4
	case radiusKm <= 630: // ~630km
		return 3
	default:
		return 2
	}
}
