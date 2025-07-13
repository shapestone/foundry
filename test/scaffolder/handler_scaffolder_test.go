// test/scaffolder/handler_scaffolder_test.go
package scaffolder_test

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/shapestone/foundry/internal/scaffolder"
)

// Mock implementations for testing

type mockFileSystem struct {
	files       map[string][]byte
	directories map[string]bool
	existsFunc  func(path string) bool
}

func newMockFileSystem() *mockFileSystem {
	return &mockFileSystem{
		files:       make(map[string][]byte),
		directories: make(map[string]bool),
	}
}

func (m *mockFileSystem) Exists(path string) bool {
	if m.existsFunc != nil {
		return m.existsFunc(path)
	}
	_, exists := m.files[path]
	return exists || m.directories[path]
}

func (m *mockFileSystem) ReadFile(path string) ([]byte, error) {
	content, exists := m.files[path]
	if !exists {
		return nil, &mockError{"file not found: " + path}
	}
	return content, nil
}

func (m *mockFileSystem) WriteFile(path string, data []byte, perm uint32) error {
	m.files[path] = data
	return nil
}

func (m *mockFileSystem) MkdirAll(path string, perm uint32) error {
	m.directories[path] = true
	return nil
}

func (m *mockFileSystem) Remove(path string) error {
	delete(m.files, path)
	delete(m.directories, path)
	return nil
}

func (m *mockFileSystem) Stat(path string) (scaffolder.FileInfo, error) {
	return &mockFileInfo{name: path}, nil
}

type mockFileInfo struct {
	name string
}

func (m *mockFileInfo) Name() string { return m.name }
func (m *mockFileInfo) Size() int64  { return 0 }
func (m *mockFileInfo) Mode() uint32 { return 0644 }
func (m *mockFileInfo) IsDir() bool  { return false }

type mockTemplate struct {
	name    string
	content string
}

func (m *mockTemplate) Name() string { return m.name }

func (m *mockTemplate) Execute(wr io.Writer, data interface{}) error {
	// Simple template rendering for testing
	content := m.content
	if dataMap, ok := data.(map[string]interface{}); ok {
		for key, value := range dataMap {
			placeholder := "{{." + key + "}}"
			if strValue, ok := value.(string); ok {
				content = strings.ReplaceAll(content, placeholder, strValue)
			}
		}
	}
	_, err := wr.Write([]byte(content))
	return err
}

type mockTemplateRenderer struct {
	templates map[string]*mockTemplate
}

func newMockTemplateRenderer() *mockTemplateRenderer {
	return &mockTemplateRenderer{
		templates: make(map[string]*mockTemplate),
	}
}

func (m *mockTemplateRenderer) LoadTemplate(name string) (scaffolder.Template, error) {
	template, exists := m.templates[name]
	if !exists {
		return nil, &mockError{"template not found: " + name}
	}
	return template, nil
}

func (m *mockTemplateRenderer) RenderTemplate(tmpl scaffolder.Template, data interface{}) (string, error) {
	var buf strings.Builder
	err := tmpl.Execute(&buf, data)
	return buf.String(), err
}

func (m *mockTemplateRenderer) addTemplate(name, content string) {
	m.templates[name] = &mockTemplate{name: name, content: content}
}

type mockProjectAnalyzer struct {
	moduleName     string
	projectName    string
	isGoProjectVal bool
}

func (m *mockProjectAnalyzer) GetModuleName(projectRoot string) (string, error) {
	if m.moduleName == "" {
		return "", &mockError{"module name not found"}
	}
	return m.moduleName, nil
}

func (m *mockProjectAnalyzer) GetProjectName(projectRoot string) (string, error) {
	if m.projectName == "" {
		return "", &mockError{"project name not found"}
	}
	return m.projectName, nil
}

func (m *mockProjectAnalyzer) ValidateProjectStructure(projectRoot string) error {
	return nil
}

func (m *mockProjectAnalyzer) IsGoProject(projectRoot string) bool {
	return m.isGoProjectVal
}

type mockUserInteraction struct {
	confirmResult     bool
	showPreviewResult bool
}

func (m *mockUserInteraction) Confirm(message string) bool {
	return m.confirmResult
}

func (m *mockUserInteraction) ShowPreview(title string, changes []string, message string) bool {
	return m.showPreviewResult
}

type mockError struct {
	message string
}

func (e *mockError) Error() string {
	return e.message
}

// Test functions

func TestScaffolder_CreateHandler_Success(t *testing.T) {
	// Given
	mockFS := newMockFileSystem()
	mockTemplateRenderer := newMockTemplateRenderer()
	mockProjectAnalyzer := &mockProjectAnalyzer{
		moduleName:     "github.com/example/myapp",
		isGoProjectVal: true,
	}
	mockUserInteraction := &mockUserInteraction{}

	// Set up template
	mockTemplateRenderer.addTemplate("handler.go.tmpl", `package handlers

import "net/http"

type {{.HandlerName}}Handler struct{}

func New{{.HandlerName}}Handler() *{{.HandlerName}}Handler {
	return &{{.HandlerName}}Handler{}
}

func (h *{{.HandlerName}}Handler) Routes() http.Handler {
	// Handler implementation
	return nil
}
`)

	scaffolderInstance := scaffolder.New(
		mockFS,
		mockTemplateRenderer,
		mockProjectAnalyzer,
		mockUserInteraction,
	)

	spec := &scaffolder.HandlerSpec{
		Name:        "user",
		Type:        "REST",
		AutoWire:    false,
		DryRun:      false,
		ProjectRoot: "/test/project",
		Module:      "github.com/example/myapp",
	}

	// When
	result, err := scaffolderInstance.CreateHandler(context.Background(), spec)

	// Then
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !result.Success {
		t.Fatalf("Expected success, got failure: %s", result.Message)
	}

	if len(result.FilesCreated) != 1 {
		t.Fatalf("Expected 1 file created, got %d", len(result.FilesCreated))
	}

	expectedPath := "/test/project/internal/handlers/user.go"
	if result.FilesCreated[0] != expectedPath {
		t.Fatalf("Expected file %s, got %s", expectedPath, result.FilesCreated[0])
	}

	// Check file content
	content, exists := mockFS.files[expectedPath]
	if !exists {
		t.Fatalf("File was not created: %s", expectedPath)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "UserHandler") {
		t.Fatalf("Generated content should contain 'UserHandler', got: %s", contentStr)
	}
}

func TestScaffolder_CreateHandler_ValidationErrors(t *testing.T) {
	scaffolderInstance := scaffolder.New(nil, nil, nil, nil)

	tests := []struct {
		name        string
		spec        *scaffolder.HandlerSpec
		expectedErr string
	}{
		{
			name: "empty name",
			spec: &scaffolder.HandlerSpec{
				Name:        "",
				ProjectRoot: "/test",
			},
			expectedErr: "handler name is required",
		},
		{
			name: "name too short",
			spec: &scaffolder.HandlerSpec{
				Name:        "a",
				ProjectRoot: "/test",
			},
			expectedErr: "handler name must be at least 2 characters",
		},
		{
			name: "name too long",
			spec: &scaffolder.HandlerSpec{
				Name:        strings.Repeat("a", 51),
				ProjectRoot: "/test",
			},
			expectedErr: "handler name must be less than 50 characters",
		},
		{
			name: "empty project root",
			spec: &scaffolder.HandlerSpec{
				Name:        "user",
				ProjectRoot: "",
			},
			expectedErr: "project root is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// When
			_, err := scaffolderInstance.CreateHandler(context.Background(), tt.spec)

			// Then
			if err == nil {
				t.Fatalf("Expected error, got nil")
			}

			if !strings.Contains(err.Error(), tt.expectedErr) {
				t.Fatalf("Expected error containing '%s', got '%s'", tt.expectedErr, err.Error())
			}
		})
	}
}

func TestScaffolder_CreateHandler_NotGoProject(t *testing.T) {
	// Given
	mockProjectAnalyzer := &mockProjectAnalyzer{
		isGoProjectVal: false,
	}

	scaffolderInstance := scaffolder.New(
		nil,
		nil,
		mockProjectAnalyzer,
		nil,
	)

	spec := &scaffolder.HandlerSpec{
		Name:        "user",
		ProjectRoot: "/test/project",
	}

	// When
	_, err := scaffolderInstance.CreateHandler(context.Background(), spec)

	// Then
	if err == nil {
		t.Fatalf("Expected error for non-Go project")
	}

	if !strings.Contains(err.Error(), "not a Go project") {
		t.Fatalf("Expected 'not a Go project' error, got: %s", err.Error())
	}
}

func TestScaffolder_CreateHandler_HandlerExists(t *testing.T) {
	// Given
	mockFS := newMockFileSystem()
	mockFS.existsFunc = func(path string) bool {
		return strings.Contains(path, "user.go") // Handler already exists
	}

	mockProjectAnalyzer := &mockProjectAnalyzer{
		moduleName:     "github.com/example/myapp",
		isGoProjectVal: true,
	}

	scaffolderInstance := scaffolder.New(
		mockFS,
		nil,
		mockProjectAnalyzer,
		nil,
	)

	spec := &scaffolder.HandlerSpec{
		Name:        "user",
		ProjectRoot: "/test/project",
		Module:      "github.com/example/myapp",
	}

	// When
	_, err := scaffolderInstance.CreateHandler(context.Background(), spec)

	// Then
	if err == nil {
		t.Fatalf("Expected error for existing handler")
	}

	if !strings.Contains(err.Error(), "handler already exists") {
		t.Fatalf("Expected 'handler already exists' error, got: %s", err.Error())
	}
}

func TestScaffolder_CreateHandler_DryRun(t *testing.T) {
	// Given
	mockFS := newMockFileSystem()
	mockTemplateRenderer := newMockTemplateRenderer()
	mockProjectAnalyzer := &mockProjectAnalyzer{
		moduleName:     "github.com/example/myapp",
		isGoProjectVal: true,
	}

	// Set up template (this was missing!)
	mockTemplateRenderer.addTemplate("handler.go.tmpl", `package handlers

import "net/http"

type {{.HandlerName}}Handler struct{}

func New{{.HandlerName}}Handler() *{{.HandlerName}}Handler {
	return &{{.HandlerName}}Handler{}
}

func (h *{{.HandlerName}}Handler) Routes() http.Handler {
	// Handler implementation
	return nil
}
`)

	scaffolderInstance := scaffolder.New(
		mockFS,
		mockTemplateRenderer,
		mockProjectAnalyzer,
		nil,
	)

	spec := &scaffolder.HandlerSpec{
		Name:        "user",
		ProjectRoot: "/test/project",
		Module:      "github.com/example/myapp",
		DryRun:      true,
	}

	// When
	result, err := scaffolderInstance.CreateHandler(context.Background(), spec)

	// Then
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !result.Success {
		t.Fatalf("Expected success, got failure")
	}

	if len(result.FilesCreated) != 0 {
		t.Fatalf("Expected no files created in dry run, got %d", len(result.FilesCreated))
	}

	if len(result.Changes) == 0 {
		t.Fatalf("Expected changes to be shown in dry run")
	}

	if !strings.Contains(result.Message, "Dry run completed") {
		t.Fatalf("Expected dry run message, got: %s", result.Message)
	}
}

// Benchmark tests
func BenchmarkScaffolder_CreateHandler(b *testing.B) {
	// Setup
	mockFS := newMockFileSystem()
	mockTemplateRenderer := newMockTemplateRenderer()
	mockTemplateRenderer.addTemplate("handler.go.tmpl", "package handlers\n\ntype {{.HandlerName}}Handler struct{}")

	mockProjectAnalyzer := &mockProjectAnalyzer{
		moduleName:     "github.com/example/myapp",
		isGoProjectVal: true,
	}

	scaffolderInstance := scaffolder.New(
		mockFS,
		mockTemplateRenderer,
		mockProjectAnalyzer,
		nil,
	)

	spec := &scaffolder.HandlerSpec{
		Name:        "user",
		Type:        "REST",
		ProjectRoot: "/test/project",
		Module:      "github.com/example/myapp",
		DryRun:      true, // Use dry run for benchmarking
	}

	b.ResetTimer()

	// Benchmark
	for i := 0; i < b.N; i++ {
		_, err := scaffolderInstance.CreateHandler(context.Background(), spec)
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}
