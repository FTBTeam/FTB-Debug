package main

import (
	"flag"
	"fmt"
	"github.com/eiannone/keyboard"
	"github.com/pterm/pterm"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path"
	"runtime"
	"time"
)

var(
	tmpLog string
	logMw io.Writer
)

func init(){
	verboseLogging := flag.Bool("v", false, "Enable verbose logging")
	flag.Parse()
	if *verboseLogging {
		pterm.EnableDebugMessages()
	}
	logFile, err := ioutil.TempFile("", "ftb-debug-log")
	if err != nil {
		pterm.Fatal.Println(err)
	}
	tmpLog = logFile.Name()
	defer cleanup(logFile)
	logMw = io.MultiWriter(os.Stdout, logFile)
	pterm.SetDefaultOutput(logMw)
}

func main() {
	logo, _ := pterm.DefaultBigText.WithLetters(
		pterm.NewLettersFromStringWithStyle("F", pterm.NewStyle(pterm.FgCyan)),
		pterm.NewLettersFromStringWithStyle("T", pterm.NewStyle(pterm.FgGreen)),
		pterm.NewLettersFromStringWithStyle("B", pterm.NewStyle(pterm.FgRed))).Srender()
	pterm.DefaultCenter.Println(logo)
	pterm.DefaultCenter.WithCenterEachLineSeparately().Println(fmt.Sprintf("Version: %s\n%s", "1.0.1", time.Now().UTC().Format(time.RFC1123)))
	pterm.Debug.Println("Verbose logging enabled")

	pterm.DefaultSection.Println("System Information")
	cpuInfo, _ := cpu.Info()
	memInfo, _ := mem.VirtualMemory()
	oSystem, err := getOSInfo()
	if err == nil {
		if oSystem != "" {
			pterm.Info.Println(fmt.Sprintf("OS: %s", oSystem))
		} else {
			pterm.Info.Println(fmt.Sprintf("OS: %s", runtime.GOOS))
		}
	} else {
		pterm.Info.Println(fmt.Sprintf("OS: %s", runtime.GOOS))
	}
	pterm.Info.Println(fmt.Sprintf("CPU: %s (%s)", cpuInfo[0].ModelName, cpuInfo[0].VendorID))
	pterm.Info.Println(fmt.Sprintf("Memory: %s / %s (%.2f%% used)", ByteCountIEC(int64(memInfo.Used)), ByteCountIEC(int64(memInfo.Total)), memInfo.UsedPercent))

	javaHome := os.Getenv("JAVA_HOME")
	if javaHome != "" {
		pterm.Info.Println("Java Home:", javaHome)
	}

	pterm.DefaultSection.Println("FTB App Checks")
	usr, err := user.Current()
	if err != nil {
		pterm.Error.Println("Failed to get users home directory")
	}
	log.Println(usr)

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

func checkMinecraftBin(filePath string){
	pterm.DefaultSection.WithLevel(2).Println("Validating App structure")
	binExists := checkFilePathSpinner("bin directory", path.Join(filePath, "bin"))
	if binExists {
		checkFilePathSpinner("minecraft launcher", path.Join(filePath, "bin", "launcher.exe"))
		validateJson("minecraft launcher profiles", path.Join(filePath, "bin", "launcher_profiles.json"))
	}
}

