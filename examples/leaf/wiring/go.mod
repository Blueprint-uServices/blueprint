module github.com/blueprint-uservices/blueprint/examples/leaf/wiring

go 1.20

require github.com/blueprint-uservices/blueprint/blueprint v0.0.0

require github.com/blueprint-uservices/blueprint/plugins v0.0.0

require github.com/blueprint-uservices/blueprint/examples/leaf/workflow v0.0.0

require (
	github.com/otiai10/copy v1.11.0 // indirect
	golang.org/x/exp v0.0.0-20231110203233-9a3e6036ecaa // indirect
	golang.org/x/mod v0.14.0 // indirect
	golang.org/x/sys v0.14.0 // indirect
	golang.org/x/tools v0.15.0 // indirect
)

replace github.com/blueprint-uservices/blueprint/blueprint => ../../../blueprint

replace github.com/blueprint-uservices/blueprint/plugins => ../../../plugins

replace github.com/blueprint-uservices/blueprint/examples/leaf/workflow => ../workflow
