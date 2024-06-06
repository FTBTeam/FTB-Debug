package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Gaz492/haste"
	"github.com/pterm/pterm"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

var (
	hasteClient       *haste.Haste
	checkRequestsURLs = map[string]CheckURLStruct{
		"https://api.modpacks.ch/public/api/ping": {
			method:             "GET",
			validateResponse:   true,
			expectedStatusCode: http.StatusOK,
			expectedReponse:    "{\"status\":\"success\",\"reply\":\"pong\"}",
		},
		"https://api.creeper.host/api/health": {
			method:             "HEAD",
			validateResponse:   false,
			expectedStatusCode: http.StatusOK,
			expectedReponse:    "",
		},
		"https://maven.creeperhost.net": {
			method:             "HEAD",
			validateResponse:   false,
			expectedStatusCode: http.StatusOK,
			expectedReponse:    "",
		},
		"https://maven.fabricmc.net": {
			method:             "HEAD",
			validateResponse:   false,
			expectedStatusCode: http.StatusOK,
			expectedReponse:    "",
		},
		"https://maven.minecraftforge.net/net/minecraftforge/forge/maven-metadata.xml": {
			method:             "HEAD",
			validateResponse:   false,
			expectedStatusCode: http.StatusOK,
			expectedReponse:    "",
		},
		"https://api.feed-the-beast.com/": {
			method:             "HEAD",
			validateResponse:   false,
			expectedStatusCode: http.StatusNotFound,
			expectedReponse:    "",
		},
		"https://meta.feed-the-beast.com/v1/health": {
			method:             "HEAD",
			validateResponse:   false,
			expectedStatusCode: http.StatusOK,
			expectedReponse:    "",
		},
	}
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
	jsonF := checkFilePathExistsSpinner(message, filePath)
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

func checkFilePathExistsSpinner(dirMessage string, filePath string) bool {
	dirStatus, _ := pterm.DefaultSpinner.Start("Checking ", "for ", dirMessage)
	message, success := checkFilePath(filePath)
	if !success {
		dirStatus.Warning(fmt.Sprintf("%s: %s", dirMessage, message))
		return false
	}

	dirStatus.Success(fmt.Sprintf("%s: %s", dirMessage, message))
	return true
}

func checkFilePath(filePath string) (string, bool) {
	if _, err := os.Stat(filePath); err == nil {
		return "file/directory exists", true

	} else if os.IsNotExist(err) {
		return "file/directory does not exist", false

	} else {
		return "possible permission error, could not determine if file/directory explicitly exists or not", false
	}
}

func newUploadFile(filePath string, fileName string) {
	pterm.Debug.Println(filePath)
	data, err := os.ReadFile(filePath)
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
	if err != nil {
		pterm.Warning.Println(fmt.Sprintf("Uploading %s: failed to open file\n%v", fileName, err))
	} else {
		resp, err := hasteClient.UploadBytes(data)
		if err != nil {
			pterm.Warning.Println(fmt.Sprintf("Uploading %s: failed to upload - %s", fileName, err.Error()))
			if err.Error() == "file too large" {
				pterm.Info.Println("Trying again with transfer.sh")
				uploadBigFile(filePath, fileName)
			}
		} else {
			pterm.Success.Println(fmt.Sprintf("Uploaded %s: %s", fileName, resp.GetLink(hasteClient)))
		}
	}
}

func uploadBigFile(filePath string, name string) {
	req, err := uploadFileRequest(filepath.Join(filePath))
	if err != nil {
		pterm.Error.Println(fmt.Sprintf("Uploading %s: failed to upload", name))
		pterm.Error.Println(err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		pterm.Error.Println(fmt.Sprintf("Uploading %s: failed to upload", name))
		pterm.Error.Println(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		pterm.Error.Println(fmt.Sprintf("Uploading %s: failed to upload", name))
		pterm.Error.Println(err)
	} else {
		pterm.Success.Println(fmt.Sprintf("Uploaded %s: %s", name, strings.TrimSuffix(string(body), "\n")))
	}

}

func uploadFileRequest(path string) (*http.Request, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("upload", filepath.Base(path))
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, file)

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", "https://transfer.sh", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, err
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
	binExists := checkFilePathExistsSpinner("Minecraft bin directory", filepath.Join(ftbApp.InstallLocation, "bin"))
	if binExists {
		ftbApp.Structure.Bin.Exists = true
	}
}

func runNetworkChecks() {
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
}
