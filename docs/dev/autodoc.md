# Generating code documentation in READMEs

Code documentation is generated using `gomarkdoc`.  After making substantial changes to the structs or funcs in a package, you may want to regenerate the documentation for the package.

Install `gomarkdoc`:

```
go install github.com/princjef/gomarkdoc/cmd/gomarkdoc@latest
```

Run `gomarkdoc` for a particular package; in this case we use the `blueprint` package as an example
```
cd blueprint
gomarkdoc -e --output '{{.Dir}}/README.md' ./...
```

Currently we use `gomarkdoc` to auto-generate READMEs for the following:
```
blueprint
plugins
runtime
```

For further information see the [gomarkdoc](https://github.com/princjef/gomarkdoc/tree/master) github repo.
