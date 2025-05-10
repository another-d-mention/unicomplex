package strutil

import (
	"testing"
)

func TestTextWrap(t *testing.T) {
	tests := []struct {
		input, expected string
		width           int
	}{
		{"Hello, world!", "Hello,\nworld!", 10},
		{"Hello, world!", "Hello, world!", 20},
		{"", "", 10},
		{"  ", "", 10},
		{"Hello", "Hello", 10},
		{"Hello\nworld!", "Hello\nworld!", 10},
		{"One\nTwo", "One\nTwo", 10},
	}

	for _, tt := range tests {
		if got := TextWrap(tt.input, tt.width); got != tt.expected {
			t.Errorf("TextWrap(%q, %d) = %q, want %q", tt.input, tt.width, got, tt.expected)
		}
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input, expected string
		length          int
	}{
		{"Hello, world!", "Hello, world!", 20},
		{"Hello, world!", "Hello, wo...", 12},
		{"Hello, world!", "Hello...", 8},
		{"", "", 10},
		{"  ", "  ", 10},
		{"Hello", "Hello", 10},
		{"Hello\nworld!", "Hello\nw...", 10},
		{"One\nTwo", "One\nTwo", 10},
	}
	for _, tt := range tests {
		if got := TruncateLine(tt.input, tt.length); got != tt.expected {
			t.Errorf("Truncate(%q, %d) = %q, want %q", tt.input, tt.length, got, tt.expected)
		}
	}
}

func TestTruncateLines(t *testing.T) {

}
