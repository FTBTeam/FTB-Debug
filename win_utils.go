//go:build windows

package main

import (
	"fmt"
	"github.com/pterm/pterm"
	"github.com/shirou/gopsutil/v3/process"
	"github.com/yusufpapurcu/wmi"
	"strings"
)

type (
	Win32_OperatingSystem struct {
		Caption string
		Version string
	}
)

func getSysInfo() (oSystem string, err error) {
	var dst []Win32_OperatingSystem

	q := wmi.CreateQuery(&dst, "")
	err = wmi.Query(q, &dst)
	if err != nil {
		return "", err
	}
	oSystem = fmt.Sprintf("%s (%s)", dst[0].Caption, dst[0].Version)
	return oSystem, nil
}

func getFTBProcess() {
	processes, err := process.Processes()
	if err != nil {
		pterm.Error.Println("Error getting processes\n", err)
		return
	}

	for _, p := range processes {
		n, err := p.Name()
		if err != nil {
			//pterm.Warning.Println("Error getting process name\n", err)
			continue
		}
		if n != "" && strings.ToLower(n) == "overwolf.exe" {
			p.Kill()
		}
		if n != "" && strings.ToLower(n) == "ftb electron app.exe" {
			p.Kill()
		}
	}
}
