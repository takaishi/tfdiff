package tfdiff

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestApp_Run(t *testing.T) {
	tests := []struct {
		name        string
		cli         CLI
		setupDirs   func(t *testing.T) (string, string)
		wantErr     bool
		errContains string
	}{
		{
			name: "successful comparison with text output",
			cli: CLI{
				Levels:       []string{"resources"},
				IgnoreArgs:   true,
				OutputFormat: "text",
			},
			setupDirs: func(t *testing.T) (string, string) {
				tmpDir := t.TempDir()
				leftDir := filepath.Join(tmpDir, "left")
				rightDir := filepath.Join(tmpDir, "right")
				
				createTestModuleWithResource(t, leftDir, "aws_instance", "web")
				createTestModuleWithResource(t, rightDir, "aws_instance", "web")
				
				return leftDir, rightDir
			},
			wantErr: false,
		},
		{
			name: "successful comparison with json output",
			cli: CLI{
				Levels:       []string{"resources"},
				IgnoreArgs:   true,
				OutputFormat: "json",
			},
			setupDirs: func(t *testing.T) (string, string) {
				tmpDir := t.TempDir()
				leftDir := filepath.Join(tmpDir, "left")
				rightDir := filepath.Join(tmpDir, "right")
				
				createTestModuleWithResource(t, leftDir, "aws_instance", "web")
				createTestModuleWithResource(t, rightDir, "aws_instance", "web")
				
				return leftDir, rightDir
			},
			wantErr: false,
		},
		{
			name: "left directory does not exist",
			cli: CLI{
				LeftDir:      "/nonexistent/left",
				RightDir:     "/tmp",
				Levels:       []string{"resources"},
				OutputFormat: "text",
			},
			wantErr:     true,
			errContains: "left directory validation failed",
		},
		{
			name: "right directory does not exist",
			setupDirs: func(t *testing.T) (string, string) {
				tmpDir := t.TempDir()
				leftDir := filepath.Join(tmpDir, "left")
				createTestModule(t, leftDir)
				return leftDir, "/nonexistent/right"
			},
			cli: CLI{
				Levels:       []string{"resources"},
				IgnoreArgs:   true,
				OutputFormat: "text",
			},
			wantErr:     true,
			errContains: "right directory validation failed",
		},
		{
			name: "invalid output format",
			cli: CLI{
				Levels:       []string{"resources"},
				IgnoreArgs:   true,
				OutputFormat: "invalid",
			},
			setupDirs: func(t *testing.T) (string, string) {
				tmpDir := t.TempDir()
				leftDir := filepath.Join(tmpDir, "left")
				rightDir := filepath.Join(tmpDir, "right")
				
				createTestModule(t, leftDir)
				createTestModule(t, rightDir)
				
				return leftDir, rightDir
			},
			wantErr:     true,
			errContains: "unsupported output format",
		},
		{
			name: "comparison with different resources",
			cli: CLI{
				Levels:       []string{"resources"},
				IgnoreArgs:   true,
				OutputFormat: "text",
			},
			setupDirs: func(t *testing.T) (string, string) {
				tmpDir := t.TempDir()
				leftDir := filepath.Join(tmpDir, "left")
				rightDir := filepath.Join(tmpDir, "right")
				
				createTestModuleWithResource(t, leftDir, "aws_instance", "web")
				createTestModuleWithResource(t, rightDir, "aws_s3_bucket", "bucket")
				
				return leftDir, rightDir
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup directories if needed
			if tt.setupDirs != nil {
				leftDir, rightDir := tt.setupDirs(t)
				tt.cli.LeftDir = leftDir
				tt.cli.RightDir = rightDir
			}

			app := New(&tt.cli)
			err := app.Run(context.Background())

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errContains != "" && !containsString(err.Error(), tt.errContains) {
					t.Errorf("expected error to contain %q, got %q", tt.errContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestApp_OutputJSON(t *testing.T) {
	app := &App{
		CLI: &CLI{},
	}

	result := &ComparisonResult{
		Diffs: []Diff{
			{
				Type:    DiffTypeAdded,
				Level:   "resources",
				Element: "aws_instance.test",
				Message: "Resource added",
			},
		},
	}

	err := app.outputJSON(result)
	if err != nil {
		t.Errorf("unexpected error in outputJSON: %v", err)
	}
}

func TestApp_OutputText(t *testing.T) {
	app := &App{
		CLI: &CLI{},
	}

	result := &ComparisonResult{
		Diffs: []Diff{
			{
				Type:    DiffTypeAdded,
				Level:   "resources",
				Element: "aws_instance.test",
				Message: "Resource added",
			},
		},
	}

	err := app.outputText(result)
	if err != nil {
		t.Errorf("unexpected error in outputText: %v", err)
	}
}

// Helper functions

func createTestModuleWithResource(t *testing.T, dir, resourceType, resourceName string) {
	t.Helper()
	
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("failed to create test directory %s: %v", dir, err)
	}
	
	mainTf := filepath.Join(dir, "main.tf")
	content := `
resource "` + resourceType + `" "` + resourceName + `" {
  # Test resource configuration
}

output "test_output" {
  value = "test"
}
`
	
	if err := os.WriteFile(mainTf, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create main.tf in %s: %v", dir, err)
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
		findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}