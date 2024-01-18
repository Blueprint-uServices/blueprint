module github.com/blueprint-uservices/blueprint/plugins

go 1.20

require golang.org/x/exp v0.0.0-20231110203233-9a3e6036ecaa

require (
	github.com/otiai10/copy v1.11.0
	golang.org/x/mod v0.14.0
)

require golang.org/x/sys v0.14.0 // indirect

require (
	github.com/blueprint-uservices/blueprint/blueprint v0.0.0
	github.com/pkg/errors v0.9.1
	golang.org/x/tools v0.15.0
)

replace github.com/blueprint-uservices/blueprint/blueprint => ../blueprint
