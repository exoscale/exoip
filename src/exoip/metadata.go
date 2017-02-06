package exoip

import (
	"io/ioutil"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
)

func FindMetadataServer() (string, error) {

	out, err :=  exec.Command("ip", "route", "list").Output()
	if err != nil {
		fmt.Println("could not execute:", err)
	}
	lines := strings.Split(string(out), "\n")
	for _, line := range(lines) {
		if strings.HasPrefix(line, "default via ") {
			params := strings.Split(line, " ")
			return params[2], nil
		}
	}
	return "", fmt.Errorf("could not find metadata server")
}

func FetchMetadata(mserver string, path string) (string, error) {

	url := fmt.Sprintf("http://%s/%s", mserver, path)
	resp, err := http.Get(url)
	defer resp.Body.Close()
	if err != nil {
		return "", err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}
