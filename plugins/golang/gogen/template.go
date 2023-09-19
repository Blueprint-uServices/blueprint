package gogen

import (
	"bytes"
	"os"
	"text/template"
)

func ExecuteTemplate(name string, body string, args any) (string, error) {
	t, err := template.New(name).Parse(body)
	if err != nil {
		return "", err
	}

	buf := &bytes.Buffer{}
	err = t.Execute(buf, args)
	if err != nil {
		return "", err
	}

	// This is a hacky but very convenient way of dealing with the fact that imports
	// get declared before they're used... just compile twice.  The second pass
	// will compile the correct imports.  Alternative is much more verbose.
	// In the long run we can implement this properly but for now this works just fine.
	buf = &bytes.Buffer{}
	err = t.Execute(buf, args)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
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
