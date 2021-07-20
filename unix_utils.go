// +build !windows

package main

import (
	"errors"
	"fmt"
	"github.com/pterm/pterm"
	"os"
	"os/exec"
	"path"
	"regexp"
	"runtime"
)

func getSysInfo() (oSystem string, err error) {
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

func locateApp() bool {
	if runtime.GOOS == "darwin" {
		if checkFilePathExistsSpinner("FTB App directory (Application Support)", path.Join(os.Getenv("HOME"), "Library", "Application Support")) {
			ftbApp.InstallLocation = path.Join(os.Getenv("HOME"), "Library", "Application Support")
			return true
		} else {
			pterm.Error.Println("Unable to find app install")
			return false
		}
	} else if runtime.GOOS == "linux" {
		if checkFilePathExistsSpinner("FTB App directory (~/.ftba)", path.Join(ftbApp.User.HomeDir, ".ftba")) {
			ftbApp.InstallLocation = path.Join(ftbApp.User.HomeDir, ".ftba")
			return true
		} else {
			pterm.Error.Println("Unable to find app install")
			return false
		}
	} else {
		pterm.Error.Println("Could you let us know what operating system you are using so we can add our checks?")
		return false
	}
}

func getAppVersion(){
	ftbApp.AppVersion = "N/A"


}
