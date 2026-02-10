package tfdiff

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

type App struct {
	CLI *CLI
}

func New(cli *CLI) *App {
	return &App{
		CLI: cli,
	}
}

func (app *App) Run(ctx context.Context) error {
	cli := app.CLI

	// Validate directories
	if err := ValidateModuleDirectory(cli.LeftDir); err != nil {
		return fmt.Errorf("left directory validation failed: %w", err)
	}

	if err := ValidateModuleDirectory(cli.RightDir); err != nil {
		return fmt.Errorf("right directory validation failed: %w", err)
	}

	// Parse modules
	parseOptions := ParseOptions{
		IgnoreFiles: cli.IgnoreFiles,
	}
	leftModule, err := ParseModuleWithOptions(cli.LeftDir, parseOptions)
	if err != nil {
		return fmt.Errorf("failed to parse left module: %w", err)
	}

	rightModule, err := ParseModuleWithOptions(cli.RightDir, parseOptions)
	if err != nil {
		return fmt.Errorf("failed to parse right module: %w", err)
	}

	// Build comparison config
	config := ComparisonConfig{
		Levels:          parseComparisonLevels(cli.Levels),
		IgnoreArguments: cli.IgnoreArgs,
	}

	// Compare modules
	result := CompareModules(leftModule, rightModule, config)

	// Sort diffs for consistent output
	SortDiffs(result.Diffs)

	// Output results
	switch cli.OutputFormat {
	case "json":
		return app.outputJSON(result)
	case "text":
		return app.outputText(result, config)
	default:
		return fmt.Errorf("unsupported output format: %s", cli.OutputFormat)
	}
}

func parseComparisonLevels(levels []string) []ComparisonLevel {
	var result []ComparisonLevel

	for _, level := range levels {
		switch strings.TrimSpace(level) {
		case "module_calls":
			result = append(result, ComparisonLevelModuleCalls)
		case "outputs":
			result = append(result, ComparisonLevelOutputs)
		case "resources":
			result = append(result, ComparisonLevelResources)
		case "data_sources":
			result = append(result, ComparisonLevelDataSources)
		case "variables":
			result = append(result, ComparisonLevelVariables)
		case "all":
			result = append(result, ComparisonLevelAll)
		}
	}

	return result
}

func (app *App) outputJSON(result *ComparisonResult) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

func (app *App) outputText(result *ComparisonResult, config ComparisonConfig) error {
	output := FormatTextOutput(result, config, app.CLI.NoColor)
	fmt.Print(output)
	return nil
}
