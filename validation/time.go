package validation

import (
	"time"
)

// TimeInRange checks if a given time is within a specified range.
// If start is zero, only checks if dst is before or equal to end.
// If end is zero, only checks if dst is after or equal to start.
func TimeInRange(dst, start, end time.Time) bool {
	if start.IsZero() && end.IsZero() {
		return false
	}
	if start.IsZero() {
		return dst.Before(end) || dst.Equal(end)
	}
	if end.IsZero() {
		return dst.After(start) || dst.Equal(start)
	}
	return dst.After(start) && dst.Before(end) || dst.Equal(start) || dst.Equal(end)
}
