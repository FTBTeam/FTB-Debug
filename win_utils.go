//go:build windows

package main

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-version"
	"github.com/pterm/pterm"
	"github.com/shirou/gopsutil/v3/process"
	"github.com/yusufpapurcu/wmi"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type (
	Win32_OperatingSystem struct {
		Caption string
		Version string
	}
)

// TODO implement getting app version from overwolf
func getAppVersion() {
	var rawVersions []string
	appLocal, _ := os.UserCacheDir()
	overwolfDIR := filepath.Join(appLocal, "Overwolf", "Extensions", owUID)
	files, err := os.ReadDir(overwolfDIR)
	if err != nil {
		pterm.Error.Println("Error while reading Overwolf versions")
		return
	}
	for _, file := range files {
		if file.IsDir() {
			rawVersions = append(rawVersions, file.Name())
		}
	}
	versions := make([]*version.Version, len(rawVersions))
	for i, raw := range rawVersions {
		v, _ := version.NewVersion(raw)
		versions[i] = v
	}
	sort.Slice(version.Collection(versions), func(i, j int) bool {
		return versions[i].GreaterThan(versions[j])
	})
	pterm.Debug.Println("Found versions:", versions)
	ftbApp.AppVersion = versions[0].String()

	jsonFile, err := os.Open(filepath.Join(overwolfDIR, ftbApp.AppVersion, "version.json"))
	// if we os.Open returns an error then handle it
	if err != nil {
		pterm.Error.Println("Error opening version.json:", err)
	}
	defer jsonFile.Close()
	byteValue, _ := io.ReadAll(jsonFile)
	var versionJson VersionJson
	json.Unmarshal(byteValue, &versionJson)
	ftbApp.JarVersion = versionJson.JarVersion
	ftbApp.WebVersion = versionJson.WebVersion
	ftbApp.AppBranch = versionJson.Branch
}

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
	if checkFilePathExistsSpinner("FTB App directory (AppData)", filepath.Join(os.Getenv("localappdata"), ".ftba")) {
		ftbApp.InstallLocation = filepath.Join(os.Getenv("localappdata"), ".ftba")
		return true
	} else if checkFilePathExistsSpinner("FTB App directory (home)", filepath.Join(ftbApp.User.HomeDir, ".ftba")) {
		ftbApp.InstallLocation = filepath.Join(ftbApp.User.HomeDir, ".ftba")
		return true
	} else {
		pterm.Error.Println("Unable to find app install")
		return false
	}
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
			pterm.Warning.Println("Error getting process name\n", err)
		}
		if n != "" && strings.ToLower(n) == "overwolf.exe" {
			p.Kill()
		}
	}
}
