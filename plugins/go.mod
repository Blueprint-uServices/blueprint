module gitlab.mpi-sws.org/cld/blueprint/plugins

go 1.20

require golang.org/x/exp v0.0.0-20231110203233-9a3e6036ecaa

require (
	github.com/otiai10/copy v1.11.0
	golang.org/x/mod v0.14.0
)

require (
	github.com/Azure/go-ansiterm v0.0.0-20230124172434-306776ec8161 // indirect
	github.com/Microsoft/go-winio v0.6.1 // indirect
	github.com/containerd/containerd v1.7.11 // indirect
	github.com/containerd/continuity v0.4.3 // indirect
	github.com/distribution/reference v0.5.0 // indirect
	github.com/docker/distribution v2.8.3+incompatible // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/klauspost/compress v1.16.0 // indirect
	github.com/moby/patternmatcher v0.6.0 // indirect
	github.com/moby/sys/sequential v0.5.0 // indirect
	github.com/moby/term v0.5.0 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.0-rc2.0.20221005185240-3a7f492d3f1b // indirect
	github.com/opencontainers/runc v1.1.10 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	golang.org/x/net v0.18.0 // indirect
	golang.org/x/sys v0.14.0 // indirect
	golang.org/x/time v0.5.0 // indirect
	gotest.tools/v3 v3.5.1 // indirect
)

require (
	github.com/docker/docker v23.0.3+incompatible
	gitlab.mpi-sws.org/cld/blueprint/blueprint v0.0.0
	golang.org/x/tools v0.15.0
)

replace gitlab.mpi-sws.org/cld/blueprint/blueprint => ../blueprint
