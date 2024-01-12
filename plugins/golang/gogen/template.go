package gogen

import (
	"bytes"
	"fmt"
	"os"
	"reflect"
	"strings"
	"text/template"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint"
	"github.com/blueprint-uservices/blueprint/plugins/golang/gocode"
)

/*
Any template that is executed using ExecuteTemplate will be able to use
the helper functions defined in this file within the template.
*/

// A helper function for executing [text/template] templates to string.
//
// When executing the provided template body, the body can make use of a number
// of convenience functions.  See [gogen/template.go] for details.
//
// [gogen/template.go]: https://github.com/Blueprint-uServices/blueprint/tree/main/plugins/golang/gogen/template.go
func ExecuteTemplate(name string, body string, args any) (string, error) {
	e := newTemplateExecutor(args)

	// This is a hacky but very convenient way of dealing with the fact that imports
	// get declared before they're used... just compile twice.  The second pass
	// will compile the correct imports.  Alternative is much more verbose.
	// In the long run we can implement this properly but for now this works just fine.
	_, err := e.exec(name, body, args)
	if err != nil {
		return "", err
	}
	return e.exec(name, body, args)
}

// A helper function for executing [text/template] templates to file
//
// When executing the provided template body, the body can make use of a number
// of convenience functions.  See [gogen/template.go] for details.
//
// [gogen/template.go]: https://github.com/Blueprint-uServices/blueprint/tree/main/plugins/golang/gogen/template.go
func ExecuteTemplateToFile(name string, body string, args any, filename string) error {
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		return err
	}
	defer f.Close()

	code, err := ExecuteTemplate(name, body, args)
	if err != nil {
		return err
	}
	_, err = f.WriteString(code)
	return err
}

type templateExecutor struct {
	Funcs   template.FuncMap
	Imports *Imports
}

func newTemplateExecutor(args any) *templateExecutor {
	e := &templateExecutor{
		Funcs:   template.FuncMap{},
		Imports: getImports(args),
	}

	e.Funcs["NameOf"] = e.NameOf
	e.Funcs["DeclareArgVars"] = e.DeclareArgVars
	e.Funcs["ArgVars"] = e.ArgVars
	e.Funcs["ArgVarsCutoff"] = e.ArgVarsCutoff
	e.Funcs["ArgVarsEquals"] = e.ArgVarsEquals
	e.Funcs["ArgVarsAndTypes"] = e.ArgVarsAndTypes
	e.Funcs["ArgVarsAndTypesCutoff"] = e.ArgVarsAndTypesCutoff
	e.Funcs["RetVars"] = e.RetVars
	e.Funcs["RetVarsCutoff"] = e.RetVarsCutoff
	e.Funcs["RetVarsEquals"] = e.RetVarsEquals
	e.Funcs["RetTypes"] = e.RetTypes
	e.Funcs["RetVarsAndTypes"] = e.RetVarsAndTypes
	e.Funcs["RetVarsAndTypesCutoff"] = e.RetVarsAndTypesCutoff
	e.Funcs["Signature"] = e.Signature
	e.Funcs["SignatureWithRetVars"] = e.SignatureWithRetVars
	e.Funcs["JsonField"] = e.JsonField
	e.Funcs["Title"] = e.TitleCase
	e.Funcs["HasNewReturnVars"] = e.HasNewReturnVars

	return e
}

func (e *templateExecutor) exec(name string, body string, args any) (string, error) {
	t, err := template.New(name).Funcs(e.Funcs).Parse(body)
	if err != nil {
		return "", err
	}
	buf := &bytes.Buffer{}
	err = t.Execute(buf, args)
	return buf.String(), err
}

func (e *templateExecutor) NameOf(typeName gocode.TypeName) (string, error) {
	if e.Imports != nil {
		return e.Imports.NameOf(typeName), nil
	}
	return "", blueprint.Errorf("the NameOf template function requires that the template args have a *gocode.Imports field, but none was found")
}

func (e *templateExecutor) DeclareArgVars(f gocode.Func) (string, error) {
	tmpl := `{{range $i, $arg := .Arguments}}
	var {{$arg.Name}} {{NameOf $arg.Type}}
	{{- end}}`
	return e.exec("DeclareArgVars", tmpl, f)
}

func (e *templateExecutor) ArgVars(f gocode.Func, prefix ...string) (string, error) {
	return e.ArgVarsCutoff(f, 0, prefix...)
}

func (e *templateExecutor) ArgVarsCutoff(f gocode.Func, cutoff int, prefix ...string) (string, error) {
	for _, arg := range f.Arguments {
		prefix = append(prefix, arg.Name)
	}
	return strings.Join(prefix[:len(prefix)-cutoff], ", "), nil
}

func (e *templateExecutor) ArgVarsEquals(f gocode.Func, prefix ...string) (string, error) {
	if len(f.Arguments) == 0 && len(prefix) == 0 {
		return "", nil
	}
	for _, arg := range f.Arguments {
		prefix = append(prefix, arg.Name)
	}
	return fmt.Sprintf("%v :=", strings.Join(prefix, ", ")), nil
}

func (e *templateExecutor) ArgVarsAndTypes(f gocode.Func, prefix ...string) (string, error) {
	return e.ArgVarsAndTypesCutoff(f, 0, prefix...)
}

func (e *templateExecutor) ArgVarsAndTypesCutoff(f gocode.Func, cutoff int, prefix ...string) (string, error) {
	for _, arg := range f.Arguments {
		argType, err := e.NameOf(arg.Type)
		if err != nil {
			return "", err
		}
		prefix = append(prefix, arg.Name+" "+argType)
	}
	return strings.Join(prefix[:len(prefix)-cutoff], ", "), nil
}

func (e *templateExecutor) RetVars(f gocode.Func, suffix ...string) (string, error) {
	return e.RetVarsCutoff(f, 0, suffix...)
}

func (e *templateExecutor) RetVarsCutoff(f gocode.Func, cutoff int, suffix ...string) (string, error) {
	var retvars []string
	for i := range f.Returns {
		retvars = append(retvars, fmt.Sprintf("ret%v", i))
	}
	retvars = retvars[:len(retvars)-cutoff]
	retvars = append(retvars, suffix...)
	return strings.Join(retvars, ", "), nil
}

func (e *templateExecutor) RetVarsEquals(f gocode.Func, suffix ...string) (string, error) {
	if len(f.Returns) == 0 && len(suffix) == 0 {
		return "", nil
	}

	var retvars []string
	for i := range f.Returns {
		retvars = append(retvars, fmt.Sprintf("ret%v", i))
	}
	retvars = append(retvars, suffix...)
	return fmt.Sprintf("%v = ", strings.Join(retvars, ", ")), nil
}

func (e *templateExecutor) RetTypes(f gocode.Func, suffix ...string) (string, error) {
	var rettypes []string
	for _, ret := range f.Returns {
		rettype, err := e.NameOf(ret.Type)
		if err != nil {
			return "", err
		}
		rettypes = append(rettypes, rettype)
	}
	rettypes = append(rettypes, suffix...)
	return strings.Join(rettypes, ", "), nil
}

func (e *templateExecutor) RetVarsAndTypes(f gocode.Func, suffix ...string) (string, error) {
	return e.RetVarsAndTypesCutoff(f, 0, suffix...)
}

func (e *templateExecutor) RetVarsAndTypesCutoff(f gocode.Func, cutoff int, suffix ...string) (string, error) {
	var rettypes []string
	for i, ret := range f.Returns {
		rettype, err := e.NameOf(ret.Type)
		if err != nil {
			return "", err
		}
		rettypes = append(rettypes, fmt.Sprintf("ret%v %v", i, rettype))
	}
	rettypes = rettypes[:len(rettypes)-cutoff]
	rettypes = append(rettypes, suffix...)
	return strings.Join(rettypes, ", "), nil
}

func (e *templateExecutor) Signature(f gocode.Func) (string, error) {
	tmpl := `{{.Name}}({{ArgVarsAndTypes . "ctx context.Context"}}) ({{RetTypes . "error"}})`
	return e.exec("Signature", tmpl, f)
}

func (e *templateExecutor) SignatureWithRetVars(f gocode.Func) (string, error) {
	tmpl := `{{.Name}}({{ArgVarsAndTypes . "ctx context.Context"}}) ({{RetVarsAndTypes . "err error"}})`
	return e.exec("SignatureWithRetVars", tmpl, f)
}

func (e *templateExecutor) JsonField(name string) (string, error) {
	return "`json:\"" + name + "\"`", nil
}

func (e *templateExecutor) TitleCase(arg string) (string, error) {
	return strings.Title(arg), nil
}

func (e *templateExecutor) HasNewReturnVars(f gocode.Func) (string, error) {
	if len(f.Returns) != 0 {
		return ":=", nil
	}
	return "=", nil
}

// Looks for any field of type *Imports on the provided obj
func getImports(args any) *Imports {
	v := reflect.Indirect(reflect.ValueOf(args))
	imp := reflect.TypeOf(&Imports{})
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		if f.Type().AssignableTo(imp) {
			if imports, valid := f.Interface().(*Imports); valid {
				return imports
			}
		}
	}
	return nil
}
