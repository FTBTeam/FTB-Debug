package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Gaz492/haste"
	"github.com/pterm/pterm"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

var hasteClient *haste.Haste

func cleanup(logFile *os.File) {
	if err := logFile.Close(); err != nil {
		log.Fatal("Unable to close temp log file: ", err)
	}
	if err := os.Remove(logFile.Name()); err != nil {
		log.Fatal(err)
	}
}

//TODO implement getting app version from overwolf
func getAppVersion(){
	appLocal, _ := os.UserCacheDir()
	files, err := ioutil.ReadDir(path.Join(appLocal, "Overwolf", "Extensions", "cmogmmciplgmocnhikmphehmeecmpaggknkjlbag"))
	if err != nil {
		pterm.Error.Println("Error while reading Overwolf versions")
		return
	}
	for _, file := range files {
		if file.IsDir() {
			pterm.Info.Println(file.Name())
		}
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

func validateJson(message string, filePath string) {
	jsonF := checkFilePathExistsSpinner(message, filePath)
	if jsonF {
		jsonFile, err := os.Open(filePath)
		if err != nil {
			pterm.Error.Println(message, ": failed to load file\n", err)
			return
		}

		defer jsonFile.Close()
		byteValue, _ := ioutil.ReadAll(jsonFile)
		valid := json.Valid(byteValue)
		if !valid {
			pterm.Error.Println(message, ": is invalid")
			return
		}
		pterm.Success.Println(message, ": json is valid")
	}
}

func getOSInfo(){
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
	pterm.Info.Println(fmt.Sprintf("CPU: %s (%s)", cpuInfo[0].ModelName, cpuInfo[0].VendorID))
	pterm.Info.Println(fmt.Sprintf("Memory: %s / %s (%.2f%% used)", ByteCountIEC(int64(memInfo.Used)), ByteCountIEC(int64(memInfo.Total)), memInfo.UsedPercent))

	javaHome := os.Getenv("JAVA_HOME")
	if javaHome != "" {
		pterm.Info.Println("Java Home:", javaHome)
	}
}

func checkFilePathExistsSpinner(dirMessage string, filePath string) bool {
	dirStatus, _ := pterm.DefaultSpinner.Start("Checking for ", dirMessage)
	message, success := checkFilePath(filePath)
	if !success {
		dirStatus.Warning(dirMessage, ": ", message)
		return false
	}

	dirStatus.Success(dirMessage, ": ", message)
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

func uploadFile(filePath string, name string) {
	data, err := ioutil.ReadFile(path.Join(filePath, name))
	if err != nil {
		pterm.Warning.Println("Uploading ", name, ": failed to open file")
	} else {
		resp, err := hasteClient.UploadBytes(data)
		if err != nil {
			pterm.Warning.Println("Uploading ", name, ": failed to upload - ", err.Error())
			if err.Error() == "file too large" {
				pterm.Info.Println("Trying again with transfer.sh")
				uploadBigFile(filePath, name)
			}
		} else {
			pterm.Success.Println("Uploaded ", name, ": ", resp.GetLink(hasteClient))
		}
	}
}

func uploadBigFile(filePath string, name string) {
	req, err := newfileUploadRequest(path.Join(filePath, name))
	if err != nil {
		pterm.Error.Println("Uploading " + name + ": failed to upload")
		pterm.Error.Println(err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		pterm.Error.Println("Uploading " + name + ": failed to upload")
		pterm.Error.Println(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		pterm.Error.Println("Uploading " + name + ": failed to upload")
		pterm.Error.Println(err)
	} else {
		pterm.Success.Println("Uploaded " + name + ": " + strings.TrimSuffix(string(body), "\n"))
	}

}

func newfileUploadRequest(path string) (*http.Request, error) {
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