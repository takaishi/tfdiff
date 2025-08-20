package tfdiff

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestValidateModuleDirectory(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T) string
		wantErr bool
	}{
		{
			name: "valid directory with terraform files",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				createTestModule(t, tmpDir)
				return tmpDir
			},
			wantErr: false,
		},
		{
			name: "nonexistent directory",
			setup: func(t *testing.T) string {
				return "/nonexistent/directory"
			},
			wantErr: true,
		},
		{
			name: "directory without terraform files",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				// Create directory but no .tf files
				return tmpDir
			},
			wantErr: true,
		},
		{
			name: "file instead of directory",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				filePath := filepath.Join(tmpDir, "test.txt")
				if err := os.WriteFile(filePath, []byte("test"), 0644); err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}
				return filePath
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup(t)
			err := ValidateModuleDirectory(path)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestFindTerraformFiles(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T) string
		expectedMin int // minimum expected files
	}{
		{
			name: "directory with single terraform file",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				createTestModule(t, tmpDir)
				return tmpDir
			},
			expectedMin: 1,
		},
		{
			name: "directory with multiple terraform files",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				
				// Create multiple .tf files
				files := []string{"main.tf", "variables.tf", "outputs.tf"}
				for _, file := range files {
					path := filepath.Join(tmpDir, file)
					content := `# Test terraform file`
					if err := os.WriteFile(path, []byte(content), 0644); err != nil {
						t.Fatalf("failed to create %s: %v", file, err)
					}
				}
				
				return tmpDir
			},
			expectedMin: 3,
		},
		{
			name: "directory with no terraform files",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				// Create some non-terraform files
				if err := os.WriteFile(filepath.Join(tmpDir, "readme.md"), []byte("test"), 0644); err != nil {
					t.Fatalf("failed to create readme.md: %v", err)
				}
				return tmpDir
			},
			expectedMin: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tt.setup(t)
			files, err := FindTerraformFiles(dir)
			
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			
			if len(files) < tt.expectedMin {
				t.Errorf("expected at least %d files, got %d", tt.expectedMin, len(files))
			}
			
			// Verify all returned files have .tf extension
			for _, file := range files {
				if filepath.Ext(file) != ".tf" {
					t.Errorf("found non-terraform file in results: %s", file)
				}
			}
		})
	}
}

func TestParseModule(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T) string
		wantErr bool
		check   func(t *testing.T, module *ModuleDefinition)
	}{
		{
			name: "simple module with resource",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				mainTf := filepath.Join(tmpDir, "main.tf")
				content := `
resource "aws_instance" "web" {
  ami           = "ami-0c02fb55956c7d316"
  instance_type = "t2.micro"
}

output "instance_id" {
  value = aws_instance.web.id
}
`
				if err := os.WriteFile(mainTf, []byte(content), 0644); err != nil {
					t.Fatalf("failed to create main.tf: %v", err)
				}
				return tmpDir
			},
			wantErr: false,
			check: func(t *testing.T, module *ModuleDefinition) {
				if len(module.Resources) == 0 {
					t.Errorf("expected resources to be found")
				}
				if len(module.Outputs) == 0 {
					t.Errorf("expected outputs to be found")
				}
			},
		},
		{
			name: "module with variable",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				variablesTf := filepath.Join(tmpDir, "variables.tf")
				content := `
variable "instance_type" {
  description = "Type of EC2 instance"
  type        = string
  default     = "t2.micro"
}
`
				if err := os.WriteFile(variablesTf, []byte(content), 0644); err != nil {
					t.Fatalf("failed to create variables.tf: %v", err)
				}
				return tmpDir
			},
			wantErr: false,
			check: func(t *testing.T, module *ModuleDefinition) {
				if len(module.Variables) == 0 {
					t.Errorf("expected variables to be found")
				}
			},
		},
		{
			name: "nonexistent directory",
			setup: func(t *testing.T) string {
				return "/nonexistent/directory"
			},
			wantErr: true,
			check:   nil,
		},
		{
			name: "directory without terraform files",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				// Create a non-terraform file
				if err := os.WriteFile(filepath.Join(tmpDir, "readme.md"), []byte("test"), 0644); err != nil {
					t.Fatalf("failed to create readme.md: %v", err)
				}
				return tmpDir
			},
			wantErr: false, // ParseModule will succeed but return empty module
			check: func(t *testing.T, module *ModuleDefinition) {
				// Should have empty collections since no .tf files exist
				if len(module.Resources) != 0 || len(module.Variables) != 0 || len(module.Outputs) != 0 {
					t.Errorf("expected empty module collections")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tt.setup(t)
			module, err := ParseModule(dir)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			
			if module == nil {
				t.Errorf("expected module but got nil")
				return
			}
			
			if module.Path != dir {
				t.Errorf("expected path %s, got %s", dir, module.Path)
			}
			
			if tt.check != nil {
				tt.check(t, module)
			}
		})
	}
}

func TestParseHCLResourceWithTags(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create test Terraform file
	tfContent := `
resource "aws_instance" "test" {
  ami           = "ami-12345"
  instance_type = "t2.micro"
  
  tags = {
    Name        = "TestServer"
    Environment = "production"
    Team        = "backend"
  }
}
`
	if err := os.WriteFile(filepath.Join(tmpDir, "main.tf"), []byte(tfContent), 0644); err != nil {
		t.Fatal(err)
	}
	
	// Execute parsing
	module, err := ParseModuleHCL(tmpDir)
	if err != nil {
		t.Fatalf("Failed to parse module: %v", err)
	}
	
	// Verify there is 1 resource
	if len(module.Resources) != 1 {
		t.Fatalf("Expected 1 resource, got %d", len(module.Resources))
	}
	
	resource := module.Resources[0]
	
	// Verify tags are stored as JSON
	tagsValue, exists := resource.Config["tags"]
	if !exists {
		t.Fatal("tags not found in resource config")
	}
	
	t.Logf("Tags value: %s", tagsValue)
	
	// Verify it can be parsed as JSON
	var tags map[string]interface{}
	if err := json.Unmarshal([]byte(tagsValue), &tags); err != nil {
		// Case when it's <complex_expression>
		if tagsValue == "<complex_expression>" {
			t.Errorf("Tags were not parsed as JSON, got: %s", tagsValue)
		} else {
			t.Errorf("Failed to parse tags as JSON: %v", err)
		}
	} else {
		// Verification when JSON is correctly parsed
		if tags["Name"] != "TestServer" {
			t.Errorf("Expected Name=TestServer, got %v", tags["Name"])
		}
		if tags["Environment"] != "production" {
			t.Errorf("Expected Environment=production, got %v", tags["Environment"])
		}
		if tags["Team"] != "backend" {
			t.Errorf("Expected Team=backend, got %v", tags["Team"])
		}
	}
}