package docker_tools

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	dockerMachineHelper "github.com/paul-nelson-baker/docker-machine-helper"
	"io"
	"time"
)

type LazyDockerClient struct {
	*client.Client
	Timeout time.Duration
}

func NewLazyClient() (LazyDockerClient, error) {
	if client, err := dockerMachineHelper.GetDockerClientEnvFallback(); err != nil {
		return LazyDockerClient{}, err
	} else {
		return LazyDockerClient{
			Client:  client,
			Timeout: time.Minute * 15,
		}, nil
	}
}

func (c LazyDockerClient) LazyPull(image, version string) (io.ReadCloser, error) {
	return c.LazyLibraryPull(`docker.io/library`, image, version)
}

func (c LazyDockerClient) LazyLibraryPull(library, image, version string) (io.ReadCloser, error) {
	fullyQualifiedImageName := fmt.Sprintf("%s/%s:%s", library, image, version)
	ctx, cancelFunc := c.newLazyContext()
	defer cancelFunc()
	return c.ImagePull(ctx, fullyQualifiedImageName, types.ImagePullOptions{
		All:           false,
		RegistryAuth:  "",
		PrivilegeFunc: nil,
	})
}

func (c LazyDockerClient) newLazyContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), c.Timeout)
}
