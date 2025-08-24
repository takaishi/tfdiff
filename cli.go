package tfdiff

import (
	"context"
	"fmt"

	"github.com/alecthomas/kong"
)

var Version = "dev"
var Revision = "HEAD"

type GlobalOptions struct {
}

type CLI struct {
	Version       VersionFlag  `name:"version" help:"show version"`
	LeftDir       string       `arg:"" name:"left" help:"path to left Terraform module directory"`
	RightDir      string       `arg:"" name:"right" help:"path to right Terraform module directory"`
	Levels        []string     `short:"l" name:"level" help:"comparison levels: module_calls, outputs, resources, data_sources, variables, all" default:"module_calls,outputs,resources,data_sources"`
	IgnoreArgs    bool         `name:"ignore-args" help:"ignore argument differences" default:"false"`
	OutputFormat  string       `short:"o" name:"output" help:"output format: text, json" default:"text"`
	NoColor       bool         `name:"no-color" help:"disable colored output"`
}

type VersionFlag string

func (v VersionFlag) Decode(ctx *kong.DecodeContext) error { return nil }
func (v VersionFlag) IsBool() bool                         { return true }
func (v VersionFlag) BeforeApply(app *kong.Kong, vars kong.Vars) error {
	fmt.Printf("%s-%s\n", Version, Revision)
	app.Exit(0)
	return nil
}

func RunCLI(ctx context.Context, args []string) error {
	cli := CLI{
		Version: VersionFlag("0.1.0"),
	}
	parser, err := kong.New(&cli)
	if err != nil {
		return fmt.Errorf("error creating CLI parser: %w", err)
	}
	_, err = parser.Parse(args)
	if err != nil {
		fmt.Printf("error parsing CLI: %v\n", err)
		return fmt.Errorf("error parsing CLI: %w", err)
	}
	app := New(&cli)
	return app.Run(ctx)
}
