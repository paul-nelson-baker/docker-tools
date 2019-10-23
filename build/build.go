package build

import (
	"github.com/jhoonb/archivex"
	"io/ioutil"
	"log"
	"strings"
)

type TarFileConsumer func(tar *archivex.TarFile) error

func CreateTarBuildContext(tarFileConsumers ...TarFileConsumer) (string, error) {
	tarTempFile, err := ioutil.TempFile("", "*.tar")
	if err != nil {
		return "", err
	}
	tarFile := new(archivex.TarFile)
	defer func() {
		_ = tarFile.Close()
	}()
	tarTempFilename := tarTempFile.Name()
	if err := tarFile.Create(tarTempFilename); err != nil {
		return "", err
	}
	for _, tarFileConsumer := range tarFileConsumers {
		if err := tarFileConsumer(tarFile); err != nil {
			return "", err
		}
	}
	return tarTempFilename, nil
}

type EventFunc func(event Event) error

type Event struct {
	Stream string `json:"stream"`
}

// Logs any status or progress changes to the console via `log.Println`
func LoggingFunc(event Event) error {
	if output := strings.TrimSpace(event.Stream); output != "" {
		log.Println(output)
	}
	return nil
}
