package ir

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint/ioutil"
	"golang.org/x/exp/slog"
)

type (
	// Registers default build functions for building nodes of certain types.
	registry struct {
		namespace map[reflect.Type]*namespaceBuilder
	}

	builder struct {
		name     string
		nodeType reflect.Type
	}

	namespaceBuilder struct {
		builder
		build func(outputDir string, nodes []IRNode) error
	}
)

var defaultBuilders = registry{
	namespace: make(map[reflect.Type]*namespaceBuilder),
}

func init() {
	// When building an application we don't need to do anything to IRConfig or IRMetadata nodes by default.
	// This can, however, be overridden by plugins by calling [RegisterDefaultNamespace]
	RegisterDefaultNamespace[IRConfig]("ignore_irconfig", func(string, []IRNode) error { return nil })
	RegisterDefaultNamespace[IRMetadata]("ignore_irmetadata", func(string, []IRNode) error { return nil })
}

// When building an application, any IR nodes of type T that reside within the top-level
// application will be built using the specified buildFunc.
func RegisterDefaultNamespace[T IRNode](name string, buildFunc func(outputDir string, nodes []IRNode) error) {
	nodeType := reflect.TypeOf(new(T)).Elem()
	defaultBuilders.addNamespaceBuilder(name, nodeType, buildFunc)
	slog.Info(fmt.Sprintf("%v registered as the default namespace builder for %v nodes", name, nodeType))
}

func (r *registry) addNamespaceBuilder(name string, nodeType reflect.Type, buildFunc func(outputDir string, nodes []IRNode) error) {
	r.namespace[nodeType] = &namespaceBuilder{
		builder: builder{
			name:     name,
			nodeType: nodeType,
		},
		build: buildFunc,
	}
}

/* True if this builder can build the specified node type; false otherwise */
func (b *builder) builds(node IRNode) bool {
	return reflect.TypeOf(node).AssignableTo(b.nodeType)
}

func (b *namespaceBuilder) buildCompatibleNodes(outputDir string, nodes []IRNode) ([]IRNode, error) {
	// Find compatible nodes
	toBuild := make([]IRNode, 0, len(nodes))
	remaining := make([]IRNode, 0, len(nodes))
	for _, node := range nodes {
		if b.builds(node) {
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

func buildArtifactGeneratorNodes(outputdir string, nodes []IRNode) ([]IRNode, error) {
	remaining := make([]IRNode, 0, len(nodes))
	for _, node := range nodes {
		if gen, isGen := node.(ArtifactGenerator); isGen {
			subdir, err := ioutil.CreateNodeDir(outputdir, node.Name())
			if err != nil {
				return nil, err
			}

			if err := gen.GenerateArtifacts(subdir); err != nil {
				return nil, err
			}
		} else {
			remaining = append(remaining, node)
		}
	}
	return remaining, nil
}

func (r *registry) buildAll(outputDir string, nodes []IRNode) (err error) {
	// Create output directory
	if info, err := os.Stat(outputDir); err == nil && info.IsDir() {
		return blueprint.Errorf("output directory %v already exists", outputDir)
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return blueprint.Errorf("unable to create output directory %v due to %v", outputDir, err.Error())
	}

	// Try to group like-nodes into namespaces first
	for _, builder := range r.namespace {
		nodes, err = builder.buildCompatibleNodes(outputDir, nodes)
		if err != nil {
			return err
		}
	}

	// Remaining nodes can be built individually
	nodes, err = buildArtifactGeneratorNodes(outputDir, nodes)
	if err != nil {
		return err
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
		// This should probably be a warning in general
		return blueprint.Errorf("No registered builders for node types %s", strings.Join(typeNames, ", "))
	}
	return nil
}
