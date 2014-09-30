package ipt

import (
	"regexp"
	"strconv"
	"strings"

	"btctl/util"
)

func GetNetworkUsage() (packets uint64, bytes uint64, err error) {
	var stdout string
	var chainPackets, chainBytes uint64

	for _, iptables := range(allIptables) {
		stdout, err = iptables("-L", "-nvx")
		if err != nil {
			return
		}

		chainPackets, chainBytes, err = getNetworkUsage(stdout, "network_usage_stats")
		if err != nil {
			return
		}

		packets += chainPackets
		bytes += chainBytes
	}

	return
}

func getNetworkUsage(stdout string, statChain string) (packets uint64, bytes uint64, err error) {
	ruleStatRe, err := regexp.Compile(`^\s*(\d+)\s+(\d+)\s+([^[:space:]]+)`)
	if err != nil {
		return
	}

	stdout = strings.Trim(stdout, "\n")

	for _, block := range(strings.Split(stdout, "\n\n")) {
		lines := strings.Split(strings.Trim(block, "\n"), "\n")

		if len(lines) < 2 || ! strings.HasPrefix(lines[0], "Chain") {
			err = parseError(lines[0])
			return
		}

		if len(lines) < 3 {
			continue
		}

		for _, ruleStat := range(lines[2:]) {
			matches := ruleStatRe.FindStringSubmatch(ruleStat)
			if matches == nil {
				err = parseError(ruleStat)
				return
			}

			if matches[3] != statChain {
				continue
			}

			rulePackets, err := strconv.ParseUint(matches[1], 10, 64)
			ruleBytes, err := strconv.ParseUint(matches[2], 10, 64)

			if err != nil {
				return packets, bytes, err
			}

			packets += rulePackets
			bytes += ruleBytes
		}
	}

	return
}

func ip4tables(args ...string) (stdout string, err error) {
	return runIptables("iptables", args...)
}

func ip6tables(args ...string) (stdout string, err error) {
	return runIptables("ip6tables", args...)
}

func runIptables(executable string, args ...string) (stdout string, err error) {
	stdout, err = util.RunCmd(executable, args...)
	if err != nil {
		err = util.Error("%s execution error: %s", executable, err)
	}
	return
}

func parseError(errorLine string) error {
	return util.Error("Failed to parse iptables output at line `%s`.", errorLine)
}

var allIptables = [](func(...string) (string, error)){ip4tables, ip6tables}
