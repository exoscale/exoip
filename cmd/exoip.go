package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/exoscale/egoscale"
	"github.com/exoscale/exoip"
)

type stringslice []string

var timer = flag.Int("t", 1, "Advertisement interval in seconds")
var prio = flag.Int("P", 10, "Host priority (lowest wins)")
var address = flag.String("l", fmt.Sprintf(":%d", exoip.DefaultPort), "Address to bind to")
var deadRatio = flag.Int("r", 3, "Dead ratio")
var exoToken = flag.String("xk", "", "Exoscale API Key")
var exoSecret = flag.String("xs", "", "Exoscale API Secret")
var csEndpoint = flag.String("xe", "https://api.exoscale.ch/compute", "Exoscale API Endpoint")
var exoSecurityGroup = flag.String("G", "", "Exoscale Security Group to use to create list of peers")
var eip = flag.String("xi", "", "Exoscale Elastic IP to watch over")
var instanceID = flag.String("i", "", "Exoscale Instance ID of oneself")
var verbose = flag.Bool("v", false, "Log additional information")
var validateConfig = flag.Bool("n", false, "Validate configuration and exit")
var watchMode = flag.Bool("W", false, "Watchdog mode")
var associateMode = flag.Bool("A", false, "Associate EIP and exit")
var disassociateMode = flag.Bool("D", false, "Dissociate EIP and exit")
var logStdout = flag.Bool("O", false, "Do not log to syslog, use standard output")
var peers stringslice
var resetPeers = false

func (s *stringslice) String() string {
	return strings.Join(*s, ",")
}

func (s *stringslice) Set(value string) error {
	if resetPeers {
		*s = make([]string, 0)
	}
	resetPeers = false
	peers := strings.Split(value, ",")
	for _, peer := range peers {
		*s = append(*s, peer)
	}
	return nil
}

type envEquiv struct {
	Env  string
	Flag string
}

type equivList []envEquiv

func parseEnvironment() {

	envFlags := equivList{
		envEquiv{Env: "IF_ADDRESS", Flag: "xi"},
		envEquiv{Env: "IF_BIND_TO", Flag: "l"},
		envEquiv{Env: "IF_DEAD_RATIO", Flag: "r"},
		envEquiv{Env: "IF_ADVERTISEMENT_INTERVAL", Flag: "t"},
		envEquiv{Env: "IF_HOST_PRIORITY", Flag: "P"},
		envEquiv{Env: "IF_EXOSCALE_API_KEY", Flag: "xk"},
		envEquiv{Env: "IF_EXOSCALE_API_SECRET", Flag: "xs"},
		envEquiv{Env: "IF_EXOSCALE_API_ENDPOINT", Flag: "xe"},
		envEquiv{Env: "IF_EXOSCALE_PEER_GROUP", Flag: "G"},
		envEquiv{Env: "IF_EXOSCALE_INSTANCE_ID", Flag: "i"},
		envEquiv{Env: "IF_EXOSCALE_PEERS", Flag: "p"},
	}

	for _, env := range envFlags {
		v := os.Getenv(env.Env)
		if len(v) > 0 {
			flag.Set(env.Flag, v)
		}
	}

	resetPeers = true
}

func checkConfiguration() {

	die := false

	if *verbose {
		exoip.Verbose = true
	}
	i := 0
	if *watchMode {
		i++
	}
	if *associateMode {
		i++
	}
	if *disassociateMode {
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

	if *watchMode {
		if len(peers) > 0 && len(*exoSecurityGroup) > 0 {
			exoip.Logger.Crit("ambiguous peer definition (-p and -G given)")
			fmt.Fprintln(os.Stderr, "-p and -G options are exclusive")
			die = true
		}

		if len(peers) == 0 && len(*exoSecurityGroup) == 0 {
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

	if len(*exoToken) == 0 || len(*csEndpoint) == 0 || len(*exoSecret) == 0 {
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
		fmt.Printf("\tdead-ratio: %d\n", *deadRatio)
		fmt.Printf("\texoscale-api-key: %s\n", *exoToken)
		fmt.Printf("\texoscale-api-secret: %sXXXX\n", (*exoSecret)[0:2])
		fmt.Printf("\texoscale-api-endpoint: %s\n", *csEndpoint)

		exoip.Logger.Info(fmt.Sprintf("exoip will watch over: %s\n", *eip))
		exoip.Logger.Info(fmt.Sprintf("\tbind-address: %s\n", *address))
		exoip.Logger.Info(fmt.Sprintf("\thost-priority: %d\n", *prio))
		exoip.Logger.Info(fmt.Sprintf("\tadvertisement-interval: %d\n", *timer))
		exoip.Logger.Info(fmt.Sprintf("\tdead-ratio: %d\n", *deadRatio))
		exoip.Logger.Info(fmt.Sprintf("\texoscale-api-key: %s\n", *exoToken))
		exoip.Logger.Info(fmt.Sprintf("\texoscale-api-secret: %sXXXX\n", (*exoSecret)[0:2]))
		exoip.Logger.Info(fmt.Sprintf("\texoscale-api-endpoint: %s\n", *csEndpoint))

		if len(*exoSecurityGroup) > 0 {
			fmt.Printf("\texoscale-peer-group: %s\n", *exoSecurityGroup)
			exoip.Logger.Info(fmt.Sprintf("\texoscale-peer-group: %s\n", *exoSecurityGroup))
		} else {
			for _, p := range peers {
				fmt.Printf("\tpeer: %s\n", p)
				exoip.Logger.Info(fmt.Sprintf("\tpeer: %s\n", p))
			}
		}
	}

	if *validateConfig {
		os.Exit(0)
	}
}

func main() {

	var engine *exoip.Engine

	flag.Var(&peers, "p", "peers to communicate with")

	parseEnvironment()
	flag.Parse()

	// Sanity Checks
	exoip.SetupLogger(*logStdout)
	checkConfiguration()

	if (*instanceID) == "" {
		mserver, err := exoip.FindMetadataServer()
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			os.Exit(1)
		}
		(*instanceID), err = exoip.FetchMetadata(mserver, "/latest/instance-id")
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			os.Exit(1)
		}
	}

	ego := egoscale.NewClient(*csEndpoint, *exoToken, *exoSecret)
	if *associateMode {
		engine := exoip.NewEngine(ego, *eip, *instanceID)
		if err := engine.ObtainNic(engine.NicID); err != nil {
			fmt.Fprint(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	if *disassociateMode {
		engine := exoip.NewEngine(ego, *eip, *instanceID)
		if err := engine.ReleaseMyNic(); err != nil {
			fmt.Fprint(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	if len(*exoSecurityGroup) > 0 {
		if len(peers) > 0 {
			fmt.Fprintln(os.Stderr, "-p and -G options are exclusive")
			os.Exit(1)
		}
		sgpeers, err := exoip.GetSecurityGroupPeers(ego, *exoSecurityGroup)
		if err != nil {
			exoip.Logger.Crit("cannot build peer list from security-group")
			fmt.Fprintf(os.Stderr, "cannot build peer list from security-group: %s\n", err)
			os.Exit(1)
		}
		engine = exoip.NewWatchdogEngine(ego, *eip, *instanceID, *timer, *prio, *deadRatio, sgpeers)
	} else {
		engine = exoip.NewWatchdogEngine(ego, *eip, *instanceID, *timer, *prio, *deadRatio, peers)
	}
	exoip.Logger.Info("starting watchdog")
	go engine.NetworkAdvertise()
	engine.NetworkLoop(*address)

	os.Exit(0)
}
