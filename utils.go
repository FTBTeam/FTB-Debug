package main

import (
	"fmt"
	"log"
	"os"
)

const (
	TB = 1099511600000
	GB = 1073741824
	MB = 1048576
	KB = 1024
)

func cleanup(logFile *os.File){
	if err := logFile.Close(); err != nil {
		log.Fatal("Unable to close temp log file: ", err)
	}
	if err := os.Remove(logFile.Name()); err != nil{
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

//func getOSInfo() (oSystem string, version string, err error){
//	switch runtime.GOOS {
//	case "windows":
//		oSystem, version, err = winOS()
//		return oSystem, version, err
//	case "darwin":
//		out, err := exec.Command("sw_vers").Output()
//		if err != nil {
//			return "", "", err
//		}
//		darwinRe := regexp.MustCompile(`ProductVersion:\W([0-9]*\.?[0-9?]*\.?[0-9?]*)`)
//		match := darwinRe.FindStringSubmatch(string(out))
//		return "", match[1], nil
//	default:
//		return "", "", errors.New("unable to determine operating system")
//	}
//}