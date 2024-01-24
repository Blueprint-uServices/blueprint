package thriftcodegen

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint"
	"github.com/blueprint-uservices/blueprint/plugins/golang"
	"github.com/blueprint-uservices/blueprint/plugins/golang/gocode"
	"github.com/blueprint-uservices/blueprint/plugins/golang/gogen"
	"github.com/blueprint-uservices/blueprint/plugins/golang/goparser"
	"github.com/blueprint-uservices/blueprint/plugins/workflow/workflowspec"
	"golang.org/x/exp/slog"
)

// Generates the .thrift file for the provided service interface, then compiles it using `thrift`.
// See the plugin README for the required thrift package dependencies.
func GenerateThrift(builder golang.ModuleBuilder, service *gocode.ServiceInterface, outputPackage string) error {
	if builder.Visited(outputPackage + "/" + service.BaseName + ".thrift") {
		return nil
	}

	modules := workflowspec.Get().Derive().Modules
	if err := modules.AddWorkspace(builder.Workspace().Info().Path); err != nil {
		return err
	}

	tf := NewThriftBuilder(modules)
	splits := strings.Split(outputPackage, "/")
	outputPackageName := splits[len(splits)-1]
	tf.Module = builder.Info()
	tf.Package = outputPackageName
	tf.PackageName = tf.Module.Name + "/" + outputPackage
	tf.ImportName = strings.ToLower(service.BaseName)
	tf.InternalPkg = tf.PackageName + "/" + tf.ImportName

	err := tf.AddService(service)
	if err != nil {
		return err
	}

	outputDir := filepath.Join(builder.Info().Path, filepath.Join(splits...))
	err = os.MkdirAll(outputDir, 0755)
	if err != nil {
		return blueprint.Errorf("unable to create thrift output dir %v due to %v", outputDir, err.Error())
	}

	outputFilename := filepath.Join(outputDir, service.BaseName+".thrift")
	err = tf.WriteThriftFile(outputFilename)
	if err != nil {
		return err
	}

	err = CompileThriftFile(outputFilename)
	if err != nil {
		return err
	}

	slog.Info(fmt.Sprintf("Generating %v/%v_conversions.go", tf.PackageName, service.BaseName))
	marshallFile := filepath.Join(outputDir, service.BaseName+"_conversions.go")
	return tf.GenerateMarshallingCode(marshallFile)
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

func CompileThriftFile(thriftFileName string) error {
	thrift_path, _ := filepath.Split(thriftFileName)
	cmd := exec.Command("thrift", "--gen", "go", "--out", thrift_path, thriftFileName)
	slog.Info(fmt.Sprintf("thrift --gen go --out %v %v", thrift_path, thriftFileName))
	var out strings.Builder
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	if err != nil {
		slog.Error(out.String())
		return err
	} else {
		slog.Info("thrift compilation success")
	}
	return nil
}

// A basic structural representation of the Thrift messages and services
type ThriftField struct {
	SrcType      gocode.TypeName
	ThriftType   string
	ThriftGoType gocode.TypeName
	Name         string
	Position     int
}

type ThriftStructDecl struct {
	Builder    *ThriftBuilder
	Name       string
	ThriftType *gocode.UserType
	FieldList  []*ThriftField
}

type ThriftMethodDecl struct {
	Service  *ThriftServiceDecl
	Name     string
	Request  *ThriftStructDecl
	Response *ThriftStructDecl
}

type ThriftServiceDecl struct {
	Builder *ThriftBuilder
	Name    string
	Methods map[string]*ThriftMethodDecl
}

type ThriftBuilder struct {
	Code        *goparser.ParsedModuleSet
	Package     string
	Module      golang.ModuleInfo
	PackageName string
	ImportName  string
	InternalPkg string
	Services    map[string]*ThriftServiceDecl
	Structs     map[string]*ThriftStructDecl
	GoStructs   map[gocode.UserType]*ThriftStructDecl
}

func NewThriftBuilder(code *goparser.ParsedModuleSet) *ThriftBuilder {
	t := &ThriftBuilder{}
	t.Code = code
	t.Services = make(map[string]*ThriftServiceDecl)
	t.Structs = make(map[string]*ThriftStructDecl)
	t.GoStructs = make(map[gocode.UserType]*ThriftStructDecl)
	return t
}

var thriftFileTemplate = `
{{range $_, $struct := .Structs}}
struct {{$struct.Name}} {
	{{- range $_, $field := $struct.FieldList}}
	{{$field.Position}}: {{$field.ThriftType}} {{$field.Name}},
	{{- end}}
}
{{end}}

{{range $_, $service := .Services}}
service {{$service.Name}} {
	{{- range $_, $method := $service.Methods}}
	{{$method.Response.Name}} {{$method.Name}} (1:{{$method.Request.Name}} req),
	{{- end}}
}
{{end}}
`

func (b *ThriftBuilder) WriteThriftFile(outputFilePath string) error {
	return gogen.ExecuteTemplateToFile("Thrift", thriftFileTemplate, b, outputFilePath)
}

func (b *ThriftBuilder) newStruct(name string) *ThriftStructDecl {
	s := &ThriftStructDecl{}
	s.Builder = b
	s.Name = name
	s.FieldList = nil
	s.ThriftType = &gocode.UserType{Name: name, Package: b.PackageName}
	b.Structs[name] = s
	return s
}

func (b *ThriftBuilder) newService(name string) *ThriftServiceDecl {
	s := &ThriftServiceDecl{}
	s.Builder = b
	s.Name = name
	s.Methods = make(map[string]*ThriftMethodDecl)
	b.Services[name] = s
	return s
}

func (s *ThriftServiceDecl) newMethod(name string) *ThriftMethodDecl {
	m := &ThriftMethodDecl{}
	m.Service = s
	m.Name = name
	m.Request = s.Builder.newStruct(fmt.Sprintf("%s_%s_Request", s.Name, name))
	m.Response = s.Builder.newStruct(fmt.Sprintf("%s_%s_Response", s.Name, name))
	s.Methods[name] = m
	return m
}

func (b *ThriftBuilder) makeFieldList(vars []gocode.Variable) ([]*ThriftField, error) {
	var fieldList []*ThriftField
	for i, arg := range vars {
		thriftType, goThriftType, err := b.getThriftType(arg.Type)
		if err != nil {
			return nil, blueprint.Errorf("cannot serialize %v of type %v for Thrift due to %v", arg.Name, arg.Type, err.Error())
		}

		name := arg.Name
		if name == "" {
			name = fmt.Sprintf("ret%v", i)
		}
		fieldList = append(fieldList, &ThriftField{
			SrcType:      arg.Type,
			ThriftType:   thriftType,
			ThriftGoType: goThriftType,
			Name:         name,
			Position:     i + 1,
		})
	}
	return fieldList, nil
}

// Adds a service declaration for the provided golang service interface.
func (b *ThriftBuilder) AddService(iface *gocode.ServiceInterface) error {
	serviceDecl := b.newService(iface.Name)
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

func (b *ThriftBuilder) GetOrAddMessage(t *gocode.UserType) (*ThriftStructDecl, error) {
	if structDecl, exists := b.GoStructs[*t]; exists {
		return structDecl, nil
	}

	pkg, err := b.Code.GetPackage(t.Package)
	if err != nil {
		return nil, blueprint.Errorf("could not find package %v for type %v due to: %v", t.Package, t, err)
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

	thrift_struct := b.newStruct(t.Name)
	b.GoStructs[*t] = thrift_struct
	for _, field := range struc.FieldsList {
		// We ignore promoted and anonymous struct / interface extensions
		if _, isNamed := struc.Fields[field.Name]; !isNamed {
			// TODO (not implemented yet): support promoted and anonymous, handle interfaces and promoted struct fields
			continue
		}

		fieldThrift, fieldGoThrift, err := b.getThriftType(field.Type)
		if err != nil {
			return nil, err
		}

		thrift_struct.FieldList = append(thrift_struct.FieldList, &ThriftField{
			SrcType:      field.Type,
			ThriftType:   fieldThrift,
			ThriftGoType: fieldGoThrift,
			Name:         field.Name,
			Position:     len(thrift_struct.FieldList) + 1,
		})

	}

	return thrift_struct, nil
}

var basicToThirft = map[string]string{
	"bool":   "bool",
	"string": "string",
	"int":    "i32",
	"int32":  "i32",
	"int64":  "i64",
	"int16":  "i16",
	"int8":   "byte",
	// Use 64-bit integers for unsigned integers as thrift only has support for signed ints
	"uint32":  "i64",
	"uint64":  "i64",
	"uint8":   "i64",
	"uint16":  "i64",
	"float32": "double",
	"float64": "double",
	"byte":    "byte",
	"rune":    "byte",
}

var thriftToBasic = map[string]string{
	"bool":   "bool",
	"string": "string",
	"byte":   "byte",
	"double": "double",
	"i64":    "int64",
	"i32":    "int32",
	"i16":    "int16",
}

func (b *ThriftBuilder) getThriftType(t gocode.TypeName) (string, gocode.TypeName, error) {
	switch arg := t.(type) {
	case *gocode.UserType:
		struc, err := b.GetOrAddMessage(arg)
		if err != nil {
			return "", nil, err
		}
		return struc.Name, struc.ThriftType, nil
	case *gocode.BasicType:
		if thriftType, ok := basicToThirft[arg.Name]; ok {
			return thriftType, &gocode.BasicType{Name: thriftToBasic[thriftType]}, nil
		}
		return "", nil, blueprint.Errorf("%v is not supported by Thrift", arg.Name)
	case *gocode.Pointer:
		protoType, pointerType, err := b.getThriftType(arg.PointerTo)
		if err != nil {
			return "", nil, err
		}
		thriftType := &gocode.Pointer{PointerTo: pointerType}
		return protoType, thriftType, nil
	case *gocode.Map:
		keyThrift, keyGoThrift, err := b.getThriftType(arg.KeyType)
		if err != nil {
			return "", nil, err
		}
		valueThrift, valueGoThrift, err := b.getThriftType(arg.ValueType)
		if err != nil {
			return "", nil, err
		}
		thriftType := fmt.Sprintf("map<%v,%v>", keyThrift, valueThrift)
		thriftGoType := &gocode.Map{KeyType: keyGoThrift, ValueType: valueGoThrift}
		return thriftType, thriftGoType, nil
	case *gocode.Slice:
		sliceType, sliceGoType, err := b.getThriftType(arg.SliceOf)
		if err != nil {
			return "", nil, err
		}
		thriftType := fmt.Sprintf("list<%v>", sliceType)
		thriftGoType := &gocode.Slice{SliceOf: sliceGoType}
		return thriftType, thriftGoType, nil
	default:
		return "", nil, blueprint.Errorf("Thrift cannot serialize %v", t.String())
	}
}
