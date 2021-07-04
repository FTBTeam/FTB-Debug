// +build windows

package main

import (
	"errors"
	"github.com/StackExchange/wmi"
	"runtime"
)

type Win32_OperatingSystem struct {
	Caption string
	Version string
}

func getOSInfo() (oSystem string, version string, err error){
	switch runtime.GOOS {
	case "windows":
		var dst []Win32_OperatingSystem

		q := wmi.CreateQuery(&dst, "")
		err = wmi.Query(q, &dst)
		if err != nil {
			return "", "", err
		}
		return dst[0].Caption, dst[0].Version, nil
	default:
		return "", "", errors.New("unable to determine operating system")
	}
}
