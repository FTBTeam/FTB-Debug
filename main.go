package main

import (
	"flag"
	"fmt"
	"github.com/Gaz492/haste"
	"github.com/pterm/pterm"
	"io"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"runtime"
	"strings"
	"time"
)

var(
	ftbApp FTBApp
	logFile *os.File
	logMw io.Writer
)

func init(){
	var err error
	verboseLogging := flag.Bool("v", false, "Enable verbose logging")
	hasteClient = haste.NewHaste("https://pste.ch")
	flag.Parse()
	if *verboseLogging {
		pterm.EnableDebugMessages()
	}
	logFile, err = ioutil.TempFile("", "ftb-debug-log")
	if err != nil {
		pterm.Fatal.Println(err)
	}
}

func main() {
	defer cleanup(logFile)
	logMw = io.MultiWriter(os.Stdout, logFile)
	pterm.SetDefaultOutput(logMw)

	logo, _ := pterm.DefaultBigText.WithLetters(
		pterm.NewLettersFromStringWithStyle("F", pterm.NewStyle(pterm.FgCyan)),
		pterm.NewLettersFromStringWithStyle("T", pterm.NewStyle(pterm.FgGreen)),
		pterm.NewLettersFromStringWithStyle("B", pterm.NewStyle(pterm.FgRed))).Srender()
	pterm.DefaultCenter.Println(logo)
	pterm.DefaultCenter.WithCenterEachLineSeparately().Println(fmt.Sprintf("Version: %s\n%s", "1.0.1", time.Now().UTC().Format(time.RFC1123)))
	pterm.Debug.Println("Verbose logging enabled")

	pterm.DefaultSection.Println("System Information")
	getOSInfo()

	pterm.DefaultSection.Println("FTB App Checks")
	usr, err := user.Current()
	if err != nil {
		pterm.Error.Println("Failed to get users home directory")
	}
	ftbApp.User = usr

	//App checks here
	located := locateApp()
	if !located {
		os.Exit(1)
	}
	pterm.Info.Println(fmt.Sprintf("Located app at %s", ftbApp.InstallLocation))
	getAppVersion()
	//TODO Add instance checking and settings file validation


	pterm.DefaultSection.WithLevel(2).Println("Validating App structure")
	checkMinecraftBin()

	//TODO do uploading of log files

	pterm.DefaultSection.Println("Debug Report Completed")

	tUpload, err := ioutil.ReadFile(logFile.Name())
	if err != nil {
		pterm.Error.Println("Failed to upload support file...")
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

	pterm.Info.Println("Press ESC to exit...")

	//if err := keyboard.Open(); err != nil {
	//	panic(err)
	//}
	//defer func() {
	//	_ = keyboard.Close()
	//}()
	//for {
	//	_, key, err := keyboard.GetKey()
	//	if err != nil {
	//		panic(err)
	//	}
	//	if key == keyboard.KeyEsc {
	//		break
	//	}
	//}
}

func uploadFiles() {
	appLocal, _ := os.UserCacheDir()
	hasteClient = haste.NewHaste("https://pste.ch")
	uploadFile(ftbApp.InstallLocation, path.Join("bin", "launcher_profiles.json"))
	uploadFile(ftbApp.InstallLocation, path.Join("logs", "latest.log"))
	uploadFile(ftbApp.InstallLocation, path.Join("logs", "debug.log"))
	if runtime.GOOS == "windows" && checkFilePathExistsSpinner("Overwolf Logs", path.Join(appLocal, "Overwolf", "Log", "Apps", "FTB App")) {
		uploadFile(appLocal, path.Join("Overwolf", "Log", "Apps", "FTB App", "index.html.log"))
		uploadFile(appLocal, path.Join("Overwolf", "Log", "Apps", "FTB App", "background.html.log"))
		uploadFile(appLocal, path.Join("Overwolf", "Log", "Apps", "FTB App", "chat.html.log"))
	}
}

func checkMinecraftBin(){
	binExists := checkFilePathExistsSpinner("Minecraft bin directory", path.Join(ftbApp.InstallLocation, "bin"))
	if binExists {
		checkFilePathExistsSpinner("Minecraft launcher", path.Join(ftbApp.InstallLocation, "bin", "launcher.exe"))
		validateJson("Minecraft launcher profiles", path.Join(ftbApp.InstallLocation, "bin", "launcher_profiles.json"))
	}
}

