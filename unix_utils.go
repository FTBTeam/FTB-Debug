//go:build !windows

package main

import (
	"errors"
	"fmt"
	"github.com/pterm/pterm"
	"github.com/shirou/gopsutil/v3/process"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
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
		if checkFilePathExistsSpinner("FTB App directory (Application Support)", filepath.Join(os.Getenv("HOME"), "Library", "Application Support", ".ftba")) {
			ftbApp.InstallLocation = filepath.Join(os.Getenv("HOME"), "Library", "Application Support", ".ftba")
			return true
		} else {
			pterm.Error.Println("Unable to find app install")
			return false
		}
	} else if runtime.GOOS == "linux" {
		if checkFilePathExistsSpinner("FTB App directory (~/.ftba)", filepath.Join(ftbApp.User.HomeDir, ".ftba")) {
			ftbApp.InstallLocation = filepath.Join(ftbApp.User.HomeDir, ".ftba")
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

//func oldgetAppVersion() {
//	ftbApp.AppVersion = "Electron"
//	var appPath string
//	if runtime.GOOS == "darwin" {
//		appPath = filepath.Join(ftbApp.User.HomeDir, "Applications", "FTBApp.app", "contents", "Resources", "app", "bin", "ftbapp.app", "Contents", "Resources", "app.asar")
//		installExists := checkFilePathExistsSpinner("App install (User home)", appPath)
//		if !installExists {
//			appPath = filepath.Join("/Applications", "FTBApp.app", "contents", "Resources", "app", "bin", "ftbapp.app", "Contents", "Resources", "app.asar")
//			installExists = checkFilePathExistsSpinner("App install (User home)", appPath)
//			if !installExists {
//				ftbApp.JarVersion = "N/A"
//				ftbApp.WebVersion = "N/A"
//				ftbApp.AppBranch = "N/A"
//				return
//			}
//		}
//	} else if runtime.GOOS == "linux" {
//		appPath = filepath.Join(ftbApp.User.HomeDir, "FTBA", "bin", "resources", "app.asar")
//	} else {
//		pterm.Error.Println("Could you let us know what operating system you are using so we can add our checks?")
//		ftbApp.JarVersion = "N/A"
//		ftbApp.WebVersion = "N/A"
//		ftbApp.AppBranch = "N/A"
//		return
//	}
//	f, err := os.Open(appPath)
//	if err != nil {
//		pterm.Error.Println(err)
//		return
//	}
//	defer f.Close()
//
//	archive, err := asar.Decode(f)
//	if err != nil {
//		pterm.Error.Println(err)
//		return
//	}
//
//	versionRaw := archive.Find("version.json")
//	if versionRaw == nil {
//		pterm.Error.Println("file not found")
//		return
//	}
//	var versionJson VersionJson
//	err = json.Unmarshal(versionRaw.Bytes(), &versionJson)
//	if err != nil {
//		pterm.Error.Println("JSON unmarshal error")
//		return
//	}
//	ftbApp.JarVersion = versionJson.JarVersion
//	ftbApp.WebVersion = versionJson.WebVersion
//	ftbApp.AppBranch = versionJson.Branch
//}

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
		if n != "" && strings.ToLower(n) == "ftb-app" {
			p.Kill()
		}
		if n != "" && strings.ToLower(n) == "ftb electron app" {
			p.Kill()
		}
	}
}
