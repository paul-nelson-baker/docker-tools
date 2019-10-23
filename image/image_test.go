package image

import (
	"testing"
)

func TestDockerLibraryImage(t *testing.T) {
	image := DockerLibraryImage("golang", "latest")
	if image.Library != "docker.io/library" {
		t.Errorf(`Image didn't have correct library: '%s'`, image.Library)
	}
	if image.Name != "golang" {
		t.Errorf(`Image had incorrect name: '%s'`, image.Name)
	}
	if image.Version != "latest" {
		t.Errorf(`Image had incorrect version: '%s'`, image.Version)
	}

	if fullName := image.FullName(); fullName != "docker.io/library/golang:latest" {
		t.Errorf(`Full name wasn't properly formated: '%s'`, fullName)
	}
	if shortName := image.ShortName(); shortName != "golang:latest" {
		t.Errorf(`Full name wasn't properly formated: '%s'`, shortName)
	}
}

func TestDockerHubImage(t *testing.T) {
	image := DockerHubImage("kitematic/minecraft", "latest")
	if image.Library != "registry.hub.docker.com" {
		t.Errorf(`Image didn't have correct library: '%s'`, image.Library)
	}
	if image.Name != "kitematic/minecraft" {
		t.Errorf(`Image had incorrect name: '%s'`, image.Name)
	}
	if image.Version != "latest" {
		t.Errorf(`Image had incorrect version: '%s'`, image.Version)
	}
}
