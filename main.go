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
		resp, err := uploadRequest(tUpload)
		if err != nil {
			pterm.Error.Println("Failed to upload support file...")
			pterm.Error.Println(err)
		} else {
			codeStyle := pterm.NewStyle(pterm.FgLightMagenta, pterm.Bold)
			pterm.DefaultBasicText.Printfln("Please provide this code to support: %s", codeStyle.Sprintf("dbg:%s", resp.Data.ID))
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
		uploadFile(file.Path, file.File.Name())
	}

	if appLocated {
		uploadFile(filepath.Join(ftbApp.InstallLocation, "logs", "latest.log"), "App latest.log")
		uploadFile(filepath.Join(ftbApp.InstallLocation, "logs", "debug.log"), "App debug.log")

		electronLog := filepath.Join(ftbApp.InstallLocation, "logs", "ftb-app-electron.log")
		_, exists := checkFilePath(electronLog)
		if exists {
			uploadFile(electronLog, "ftb-app-electron.log")
		}

		frontendLog := filepath.Join(ftbApp.InstallLocation, "logs", "ftb-app-frontend.log")
		_, exists = checkFilePath(frontendLog)
		if exists {
			uploadFile(frontendLog, "ftb-app-frontend.log")
		}

		installerLog := filepath.Join(ftbApp.InstallLocation, "logs", "ftb-app-installer.log")
		_, exists = checkFilePath(installerLog)
		if exists {
			uploadFile(installerLog, "ftb-app-installer.log")
		}

		runtimeInstallations := filepath.Join(ftbApp.InstallLocation, "bin", "runtime", "installations.json")
		_, exists = checkFilePath(runtimeInstallations)
		if exists {
			uploadFile(runtimeInstallations, "runtime installations.json")
		}
	}

	if runtime.GOOS == "windows" && checkFilePathExistsSpinner("Overwolf Logs", filepath.Join(appLocal, "Overwolf", "Log", "Apps", "FTB App")) {
		uploadFile(filepath.Join(appLocal, "Overwolf", "Log", "Apps", "FTB App", "index.html.log"), "Overwolf index.html.log")
		uploadFile(filepath.Join(appLocal, "Overwolf", "Log", "Apps", "FTB App", "background.html.log"), "Overwolf background.html.log")
		uploadFile(filepath.Join(appLocal, "Overwolf", "Log", "Apps", "FTB App", "chat.html.log"), "Overwolf chat.html.log")
	}
}
