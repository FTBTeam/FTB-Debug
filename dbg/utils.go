package dbg

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"ftb-debug/v2/shared"
	"github.com/google/uuid"
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
	"strings"
	"time"
	"unicode"
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
	jsonF := shared.DoesPathExist(filePath)
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

func uploadRequest(data []byte, lang string) (PsteMeResp, error) {
	// http put request to https://pste.me/v1/paste

	sanitizedData := sanitize(data)
	client := &http.Client{}
	req, err := http.NewRequest("PUT", "https://pste.me/v1/paste", bytes.NewBuffer(sanitizedData))
	if err != nil {
		return PsteMeResp{}, err
	}
	query := req.URL.Query()
	query.Add("expires_at", strconv.FormatInt(time.Now().Add(time.Hour*2190).Unix()-time.Now().Unix(), 10))
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

func sanitize(data []byte) []byte {
	reToken := regexp.MustCompile(`(^|")(ey[a-zA-Z0-9._-]+|Ew[a-zA-Z0-9._+/-]+=|M\.R3[a-zA-Z0-9._+!\*\$/-]+)`)
	reWindowsPath := regexp.MustCompile(`((?:[A-Za-z]:)\\Users\\)([^/\\\r\n\t\v]+)(\\.+)?`)
	reMacPath := regexp.MustCompile(`(/Users/)([^/\\\r\n\t\v]+)(/.+)?`)
	reLinuxPath := regexp.MustCompile(`(/home/)([^/\\\r\n\t\v]+)(/.+)?`)

	clean := reToken.ReplaceAll(data, []byte("$1******AUTHTOKEN******$3"))
	clean = reWindowsPath.ReplaceAll(clean, []byte("$1***$3"))
	clean = reMacPath.ReplaceAll(clean, []byte("$1***$3"))
	clean = reLinuxPath.ReplaceAll(clean, []byte("$1***$3"))
	return clean
}

func doesBinExist() {
	binExists := shared.DoesPathExist(filepath.Join(ftbApp.InstallLocation, "bin"))
	if binExists {
		ftbApp.Structure.Bin.Exists = true
	}
}

func locateFTBAFolder() (string, error) {
	if runtime.GOOS == "windows" {
		if shared.DoesPathExist(filepath.Join(os.Getenv("localappdata"), ".ftba")) {
			return filepath.Join(os.Getenv("localappdata"), ".ftba"), nil
		} else if shared.DoesPathExist(filepath.Join(ftbApp.User.HomeDir, ".ftba")) {
			return filepath.Join(ftbApp.User.HomeDir, ".ftba"), nil
		} else {
			return "", errors.New("unable to find .ftba directory")
		}
	} else if runtime.GOOS == "darwin" {
		if shared.DoesPathExist(filepath.Join(os.Getenv("HOME"), "Library", "Application Support", ".ftba")) {
			return filepath.Join(os.Getenv("HOME"), "Library", "Application Support", ".ftba"), nil
		} else {
			return "", errors.New("unable to find .ftba directory")
		}
	} else if runtime.GOOS == "linux" {
		if shared.DoesPathExist(filepath.Join(ftbApp.User.HomeDir, ".ftba")) {
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
		url = strings.Replace(url, "RANDOM_UUID", uuid.New().String(), 1)
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
			match, err := regexp.Match(checks.ExpectedReponse, body)
			if err != nil {
				return nil
			}
			if !match {
				nc = append(nc, NetworkCheck{URL: url, Success: false, Error: false, Status: fmt.Sprintf("Expected %s got %s", checks.ExpectedReponse, string(body))})
				continue
			}
		}
		nc = append(nc, NetworkCheck{URL: url, Success: true, Error: false, Status: "ok"})
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

// CustomWriter to strip ascii characters
type CustomWriter struct {
	writer io.Writer
}

// NewCustomWriter creates a new CustomWriter.
func NewCustomWriter(writer io.Writer) *CustomWriter {
	return &CustomWriter{writer: writer}
}

// Write implements the io.Writer interface.
func (cw *CustomWriter) Write(p []byte) (n int, err error) {

	re := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	stripped := re.ReplaceAll(p, []byte{})

	filtered := make([]byte, 0, len(stripped))
	for _, b := range stripped {
		if b == '\n' || (unicode.IsPrint(rune(b)) || b < 0x20 || b > 0x7E) {
			filtered = append(filtered, b)
		}
	}
	return cw.writer.Write(filtered)
}
