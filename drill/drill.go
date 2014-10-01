package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
	"btctl/ipt"
	"btctl/util"
)

func drill(path string) error {
	usageFile, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func () {
		err := usageFile.Close()
		if err != nil {
			log.Printf("Failed to close file '%s': %s.", usageFile.Name(), err)
		}
	}()

	scanner := bufio.NewScanner(usageFile)

	if err := readUsageData(scanner); err != nil {
		return err
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func readUsageData(scanner *bufio.Scanner) error {
	for scanner.Scan() {
		statTime, stat, err := getNetworkUsage(scanner.Text())
		if err != nil {
			return err
		}

		log.Println(statTime, stat)
	}

	return nil
}

func getNetworkUsage(data string) (statTime time.Time, stat ipt.NetworkUsage, err error) {
	// Format: date time packets bytes packets_diff bytes_diff
	usageInfo := strings.Split(data, " ")

	if len(usageInfo) != 6 {
		err = invalidDataErr(data)
		return
	}

	var e1, e2, e3 error
	statTime, e1 = time.Parse("2006.01.02 15:04:05", usageInfo[0] + " " + usageInfo[1])
	stat.Packets, e2 = strconv.ParseUint(usageInfo[2], 10, 64)
	stat.Bytes, e3 = strconv.ParseUint(usageInfo[3], 10, 64)

	if e1 != nil || e2 != nil || e3 != nil {
		err = invalidDataErr(data)
		return
	}

	return
}

func invalidDataErr(data string) error {
	return util.Error("Got an invalid network usage data: %s.", data)
}

func main() {
	usageFilePath := flag.String("source", "network-usage.txt", "collected network usage stat file")
	flag.Parse()

	if flag.NArg() != 0 {
		flag.Usage()
		os.Exit(2)
	}

	err := drill(*usageFilePath)

	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
