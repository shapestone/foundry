// internal/scaffolder/adapters.go
package scaffolder

import (
	"io"
	"os"
	"path/filepath"

	"github.com/shapestone/foundry/internal/generator"
	"github.com/shapestone/foundry/internal/interactive"
	"github.com/shapestone/foundry/internal/project"
)

// Adapters to integrate existing components with the new scaffolder interfaces

// FileSystemAdapter adapts the standard os package to our FileSystem interface
type FileSystemAdapter struct{}

func NewFileSystemAdapter() *FileSystemAdapter {
	return &FileSystemAdapter{}
}

func (f *FileSystemAdapter) Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (f *FileSystemAdapter) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func (f *FileSystemAdapter) WriteFile(path string, data []byte, perm uint32) error {
	return os.WriteFile(path, data, os.FileMode(perm))
}

func (f *FileSystemAdapter) MkdirAll(path string, perm uint32) error {
	return os.MkdirAll(path, os.FileMode(perm))
}

func (f *FileSystemAdapter) Remove(path string) error {
	return os.Remove(path)
}

func (f *FileSystemAdapter) Stat(path string) (FileInfo, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	return &FileInfoAdapter{info}, nil
}

// FileInfoAdapter adapts os.FileInfo to our FileInfo interface
type FileInfoAdapter struct {
	info os.FileInfo
}

func (f *FileInfoAdapter) Name() string { return f.info.Name() }
func (f *FileInfoAdapter) Size() int64  { return f.info.Size() }
func (f *FileInfoAdapter) Mode() uint32 { return uint32(f.info.Mode()) }
func (f *FileInfoAdapter) IsDir() bool  { return f.info.IsDir() }

// TemplateRendererAdapter adapts the existing generator to our TemplateRenderer interface
type TemplateRendererAdapter struct {
	generator *generator.FileGenerator
}

func NewTemplateRendererAdapter() *TemplateRendererAdapter {
	return &TemplateRendererAdapter{
		generator: generator.NewFileGenerator(),
	}
}

func (t *TemplateRendererAdapter) LoadTemplate(name string) (Template, error) {
	// For now, we'll use the embedded templates from the foundry package
	// This will be improved when we move templates to external files
	return &TemplateAdapter{name: name}, nil
}

func (t *TemplateRendererAdapter) RenderTemplate(tmpl Template, data interface{}) (string, error) {
	// This is a simplified implementation
	// In practice, we'll need to integrate with your existing template system
	return "// Generated handler placeholder\n", nil
}

// TemplateAdapter represents a template
type TemplateAdapter struct {
	name string
}

func (t *TemplateAdapter) Name() string {
	return t.name
}

func (t *TemplateAdapter) Execute(wr io.Writer, data interface{}) error {
	// Simplified implementation for now
	_, err := wr.Write([]byte("// Generated handler placeholder\n"))
	return err
}

// ProjectAnalyzerAdapter adapts the existing project package
type ProjectAnalyzerAdapter struct{}

func NewProjectAnalyzerAdapter() *ProjectAnalyzerAdapter {
	return &ProjectAnalyzerAdapter{}
}

func (p *ProjectAnalyzerAdapter) GetModuleName(projectRoot string) (string, error) {
	// Change to project root temporarily
	originalDir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(projectRoot); err != nil {
		return "", err
	}

	return project.GetCurrentModule(), nil
}

func (p *ProjectAnalyzerAdapter) GetProjectName(projectRoot string) (string, error) {
	// Change to project root temporarily
	originalDir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(projectRoot); err != nil {
		return "", err
	}

	return project.GetProjectName(), nil
}

func (p *ProjectAnalyzerAdapter) ValidateProjectStructure(projectRoot string) error {
	// Basic validation - check if go.mod exists
	if !p.IsGoProject(projectRoot) {
		return &ValidationError{
			Field:   "project_root",
			Message: "not a valid Go project (go.mod not found)",
		}
	}
	return nil
}

func (p *ProjectAnalyzerAdapter) IsGoProject(projectRoot string) bool {
	goModPath := filepath.Join(projectRoot, "go.mod")
	_, err := os.Stat(goModPath)
	return err == nil
}

// UserInteractionAdapter adapts the existing interactive package
type UserInteractionAdapter struct {
	prompter interactive.Prompter
}

func NewUserInteractionAdapter() *UserInteractionAdapter {
	return &UserInteractionAdapter{
		prompter: interactive.NewConsolePrompter(),
	}
}

func (u *UserInteractionAdapter) Confirm(message string) bool {
	return u.prompter.Confirm(message)
}

func (u *UserInteractionAdapter) ShowPreview(title string, changes []string, message string) bool {
	return u.prompter.ShowPreview(title, changes, message)
}

// Factory function to create a scaffolder with real dependencies
func NewScaffolderWithAdapters() Scaffolder {
	return New(
		NewFileSystemAdapter(),
		NewTemplateRendererAdapter(),
		NewProjectAnalyzerAdapter(),
		NewUserInteractionAdapter(),
	)
}

// Factory function to create a scaffolder for testing
func NewScaffolderForTesting(
	fileSystem FileSystem,
	templateRenderer TemplateRenderer,
	projectAnalyzer ProjectAnalyzer,
	userInteraction UserInteraction,
) Scaffolder {
	return New(fileSystem, templateRenderer, projectAnalyzer, userInteraction)
}
