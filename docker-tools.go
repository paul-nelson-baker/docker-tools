package docker_tools

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"io"
	"log"
	"strings"
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
func (c LazyDockerClient) LazyPullCallback(lazyImage LazyImage, callback DockerPullEventFunc) error {
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
	var event DockerPullEvent
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

type DockerPullEventFunc func(lazyImage LazyImage, event DockerPullEvent) error

// Logs any status or progress changes to the console via `log.Println`
func LazyLogPullEventCallback(lazyImage LazyImage, event DockerPullEvent) error {
	if event.Status != "" || event.Progress != "" {
		output := strings.TrimSpace(fmt.Sprintf("%s %s", event.Status, event.Progress))
		log.Println(output)
	}
	return nil
}

// An event that will be async returned to the client from the docker
// daemon as it pulls a remote image
type DockerPullEvent struct {
	Status         string `json:"status"`
	Error          string `json:"error"`
	Progress       string `json:"progress"`
	ProgressDetail struct {
		Current int `json:"current"`
		Total   int `json:"total"`
	} `json:"progressDetail"`
}

// Simplest way to pull which will allow you to control the context and cancelation
func (c LazyDockerClient) LazyPull(lazyImage LazyImage) (io.ReadCloser, context.CancelFunc, error) {
	fullyQualifiedImageName := lazyImage.FullName()
	ctx, cancelFunc := c.newLazyContext()
	closer, err := c.ImagePull(ctx, fullyQualifiedImageName, types.ImagePullOptions{
		All:           false,
		RegistryAuth:  "",
		PrivilegeFunc: nil,
	})
	return closer, cancelFunc, err
}

// Use this to reference official images
func DockerLibraryImage(name, version string) LazyImage {
	return LazyImage{
		Library: "docker.io/library",
		Name:    name,
		Version: version,
	}
}

// Use this to reference other third party images on docker hub
func DockerHubImage(name, version string) LazyImage {
	return LazyImage{
		Library: "registry.hub.docker.com",
		Name:    name,
		Version: version,
	}
}

// Used to reference a docker image as a 3-tuple
type LazyImage struct {
	Library string
	Name    string
	Version string
}

// A fully qualified name for an image which will all elements
func (l LazyImage) FullName() string {
	fullyQualifiedImageName := fmt.Sprintf("%s/%s:%s", l.Library, l.Name, l.Version)
	return fullyQualifiedImageName
}

// A short name. This is the name and version as a person would enter
// into a docker command. If the version is not present only the name
// is returned.
func (l LazyImage) ShortName() string {
	if l.Version == "" {
		return l.Name
	}
	shortImageName := fmt.Sprintf("%s:%s", l.Name, l.Version)
	return shortImageName
}

func (c LazyDockerClient) newLazyContext() (context.Context, context.CancelFunc) {
	return context.WithCancel(context.Background())
}
