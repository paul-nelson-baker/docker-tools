package image

import "fmt"

// Used to reference a docker image as a 3-tuple
type LazyImage struct {
	Library string
	Name    string
	Version string
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
