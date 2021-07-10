package main

import (
	"flag"
	"fmt"
	"github.com/Gaz492/haste"
	"github.com/pterm/pterm"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path"
	"strings"
	"time"
)

var(
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
	log.Println(usr)

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

func checkMinecraftBin(filePath string){
	pterm.DefaultSection.WithLevel(2).Println("Validating App structure")
	binExists := checkFilePathExistsSpinner("bin directory", path.Join(filePath, "bin"))
	if binExists {
		checkFilePathExistsSpinner("minecraft launcher", path.Join(filePath, "bin", "launcher.exe"))
		validateJson("minecraft launcher profiles", path.Join(filePath, "bin", "launcher_profiles.json"))
	}
}

