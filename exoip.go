package main

import (
	"flag"
	"os"
	"strings"
	"exoip"
)

type stringslice []string

var adv_timer = flag.Int("b", 1, "advertisement interval in seconds")
var prio = flag.Int("P", 10, "host priority (lowest wins)")
var vhid = flag.Int("i", 10, "server ID advertised")
var address = flag.String("l", ":12345", "address to listen to")
var dead_ratio = flag.Int("r", 3, "number of missed advertisements before promotion")
var peers stringslice

func (s *stringslice) String() string {
	return strings.Join(*s, ",")
}

func (s *stringslice) Set(value string) error {
	*s = append(*s, value)
	return nil
}

func main() {

	flag.Var(&peers, "p", "peers to communicate with")
	flag.Parse()
	exoip.SetupLogger()

	engine := exoip.NewEngine(*adv_timer, *vhid, *prio, *dead_ratio, peers)
	go engine.NetworkAdvertise()
	engine.NetworkLoop(*address)

	os.Exit(0)
}
