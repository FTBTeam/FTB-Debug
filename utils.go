package main

import (
	"encoding/json"
	"fmt"
	"github.com/pterm/pterm"
	"io/ioutil"
	"log"
	"os"
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

func validateJson(message string, filePath string) {
	jsonF := checkFilePathSpinner(message, filePath)
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

func checkFilePathSpinner(dirMessage string, filePath string) bool {
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
