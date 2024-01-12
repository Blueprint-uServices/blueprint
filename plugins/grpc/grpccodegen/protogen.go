package grpccodegen

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint"
	"github.com/blueprint-uservices/blueprint/plugins/golang"
	"github.com/blueprint-uservices/blueprint/plugins/golang/gocode"
	"github.com/blueprint-uservices/blueprint/plugins/golang/goparser"
	"golang.org/x/exp/slog"
)

/*
Generates the GRPC .proto file for the provided service interface, then compiles it using `protoc`.

See the plugin README for the required GRPC and protocol buffers package dependencies.
*/
func GenerateGRPCProto(builder golang.ModuleBuilder, service *gocode.ServiceInterface, outputPackage string) error {
	// No need to generate the proto more than once
	if builder.Visited(outputPackage + "/" + service.BaseName + ".proto") {
		return nil
	}

	// Re-parse all of the modules, which can include generated code from other plugins
	modules, err := goparser.ParseWorkspace(builder.Workspace().Info().Path)
	if err != nil {
		return err
	}

	// Construct and validate the GRPC proto builder for the service
	pb := newProtoBuilder(modules, service.BaseName)
	splits := strings.Split(outputPackage, "/")
	outputPackageName := splits[len(splits)-1]
	pb.Module = builder.Info()
	pb.Package = outputPackageName
	pb.PackageName = pb.Module.Name + "/" + outputPackage

	err = pb.AddService(service)
	if err != nil {
		return err
	}

	// Filename munging
	outputDir := filepath.Join(builder.Info().Path, filepath.Join(splits...))
	err = os.MkdirAll(outputDir, 0755)
	if err != nil {
		return blueprint.Errorf("unable to create grpc output dir %v due to %v", outputDir, err.Error())
	}

	// Write the proto file
	outputFilename := filepath.Join(outputDir, service.BaseName+".proto")
	err = pb.WriteProtoFile(outputFilename)
	if err != nil {
		return err
	}

	// Compile the proto file
	err = CompileProtoFile(outputFilename)
	if err != nil {
		return err
	}

	// Generate the marshalling code
	slog.Info(fmt.Sprintf("Generating %v/%v_conversions.go", pb.PackageName, service.BaseName))
	marshallFile := filepath.Join(outputDir, service.BaseName+"_conversions.go")
	return pb.GenerateMarshallingCode(marshallFile)
}

func rel(path string) string {
	pwd, err := os.Getwd()
	if err != nil {
		return path
	}
	s, err := filepath.Rel(pwd, path)
	if err != nil {
		return path
	}
	return s
}

// Runs protoc on the specified protoFileName
func CompileProtoFile(protoFileName string) error {
	proto_path, _ := filepath.Split(protoFileName)
	cmd := exec.Command("protoc", protoFileName, "--go_out="+proto_path, "--go-grpc_out="+proto_path, "--proto_path="+proto_path)
	var out strings.Builder
	cmd.Stdout = &out
	cmd.Stderr = &out
	slog.Info(fmt.Sprintf("protoc %v --go_out=%v --go-grpc_out=%v --proto_path=%v", rel(protoFileName), rel(proto_path), rel(proto_path), rel(proto_path)))
	err := cmd.Run()
	if err != nil {
		slog.Error(out.String())
		return err
	} else {
		slog.Info("protoc compilation success")
	}
	return nil
}

/* A basic structural representation of the GRPC messages and services */
type (
	gRPCField struct {
		SrcType   gocode.TypeName // The source type
		ProtoType string          // The GRPC type in proto
		GRPCType  gocode.TypeName // The GRPC type in golang
		Name      string
		Position  int
	}

	gRPCMessageDecl struct {
		Builder   *gRPCProtoBuilder
		Name      string
		GRPCType  *gocode.UserType // The GRPC-generated type for this message
		FieldList []*gRPCField
	}

	gRPCMethodDecl struct {
		Service  *gRPCServiceDecl
		Name     string
		Request  *gRPCMessageDecl
		Response *gRPCMessageDecl
	}

	gRPCServiceDecl struct {
		Builder *gRPCProtoBuilder
		Name    string
		Methods map[string]*gRPCMethodDecl
	}

	gRPCProtoBuilder struct {
		Name        string
		Code        *goparser.ParsedModuleSet
		Package     string // Package shortname
		Module      golang.ModuleInfo
		PackageName string // Fully qualified package
		Services    map[string]*gRPCServiceDecl
		Messages    map[string]*gRPCMessageDecl
		Structs     map[gocode.UserType]*gRPCMessageDecl // Mapping from golang struct to the corresponding message
	}
)

func newProtoBuilder(code *goparser.ParsedModuleSet, name string) *gRPCProtoBuilder {
	b := &gRPCProtoBuilder{}
	b.Name = name
	b.Code = code
	b.Services = make(map[string]*gRPCServiceDecl)
	b.Messages = make(map[string]*gRPCMessageDecl)
	b.Structs = make(map[gocode.UserType]*gRPCMessageDecl)
	return b
}

var protoFileTemplate = `syntax="proto3";
option go_package="./;{{ .Package }}";
package {{ .Package }};

{{ range $k, $msg := .Messages }}
message {{$msg.Name}} {
    {{- range $k, $field := $msg.FieldList}}
    {{$field.ProtoType}} {{$field.Name}} = {{$field.Position}};
    {{- end}}
}
{{ end -}}

{{ range $k, $service := .Services }}
service {{$service.Name}} {
    {{- range $k, $method := $service.Methods}}
    rpc {{$method.Name}} ({{$method.Request.Name}}) returns ({{$method.Response.Name}}) {}
    {{- end}}
}
{{ end }}
`

func (b *gRPCProtoBuilder) WriteProtoFile(outputFilePath string) error {
	t, err := template.New("protofile").Parse(protoFileTemplate)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(outputFilePath, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		return err
	}

	return t.Execute(f, b)

}

func (b *gRPCProtoBuilder) newMessage(name string) *gRPCMessageDecl {
	s := &gRPCMessageDecl{}
	s.Builder = b
	s.Name = name
	s.FieldList = nil
	s.GRPCType = &gocode.UserType{Name: name, Package: b.PackageName}
	b.Messages[name] = s
	return s
}

func (b *gRPCProtoBuilder) newService(name string) *gRPCServiceDecl {
	s := &gRPCServiceDecl{}
	s.Builder = b
	s.Name = name
	s.Methods = make(map[string]*gRPCMethodDecl)
	b.Services[name] = s
	return s
}

func (b *gRPCServiceDecl) newMethod(name string) *gRPCMethodDecl {
	m := &gRPCMethodDecl{}
	m.Service = b
	m.Name = name
	m.Request = b.Builder.newMessage(fmt.Sprintf("%s_%s_Request", b.Name, name))
	m.Response = b.Builder.newMessage(fmt.Sprintf("%s_%s_Response", b.Name, name))
	b.Methods[name] = m
	return m
}

func (b *gRPCProtoBuilder) makeFieldList(vars []gocode.Variable) ([]*gRPCField, error) {
	var fieldList []*gRPCField
	for i, arg := range vars {
		protoType, grpcType, err := b.getGRPCType(arg.Type)
		if err != nil {
			return nil, blueprint.Errorf("cannot serialize %v of type %v for GRPC due to %v", arg.Name, arg.Type, err.Error())
		}

		name := arg.Name
		if name == "" {
			name = fmt.Sprintf("ret%v", i)
		}
		fieldList = append(fieldList, &gRPCField{
			SrcType:   arg.Type,
			ProtoType: protoType,
			GRPCType:  grpcType,
			Name:      name,
			Position:  i + 1,
		})
	}
	return fieldList, nil
}

/*
Adds a service declaration for the provided golang service interface.

# This will create message and service definitions within the grpc proto

For arguments and return values on methods in the interface, corresponding GRPC message objects
are needed.  The ProtoBuilder will consult the parsed code to find the definitions of arguments
and return values.
*/
func (b *gRPCProtoBuilder) AddService(iface *gocode.ServiceInterface) error {
	serviceDecl := b.newService(iface.Name) // TODO: (not implemented yet) possibility of name collisions
	for _, method := range iface.Methods {
		argList, err := b.makeFieldList(method.Arguments)
		if err != nil {
			return err
		}

		retList, err := b.makeFieldList(method.Returns)
		if err != nil {
			return err
		}

		methodDecl := serviceDecl.newMethod(method.Name)
		methodDecl.Request.FieldList = argList
		methodDecl.Response.FieldList = retList
	}
	return nil
}

func (b *gRPCProtoBuilder) GetOrAddMessage(t *gocode.UserType) (*gRPCMessageDecl, error) {
	// Message might already exist
	if msgDecl, exists := b.Structs[*t]; exists {
		return msgDecl, nil
	}

	// Find the struct definition in the module
	pkg := b.Code.GetPackage(t.Package)
	if pkg == nil {
		return nil, blueprint.Errorf("could not find package %v for type %v", t.Package, t)
	}
	struc, hasStruct := pkg.Structs[t.Name]
	if !hasStruct {
		// It's possible that the type does exist but it wasn't declared as a struct, e.g. it is
		// an enum or a type alias. Non-struct types are not-yet-implemented
		if _, hasTypeDef := pkg.DeclaredTypes[t.Name]; hasTypeDef {
			return nil, blueprint.Errorf("expected %v to be a struct but it is an unsupported type", t.String())
		} else {
			return nil, blueprint.Errorf("could not find %v within %v", t.Name, t.Package)
		}
	}

	// Create the message
	msg := b.newMessage(fmt.Sprintf("%v_%v", b.Name, t.Name))
	b.Structs[*t] = msg
	for _, field := range struc.FieldsList {
		// We ignore promoted and anonymous struct / interface extensions
		if _, isNamed := struc.Fields[field.Name]; !isNamed {
			// TODO (not implemented yet): support promoted and anonymous, handle interfaces and promoted struct fields
			continue
		}

		// Gets the type name of this field, possibly internally creating the GRPC message if it's a struct
		fieldProto, fieldGRPC, err := b.getGRPCType(field.Type)
		if err != nil {
			return nil, err
		}

		msg.FieldList = append(msg.FieldList, &gRPCField{
			SrcType:   field.Type,
			ProtoType: fieldProto,
			GRPCType:  fieldGRPC,
			Name:      field.Name,
			Position:  len(msg.FieldList) + 1,
		})
	}

	return msg, nil
}

var basicToGrpc = map[string]string{
	"bool":   "bool",
	"string": "string",
	"int":    "sint64", "int8": "sint32", "int16": "sint32", "int32": "sint32", "int64": "sint64",
	"uint": "uint64", "uint8": "uint32", "uint16": "uint32", "uint32": "uint32", "uint64": "uint64",
	"byte":    "uint8",
	"rune":    "uint8",
	"float32": "float", "float64": "double",
}

var grpcToBasic = map[string]string{
	"bool":   "bool",
	"string": "string",
	"sint32": "int32", "sint64": "int64",
	"uint32": "uint32", "uint64": "uint64",
	"uint8":   "byte",
	"float32": "float",
	"float64": "double",
}

var acceptableMapKeys map[string]struct{}

func getMapKeyType(t gocode.TypeName) (string, gocode.TypeName, bool) {
	if acceptableMapKeys == nil {
		keys := []string{
			"int32", "int64", "uint32", "uint64", "sint32", "sint64",
			"fixed32", "fixed64", "sfixed32", "sfixed64", "bool", "string",
		}
		acceptableMapKeys = make(map[string]struct{})
		for _, key := range keys {
			acceptableMapKeys[key] = struct{}{}
		}
	}
	if basic, isBasic := t.(*gocode.BasicType); isBasic {
		if grpcType, hasGrpcType := basicToGrpc[basic.Name]; hasGrpcType {
			if _, isValid := acceptableMapKeys[grpcType]; isValid {
				return grpcType, t, true
			}
		}
	}
	return "", nil, false
}

// Returns the name of the type for the .proto declaration and the corresponding golang type,
// which may be different from the source type
func (b *gRPCProtoBuilder) getGRPCType(t gocode.TypeName) (string, gocode.TypeName, error) {
	switch arg := t.(type) {
	case *gocode.UserType:
		{
			msg, err := b.GetOrAddMessage(arg)
			if err != nil {
				return "", nil, err
			}
			return msg.Name, msg.GRPCType, nil
		}
	case *gocode.BasicType:
		{
			if grpcType, hasGrpcType := basicToGrpc[arg.Name]; hasGrpcType {
				return grpcType, &gocode.BasicType{Name: grpcToBasic[grpcType]}, nil
			}
			return "", nil, blueprint.Errorf("%v is not supported by GRPC", arg.Name)
		}
	case *gocode.Pointer:
		{
			protoType, pointerToGrpcType, err := b.getGRPCType(arg.PointerTo)
			if err != nil {
				return "", nil, err
			}
			grpcType := &gocode.Pointer{PointerTo: pointerToGrpcType}
			return protoType, grpcType, nil
		}
	case *gocode.Map:
		{
			keyProto, keyGRPC, isValidKey := getMapKeyType(arg.KeyType)
			if !isValidKey {
				return "", nil, blueprint.Errorf("GRPC cannot use %v as a map key", arg.KeyType)
			}
			valueProto, valueGRPC, err := b.getGRPCType(arg.ValueType)
			if err != nil {
				return "", nil, err
			}
			protoType := fmt.Sprintf("map<%v,%v>", keyProto, valueProto)
			grpcType := &gocode.Map{KeyType: keyGRPC, ValueType: valueGRPC}
			return protoType, grpcType, nil
		}
	case *gocode.Slice:
		{
			// []byte is a special case where the type is 'bytes', everything else is a repeated
			if basic, isBasic := arg.SliceOf.(*gocode.BasicType); isBasic && basic.Name == "byte" {
				return "bytes", t, nil
			}
			// map is a special case that can't be repeated
			if _, isMap := arg.SliceOf.(*gocode.Map); isMap {
				return "", nil, blueprint.Errorf("GRPC does not support arrays of maps %v", t.String())
			}
			sliceProto, sliceGRPC, err := b.getGRPCType(arg.SliceOf)
			if err != nil {
				return "", nil, err
			}
			protoType := fmt.Sprintf("repeated %v", sliceProto)
			grpcType := &gocode.Slice{SliceOf: sliceGRPC}
			return protoType, grpcType, nil
		}
	default:
		{
			// all others are invalid or not yet supported
			return "", nil, blueprint.Errorf("GRPC cannot serialize %v", t.String())
		}
	}
}
