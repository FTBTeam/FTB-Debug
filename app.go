package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/pterm/pterm"
	"io"
	"os"
	"path"
	"path/filepath"
	"runtime"
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

	//TODO Add instance checking and settings file validation
	err = loadAppSettings()
	if err != nil {
		pterm.Error.Println("Failed to load app settings:\n", err)
	}
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

func getInstances() (map[string]Instances, error) {
	instancesExists := checkFilePathExistsSpinner("instances directory", ftbApp.Settings.InstanceLocation)
	if instancesExists {
		pterm.Info.Println("Instance Location: ", ftbApp.Settings.InstanceLocation)
		instances, _ := os.ReadDir(filepath.Join(ftbApp.Settings.InstanceLocation))
		pIM := make(map[string]Instances)
		for _, instance := range instances {
			name := instance.Name()
			if instance.IsDir() {
				if name != ".localCache" {
					pterm.Info.Println("found instance: ", name)
					var i Instance
					instancePaths = append(instancePaths, filepath.Join(ftbApp.Settings.InstanceLocation, name))
					iJsonStat, err := os.Stat(filepath.Join(ftbApp.Settings.InstanceLocation, name, "instance.json"))
					if err != nil {
						return nil, fmt.Errorf("error getting instance.json file stat: %s", err.Error())
					}
					data, err := os.ReadFile(filepath.Join(ftbApp.Settings.InstanceLocation, name, "instance.json"))
					if err := json.Unmarshal(data, &i); err != nil {
						filesToUpload = append(filesToUpload, FilesToUploadStruct{File: iJsonStat, Path: path.Join(ftbApp.Settings.InstanceLocation, name, iJsonStat.Name())})
						return nil, fmt.Errorf("error reading instance.json: %s", err.Error())
					} else {
						pIM[i.UUID] = Instances{
							Name:        i.Name,
							PackType:    i.PackType,
							PackId:      i.ID,
							PackVersion: i.VersionID,
						}
					}

					_, err = validateJson(name+" instance.json", filepath.Join(ftbApp.Settings.InstanceLocation, name, "instance.json"))
					if err != nil {
						return nil, fmt.Errorf("instance.json failed to validate: %s", err.Error())
					}
				}
			}
		}
		return pIM, nil
	}
	return nil, errors.New("instances directory not found")
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

func getProfiles() (Profiles, error) {
	profilesPath := filepath.Join(ftbApp.InstallLocation, "profiles.json")
	profilesExists := checkFilePathExistsSpinner("profiles.json", profilesPath)
	if profilesExists {
		profilesRaw, err := os.ReadFile(profilesPath)
		if err != nil {
			return Profiles{}, err
		}
		var profiles Profiles
		if err := json.Unmarshal(profilesRaw, &profiles); err != nil {
			return Profiles{}, err
		}
		return profiles, nil
	}
	return Profiles{}, errors.New("profiles.json not found")
}

func getAppLogs() (map[string]string, error) {
	lPath := filepath.Join(ftbApp.InstallLocation, "logs")
	files, err := os.ReadDir(lPath)
	if err != nil {
		return nil, err
	}
	logFile := make(map[string]string)
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".log" || filepath.Ext(file.Name()) == ".txt" || filepath.Ext(file.Name()) == ".gz" {
			data, err := os.ReadFile(filepath.Join(lPath, file.Name()))
			if err != nil {
				pterm.Error.Println("Error reading log file:", err)
				continue
			}

			if filepath.Ext(file.Name()) == ".gz" {
				reader, err := gzip.NewReader(bytes.NewReader(data))
				if err != nil {
					return nil, err
				}
				data, err = io.ReadAll(reader)
			}

			request, err := uploadRequest(data, "log")
			if err != nil {
				pterm.Error.Println("Error creating upload request:", err)
				continue
			}
			logFile[file.Name()] = request.Data.ID
		}
	}
	return logFile, nil
}

func getInstanceLogs(path string) (map[string]string, error) {
	if !checkFilePathExistsSpinner("Instance logs folder", path) {
		return nil, errors.New("instance logs folder not found")
	}

	files, err := os.ReadDir(filepath.Join(path))
	if err != nil {
		return nil, err
	}
	logFile := make(map[string]string)
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".log" || filepath.Ext(file.Name()) == ".txt" {
			data, err := os.ReadFile(filepath.Join(path, file.Name()))
			if err != nil {
				pterm.Error.Println("Error reading log file:", err)
				continue
			}
			request, err := uploadRequest(data, "log")
			if err != nil {
				pterm.Error.Println("Error creating upload request:", err)
				continue
			}
			logFile[file.Name()] = request.Data.ID
		}
	}
	return logFile, nil
}
