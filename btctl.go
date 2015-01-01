package main

import (
	"btctl/ipt"
	"btctl/util"
	"flag"
	"os"
	"strings"
	"time"
)

type controller struct {
	iptablesChain string
	controllerBinaryPath string
	users []string

	stat ipt.NetworkUsage
}

var log = util.MustGetLogger("btctl")

func (c *controller) Init() error {
	return nil
}

func (c *controller) OnTick() (err error) {
	stat, err := ipt.GetNetworkUsage(c.iptablesChain)
	if err != nil {
		return util.Error("Unable to get network usage stats: %s", err)
	}

	c.stat = stat

	return nil
}

func (c *controller) Close() {
}

func main() {
	iptablesChain := flag.String("c", "network_usage_stat", "iptables chain name")
	controllerBinaryPath := flag.String("t", "transmission-controller", "transmission-controller binary")
	users := flag.String("u", "", "comma-separated list of users to control Transmission for")
	util.InitFlags()

	if flag.NArg() != 0 || *users == "" {
		flag.Usage()
		os.Exit(2)
	}

	util.MustInitLogging(false)

	util.Loop(&controller{
		iptablesChain: *iptablesChain,
		controllerBinaryPath: *controllerBinaryPath,
		users: strings.Split(*users, ","),
	}, time.Second)
}
