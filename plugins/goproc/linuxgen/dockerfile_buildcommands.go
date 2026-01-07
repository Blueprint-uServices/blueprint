package linuxgen

import "github.com/blueprint-uservices/blueprint/plugins/golang/gogen"

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

FROM golang:1.24-bookworm AS {{.ProcName}}

COPY ./{{.ProcName}} /src

WORKDIR /src
RUN go mod download

RUN mkdir /{{.ProcName}}
RUN go build -o /{{.ProcName}} ./{{.ProcName}}

#
# custom docker build commands provided by goproc.Process {{.ProcName}}
######## END
`
