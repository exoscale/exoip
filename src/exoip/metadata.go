package exoip

import (
	"io/ioutil"
	"fmt"
	"net/http"
)

func FindMetadataServer() (string, error) {
	return "159.100.241.1", nil
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
