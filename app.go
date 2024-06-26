package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hashicorp/go-version"
	"github.com/pterm/pterm"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
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
		failedToLoadSettings = true
		pterm.Error.Println("Failed to load app settings:\n", err)
	}
}

func loadAppSettings() error {
	if ftbApp.Structure.Bin.Exists {
		var appSettings []byte
		var err error
		appSettingsPath := filepath.Join(ftbApp.InstallLocation, "storage", "settings.json")
		doesAppSettingsExist := doesPathExist(appSettingsPath)
		if doesAppSettingsExist {
			appSettings, err = os.ReadFile(appSettingsPath)
			if err != nil {
				pterm.Error.Println("Error reading settings.json:", err)
			}
		}

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

func getInstances() (map[string]Instances, []InstanceLogs, error) {
	instancesExists := doesPathExist(ftbApp.Settings.InstanceLocation)
	if instancesExists {
		pterm.Info.Println("Instance Location: ", ftbApp.Settings.InstanceLocation)
		instances, _ := os.ReadDir(filepath.Join(ftbApp.Settings.InstanceLocation))
		pIM := make(map[string]Instances)
		var instanceLogs []InstanceLogs
		for _, instance := range instances {
			name := instance.Name()
			if instance.IsDir() {
				if name != ".localCache" {
					pterm.Info.Println("found instance: ", name)
					var i Instance
					data, err := os.ReadFile(filepath.Join(ftbApp.Settings.InstanceLocation, name, "instance.json"))
					if err := json.Unmarshal(data, &i); err != nil {
						pterm.Error.Printfln("error reading instance.json: %s", err.Error())
						continue
					} else {
						pIM[i.UUID] = Instances{
							Name:        i.Name,
							PackType:    i.PackType,
							PackId:      i.ID,
							PackVersion: i.VersionID,
						}

						// Check for logs
						logsPath := filepath.Join(ftbApp.Settings.InstanceLocation, name, "logs")
						logs := make(map[string]string)
						if doesPathExist(logsPath) {
							logs, err = getInstanceLogs(logsPath)
							if err != nil {
								pterm.Error.Printfln("Error getting instance logs: %s", err.Error())
							}
						}

						// Check for crash-reports
						crashLogsPath := filepath.Join(ftbApp.Settings.InstanceLocation, name, "crash-reports")
						crashLogs := make(map[string]string)
						if doesPathExist(crashLogsPath) {
							crashLogs, err = getInstanceLogs(crashLogsPath)
							if err != nil {
								pterm.Error.Printfln("Error getting instance crash logs: %s", err.Error())
							}
						}
						instanceLogs = append(instanceLogs, InstanceLogs{
							Created:   0,
							Name:      i.Name,
							UUID:      i.UUID,
							McVersion: i.McVersion,
							ModLoader: i.ModLoader,
							Logs:      logs,
							CrashLogs: crashLogs,
						})

					}

					_, err = validateJson(name+" instance.json", filepath.Join(ftbApp.Settings.InstanceLocation, name, "instance.json"))
					if err != nil {
						pterm.Error.Printfln("instance.json failed to validate: %s", err.Error())
						continue
					}
				}
			}
		}
		return pIM, instanceLogs, nil
	}
	return nil, nil, errors.New("instances directory not found")
}

// NEW STUFF HERE

func getAppVersion() (AppMeta, error) {
	var metaPath string
	if runtime.GOOS == "windows" {
		metaPath = filepath.Join(windowsAppPath, "resources", "meta.json")

		// checking overwolf
		var rawVersions []string
		files, err := os.ReadDir(overwolfAppPath)
		if err == nil {
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
			if len(versions) > 0 {
				foundOverwolfVersion = true
				sort.Slice(version.Collection(versions), func(i, j int) bool {
					return versions[i].GreaterThan(versions[j])
				})
				pterm.Debug.Println("Found versions:", versions)
				if !doesPathExist(metaPath) {
					metaPath = filepath.Join(overwolfAppPath, versions[0].String(), "meta.json")
				}
			}
		}
	} else if runtime.GOOS == "darwin" {
		metaPath = filepath.Join(macAppPath, "contents", "Resources", "meta.json")
	} else if runtime.GOOS == "linux" {
		return AppMeta{}, errors.New("linux not supported yet")
	} else {
		return AppMeta{}, errors.New("unknown OS, could you let us know what operating system you are using so we can add our checks")
	}

	installExists := doesPathExist(metaPath)
	if !installExists {
		return AppMeta{}, errors.New("app meta not found")
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
	profilesExists := doesPathExist(profilesPath)
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

		re, err := regexp.Compile(`^([0-9]{4}-[0-9]{2}-[0-9]{2}).*`)
		if err != nil {
			fmt.Println("Error compiling regex:", err)
			return nil, err
		}

		if re.MatchString(file.Name()) {
			continue
		}

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

			if len(data) == 0 {
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

func getInstanceLogs(path string) (map[string]string, error) {
	files, err := os.ReadDir(path)
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
			if len(data) == 0 {
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

func getMiscFile(path string) (string, error) {
	exists := doesPathExist(path)
	if !exists {
		return "", fmt.Errorf("file %s does not exist", path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	if filepath.Ext(path) == ".gz" {
		reader, err := gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			return "", err
		}
		data, err = io.ReadAll(reader)
	}

	lang := ""
	if filepath.Ext(path) == ".json" {
		lang = "json"
	}
	if len(data) == 0 {
		return "", errors.New("file is empty")
	}
	request, err := uploadRequest(data, lang)
	if err != nil {
		return "", err
	}
	return request.Data.ID, nil
}
