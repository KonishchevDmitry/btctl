package main

import (
	"fmt"
	"log"
	"time"
	"btctl/ipt"
)

func main() {
	time.Sleep(60 * time.Second)
	packets, bytes, err := ipt.GetNetworkUsage()
	if err != nil {
		log.Printf("Unable to get network usage stats: %s", err)
		return
	}
	fmt.Println(time.Now().Unix(), packets, bytes)
}
