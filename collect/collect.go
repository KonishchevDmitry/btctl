package main

import (
	"fmt"
	"log"
	"os"
	"time"
	"btctl/ipt"
	"btctl/util"
)

type MainLoop struct {
	packets uint64
	bytes uint64
	usageFile *os.File
}

func (loop *MainLoop) Init() (err error) {
	loop.usageFile, err = os.Create("network-usage.txt")
	return
}

func (loop *MainLoop) TickHandler() (err error) {
	loop.packets, loop.bytes, err = collect(loop.packets, loop.bytes, loop.usageFile)
	return
}

func (loop *MainLoop) Close() {
	if loop.usageFile != nil {
		if err := loop.usageFile.Close(); err != nil {
			log.Printf("Failed to close file '%s': %s.", loop.usageFile.Name(), err)
		}
		loop.usageFile = nil
	}
}

func collect(prevPackets uint64, prevBytes uint64, usageFile *os.File) (uint64, uint64, error) {
	packets, bytes, err := ipt.GetNetworkUsage()
	if err != nil {
		return prevPackets, prevBytes, util.Error("Unable to get network usage stats: %s", err)
	}

	_, err = fmt.Fprintln(usageFile, time.Now().Format("2006.01.02 15:04:05"), packets, bytes, packets - prevPackets, bytes - prevBytes)
	if err != nil {
		return packets, bytes, util.Error("Failed to write network usage stats: %s.", err)
	}

	return packets, bytes, nil
}

func main() {
	var mainLoop MainLoop
	util.Loop(&mainLoop, time.Second)
}
