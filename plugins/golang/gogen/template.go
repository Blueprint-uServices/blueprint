package gogen

import (
	"bytes"
	"os"
	"reflect"
	"strings"
	"text/template"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang/gocode"
)

/*
Any template that is executed using ExecuteTemplate will be able to use
the helper functions defined in this file within the template.
*/

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
	e.Funcs["ArgVarsAndTypes"] = e.ArgVarsAndTypes
	e.Funcs["RetVars"] = e.RetVars
	e.Funcs["RetTypes"] = e.RetTypes
	e.Funcs["RetVarsAndTypes"] = e.RetVarsAndTypes
	e.Funcs["Signature"] = e.Signature
	e.Funcs["SignatureWithRetVars"] = e.SignatureWithRetVars
	e.Funcs["JsonField"] = e.JsonField
	e.Funcs["Title"] = e.TitleCase

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
	tmpl := "{{range $i, $arg := .Arguments}}{{if $i}}, {{end}}{{$arg.Name}}{{end}}"
	s, err := e.exec("ArgVars", tmpl, f)
	prefix = append(prefix, s)
	return strings.Join(prefix, ", "), err
}

func (e *templateExecutor) ArgVarsAndTypes(f gocode.Func, prefix ...string) (string, error) {
	tmpl := "{{range $i, $arg := .Arguments}}{{if $i}}, {{end}}{{$arg.Name}} {{NameOf $arg.Type}}{{end}}"
	s, err := e.exec("ArgVarsAndTypes", tmpl, f)
	prefix = append(prefix, s)
	return strings.Join(prefix, ", "), err
}

func (e *templateExecutor) RetVars(f gocode.Func, suffix ...string) (string, error) {
	tmpl := "{{range $i, $_ := .Returns}}{{if $i}}, {{end}}ret{{$i}}{{end}}"
	s, err := e.exec("RetVars", tmpl, f)
	suffix = append([]string{s}, suffix...)
	return strings.Join(suffix, ", "), err
}

func (e *templateExecutor) RetTypes(f gocode.Func, suffix ...string) (string, error) {
	tmpl := "{{range $i, $ret := .Returns}}{{if $i}}, {{end}}{{NameOf $ret.Type}}{{end}}"
	s, err := e.exec("RetTypes", tmpl, f)
	suffix = append([]string{s}, suffix...)
	return strings.Join(suffix, ", "), err
}

func (e *templateExecutor) RetVarsAndTypes(f gocode.Func, suffix ...string) (string, error) {
	tmpl := "{{range $i, $ret := .Returns}}{{if $i}}, {{end}}ret{{$i}} {{NameOf $ret.Type}}{{end}}"
	s, err := e.exec("RetVarsAndTypes", tmpl, f)
	suffix = append([]string{s}, suffix...)
	return strings.Join(suffix, ", "), err
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
