// +build !windows

package main

import (
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"runtime"
)

func getOSInfo() (oSystem string, err error){
	switch runtime.GOOS {
	case "linux":
		out, err := exec.Command("hostnamectl").Output()
		if err != nil {
			return "", err
		}
		osRe := regexp.MustCompile(`(?m)Operating System: (.+$)`)
		match := osRe.FindStringSubmatch(string(out))
		if len(match[1]) > 1 {
			return match[1], nil
		}
		return "", errors.New("Failed to fetch os Info")
	case "darwin":
		out, err := exec.Command("sw_vers").Output()
		if err != nil {
			return "", err
		}
		darwinNameRe := regexp.MustCompile(`(?m)ProductName:\W(.+$)`)
		darwinVerRe := regexp.MustCompile(`(?m)ProductVersion:\W(.+$)`)
		nameMatch := darwinNameRe.FindStringSubmatch(string(out))
		verMatch := darwinVerRe.FindStringSubmatch(string(out))
		if len(nameMatch) > 1 && len(verMatch) > 1 {
			oSystem = fmt.Sprintf("%s (%s)", nameMatch[1], verMatch[1])
			return oSystem, nil
		}
		return "", nil
	default:
		return "", errors.New("unable to determine operating system")
	}
}
