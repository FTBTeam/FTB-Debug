// +build windows

package main

import (
	"encoding/json"
	"fmt"
	"github.com/StackExchange/wmi"
	"github.com/hashicorp/go-version"
	"github.com/pterm/pterm"
	"io/ioutil"
	"os"
	"path"
	"sort"
)

type (
	Win32_OperatingSystem struct {
		Caption string
		Version string
	}
)

//TODO implement getting app version from overwolf
func getAppVersion(){
	var rawVersions []string
	appLocal, _ := os.UserCacheDir()
	overwolfDIR := path.Join(appLocal, "Overwolf", "Extensions", owUID)
	files, err := ioutil.ReadDir(overwolfDIR)
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
	sort.Sort(version.Collection(versions))
	pterm.Debug.Println("Found versions:", versions)
	ftbApp.AppVersion = versions[0].String()

	jsonFile, err := os.Open(path.Join(overwolfDIR, ftbApp.AppVersion, "version.json"))
	// if we os.Open returns an error then handle it
	if err != nil {
		pterm.Error.Println("Error opening version.json:", err)
	}
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
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
