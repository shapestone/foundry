package integration_test

import (
	"testing"

	"github.com/shapestone/foundry/test/testutil"
)

func TestNewCommand(t *testing.T) {
	suite := testutil.NewTestSuite(t)

	t.Run("CreateBasicProject", func(t *testing.T) {
		// Given
		projectName := "testapp"

		// When
		output, err := suite.RunFoundry("new", projectName)

		// Then
		suite.AssertNoError(err)
		suite.AssertCommandOutput(output, "✅ Project created successfully!")
		suite.AssertCommandOutput(output, projectName)

		// Verify project structure
		expectedFiles := []string{
			"main.go",
			"go.mod",
			"README.md",
			".gitignore",
			"internal/handlers",
			"internal/middleware",
			"internal/routes/routes.go",
		}
		suite.AssertProjectStructure(projectName, expectedFiles)

		// Verify go.mod content
		suite.AssertFileContains(projectName, "go.mod", "module "+projectName)
		suite.AssertFileContains(projectName, "go.mod", "go 1.21")

		// Verify generated code compiles
		suite.AssertGeneratedCodeCompiles(projectName)
	})

	t.Run("ProjectNameWithHyphens", func(t *testing.T) {
		// Given
		projectName := "my-test-app"

		// When
		output, err := suite.RunFoundry("new", projectName)

		// Then
		suite.AssertNoError(err)
		suite.AssertCommandOutput(output, "✅ Project created successfully!")
		suite.AssertProjectStructure(projectName, []string{"main.go", "go.mod"})
	})

	t.Run("ProjectAlreadyExists", func(t *testing.T) {
		// Given
		projectName := "existingapp"
		suite.CreateTestProject(projectName)

		// When
		_, err := suite.RunFoundry("new", projectName)

		// Then
		if err == nil {
			t.Error("Expected error when creating project that already exists")
		}
	})

	t.Run("ProjectNameWithSpaces", func(t *testing.T) {
		// Given
		projectName := "my app"

		// When
		_, err := suite.RunFoundry("new", projectName)

		// Then
		if err == nil {
			t.Error("Expected error for project name with spaces")
		}
	})
}

func TestNewCommandGeneratedCode(t *testing.T) {
	suite := testutil.NewTestSuite(t)

	t.Run("GeneratedProjectRuns", func(t *testing.T) {
		// Given
		projectName := "runnable"

		// When
		_, err := suite.RunFoundry("new", projectName)

		// Then
		suite.AssertNoError(err)

		// This will compile and briefly run the server
		suite.AssertGeneratedCodeRuns(projectName)
	})
}
