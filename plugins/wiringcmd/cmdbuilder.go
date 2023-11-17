// Package wiringbuilder doesn't provide any blueprint IR or wiring spec extensions.
//
// It is a helper package for building wiring spec command line programs.
// The Blueprint example applications use the cmdbuilder
package wiringcmd

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint/logging"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/wiring"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/dockerdeployment"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/goproc"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/linuxcontainer"
	"golang.org/x/exp/slog"
)

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
	Spec      SpecOption
	Wiring    wiring.WiringSpec
	IR        *ir.ApplicationNode

	Registry map[string]SpecOption
}

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

	flag.Parse()

	b.OutputDir = *output_dir
	b.Quiet = *quiet
	b.SpecName = *spec_name
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
	dockerdeployment.RegisterAsDefaultBuilder()

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
