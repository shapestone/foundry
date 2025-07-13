package layout

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode"
)

// String case conversion helpers

// toLower converts a string to lowercase
func toLower(s string) string {
	return strings.ToLower(s)
}

// toUpper converts a string to uppercase
func toUpper(s string) string {
	return strings.ToUpper(s)
}

// capitalize capitalizes the first letter of a string
func capitalize(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

// toSnakeCase converts a string to snake_case
func toSnakeCase(s string) string {
	var result []rune
	for i, r := range s {
		if unicode.IsUpper(r) && i > 0 {
			// Add underscore before uppercase letters (except first)
			if i > 0 && !unicode.IsUpper([]rune(s)[i-1]) {
				result = append(result, '_')
			}
		}
		result = append(result, unicode.ToLower(r))
	}
	return string(result)
}

// toCamelCase converts a string to camelCase
func toCamelCase(s string) string {
	parts := splitWords(s)
	if len(parts) == 0 {
		return ""
	}

	// First word lowercase
	result := strings.ToLower(parts[0])

	// Rest of words capitalized
	for i := 1; i < len(parts); i++ {
		result += capitalize(strings.ToLower(parts[i]))
	}

	return result
}

// toPascalCase converts a string to PascalCase
func toPascalCase(s string) string {
	parts := splitWords(s)
	result := ""

	for _, part := range parts {
		result += capitalize(strings.ToLower(part))
	}

	return result
}

// toKebabCase converts a string to kebab-case
func toKebabCase(s string) string {
	parts := splitWords(s)
	return strings.Join(parts, "-")
}

// splitWords splits a string into words
func splitWords(s string) []string {
	// Handle different separators
	s = strings.ReplaceAll(s, "-", " ")
	s = strings.ReplaceAll(s, "_", " ")

	// Handle camelCase and PascalCase
	var result []rune
	for i, r := range s {
		if unicode.IsUpper(r) && i > 0 {
			prev := []rune(s)[i-1]
			if !unicode.IsSpace(prev) && !unicode.IsUpper(prev) {
				result = append(result, ' ')
			}
		}
		result = append(result, r)
	}

	// Split on spaces and filter empty strings
	parts := strings.Fields(string(result))

	// Convert to lowercase
	for i := range parts {
		parts[i] = strings.ToLower(parts[i])
	}

	return parts
}

// Archive extraction helpers

// extractTarGz extracts a tar.gz archive to a destination directory
func extractTarGz(archivePath, destDir string) error {
	// Open the archive file
	file, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("failed to open archive: %w", err)
	}
	defer file.Close()

	// Create gzip reader
	gzr, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzr.Close()

	// Create tar reader
	tr := tar.NewReader(gzr)

	// Extract files
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar header: %w", err)
		}

		// Construct the file path
		target := filepath.Join(destDir, header.Name)

		// Check for directory traversal
		if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(destDir)) {
			return fmt.Errorf("invalid file path in archive: %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			// Create directory
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}

		case tar.TypeReg:
			// Create the directory if needed
			dir := filepath.Dir(target)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}

			// Create the file
			outFile, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("failed to create file: %w", err)
			}

			// Copy file contents
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return fmt.Errorf("failed to write file: %w", err)
			}
			outFile.Close()

		case tar.TypeSymlink:
			// Create symlink
			if err := os.Symlink(header.Linkname, target); err != nil {
				return fmt.Errorf("failed to create symlink: %w", err)
			}

		default:
			// Skip unknown types
			continue
		}
	}

	return nil
}

// Checksum helpers

// calculateSHA256 calculates the SHA256 checksum of a file
func calculateSHA256(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", fmt.Errorf("failed to calculate checksum: %w", err)
	}

	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

// verifySHA256 verifies a file's SHA256 checksum
func verifySHA256(filePath, expectedChecksum string) error {
	actualChecksum, err := calculateSHA256(filePath)
	if err != nil {
		return err
	}

	if actualChecksum != expectedChecksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum)
	}

	return nil
}

// Path helpers

// expandPath expands ~ and environment variables in a path
func expandPath(path string) string {
	// Expand tilde to home directory
	if strings.HasPrefix(path, "~/") {
		if home := getHomeDir(); home != "" {
			path = filepath.Join(home, path[2:])
		}
	}

	// Expand environment variables
	path = os.ExpandEnv(path)

	return path
}

// getHomeDir returns the user's home directory
func getHomeDir() string {
	// Try multiple methods to get home directory

	// Method 1: Use os.UserHomeDir (Go 1.12+)
	if home, err := os.UserHomeDir(); err == nil {
		return home
	}

	// Method 2: Check HOME environment variable
	if home := os.Getenv("HOME"); home != "" {
		return home
	}

	// Method 3: Windows specific
	if runtime := os.Getenv("USERPROFILE"); runtime != "" {
		return runtime
	}

	// Method 4: Construct from HOMEDRIVE and HOMEPATH (Windows)
	drive := os.Getenv("HOMEDRIVE")
	path := os.Getenv("HOMEPATH")
	if drive != "" && path != "" {
		return drive + path
	}

	return ""
}

// Git helpers

// cloneGitRepository clones a git repository
func cloneGitRepository(url, ref, destDir string) error {
	// Check if git is available
	if _, err := exec.LookPath("git"); err != nil {
		return fmt.Errorf("git is not installed or not in PATH")
	}

	// Clone the repository
	args := []string{"clone"}

	// Add branch/tag/ref if specified
	if ref != "" && ref != "main" && ref != "master" {
		args = append(args, "--branch", ref)
	}

	// Add depth to make clone faster
	args = append(args, "--depth", "1")

	// Add URL and destination
	args = append(args, url, destDir)

	// Execute git clone
	cmd := exec.Command("git", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git clone failed: %w\nOutput: %s", err, string(output))
	}

	// If we need to checkout a specific ref that wasn't a branch/tag
	if ref != "" && ref != "main" && ref != "master" {
		// Check if the ref is a commit hash
		if len(ref) == 40 { // SHA-1 hash length
			checkoutCmd := exec.Command("git", "-C", destDir, "checkout", ref)
			if output, err := checkoutCmd.CombinedOutput(); err != nil {
				return fmt.Errorf("git checkout failed: %w\nOutput: %s", err, string(output))
			}
		}
	}

	// Remove .git directory to save space
	gitDir := filepath.Join(destDir, ".git")
	if err := os.RemoveAll(gitDir); err != nil {
		// Non-fatal error
		fmt.Printf("Warning: failed to remove .git directory: %v\n", err)
	}

	return nil
}

// Validation helpers

// isValidLayoutName checks if a layout name is valid
func isValidLayoutName(name string) bool {
	if name == "" {
		return false
	}

	// Allow alphanumeric, hyphens, and underscores
	for _, r := range name {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '-' && r != '_' {
			return false
		}
	}

	return true
}

// isValidProjectName checks if a project name is valid
func isValidProjectName(name string) bool {
	return isValidLayoutName(name) // Same rules for now
}

// Template helpers

// mergeMaps merges multiple maps into one
func mergeMaps(maps ...map[string]string) map[string]string {
	result := make(map[string]string)

	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}

	return result
}

// defaultString returns the default value if the string is empty
func defaultString(s, defaultValue string) string {
	if s == "" {
		return defaultValue
	}
	return s
}

// contains checks if a string slice contains a value
func contains(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

// unique returns unique values from a string slice
func unique(slice []string) []string {
	seen := make(map[string]bool)
	result := []string{}

	for _, v := range slice {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}

	return result
}

// File helpers

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	// Create destination directory if needed
	destDir := filepath.Dir(dst)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	// Copy file permissions
	sourceInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to stat source file: %w", err)
	}

	return os.Chmod(dst, sourceInfo.Mode())
}

// ensureDir ensures a directory exists
func ensureDir(dir string) error {
	return os.MkdirAll(dir, 0755)
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// isDir checks if a path is a directory
func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// pluralize adds 's' to make a word plural (simple implementation)
func pluralize(s string) string {
	if s == "" {
		return s
	}

	// Simple pluralization rules
	switch {
	case strings.HasSuffix(s, "y"):
		return s[:len(s)-1] + "ies"
	case strings.HasSuffix(s, "s"), strings.HasSuffix(s, "sh"), strings.HasSuffix(s, "ch"), strings.HasSuffix(s, "x"), strings.HasSuffix(s, "z"):
		return s + "es"
	default:
		return s + "s"
	}
}
