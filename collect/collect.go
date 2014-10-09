package main

import (
	"btctl/ipt"
	"btctl/util"
	"bufio"
	"flag"
	"fmt"
	"os"
	"time"
)

type collector struct {
	iptablesChain string

	stat ipt.NetworkUsage

	usageFilePath string
	usageFile *os.File
	usageWriter *bufio.Writer
}

var log = util.MustGetLogger("collect")

func (c *collector) Init() error {
	var err error

	c.usageFile, err = os.Create(c.usageFilePath)
	if err != nil {
		return err
	}

	c.usageWriter = bufio.NewWriter(c.usageFile)

	return nil
}

func (c *collector) OnTick() (err error) {
	return c.collect()
}

func (c *collector) Close() {
	if c.usageWriter != nil {
		if err := c.usageWriter.Flush(); err != nil {
			log.Error("Failed to write network usage stats: %s.", err)
		}
		c.usageWriter = nil
	}

	if c.usageFile != nil {
		if err := c.usageFile.Close(); err != nil {
			log.Error("Failed to close file '%s': %s.", c.usageFile.Name(), err)
		}
		c.usageFile = nil
	}
}

func (c *collector) collect() error {
	stat, err := ipt.GetNetworkUsage(c.iptablesChain)
	if err != nil {
		return util.Error("Unable to get network usage stats: %s", err)
	}

	_, err = fmt.Fprintln(c.usageWriter, time.Now().Format("2006.01.02 15:04:05"),
		stat.Packets, stat.Bytes, stat.Packets - c.stat.Packets, stat.Bytes - c.stat.Bytes)

	if err != nil {
		return util.Error("Failed to write network usage stats: %s.", err)
	}

	c.stat = stat

	return nil
}

func main() {
	iptablesChain := flag.String("c", "network_usage_stat", "iptables chain name")
	usageFilePath := flag.String("o", "network-usage.txt", "path to write network usage statistics to")
	util.InitFlags()

	util.MustInitLogging(false)

	loop := &collector{iptablesChain: *iptablesChain, usageFilePath: *usageFilePath}
	util.Loop(loop, time.Second)
}
