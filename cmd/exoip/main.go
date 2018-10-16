package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/exoscale/egoscale"
	"github.com/exoscale/exoip"
)

// defaultPort used by exoip, can be changed via "-l"
const defaultPort = 12345

type stringslice []string

var timer = flag.Int("t", 1, "Advertisement interval in seconds")
var prio = flag.Int("P", 10, "Host priority (lowest wins)")
var address = flag.String("l", fmt.Sprintf(":%d", defaultPort), "Address to bind to")
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
var printVersion = flag.Bool("version", false, "Print version and quit")
var peers stringslice
var resetPeers = false

func (s *stringslice) String() string {
	return strings.Join(*s, ",")
}

func (s *stringslice) Set(value string) error { // nolint: unparam
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
			if err := flag.Set(env.Flag, v); err != nil {
				panic(err)
			}
		}
	}

	resetPeers = true
}

func setupLogger() {
	exoip.SetupLogger(*logStdout)

	if *verbose {
		exoip.Verbose = true
	}
}

func checkConfiguration() {
	die := !checkMode() || !checkEIP()
	if *watchMode {
		die = die || !checkPeerAndSecurityGroups() || !checkPeerDefinition() || !checkHostPriority()
	}

	die = die || !checkAPI()

	if die {
		os.Exit(1)
	}
}

func checkMode() bool {
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
		if _, err := fmt.Fprintln(os.Stderr, "need exactly one of -A, -D, or -W"); err != nil {
			panic(err)
		}
		exoip.Logger.Info("invalid mode: need exactly one of -A, -D, or -W")
		return false
	}

	return true
}

func checkEIP() bool {
	if len(*eip) == 0 {
		exoip.Logger.Crit("no Exoscale IP provided")
		if _, err := fmt.Fprintln(os.Stderr, "no Exoscale IP provided"); err != nil {
			panic(err)
		}
		return false
	}
	return true
}

func checkPeerAndSecurityGroups() bool {
	if len(peers) > 0 && len(*exoSecurityGroup) > 0 {
		exoip.Logger.Crit("ambiguous peer definition (-p and -G given)")
		if _, err := fmt.Fprintln(os.Stderr, "-p and -G options are exclusive"); err != nil {
			panic(err)
		}
		return false
	}
	return true
}

func checkPeerDefinition() bool {
	if len(peers) == 0 && len(*exoSecurityGroup) == 0 {
		exoip.Logger.Crit("need peer definition (either -p or -G)")
		if _, err := fmt.Fprintln(os.Stderr, "need peer definition (either -p or -G)"); err != nil {
			panic(err)
		}
		return false
	}
	return true
}

func checkHostPriority() bool {
	if *prio < 0 || *prio > 255 {
		exoip.Logger.Crit("invalid host priority (must be 0-255)")
		if _, err := fmt.Fprintln(os.Stderr, "invalid host priority (must be 0-255)"); err != nil {
			panic(err)
		}
		return false
	}

	return true
}

func checkAPI() bool {
	if len(*exoToken) == 0 || len(*csEndpoint) == 0 || len(*exoSecret) == 0 {
		exoip.Logger.Crit("insufficient API credentials")
		if _, err := fmt.Fprintln(os.Stderr, "insufficient API credentials"); err != nil {
			panic(err)
		}
		return false
	}
	return true
}

func printConfiguration() {
	fmt.Printf("exoip will watch over: %s\n", *eip)
	fmt.Printf("\tbind-address: %s\n", *address)
	fmt.Printf("\thost-priority: %d\n", *prio)
	fmt.Printf("\tadvertisement-interval: %d\n", *timer)
	fmt.Printf("\tdead-ratio: %d\n", *deadRatio)
	fmt.Printf("\texoscale-api-key: %s\n", *exoToken)
	fmt.Printf("\texoscale-api-secret: %sXXXX\n", (*exoSecret)[0:2])
	fmt.Printf("\texoscale-api-endpoint: %s\n", *csEndpoint)

	exoip.Logger.Info("exoip will watch over: %s\n", *eip)
	exoip.Logger.Info("\tbind-address: %s\n", *address)
	exoip.Logger.Info("\thost-priority: %d\n", *prio)
	exoip.Logger.Info("\tadvertisement-interval: %d\n", *timer)
	exoip.Logger.Info("\tdead-ratio: %d\n", *deadRatio)
	exoip.Logger.Info("\texoscale-api-key: %s\n", *exoToken)
	exoip.Logger.Info("\texoscale-api-secret: %sXXXX\n", (*exoSecret)[0:2])
	exoip.Logger.Info("\texoscale-api-endpoint: %s\n", *csEndpoint)

	if len(*exoSecurityGroup) > 0 {
		fmt.Printf("\texoscale-peer-group: %s\n", *exoSecurityGroup)
		exoip.Logger.Info("\texoscale-peer-group: %s\n", *exoSecurityGroup)
	} else {
		for _, p := range peers {
			fmt.Printf("\tpeer: %s\n", p)
			exoip.Logger.Info("\tpeer: %s\n", p)
		}
	}
}

func main() {

	var engine *exoip.Engine

	flag.Var(&peers, "p", "peers to communicate with")

	parseEnvironment()
	flag.Parse()

	if *printVersion {
		fmt.Printf("%v\n", exoip.Version)
		os.Exit(0)
	}

	// Sanity Checks
	setupLogger()
	checkConfiguration()
	if exoip.Verbose {
		printConfiguration()
	}
	if *validateConfig {
		os.Exit(0)
	}

	if (*instanceID) == "" {
		mserver, err := exoip.FindMetadataServer()
		if err != nil {
			if _, errP := fmt.Fprintln(os.Stderr, err); errP != nil {
				panic(errP)
			}
			os.Exit(1)
		}
		(*instanceID), err = exoip.FetchMetadata(mserver, "/latest/instance-id")
		if err != nil {
			if _, errP := fmt.Fprintln(os.Stderr, err); errP != nil {
				panic(errP)
			}
			os.Exit(1)
		}
	}

	ego := egoscale.NewClient(*csEndpoint, *exoToken, *exoSecret)

	ip := net.ParseIP(*eip)
	if ip == nil {
		if _, errP := fmt.Fprintln(os.Stderr, "not a valid IP Address"); errP != nil {
			panic(errP)
		}
		os.Exit(1)
	}

	if *associateMode {
		engine = exoip.NewEngine(ego, ip, *egoscale.MustParseUUID(*instanceID))
		if err := engine.ObtainNic(*engine.NicID); err != nil {
			if _, errP := fmt.Fprintln(os.Stderr, err); errP != nil {
				panic(errP)
			}
			os.Exit(1)
		}
		os.Exit(0)
	}

	if *disassociateMode {
		engine = exoip.NewEngine(ego, ip, *egoscale.MustParseUUID(*instanceID))
		if err := engine.ReleaseMyNic(); err != nil {
			if _, errP := fmt.Fprintln(os.Stderr, err); errP != nil {
				panic(errP)
			}
			os.Exit(1)
		}
		os.Exit(0)
	}

	if (*address)[0] == ':' {
		interfaces, err := net.Interfaces()
		if err == nil {
		outterfor:
			for _, iface := range interfaces {
				if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
					continue
				}

				addrs, err := iface.Addrs()
				if err != nil {
					continue
				}
				for _, addr := range addrs {
					ipAddress, _, err := net.ParseCIDR(addr.String())
					if err != nil {
						continue
					}
					if ip != nil && ipAddress.To4() != nil {
						*address = fmt.Sprintf("%s%s", ipAddress.String(), *address)
						exoip.Logger.Info("using IP address from %s", iface.Name)
						break outterfor
					}
				}
			}
		}
	}

	if len(*exoSecurityGroup) > 0 {
		if len(peers) > 0 {
			if _, err := fmt.Fprintln(os.Stderr, "-p and -G options are exclusive"); err != nil {
				panic(err)
			}
			os.Exit(1)
		}

		engine = exoip.NewEngineWatchdog(ego, *address, ip, *egoscale.MustParseUUID(*instanceID), *timer, *prio, *deadRatio, nil, *exoSecurityGroup)
	} else {
		engine = exoip.NewEngineWatchdog(ego, *address, ip, *egoscale.MustParseUUID(*instanceID), *timer, *prio, *deadRatio, peers, "")
	}

	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGTERM)
	signal.Notify(sigs, syscall.SIGINT)
	signal.Notify(sigs, syscall.SIGUSR1)
	signal.Notify(sigs, syscall.SIGUSR2)

	stopping := false

	go func() {
		for {
			sig := <-sigs
			exoip.Logger.Info("got sig: %+v", sig)
			if _, err := fmt.Fprintf(os.Stderr, "got sig: %+v\n", sig); err != nil {
				exoip.Logger.Crit(err.Error())
			}
			switch sig {
			case syscall.SIGUSR1:
				prio, err := engine.LowerPriority()
				if err != nil {
					exoip.Logger.Warning(err.Error())
				} else {
					exoip.Logger.Info("new priority: %d", prio)
				}
			case syscall.SIGUSR2:
				prio, err := engine.RaisePriority()
				if err != nil {
					exoip.Logger.Warning(err.Error())
				} else {
					exoip.Logger.Info("new priority: %d", prio)
				}
			default:
				if engine.State == exoip.StateMaster {
					stopping = true
					exoip.Logger.Info("releasing the Nic and stopping.")
					if _, err := fmt.Fprintln(os.Stderr, "releasing the Nic and stopping"); err != nil {
						exoip.Logger.Crit(err.Error())
					}
					if err := engine.ReleaseMyNic(); err != nil {
						exoip.Logger.Crit(err.Error())
						os.Exit(1)
					}
				}
				os.Exit(0)
			}
		}
	}()

	if err := engine.UpdatePeers(); err != nil {
		exoip.Logger.Crit(err.Error())
	}

	go func() {
		// update list of peers, every 5 minutes
		interval := 5 * time.Minute
		var elapsed time.Duration
		for !stopping {
			start := time.Now()
			if err := engine.UpdatePeers(); err != nil {
				exoip.Logger.Crit(err.Error())
			}

			if err := engine.UpdateNic(); err != nil {
				exoip.Logger.Crit(err.Error())
			}
			elapsed = time.Since(start)
			time.Sleep(interval - elapsed)
		}
	}()

	go func() {
		// pings our peers, every interval
		var elapsed time.Duration
		for !stopping {
			start := time.Now()
			if err := engine.PingPeers(); err != nil {
				exoip.Logger.Crit(err.Error())
			}
			elapsed = time.Since(start)
			if elapsed > engine.Interval {
				exoip.Logger.Warning("PingPeers took longer than allowed interval (%dms): %dms", engine.Interval/time.Millisecond, elapsed/time.Millisecond)
			}

			time.Sleep(engine.Interval - elapsed)
		}
	}()

	go func() {
		// act upon the peers state, every interval
		var elapsed time.Duration
		for !stopping {
			start := time.Now()
			engine.CheckState()
			elapsed = time.Since(start)
			if elapsed > engine.Interval {
				exoip.Logger.Warning("CheckState took longer than allowed interval (%dms): %dms", engine.Interval/time.Millisecond, elapsed/time.Millisecond)
			}

			time.Sleep(engine.Interval - elapsed)
		}
	}()

	exoip.Logger.Info("starting watchdog")
	err := engine.NetworkLoop()
	if err != nil {
		panic(err)
	}
	os.Exit(0)
}
