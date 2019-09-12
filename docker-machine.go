package docker_tools

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/client"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"regexp"
	"strings"
)

// A function that will either return a
type DockerClientSupplier func() (*client.Client, error)

func GetDockerClient() (*client.Client, error) {
	return GetDockerClientOrFallback(client.NewEnvClient, GetDockerMachineClient)
}

func GetDockerClientOrFallback(dockerClientSuppliers ...DockerClientSupplier) (dockerClient *client.Client, err error) {
	if len(dockerClientSuppliers) == 0 {
		err = fmt.Errorf("no docker suppliers were provided")
		return
	}
	for _, supplier := range dockerClientSuppliers {
		if dockerClient, err = supplier(); err == nil {
			return
		}
	}
	return
}

func GetDockerMachineClient() (*client.Client, error) {
	dockerMachineConfig, err := getDockerMachineConfig()
	if err != nil {
		return nil, err
	}
	tlsConfig, err := loadDockerMachineCerts(dockerMachineConfig.tlsCaCert, dockerMachineConfig.tlsCert, dockerMachineConfig.tlsKey)
	if err != nil {
		return nil, err
	}
	transport := &http.Transport{TLSClientConfig: tlsConfig}
	httpClient := &http.Client{Transport: transport}
	apiVersion, err := determineApiVersion(dockerMachineConfig.url, httpClient)
	if err != nil {
		return nil, err
	}
	return client.NewClient(dockerMachineConfig.url, apiVersion, httpClient, map[string]string{})
}

func determineApiVersion(host string, client *http.Client) (string, error) {
	regex := regexp.MustCompile("^tcp")
	host = regex.ReplaceAllString(host, "https")
	response, err := client.Get(host + "/version")
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	bytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	responseBody := map[string]interface{}{}
	err = json.Unmarshal(bytes, &responseBody)
	if err != nil {
		return "", fmt.Errorf("could not determine ApiVersion (%s): %+v", err, string(bytes))
	}
	if apiVersion, ok := responseBody["ApiVersion"].(string); !ok {
		return "", fmt.Errorf("could not determine ApiVersion: %+v", responseBody)
	} else {
		return apiVersion, nil
	}
}

func getDockerMachineConfig() (DockerMachineConfig, error) {
	items, err := getOutputItemsFromDockerMachine("config")
	if err != nil {
		return DockerMachineConfig{}, err
	}
	config := parseDockerMachineOutput(items)
	return config, nil
}

// Important references:
// 	https://forfuncsake.github.io/post/2017/08/trust-extra-ca-cert-in-go-app/
// 	https://medium.com/@sirsean/mutually-authenticated-tls-from-a-go-client-92a117e605a1
func loadDockerMachineCerts(caCertFilePath, certFilePath, keyFilePath string) (*tls.Config, error) {
	// Append our certificate-authority cert to the system pool
	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}
	certs, err := ioutil.ReadFile(caCertFilePath)
	if err != nil {
		return nil, err
	}
	if ok := rootCAs.AppendCertsFromPEM(certs); !ok {
		return nil, fmt.Errorf("no certs appended, using system certs only")
	}
	// Get the actual client certificate
	certificate, err := tls.LoadX509KeyPair(certFilePath, keyFilePath)
	if err != nil {
		return nil, err
	}
	config := &tls.Config{
		InsecureSkipVerify: false,
		RootCAs:            rootCAs,
		Certificates:       []tls.Certificate{certificate},
	}
	return config, nil
}

func getOutputItemsFromDockerMachine(args ...string) ([]string, error) {
	command := exec.Command("docker-machine", args...)
	output := bytes.Buffer{}
	command.Stdout = &output
	err := command.Run()
	if err != nil {
		return []string{}, err
	}
	return strings.Split(output.String(), "\n"), nil
}

func parseDockerMachineOutput(outputItems []string) (config DockerMachineConfig) {
	for _, line := range outputItems {
		scrubValue := func(value string) string {
			value = strings.TrimLeft(value, `"`)
			value = strings.TrimRight(value, `"`)
			value = strings.ReplaceAll(value, `\\`, `\`)
			return value
		}
		stuff := strings.SplitN(strings.TrimLeft(line, "-"), "=", 2)
		if len(stuff) == 0 {
			continue
		}
		key := stuff[0]
		switch key {
		case "tlsverify":
			config.tlsVerify = true
		case "tlscacert":
			config.tlsCaCert = scrubValue(stuff[1])
		case "tlscert":
			config.tlsCert = scrubValue(stuff[1])
		case "tlskey":
			config.tlsKey = scrubValue(stuff[1])
		case "H":
			config.url = scrubValue(stuff[1])
		default:
			log.Println("Unknown config:", line)
		}
	}
	return
}

type DockerMachineConfig struct {
	url       string
	tlsVerify bool
	tlsCaCert string
	tlsCert   string
	tlsKey    string
}
