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
	usageFile *os.File
}

func (loop *MainLoop) Init() (err error) {
	loop.usageFile, err = os.Create("network-usage.txt")
	return
}

func (loop *MainLoop) TickHandler() error {
	return collect(loop.usageFile)
}

func (loop *MainLoop) Close() {
	if loop.usageFile != nil {
		if err := loop.usageFile.Close(); err != nil {
			log.Printf("Failed to close file '%s': %s.", loop.usageFile.Name(), err)
		}
		loop.usageFile = nil
	}
}

func collect(usageFile *os.File) error {
	packets, bytes, err := ipt.GetNetworkUsage()
	if err != nil {
		return util.Error("Unable to get network usage stats: %s", err)
	}

	_, err = fmt.Fprintln(usageFile, time.Now().Format("2006.01.02 15:04:05"), packets, bytes)
	if err != nil {
		return util.Error("Failed to write network usage stats: %s.", err)
	}

	return nil
}

func main() {
	var mainLoop MainLoop
	util.Loop(&mainLoop, time.Second)
}
