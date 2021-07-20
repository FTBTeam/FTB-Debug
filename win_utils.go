// +build windows

package main

import (
	"fmt"
	"github.com/StackExchange/wmi"
	"github.com/pterm/pterm"
	"os"
	"path"
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

func locateApp() bool {
	if checkFilePathExistsSpinner("FTB App directory (AppData)", path.Join(os.Getenv("localappdata"), ".ftba")) {
		ftbApp.InstallLocation = path.Join(os.Getenv("localappdata"), ".ftba")
		return true
	} else if checkFilePathExistsSpinner("FTB App directory (home)", path.Join(ftbApp.User.HomeDir, ".ftba")) {
		ftbApp.InstallLocation = path.Join(ftbApp.User.HomeDir, ".ftba")
		return true
	} else {
		pterm.Error.Println("Unable to find app install")
		return false
	}
}
