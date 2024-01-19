# Generating code documentation in READMEs

Code documentation is generated using `gomarkdoc`.  After making substantial changes to the structs or funcs in a package, you may want to regenerate the documentation for the package.

Install `gomarkdoc`:

```
go install github.com/princjef/gomarkdoc/cmd/gomarkdoc@latest
```

Run `gomarkdoc` for a particular package; in this case we use the `blueprint` package as an example
```
cd blueprint
gomarkdoc --output '{{.Dir}}/README.md' --repository.default-branch main --repository.url https://github.com/blueprint-uservices/blueprint ./...
```

Currently we use `gomarkdoc` to auto-generate READMEs for the following:
```
blueprint
plugins
runtime
examples/*/workflow
examples/*/wiring
examples/*/workload
```

The following command, when run from the root of the blueprint repository, refreshes all documentation:

```
gomarkdoc --output '{{.Dir}}/README.md' --repository.default-branch main --repository.url https://github.com/blueprint-uservices/blueprint ./blueprint/... ./plugins/... ./runtime/... ./examples/sockshop/wiring/... ./examples/sockshop/workflow/... ./examples/sockshop/workload/... ./examples/dsb_hotel/wiring/... ./examples/dsb_hotel/workflow/... ./examples/dsb_sn/wiring/... ./examples/dsb_sn/workflow/... ./examples/leaf/wiring/... ./examples/leaf/workflow/... ./examples/train_ticket/wiring/... ./examples/train_ticket/workflow/... 
```

For further information see the [gomarkdoc](https://github.com/princjef/gomarkdoc/tree/master) github repo.