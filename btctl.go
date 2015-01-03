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

	dmc dmc.Dmc
	lastDecision dmc.Decision
	controllers []*controller.Controller
}

var log = util.MustGetLogger("btctl")

func (b *btctl) Init() error {
	return nil
}

func (b *btctl) OnTick() error {
	stat, err := ipt.GetNetworkUsage(b.iptablesChain)
	if err != nil {
		log.Error("Unable to get network usage stats: %s", err)
		return nil
	}

	decision := b.dmc.OnNetworkUsageStat(time.Now(), stat)
	if decision != dmc.NO_DECISION && b.lastDecision != decision {
		for _, c := range b.controllers {
			c.OnDecision(decision)
		}

		b.lastDecision = decision
	}

	return nil
}

func (b *btctl) Close() {
	for _, c := range b.controllers {
		c.Close()
	}
}

func main() {
	iptablesChain := flag.String("c", "network_usage_stat", "iptables chain name")
	controllerBinaryPath := flag.String("t", "transmission-controller", "transmission-controller binary")
	commaUsers := flag.String("u", "", "comma-separated list of users to control Transmission for")
	util.InitFlags()

	if flag.NArg() != 0 || *commaUsers == "" {
		flag.Usage()
		os.Exit(2)
	}

	util.MustInitLogging(true)

	users := strings.Split(*commaUsers, ",")
	controllers := make([]*controller.Controller, len(users))

	for i, user := range users {
		controllers[i] = controller.New(user, *controllerBinaryPath)
	}

	util.Loop(&btctl{iptablesChain: *iptablesChain, controllers: controllers}, dmc.NetworkUsageCollectionPeriod)
}
