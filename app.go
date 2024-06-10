package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/pterm/pterm"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"time"
)

func runAppChecks() {
	ftbaPath, err := locateFTBAFolder()
	if err != nil {
		pterm.Error.Println("Error locating app:", err)
		return
	}
	ftbApp.InstallLocation = ftbaPath
	// Validate bin folder exists
	doesBinExist()

	pterm.DefaultSection.WithLevel(2).Println("App info")
	pterm.Info.Println(fmt.Sprintf("Located app at %s", ftbApp.InstallLocation))
	appVerData, err := getAppVersion()
	if err != nil {
		pterm.Error.Println("Error getting app version:", err)
		return
	}
	pterm.Info.Println("App version:", appVerData.AppVersion)
	pterm.Info.Println("App release date:", time.Unix(int64(appVerData.Released), 0))
	pterm.Info.Println("Branch:", appVerData.Branch)

	//TODO Add instance checking and settings file validation
	err = loadAppSettings()
	if err != nil {
		pterm.Error.Println("Failed to load app settings:\n", err)
	} else {
		pterm.Info.Println("Instance Location: ", ftbApp.Settings.InstanceLocation)
		if ftbApp.Settings.Jvmargs != "" {
			pterm.Info.Println("Custom Args: ", ftbApp.Settings.Jvmargs)
		}

	}

	pterm.DefaultSection.Println("Check for instances")
	listInstances()

}

func loadAppSettings() error {
	if ftbApp.Structure.Bin.Exists {
		var appSettings []byte
		var err error
		doesAppSettingsExist := checkFilePathExistsSpinner("Does settings.json exist?", filepath.Join(ftbApp.InstallLocation, "bin", "settings.json"))
		if doesAppSettingsExist {
			appSettings, err = os.ReadFile(filepath.Join(ftbApp.InstallLocation, "bin", "settings.json"))
			if err != nil {
				pterm.Error.Println("Error reading settings.json:", err)
			} else {
				settingsFile, err := os.Stat(filepath.Join(ftbApp.InstallLocation, "bin", "settings.json"))
				if err != nil {
					pterm.Error.Println("Error getting settings.json stat:", err)
				} else {
					filesToUpload = append(filesToUpload, FilesToUploadStruct{
						File: settingsFile,
						Path: filepath.Join(ftbApp.InstallLocation, "bin", "settings.json"),
					})
				}
			}
		} else {
			appSettings, err = os.ReadFile(filepath.Join(ftbApp.InstallLocation, "app_settings.json"))
			if err != nil {
				pterm.Error.Println("Error reading app_settings.json:", err)
				return errors.New("error reading app_settings.json")
			}
		}
		doesVersionsManifestExist := checkFilePathExistsSpinner("Does version_manifest.json exist?", filepath.Join(ftbApp.InstallLocation, "bin", "versions", "version_manifest.json"))
		if doesVersionsManifestExist {
			vManifest, err := os.Stat(filepath.Join(ftbApp.InstallLocation, "bin", "versions", "version_manifest.json"))
			if err != nil {
				pterm.Error.Println("Error getting file stat for version_manifest.json:", err)
			}
			filesToUpload = append(filesToUpload, FilesToUploadStruct{
				File: vManifest,
				Path: filepath.Join(ftbApp.InstallLocation, "bin", "versions", "version_manifest.json"),
			})
		}
		uploadFile(filepath.Join(ftbApp.InstallLocation, "bin", "versions", "version_manifest.json"), "version_manifest.json")

		var i AppSettings
		if err := json.Unmarshal(appSettings, &i); err != nil {
			pterm.Error.Println("Error reading app settings:", err)
			pterm.Debug.Println("JSON data:", string(appSettings))
			return err
		}
		ftbApp.Settings = i
		return nil
	} else {
		return errors.New("MC bin folder missing")
	}
}

func listInstances() {
	instancesExists := checkFilePathExistsSpinner("instances directory", ftbApp.Settings.InstanceLocation)
	if instancesExists {
		instances, _ := os.ReadDir(filepath.Join(ftbApp.Settings.InstanceLocation))
		for _, instance := range instances {
			name := instance.Name()
			if instance.IsDir() {
				if name != ".localCache" {
					pterm.Info.Println("found instance: ", name)
					var i Instance
					iJsonStat, err := os.Stat(filepath.Join(ftbApp.Settings.InstanceLocation, name, "instance.json"))
					if err != nil {
						pterm.Error.Println("Error getting instance.json file stat:", err)
					}
					data, err := os.ReadFile(filepath.Join(ftbApp.Settings.InstanceLocation, name, "instance.json"))
					if err := json.Unmarshal(data, &i); err != nil {
						pterm.Error.Println("Error reading instance.json:", err)
						pterm.Debug.Println("JSON data:", string(data))
						filesToUpload = append(filesToUpload, FilesToUploadStruct{File: iJsonStat, Path: path.Join(ftbApp.Settings.InstanceLocation, name, iJsonStat.Name())})
					} else {
						pterm.Info.Println("Name:", i.Name)
						pterm.Info.Println("Version:", i.Version)
						pterm.Info.Println("Version ID:", i.VersionID)
						pterm.Info.Println("Memory:", i.Memory)
						pterm.Info.Println(fmt.Sprintf("Min/Rec Memory: %d/%d", i.MinMemory, i.RecMemory))
						pterm.Info.Println("Custom Args:", i.JvmArgs)
						pterm.Info.Println("Embedded JRE:", i.EmbeddedJre)
						pterm.Info.Println("Is Modified:", i.IsModified)
					}
					pterm.Info.Println("Name:", i.Name)
					pterm.Info.Println("Version:", i.Version)
					pterm.Info.Println("Version ID:", i.VersionID)
					pterm.Info.Println("Memory:", i.Memory)
					pterm.Info.Println(fmt.Sprintf("Min/Rec Memory: %d/%d", i.MinMemory, i.RecMemory))
					pterm.Info.Println("Custom Args:", i.JvmArgs)
					pterm.Info.Println("Embedded JRE:", i.EmbeddedJre)
					pterm.Info.Println("Is Modified:", i.IsModified)

					_, err = validateJson(name+" instance.json", filepath.Join(ftbApp.Settings.InstanceLocation, name, "instance.json"))
					if err != nil {
						pterm.Error.Println("instance.json failed to validate")
					}
				}
			} else {
				pterm.Info.Println("found extra file in instances directory: ", name)
			}
		}
	}
}

// NEW STUFF HERE

func getAppVersion() (AppMeta, error) {
	var metaPath string
	if runtime.GOOS == "windows" {
		// TODO: Implement windows version
	} else if runtime.GOOS == "darwin" {
		metaPath = filepath.Join(macAppPath, "contents", "Resources", "meta.json")
		installExists := checkFilePathExistsSpinner("App metadata", metaPath)
		if !installExists {
			return AppMeta{}, errors.New("app meta not found")
		}
	} else if runtime.GOOS == "linux" {

	} else {
		return AppMeta{}, errors.New("unknown OS, could you let us know what operating system you are using so we can add our checks")
	}

	// Read json file
	metaRaw, err := os.ReadFile(metaPath)
	if err != nil {
		return AppMeta{}, err
	}
	var metaJson AppMeta
	if err := json.Unmarshal(metaRaw, &metaJson); err != nil {
		return AppMeta{}, err
	}
	return metaJson, nil
}
