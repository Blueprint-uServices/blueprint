package docker

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
)

// For use
func BuildDockerfile(dockerfileName, dockerContext string, tags []string) error {
	ctx := context.Background()

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return err
	}
	fmt.Println(cli.ClientVersion())

	buildCtx, err := archive.TarWithOptions(dockerContext, &archive.TarOptions{})
	if err != nil {
		return err
	}

	buildOpts := types.ImageBuildOptions{
		Dockerfile: dockerfileName,
		Tags:       tags,
		Remove:     true,
	}

	rsp, err := cli.ImageBuild(ctx, buildCtx, buildOpts)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()

	_, err = io.Copy(os.Stdout, rsp.Body)
	return err
}

func PruneBuildCache() error {

	ctx := context.Background()

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return err
	}

	_, err = cli.BuildCachePrune(ctx, types.BuildCachePruneOptions{
		All: true,
	})
	return err
}

func makeDockerfileTar(name string, path string) (*bytes.Reader, error) {
	// Read the contents of the dockerfile
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	bs, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	// Write to a tar file
	buf := new(bytes.Buffer)
	w := tar.NewWriter(buf)
	defer w.Close()

	hdr := &tar.Header{Name: name, Size: int64(len(bs))}
	if err := w.WriteHeader(hdr); err != nil {
		return nil, err
	}

	_, err = w.Write(bs)
	return bytes.NewReader(buf.Bytes()), err
}
