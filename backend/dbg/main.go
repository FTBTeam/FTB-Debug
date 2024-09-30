package dbg

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/pterm/pterm"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"time"
)

var (
	ftbApp               FTBApp
	logFile              *os.File
	logMw                io.Writer
	owUID                = "cmogmmciplgmocnhikmphehmeecmpaggknkjlbag"
	re                   = regexp.MustCompile(`(?m)[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}`)
	foundOverwolfVersion = false
	failedToLoadSettings = false
)

func RunDebug() (string, error) {
	var err error
	logFile, err = os.CreateTemp("", "ftb-dbg-log")
	if err != nil {
		pterm.Fatal.Println(err)
	}

	var manifest Manifest

	defer cleanup(logFile)
	logMw = io.MultiWriter(os.Stdout, NewCustomWriter(logFile))
	pterm.SetDefaultOutput(logMw)

	pterm.DefaultHeader.Println("System Info")
	getOSInfo()
	usr, err := user.Current()
	if err != nil {
		pterm.Error.Println("Failed to get users home directory")
	}
	ftbApp.User = usr

	pterm.DefaultHeader.Println("Running Network Checks")
	nc := runNetworkChecks()
	for _, n := range nc {
		if n.Error {
			pterm.Error.Println(n.Status)
		} else if !n.Success && !n.Error {
			pterm.Warning.Println(n.Status)
		} else {
			pterm.Success.Printfln("%s: %s", n.URL, n.Status)
		}
	}

	pterm.DefaultHeader.Println("Running App Checks")
	runAppChecks()
	profiles, err := getProfiles()
	hasActiveAccount := false
	if err != nil {
		pterm.Error.Println("Failed to get profiles:", err)
	} else {
		hasActiveAccount = isActiveProfileInProfiles(profiles)
	}

	pterm.DefaultSection.WithLevel(2).Println("App info")
	pterm.Info.Println(fmt.Sprintf("Located app at %s", ftbApp.InstallLocation))
	appVerData, err := getAppVersion()
	if err != nil {
		pterm.Error.Println("Error getting app version:", err)
	} else {
		pterm.Info.Println("App version:", appVerData.AppVersion)
		pterm.Info.Println("App release date:", time.Unix(int64(appVerData.Released), 0))
		pterm.Info.Println("Branch:", appVerData.Branch)
	}

	appLogs := make(map[string]string)
	instances := make(map[string]Instances)
	instanceLogs := make([]InstanceLogs, 0)

	appLogs, err = getAppLogs()
	if err != nil {
		pterm.Error.Println("Failed to get app logs:", err)
		return "", errors.New("failed to get app logs: " + err.Error())
	}
	if !failedToLoadSettings {
		pterm.DefaultSection.Println("Check for instances")
		instances, instanceLogs, err = getInstances()
		if err != nil {
			pterm.Error.Println("Failed to get instances:", err)
		}
	}

	// Additional files to upload
	miscFiles := []string{
		filepath.Join(ftbApp.InstallLocation, "storage", "settings.json"),
		filepath.Join(ftbApp.InstallLocation, "bin", "runtime", "installations.json"),
	}
	if foundOverwolfVersion {
		miscFiles = append(miscFiles, filepath.Join(overwolfAppLogs, "index.html.log"))
		miscFiles = append(miscFiles, filepath.Join(overwolfAppLogs, "background.html.log"))
		miscFiles = append(miscFiles, filepath.Join(overwolfAppLogs, "chat.html.log"))
	}

	for _, mf := range miscFiles {
		id, err := getMiscFile(mf)
		if err != nil {
			pterm.Error.Println("Error getting file:", err)
			continue
		}
		appLogs[filepath.Base(mf)] = id
	}

	tUpload, err := os.ReadFile(logFile.Name())
	if err != nil {
		pterm.Error.Println("Failed to read dbg output", logFile.Name())
		pterm.Error.Println(err)
	} else {
		if len(tUpload) > 0 {
			resp, err := uploadRequest(tUpload, "")
			if err != nil {
				pterm.Error.Println("Failed to upload support file...")
				pterm.Error.Println(err)
			} else {
				appLogs["dbg-tool-output"] = resp.Data.ID
			}
		}
	}

	// Compile manifest
	manifest.Version = "v2.0.6-go"
	manifest.MetaDetails = MetaDetails{
		InstanceCount:     len(instances),
		Today:             time.Now().UTC().Format(time.DateOnly),
		Time:              time.Now().Unix(),
		AddedAccounts:     len(profiles.Profiles),
		HasActiveAccounts: hasActiveAccount,
	}
	manifest.AppDetails = AppDetails{
		App:           appVerData.Commit,
		SharedVersion: appVerData.AppVersion,
		Meta:          appVerData,
	}
	manifest.AppLogs = appLogs
	manifest.ProviderInstanceMapping = instances
	manifest.InstanceLogs = instanceLogs
	manifest.NetworkChecks = nc

	pterm.DefaultHeader.Println("Manifest")
	jsonManifest, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		pterm.Error.Println("Error marshalling manifest:", err)
		return "", errors.New("error marshalling manifest: " + err.Error())
	}
	if len(jsonManifest) > 0 {
		request, err := uploadRequest(jsonManifest, "json")
		if err != nil {
			pterm.Error.Println("Failed to upload manifest:", err)
			return "", errors.New("failed to upload manifest: " + err.Error())
		}
		//codeStyle := pterm.NewStyle(pterm.FgLightMagenta, pterm.Bold)
		//pterm.DefaultBasicText.Printfln("Please provide this code to support: %s", codeStyle.Sprintf("dbg:%s", request.Data.ID))
		return request.Data.ID, nil
	}
	return "", errors.New("something went really wrong")
}
