// Package utils provides internal utility functions
package utils

import (
	"fmt"
	"strings"
)

// NormalizeDeviceID normalizes a device ID for consistent comparison
func NormalizeDeviceID(id string) string {
	return strings.TrimSpace(strings.ToLower(id))
}

// ClampVolume ensures volume is within valid range [0.0, 1.0]
func ClampVolume(volume float64) float64 {
	if volume < 0.0 {
		return 0.0
	}
	if volume > 1.0 {
		return 1.0
	}
	return volume
}

// FormatDeviceInfo formats device information for display
func FormatDeviceInfo(name string, id string, deviceType string) string {
	return fmt.Sprintf("%s (%s) [%s]", name, id, deviceType)
}