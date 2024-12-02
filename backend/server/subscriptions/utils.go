package subscriptions

import (
	"time"
)

// AddDate is similar to time.AddDate, but defaults to maxing out an available
// month's number of days rather than rolling over into the following month.
// For example: January 31 + 1 month will be February 28.
func AddDate(years int, months int) time.Time {
	now := time.Now()
	future := now.AddDate(years, months, 0)
	if d := future.Day(); d != now.Day() {
		return future.AddDate(0, 0, -d)
	}

	return future
}
