package strutil

import (
	"strings"
)

func TextWrap(text string, width int) string {
	var result strings.Builder
	for _, line := range strings.Split(text, "\n") {
		lineEmpty := true
		for _, word := range strings.Fields(line) {
			if !lineEmpty {
				if result.Len() > 0 && result.String()[result.Len()-1] != '\n' {
					if len(result.String())+len(word)+1 > width {
						result.WriteByte('\n')
					} else {
						result.WriteByte(' ')
					}
				}
			}
			result.WriteString(word)
			lineEmpty = false
		}
		if !lineEmpty {
			result.WriteByte('\n')
		}
	}
	return strings.TrimSpace(result.String())
}

// TruncateLine shortens a string to a specified length, appending "..." if truncation occurs.
//
// Parameters:
//
//	s - The string to be truncated.
//	length - The maximum length of the resulting string, including the ellipsis if truncation occurs.
//
// Returns:
//
//	A string that is either the original string if its length is less than or equal to the specified length,
//	or a truncated version of the string with "..." appended if the original string exceeds the specified length.
func TruncateLine(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length-3] + "..."
}
