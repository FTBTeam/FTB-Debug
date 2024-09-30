package fixes

import (
	"encoding/json"
	"errors"
	"fmt"
	"ftb-debug-ui/backend/shared"
	semVer "github.com/hashicorp/go-version"
	"io"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
)

const adoptiumApiUrl = "https://api.adoptium.net"

func DoGet(url string) (*http.Response, error) {
	headers := map[string][]string{
		"User-Agent": {"FTB-Debug-Tool/" + shared.Version},
	}
	resp, err := makeRequest("GET", url, headers)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return nil, errors.New(fmt.Sprintf("Error: %d\n%s", resp.StatusCode, b))
	}
	return resp, nil
}

func makeRequest(method, url string, requestHeaders map[string][]string) (*http.Response, error) {
	headers := map[string][]string{}
	for k, v := range requestHeaders {
		headers[k] = v
	}
	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header = headers

	return client.Do(req)
}

func GetJava(version string) (File, error) {
	adoptiumUrl, err := makeAdoptiumUrl(version)
	if err != nil {
		return File{}, err
	}

	get, err := DoGet(adoptiumUrl)
	if err != nil {
		return File{}, err
	}
	defer get.Body.Close()

	var adoptium Adoptium

	err = json.NewDecoder(get.Body).Decode(&adoptium)
	if err != nil {
		return File{}, err
	}

	var fileExt string
	if strings.HasSuffix(adoptium[0].Binary.Package.Name, ".zip") {
		fileExt = ".zip"
	} else if strings.HasSuffix(adoptium[0].Binary.Package.Name, ".tar.gz") {
		fileExt = ".tar.gz"
	} else {
		fileExt = "" // shrug
	}

	return File{
		Name:     "jre" + fileExt,
		Url:      adoptium[0].Binary.Package.Link,
		Hash:     adoptium[0].Binary.Package.Checksum,
		HashType: "sha256",
	}, nil
}

func makeAdoptiumUrl(version string) (string, error) {
	parsedUrl, err := url.Parse(adoptiumApiUrl + "/v3/assets/latest/" + version + "/hotspot")
	if err != nil {
		return "", err
	}

	q := parsedUrl.Query()
	q.Add("image_type", "jre")
	if runtime.GOOS == "windows" {
		q.Add("os", "windows")
	}
	if runtime.GOOS == "darwin" {
		q.Add("os", "mac")
	}
	if runtime.GOOS == "linux" {
		if _, err := os.Stat("/etc/alpine-release"); !os.IsNotExist(err) {
			q.Add("os", "alpine-linux")
		} else {
			q.Add("os", "linux")
		}
	}

	arch, err := validJavaArch(version)
	if err != nil {
		return "", err
	}
	q.Add("architecture", arch)

	parsedUrl.RawQuery = q.Encode()

	return parsedUrl.String(), nil
}

func validJavaArch(version string) (string, error) {
	targetVersion, err := semVer.NewVersion(version)
	if err != nil {
		return "", err
	}
	switch runtime.GOOS {
	case "darwin":
		if runtime.GOARCH == "arm64" {
			limit, err := semVer.NewVersion("11.0.0")
			if err != nil {
				return "", err
			}
			if targetVersion.LessThan(limit) {
				return "x64", nil
			}
			return "aarch64", nil
		}
		if runtime.GOARCH == "amd64" {
			return "x64", nil
		}
		if runtime.GOARCH == "386" {
			return "x86", nil
		}
	case "windows":
		if runtime.GOARCH == "amd64" || runtime.GOARCH == "arm64" {
			return "x64", nil
		}
		if runtime.GOARCH == "386" || runtime.GOARCH == "arm" {
			return "x86", nil
		}
	case "linux":
		if runtime.GOARCH == "amd64" {
			return "x64", nil
		}
		if runtime.GOARCH == "386" {
			return "x86", nil
		}
		if runtime.GOARCH == "arm64" {
			return "aarch64", nil
		}
		if runtime.GOARCH == "arm" {
			return "arm", nil
		}
	}
	return "", errors.New("unsupported architecture, please contact FTB support")
}
