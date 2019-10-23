package pull

import (
	"fmt"
	"github.com/paul-nelson-baker/docker-tools/image"
	"log"
	"strings"
)

type EventFunc func(lazyImage image.LazyImage, event Event) error

// Logs any status or progress changes to the console via `log.Println`
func LoggingFunc(lazyImage image.LazyImage, event Event) error {
	if event.Status != "" || event.Progress != "" {
		output := strings.TrimSpace(fmt.Sprintf("%s %s", event.Status, event.Progress))
		log.Println(output)
	}
	return nil
}

// An event that will be async returned to the client from the docker
// daemon as it pulls a remote image
type Event struct {
	Status         string `json:"status"`
	Error          string `json:"error"`
	Progress       string `json:"progress"`
	ProgressDetail struct {
		Current int `json:"current"`
		Total   int `json:"total"`
	} `json:"progressDetail"`
}
