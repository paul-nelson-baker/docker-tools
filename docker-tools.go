package docker_tools

import (
	"context"
	"encoding/json"
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

// Simple client with a default 15 minute timeout
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

// Let's pull an image, and key off of the events we get back as we're pulling
func (c LazyDockerClient) LazyPullCallback(image, version string, callback DockerPullEventFunc) error {
	readCloser, cancelFunc, err := c.LazyPull(image, version)
	// Make sure that any resources that need to be closed get closed
	// but make sure that we still error check where necessary
	if readCloser != nil {
		defer readCloser.Close()
	}
	if cancelFunc != nil {
		defer cancelFunc()
	}
	if err != nil {
		return err
	}
	// process the pull events until there aren't any left and pass
	// that event back to the function that we want to process it
	// terminate the processing if we encounter an error along the way.
	var event DockerPullEvent
	decoder := json.NewDecoder(readCloser)
	for {
		if err := decoder.Decode(&event); err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		if err := callback(image, version, event); err != nil {
			return err
		}
	}
	// everything went well
	return nil
}

type DockerPullEventFunc func(image string, version string, event DockerPullEvent) error

type DockerPullEvent struct {
	Status         string `json:"status"`
	Error          string `json:"error"`
	Progress       string `json:"progress"`
	ProgressDetail struct {
		Current int `json:"current"`
		Total   int `json:"total"`
	} `json:"progressDetail"`
}

func (c LazyDockerClient) LazyPull(image, version string) (io.ReadCloser, context.CancelFunc, error) {
	return c.LazyLibraryPull(`docker.io/library`, image, version)
}

func (c LazyDockerClient) LazyLibraryPull(library, image, version string) (io.ReadCloser, context.CancelFunc, error) {
	fullyQualifiedImageName := fmt.Sprintf("%s/%s:%s", library, image, version)
	ctx, cancelFunc := c.newLazyContext()
	closer, err := c.ImagePull(ctx, fullyQualifiedImageName, types.ImagePullOptions{
		All:           false,
		RegistryAuth:  "",
		PrivilegeFunc: nil,
	})
	return closer, cancelFunc, err
}

//func (c LazyDockerClient) LazyRun(image, version, containerName string) (container.ContainerCreateCreatedBody, error) {
//	lazyContext, cancelFunc := c.newLazyContext()
//	defer cancelFunc()
//	config := &container.Config{}
//	hostConfig := &container.HostConfig{}
//	networkingConfig := &network.NetworkingConfig{}
//	return c.ContainerCreate(lazyContext, config, hostConfig, networkingConfig, containerName)
//}

func (c LazyDockerClient) newLazyContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), c.Timeout)
}
