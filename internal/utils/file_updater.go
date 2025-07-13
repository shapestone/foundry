package utils

import (
	"fmt"
	"github.com/shapestone/foundry/internal/interactive"
	"github.com/shapestone/foundry/internal/project"
	"github.com/shapestone/foundry/internal/routes"
)

// FileUpdate represents a file modification
// Deprecated: Use routes.Update instead
type FileUpdate = routes.Update

// UpdateRoutesFile safely updates the routes file with preview and rollback
func UpdateRoutesFile(handlerName string, dryRun bool) error {
	updater := routes.NewFileUpdater()
	prompter := interactive.NewConsolePrompter()
	moduleName := project.GetCurrentModule()

	// Calculate changes
	update, err := updater.UpdateRoutes(handlerName, moduleName)
	if err != nil {
		return err
	}

	// Show preview
	message := fmt.Sprintf("This will add the %s handler to your routes", handlerName)
	if !prompter.ShowPreview(fmt.Sprintf("Preview changes to %s:", update.Path), update.Changes, message) {
		return fmt.Errorf("update cancelled by user")
	}

	if dryRun {
		fmt.Println("🔍 Dry run complete - no changes made")
		return nil
	}

	// Apply with rollback protection
	return routes.ApplyUpdate(update, updater)
}

// calculateRoutesUpdate calculates the routes update
// Deprecated: Use routes.FileUpdater.UpdateRoutes instead
func calculateRoutesUpdate(content, handlerName string) *FileUpdate {
	// This function is now internal to the routes package
	// Keeping for backward compatibility if needed
	panic("calculateRoutesUpdate is deprecated, use routes.FileUpdater.UpdateRoutes")
}

// showUpdatePreview shows update preview
// Deprecated: Use interactive.ConsolePrompter.ShowPreview instead
func showUpdatePreview(update *FileUpdate, handlerName string) bool {
	prompter := interactive.NewConsolePrompter()
	message := fmt.Sprintf("This will add the %s handler to your routes", handlerName)
	return prompter.ShowPreview(fmt.Sprintf("Preview changes to %s:", update.Path), update.Changes, message)
}

// applyFileUpdate applies file update
// Deprecated: Use routes.ApplyUpdate instead
func applyFileUpdate(update *FileUpdate) error {
	updater := routes.NewFileUpdater()
	return routes.ApplyUpdate(update, updater)
}

// validateGoFile validates Go file syntax
// Deprecated: Use routes.FileUpdater.ValidateGoFile instead
func validateGoFile(path string) error {
	updater := routes.NewFileUpdater()
	return updater.ValidateGoFile(path)
}
