package goprocgen

import "gitlab.mpi-sws.org/cld/blueprint/plugins/golang/gogen"

/*
If the goproc is being deployed to Docker, we can provide some custom
build commands to add to the Dockerfile
*/
func GenerateDockerfileBuildCommands(goProcName string) (string, error) {
	args := dockerfileBuildTemplateArgs{
		ProcName: goProcName,
	}
	return gogen.ExecuteTemplate("dockerfile_buildgoproc", dockerfileBuildTemplate, args)
}

type dockerfileBuildTemplateArgs struct {
	ProcName string
}

var dockerfileBuildTemplate = `
####### BEGIN
#  custom docker build commands provided by goproc.Process {{.ProcName}}
#

FROM golang:1.18-buster AS {{.ProcName}}

COPY ./{{.ProcName}} /src

WORKDIR /src
RUN go mod download

WORKDIR /
RUN mkdir /{{.ProcName}}
RUN go build -o /{{.ProcName}} /src/{{.ProcName}}

#
# custom docker build commands provided by goproc.Process {{.ProcName}}
######## END
`
