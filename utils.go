package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/pterm/pterm"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"time"
)

func cleanup(logFile *os.File) {
	if err := logFile.Close(); err != nil {
		log.Fatal("Unable to close temp log file: ", err)
	}
	if err := os.Remove(logFile.Name()); err != nil {
		log.Fatal(err)
	}
}

func ByteCountIEC(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %ciB",
		float64(b)/float64(div), "KMGTPE"[exp])
}

func validateJson(message string, filePath string) (bool, error) {
	jsonF := doesPathExist(filePath)
	if jsonF {
		jsonFile, err := os.Open(filePath)
		if err != nil {
			pterm.Error.Println(message, ": failed to load file\n", err)
			return false, err
		}

		defer jsonFile.Close()
		byteValue, _ := io.ReadAll(jsonFile)
		valid := json.Valid(byteValue)
		if !valid {
			pterm.Error.Println(fmt.Sprintf("%s: is invalid", message))
			return false, err
		}
		pterm.Success.Println(fmt.Sprintf("%s: json is valid", message))
		return true, nil
	}
	return false, errors.New(fmt.Sprintf("Unable to validate %s\nUnable to find file %s", message, filePath))
}

func getOSInfo() {
	cpuInfo, _ := cpu.Info()
	memInfo, _ := mem.VirtualMemory()
	oSystem, err := getSysInfo()
	if err == nil {
		if oSystem != "" {
			pterm.Info.Println(fmt.Sprintf("OS: %s", oSystem))
		} else {
			pterm.Info.Println(fmt.Sprintf("OS: %s", runtime.GOOS))
		}
	} else {
		pterm.Info.Println(fmt.Sprintf("OS: %s", runtime.GOOS))
	}

	if len(cpuInfo) > 0 {
		pterm.Info.Println(fmt.Sprintf("CPU: %s (%s)", cpuInfo[0].ModelName, cpuInfo[0].VendorID))
	} else {
		pterm.Info.Println("CPU: Unable to calculate")
	}

	pterm.Info.Println(fmt.Sprintf("Memory: %s / %s (%.2f%% used)", ByteCountIEC(int64(memInfo.Used)), ByteCountIEC(int64(memInfo.Total)), memInfo.UsedPercent))

	javaHome := os.Getenv("JAVA_HOME")
	if javaHome != "" {
		pterm.Info.Println("Java Home:", javaHome)
	}
}

//func checkFilePathExistsSpinner(dirMessage string, filePath string) bool {
//	dirStatus, _ := pterm.DefaultSpinner.Start("Checking ", "for ", dirMessage)
//	message, success := checkFilePath(filePath)
//	if !success {
//		dirStatus.Warning(fmt.Sprintf("%s: %s", dirMessage, message))
//		return false
//	}
//
//	dirStatus.Success(fmt.Sprintf("%s: %s", dirMessage, message))
//	return true
//}

func doesPathExist(filePath string) bool {
	if _, err := os.Stat(filePath); err == nil {
		return true
	}
	return false
}

//func checkFilePath(filePath string) (string, bool) {
//	if _, err := os.Stat(filePath); err == nil {
//		return "file/directory exists", true
//
//	} else if os.IsNotExist(err) {
//		return "file/directory does not exist", false
//
//	} else {
//		return "possible permission error, could not determine if file/directory explicitly exists or not", false
//	}
//}

func uploadFile(filePath string, comment string) {
	pterm.Debug.Println(filePath)
	fileName := filepath.Base(filePath)
	data, err := os.ReadFile(filePath)
	if err != nil {
		pterm.Error.Printfln("Uploading %s: failed to open file\n%s", comment, err)
		return
	}
	if fileName == "launcher_profiles.json" {
		data, err = sanitizeProfile(data)
		if err != nil {
			pterm.Warning.Println("Error sanitizing launcher_profiles.json")
			return
		}
	} else if fileName == "settings.json" {
		data, err = sanitizeSettings(data)
		if err != nil {
			pterm.Warning.Println("Error sanitizing settings.json")
			return
		}
	} else {
		data = sanitizeLogs(data)
	}
	r, err := uploadRequest(data, "")
	if err != nil {
		pterm.Error.Printfln("Uploading %s: failed to upload\n%s", comment, err)
		return
	}
	pterm.Info.Printfln("Uploaded [%s#%s]", r.Data.ID, comment)
}

func uploadRequest(data []byte, lang string) (PsteMeResp, error) {
	// http put request to https://pste.me/v1/paste
	client := &http.Client{}
	req, err := http.NewRequest("PUT", "https://pste.me/v1/paste", bytes.NewBuffer(data))
	if err != nil {
		return PsteMeResp{}, err
	}
	query := req.URL.Query()
	query.Add("expires_at", strconv.FormatInt(time.Now().Add(time.Hour*1440).Unix(), 10))
	if lang != "" {
		query.Add("language", lang)
	}
	req.URL.RawQuery = query.Encode()
	resp, err := client.Do(req)
	if err != nil {
		return PsteMeResp{}, err
	}
	defer resp.Body.Close()
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return PsteMeResp{}, err
	}
	if resp.StatusCode != 200 {
		return PsteMeResp{}, fmt.Errorf("invalid status code: %d\n%s", resp.StatusCode, string(content))
	}
	var r PsteMeResp
	if err := json.Unmarshal(content, &r); err != nil {
		return PsteMeResp{}, err
	}
	return r, nil
}

func sanitizeProfile(data []byte) ([]byte, error) {
	var i interface{}
	if data != nil {
		if err := json.Unmarshal(data, &i); err != nil {
			pterm.Error.Println("Error reading launcher profile:", err)
			pterm.Debug.Println("JSON data:", string(data))
			return nil, err
		}
		if m, ok := i.(map[string]interface{}); ok {
			delete(m, "authenticationDatabase")
			delete(m, "clientToken")
		}
		output, err := json.MarshalIndent(i, "", "  ")
		if err != nil {
			pterm.Error.Println("Error marshaling json:", err)
			return nil, err
		}
		return output, nil
	}
	return nil, fmt.Errorf("unable to sanitize profile, parameter data is nil")
}

func sanitizeSettings(data []byte) ([]byte, error) {
	if data != nil {
		var i AppSettings
		if err := json.Unmarshal(data, &i); err != nil {
			pterm.Error.Println("Error reading app settings:", err)
			pterm.Debug.Println("JSON data:", string(data))
			return nil, err
		}
		i.SessionString = "************************"
		output, err := json.MarshalIndent(i, "", "  ")
		if err != nil {
			pterm.Error.Println("Error marshaling json:", err)
			return nil, err
		}
		return output, nil
	}
	return nil, fmt.Errorf("unable to sanitize settings, parameter data is nil")
}

func sanitizeLogs(data []byte) []byte {
	reToken := regexp.MustCompile(`(^|")(ey[a-zA-Z0-9._-]+|Ew[a-zA-Z0-9._+/-]+=|M\.R3[a-zA-Z0-9._+!\*\$/-]+)`)

	clean := reToken.ReplaceAll(data, []byte("${1}******AUTHTOKEN******$3"))
	return clean
}

func logToConsole(b bool) {
	if b {
		logMw = io.MultiWriter(os.Stdout, logFile)
		pterm.SetDefaultOutput(logMw)
	} else {
		pterm.SetDefaultOutput(logFile)
	}
}

func doesBinExist() {
	binExists := doesPathExist(filepath.Join(ftbApp.InstallLocation, "bin"))
	if binExists {
		ftbApp.Structure.Bin.Exists = true
	}
}

//func locateFTBApp() (string, error) {
//	if runtime.GOOS == "windows" {
//		if doesPathExist(windowsAppPath) {
//			return windowsAppPath, nil
//		} else {
//			return "", errors.New("unable to find .ftba directory")
//		}
//	} else if runtime.GOOS == "darwin" {
//		if doesPathExist(macAppPath) {
//			return macAppPath, nil
//		} else {
//			return "", errors.New("unable to find .ftba directory")
//		}
//	} else if runtime.GOOS == "linux" {
//		if doesPathExist(linuxAppPath) {
//			return linuxAppPath, nil
//		} else {
//			return "", errors.New("unable to find .ftba directory")
//		}
//	} else {
//		return "", errors.New("unknown OS, could you let us know what operating system you are using so we can add our checks")
//	}
//}

func locateFTBAFolder() (string, error) {
	if runtime.GOOS == "windows" {
		if doesPathExist(filepath.Join(os.Getenv("localappdata"), ".ftba")) {
			return filepath.Join(os.Getenv("localappdata"), ".ftba"), nil
		} else if doesPathExist(filepath.Join(ftbApp.User.HomeDir, ".ftba")) {
			return filepath.Join(ftbApp.User.HomeDir, ".ftba"), nil
		} else {
			return "", errors.New("unable to find .ftba directory")
		}
	} else if runtime.GOOS == "darwin" {
		if doesPathExist(filepath.Join(os.Getenv("HOME"), "Library", "Application Support", ".ftba")) {
			return filepath.Join(os.Getenv("HOME"), "Library", "Application Support", ".ftba"), nil
		} else {
			return "", errors.New("unable to find .ftba directory")
		}
	} else if runtime.GOOS == "linux" {
		if doesPathExist(filepath.Join(ftbApp.User.HomeDir, ".ftba")) {
			return filepath.Join(ftbApp.User.HomeDir, ".ftba"), nil
		} else {
			return "", errors.New("unable to find .ftba directory")
		}
	} else {
		return "", errors.New("unknown OS, could you let us know what operating system you are using so we can add our checks")
	}
}

func runNetworkChecks() []NetworkCheck {
	var nc []NetworkCheck
	for url, checks := range checkRequestsURLs {
		client := &http.Client{}
		req, err := http.NewRequest(checks.Method, url, nil)
		if err != nil {
			nc = append(nc, NetworkCheck{URL: url, Success: false, Error: true, Status: fmt.Sprintf("Error creating request to %s\n%s", url, err.Error())})
			continue
		}
		resp, err := client.Do(req)
		if err != nil {
			nc = append(nc, NetworkCheck{URL: url, Success: false, Error: true, Status: fmt.Sprintf("Error making request to %s\n%s", url, err.Error())})
			continue
		}

		// DO checks
		if resp.StatusCode != checks.ExpectedStatusCode {
			nc = append(nc, NetworkCheck{URL: url, Success: false, Error: false, Status: fmt.Sprintf("%s: Expected %d got %d (%s)", url, checks.ExpectedStatusCode, resp.StatusCode, resp.Status)})
			continue
		}
		if checks.ValidateResponse {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				nc = append(nc, NetworkCheck{URL: url, Success: false, Error: true, Status: fmt.Sprintf("Error reading response body\n%s", err.Error())})
				continue
			}
			if string(body) != checks.ExpectedReponse {
				nc = append(nc, NetworkCheck{URL: url, Success: false, Error: false, Status: fmt.Sprintf("%s: Expected %s got %s", url, checks.ExpectedReponse, string(body))})
				continue
			}
		}
		nc = append(nc, NetworkCheck{URL: url, Success: true, Error: false, Status: fmt.Sprintf("%s returned expected results", url)})
	}
	return nc
}

func isActiveProfileInProfiles(profiles Profiles) bool {
	for _, profile := range profiles.Profiles {
		if profile.UUID == profiles.ActiveProfile {
			return true
		}
	}
	return false
}
