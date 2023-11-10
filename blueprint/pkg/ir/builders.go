package ir

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"golang.org/x/exp/slog"
)

/*
Plugins can register their ability to build different nodes and namespaces here

Compiling a Blueprint application will use the builders registered here
*/

type (
	registry struct {
		node      map[reflect.Type]*nodeBuilder
		namespace map[reflect.Type]*namespaceBuilder
	}

	builder struct {
		name     string
		nodeType reflect.Type
	}

	/*
		For a plugin that can combine nodes of the same type into a single namespace
		(e.g. goproc combines go nodes into a goproc namespace)
	*/
	namespaceBuilder struct {
		builder
		build func(outputDir string, nodes []IRNode) error
	}

	/*
		For a plugin that can build individual nodes of a given type
	*/
	nodeBuilder struct {
		builder
		build func(outputDir string, node IRNode) error
	}
)

/*
If the root Blueprint application contains nodes of type T, this function enables
plugins to register a default namespace to combine and build those nodes.
*/
func RegisterDefaultNamespace[T IRNode](name string, buildFunc func(outputDir string, nodes []IRNode) error) {
	nodeType := reflect.TypeOf(new(T)).Elem()
	defaultBuilders.AddNamespaceBuilder(name, nodeType, buildFunc)
	slog.Info(fmt.Sprintf("%v registered as the default namespace builder for %v nodes", name, nodeType))
}

/*
If the root Blueprint application contains nodes of type T, this function enables
plugins to register a default builder to individually build those nodes
*/
func RegisterDefaultBuilder[T IRNode](name string, buildFunc func(outputDir string, node IRNode) error) {
	nodeType := reflect.TypeOf(new(T)).Elem()
	defaultBuilders.AddNodeBuilder(name, nodeType, buildFunc)
	slog.Info(fmt.Sprintf("%v registered as the default node builder for %v nodes", name, nodeType))
}

var defaultBuilders = registry{
	node:      make(map[reflect.Type]*nodeBuilder),
	namespace: make(map[reflect.Type]*namespaceBuilder),
}

func (r *registry) AddNamespaceBuilder(name string, nodeType reflect.Type, buildFunc func(outputDir string, nodes []IRNode) error) {
	r.namespace[nodeType] = &namespaceBuilder{
		builder: builder{
			name:     name,
			nodeType: nodeType,
		},
		build: buildFunc,
	}
}

func (r *registry) AddNodeBuilder(name string, nodeType reflect.Type, buildFunc func(outputDir string, nodes IRNode) error) {
	r.node[nodeType] = &nodeBuilder{
		builder: builder{
			name:     name,
			nodeType: nodeType,
		},
		build: buildFunc,
	}
}

/* True if this builder can build the specified node type; false otherwise */
func (b *builder) Builds(node IRNode) bool {
	return reflect.TypeOf(node).AssignableTo(b.nodeType)
}

func (b *namespaceBuilder) BuildCompatibleNodes(outputDir string, nodes []IRNode) ([]IRNode, error) {
	// Find compatible nodes
	toBuild := make([]IRNode, 0, len(nodes))
	remaining := make([]IRNode, 0, len(nodes))
	for _, node := range nodes {
		if b.Builds(node) {
			toBuild = append(toBuild, node)
		} else {
			remaining = append(remaining, node)
		}
	}

	// Build them
	if len(toBuild) > 0 {
		if err := b.build(outputDir, toBuild); err != nil {
			return nil, err
		}
	}
	return remaining, nil
}

func (b *nodeBuilder) BuildCompatibleNodes(outputDir string, nodes []IRNode) ([]IRNode, error) {
	remaining := make([]IRNode, 0, len(nodes))
	for _, node := range nodes {
		if b.Builds(node) {
			if err := b.build(outputDir, node); err != nil {
				return nil, err
			}
		} else {
			remaining = append(remaining, node)
		}
	}
	return remaining, nil
}

func (r *registry) BuildAll(outputDir string, nodes []IRNode) error {
	// Create output directory
	if info, err := os.Stat(outputDir); err == nil && info.IsDir() {
		return blueprint.Errorf("output directory %v already exists", outputDir)
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return blueprint.Errorf("unable to create output directory %v due to %v", outputDir, err.Error())
	}

	// Exclude metadata nodes and config nodes (for now)
	nodes = Remove[IRMetadata](nodes)
	nodes = Remove[IRConfig](nodes)

	// Build namespaces first
	for _, builder := range r.namespace {
		var err error
		nodes, err = builder.BuildCompatibleNodes(outputDir, nodes)
		if err != nil {
			return err
		}
	}

	// Build individual nodes
	for _, builder := range r.node {
		var err error
		nodes, err = builder.BuildCompatibleNodes(outputDir, nodes)
		if err != nil {
			return err
		}
	}

	if len(nodes) > 0 {
		unbuiltTypes := make(map[reflect.Type]struct{})
		for _, node := range nodes {
			unbuiltTypes[reflect.TypeOf(node)] = struct{}{}
		}
		typeNames := []string{}
		for t := range unbuiltTypes {
			typeNames = append(typeNames, t.String())
		}
		return blueprint.Errorf("No registered builders for node types %s", strings.Join(typeNames, ", "))
	}
	return nil
}
