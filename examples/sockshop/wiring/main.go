// Package main provides an application for compiling a number of different
// wiring specs for the SockShop application.
//
// Run with go run examples/sockshop/wiring/main.go
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint/logging"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/wiring"
	"gitlab.mpi-sws.org/cld/blueprint/examples/sockshop/wiring/specs"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/dockerdeployment"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/goproc"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/linuxcontainer"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/workflow"
	"golang.org/x/exp/slices"
	"golang.org/x/exp/slog"
)

func parseArgs() (outdir, spec string, err error) {
	supportedSpecs := []string{
		"basic",
	}

	output_dir := flag.String("o", "", "Target output directory for compilation.")
	spec_name := flag.String("w", "", "Wiring spec to compile.  One of "+strings.Join(supportedSpecs, ", "))
	quiet := flag.Bool("quiet", false, "Suppress verbose compiler output.")

	flag.Parse()

	outdir = *output_dir
	spec = *spec_name

	if outdir == "" {
		err = fmt.Errorf("output directory not specified, specify with -o")
		return
	}

	if spec == "" {
		err = fmt.Errorf("wiring spec not specified, specify with -w")
		return
	}
	if !slices.Contains(supportedSpecs, spec) {
		err = fmt.Errorf("unknown wiring spec \"%v\", expected one of [%v]", spec, strings.Join(supportedSpecs, ", "))
		return
	}

	if *quiet {
		slog.Info("Suppressing compiler logging")
		logging.DisableCompilerLogging()
	}

	return
}

func getSpec(specname string) (spec wiring.WiringSpec, nodesToBuild []string, err error) {
	spec = wiring.NewWiringSpec("SockShop")

	switch specname {
	case "basic":
		nodesToBuild, err = specs.BasicWiringSpec(spec)
	default:
		err = fmt.Errorf("unknown wiring spec %v", specname)
	}

	return
}

func initializeBlueprint() {
	// Configure the default builders for Blueprint
	goproc.RegisterAsDefaultBuilder()
	linuxcontainer.RegisterAsDefaultBuilder()
	dockerdeployment.RegisterAsDefaultBuilder()

	// Point the workflow plugin to the location of our workflow spec
	workflow.Init("../workflow")
}

func main() {
	// Parse command line args
	outdir, specname, err := parseArgs()
	if err != nil {
		fmt.Printf("Error: %v\n", err.Error())
		os.Exit(1)
	}
	slog.Info(fmt.Sprintf("Compiling SockShop \"%v\" to %v", specname, outdir))

	// Initialize Blueprint
	slog.Info("Initializing Blueprint compiler")
	initializeBlueprint()

	// Define the wiring spec
	spec, nodesToBuild, err := getSpec(specname)
	if err != nil {
		slog.Error(fmt.Sprintf("Unable to build %v due to %v", specname, err.Error()))
		os.Exit(2)
	}
	slog.Info("Constructed SockShop WiringSpec: \n" + spec.String())

	// Build the IR
	ir, err := spec.BuildIR(nodesToBuild...)
	slog.Info("SockShop IR: \n" + ir.String())
	if err != nil {
		slog.Error("Error building application IR: " + err.Error())
		os.Exit(3)
	}

	// Compile the IR
	slog.Info(fmt.Sprintf("Generating artifacts to %v", outdir))
	err = ir.GenerateArtifacts(outdir)
	if err != nil {
		slog.Error("Error generating artifacts: " + err.Error())
		os.Exit(4)
	}

	slog.Info(fmt.Sprintf("Successfully compiled SockShop %v to %v", specname, outdir))
}
