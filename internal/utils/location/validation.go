package location

import (
	"errors"
)

func ValidateCoordinates(lat, lon float64) error {
	if lat < -90 || lat > 90 {
		return errors.New("latitude must be between -90 and 90")
	}
	if lon < -180 || lon > 180 {
		return errors.New("longitude must be between -180 and 180")
	}
	return nil
}

func ValidateHeading(heading int) error {
	if heading < 0 || heading > 360 {
		return errors.New("heading must be between 0 and 360")
	}
	return nil
}

func NormalizeHeading(heading int) int {
	normalized := heading % 360
	if normalized < 0 {
		normalized += 360
	}
	return normalized
}

func ValidateSpeed(speed float64) error {
	if speed < 0 || speed > 300 {
		return errors.New("speed must be between 0 and 300 km/h")
	}
	return nil
}

func ValidateAccuracy(accuracy float64) error {
	if accuracy < 0 || accuracy > 10000 {
		return errors.New("accuracy must be between 0 and 10000 meters")
	}
	return nil
}

func IsValidLocation(lat, lon float64) bool {
	return lat >= -90 && lat <= 90 && lon >= -180 && lon <= 180
}
