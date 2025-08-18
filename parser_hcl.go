package tfdiff

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

// ParseModuleHCL parses a Terraform module directory using HCL parser directly
func ParseModuleHCL(modulePath string) (*ModuleDefinition, error) {
	// Check if directory exists
	if _, err := os.Stat(modulePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("directory does not exist: %s", modulePath)
	}

	parser := hclparse.NewParser()

	// Find all .tf files in the directory
	files, err := filepath.Glob(filepath.Join(modulePath, "*.tf"))
	if err != nil {
		return nil, fmt.Errorf("failed to find .tf files: %w", err)
	}

	def := &ModuleDefinition{
		Path: modulePath,
	}

	// If no .tf files found, return empty module definition
	if len(files) == 0 {
		return def, nil
	}

	// Parse each .tf file
	for _, file := range files {
		if err := parseFile(parser, file, def); err != nil {
			return nil, fmt.Errorf("failed to parse file %s: %w", file, err)
		}
	}

	return def, nil
}

func parseFile(parser *hclparse.Parser, filename string, def *ModuleDefinition) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	file, diags := parser.ParseHCL(content, filename)
	if diags.HasErrors() {
		return fmt.Errorf("failed to parse HCL file %s: %s", filename, diags.Error())
	}

	body, ok := file.Body.(*hclsyntax.Body)
	if !ok {
		return fmt.Errorf("expected HCL syntax body in file %s", filename)
	}

	// Parse different block types
	for _, block := range body.Blocks {
		switch block.Type {
		case "module":
			if err := parseModuleBlock(block, def, filename); err != nil {
				return fmt.Errorf("failed to parse module block: %w", err)
			}
		case "resource":
			if err := parseResourceBlock(block, def, filename); err != nil {
				return fmt.Errorf("failed to parse resource block: %w", err)
			}
		case "data":
			if err := parseDataBlock(block, def, filename); err != nil {
				return fmt.Errorf("failed to parse data block: %w", err)
			}
		case "output":
			if err := parseOutputBlock(block, def, filename); err != nil {
				return fmt.Errorf("failed to parse output block: %w", err)
			}
		case "variable":
			if err := parseVariableBlock(block, def, filename); err != nil {
				return fmt.Errorf("failed to parse variable block: %w", err)
			}
		}
	}

	return nil
}

func parseModuleBlock(block *hclsyntax.Block, def *ModuleDefinition, filename string) error {
	if len(block.Labels) != 1 {
		return fmt.Errorf("module block must have exactly one label")
	}

	moduleCall := ModuleCall{
		Name:     block.Labels[0],
		Args:     make(map[string]string),
		Position: fmt.Sprintf("%s:%d", filepath.Base(filename), block.DefRange().Start.Line),
	}

	// Parse attributes
	for name, attr := range block.Body.Attributes {
		value, err := evaluateExpression(attr.Expr)
		if err != nil {
			// If we can't evaluate, store the raw expression
			value = string(attr.Expr.Range().SliceBytes(block.Body.SrcRange.SliceBytes([]byte{})))
		}

		switch name {
		case "source":
			moduleCall.Source = value
		case "version":
			moduleCall.Version = value
		default:
			moduleCall.Args[name] = value
		}
	}

	def.ModuleCalls = append(def.ModuleCalls, moduleCall)
	return nil
}

func parseResourceBlock(block *hclsyntax.Block, def *ModuleDefinition, filename string) error {
	if len(block.Labels) != 2 {
		return fmt.Errorf("resource block must have exactly two labels")
	}

	resource := Resource{
		Type:     block.Labels[0],
		Name:     block.Labels[1],
		Config:   make(map[string]string),
		Position: fmt.Sprintf("%s:%d", filepath.Base(filename), block.DefRange().Start.Line),
	}

	// Parse attributes
	for name, attr := range block.Body.Attributes {
		value, err := evaluateExpression(attr.Expr)
		if err != nil {
			// Try to evaluate as HCL expression
			if val, diags := attr.Expr.Value(nil); !diags.HasErrors() {
				jsonStr, err := convertCtyToJSON(val)
				if err == nil {
					value = jsonStr
				} else {
					value = "<complex_expression>"
				}
			} else {
				value = "<complex_expression>"
			}
		}
		resource.Config[name] = value
	}

	// For nested blocks, we'll just mark them as complex expressions
	// This branch focuses only on JSON object support for attributes
	for _, nestedBlock := range block.Body.Blocks {
		blockKey := nestedBlock.Type
		if len(nestedBlock.Labels) > 0 {
			blockKey = fmt.Sprintf("%s.%s", nestedBlock.Type, strings.Join(nestedBlock.Labels, "."))
		}

		resource.Config[blockKey] = fmt.Sprintf("<%s_block>", nestedBlock.Type)
	}

	def.Resources = append(def.Resources, resource)
	return nil
}

func parseDataBlock(block *hclsyntax.Block, def *ModuleDefinition, filename string) error {
	if len(block.Labels) != 2 {
		return fmt.Errorf("data block must have exactly two labels")
	}

	dataSource := DataSource{
		Type:     block.Labels[0],
		Name:     block.Labels[1],
		Config:   make(map[string]string),
		Position: fmt.Sprintf("%s:%d", filepath.Base(filename), block.DefRange().Start.Line),
	}

	// Parse attributes
	for name, attr := range block.Body.Attributes {
		value, err := evaluateExpression(attr.Expr)
		if err != nil {
			value = "<complex_expression>"
		}
		dataSource.Config[name] = value
	}

	def.DataSources = append(def.DataSources, dataSource)
	return nil
}

func parseOutputBlock(block *hclsyntax.Block, def *ModuleDefinition, filename string) error {
	if len(block.Labels) != 1 {
		return fmt.Errorf("output block must have exactly one label")
	}

	output := Output{
		Name:     block.Labels[0],
		Position: fmt.Sprintf("%s:%d", filepath.Base(filename), block.DefRange().Start.Line),
	}

	// Parse attributes
	for name, attr := range block.Body.Attributes {
		value, err := evaluateExpression(attr.Expr)
		if err != nil {
			value = "<complex_expression>"
		}

		switch name {
		case "description":
			output.Description = value
		case "sensitive":
			output.Sensitive = (value == "true")
		case "value":
			output.Value = value
		}
	}

	def.Outputs = append(def.Outputs, output)
	return nil
}

func parseVariableBlock(block *hclsyntax.Block, def *ModuleDefinition, filename string) error {
	if len(block.Labels) != 1 {
		return fmt.Errorf("variable block must have exactly one label")
	}

	variable := Variable{
		Name:     block.Labels[0],
		Position: fmt.Sprintf("%s:%d", filepath.Base(filename), block.DefRange().Start.Line),
	}

	// Parse attributes
	for name, attr := range block.Body.Attributes {
		value, err := evaluateExpression(attr.Expr)
		if err != nil {
			value = "<complex_expression>"
		}

		switch name {
		case "type":
			variable.Type = value
		case "description":
			variable.Description = value
		case "default":
			variable.DefaultValue = value
		}
	}

	def.Variables = append(def.Variables, variable)
	return nil
}

// evaluateExpression tries to evaluate a simple HCL expression to a string
func evaluateExpression(expr hcl.Expression) (string, error) {
	// Handle simple literal values
	switch e := expr.(type) {
	case *hclsyntax.LiteralValueExpr:
		val := e.Val
		switch val.Type() {
		case cty.String:
			return val.AsString(), nil
		case cty.Number:
			f, _ := val.AsBigFloat().Float64()
			return fmt.Sprintf("%g", f), nil
		case cty.Bool:
			if val.True() {
				return "true", nil
			}
			return "false", nil
		}
	case *hclsyntax.TemplateExpr:
		// For simple templates, try to extract the string
		if len(e.Parts) == 1 {
			if lit, ok := e.Parts[0].(*hclsyntax.LiteralValueExpr); ok {
				if lit.Val.Type() == cty.String {
					return lit.Val.AsString(), nil
				}
			}
		}
	}

	// For complex expressions, return an error so caller can handle
	return "", fmt.Errorf("complex expression cannot be evaluated")
}

// convertCtyToJSON converts a cty.Value to JSON string
func convertCtyToJSON(val cty.Value) (string, error) {
	if val.IsNull() {
		return "null", nil
	}

	switch {
	case val.Type() == cty.String:
		return val.AsString(), nil
	case val.Type() == cty.Number:
		f, _ := val.AsBigFloat().Float64()
		return fmt.Sprintf("%g", f), nil
	case val.Type() == cty.Bool:
		if val.True() {
			return "true", nil
		}
		return "false", nil
	case val.Type().IsObjectType() || val.Type().IsMapType():
		result := make(map[string]interface{})
		valMap := val.AsValueMap()
		for k, v := range valMap {
			converted, err := convertCtyValue(v)
			if err != nil {
				return "", err
			}
			result[k] = converted
		}
		jsonBytes, err := json.Marshal(result)
		return string(jsonBytes), err
	case val.Type().IsListType() || val.Type().IsTupleType():
		var result []interface{}
		valSlice := val.AsValueSlice()
		for _, v := range valSlice {
			converted, err := convertCtyValue(v)
			if err != nil {
				return "", err
			}
			result = append(result, converted)
		}
		jsonBytes, err := json.Marshal(result)
		return string(jsonBytes), err
	default:
		return "", fmt.Errorf("unsupported cty type: %s", val.Type().FriendlyName())
	}
}

// convertCtyValue converts a cty.Value to a Go value
func convertCtyValue(val cty.Value) (interface{}, error) {
	if val.IsNull() {
		return nil, nil
	}

	switch {
	case val.Type() == cty.String:
		return val.AsString(), nil
	case val.Type() == cty.Number:
		f, _ := val.AsBigFloat().Float64()
		return f, nil
	case val.Type() == cty.Bool:
		return val.True(), nil
	case val.Type().IsObjectType() || val.Type().IsMapType():
		result := make(map[string]interface{})
		valMap := val.AsValueMap()
		for k, v := range valMap {
			converted, err := convertCtyValue(v)
			if err != nil {
				return nil, err
			}
			result[k] = converted
		}
		return result, nil
	case val.Type().IsListType() || val.Type().IsTupleType():
		var result []interface{}
		valSlice := val.AsValueSlice()
		for _, v := range valSlice {
			converted, err := convertCtyValue(v)
			if err != nil {
				return nil, err
			}
			result = append(result, converted)
		}
		return result, nil
	default:
		return nil, fmt.Errorf("unsupported cty type: %s", val.Type().FriendlyName())
	}
}
