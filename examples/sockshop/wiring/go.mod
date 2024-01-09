module github.com/Blueprint-uServices/blueprint/examples/sockshop/wiring

go 1.20

require github.com/Blueprint-uServices/blueprint/blueprint v0.0.0

require github.com/Blueprint-uServices/blueprint/plugins v0.0.0

require (
	github.com/otiai10/copy v1.11.0 // indirect
	golang.org/x/exp v0.0.0-20231110203233-9a3e6036ecaa // indirect
	golang.org/x/mod v0.14.0 // indirect
	golang.org/x/sys v0.14.0 // indirect
	golang.org/x/tools v0.15.0 // indirect
)

replace github.com/Blueprint-uServices/blueprint/blueprint => ../../../blueprint

replace github.com/Blueprint-uServices/blueprint/plugins => ../../../plugins
