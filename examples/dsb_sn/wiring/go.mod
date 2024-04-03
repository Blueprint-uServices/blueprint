module github.com/blueprint-uservices/blueprint/examples/dsb_sn/wiring

go 1.20

require github.com/blueprint-uservices/blueprint/blueprint v0.0.0

require github.com/blueprint-uservices/blueprint/plugins v0.0.0

require github.com/blueprint-uservices/blueprint/examples/dsb_sn/workflow v0.0.0

require github.com/blueprint-uservices/blueprint/examples/dsb_sn/tests v0.0.0

require (
	github.com/blueprint-uservices/blueprint/runtime v0.0.0-20240120085724-a66c24cd32b1 // indirect
	github.com/bradfitz/gomemcache v0.0.0-20230905024940-24af94b03874 // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible // indirect
	github.com/go-logr/logr v1.3.0 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/klauspost/compress v1.16.6 // indirect
	github.com/montanaflynn/stats v0.0.0-20171201202039-1bf9dbcd8cbe // indirect
	github.com/otiai10/copy v1.11.0 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/youmark/pkcs8 v0.0.0-20181117223130-1be2e3e5546d // indirect
	go.mongodb.org/mongo-driver v1.13.0 // indirect
	go.opentelemetry.io/otel v1.21.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdoutmetric v0.44.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.20.0 // indirect
	go.opentelemetry.io/otel/metric v1.21.0 // indirect
	go.opentelemetry.io/otel/sdk v1.21.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.21.0 // indirect
	go.opentelemetry.io/otel/trace v1.21.0 // indirect
	golang.org/x/crypto v0.11.0 // indirect
	golang.org/x/exp v0.0.0-20231110203233-9a3e6036ecaa // indirect
	golang.org/x/mod v0.14.0 // indirect
	golang.org/x/sync v0.5.0 // indirect
	golang.org/x/sys v0.14.0 // indirect
	golang.org/x/text v0.11.0 // indirect
	golang.org/x/tools v0.15.0 // indirect
)

replace github.com/blueprint-uservices/blueprint/blueprint => ../../../blueprint

replace github.com/blueprint-uservices/blueprint/plugins => ../../../plugins

replace github.com/blueprint-uservices/blueprint/examples/dsb_sn/workflow => ../workflow

replace github.com/blueprint-uservices/blueprint/examples/dsb_sn/tests => ../tests
