package layouts

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DefaultLayoutManager implements the LayoutManager interface
type DefaultLayoutManager struct {
	layouts       map[string]Layout
	defaultLayout string
}

// NewDefaultLayoutManager creates a new layout manager with built-in layouts
func NewDefaultLayoutManager() LayoutManager {
	manager := &DefaultLayoutManager{
		layouts:       make(map[string]Layout),
		defaultLayout: "standard",
	}

	// Register built-in layouts
	manager.registerBuiltinLayouts()

	return manager
}

// GetLayout retrieves a layout by name
func (m *DefaultLayoutManager) GetLayout(name string) (Layout, error) {
	layout, exists := m.layouts[name]
	if !exists {
		return nil, NewLayoutError("not_found", name,
			fmt.Errorf("layout '%s' not found", name))
	}

	return layout, nil
}

// ListLayouts returns all available layouts
func (m *DefaultLayoutManager) ListLayouts() []Layout {
	layouts := make([]Layout, 0, len(m.layouts))

	for _, layout := range m.layouts {
		layouts = append(layouts, layout)
	}

	return layouts
}

// DetectLayout attempts to detect the layout of an existing project
func (m *DefaultLayoutManager) DetectLayout(projectPath string) (Layout, error) {
	var bestMatch Layout
	var bestScore int

	for _, layout := range m.layouts {
		score := m.calculateLayoutScore(layout, projectPath)
		if score > bestScore {
			bestScore = score
			bestMatch = layout
		}
	}

	if bestMatch == nil || bestScore < 50 { // Minimum confidence threshold
		return m.GetDefaultLayout(), NewLayoutError("detection", "",
			fmt.Errorf("could not detect layout with sufficient confidence"))
	}

	return bestMatch, nil
}

// ValidateLayoutName checks if a layout name is valid
func (m *DefaultLayoutManager) ValidateLayoutName(name string) error {
	if name == "" {
		return NewLayoutError("validation", name,
			fmt.Errorf("layout name cannot be empty"))
	}

	// Check if layout exists
	if _, exists := m.layouts[name]; !exists {
		availableLayouts := make([]string, 0, len(m.layouts))
		for layoutName := range m.layouts {
			availableLayouts = append(availableLayouts, layoutName)
		}

		return NewLayoutError("validation", name,
			fmt.Errorf("layout '%s' not found. Available layouts: %s",
				name, strings.Join(availableLayouts, ", ")))
	}

	return nil
}

// GetDefaultLayout returns the default layout
func (m *DefaultLayoutManager) GetDefaultLayout() Layout {
	layout, _ := m.GetLayout(m.defaultLayout)
	return layout
}

// RegisterLayout registers a new layout
func (m *DefaultLayoutManager) RegisterLayout(layout Layout) error {
	if layout == nil {
		return NewLayoutError("validation", "",
			fmt.Errorf("layout cannot be nil"))
	}

	metadata := layout.Metadata()
	if metadata.Name == "" {
		return NewLayoutError("validation", "",
			fmt.Errorf("layout name cannot be empty"))
	}

	m.layouts[metadata.Name] = layout
	return nil
}

// UnregisterLayout removes a layout
func (m *DefaultLayoutManager) UnregisterLayout(name string) error {
	if name == m.defaultLayout {
		return NewLayoutError("validation", name,
			fmt.Errorf("cannot unregister default layout"))
	}

	if _, exists := m.layouts[name]; !exists {
		return NewLayoutError("not_found", name,
			fmt.Errorf("layout '%s' not found", name))
	}

	delete(m.layouts, name)
	return nil
}

// SetDefaultLayout sets the default layout
func (m *DefaultLayoutManager) SetDefaultLayout(name string) error {
	if err := m.ValidateLayoutName(name); err != nil {
		return err
	}

	m.defaultLayout = name
	return nil
}

// GetLayoutNames returns all layout names
func (m *DefaultLayoutManager) GetLayoutNames() []string {
	names := make([]string, 0, len(m.layouts))
	for name := range m.layouts {
		names = append(names, name)
	}
	return names
}

// registerBuiltinLayouts registers all built-in layouts
func (m *DefaultLayoutManager) registerBuiltinLayouts() {
	// Register standard layout
	m.RegisterLayout(NewStandardLayout())

	// TODO: Register other layouts as they are implemented
	// m.RegisterLayout(NewSinglePackageLayout())
	// m.RegisterLayout(NewFlatStructureLayout())
	// m.RegisterLayout(NewWebAppLayout())
	// m.RegisterLayout(NewMicroserviceLayout())
	// m.RegisterLayout(NewDDDLayout())
	// m.RegisterLayout(NewCleanLayout())
	// m.RegisterLayout(NewHexagonalLayout())
}

// calculateLayoutScore calculates how well a layout matches a project structure
func (m *DefaultLayoutManager) calculateLayoutScore(layout Layout, projectPath string) int {
	score := 0

	// Check if layout has detection rules
	if detectableLayout, ok := layout.(DetectableLayout); ok {
		rules := detectableLayout.GetDetectionRules()
		for _, rule := range rules {
			ruleScore := m.evaluateDetectionRule(rule, projectPath)
			if ruleScore > score {
				score = ruleScore
			}
		}
	} else {
		// Fallback: use basic validation
		if err := layout.ValidateProjectStructure(projectPath); err == nil {
			score = 70 // Medium confidence
		}
	}

	return score
}

// evaluateDetectionRule evaluates a single detection rule against a project
func (m *DefaultLayoutManager) evaluateDetectionRule(rule LayoutDetectionRule, projectPath string) int {
	// Check required directories
	for _, dir := range rule.RequiredDirs {
		dirPath := filepath.Join(projectPath, dir)
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			return 0 // Required directory missing
		}
	}

	// Check required files
	for _, file := range rule.RequiredFiles {
		filePath := filepath.Join(projectPath, file)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return 0 // Required file missing
		}
	}

	// Check prohibited paths
	for _, path := range rule.ProhibitedPaths {
		prohibitedPath := filepath.Join(projectPath, path)
		if _, err := os.Stat(prohibitedPath); err == nil {
			return 0 // Prohibited path exists
		}
	}

	// All checks passed, return confidence score
	return rule.Confidence
}

// DetectableLayout interface for layouts that can be auto-detected
type DetectableLayout interface {
	Layout
	GetDetectionRules() []LayoutDetectionRule
}

// LayoutManagerConfig represents configuration for the layout manager
type LayoutManagerConfig struct {
	DefaultLayout string            `yaml:"default_layout"`
	CustomLayouts map[string]string `yaml:"custom_layouts,omitempty"`
}

// DefaultLayoutManagerConfig returns default configuration
func DefaultLayoutManagerConfig() LayoutManagerConfig {
	return LayoutManagerConfig{
		DefaultLayout: "standard",
		CustomLayouts: make(map[string]string),
	}
}

// GetLayoutByComponentPath attempts to determine layout from component path
func (m *DefaultLayoutManager) GetLayoutByComponentPath(componentPath string) (Layout, ComponentType, error) {
	for _, layout := range m.layouts {
		// Check each component type to see if the path matches
		componentTypes := []ComponentType{
			ComponentHandler,
			ComponentModel,
			ComponentMiddleware,
			ComponentService,
			ComponentRepository,
		}

		for _, componentType := range componentTypes {
			if layout.IsComponentSupported(componentType) {
				placement, err := layout.GetComponentPlacement(componentType)
				if err != nil {
					continue
				}

				// Check if the path starts with the component directory
				if strings.HasPrefix(componentPath, placement.Directory) {
					return layout, componentType, nil
				}
			}
		}
	}

	return nil, "", NewLayoutError("detection", "",
		fmt.Errorf("could not determine layout from component path: %s", componentPath))
}

// CreateLayoutConfig creates a layout configuration for a project
func (m *DefaultLayoutManager) CreateLayoutConfig(layout Layout) *LayoutConfig {
	metadata := layout.Metadata()

	return &LayoutConfig{
		Layout: metadata.Name,
		Config: map[string]interface{}{
			"description": metadata.Description,
			"use_case":    metadata.UseCase,
			"team_size":   metadata.TeamSize,
			"complexity":  metadata.Complexity,
		},
	}
}

// ValidateLayoutConfig validates a layout configuration
func (m *DefaultLayoutManager) ValidateLayoutConfig(config *LayoutConfig) error {
	if config.Layout == "" {
		return NewLayoutError("validation", "",
			fmt.Errorf("layout name is required in configuration"))
	}

	return m.ValidateLayoutName(config.Layout)
}
