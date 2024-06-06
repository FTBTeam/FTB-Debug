package main

import (
	"flag"
	"fmt"
	"github.com/Gaz492/haste"
	"github.com/eiannone/keyboard"
	"github.com/pterm/pterm"
	"github.com/pterm/pterm/putils"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
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
	appLocated    bool
)

func init() {
	var err error
	verboseLogging := flag.Bool("v", false, "Enable verbose logging")
	cli = flag.Bool("cli", false, "Only output the support code in console")
	hasteClient = haste.NewHaste("https://pste.ch")
	flag.Parse()

	if *verboseLogging {
		pterm.EnableDebugMessages()
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
	pterm.DefaultCenter.WithCenterEachLineSeparately().Println(fmt.Sprintf("Version: %s-%s\n%s", "1.1.0", GitCommit, time.Now().UTC().Format(time.RFC1123)))
	pterm.Debug.Println("Verbose logging enabled")

	pterm.DefaultSection.Println("System Information")
	getOSInfo()

	pterm.Info.Println("Killing FTB App")
	getFTBProcess()

	pterm.DefaultSection.Println("FTB App Checks")
	usr, err := user.Current()
	if err != nil {
		pterm.Error.Println("Failed to get users home directory")
	}
	ftbApp.User = usr

	pterm.DefaultSection.Println("Network requests checks")
	runNetworkChecks()

	//App checks here
	runAppChecks()

	// Upload info and logs
	pterm.DefaultSection.Println("Upload logs")
	uploadFiles()

	pterm.DefaultSection.Println("Debug Report Completed")
	if *cli {
		logToConsole(true)
	}

	tUpload, err := os.ReadFile(logFile.Name())
	if err != nil {
		pterm.Error.Println("Failed to upload log file", logFile.Name())
		pterm.Error.Println(err)
	} else {
		resp, err := hasteClient.UploadBytes(tUpload)
		if err != nil {
			pterm.Error.Println("Failed to upload support file...")
			pterm.Error.Println(err)
		} else {
			pterm.DefaultBasicText.WithStyle(pterm.NewStyle(pterm.Bold)).Println(fmt.Sprintf("Please provide this code to support: FTB-DBG%s", strings.ToUpper(resp.Key)))
		}
	}

	if !*cli {
		pterm.Info.Println("Press ESC to exit...")

		if err := keyboard.Open(); err != nil {
			panic(err)
		}
		defer func() {
			_ = keyboard.Close()
		}()
		for {
			_, key, err := keyboard.GetKey()
			if err != nil {
				panic(err)
			}
			if key == keyboard.KeyEsc {
				break
			}
		}
	}
}

func uploadFiles() {
	appLocal, _ := os.UserCacheDir()
	hasteClient = haste.NewHaste("https://pste.ch")

	for _, file := range filesToUpload {
		pterm.Debug.Println("[fileToUpload] Uploading file:", file.File.Name())
		newUploadFile(file.Path, file.File.Name())
	}

	if appLocated {
		newUploadFile(filepath.Join(ftbApp.InstallLocation, "logs", "latest.log"), "latest.log")
		newUploadFile(filepath.Join(ftbApp.InstallLocation, "logs", "debug.log"), "debug.log")
	}

	if runtime.GOOS == "windows" && checkFilePathExistsSpinner("Overwolf Logs", filepath.Join(appLocal, "Overwolf", "Log", "Apps", "FTB App")) {
		newUploadFile(filepath.Join(appLocal, "Overwolf", "Log", "Apps", "FTB App", "index.html.log"), "index.html.log")
		newUploadFile(filepath.Join(appLocal, "Overwolf", "Log", "Apps", "FTB App", "background.html.log"), "background.html.log")
		newUploadFile(filepath.Join(appLocal, "Overwolf", "Log", "Apps", "FTB App", "chat.html.log"), "chat.html.log")
	}
}
