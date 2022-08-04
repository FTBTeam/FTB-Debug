package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/Gaz492/haste"
	"github.com/eiannone/keyboard"
	"github.com/getsentry/sentry-go"
	"github.com/pterm/pterm"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path"
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
	betaApp       *bool
	GitCommit     string
	filesToUpload []FilesToUploadStruct
	appLocated    bool
)

func init() {
	var err error
	verboseLogging := flag.Bool("v", false, "Enable verbose logging")
	betaApp = flag.Bool("beta", false, "Use beta version of FTB")
	hasteClient = haste.NewHaste("https://pste.ch")
	flag.Parse()

	if *betaApp {
		owUID = "nelapelmednbnaigieobbdgbinpgcgkfmmdjembg"
	}

	if *verboseLogging {
		pterm.EnableDebugMessages()
	}
	logFile, err = ioutil.TempFile("", "ftb-debug-log")
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

	err := sentry.Init(sentry.ClientOptions{
		// Either set your DSN here or set the SENTRY_DSN environment variable.
		Dsn: "https://50d1f640c19c4a4d84297643f695d5a7@sentry.creeperhost.net/11",
		// Either set environment and release here or set the SENTRY_ENVIRONMENT
		// and SENTRY_RELEASE environment variables.
		Environment: "",
		Release:     fmt.Sprintf("ftb-debug-%s", GitCommit),
		// Enable printing of SDK debug messages.
		// Useful when getting started or trying to figure something out.
		Debug: false,
	})
	if err != nil {
		log.Fatalf("sentry.Init: %s", err)
	}

	defer sentry.Flush(2 * time.Second)
	defer cleanup(logFile)
	logMw = io.MultiWriter(os.Stdout, logFile)
	pterm.SetDefaultOutput(logMw)

	logo, _ := pterm.DefaultBigText.WithLetters(
		pterm.NewLettersFromStringWithStyle("F", pterm.NewStyle(pterm.FgCyan)),
		pterm.NewLettersFromStringWithStyle("T", pterm.NewStyle(pterm.FgGreen)),
		pterm.NewLettersFromStringWithStyle("B", pterm.NewStyle(pterm.FgRed))).Srender()
	pterm.DefaultCenter.Println(logo)
	pterm.DefaultCenter.WithCenterEachLineSeparately().Println(fmt.Sprintf("Version: %s-%s\n%s", "1.0.1", GitCommit, time.Now().UTC().Format(time.RFC1123)))
	pterm.Debug.Println("Verbose logging enabled")

	pterm.DefaultSection.Println("System Information")
	getOSInfo()

	pterm.DefaultSection.Println("FTB App Checks")
	usr, err := user.Current()
	if err != nil {
		sentry.CaptureException(err)
		pterm.Error.Println("Failed to get users home directory")
	}
	ftbApp.User = usr

	//App checks here
	appLocated = locateApp()
	if appLocated {
		pterm.DefaultSection.WithLevel(2).Println("Validating App structure")
		// Validate Minecraft bin folder exists
		doesBinExist()
		pterm.DefaultSection.WithLevel(2).Println("App info")
		pterm.Info.Println(fmt.Sprintf("Located app at %s", ftbApp.InstallLocation))
		getAppVersion()
		pterm.Info.Println("App version:", ftbApp.AppVersion)
		pterm.Info.Println("Backend version:", ftbApp.JarVersion)
		pterm.Info.Println("Web version:", ftbApp.WebVersion)
		pterm.Info.Println("Branch:", ftbApp.AppBranch)

		//TODO Add instance checking and settings file validation
		err := loadAppSettings()
		if err != nil {
			sentry.CaptureException(err)
			pterm.Error.Println("Failed to load app settings:\n", err)
		} else {
			pterm.Info.Println("Instance Location: ", ftbApp.Settings.InstanceLocation)
			if ftbApp.Settings.Jvmargs != "" {
				pterm.Info.Println("Custom Args: ", ftbApp.Settings.Jvmargs)
			}

		}

		pterm.DefaultSection.Println("Check for instances")
		listInstances()
	}

	// Upload info and logs
	pterm.DefaultSection.Println("Upload logs")
	uploadFiles()

	pterm.DefaultSection.Println("Debug Report Completed")

	tUpload, err := ioutil.ReadFile(logFile.Name())
	if err != nil {
		sentry.CaptureException(err)
		pterm.Error.Println("Failed to upload log file", logFile.Name())
		pterm.Error.Println(err)
	} else {
		resp, err := hasteClient.UploadBytes(tUpload)
		if err != nil {
			sentry.CaptureException(err)
			pterm.Error.Println("Failed to upload support file...")
			pterm.Error.Println(err)
		} else {
			pterm.DefaultBasicText.WithStyle(pterm.NewStyle(pterm.Bold)).Println(fmt.Sprintf("Please provide this code to support: FTB-DBG%s", strings.ToUpper(resp.Key)))
		}
	}

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

func uploadFiles() {
	appLocal, _ := os.UserCacheDir()
	hasteClient = haste.NewHaste("https://pste.ch")

	if appLocated {
		if ftbApp.Structure.Bin.Exists {
			uploadFile(ftbApp.InstallLocation, path.Join("bin", "settings.json"))
			uploadFile(ftbApp.InstallLocation, path.Join("bin", "launcher_log.txt"))
			uploadFile(ftbApp.InstallLocation, path.Join("bin", "launcher_cef_log.txt"))
			uploadFile(ftbApp.InstallLocation, path.Join("bin", "versions", "version_manifest.json"))
		}
		for _, file := range filesToUpload {
			pterm.Debug.Println("[fileToUpload] Uploading file:", file.File.Name())
			newUploadFile(file.Path, file.File.Name())
		}
		uploadFile(ftbApp.InstallLocation, path.Join("logs", "latest.log"))
		uploadFile(ftbApp.InstallLocation, path.Join("logs", "debug.log"))
	}

	if !*betaApp && runtime.GOOS == "windows" && checkFilePathExistsSpinner("Overwolf Logs", path.Join(appLocal, "Overwolf", "Log", "Apps", "FTB App")) {
		uploadFile(appLocal, path.Join("Overwolf", "Log", "Apps", "FTB App", "index.html.log"))
		uploadFile(appLocal, path.Join("Overwolf", "Log", "Apps", "FTB App", "background.html.log"))
		uploadFile(appLocal, path.Join("Overwolf", "Log", "Apps", "FTB App", "chat.html.log"))
	} else if *betaApp && runtime.GOOS == "windows" && checkFilePathExistsSpinner("Overwolf Logs", path.Join(appLocal, "Overwolf", "Log", "Apps", "FTB App Preview")) {
		uploadFile(appLocal, path.Join("Overwolf", "Log", "Apps", "FTB App Preview", "index.html.log"))
		uploadFile(appLocal, path.Join("Overwolf", "Log", "Apps", "FTB App Preview", "background.html.log"))
		uploadFile(appLocal, path.Join("Overwolf", "Log", "Apps", "FTB App Preview", "chat.html.log"))
	}
}

func doesBinExist() {
	binExists := checkFilePathExistsSpinner("Minecraft bin directory", path.Join(ftbApp.InstallLocation, "bin"))
	if binExists {
		ftbApp.Structure.Bin.Exists = true
	}
}

//func checkMinecraftBin() {
//	binExists := checkFilePathExistsSpinner("Minecraft bin directory", path.Join(ftbApp.InstallLocation, "bin"))
//	if binExists {
//		ftbApp.Structure.MCBin.Exists = true
//		checkFilePathExistsSpinner("Minecraft launcher", path.Join(ftbApp.InstallLocation, "bin", "launcher.exe"))
//		_, err := validateJson("Minecraft launcher profiles", path.Join(ftbApp.InstallLocation, "bin", "launcher_profiles.json"))
//		if err != nil {
//			return
//		}
//		ftbApp.Structure.MCBin.Profile = true
//	}
//}

func loadAppSettings() error {
	if ftbApp.Structure.Bin.Exists {
		var appSettings []byte
		var err error
		doesAppSettingsExist := checkFilePathExistsSpinner("Does app_settings.json exist?", path.Join(ftbApp.InstallLocation, "bin", "settings.json"))
		if doesAppSettingsExist {
			appSettings, err = ioutil.ReadFile(path.Join(ftbApp.InstallLocation, "bin", "settings.json"))
			if err != nil {
				sentry.CaptureException(err)
				pterm.Error.Println("Error reading settings.json:", err)
				return errors.New("error reading settings.json")
			}

		} else {
			appSettings, err = ioutil.ReadFile(path.Join(ftbApp.InstallLocation, "app_settings.json"))
			if err != nil {
				sentry.CaptureException(err)
				pterm.Error.Println("Error reading app_settings.json:", err)
				return errors.New("error reading app_settings.json")
			}
		}
		var i AppSettings
		if err := json.Unmarshal(appSettings, &i); err != nil {
			sentry.CaptureException(err)
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

func listInstances() {
	instancesExists := checkFilePathExistsSpinner("instances directory", ftbApp.Settings.InstanceLocation)
	if instancesExists {
		instances, _ := ioutil.ReadDir(path.Join(ftbApp.Settings.InstanceLocation))
		for _, instance := range instances {
			name := instance.Name()
			if instance.IsDir() {
				if name != ".localCache" {
					pterm.Info.Println("found instance: ", name)
					var i Instance
					data, err := ioutil.ReadFile(path.Join(ftbApp.Settings.InstanceLocation, name, "instance.json"))
					if err := json.Unmarshal(data, &i); err != nil {
						sentry.CaptureException(err)
						pterm.Error.Println("Error reading instance.json:", err)
						pterm.Debug.Println("JSON data:", string(data))
					}
					pterm.Info.Println("Name:", i.Name)
					pterm.Info.Println("Version:", i.Version)
					pterm.Info.Println("Version ID:", i.VersionID)
					pterm.Info.Println("Memory:", i.Memory)
					pterm.Info.Println(fmt.Sprintf("Min/Rec Memory: %d/%d", i.MinMemory, i.RecMemory))
					pterm.Info.Println("Custom Args:", i.JvmArgs)
					pterm.Info.Println("Embedded JRE:", i.EmbeddedJre)
					pterm.Info.Println("Is Modified:", i.IsModified)

					baseFiles, err := os.ReadDir(path.Join(ftbApp.Settings.InstanceLocation, name))
					for _, baseFile := range baseFiles {
						if strings.HasPrefix(baseFile.Name(), "hs_err_") {
							pterm.Debug.Println("Found java segfault log:", baseFile.Name())
							filesToUpload = append(filesToUpload, FilesToUploadStruct{File: baseFile, Path: path.Join(ftbApp.Settings.InstanceLocation, name, baseFile.Name())})
							//uploadFile(path.Join(ftbApp.Settings.InstanceLocation), path.Join(name, baseFile.Name()))
						}
					}

					logFolderExists := checkFilePathExistsSpinner(name+" logs folder", path.Join(ftbApp.Settings.InstanceLocation, name, "logs"))
					if logFolderExists {
						files, err := os.ReadDir(path.Join(ftbApp.Settings.InstanceLocation, name, "logs"))
						if err != nil {
							sentry.CaptureException(err)
							pterm.Error.Println("Error getting file list at:", path.Join(ftbApp.Settings.InstanceLocation, name, "logs"))
						} else {
							for _, file := range files {
								if filepath.Ext(file.Name()) == ".log" || filepath.Ext(file.Name()) == ".txt" {
									pterm.Debug.Println("Found log file:", file.Name())
									fInfo, err := os.Stat(path.Join(ftbApp.Settings.InstanceLocation, name, "logs", file.Name()))
									if err != nil {
										sentry.CaptureException(err)
										pterm.Error.Println("Error getting file info:", err)
									}
									//pterm.Info.Println(file.Name(), "last modified:", fInfo.ModTime().Format("2006-01-02 15:04:05"))
									pterm.Info.Println(file.Name(), "last modified:", fInfo.ModTime().Format("02/01/2006 15:04:05"))
									filesToUpload = append(filesToUpload, FilesToUploadStruct{File: file, Path: path.Join(ftbApp.Settings.InstanceLocation, name, file.Name())})
									//uploadFile(path.Join(ftbApp.Settings.InstanceLocation), path.Join(name, "logs", file.Name()))
								}
							}
						}
						//fInfo, err := os.Stat(path.Join(ftbApp.Settings.InstanceLocation, name, "logs", "latest.log"))
						//if err != nil {
						//	pterm.Error.Println("Error reading latest.log:", err)
						//}
						//pterm.Info.Println("Latest log timestamp:", fInfo.ModTime().Format("2006-01-02 15:04:05"))
						//uploadFile(path.Join(ftbApp.Settings.InstanceLocation), path.Join(name, "logs", "latest.log"))
					}
					validUuid := re.Find([]byte(name))
					if validUuid == nil {
						pterm.Error.Println(name, " instance name: invalid uuid")
					}
					_, err = validateJson(name+" instance.json", path.Join(ftbApp.Settings.InstanceLocation, name, "instance.json"))
					if err != nil {
						sentry.CaptureException(err)
						pterm.Error.Println("instance.json failed to validate")
					}
				}
			} else {
				pterm.Info.Println("found extra file in instances directory: ", name)
			}
		}
	}
}
