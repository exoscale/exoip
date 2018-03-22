package exoip

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/vishvananda/netlink"
)

// FindMetadataServer finds the Virtual Router / Metadata server IP address
func FindMetadataServer() (string, error) {
	links, err := netlink.LinkList()
	if err != nil {
		return "", err
	}

	for _, link := range links {
		if link.Attrs().EncapType != "ether" {
			continue
		}

		routes, err := netlink.RouteList(link, netlink.FAMILY_V4)
		if err != nil {
			return "", err
		}

		for _, route := range routes {
			// The default route has Dst set to nil
			if route.Dst == nil && route.Gw != nil {
				return route.Gw.String(), nil
			}
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
