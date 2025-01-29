package utils

import (
	"fmt"
	"time"
)

// Helper function to parse datetime with multiple formats
func ParseDatetime(dateStr string, formats []string, isStartDate bool, defaultDate time.Time) (time.Time, error) {
	if dateStr == "" {
		return defaultDate, nil
	}

	for _, format := range formats {
		parsedDate, err := time.Parse(format, dateStr)
		if err != nil {
			continue
		}

		// Handle start and end date logic for date only format (%Y-%m-%d)
		if format == "%Y-%m-%d" {
			if isStartDate {
				parsedDate = time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(), 0, 0, 0, 0, time.UTC)
			} else {
				parsedDate = time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(), 23, 59, 59, 999999999, time.UTC)
			}
		}

		return parsedDate, nil
	}

	return time.Time{}, fmt.Errorf("time data '%s' does not match any supported format", dateStr)
}
