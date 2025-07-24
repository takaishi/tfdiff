package tfdiff

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestCLI_ParseArguments(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantErr  bool
		expected CLI
	}{
		{
			name: "basic arguments",
			args: []string{"/path/left", "/path/right"},
			expected: CLI{
				LeftDir:      "/path/left",
				RightDir:     "/path/right",
				Levels:       []string{"module_calls", "outputs", "resources", "data_sources"},
				IgnoreArgs:   true,
				OutputFormat: "text",
			},
		},
		{
			name: "with custom levels",
			args: []string{"/path/left", "/path/right", "--level", "all"},
			expected: CLI{
				LeftDir:      "/path/left",
				RightDir:     "/path/right",
				Levels:       []string{"all"},
				IgnoreArgs:   true,
				OutputFormat: "text",
			},
		},
		{
			name: "with JSON output",
			args: []string{"/path/left", "/path/right", "--output", "json"},
			expected: CLI{
				LeftDir:      "/path/left",
				RightDir:     "/path/right",
				Levels:       []string{"module_calls", "outputs", "resources", "data_sources"},
				IgnoreArgs:   true,
				OutputFormat: "json",
			},
		},
		{
			name: "with ignore-args disabled",
			args: []string{"/path/left", "/path/right", "--ignore-args=false"},
			expected: CLI{
				LeftDir:      "/path/left",
				RightDir:     "/path/right",
				Levels:       []string{"module_calls", "outputs", "resources", "data_sources"},
				IgnoreArgs:   false,
				OutputFormat: "text",
			},
		},
		{
			name:    "missing arguments",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "missing right argument",
			args:    []string{"/path/left"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directories for valid test cases
			if !tt.wantErr {
				tmpDir := t.TempDir()
				leftDir := filepath.Join(tmpDir, "left")
				rightDir := filepath.Join(tmpDir, "right")
				
				// Create directories with minimal Terraform files
				createTestModule(t, leftDir)
				createTestModule(t, rightDir)
				
				// Replace test paths with actual temporary directories
				for i, arg := range tt.args {
					if arg == "/path/left" {
						tt.args[i] = leftDir
						tt.expected.LeftDir = leftDir
					}
					if arg == "/path/right" {
						tt.args[i] = rightDir
						tt.expected.RightDir = rightDir
					}
				}
			}

			err := RunCLI(context.Background(), tt.args)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestParseComparisonLevels(t *testing.T) {
	tests := []struct {
		name     string
		levels   []string
		expected []ComparisonLevel
	}{
		{
			name:   "single level",
			levels: []string{"outputs"},
			expected: []ComparisonLevel{ComparisonLevelOutputs},
		},
		{
			name:   "multiple levels",
			levels: []string{"resources", "variables", "outputs"},
			expected: []ComparisonLevel{
				ComparisonLevelResources,
				ComparisonLevelVariables,
				ComparisonLevelOutputs,
			},
		},
		{
			name:   "all level",
			levels: []string{"all"},
			expected: []ComparisonLevel{ComparisonLevelAll},
		},
		{
			name:   "with whitespace",
			levels: []string{" module_calls ", "  outputs  "},
			expected: []ComparisonLevel{
				ComparisonLevelModuleCalls,
				ComparisonLevelOutputs,
			},
		},
		{
			name:   "unknown level ignored",
			levels: []string{"unknown", "outputs"},
			expected: []ComparisonLevel{ComparisonLevelOutputs},
		},
		{
			name:     "empty levels",
			levels:   []string{},
			expected: []ComparisonLevel{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseComparisonLevels(tt.levels)
			
			if len(result) != len(tt.expected) {
				t.Errorf("expected %d levels, got %d", len(tt.expected), len(result))
				return
			}
			
			for i, expected := range tt.expected {
				if result[i] != expected {
					t.Errorf("expected level %v at index %d, got %v", expected, i, result[i])
				}
			}
		})
	}
}

// Helper function to create a minimal test Terraform module
func createTestModule(t *testing.T, dir string) {
	t.Helper()
	
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("failed to create test directory %s: %v", dir, err)
	}
	
	// Create a minimal main.tf file
	mainTf := filepath.Join(dir, "main.tf")
	content := `
resource "null_resource" "test" {
  triggers = {
    timestamp = timestamp()
  }
}

output "test_output" {
  value = "test"
}
`
	
	if err := os.WriteFile(mainTf, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create main.tf in %s: %v", dir, err)
	}
}