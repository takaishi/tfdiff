package tfdiff

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
)

// CompareModules compares two module definitions and returns differences
func CompareModules(left, right *ModuleDefinition, config ComparisonConfig) *ComparisonResult {
	result := &ComparisonResult{
		LeftPath:  left.Path,
		RightPath: right.Path,
		Diffs:     []Diff{},
	}

	// If "all" is specified, compare everything
	if containsLevel(config.Levels, ComparisonLevelAll) {
		diffs := compareModuleCalls(left.ModuleCalls, right.ModuleCalls, config)
		result.Diffs = append(result.Diffs, diffs...)
		
		diffs = compareOutputs(left.Outputs, right.Outputs)
		result.Diffs = append(result.Diffs, diffs...)
		
		diffs = compareResources(left.Resources, right.Resources, config)
		result.Diffs = append(result.Diffs, diffs...)
		
		diffs = compareDataSources(left.DataSources, right.DataSources, config)
		result.Diffs = append(result.Diffs, diffs...)
		
		diffs = compareVariables(left.Variables, right.Variables)
		result.Diffs = append(result.Diffs, diffs...)
	} else {
		// Compare based on configured levels
		for _, level := range config.Levels {
			switch level {
			case ComparisonLevelModuleCalls:
				diffs := compareModuleCalls(left.ModuleCalls, right.ModuleCalls, config)
				result.Diffs = append(result.Diffs, diffs...)
			case ComparisonLevelOutputs:
				diffs := compareOutputs(left.Outputs, right.Outputs)
				result.Diffs = append(result.Diffs, diffs...)
			case ComparisonLevelResources:
				diffs := compareResources(left.Resources, right.Resources, config)
				result.Diffs = append(result.Diffs, diffs...)
			case ComparisonLevelDataSources:
				diffs := compareDataSources(left.DataSources, right.DataSources, config)
				result.Diffs = append(result.Diffs, diffs...)
			case ComparisonLevelVariables:
				diffs := compareVariables(left.Variables, right.Variables)
				result.Diffs = append(result.Diffs, diffs...)
			}
		}
	}

	// Calculate summary
	for _, diff := range result.Diffs {
		switch diff.Type {
		case DiffTypeAdded:
			result.Summary.Added++
		case DiffTypeRemoved:
			result.Summary.Removed++
		case DiffTypeModified:
			result.Summary.Modified++
		}
	}
	result.Summary.Total = len(result.Diffs)

	return result
}

// compareModuleCalls compares module calls between two modules
func compareModuleCalls(left, right []ModuleCall, config ComparisonConfig) []Diff {
	var diffs []Diff

	leftMap := make(map[string]ModuleCall)
	rightMap := make(map[string]ModuleCall)

	for _, mc := range left {
		leftMap[mc.Name] = mc
	}
	for _, mc := range right {
		rightMap[mc.Name] = mc
	}

	// Find added module calls
	for name, rightCall := range rightMap {
		if _, exists := leftMap[name]; !exists {
			diffs = append(diffs, Diff{
				Type:    DiffTypeAdded,
				Level:   "module_call",
				Element: name,
				After:   rightCall,
				Message: fmt.Sprintf("Module call '%s' was added", name),
			})
		}
	}

	// Find removed module calls
	for name, leftCall := range leftMap {
		if _, exists := rightMap[name]; !exists {
			diffs = append(diffs, Diff{
				Type:    DiffTypeRemoved,
				Level:   "module_call",
				Element: name,
				Before:  leftCall,
				Message: fmt.Sprintf("Module call '%s' was removed", name),
			})
		}
	}

	// Find modified module calls
	for name, leftCall := range leftMap {
		if rightCall, exists := rightMap[name]; exists {
			if !moduleCallsEqual(leftCall, rightCall, config) {
				diffs = append(diffs, Diff{
					Type:    DiffTypeModified,
					Level:   "module_call",
					Element: name,
					Before:  leftCall,
					After:   rightCall,
					Message: fmt.Sprintf("Module call '%s' was modified", name),
				})
			}
		}
	}

	return diffs
}

// compareOutputs compares outputs between two modules
func compareOutputs(left, right []Output) []Diff {
	var diffs []Diff

	leftMap := make(map[string]Output)
	rightMap := make(map[string]Output)

	for _, o := range left {
		leftMap[o.Name] = o
	}
	for _, o := range right {
		rightMap[o.Name] = o
	}

	// Find added outputs
	for name, rightOutput := range rightMap {
		if _, exists := leftMap[name]; !exists {
			diffs = append(diffs, Diff{
				Type:    DiffTypeAdded,
				Level:   "output",
				Element: name,
				After:   rightOutput,
				Message: fmt.Sprintf("Output '%s' was added", name),
			})
		}
	}

	// Find removed outputs
	for name, leftOutput := range leftMap {
		if _, exists := rightMap[name]; !exists {
			diffs = append(diffs, Diff{
				Type:    DiffTypeRemoved,
				Level:   "output",
				Element: name,
				Before:  leftOutput,
				Message: fmt.Sprintf("Output '%s' was removed", name),
			})
		}
	}

	// Find modified outputs
	for name, leftOutput := range leftMap {
		if rightOutput, exists := rightMap[name]; exists {
			if !outputsEqual(leftOutput, rightOutput) {
				diffs = append(diffs, Diff{
					Type:    DiffTypeModified,
					Level:   "output",
					Element: name,
					Before:  leftOutput,
					After:   rightOutput,
					Message: fmt.Sprintf("Output '%s' was modified", name),
				})
			}
		}
	}

	return diffs
}

// compareResources compares resources between two modules
func compareResources(left, right []Resource, config ComparisonConfig) []Diff {
	var diffs []Diff

	leftMap := make(map[string]Resource)
	rightMap := make(map[string]Resource)

	for _, r := range left {
		key := fmt.Sprintf("%s.%s", r.Type, r.Name)
		leftMap[key] = r
	}
	for _, r := range right {
		key := fmt.Sprintf("%s.%s", r.Type, r.Name)
		rightMap[key] = r
	}

	// Find added resources
	for key, rightResource := range rightMap {
		if _, exists := leftMap[key]; !exists {
			diffs = append(diffs, Diff{
				Type:    DiffTypeAdded,
				Level:   "resource",
				Element: key,
				After:   rightResource,
				Message: fmt.Sprintf("Resource '%s' was added", key),
			})
		}
	}

	// Find removed resources
	for key, leftResource := range leftMap {
		if _, exists := rightMap[key]; !exists {
			diffs = append(diffs, Diff{
				Type:    DiffTypeRemoved,
				Level:   "resource",
				Element: key,
				Before:  leftResource,
				Message: fmt.Sprintf("Resource '%s' was removed", key),
			})
		}
	}

	// Find modified resources
	for key, leftResource := range leftMap {
		if rightResource, exists := rightMap[key]; exists {
			if !resourcesEqual(leftResource, rightResource, config) {
				diffs = append(diffs, Diff{
					Type:    DiffTypeModified,
					Level:   "resource",
					Element: key,
					Before:  leftResource,
					After:   rightResource,
					Message: fmt.Sprintf("Resource '%s' was modified", key),
				})
			}
		}
	}

	return diffs
}

// compareDataSources compares data sources between two modules
func compareDataSources(left, right []DataSource, config ComparisonConfig) []Diff {
	var diffs []Diff

	leftMap := make(map[string]DataSource)
	rightMap := make(map[string]DataSource)

	for _, ds := range left {
		key := fmt.Sprintf("%s.%s", ds.Type, ds.Name)
		leftMap[key] = ds
	}
	for _, ds := range right {
		key := fmt.Sprintf("%s.%s", ds.Type, ds.Name)
		rightMap[key] = ds
	}

	// Find added data sources
	for key, rightDS := range rightMap {
		if _, exists := leftMap[key]; !exists {
			diffs = append(diffs, Diff{
				Type:    DiffTypeAdded,
				Level:   "data_source",
				Element: key,
				After:   rightDS,
				Message: fmt.Sprintf("Data source '%s' was added", key),
			})
		}
	}

	// Find removed data sources
	for key, leftDS := range leftMap {
		if _, exists := rightMap[key]; !exists {
			diffs = append(diffs, Diff{
				Type:    DiffTypeRemoved,
				Level:   "data_source",
				Element: key,
				Before:  leftDS,
				Message: fmt.Sprintf("Data source '%s' was removed", key),
			})
		}
	}

	// Find modified data sources
	for key, leftDS := range leftMap {
		if rightDS, exists := rightMap[key]; exists {
			if !dataSourcesEqual(leftDS, rightDS, config) {
				diffs = append(diffs, Diff{
					Type:    DiffTypeModified,
					Level:   "data_source",
					Element: key,
					Before:  leftDS,
					After:   rightDS,
					Message: fmt.Sprintf("Data source '%s' was modified", key),
				})
			}
		}
	}

	return diffs
}

// compareVariables compares variables between two modules
func compareVariables(left, right []Variable) []Diff {
	var diffs []Diff

	leftMap := make(map[string]Variable)
	rightMap := make(map[string]Variable)

	for _, v := range left {
		leftMap[v.Name] = v
	}
	for _, v := range right {
		rightMap[v.Name] = v
	}

	// Find added variables
	for name, rightVar := range rightMap {
		if _, exists := leftMap[name]; !exists {
			diffs = append(diffs, Diff{
				Type:    DiffTypeAdded,
				Level:   "variable",
				Element: name,
				After:   rightVar,
				Message: fmt.Sprintf("Variable '%s' was added", name),
			})
		}
	}

	// Find removed variables
	for name, leftVar := range leftMap {
		if _, exists := rightMap[name]; !exists {
			diffs = append(diffs, Diff{
				Type:    DiffTypeRemoved,
				Level:   "variable",
				Element: name,
				Before:  leftVar,
				Message: fmt.Sprintf("Variable '%s' was removed", name),
			})
		}
	}

	// Find modified variables
	for name, leftVar := range leftMap {
		if rightVar, exists := rightMap[name]; exists {
			if !variablesEqual(leftVar, rightVar) {
				diffs = append(diffs, Diff{
					Type:    DiffTypeModified,
					Level:   "variable",
					Element: name,
					Before:  leftVar,
					After:   rightVar,
					Message: fmt.Sprintf("Variable '%s' was modified", name),
				})
			}
		}
	}

	return diffs
}

// Equality comparison functions

func moduleCallsEqual(left, right ModuleCall, config ComparisonConfig) bool {
	if left.Name != right.Name || left.Source != right.Source || left.Version != right.Version {
		return false
	}

	if !config.IgnoreArguments && !reflect.DeepEqual(left.Args, right.Args) {
		return false
	}

	return true
}

func outputsEqual(left, right Output) bool {
	if left.Name != right.Name || left.Sensitive != right.Sensitive || left.Description != right.Description {
		return false
	}

	return true
}

func resourcesEqual(left, right Resource, config ComparisonConfig) bool {
	if left.Type != right.Type || left.Name != right.Name {
		return false
	}

	if !config.IgnoreArguments && !configsEqual(left.Config, right.Config) {
		return false
	}

	return true
}

// configsEqual compares two config maps, handling nested blocks and JSON strings semantically
func configsEqual(left, right map[string]interface{}) bool {
	if len(left) != len(right) {
		return false
	}

	for key, leftValue := range left {
		rightValue, exists := right[key]
		if !exists {
			return false
		}

		if !valuesEqual(leftValue, rightValue) {
			return false
		}
	}

	return true
}

// valuesEqual compares two interface{} values recursively
func valuesEqual(left, right interface{}) bool {
	switch leftVal := left.(type) {
	case string:
		rightVal, ok := right.(string)
		if !ok {
			return false
		}
		// Try to parse as JSON for semantic comparison
		if isJSON(leftVal) && isJSON(rightVal) {
			return jsonEqual(leftVal, rightVal)
		}
		return leftVal == rightVal
		
	case map[string]interface{}:
		rightVal, ok := right.(map[string]interface{})
		if !ok {
			return false
		}
		return configsEqual(leftVal, rightVal)
		
	case []map[string]interface{}:
		rightVal, ok := right.([]map[string]interface{})
		if !ok {
			return false
		}
		if len(leftVal) != len(rightVal) {
			return false
		}
		// For nested blocks, we need to compare them in a way that's order-independent
		// for blocks that can be reordered (like security group rules)
		return sliceMapsEqual(leftVal, rightVal)
		
	case []interface{}:
		rightVal, ok := right.([]interface{})
		if !ok {
			return false
		}
		if len(leftVal) != len(rightVal) {
			return false
		}
		for i, leftItem := range leftVal {
			if !valuesEqual(leftItem, rightVal[i]) {
				return false
			}
		}
		return true
		
	default:
		// For other types, use reflection for deep comparison to avoid panic
		return reflect.DeepEqual(left, right)
	}
}

// sliceMapsEqual compares two slices of maps, handling order-independent comparison
func sliceMapsEqual(left, right []map[string]interface{}) bool {
	if len(left) != len(right) {
		return false
	}
	
	// For blocks like ingress/egress that can be reordered, we need to find matching pairs
	matched := make([]bool, len(right))
	
	for _, leftItem := range left {
		found := false
		for j, rightItem := range right {
			if !matched[j] && configsEqual(leftItem, rightItem) {
				matched[j] = true
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	
	return true
}

// isJSON checks if a string is valid JSON
func isJSON(str string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(str), &js) == nil
}

// jsonEqual compares two JSON strings semantically
func jsonEqual(a, b string) bool {
	var aData, bData interface{}
	
	if err := json.Unmarshal([]byte(a), &aData); err != nil {
		return false
	}
	
	if err := json.Unmarshal([]byte(b), &bData); err != nil {
		return false
	}
	
	return reflect.DeepEqual(aData, bData)
}

func dataSourcesEqual(left, right DataSource, config ComparisonConfig) bool {
	if left.Type != right.Type || left.Name != right.Name {
		return false
	}

	if !config.IgnoreArguments && !configsEqual(left.Config, right.Config) {
		return false
	}

	return true
}

func variablesEqual(left, right Variable) bool {
	if left.Name != right.Name || left.Type != right.Type || left.Description != right.Description {
		return false
	}

	// Handle DefaultValue comparison
	leftVal := left.DefaultValue
	rightVal := right.DefaultValue

	// If both are JSON strings, parse and compare structurally
	if isJSON(leftVal) && isJSON(rightVal) {
		return jsonEqual(leftVal, rightVal)
	}

	// Otherwise, use simple string comparison
	return leftVal == rightVal
}

// containsLevel checks if a slice of ComparisonLevel contains a specific level
func containsLevel(levels []ComparisonLevel, target ComparisonLevel) bool {
	for _, level := range levels {
		if level == target {
			return true
		}
	}
	return false
}

// SortDiffs sorts diffs by level, type, and element name for consistent output
func SortDiffs(diffs []Diff) {
	sort.Slice(diffs, func(i, j int) bool {
		if diffs[i].Level != diffs[j].Level {
			return diffs[i].Level < diffs[j].Level
		}
		if diffs[i].Type != diffs[j].Type {
			return diffs[i].Type < diffs[j].Type
		}
		return diffs[i].Element < diffs[j].Element
	})
}