package main

import (
	"fmt"
	"btctl/ipt"
)

func main() {
	usage, err := ipt.GetUsage()
	if err != nil {
		fmt.Println("ERROR:", err)
		return
	}
	fmt.Println(usage)
}
