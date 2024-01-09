package dockergen

import (
	"bytes"
	"os"
	"strings"
	"text/template"

	"github.com/blueprint-uservices/blueprint/plugins/linux"
)

/*
Any template that is executed using ExecuteTemplate will be able to use
the helper functions defined in this file within the template.
*/

func ExecuteTemplate(name string, body string, args any) (string, error) {
	return newTemplateExecutor(args).exec(name, body, args)
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
	Funcs template.FuncMap
}

func newTemplateExecutor(args any) *templateExecutor {
	e := &templateExecutor{
		Funcs: template.FuncMap{},
	}

	e.Funcs["EnvVarName"] = e.EnvVarName
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

func (e *templateExecutor) EnvVarName(name string) (string, error) {
	return linux.EnvVar(name), nil
}

func (e *templateExecutor) TitleCase(arg string) (string, error) {
	return strings.Title(arg), nil
}
