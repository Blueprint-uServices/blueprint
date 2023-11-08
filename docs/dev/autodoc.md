# Generating code documentation in READMEs

Code documentation is generated using `gomarkdoc`.  After making substantial changes to the structs or funcs in a package, you may want to regenerate the documentation for the package.

Install `gomarkdoc`:

```
go install github.com/princjef/gomarkdoc/cmd/gomarkdoc@latest
```

Run `gomarkdoc` for a particular package; in this case we use the `blueprint` package as an example
```
cd blueprint
gomarkdoc --output '{{.Dir}}/README.md' ./...
```

For further information see the [gomarkdoc](https://github.com/princjef/gomarkdoc/tree/master) github repo.
