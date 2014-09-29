package ipt

import (
	"btctl/util"
)

func GetUsage() (string, error) {
	stdout, err := util.RunCmd("date", "ERROR")
	return stdout, err
}
