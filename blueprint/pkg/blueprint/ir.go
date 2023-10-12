package blueprint

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"golang.org/x/exp/slog"
)

// This file contains some of the core IR nodes for Blueprint

// The base IRNode type
type IRNode interface {
	Name() string
	String() string
	// TODO: all nodes should have a Build() function
}

type IRMetadata interface {
	ImplementsIRMetadata()
}

// The IR Node that represents the whole application
type ApplicationNode struct {
	IRNode

	name     string
	Children []IRNode
}

func (node *ApplicationNode) Name() string {
	return node.name
}

// Print the IR graph
func (node *ApplicationNode) String() string {
	var b strings.Builder
	b.WriteString(node.name)
	b.WriteString(" = BlueprintApplication() {\n")
	var children []string
	for _, node := range node.Children {
		children = append(children, node.String())
	}
	b.WriteString(Indent(strings.Join(children, "\n"), 2))
	b.WriteString("\n}")
	return b.String()
}

var defaultBuilders = make(map[reflect.Type]func(string, []IRNode) error)

/*
If the root Blueprint application contains nodes of type T, this function enables
plugins to register a default builder to build those nodes.
*/
func RegisterDefaultBuilder[T IRNode](buildFunc func(outputDir string, nodes []IRNode) error) {
	var t T
	key := reflect.TypeOf(t)
	defaultBuilders[key] = buildFunc
}

func (app *ApplicationNode) Compile(outputDir string) error {
	// Create output directory
	if info, err := os.Stat(outputDir); err == nil && info.IsDir() {
		return Errorf("output directory %v already exists", outputDir)
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return Errorf("unable to create output directory %v due to %v", outputDir, err.Error())
	}

	// Group the nodes that we must build, according to their type
	nodesToBuild := make(map[reflect.Type][]IRNode)
	for _, node := range app.Children {
		switch node.(type) {
		case IRMetadata: // skip metadata nodes
			continue
		default:
			key := reflect.TypeOf(node)
			nodesToBuild[key] = append(nodesToBuild[key], node)
		}
	}

	// Invoke the default builders for each node type
	for t, nodes := range nodesToBuild {
		if buildFunc, hasBuilder := defaultBuilders[t]; hasBuilder {
			err := buildFunc(outputDir, nodes)
			if err != nil {
				return Errorf("encountered error while building to %v, aborting: %v", outputDir, err)
			}
		} else {
			slog.Info(fmt.Sprintf("No default builder registered for type %v", t))
		}
	}
	return nil
}
