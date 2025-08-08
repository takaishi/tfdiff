package tfdiff

import (
	"os"
	"path/filepath"
)

// ParseModule parses a Terraform module directory and extracts its definitions
func ParseModule(modulePath string) (*ModuleDefinition, error) {
	// Use the HCL parser implementation
	return ParseModuleHCL(modulePath)
}

// FindTerraformFiles finds all .tf files in the specified directory
func FindTerraformFiles(path string) ([]string, error) {
	files, err := filepath.Glob(filepath.Join(path, "*.tf"))
	if err != nil {
		return nil, err
	}
	return files, nil
}

// ValidateModuleDirectory validates that a directory exists and contains Terraform files
func ValidateModuleDirectory(path string) error {
	// Check if directory exists
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return err
	}
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return os.ErrNotExist
	}

	// Check if directory contains .tf files
	files, err := FindTerraformFiles(path)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return os.ErrNotExist
	}

	return nil
}