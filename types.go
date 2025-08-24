package tfdiff

// ModuleCall represents a module call in Terraform configuration
type ModuleCall struct {
	Name     string            `json:"name"`
	Source   string            `json:"source"`
	Version  string            `json:"version,omitempty"`
	Args     map[string]string `json:"args,omitempty"`
	Position string            `json:"position,omitempty"`
}

// Output represents a Terraform output value
type Output struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Sensitive   bool   `json:"sensitive,omitempty"`
	Value       string `json:"value,omitempty"`
	Position    string `json:"position,omitempty"`
}

// Resource represents a Terraform resource
type Resource struct {
	Type      string                 `json:"type"`
	Name      string                 `json:"name"`
	Config    map[string]interface{} `json:"config,omitempty"`
	Position  string                 `json:"position,omitempty"`
}

// DataSource represents a Terraform data source
type DataSource struct {
	Type     string                 `json:"type"`
	Name     string                 `json:"name"`
	Config   map[string]interface{} `json:"config,omitempty"`
	Position string                 `json:"position,omitempty"`
}

// Variable represents a Terraform variable
type Variable struct {
	Name         string `json:"name"`
	Type         string `json:"type,omitempty"`
	Description  string `json:"description,omitempty"`
	DefaultValue string `json:"default_value,omitempty"`
	Position     string `json:"position,omitempty"`
}

// ModuleDefinition represents the complete definition of a Terraform module
type ModuleDefinition struct {
	Path        string            `json:"path"`
	ModuleCalls []ModuleCall      `json:"module_calls,omitempty"`
	Outputs     []Output          `json:"outputs,omitempty"`
	Resources   []Resource        `json:"resources,omitempty"`
	DataSources []DataSource      `json:"data_sources,omitempty"`
	Variables   []Variable        `json:"variables,omitempty"`
}

// ComparisonLevel defines what elements to compare
type ComparisonLevel string

const (
	ComparisonLevelModuleCalls ComparisonLevel = "module_calls"
	ComparisonLevelOutputs     ComparisonLevel = "outputs"
	ComparisonLevelResources   ComparisonLevel = "resources"
	ComparisonLevelDataSources ComparisonLevel = "data_sources"
	ComparisonLevelVariables   ComparisonLevel = "variables"
	ComparisonLevelAll         ComparisonLevel = "all"
)

// ComparisonConfig defines configuration for comparison
type ComparisonConfig struct {
	Levels          []ComparisonLevel `json:"levels"`
	IgnoreArguments bool              `json:"ignore_arguments"`
}

// DefaultComparisonConfig returns default comparison configuration
func DefaultComparisonConfig() ComparisonConfig {
	return ComparisonConfig{
		Levels: []ComparisonLevel{
			ComparisonLevelModuleCalls,
			ComparisonLevelOutputs,
			ComparisonLevelResources,
			ComparisonLevelDataSources,
			ComparisonLevelVariables,
		},
		IgnoreArguments: true,
	}
}

// DiffType represents the type of difference found
type DiffType string

const (
	DiffTypeAdded    DiffType = "added"
	DiffTypeRemoved  DiffType = "removed"
	DiffTypeModified DiffType = "modified"
)

// Diff represents a difference between two elements
type Diff struct {
	Type     DiffType    `json:"type"`
	Level    string      `json:"level"`
	Element  string      `json:"element"`
	Before   interface{} `json:"before,omitempty"`
	After    interface{} `json:"after,omitempty"`
	Message  string      `json:"message,omitempty"`
}

// ComparisonResult represents the result of comparing two modules
type ComparisonResult struct {
	LeftPath  string `json:"left_path"`
	RightPath string `json:"right_path"`
	Diffs     []Diff `json:"diffs"`
	Summary   struct {
		Added    int `json:"added"`
		Removed  int `json:"removed"`
		Modified int `json:"modified"`
		Total    int `json:"total"`
	} `json:"summary"`
}