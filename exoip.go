package main

import (
	"flag"
	"os"
	"fmt"
	"strings"
	"strconv"
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

func ParseEnvironment() {
	var x string
	x = os.Getenv("IF_ADDRESS")
	if len(x) > 0 {
		*eip = x
	}

	x = os.Getenv("IF_BIND_TO")
	if len(x) > 0 {
		*address = x
	}

	x = os.Getenv("IF_DEAD_RATIO")
	if len(x) > 0 {
		i, err := strconv.Atoi(x)
		if err == nil {
			*dead_ratio = i
		}
	}

	x = os.Getenv("IF_ADVERTISEMENT_INTERVAL")
	if len(x) > 0 {
		i, err := strconv.Atoi(x)
		if err == nil {
			*timer = i
		}
	}

	x = os.Getenv("IF_HOST_PRIORITY")
	if len(x) > 0 {
		i, err := strconv.Atoi(x)
		if err == nil {
			*prio = i
		}
	}

	x = os.Getenv("IF_EXOSCALE_API_KEY")
	if len(x) > 0 {
		*exo_key = x
	}
	x = os.Getenv("IF_EXOSCALE_API_SECRET")
	if len(x) > 0 {
		*exo_secret = x
	}
	x = os.Getenv("IF_EXOSCALE_API_ENDPOINT")
	if len(x) > 0 {
		*exo_endpoint = x
	}
	x = os.Getenv("IF_EXOSCALE_PEER_GROUP")
	if len(x) > 0 {
		*exo_sg = x
	}
	x = os.Getenv("IF_EXOSCALE_PEERS")
	if len(x) > 0 {
		peers = strings.Split(x, ",")
	}
}

func CheckConfiguration() {

	die := false
	if len(peers) > 0 && len(*exo_sg) > 0 {
		exoip.Logger.Crit("ambiguous peer definition (-p and -G given)")
		fmt.Fprintln(os.Stderr, "-p and -G options are exclusive")
		die = true
	}

	if len(peers) == 0 && len(*exo_sg) == 0 {
		exoip.Logger.Crit("need peer definition (either -p or -G)")
		fmt.Fprintln(os.Stderr, "need peer definition (either -p or -G)")
		die = true
	}

	if *prio < 0 || *prio > 255 {
		exoip.Logger.Crit("invalid host priority (must be 0-255)")
		fmt.Fprintln(os.Stderr, "invalid host priority (must be 0-255)")
		die = true
	}

	if len(*exo_key) == 0 || len(*exo_endpoint) == 0 || len(*exo_secret) == 0 {
		exoip.Logger.Crit("insufficient API credentials")
		fmt.Fprintln(os.Stderr, "insufficient API credentials")
		die = true
	}
	if die {
		os.Exit(1)
	}
}

func main() {

	var engine *exoip.Engine

	flag.Var(&peers, "p", "peers to communicate with")

	ParseEnvironment()
	flag.Parse()

	// Sanity Checks
	exoip.SetupLogger()
	CheckConfiguration()

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
