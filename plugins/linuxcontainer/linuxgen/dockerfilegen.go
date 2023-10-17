package linuxgen

import "path/filepath"

type Dockerfile struct {
	WorkspaceDir string
	FilePath     string
}

func NewDockerfile(workspaceDir string) *Dockerfile {
	return &Dockerfile{
		WorkspaceDir: workspaceDir,
		FilePath:     filepath.Join(workspaceDir, "Dockerfile"),
	}
}

func (d *Dockerfile) Generate() error {
	return ExecuteTemplateToFile("Dockerfile", dockerfileTemplate, d, d.FilePath)
}

var dockerfileTemplate = `FROM gcr.io/distroless/base-debian10
WORKDIR /

ENTRYPOINT ["/run.sh"]`
