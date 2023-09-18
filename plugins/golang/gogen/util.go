package gogen

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"text/template"
)

// Returns true if the specified path exists and is a directory; false otherwise
func IsDir(path string) bool {
	if info, err := os.Stat(path); err == nil && info.IsDir() {
		return true
	}
	return false
}

/*
Checks if the specified path exists and is a directory.
If `createIfAbsent` is true, then this will attempt to create the directory
*/
func CheckDir(path string, createIfAbsent bool) error {
	if info, err := os.Stat(path); err == nil {
		if info.IsDir() {
			return nil
		} else {
			return fmt.Errorf("expected %s to be a directory but it is not", path)
		}
	} else if errors.Is(err, os.ErrNotExist) {
		if !createIfAbsent {
			return fmt.Errorf("expected directory %s but it does not exist", path)
		}
		err = os.MkdirAll(path, 0755)
		if err != nil {
			return fmt.Errorf("unable to create directory %s due to %s", path, err.Error())
		}
		return nil
	} else {
		return fmt.Errorf("unexpected error for directory %s due to %s", path, err.Error())
	}
}

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

	return buf.String(), nil
}

func ExecuteTemplateToFile(name string, body string, args any, filename string) error {
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		return err
	}

	t, err := template.New(name).Parse(body)
	if err != nil {
		return err
	}

	return t.Execute(f, args)
}
