package tfdiff

import (
	"fmt"
	"sort"
	"strings"
)

// FormatTextOutput formats the comparison result in a unified diff format
func FormatTextOutput(result *ComparisonResult, config ComparisonConfig) string {
	return FormatDiffOutput(result, config)
}

// FormatDiffOutput formats the comparison result in a unified diff format
func FormatDiffOutput(result *ComparisonResult, config ComparisonConfig) string {
	var output strings.Builder
	
	output.WriteString(fmt.Sprintf("--- %s\n", result.LeftPath))
	output.WriteString(fmt.Sprintf("+++ %s\n", result.RightPath))
	
	if len(result.Diffs) == 0 {
		return output.String()
	}
	
	// Sort diffs by level and then by name for consistent output
	sortedDiffs := sortDiffsForDiffOutput(result.Diffs)
	
	for _, diff := range sortedDiffs {
		switch diff.Type {
		case DiffTypeRemoved:
			lines := strings.Split(formatDiffLine(diff, diff.Before, config), "\n")
			for _, line := range lines {
				output.WriteString(fmt.Sprintf("-%s\n", line))
			}
		case DiffTypeAdded:
			lines := strings.Split(formatDiffLine(diff, diff.After, config), "\n")
			for _, line := range lines {
				output.WriteString(fmt.Sprintf("+%s\n", line))
			}
		case DiffTypeModified:
			// For modified items, show attribute-level diffs
			attributeDiffs := formatAttributeDiff(diff, config)
			for _, line := range attributeDiffs {
				output.WriteString(line + "\n")
			}
		}
	}
	
	return output.String()
}

// sortDiffsForDiffOutput sorts diffs by level and name for consistent output
func sortDiffsForDiffOutput(diffs []Diff) []Diff {
	sorted := make([]Diff, len(diffs))
	copy(sorted, diffs)
	
	// Sort by level first, then by name within each level
	// This provides a consistent ordering for diff output
	return sorted
}

// formatDiffLine formats a single line for diff output
func formatDiffLine(diff Diff, item interface{}, config ComparisonConfig) string {
	switch diff.Level {
	case "module_call":
		if mc, ok := item.(ModuleCall); ok {
			lines := []string{fmt.Sprintf("module \"%s\" {", mc.Name)}
			lines = append(lines, fmt.Sprintf("  source  = \"%s\"", mc.Source))
			if mc.Version != "" {
				lines = append(lines, fmt.Sprintf("  version = \"%s\"", mc.Version))
			}
			
			if !config.IgnoreArguments && len(mc.Args) > 0 {
				for key, value := range mc.Args {
					if value != "" && value != "<complex_expression>" {
						lines = append(lines, fmt.Sprintf("  %s = \"%s\"", key, value))
					}
				}
			}
			lines = append(lines, "}")
			return strings.Join(lines, "\n")
		}
	case "output":
		if out, ok := item.(Output); ok {
			lines := []string{fmt.Sprintf("output \"%s\" {", out.Name)}
			if out.Description != "" {
				lines = append(lines, fmt.Sprintf("  description = \"%s\"", out.Description))
			}
			lines = append(lines, fmt.Sprintf("  sensitive = %t", out.Sensitive))
			if out.Value != "" {
				lines = append(lines, fmt.Sprintf("  value = \"%s\"", out.Value))
			}
			lines = append(lines, "}")
			return strings.Join(lines, "\n")
		}
	case "resource":
		if res, ok := item.(Resource); ok {
			lines := []string{fmt.Sprintf("resource \"%s\" \"%s\" {", res.Type, res.Name)}
			if !config.IgnoreArguments && len(res.Config) > 0 {
				for key, value := range res.Config {
					if value != "" && value != "<complex_expression>" && !strings.HasPrefix(value, "<") {
						lines = append(lines, fmt.Sprintf("  %s = \"%s\"", key, value))
					}
				}
			}
			lines = append(lines, "}")
			return strings.Join(lines, "\n")
		}
	case "data_source":
		if ds, ok := item.(DataSource); ok {
			lines := []string{fmt.Sprintf("data \"%s\" \"%s\" {", ds.Type, ds.Name)}
			if !config.IgnoreArguments && len(ds.Config) > 0 {
				for key, value := range ds.Config {
					if value != "" && value != "<complex_expression>" && !strings.HasPrefix(value, "<") {
						lines = append(lines, fmt.Sprintf("  %s = \"%s\"", key, value))
					}
				}
			}
			lines = append(lines, "}")
			return strings.Join(lines, "\n")
		}
	case "variable":
		if v, ok := item.(Variable); ok {
			lines := []string{fmt.Sprintf("variable \"%s\" {", v.Name)}
			if v.Type != "" {
				lines = append(lines, fmt.Sprintf("  type = \"%s\"", v.Type))
			}
			if v.Description != "" {
				lines = append(lines, fmt.Sprintf("  description = \"%s\"", v.Description))
			}
			if v.DefaultValue != "" {
				lines = append(lines, fmt.Sprintf("  default = %s", v.DefaultValue))
			}
			lines = append(lines, "}")
			return strings.Join(lines, "\n")
		}
	}
	return diff.Message
}

// formatAttributeDiff formats attribute-level differences for modified items
func formatAttributeDiff(diff Diff, config ComparisonConfig) []string {
	var lines []string
	
	switch diff.Level {
	case "module_call":
		if before, okBefore := diff.Before.(ModuleCall); okBefore {
			if after, okAfter := diff.After.(ModuleCall); okAfter {
				lines = append(lines, fmt.Sprintf(" module \"%s\" {", before.Name))
				
				// Compare basic attributes
				if before.Source != after.Source {
					lines = append(lines, fmt.Sprintf("-  source  = \"%s\"", before.Source))
					lines = append(lines, fmt.Sprintf("+  source  = \"%s\"", after.Source))
				}
				if before.Version != after.Version {
					lines = append(lines, fmt.Sprintf("-  version = \"%s\"", before.Version))
					lines = append(lines, fmt.Sprintf("+  version = \"%s\"", after.Version))
				}
				
				// Compare arguments if not ignoring them
				if !config.IgnoreArguments {
					lines = append(lines, compareMapAttributes(before.Args, after.Args)...)
				}
				
				lines = append(lines, " }")
			}
		}
	case "resource":
		if before, okBefore := diff.Before.(Resource); okBefore {
			if after, okAfter := diff.After.(Resource); okAfter {
				lines = append(lines, fmt.Sprintf(" resource \"%s\" \"%s\" {", before.Type, before.Name))
				
				// Compare config if not ignoring arguments
				if !config.IgnoreArguments {
					lines = append(lines, compareMapAttributes(before.Config, after.Config)...)
				}
				
				lines = append(lines, " }")
			}
		}
	case "data_source":
		if before, okBefore := diff.Before.(DataSource); okBefore {
			if after, okAfter := diff.After.(DataSource); okAfter {
				lines = append(lines, fmt.Sprintf(" data \"%s\" \"%s\" {", before.Type, before.Name))
				
				// Compare config if not ignoring arguments
				if !config.IgnoreArguments {
					lines = append(lines, compareMapAttributes(before.Config, after.Config)...)
				}
				
				lines = append(lines, " }")
			}
		}
	case "output":
		if before, okBefore := diff.Before.(Output); okBefore {
			if after, okAfter := diff.After.(Output); okAfter {
				lines = append(lines, fmt.Sprintf(" output \"%s\" {", before.Name))
				
				if before.Description != after.Description {
					if before.Description != "" {
						lines = append(lines, fmt.Sprintf("-  description = \"%s\"", before.Description))
					}
					if after.Description != "" {
						lines = append(lines, fmt.Sprintf("+  description = \"%s\"", after.Description))
					}
				}
				if before.Sensitive != after.Sensitive {
					lines = append(lines, fmt.Sprintf("-  sensitive = %t", before.Sensitive))
					lines = append(lines, fmt.Sprintf("+  sensitive = %t", after.Sensitive))
				}
				if before.Value != after.Value {
					if before.Value != "" {
						lines = append(lines, fmt.Sprintf("-  value = \"%s\"", before.Value))
					}
					if after.Value != "" {
						lines = append(lines, fmt.Sprintf("+  value = \"%s\"", after.Value))
					}
				}
				
				lines = append(lines, " }")
			}
		}
	case "variable":
		if before, okBefore := diff.Before.(Variable); okBefore {
			if after, okAfter := diff.After.(Variable); okAfter {
				lines = append(lines, fmt.Sprintf(" variable \"%s\" {", before.Name))
				
				if before.Type != after.Type {
					if before.Type != "" {
						lines = append(lines, fmt.Sprintf("-  type = \"%s\"", before.Type))
					}
					if after.Type != "" {
						lines = append(lines, fmt.Sprintf("+  type = \"%s\"", after.Type))
					}
				}
				if before.Description != after.Description {
					if before.Description != "" {
						lines = append(lines, fmt.Sprintf("-  description = \"%s\"", before.Description))
					}
					if after.Description != "" {
						lines = append(lines, fmt.Sprintf("+  description = \"%s\"", after.Description))
					}
				}
				if before.DefaultValue != after.DefaultValue {
					if before.DefaultValue != "" {
						lines = append(lines, fmt.Sprintf("-  default = %s", before.DefaultValue))
					}
					if after.DefaultValue != "" {
						lines = append(lines, fmt.Sprintf("+  default = %s", after.DefaultValue))
					}
				}
				
				lines = append(lines, " }")
			}
		}
	}
	
	return lines
}

// compareMapAttributes compares two string maps and returns diff lines
func compareMapAttributes(before, after map[string]string) []string {
	var lines []string
	
	// Get all keys from both maps
	allKeys := make(map[string]bool)
	for key := range before {
		allKeys[key] = true
	}
	for key := range after {
		allKeys[key] = true
	}
	
	// Sort keys for consistent output
	var sortedKeys []string
	for key := range allKeys {
		sortedKeys = append(sortedKeys, key)
	}
	sort.Strings(sortedKeys)
	
	// Compare each key
	for _, key := range sortedKeys {
		beforeVal, beforeExists := before[key]
		afterVal, afterExists := after[key]
		
		// Skip complex expressions and empty values
		if (beforeExists && (beforeVal == "" || beforeVal == "<complex_expression>" || strings.HasPrefix(beforeVal, "<"))) &&
		   (afterExists && (afterVal == "" || afterVal == "<complex_expression>" || strings.HasPrefix(afterVal, "<"))) {
			continue
		}
		
		if !beforeExists && afterExists && afterVal != "" && afterVal != "<complex_expression>" && !strings.HasPrefix(afterVal, "<") {
			// Added attribute
			lines = append(lines, fmt.Sprintf("+  %s = \"%s\"", key, afterVal))
		} else if beforeExists && !afterExists && beforeVal != "" && beforeVal != "<complex_expression>" && !strings.HasPrefix(beforeVal, "<") {
			// Removed attribute
			lines = append(lines, fmt.Sprintf("-  %s = \"%s\"", key, beforeVal))
		} else if beforeExists && afterExists && beforeVal != afterVal {
			// Modified attribute
			if beforeVal != "" && beforeVal != "<complex_expression>" && !strings.HasPrefix(beforeVal, "<") {
				lines = append(lines, fmt.Sprintf("-  %s = \"%s\"", key, beforeVal))
			}
			if afterVal != "" && afterVal != "<complex_expression>" && !strings.HasPrefix(afterVal, "<") {
				lines = append(lines, fmt.Sprintf("+  %s = \"%s\"", key, afterVal))
			}
		}
	}
	
	return lines
}

// groupDiffsByLevel groups diffs by their level
func groupDiffsByLevel(diffs []Diff) map[string][]Diff {
	grouped := make(map[string][]Diff)
	
	for _, diff := range diffs {
		grouped[diff.Level] = append(grouped[diff.Level], diff)
	}
	
	return grouped
}

// formatLevelName converts internal level names to user-friendly names
func formatLevelName(level string) string {
	switch level {
	case "module_call":
		return "📦 Module Calls"
	case "output":
		return "📤 Outputs"
	case "resource":
		return "🏗️  Resources"
	case "data_source":
		return "📊 Data Sources"
	case "variable":
		return "🔧 Variables"
	default:
		return strings.Title(strings.ReplaceAll(level, "_", " "))
	}
}

// getDiffIcon returns an appropriate icon for each diff type
func getDiffIcon(diffType DiffType) string {
	switch diffType {
	case DiffTypeAdded:
		return "➕"
	case DiffTypeRemoved:
		return "➖"
	case DiffTypeModified:
		return "📝"
	default:
		return "❓"
	}
}

// formatModificationDetails provides additional details for modified elements
func formatModificationDetails(diff Diff) string {
	var details strings.Builder
	
	switch diff.Level {
	case "module_call":
		if beforeCall, ok := diff.Before.(ModuleCall); ok {
			if afterCall, ok := diff.After.(ModuleCall); ok {
				details.WriteString(formatModuleCallDetails(beforeCall, afterCall))
			}
		}
	case "output":
		if beforeOutput, ok := diff.Before.(Output); ok {
			if afterOutput, ok := diff.After.(Output); ok {
				details.WriteString(formatOutputDetails(beforeOutput, afterOutput))
			}
		}
	case "resource":
		if beforeResource, ok := diff.Before.(Resource); ok {
			if afterResource, ok := diff.After.(Resource); ok {
				details.WriteString(formatResourceDetails(beforeResource, afterResource))
			}
		}
	case "data_source":
		if beforeDS, ok := diff.Before.(DataSource); ok {
			if afterDS, ok := diff.After.(DataSource); ok {
				details.WriteString(formatDataSourceDetails(beforeDS, afterDS))
			}
		}
	case "variable":
		if beforeVar, ok := diff.Before.(Variable); ok {
			if afterVar, ok := diff.After.(Variable); ok {
				details.WriteString(formatVariableDetails(beforeVar, afterVar))
			}
		}
	}
	
	return details.String()
}

func formatModuleCallDetails(before, after ModuleCall) string {
	var details strings.Builder
	
	if before.Source != after.Source {
		details.WriteString(fmt.Sprintf("      Source: %s → %s\n", before.Source, after.Source))
	}
	if before.Version != after.Version {
		details.WriteString(fmt.Sprintf("      Version: %s → %s\n", before.Version, after.Version))
	}
	
	return details.String()
}

func formatOutputDetails(before, after Output) string {
	var details strings.Builder
	
	if before.Description != after.Description {
		details.WriteString(fmt.Sprintf("      Description: %s → %s\n", before.Description, after.Description))
	}
	if before.Sensitive != after.Sensitive {
		details.WriteString(fmt.Sprintf("      Sensitive: %t → %t\n", before.Sensitive, after.Sensitive))
	}
	
	return details.String()
}

func formatResourceDetails(before, after Resource) string {
	var details strings.Builder
	
	if before.Type != after.Type {
		details.WriteString(fmt.Sprintf("      Type: %s → %s\n", before.Type, after.Type))
	}
	
	return details.String()
}

func formatDataSourceDetails(before, after DataSource) string {
	var details strings.Builder
	
	if before.Type != after.Type {
		details.WriteString(fmt.Sprintf("      Type: %s → %s\n", before.Type, after.Type))
	}
	
	return details.String()
}

func formatVariableDetails(before, after Variable) string {
	var details strings.Builder
	
	if before.Type != after.Type {
		details.WriteString(fmt.Sprintf("      Type: %s → %s\n", before.Type, after.Type))
	}
	if before.Description != after.Description {
		details.WriteString(fmt.Sprintf("      Description: %s → %s\n", before.Description, after.Description))
	}
	if before.DefaultValue != after.DefaultValue {
		details.WriteString(fmt.Sprintf("      Default: %s → %s\n", before.DefaultValue, after.DefaultValue))
	}
	
	return details.String()
}