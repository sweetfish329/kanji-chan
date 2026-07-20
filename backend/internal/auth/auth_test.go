package auth

import (
	"testing"
)

func TestSecureCompare(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		want     bool
	}{
		{"identical strings", "secret_token_123", "secret_token_123", true},
		{"empty strings", "", "", true},
		{"different length", "secret_token_123", "secret_token_12", false},
		{"one character different", "secret_token_123", "secret_token_124", false},
		{"completely different", "abc", "xyz", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SecureCompare(tt.input, tt.expected)
			if got != tt.want {
				t.Errorf("SecureCompare(%q, %q) = %v; want %v", tt.input, tt.expected, got, tt.want)
			}
		})
	}
}
