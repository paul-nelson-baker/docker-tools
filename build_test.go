package dockerTools

import (
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/jhoonb/archivex"
	"github.com/paul-nelson-baker/docker-tools/build"
	"os"
	"testing"
)

func TestCreateTarBuildContext(t *testing.T) {
	tarBuildContextFilename, err := build.CreateTarBuildContext(func(tar *archivex.TarFile) error {
		dockerfile, err := os.Open("testdata/Dockerfile")
		if err != nil {
			return err
		} else if dockerfile == nil {
			return fmt.Errorf("dockerfile was nil")
		}
		defer dockerfile.Close()
		info, err := dockerfile.Stat()
		if err != nil {
			return err
		}
		return tar.Add("Dockerfile", dockerfile, info)
	})
	if err != nil {
		t.Errorf("couldn't build docker tar-file context: %v", err)
		t.Fail()
	}

	dockerClient, err := NewLazyClient()
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	options := types.ImageBuildOptions{
		PullParent:     true,
		Tags:           []string{"docker-build-test:latest"},
		SuppressOutput: false,
		Remove:         true,
		ForceRemove:    true,
		Squash:         false,
	}
	if err := dockerClient.LazyBuildArchiveCallback(tarBuildContextFilename, options, build.LoggingFunc); err != nil {
		t.Error(err)
	}
}
