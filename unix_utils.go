// +build !windows

package main

import (
	"errors"
	"os/exec"
	"regexp"
	"runtime"
)

func getOSInfo() (oSystem string, version string, err error){
	switch runtime.GOOS {
	case "darwin":
		out, err := exec.Command("sw_vers").Output()
		if err != nil {
			return "", "", err
		}
		darwinRe := regexp.MustCompile(`ProductVersion:\W([0-9]*\.?[0-9?]*\.?[0-9?]*)`)
		match := darwinRe.FindStringSubmatch(string(out))
		return "", match[1], nil
	default:
		return "", "", errors.New("unable to determine operating system")
	}
}
