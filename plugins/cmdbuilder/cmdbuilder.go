// Package cmdbuilder is a helper package for building wiring spec command line programs.
// It doesn't provide any wiring spec commands or IR.
//
// The [CmdBuilder] struct enables an application to register multiple wiring spec options
// in a single main.go.  It adds command line arguments for selecting which wiring spec to compile,
// and takes care of argument parsing and spec building.
//
// Specify the name of a wiring spec with the -w argument, and the output directory with -o.
//
// # Usage
//
// Define one or more wiring specs.  Each wiring spec should be implemented inside a function
// with the following signature:
//
//	func (spec wiring.WiringSpec) ([]string, error)
//
// The function should behave like a typical wiring spec: instantiating workflow services, deploying
// them inside processes or containers, etc.
//
// Next, define a [SpecOption]:
//
//	func buildMySpec(spec wiring.WiringSpec) ([]string, error) {
//		... // does wiring spec stuff
//	}
//
//	var MySpec = cmdbuilder.SpecOption{
//		Name:        "myspec",
//		Description: "My example wiring spec",
//		Build:       buildMySpec,
//	}
//
// Lastly, in the main file, call [MakeAndExecute]:
//
//	cmdbuilder.MakeAndExecute(
//		"MyApplication",
//		MySpec
//	)
//
// [MakeAndExecute] accepts any number of specs.  cmdbuilder takes care of parsing command line args
//
// # Runtime Usage
//
// Run your program with
//
//	go run main.go -h
//
// The cmdbuilder plugin takes care of argument parsing, and will list the wiring specs that can be compiled.
//
// To compile a spec, run
//
//	go run main.go -o build -w myspec
//
// [wiring/main.go]: https://github.com/Blueprint-uServices/blueprint/blob/main/examples/sockshop/wiring/main.go
package cmdbuilder

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint/logging"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
	"github.com/blueprint-uservices/blueprint/plugins/dockercompose"
	"github.com/blueprint-uservices/blueprint/plugins/environment"
	"github.com/blueprint-uservices/blueprint/plugins/goproc"
	"github.com/blueprint-uservices/blueprint/plugins/linuxcontainer"
	"golang.org/x/exp/slog"
)

// A wiring spec option used by [CmdBuilder].  When running the program,
// this wiring spec can be selected by specifying its [Name] with the -w flag,
// e.g.
//
//	-w {{.Name}}
type SpecOption struct {
	Name        string
	Description string
	Build       func(wiring.WiringSpec) ([]string, error)
}

// A helper struct when a Blueprint application supports multiple different
// wiring specs.  Makes it easy to choose which spec to compile.
// See the Blueprint example applications for usage
type CmdBuilder struct {
	Name      string
	OutputDir string
	Quiet     bool
	SpecName  string
	Env       bool
	Port      uint16
	Spec      SpecOption
	Wiring    wiring.WiringSpec
	IR        *ir.ApplicationNode

	Registry map[string]SpecOption
}

// Parses command line flags, and if a valid spec is specified with the -w
// flag, that exists within specs, executes that spec.
func MakeAndExecute(name string, specs ...SpecOption) {
	builder := NewCmdBuilder(name)
	builder.Add(specs...)

	builder.ParseArgs()
	if err := builder.ValidateArgs(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	if err := builder.Build(); err != nil {
		slog.Error(err.Error())
		os.Exit(2)
	}
}

func NewCmdBuilder(applicationName string) *CmdBuilder {
	builder := CmdBuilder{}
	builder.Name = applicationName
	builder.Registry = make(map[string]SpecOption)
	return &builder
}

func (b *CmdBuilder) Add(specs ...SpecOption) {
	for _, spec := range specs {
		b.Registry[spec.Name] = spec
	}
}

func (b *CmdBuilder) ParseArgs() {
	output_dir := flag.String("o", "", "Target output directory for compilation.")
	spec_name := flag.String("w", "", "Wiring spec to compile.  One of:\n"+b.List())
	quiet := flag.Bool("quiet", false, "Suppress verbose compiler output.")
	env := flag.Bool("env", true, "Generate a .env file that sets service address and port environment variables")
	port := flag.Uint("port", 12345, "Sets the port to start at when assigning service ports.  Only used when generating a .env file.")

	flag.Parse()

	b.OutputDir = *output_dir
	b.Quiet = *quiet
	b.SpecName = *spec_name
	b.Env = *env
	b.Port = uint16(*port)
}

func (b *CmdBuilder) ValidateArgs() error {
	if b.OutputDir == "" {
		return fmt.Errorf("output directory not specified, specify with -o")
	}

	if b.SpecName == "" {
		return fmt.Errorf("wiring spec not specified, specify with -w")
	}

	if spec, specExists := b.Registry[b.SpecName]; specExists {
		b.Spec = spec
	} else {
		return fmt.Errorf("unknown wiring spec \"%v\", expected one of:\n%v", b.SpecName, b.List())
	}

	if b.Quiet {
		slog.Info("Suppressing compiler logging")
		logging.DisableCompilerLogging()
	}

	return nil
}

// Returns a list of configured wiring specs
func (builder *CmdBuilder) List() string {
	var b strings.Builder
	for _, spec := range builder.Registry {
		b.WriteString(fmt.Sprintf("  %v: %v\n", spec.Name, spec.Description))
	}
	return b.String()
}

func (b *CmdBuilder) Build() error {
	// Configure the default builders for Blueprint
	slog.Info("Initializing Blueprint compiler")
	goproc.RegisterAsDefaultBuilder()
	linuxcontainer.RegisterAsDefaultBuilder()
	dockercompose.RegisterAsDefaultBuilder()
	if b.Env {
		environment.AssignPorts(b.Port)
	}

	// Define the wiring spec
	slog.Info(fmt.Sprintf("Building %v-%v to %v", b.Name, b.SpecName, b.OutputDir))
	b.Wiring = wiring.NewWiringSpec(b.Name)
	nodesToBuild, err := b.Spec.Build(b.Wiring)
	if err != nil {
		return fmt.Errorf("unable to build %v-%v wiring due to %v", b.Name, b.SpecName, err.Error())
	}
	slog.Info(fmt.Sprintf("Constructed %v WiringSpec %v: \n%v", b.Name, b.SpecName, b.Wiring))

	// Construct the IR
	b.IR, err = b.Wiring.BuildIR(nodesToBuild...)
	slog.Info(fmt.Sprintf("%v %v IR: \n%v", b.Name, b.SpecName, b.IR))
	if err != nil {
		return fmt.Errorf("unable to construct %v-%v IR due to %v", b.Name, b.SpecName, err.Error())
	}

	// Generate artifacts
	slog.Info(fmt.Sprintf("Generating %v-%v artifacts to %v", b.Name, b.SpecName, b.OutputDir))
	err = b.IR.GenerateArtifacts(b.OutputDir)
	if err != nil {
		return fmt.Errorf("unable to generate %v-%v artifacts due to %v", b.Name, b.SpecName, err.Error())
	}

	slog.Info(fmt.Sprintf("Successfully generated %v-%v to %v", b.Name, b.SpecName, b.OutputDir))
	return nil
}
