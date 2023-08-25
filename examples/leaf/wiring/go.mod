module gitlab.mpi-sws.org/cld/blueprint/examples/leaf/wiring

go 1.19

require golang.org/x/exp v0.0.0-20230728194245-b0cb94b80691

require gitlab.mpi-sws.org/cld/blueprint/blueprint v0.0.0
replace gitlab.mpi-sws.org/cld/blueprint/blueprint => ../../../blueprint

require gitlab.mpi-sws.org/cld/blueprint/plugins v0.0.0
replace gitlab.mpi-sws.org/cld/blueprint/plugins => ../../../plugins

require (
	github.com/otiai10/copy v1.11.0 // indirect
	golang.org/x/mod v0.11.0 // indirect
	golang.org/x/sys v0.1.0 // indirect
)

