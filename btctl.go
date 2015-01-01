package main

import (
	"btctl/controller"
	"btctl/dmc"
	"btctl/ipt"
	"btctl/util"
	"flag"
	"os"
	"strings"
	"time"
)

type btctl struct {
	iptablesChain string
	controllerBinaryPath string
	users []string

	dmc dmc.Dmc
	controllers []controller.Controller
}

var log = util.MustGetLogger("btctl")

func (b *btctl) Init() error {
	for i := range b.controllers {
		b.controllers[i].Init()
	}

	return nil
}

func (b *btctl) OnTick() (err error) {
	stat, err := ipt.GetNetworkUsage(b.iptablesChain)
	if err != nil {
		return util.Error("Unable to get network usage stats: %s", err)
	}

	// TODO: remember last decision
	decision := b.dmc.OnNetworkUsageStat(time.Now(), stat)
	if decision != dmc.NO_DECISION {
		for i := range b.controllers {
			b.controllers[i].OnDecision(decision)
		}
	}

	return nil
}

func (b *btctl) Close() {
	for i := range b.controllers {
		b.controllers[i].Close()
	}
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

	util.Loop(&btctl{
		iptablesChain: *iptablesChain,
		controllerBinaryPath: *controllerBinaryPath,
		users: strings.Split(*users, ","),
		controllers: []controller.Controller{{}},
	}, time.Second)
}
