// File: test/cmd/foundry/cmd/validation_test.go

package cmd

import (
	"github.com/shapestone/foundry/cmd/foundry/cmd"
	"strings"
	"testing"
)

func TestValidateComponentName(t *testing.T) {
	tests := []struct {
		name          string
		componentName string
		expectError   bool
		errorContains string
	}{
		// Happy path cases
		{
			name:          "ValidSimpleName_ReturnsNoError",
			componentName: "user",
			expectError:   false,
		},
		{
			name:          "ValidCamelCaseName_ReturnsNoError",
			componentName: "userProfile",
			expectError:   false,
		},
		{
			name:          "ValidKebabCaseName_ReturnsNoError",
			componentName: "user-profile",
			expectError:   false,
		},
		{
			name:          "ValidSnakeCaseName_ReturnsNoError",
			componentName: "user_profile",
			expectError:   false,
		},
		{
			name:          "ValidNameWithNumbers_ReturnsNoError",
			componentName: "user123",
			expectError:   false,
		},
		{
			name:          "ValidMixedCase_ReturnsNoError",
			componentName: "UserProfile",
			expectError:   false,
		},

		// Error cases - empty and length
		{
			name:          "EmptyName_ReturnsError",
			componentName: "",
			expectError:   true,
			errorContains: "cannot be empty",
		},
		{
			name:          "TooLongName_ReturnsError",
			componentName: strings.Repeat("a", 51),
			expectError:   true,
			errorContains: "too long",
		},
		{
			name:          "TooShortName_ReturnsError",
			componentName: "a",
			expectError:   true,
			errorContains: "too short",
		},

		// Error cases - whitespace
		{
			name:          "NameWithSpaces_ReturnsError",
			componentName: "user profile",
			expectError:   true,
			errorContains: "cannot contain spaces",
		},
		{
			name:          "NameWithTab_ReturnsError",
			componentName: "user\tprofile",
			expectError:   true,
			errorContains: "cannot contain whitespace",
		},
		{
			name:          "NameWithNewline_ReturnsError",
			componentName: "user\nprofile",
			expectError:   true,
			errorContains: "cannot contain whitespace",
		},

		// Error cases - Go keywords
		{
			name:          "GoKeywordFunc_ReturnsError",
			componentName: "func",
			expectError:   true,
			errorContains: "Go reserved keyword",
		},
		{
			name:          "GoKeywordVar_ReturnsError",
			componentName: "var",
			expectError:   true,
			errorContains: "Go reserved keyword",
		},
		{
			name:          "GoKeywordInterface_ReturnsError",
			componentName: "interface",
			expectError:   true,
			errorContains: "Go reserved keyword",
		},

		// Error cases - invalid patterns
		{
			name:          "StartsWithNumber_ReturnsError",
			componentName: "123user",
			expectError:   true,
			errorContains: "cannot start with a number",
		},
		{
			name:          "InvalidCharacters_ReturnsError",
			componentName: "user@profile",
			expectError:   true,
			errorContains: "can only contain letters, numbers, underscores, and hyphens",
		},
		{
			name:          "ConsecutiveHyphens_ReturnsError",
			componentName: "user--profile",
			expectError:   true,
			errorContains: "cannot contain consecutive hyphens",
		},
		{
			name:          "ConsecutiveUnderscores_ReturnsError",
			componentName: "user__profile",
			expectError:   true,
			errorContains: "cannot contain consecutive underscores",
		},
		{
			name:          "StartsWithHyphen_ReturnsError",
			componentName: "-user",
			expectError:   true,
			errorContains: "cannot start with a hyphen",
		},
		{
			name:          "EndsWithHyphen_ReturnsError",
			componentName: "user-",
			expectError:   true,
			errorContains: "cannot end with a hyphen",
		},
		{
			name:          "StartsWithUnderscore_ReturnsError",
			componentName: "_user",
			expectError:   true,
			errorContains: "cannot start with an underscore",
		},
		{
			name:          "EndsWithUnderscore_ReturnsError",
			componentName: "user_",
			expectError:   true,
			errorContains: "cannot end with an underscore",
		},

		// Error cases - problematic names
		{
			name:          "ProblematicNameMain_ReturnsError",
			componentName: "main",
			expectError:   true,
			errorContains: "conflicts with main package",
		},
		{
			name:          "ProblematicNameTest_ReturnsError",
			componentName: "test",
			expectError:   true,
			errorContains: "conflicts with Go testing",
		},
		{
			name:          "ProblematicNameString_ReturnsError",
			componentName: "string",
			expectError:   true,
			errorContains: "conflicts with built-in string type",
		},
		{
			name:          "ProblematicNameError_ReturnsError",
			componentName: "error",
			expectError:   true,
			errorContains: "conflicts with built-in error type",
		},

		// Edge cases
		{
			name:          "MinimumValidLength_ReturnsNoError",
			componentName: "ab",
			expectError:   false,
		},
		{
			name:          "MaximumValidLength_ReturnsNoError",
			componentName: strings.Repeat("a", 50),
			expectError:   false,
		},
		{
			name:          "CaseInsensitiveKeywordCheck_ReturnsError",
			componentName: "FUNC",
			expectError:   true,
			errorContains: "Go reserved keyword",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			err := cmd.ValidateComponentName(tt.componentName)

			// Assert
			if tt.expectError {
				if err == nil {
					t.Errorf("ValidateComponentName(%q) expected error, got nil", tt.componentName)
					return
				}
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("ValidateComponentName(%q) error = %v, expected to contain %q", tt.componentName, err, tt.errorContains)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateComponentName(%q) unexpected error = %v", tt.componentName, err)
				}
			}
		})
	}
}

func TestToGoIdentifier(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "SimpleKebabCase_ConvertsToCapitalizedSnakeCase",
			input:    "user-profile",
			expected: "User_profile",
		},
		{
			name:     "MultipleHyphens_ReplacesAllWithUnderscores",
			input:    "user-profile-data",
			expected: "User_profile_data",
		},
		{
			name:     "NoHyphens_CapitalizesFirstLetter",
			input:    "user",
			expected: "User",
		},
		{
			name:     "AlreadyCapitalized_RemainsCapitalized",
			input:    "User",
			expected: "User",
		},
		{
			name:     "EmptyString_ReturnsEmptyString",
			input:    "",
			expected: "",
		},
		{
			name:     "SingleCharacter_CapitalizesChar",
			input:    "a",
			expected: "A",
		},
		{
			name:     "MixedCaseWithHyphens_ConvertsCorrectly",
			input:    "userProfile-Data",
			expected: "UserProfile_Data",
		},
		{
			name:     "NumbersAndHyphens_HandlesCorrectly",
			input:    "user-123-profile",
			expected: "User_123_profile",
		},
		{
			name:     "UnderscoresPreserved_CapitalizesOnly",
			input:    "user_profile",
			expected: "User_profile",
		},
		{
			name:     "HyphensAndUnderscoresMixed_ConvertsHyphensOnly",
			input:    "user-profile_data",
			expected: "User_profile_data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := cmd.ToGoIdentifier(tt.input)

			// Assert
			if result != tt.expected {
				t.Errorf("ToGoIdentifier(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestValidateComponentType(t *testing.T) {
	tests := []struct {
		name          string
		componentType string
		expectError   bool
		errorContains string
	}{
		// Happy path cases
		{
			name:          "ValidTypeHandler_ReturnsNoError",
			componentType: "handler",
			expectError:   false,
		},
		{
			name:          "ValidTypeModel_ReturnsNoError",
			componentType: "model",
			expectError:   false,
		},
		{
			name:          "ValidTypeMiddleware_ReturnsNoError",
			componentType: "middleware",
			expectError:   false,
		},

		// Error cases
		{
			name:          "InvalidType_ReturnsError",
			componentType: "invalid",
			expectError:   true,
			errorContains: "unsupported component type",
		},
		{
			name:          "EmptyType_ReturnsError",
			componentType: "",
			expectError:   true,
			errorContains: "unsupported component type",
		},
		{
			name:          "CaseSensitive_ReturnsError",
			componentType: "Handler",
			expectError:   true,
			errorContains: "unsupported component type",
		},
		{
			name:          "RandomString_ReturnsError",
			componentType: "randomtype",
			expectError:   true,
			errorContains: "unsupported component type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			err := cmd.ValidateComponentType(tt.componentType)

			// Assert
			if tt.expectError {
				if err == nil {
					t.Errorf("ValidateComponentType(%q) expected error, got nil", tt.componentType)
					return
				}
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("ValidateComponentType(%q) error = %v, expected to contain %q", tt.componentType, err, tt.errorContains)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateComponentType(%q) unexpected error = %v", tt.componentType, err)
				}
			}
		})
	}
}

func TestValidateAddCommandArgs(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		expectError   bool
		errorContains string
	}{
		// Happy path cases
		{
			name:        "ValidHandlerCommand_ReturnsNoError",
			args:        []string{"handler", "user"},
			expectError: false,
		},
		{
			name:        "ValidModelCommand_ReturnsNoError",
			args:        []string{"model", "product"},
			expectError: false,
		},
		{
			name:        "ValidMiddlewareCommand_ReturnsNoError",
			args:        []string{"middleware", "auth"},
			expectError: false,
		},

		// Error cases - insufficient arguments
		{
			name:          "NoArguments_ReturnsError",
			args:          []string{},
			expectError:   true,
			errorContains: "requires component type and name",
		},
		{
			name:          "OnlyOneArgument_ReturnsError",
			args:          []string{"handler"},
			expectError:   true,
			errorContains: "requires component type and name",
		},

		// Error cases - invalid component type
		{
			name:          "InvalidComponentType_ReturnsError",
			args:          []string{"invalid", "user"},
			expectError:   true,
			errorContains: "unsupported component type",
		},

		// Error cases - invalid component name
		{
			name:          "InvalidComponentName_ReturnsError",
			args:          []string{"handler", "123invalid"},
			expectError:   true,
			errorContains: "cannot start with a number",
		},
		{
			name:          "EmptyComponentName_ReturnsError",
			args:          []string{"handler", ""},
			expectError:   true,
			errorContains: "cannot be empty",
		},
		{
			name:          "ComponentNameWithSpaces_ReturnsError",
			args:          []string{"handler", "user profile"},
			expectError:   true,
			errorContains: "cannot contain spaces",
		},

		// Edge cases
		{
			name:        "ExtraArguments_ValidatesFirstTwo",
			args:        []string{"handler", "user", "extra", "args"},
			expectError: false,
		},
		{
			name:          "ValidTypeInvalidName_ReturnsNameError",
			args:          []string{"model", "func"},
			expectError:   true,
			errorContains: "Go reserved keyword",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			err := cmd.ValidateAddCommandArgs(tt.args)

			// Assert
			if tt.expectError {
				if err == nil {
					t.Errorf("ValidateAddCommandArgs(%v) expected error, got nil", tt.args)
					return
				}
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("ValidateAddCommandArgs(%v) error = %v, expected to contain %q", tt.args, err, tt.errorContains)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateAddCommandArgs(%v) unexpected error = %v", tt.args, err)
				}
			}
		})
	}
}

func TestToGoIdentifier_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "OnlyHyphens_ConvertsToUnderscores",
			input:    "---",
			expected: "___",
		},
		{
			name:     "SpecialUnicodeCharacters_PreservesAsIs",
			input:    "userα",
			expected: "Userα",
		},
		{
			name:     "NumbersOnly_CapitalizesFirstDigit",
			input:    "123",
			expected: "123",
		},
		{
			name:     "HyphenAtStart_ConvertsToUnderscore",
			input:    "-user",
			expected: "_user",
		},
		{
			name:     "HyphenAtEnd_ConvertsToUnderscore",
			input:    "user-",
			expected: "User_",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := cmd.ToGoIdentifier(tt.input)

			// Assert
			if result != tt.expected {
				t.Errorf("ToGoIdentifier(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}
