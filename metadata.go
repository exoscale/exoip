package exoip

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"strings"
)

// FindMetadataServer finds the Virtual Router / Metadata server IP address
func FindMetadataServer() (string, error) {

	out, err := exec.Command("ip", "route", "list").Output()
	if err != nil {
		fmt.Println("could not execute:", err)
	}
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "default via ") {
			params := strings.Split(line, " ")
			return params[2], nil
		}
	}
	return "", fmt.Errorf("could not find metadata server")
}

// FetchMetadata reads the metadata from the Virtual Router
func FetchMetadata(mserver string, path string) (string, error) {
	url := fmt.Sprintf("http://%s/%s", mserver, path)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
