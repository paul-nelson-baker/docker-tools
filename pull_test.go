package dockerTools

import (
	"github.com/paul-nelson-baker/docker-tools/image"
	"github.com/paul-nelson-baker/docker-tools/pull"
	"testing"
)

func TestLazyDockerClient_LazyPullCallback(t *testing.T) {
	dockerClient, err := NewLazyClient()
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	if err = dockerClient.LazyPullCallback(image.DockerLibraryImage("alpine", "latest"), pull.LoggingFunc); err != nil {
		t.Error(err)
	}
}
