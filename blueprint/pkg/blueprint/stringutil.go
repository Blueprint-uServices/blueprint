package blueprint

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"golang.org/x/mod/modfile"
)

func Indent(str string, amount int) string {
	lines := strings.Split(str, "\n")
	for i, line := range lines {
		lines[i] = strings.Repeat(" ", amount) + line
	}
	return strings.Join(lines, "\n")
}

type SourceFileInfo struct {
	Filename       string // Local filename
	Module         string // Fully qualified module name
	ModulePath     string // path to module on disk
	ModuleFilename string // Filename within module
}

func (info *SourceFileInfo) String() string {
	return info.ModuleFilename
}

var fileInfoCache = make(map[string]*SourceFileInfo)

func GetSourceFileInfo(fileName string) *SourceFileInfo {
	if info, exists := fileInfoCache[fileName]; exists {
		return info
	}

	dir, _ := filepath.Split(fileName)
	for dir != "" {
		modFileName := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(modFileName); err == nil {
			modFileData, err := os.ReadFile(modFileName)
			if err == nil {
				f, err := modfile.Parse(modFileName, modFileData, nil)
				if err == nil {
					relFileName, _ := filepath.Rel(dir, fileName)
					info := &SourceFileInfo{
						Filename:       fileName,
						Module:         f.Module.Mod.Path,
						ModulePath:     dir,
						ModuleFilename: relFileName,
					}

					fileInfoCache[fileName] = info
					return info
				}
			}
		}
		dir, _ = filepath.Split(filepath.Clean(dir))
	}
	return nil
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
