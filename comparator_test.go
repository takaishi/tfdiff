package tfdiff

import (
	"reflect"
	"strings"
	"testing"
)

// stringMapToInterface converts map[string]string to map[string]interface{}
func stringMapToInterface(m map[string]string) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range m {
		result[k] = v
	}
	return result
}

func TestResourcesEqual_WithNestedObjects(t *testing.T) {
	tests := []struct {
		name     string
		left     Resource
		right    Resource
		config   ComparisonConfig
		expected bool
	}{
		{
			name: "Resources with same nested objects should be equal",
			left: Resource{
				Type: "aws_instance",
				Name: "example",
				Config: stringMapToInterface(map[string]string{
					"ami":           "ami-12345",
					"instance_type": "t2.micro",
					"tags":          `{"Name": "test", "Environment": "dev"}`,
				}),
			},
			right: Resource{
				Type: "aws_instance",
				Name: "example",
				Config: stringMapToInterface(map[string]string{
					"ami":           "ami-12345",
					"instance_type": "t2.micro",
					"tags":          `{"Name": "test", "Environment": "dev"}`,
				}),
			},
			config:   ComparisonConfig{IgnoreArguments: false},
			expected: true,
		},
		{
			name: "JSON objects with different key ordering but same content should be equal",
			left: Resource{
				Type: "aws_instance",
				Name: "example",
				Config: stringMapToInterface(map[string]string{
					"tags": `{"Environment": "dev", "Name": "test"}`,
				}),
			},
			right: Resource{
				Type: "aws_instance",
				Name: "example",
				Config: stringMapToInterface(map[string]string{
					"tags": `{"Name": "test", "Environment": "dev"}`,
				}),
			},
			config:   ComparisonConfig{IgnoreArguments: false},
			expected: true, // Should be true as JSON content is the same
		},
		{
			name: "Resources with different nested objects should not be equal",
			left: Resource{
				Type: "aws_instance",
				Name: "example",
				Config: stringMapToInterface(map[string]string{
					"tags": `{"Name": "test", "Environment": "dev"}`,
				}),
			},
			right: Resource{
				Type: "aws_instance",
				Name: "example",
				Config: stringMapToInterface(map[string]string{
					"tags": `{"Name": "test", "Environment": "prod"}`,
				}),
			},
			config:   ComparisonConfig{IgnoreArguments: false},
			expected: false,
		},
		{
			name: "Complex nested JSON should be compared correctly",
			left: Resource{
				Type: "aws_security_group",
				Name: "example",
				Config: stringMapToInterface(map[string]string{
					"ingress": `[{"from_port": 80, "to_port": 80, "protocol": "tcp", "cidr_blocks": ["0.0.0.0/0"]}]`,
				}),
			},
			right: Resource{
				Type: "aws_security_group",
				Name: "example",
				Config: stringMapToInterface(map[string]string{
					"ingress": `[{"protocol": "tcp", "from_port": 80, "to_port": 80, "cidr_blocks": ["0.0.0.0/0"]}]`,
				}),
			},
			config:   ComparisonConfig{IgnoreArguments: false},
			expected: true, // Should be true if content is same even if object order in array differs
		},
		{
			name: "When IgnoreArguments is true, resources with different configs should be equal",
			left: Resource{
				Type: "aws_instance",
				Name: "example",
				Config: stringMapToInterface(map[string]string{
					"tags": `{"Name": "test1"}`,
				}),
			},
			right: Resource{
				Type: "aws_instance",
				Name: "example",
				Config: stringMapToInterface(map[string]string{
					"tags": `{"Name": "test2"}`,
				}),
			},
			config:   ComparisonConfig{IgnoreArguments: true},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resourcesEqual(tt.left, tt.right, tt.config)
			if result != tt.expected {
				t.Errorf("resourcesEqual() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCompareResources_WithNestedObjects(t *testing.T) {
	config := ComparisonConfig{IgnoreArguments: false}

	left := []Resource{
		{
			Type: "aws_instance",
			Name: "web",
			Config: stringMapToInterface(map[string]string{
				"ami":  "ami-12345",
				"tags": `{"Name": "web-server", "Environment": "prod"}`,
			}),
		},
	}

	right := []Resource{
		{
			Type: "aws_instance",
			Name: "web",
			Config: stringMapToInterface(map[string]string{
				"ami":  "ami-12345",
				"tags": `{"Environment": "prod", "Name": "web-server"}`, // Different ordering
			}),
		},
	}

	diffs := compareResources(left, right, config)

	// After fix, no differences should be detected for same JSON content with different ordering
	if len(diffs) != 0 {
		t.Errorf("Expected no differences for same JSON content, but got %d diffs", len(diffs))
	}
}

func TestMapComparison(t *testing.T) {
	// Test map comparison with reflect.DeepEqual
	map1 := map[string]string{
		"key1": "value1",
		"key2": `{"a": "1", "b": "2"}`,
	}

	map2 := map[string]string{
		"key1": "value1",
		"key2": `{"b": "2", "a": "1"}`, // Different JSON ordering
	}

	if reflect.DeepEqual(map1, map2) {
		t.Error("Maps with different JSON string values should not be equal")
	}
}

func TestVariableComparison_ArrayDifference(t *testing.T) {
	tests := []struct {
		name     string
		left     Variable
		right    Variable
		expected bool
	}{
		{
			name: "Same array values should be equal",
			left: Variable{
				Name:         "availability_zones",
				Type:         "list(string)",
				DefaultValue: `["us-west-2a", "us-west-2b"]`,
			},
			right: Variable{
				Name:         "availability_zones",
				Type:         "list(string)",
				DefaultValue: `["us-west-2a", "us-west-2b"]`,
			},
			expected: true,
		},
		{
			name: "Different array values should not be equal",
			left: Variable{
				Name:         "availability_zones",
				Type:         "list(string)",
				DefaultValue: `["us-west-2a", "us-west-2b"]`,
			},
			right: Variable{
				Name:         "availability_zones",
				Type:         "list(string)",
				DefaultValue: `["us-west-2a", "us-west-2b", "us-west-2c"]`,
			},
			expected: false,
		},
		{
			name: "Array with same elements in different order should be equal",
			left: Variable{
				Name:         "availability_zones",
				Type:         "list(string)",
				DefaultValue: `["us-west-2a", "us-west-2b"]`,
			},
			right: Variable{
				Name:         "availability_zones",
				Type:         "list(string)",
				DefaultValue: `["us-west-2b", "us-west-2a"]`,
			},
			expected: false, // Arrays should preserve order
		},
		{
			name: "String values should work normally",
			left: Variable{
				Name:         "environment",
				Type:         "string",
				DefaultValue: `"production"`,
			},
			right: Variable{
				Name:         "environment",
				Type:         "string",
				DefaultValue: `"staging"`,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := variablesEqual(tt.left, tt.right)
			if result != tt.expected {
				t.Errorf("variablesEqual() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestCompareVariables_WithArrayDifference(t *testing.T) {
	left := []Variable{
		{
			Name:         "availability_zones",
			Type:         "list(string)",
			DefaultValue: `["us-west-2a", "us-west-2b"]`,
		},
		{
			Name:         "environment",
			Type:         "string", 
			DefaultValue: "production",
		},
	}

	right := []Variable{
		{
			Name:         "availability_zones",
			Type:         "list(string)",
			DefaultValue: `["us-west-2a", "us-west-2b", "us-west-2c"]`,
		},
		{
			Name:         "environment",
			Type:         "string",
			DefaultValue: "staging",
		},
	}

	diffs := compareVariables(left, right)

	// Should detect 2 modified variables (availability_zones and environment)
	modifiedCount := 0
	for _, diff := range diffs {
		if diff.Type == DiffTypeModified {
			modifiedCount++
		}
	}

	if modifiedCount != 2 {
		t.Errorf("Expected 2 modified variables, got %d", modifiedCount)
	}

	// Check if availability_zones difference is detected
	foundArrayDiff := false
	for _, diff := range diffs {
		if diff.Element == "availability_zones" && diff.Type == DiffTypeModified {
			foundArrayDiff = true
			break
		}
	}

	if !foundArrayDiff {
		t.Error("Array difference for availability_zones was not detected")
	}
}

func TestRealFileVariableComparison(t *testing.T) {
	// Parse actual files to see what's happening
	leftDef, err := ParseModuleHCL("./examples/basic/left")
	if err != nil {
		t.Fatalf("Failed to parse left module: %v", err)
	}

	rightDef, err := ParseModuleHCL("./examples/basic/right")
	if err != nil {
		t.Fatalf("Failed to parse right module: %v", err)
	}

	// Print actual variable values for debugging
	t.Logf("Left variables count: %d", len(leftDef.Variables))
	for _, v := range leftDef.Variables {
		t.Logf("Left var: %s = %s", v.Name, v.DefaultValue)
	}

	t.Logf("Right variables count: %d", len(rightDef.Variables))
	for _, v := range rightDef.Variables {
		t.Logf("Right var: %s = %s", v.Name, v.DefaultValue)
	}

	// Compare variables
	diffs := compareVariables(leftDef.Variables, rightDef.Variables)
	t.Logf("Variable differences found: %d", len(diffs))
	
	for _, diff := range diffs {
		t.Logf("Diff: %s - %s: %s", diff.Type, diff.Element, diff.Message)
	}

	// The availability_zones difference should be detected
	foundArrayDiff := false
	for _, diff := range diffs {
		if diff.Element == "availability_zones" && diff.Type == DiffTypeModified {
			foundArrayDiff = true
			break
		}
	}

	if !foundArrayDiff {
		t.Error("availability_zones array difference was not detected in real files")
	}
}

func TestFullComparisonWithVariables(t *testing.T) {
	// Test the full pipeline like tfdiff command
	leftDef, err := ParseModuleHCL("./examples/basic/left")
	if err != nil {
		t.Fatalf("Failed to parse left module: %v", err)
	}

	rightDef, err := ParseModuleHCL("./examples/basic/right")
	if err != nil {
		t.Fatalf("Failed to parse right module: %v", err)
	}

	// Use default config with variables enabled
	config := DefaultComparisonConfig()
	config.IgnoreArguments = false // Enable argument comparison

	// Compare modules
	result := CompareModules(leftDef, rightDef, config)

	t.Logf("Total diffs found: %d", len(result.Diffs))

	// Log all diffs for debugging
	for i, diff := range result.Diffs {
		t.Logf("Diff %d: %s - %s: %s", i, diff.Type, diff.Level, diff.Element)
	}

	// Check if variable diffs exist
	variableDiffs := 0
	for _, diff := range result.Diffs {
		if diff.Level == "variable" {
			variableDiffs++
			t.Logf("Variable diff: %s %s", diff.Type, diff.Element)
		}
	}

	if variableDiffs == 0 {
		t.Error("No variable differences found, but they should exist")
	}

	// Format output
	output := FormatTextOutput(result, config, true)
	t.Logf("Formatted output:\n%s", output)

	// Check if availability_zones appears in output
	if !strings.Contains(output, "availability_zones") {
		t.Error("availability_zones should appear in formatted output")
	}
}
