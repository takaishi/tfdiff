package tfdiff

import (
	"reflect"
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