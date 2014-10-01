package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"time"
	"btctl/ipt"
	"btctl/util"
)


type collector struct {
	usageFile *os.File
	usageWriter *bufio.Writer

	stat ipt.NetworkUsage
}

func (c *collector) Init() error {
	var err error

	c.usageFile, err = os.Create("network-usage.txt")
	if err != nil {
		return err
	}

	c.usageWriter = bufio.NewWriter(c.usageFile)

	return nil
}

func (c *collector) Collect() error {
	stat, err := ipt.GetNetworkUsage()
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

func (c *collector) Close() {
	if c.usageWriter != nil {
		if err := c.usageWriter.Flush(); err != nil {
			log.Printf("Failed to write network usage stats: %s.", err)
		}
		c.usageWriter = nil
	}

	if c.usageFile != nil {
		if err := c.usageFile.Close(); err != nil {
			log.Printf("Failed to close file '%s': %s.", c.usageFile.Name(), err)
		}
		c.usageFile = nil
	}
}


type mainLoop struct {
	collector collector
}

func (loop *mainLoop) Init() (err error) {
	return loop.collector.Init()
}

func (loop *mainLoop) TickHandler() (err error) {
	err = loop.collector.Collect()
	return
}

func (loop *mainLoop) Close() {
	loop.collector.Close()
}


func main() {
	util.Loop(new(mainLoop), time.Second)
}
