package dockerTools

import (
	"context"
	"encoding/json"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/paul-nelson-baker/docker-tools/image"
	"github.com/paul-nelson-baker/docker-tools/pull"
	"io"
)

type LazyDockerClient struct {
	*client.Client
}

// Simple client with a default 15 minute timeout
func NewLazyClient() (LazyDockerClient, error) {
	if client, err := GetDockerClientOrFallback(client.NewEnvClient, GetDockerMachineClient); err != nil {
		return LazyDockerClient{}, err
	} else {
		return LazyDockerClient{
			Client: client,
		}, nil
	}
}

// Let's pull an image, and key off of the events we get back as we're pulling
func (c LazyDockerClient) LazyPullCallback(lazyImage image.LazyImage, callback pull.DockerPullEventFunc) error {
	readCloser, cancelFunc, err := c.LazyPull(lazyImage)
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
	var event pull.DockerPullEvent
	decoder := json.NewDecoder(readCloser)
	for {
		if err := decoder.Decode(&event); err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		if err := callback(lazyImage, event); err != nil {
			return err
		}
	}
	// everything went well
	return nil
}

// Simplest way to pull which will allow you to control the context and cancelation
func (c LazyDockerClient) LazyPull(lazyImage image.LazyImage) (io.ReadCloser, context.CancelFunc, error) {
	fullyQualifiedImageName := lazyImage.FullName()
	ctx, cancelFunc := c.newLazyContext()
	closer, err := c.ImagePull(ctx, fullyQualifiedImageName, types.ImagePullOptions{
		All:           false,
		RegistryAuth:  "",
		PrivilegeFunc: nil,
	})
	return closer, cancelFunc, err
}

func (c LazyDockerClient) newLazyContext() (context.Context, context.CancelFunc) {
	return context.WithCancel(context.Background())
}
