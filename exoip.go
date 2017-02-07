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
var verbose = flag.Bool("v", false, "Log additional information")
var validate_config = flag.Bool("n", false, "Validate configuration and exit")
var watch_mode = flag.Bool("W", false, "Watchdog mode")
var associate_mode = flag.Bool("A", false, "Associate EIP and exit")
var dissociate_mode = flag.Bool("D", false, "Dissociate EIP and exit")
var peers stringslice
var reset_peers bool = false

func (s *stringslice) String() string {
	return strings.Join(*s, ",")
}

func (s *stringslice) Set(value string) error {
	if reset_peers {
		*s = make([]string, 0)
	}
	reset_peers = false
	peers := strings.Split(value, ",")
	for _, peer := range(peers) {
		*s = append(*s, peer)
	}
	return nil
}

type EnvEquiv struct {
	Env string
	Flag string
}

type EquivList []EnvEquiv

func ParseEnvironment() {

	env_flags := EquivList{
		EnvEquiv{Env: "IF_ADDRESS", Flag: "xi"},
		EnvEquiv{Env: "IF_BIND_TO", Flag: "l"},
		EnvEquiv{Env: "IF_DEAD_RATIO", Flag: "r"},
		EnvEquiv{Env: "IF_ADVERTISEMENT_INTERVAL", Flag: "t"},
		EnvEquiv{Env: "IF_HOST_PRIORITY", Flag: "P"},
		EnvEquiv{Env: "IF_EXOSCALE_API_KEY", Flag: "xk"},
		EnvEquiv{Env: "IF_EXOSCALE_API_SECRET", Flag: "xs"},
		EnvEquiv{Env: "IF_EXOSCALE_API_ENDPOINT", Flag: "xe"},
		EnvEquiv{Env: "IF_EXOSCALE_PEER_GROUP", Flag: "G"},
		EnvEquiv{Env: "IF_EXOSCALE_PEERS", Flag: "p"},
	}

	for _, env := range(env_flags) {
		v := os.Getenv(env.Env)
		if len(v) > 0 {
			flag.Set(env.Flag, v)
		}
	}

	reset_peers = true
}

func CheckConfiguration() {

	die := false

	if (*verbose) {
		exoip.Verbose = true
	}
	i := 0
	if (*watch_mode) {
		i++
	}
	if (*associate_mode) {
		i++
	}
	if (*dissociate_mode) {
		i++
	}

	if i != 1 {
		fmt.Fprintln(os.Stderr, "need exactly one of -A, -D, or -W")
		exoip.Logger.Info(fmt.Sprintf("invalid mode: need exactly one of -A, -D, or -W"))
		die = true
	}

	if len(*eip) == 0 {
		exoip.Logger.Crit("no Exoscale IP provided")
		fmt.Fprintln(os.Stderr, "no Exoscale IP provided")
		die = true
	}

	if *watch_mode {
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
	}

	if len(*exo_key) == 0 || len(*exo_endpoint) == 0 || len(*exo_secret) == 0 {
		exoip.Logger.Crit("insufficient API credentials")
		fmt.Fprintln(os.Stderr, "insufficient API credentials")
		die = true
	}
	if die {
		os.Exit(1)
	}
	if exoip.Verbose {
		fmt.Printf("exoip will watch over: %s\n", *eip)
		fmt.Printf("\tbind-address: %s\n", *address)
		fmt.Printf("\thost-priority: %d\n", *prio)
		fmt.Printf("\tadvertisement-interval: %d\n", *timer)
		fmt.Printf("\tdead-ratio: %d\n", *dead_ratio)
		fmt.Printf("\texoscale-api-key: %s\n", *exo_key)
		fmt.Printf("\texoscale-api-secret: %sXXXX\n", (*exo_secret)[0:2])
		fmt.Printf("\texoscale-api-endpoint: %s\n", *exo_endpoint)

		exoip.Logger.Info(fmt.Sprintf("exoip will watch over: %s\n", *eip))
		exoip.Logger.Info(fmt.Sprintf("\tbind-address: %s\n", *address))
		exoip.Logger.Info(fmt.Sprintf("\thost-priority: %d\n", *prio))
		exoip.Logger.Info(fmt.Sprintf("\tadvertisement-interval: %d\n", *timer))
		exoip.Logger.Info(fmt.Sprintf("\tdead-ratio: %d\n", *dead_ratio))
		exoip.Logger.Info(fmt.Sprintf("\texoscale-api-key: %s\n", *exo_key))
		exoip.Logger.Info(fmt.Sprintf("\texoscale-api-secret: %sXXXX\n", (*exo_secret)[0:2]))
		exoip.Logger.Info(fmt.Sprintf("\texoscale-api-endpoint: %s\n", *exo_endpoint))

		if (len(*exo_sg) > 0) {
			fmt.Printf("\texoscale-peer-group: %s\n", *exo_sg)
			exoip.Logger.Info(fmt.Sprintf("\texoscale-peer-group: %s\n", *exo_sg))
		} else {
			for _, p := range(peers) {
				fmt.Printf("\tpeer: %s\n", p)
				exoip.Logger.Info(fmt.Sprintf("\tpeer: %s\n", p))
			}
		}
	}


	if (*validate_config) {
		os.Exit(0)
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
	if (*associate_mode) {
		engine := exoip.NewEngine(ego, *eip)
		if err := engine.ObtainNic(engine.NicId); err != nil {
			fmt.Fprintf(os.Stderr, "%s", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	if (*dissociate_mode) {
		engine := exoip.NewEngine(ego, *eip)
		if err := engine.ReleaseMyNic(); err != nil {
			fmt.Fprintf(os.Stderr, "%s", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

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
		engine = exoip.NewWatchdogEngine(ego, *eip, *timer, *prio, *dead_ratio, sgpeers)
	} else {
		engine = exoip.NewWatchdogEngine(ego, *eip, *timer, *prio, *dead_ratio, peers)
	}
	exoip.Logger.Info("starting watchdog")
	go engine.NetworkAdvertise()
	engine.NetworkLoop(*address)

	os.Exit(0)
}
