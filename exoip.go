package main

import (
	"flag"
	"os"
	"fmt"
	"strings"
	"exoip"
	"github.com/pyr/egoscale/src/egoscale"
)

type stringslice []string

var timer = flag.Int("t", 1, "Advertisement interval in seconds")
var prio = flag.Int("P", 10, "Host priority (lowest wins)")
var address = flag.String("l", fmt.Sprintf(":%d", exoip.DefaultPort), "Address to bind to")
var dead_ratio = flag.Int("r", 3, "Dead ratio")
var exo_key = flag.String("xk", "", "Exoscale API Key")
var exo_secret = flag.String("xs", "", "Exoscale API Secret")
var exo_endpoint = flag.String("xe", "https://api.exoscale.ch/compute", "Exoscale API Endpoint")
var exo_sg = flag.String("G", "", "Exoscale Security Group to use to create list of peers")
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
	var engine *exoip.Engine

	ego := egoscale.NewClient(*exo_endpoint, *exo_key, *exo_secret)
	if (len(*exo_sg) > 0) {
		if len(peers) > 0 {
			fmt.Fprintln(os.Stderr, "-p and -G options are exclusive")
			os.Exit(1)
		}
		sgpeers, err := exoip.GetSecurityGroupPeers(ego, *exo_sg)
		if err != nil {
			exoip.Logger.Crit("cannot build peer list from security-group")
			fmt.Fprintf(os.Stderr, "cannot build peer list from security-group: %s\n", err)
			os.Exit(1)
		}
		engine = exoip.NewEngine(ego, *eip, *timer, *prio, *dead_ratio, sgpeers)
	} else {
		engine = exoip.NewEngine(ego, *eip, *timer, *prio, *dead_ratio, peers)
	}
	exoip.Logger.Info("starting up")
	go engine.NetworkAdvertise()
	engine.NetworkLoop(*address)

	os.Exit(0)
}
