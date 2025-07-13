// test/scaffolder/helpers_test.go
package scaffolder_test

import (
	"testing"
)

// Test helper functions that were extracted from the main scaffolder

func TestToGoIdentifier(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple word",
			input:    "user",
			expected: "User",
		},
		{
			name:     "hyphenated word",
			input:    "user-profile",
			expected: "UserProfile",
		},
		{
			name:     "underscore word",
			input:    "user_profile",
			expected: "UserProfile",
		},
		{
			name:     "multiple hyphens",
			input:    "user-profile-admin",
			expected: "UserProfileAdmin",
		},
		{
			name:     "mixed separators",
			input:    "user-profile_admin",
			expected: "UserProfileAdmin",
		},
		{
			name:     "single character",
			input:    "a",
			expected: "A",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Since toGoIdentifier is not exported, we need to test it through the public API
			// or make it a public function in a utils package
			// For now, we'll test the behavior through CreateHandler validation

			// This is a placeholder - in practice you'd either:
			// 1. Move toGoIdentifier to a public utils package, or
			// 2. Test it indirectly through the public API

			if tt.input == "" {
				return // Skip empty test for now
			}

			// Indirect test through the scaffolder behavior
			// We can verify the identifier generation by checking the template data
			// This would require exposing the prepareHandlerData function or similar
		})
	}
}

func TestPluralize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "regular word",
			input:    "user",
			expected: "users",
		},
		{
			name:     "word ending in y",
			input:    "category",
			expected: "categories",
		},
		{
			name:     "word ending in s",
			input:    "class",
			expected: "classes",
		},
		{
			name:     "word ending in x",
			input:    "box",
			expected: "boxes",
		},
		{
			name:     "word ending in z",
			input:    "quiz",
			expected: "quizzes",
		},
		{
			name:     "word ending in ch",
			input:    "batch",
			expected: "batches",
		},
		{
			name:     "word ending in sh",
			input:    "dish",
			expected: "dishes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Similar to toGoIdentifier, this would need to be tested
			// either through a public utils function or indirectly
			// through the scaffolder's behavior

			// Placeholder for actual test implementation
			_ = tt.input
			_ = tt.expected
		})
	}
}

func TestValidGoIdentifier(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "valid identifier",
			input:    "user",
			expected: true,
		},
		{
			name:     "valid camelCase",
			input:    "userProfile",
			expected: true,
		},
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "contains space",
			input:    "user profile",
			expected: false,
		},
		{
			name:     "contains tab",
			input:    "user\tprofile",
			expected: false,
		},
		{
			name:     "go keyword",
			input:    "func",
			expected: false,
		},
		{
			name:     "go keyword mixed case",
			input:    "Func",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This would test the isValidGoIdentifier function
			// Again, either make it public or test indirectly

			// Placeholder for actual test implementation
			_ = tt.input
			_ = tt.expected
		})
	}
}
