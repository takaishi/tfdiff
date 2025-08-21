package tfdiff

import (
	"os"
	"testing"
)

// writeTestFile creates a temporary file with the given content
func writeTestFile(t *testing.T, filename, content string) {
	err := os.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file %s: %v", filename, err)
	}
}

// removeTestFile removes a temporary file
func removeTestFile(filename string) {
	os.Remove(filename)
}

func TestNestedBlockParsing(t *testing.T) {
	// Test parsing of nested blocks like metadata_options
	leftContent := `
resource "aws_instance" "example" {
  ami           = "ami-0c02fb55956c7d316"
  instance_type = "t3.micro"
  
  metadata_options {
    http_endpoint = "enabled"
    http_tokens   = "required"
  }
}
`

	rightContent := `
resource "aws_instance" "example" {
  ami           = "ami-0c02fb55956c7d316"
  instance_type = "t3.micro"
  
  metadata_options {
    http_tokens   = "required"
    http_endpoint = "enabled"
  }
}
`

	// Create temporary directories for each test
	leftDir := t.TempDir()
	rightDir := t.TempDir()
	
	leftFile := leftDir + "/main.tf"
	rightFile := rightDir + "/main.tf"
	
	writeTestFile(t, leftFile, leftContent)
	writeTestFile(t, rightFile, rightContent)

	// Parse both files
	leftDef, err := ParseModuleHCL(leftDir)
	if err != nil {
		t.Fatalf("Failed to parse left module: %v", err)
	}

	rightDef, err := ParseModuleHCL(rightDir)
	if err != nil {
		t.Fatalf("Failed to parse right module: %v", err)
	}

	// Verify that nested blocks are parsed correctly
	if len(leftDef.Resources) != 1 {
		t.Fatalf("Expected 1 resource in left module, got %d", len(leftDef.Resources))
	}

	leftResource := leftDef.Resources[0]
	if leftResource.Type != "aws_instance" || leftResource.Name != "example" {
		t.Fatalf("Unexpected resource: %s.%s", leftResource.Type, leftResource.Name)
	}

	// Check that _blocks key exists and contains metadata_options
	blocks, ok := leftResource.Config["_blocks"]
	if !ok {
		t.Fatalf("Expected _blocks key in config, but not found")
	}

	blocksMap, ok := blocks.(map[string][]map[string]interface{})
	if !ok {
		t.Fatalf("Expected _blocks to be map[string][]map[string]interface{}, got %T", blocks)
	}

	metadataOptions, ok := blocksMap["metadata_options"]
	if !ok {
		t.Fatalf("Expected metadata_options block, but not found")
	}

	if len(metadataOptions) != 1 {
		t.Fatalf("Expected 1 metadata_options block, got %d", len(metadataOptions))
	}

	metadataBlock := metadataOptions[0]
	
	// Check that both http_endpoint and http_tokens are present
	if metadataBlock["http_endpoint"] != "enabled" {
		t.Errorf("Expected http_endpoint=enabled, got %v", metadataBlock["http_endpoint"])
	}
	
	if metadataBlock["http_tokens"] != "required" {
		t.Errorf("Expected http_tokens=required, got %v", metadataBlock["http_tokens"])
	}

	// Compare the modules - they should be equal despite different ordering
	config := DefaultComparisonConfig()
	config.IgnoreArguments = false
	
	result := CompareModules(leftDef, rightDef, config)
	
	// Should have no differences since only order changed
	if len(result.Diffs) != 0 {
		t.Errorf("Expected no differences, but got %d differences", len(result.Diffs))
		for _, diff := range result.Diffs {
			t.Logf("Diff: %+v", diff)
		}
	}
}

func TestNestedBlockWithIngressEgress(t *testing.T) {
	// Test more complex nested blocks like security group rules
	leftContent := `
resource "aws_security_group" "example" {
  name = "example"
  
  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
  
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}
`

	rightContent := `
resource "aws_security_group" "example" {
  name = "example"
  
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
  
  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}
`

	// Create temporary directories for each test
	leftDir := t.TempDir()
	rightDir := t.TempDir()
	
	leftFile := leftDir + "/main.tf"
	rightFile := rightDir + "/main.tf"
	
	writeTestFile(t, leftFile, leftContent)
	writeTestFile(t, rightFile, rightContent)

	// Parse both files
	leftDef, err := ParseModuleHCL(leftDir)
	if err != nil {
		t.Fatalf("Failed to parse left module: %v", err)
	}

	rightDef, err := ParseModuleHCL(rightDir)
	if err != nil {
		t.Fatalf("Failed to parse right module: %v", err)
	}

	// Compare the modules - they should be equal despite different ordering
	config := DefaultComparisonConfig()
	config.IgnoreArguments = false
	
	result := CompareModules(leftDef, rightDef, config)
	
	// Should have no differences since only order changed
	if len(result.Diffs) != 0 {
		t.Errorf("Expected no differences, but got %d differences", len(result.Diffs))
		for _, diff := range result.Diffs {
			t.Logf("Diff: %+v", diff)
		}
	}

	// Verify that nested blocks are parsed correctly
	leftResource := leftDef.Resources[0]
	blocks, ok := leftResource.Config["_blocks"]
	if !ok {
		t.Fatalf("Expected _blocks key in config, but not found")
	}

	blocksMap, ok := blocks.(map[string][]map[string]interface{})
	if !ok {
		t.Fatalf("Expected _blocks to be map[string][]map[string]interface{}, got %T", blocks)
	}

	// Check that both ingress and egress blocks exist
	ingress, hasIngress := blocksMap["ingress"]
	egress, hasEgress := blocksMap["egress"]
	
	if !hasIngress {
		t.Errorf("Expected ingress blocks, but not found")
	}
	if !hasEgress {
		t.Errorf("Expected egress blocks, but not found")
	}
	
	if len(ingress) != 1 {
		t.Errorf("Expected 1 ingress block, got %d", len(ingress))
	}
	if len(egress) != 1 {
		t.Errorf("Expected 1 egress block, got %d", len(egress))
	}
}

func TestNestedBlockModification(t *testing.T) {
	// Test detection of modifications in nested blocks
	leftContent := `
resource "aws_instance" "example" {
  ami           = "ami-0c02fb55956c7d316"
  instance_type = "t3.micro"
  
  metadata_options {
    http_endpoint = "enabled"
    http_tokens   = "required"
  }
}
`

	rightContent := `
resource "aws_instance" "example" {
  ami           = "ami-0c02fb55956c7d316"
  instance_type = "t3.micro"
  
  metadata_options {
    http_endpoint = "disabled"
    http_tokens   = "required"
  }
}
`

	// Create temporary directories for each test
	leftDir := t.TempDir()
	rightDir := t.TempDir()
	
	leftFile := leftDir + "/main.tf"
	rightFile := rightDir + "/main.tf"
	
	writeTestFile(t, leftFile, leftContent)
	writeTestFile(t, rightFile, rightContent)

	// Parse both files
	leftDef, err := ParseModuleHCL(leftDir)
	if err != nil {
		t.Fatalf("Failed to parse left module: %v", err)
	}

	rightDef, err := ParseModuleHCL(rightDir)
	if err != nil {
		t.Fatalf("Failed to parse right module: %v", err)
	}

	// Compare the modules - should detect the difference
	config := DefaultComparisonConfig()
	config.IgnoreArguments = false
	
	result := CompareModules(leftDef, rightDef, config)
	
	// Should have differences since http_endpoint changed
	if len(result.Diffs) == 0 {
		t.Errorf("Expected differences, but got none")
	}
}