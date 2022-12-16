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
	"github.com/pterm/pterm/putils"
	"io"
	"log"
	"net/http"
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
	silent        *bool
	GitCommit     string
	filesToUpload []FilesToUploadStruct
	appLocated    bool
)

func init() {
	defer sentry.Recover()
	var err error
	verboseLogging := flag.Bool("v", false, "Enable verbose logging")
	betaApp = flag.Bool("beta", false, "Use beta version of FTB")
	silent = flag.Bool("silent", false, "Only output the support code in console")
	hasteClient = haste.NewHaste("https://pste.ch")
	flag.Parse()

	if *betaApp {
		owUID = "nelapelmednbnaigieobbdgbinpgcgkfmmdjembg"
	}

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
	defer sentry.Recover()
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
	if *silent {
		logToConsole(false)
	} else {
		logToConsole(true)
	}

	logo, _ := pterm.DefaultBigText.WithLetters(
		putils.LettersFromStringWithStyle("F", pterm.NewStyle(pterm.FgCyan)),
		putils.LettersFromStringWithStyle("T", pterm.NewStyle(pterm.FgGreen)),
		putils.LettersFromStringWithStyle("B", pterm.NewStyle(pterm.FgRed))).Srender()
	pterm.DefaultCenter.Println(logo)
	pterm.DefaultCenter.WithCenterEachLineSeparately().Println(fmt.Sprintf("Version: %s-%s\n%s", "1.0.1", GitCommit, time.Now().UTC().Format(time.RFC1123)))
	pterm.Debug.Println("Verbose logging enabled")

	pterm.DefaultSection.Println("System Information")
	getOSInfo()

	pterm.Info.Println("Killing FTB App")
	getFTBProcess()

	pterm.DefaultSection.Println("FTB App Checks")
	usr, err := user.Current()
	if err != nil {
		sentry.CaptureException(err)
		pterm.Error.Println("Failed to get users home directory")
	}
	ftbApp.User = usr

	pterm.DefaultSection.Println("Network requests checks")
	for url, checks := range checkRequestsURLs {
		client := &http.Client{}
		req, err := http.NewRequest(checks.method, url, nil)
		if err != nil {
			pterm.Error.Println("Error creating request to %s\n%s", url, err.Error())
			continue
		}
		resp, err := client.Do(req)
		if err != nil {
			pterm.Error.Println("Error making request to %s\n%s", url, err.Error())
			continue
		}

		// DO checks
		if resp.StatusCode != checks.expectedStatusCode {
			pterm.Warning.Printfln("%s returned unexpected status code, expected %d got %d (%s)", url, checks.expectedStatusCode, resp.StatusCode, resp.Status)
			continue
		}
		if checks.validateResponse {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				pterm.Error.Println("Error reading response body\n", err)
				continue
			}
			if string(body) != checks.expectedReponse {
				pterm.Warning.Printfln("%s did not match expected response\n%s", url, string(body))
				continue
			}
		}
		pterm.Success.Printfln("%s returned expected results", url)
	}

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
	if *silent {
		logToConsole(true)
	}

	tUpload, err := os.ReadFile(logFile.Name())
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
	defer sentry.Recover()
	appLocal, _ := os.UserCacheDir()
	hasteClient = haste.NewHaste("https://pste.ch")

	if appLocated {
		if ftbApp.Structure.Bin.Exists {
			newUploadFile(filepath.Join(ftbApp.InstallLocation, "bin", "settings.json"), "settings.json")
			newUploadFile(filepath.Join(ftbApp.InstallLocation, "bin", "versions", "version_manifest.json"), "version_manifest.json")
		}
		for _, file := range filesToUpload {
			pterm.Debug.Println("[fileToUpload] Uploading file:", file.File.Name())
			newUploadFile(file.Path, file.File.Name())
		}
		newUploadFile(filepath.Join(ftbApp.InstallLocation, "logs", "latest.log"), "latest.log")
		newUploadFile(filepath.Join(ftbApp.InstallLocation, "logs", "debug.log"), "debug.log")
	}

	if !*betaApp && runtime.GOOS == "windows" && checkFilePathExistsSpinner("Overwolf Logs", path.Join(appLocal, "Overwolf", "Log", "Apps", "FTB App")) {
		newUploadFile(filepath.Join(appLocal, "Overwolf", "Log", "Apps", "FTB App", "index.html.log"), "index.html.log")
		newUploadFile(filepath.Join(appLocal, "Overwolf", "Log", "Apps", "FTB App", "background.html.log"), "background.html.log")
		newUploadFile(filepath.Join(appLocal, "Overwolf", "Log", "Apps", "FTB App", "chat.html.log"), "chat.html.log")
	} else if *betaApp && runtime.GOOS == "windows" && checkFilePathExistsSpinner("Overwolf Logs", path.Join(appLocal, "Overwolf", "Log", "Apps", "FTB App Preview")) {
		newUploadFile(filepath.Join(appLocal, "Overwolf", "Log", "Apps", "FTB App Preview", "index.html.log"), "index.html.log")
		newUploadFile(filepath.Join(appLocal, "Overwolf", "Log", "Apps", "FTB App Preview", "background.html.log"), "background.html.log")
		newUploadFile(filepath.Join(appLocal, "Overwolf", "Log", "Apps", "FTB App Preview", "chat.html.log"), "chat.html.log")
	}
}

func doesBinExist() {
	defer sentry.Recover()
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
	defer sentry.Recover()
	if ftbApp.Structure.Bin.Exists {
		var appSettings []byte
		var err error
		doesAppSettingsExist := checkFilePathExistsSpinner("Does app_settings.json exist?", filepath.Join(ftbApp.InstallLocation, "bin", "settings.json"))
		if doesAppSettingsExist {
			appSettings, err = os.ReadFile(filepath.Join(ftbApp.InstallLocation, "bin", "settings.json"))
			if err != nil {
				sentry.CaptureException(err)
				pterm.Error.Println("Error reading settings.json:", err)
				return errors.New("error reading settings.json")
			}

		} else {
			appSettings, err = os.ReadFile(filepath.Join(ftbApp.InstallLocation, "app_settings.json"))
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
	defer sentry.Recover()
	instancesExists := checkFilePathExistsSpinner("instances directory", ftbApp.Settings.InstanceLocation)
	if instancesExists {
		instances, _ := os.ReadDir(path.Join(ftbApp.Settings.InstanceLocation))
		for _, instance := range instances {
			name := instance.Name()
			if instance.IsDir() {
				if name != ".localCache" {
					pterm.Info.Println("found instance: ", name)
					var i Instance
					data, err := os.ReadFile(path.Join(ftbApp.Settings.InstanceLocation, name, "instance.json"))
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
						}
					}

					logFolderExists := checkFilePathExistsSpinner(name+" logs folder", path.Join(ftbApp.Settings.InstanceLocation, name, "logs"))
					if logFolderExists {
						files, err := os.ReadDir(filepath.Join(ftbApp.Settings.InstanceLocation, name, "logs"))
						if err != nil {
							sentry.CaptureException(err)
							pterm.Error.Println("Error getting file list at:", path.Join(ftbApp.Settings.InstanceLocation, name, "logs"))
						} else {
							for _, file := range files {
								if filepath.Ext(file.Name()) == ".log" || filepath.Ext(file.Name()) == ".txt" {
									pterm.Debug.Println("Found log file:", file.Name())
									fInfo, err := os.Stat(filepath.Join(ftbApp.Settings.InstanceLocation, name, "logs", file.Name()))
									if err != nil {
										sentry.CaptureException(err)
										pterm.Error.Println("Error getting file info:", err)
									}
									pterm.Info.Println(file.Name(), "last modified:", fInfo.ModTime().Format("02/01/2006 15:04:05"))
									filesToUpload = append(filesToUpload, FilesToUploadStruct{File: file, Path: filepath.Join(ftbApp.Settings.InstanceLocation, name, "logs", file.Name())})
								}
							}
						}
					}
					validUuid := re.Find([]byte(name))
					if validUuid == nil {
						pterm.Error.Println(name, " instance name: invalid uuid")
					}
					_, err = validateJson(name+" instance.json", filepath.Join(ftbApp.Settings.InstanceLocation, name, "instance.json"))
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
