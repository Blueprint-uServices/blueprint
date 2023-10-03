package blueprint

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"unicode"

	"golang.org/x/mod/modfile"
)

func Indent(str string, amount int) string {
	lines := strings.Split(str, "\n")
	for i, line := range lines {
		lines[i] = strings.Repeat(" ", amount) + line
	}
	return strings.Join(lines, "\n")
}

func Capitalize(s string) string {
	r := []rune(s)
	return string(append([]rune{unicode.ToUpper(r[0])}, r[1:]...))
}

type SourceFileInfo struct {
	Filename          string // Local filename
	Module            string // Fully qualified module name
	ModulePath        string // path to module on disk
	ModuleFilename    string // Filename within module
	WorkspaceFilename string // Filename within workspace, if the module is in a workspace; otherwise ModuleFilename
}

func (info *SourceFileInfo) String() string {
	return info.ModuleFilename
}

var fileInfoCache = make(map[string]*SourceFileInfo)

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

func GetSourceFileInfo(fileName string) *SourceFileInfo {
	if info, exists := fileInfoCache[fileName]; exists {
		return info
	}

	// Start constructing the file info
	dir, _ := filepath.Split(fileName)
	info := &SourceFileInfo{
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

type WiringCallsite struct {
	Source     *SourceFileInfo
	LineNumber int
	Func       string
	FuncName   string
}

type WiringCallstack struct {
	Stack []WiringCallsite
}

func (cs WiringCallsite) String() string {
	return fmt.Sprintf("%s:%v %s", cs.Source.ModuleFilename, cs.LineNumber, cs.Func)
}

func (stack *WiringCallstack) String() string {
	var s []string
	for _, callsite := range stack.Stack {
		s = append(s, callsite.String())
	}
	return strings.Join(s, "\n")
}

func getWiringCallsite() *WiringCallstack {
	pc := make([]uintptr, 10)
	n := runtime.Callers(3, pc)
	if n == 0 {
		return nil
	}

	frames := runtime.CallersFrames(pc[:n-2])
	callstack := &WiringCallstack{}
	for {
		frame, more := frames.Next()

		splits := strings.Split(frame.Function, "/")
		callsite := WiringCallsite{
			Source:     GetSourceFileInfo(frame.File),
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
