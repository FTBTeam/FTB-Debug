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
	"regexp"
	"time"
)

var (
	ftbApp        FTBApp
	logFile       *os.File
	logMw         io.Writer
	owUID         = "cmogmmciplgmocnhikmphehmeecmpaggknkjlbag"
	re            = regexp.MustCompile(`(?m)[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}`)
	cli           *bool
	GitCommit     string
	filesToUpload []FilesToUploadStruct
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

	// Compile manifest
	manifest.NetworkChecks = nc

	pterm.DefaultHeader.Println("Manifest")
	jsonManifest, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		pterm.Error.Println("Error marshalling manifest:", err)
		return
	}
	pterm.Info.Println(string(jsonManifest))

}
