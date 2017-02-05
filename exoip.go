package main

import (
	"flag"
	"os"
	"strings"
	"exoip"
	"github.com/pyr/egoscale/src/egoscale"
)

type stringslice []string

var timer = flag.Int("t", 1, "Advertisement interval in seconds")
var prio = flag.Int("P", 10, "Host priority (lowest wins)")
var address = flag.String("l", ":12345", "Address to bind to")
var dead_ratio = flag.Int("r", 3, "Dead ratio")
var exo_key = flag.String("xk", "", "Exoscale API Key")
var exo_secret = flag.String("xs", "", "Exoscale API Secret")
var exo_endpoint = flag.String("xe", "https://api.exoscale.ch/compute", "Exoscale API Endpoint")
var eip = flag.String("xi", "", "Exoscale Elastic IP to watch over")
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

	ego := egoscale.NewClient(*exo_endpoint, *exo_key, *exo_secret)
	engine := exoip.NewEngine(ego, *eip, *timer, *prio, *dead_ratio, peers)
	go engine.NetworkAdvertise()
	engine.NetworkLoop(*address)

	os.Exit(0)
}
