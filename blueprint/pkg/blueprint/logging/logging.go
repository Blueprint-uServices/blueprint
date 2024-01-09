// Package logging implements a custom slog logger for Blueprint.
//
// The custom logger adds additional callsite information to logging statements,
// to provide more information during the compilation process about which
// plugins are producing logs or errors, and to tie that information back
// to the corresponding wiring line.
package logging

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"golang.org/x/exp/slog"
	"golang.org/x/mod/modfile"
)

// Implementation of Blueprint's custom Handler of slog.Logger
type blueprintLoggerHandler struct {
	slog.Handler
	l       *log.Logger
	enabled bool
}

type blueprintLoggerHandlerOptions struct {
	SlogOpts slog.HandlerOptions
}

var ignore = make(map[string]struct{})

func init() {
	funcNames := []string{
		"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring.(*namespaceimpl).Info",
		"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring.(*namespaceimpl).Warn",
		"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring.(*namespaceimpl).Error",
		"golang.org/x/exp/slog.Info",
		"golang.org/x/exp/slog.(*Logger).log",
	}
	for _, funcName := range funcNames {
		ignore[funcName] = struct{}{}
	}
}

// Implementation of a slog logger
func (h *blueprintLoggerHandler) Handle(ctx context.Context, r slog.Record) error {
	if !h.enabled {
		return nil
	}
	level := r.Level.String() + ":"

	timeStr := r.Time.Format("[15:04:05.000]")

	fields := make(map[string]interface{}, r.NumAttrs())
	r.Attrs(func(a slog.Attr) bool {
		fields[a.Key] = a.Value.Any()

		return true
	})

	b, err := json.MarshalIndent(fields, "", "  ")
	if err != nil {
		return err
	}

	// fs := runtime.CallersFrames([]uintptr{r.PC})
	// f, _ := fs.Next()
	// info := getSourceFileInfo(f.File)
	// if info.WorkspaceFilename == filepath.Join("blueprint", "pkg", "wiring", "namespace.go") {
	// 	f, _ = fs.Next()
	// }
	// source_str := fmt.Sprintf("[%v:%v]", info.WorkspaceFilename, f.Line)
	cs := GetCallstack()
	frameNumber := 0
	for ; frameNumber < len(cs.Stack)-1; frameNumber++ {
		if _, ignoreFunc := ignore[cs.Stack[frameNumber].FuncName]; !ignoreFunc {
			break
		}
	}

	f := cs.Stack[frameNumber]
	source_str := fmt.Sprintf("[%v:%v]", f.Source.WorkspaceFilename, f.LineNumber)

	if len(fields) != 0 {
		h.l.Println(timeStr, source_str, level, r.Message, string(b))
	} else {
		h.l.Println(timeStr, source_str, level, r.Message)
	}

	return nil
}

func newBlueprintLoggerHandler(out io.Writer, opts blueprintLoggerHandlerOptions) *blueprintLoggerHandler {
	h := &blueprintLoggerHandler{
		Handler: slog.NewTextHandler(out, &opts.SlogOpts),
		l:       log.New(out, "", 0),
		enabled: true,
	}

	return h
}

var loggerhandler *blueprintLoggerHandler

// Initializes the logger when this package is first loaded. This function is guaranteed to be invoked only once so the logger will be initialized only once.
func init() {

	opts := slog.HandlerOptions{
		AddSource: true,
	}
	blOpts := blueprintLoggerHandlerOptions{SlogOpts: opts}

	loggerhandler = newBlueprintLoggerHandler(os.Stdout, blOpts)
	logger := slog.New(loggerhandler)
	slog.SetDefault(logger)
}

// Compiler logging is enabled by default; this method is useful for tests to disable and enable logging
// in order to suppress output.
func EnableCompilerLogging() {
	if loggerhandler != nil {
		loggerhandler.enabled = true
	}
}

// Disables logging by the compiler; useful when running tests to suppress verbose output.
func DisableCompilerLogging() {
	if loggerhandler != nil {
		loggerhandler.enabled = false
	}
}

type sourceFileInfo struct {
	Filename          string // Local filename
	Module            string // Fully qualified module name
	ModulePath        string // path to module on disk
	ModuleFilename    string // Filename within module
	WorkspaceFilename string // Filename within workspace, if the module is in a workspace; otherwise ModuleFilename
}

func (info *sourceFileInfo) String() string {
	return info.ModuleFilename
}

var fileInfoCache = make(map[string]*sourceFileInfo)

/*
Starting from the specified subdirectory, recurses through parent
directories until finding a file with the specified name.
Returns the parent directory containing the file.  If empty
string is returned, then file is not found
*/
func findFileInParentDirectory(dir string, fileName string) string {
	dir = filepath.Clean(dir)
	for dir != "" && dir[len(dir)-1] != filepath.Separator {
		if _, err := os.Stat(filepath.Join(dir, fileName)); err == nil {
			return filepath.Clean(dir)
		}
		dir, _ = filepath.Split(filepath.Clean(dir))
		dir = filepath.Clean(dir)
	}
	return ""
}

func getSourceFileInfo(fileName string) *sourceFileInfo {
	if info, exists := fileInfoCache[fileName]; exists {
		return info
	}

	// Start constructing the file info
	dir, _ := filepath.Split(fileName)
	info := &sourceFileInfo{
		Filename:          fileName,
		Module:            "",
		ModulePath:        dir,
		ModuleFilename:    fileName,
		WorkspaceFilename: fileName,
	}
	fileInfoCache[fileName] = info

	if fileName == "" {
		return info
	}

	// Find the module directory
	modDir := findFileInParentDirectory(dir, "go.mod")
	if modDir == "" {
		// File is not within a module; return default info
		return info
	}

	modfileName := filepath.Join(modDir, "go.mod")
	modfileData, err := os.ReadFile(modfileName)
	if err != nil {
		// Invalid modfile; return default info
		return info
	}

	modfile, err := modfile.Parse(modfileName, modfileData, nil)
	if err != nil {
		// Invalid modfile; return default info
		return info
	}

	// Fill in the module info
	relFileName, _ := filepath.Rel(modDir, fileName)
	info.Module = modfile.Module.Mod.Path
	info.ModuleFilename = relFileName
	info.WorkspaceFilename = relFileName

	// Find the workspace dir
	workDir := findFileInParentDirectory(modDir, "go.work")
	if workDir == "" {
		// File is not within a workspace; return module info
		return info
	}

	// Don't bother validating the go.work file
	relFileName, _ = filepath.Rel(workDir, fileName)
	info.WorkspaceFilename = relFileName

	return info
}

// Used to tie logging statements and errors back to the wiring file line that caused the error
type Callsite struct {
	Source     *sourceFileInfo
	LineNumber int
	Func       string
	FuncName   string
}

// Used to tie logging statements and errors back to the wiring file line that caused the error
type Callstack struct {
	Stack []Callsite
}

func (cs Callsite) String() string {
	return fmt.Sprintf("%s:%v %s", cs.Source.ModuleFilename, cs.LineNumber, cs.Func)
}

func (stack *Callstack) String() string {
	var s []string
	for _, callsite := range stack.Stack {
		s = append(s, callsite.String())
	}
	return strings.Join(s, "\n")
}

// Gets the current callstack including file information.
// Blueprint's wiring spec uses this so that logging statements and error messages
// can be attributed back to the appropriate wiring spec line.
func GetCallstack() *Callstack {
	pc := make([]uintptr, 10)
	n := runtime.Callers(3, pc)
	if n == 0 {
		return nil
	}

	frames := runtime.CallersFrames(pc[:n-2])
	callstack := &Callstack{}
	for {
		frame, more := frames.Next()

		splits := strings.Split(frame.Function, "/")
		callsite := Callsite{
			Source:     getSourceFileInfo(frame.File),
			LineNumber: frame.Line,
			Func:       splits[len(splits)-1],
			FuncName:   frame.Function,
		}

		if callsite.Source == nil {
			break
		}
		callstack.Stack = append(callstack.Stack, callsite)
		if !more {
			break
		}
	}
	return callstack
}
