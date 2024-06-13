package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/pterm/pterm"
	"github.com/pterm/pterm/putils"
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
	cli                  *bool
	GitCommit            string
	foundOverwolfVersion = false
)

func init() {
	var err error
	verboseLogging := flag.Bool("v", false, "Enable verbose logging")
	cli = flag.Bool("cli", false, "Only output the support code in console")
	noColours := flag.Bool("no-colours", false, "Disable colours in output")
	flag.Parse()

	if *verboseLogging {
		pterm.EnableDebugMessages()
	}
	if *noColours {
		pterm.DisableColor()
	}
	logFile, err = os.CreateTemp("", "ftb-debug-log")
	if err != nil {
		pterm.Fatal.Println(err)
	}
	pterm.Debug.Prefix = pterm.Prefix{
		Text:  "DEBUG",
		Style: pterm.NewStyle(pterm.BgLightMagenta, pterm.FgBlack),
	}
	pterm.Debug.MessageStyle = pterm.NewStyle(98)
}

func main() {
	if GitCommit == "" {
		GitCommit = "Dev"
	}

	var manifest Manifest

	defer cleanup(logFile)
	if *cli {
		logToConsole(false)
	} else {
		logToConsole(true)
	}

	logo, _ := pterm.DefaultBigText.WithLetters(
		putils.LettersFromStringWithStyle("F", pterm.NewStyle(pterm.FgCyan)),
		putils.LettersFromStringWithStyle("T", pterm.NewStyle(pterm.FgGreen)),
		putils.LettersFromStringWithStyle("B", pterm.NewStyle(pterm.FgRed))).Srender()
	pterm.DefaultCenter.Println(logo)
	pterm.DefaultCenter.WithCenterEachLineSeparately().Println(fmt.Sprintf("Version: %s-%s\n%s", "2.0.0", GitCommit, time.Now().UTC().Format(time.RFC1123)))
	pterm.Debug.Println("Verbose logging enabled")

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
			pterm.Success.Println(n.Status)
		}
	}

	pterm.DefaultHeader.Println("Running App Checks")
	runAppChecks()
	profiles, err := getProfiles()
	if err != nil {
		pterm.Error.Println("Failed to get profiles:", err)
		return
	}
	hasActiveAccount := isActiveProfileInProfiles(profiles)

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

	appLogs, err := getAppLogs()
	if err != nil {
		return
	}

	pterm.DefaultSection.Println("Check for instances")
	instances, instanceLogs, err := getInstances()
	if err != nil {
		pterm.Error.Println("Failed to get instances:", err)
		return
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
		pterm.Error.Println("Failed to read debug output", logFile.Name())
		pterm.Error.Println(err)
	} else {
		resp, err := uploadRequest(tUpload, "")
		if err != nil {
			pterm.Error.Println("Failed to upload support file...")
			pterm.Error.Println(err)
		} else {
			appLogs["debug-tool-output"] = resp.Data.ID
		}
	}

	// Compile manifest
	manifest.Version = "2.0.1-Go"
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
	}
	manifest.AppLogs = appLogs
	manifest.ProviderInstanceMapping = instances
	manifest.InstanceLogs = instanceLogs
	manifest.NetworkChecks = nc

	pterm.DefaultHeader.Println("Manifest")
	jsonManifest, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		pterm.Error.Println("Error marshalling manifest:", err)
		return
	}
	request, err := uploadRequest(jsonManifest, "json")
	if err != nil {
		return
	}
	pterm.Info.Println(request.Data.ID)

}
