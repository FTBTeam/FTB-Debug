package fixes

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"ftb-debug/v2/shared"
	"github.com/cavaliergopher/grab/v3"
	"github.com/codeclysm/extract/v3"
	"github.com/pterm/pterm"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var (
	winPath   = filepath.Join(os.Getenv("LOCALAPPDATA"), ".ftba")
	linuxPath = filepath.Join(os.Getenv("HOME"), ".ftba")
	macPath   = filepath.Join(os.Getenv("HOME"), "Library", "Application Support", ".ftba")
)

func FixCommonIssues() {
	_, _ = pterm.DefaultInteractiveContinue.
		WithOptions([]string{"Ok"}).
		WithDefaultText("Please make sure the FTB App is closed before continuing").
		WithTextStyle(pterm.Error.MessageStyle).
		Show()

	ftbaPath := getFTBAPath()
	if ftbaPath == "" {
		pterm.Error.Println("Unsupported OS")
		return
	}

	if !shared.DoesPathExist(ftbaPath) {
		pterm.Error.Println("FTB App not found")
		return
	}

	dir, err := os.ReadDir(ftbaPath)
	if err != nil {
		pterm.Error.Println("Failed to read directory:", err)
		return
	}
	for _, file := range dir {
		if file.Name() == "bin" || file.Name() == "runtime" || file.Name() == "profiles.json" {
			err := os.RemoveAll(filepath.Join(ftbaPath, file.Name()))
			if err != nil {
				pterm.Error.Println("Failed to remove file:", err)
				return
			}
		}
	}

	err = installRuntime()
	if err != nil {
		pterm.Error.Println("Failed to install runtime:", err)
		return
	}

	pterm.Success.Println("Common issues fixed, you may now open the FTB App")
	pterm.Println()
	pterm.Warning.WithPrefix(pterm.Prefix{Text: "Info", Style: pterm.Warning.Prefix.Style}).
		WithMessageStyle(pterm.Warning.MessageStyle).
		Println("You will need to log your Minecraft account back into the app.\n" +
			"If you had a modpack(s) installed, right click on the modpack and click on settings and then click on the repair button.\n" +
			"A guide on this can be found here: https://docs.feed-the-beast.com/docs/app/Instances/repair")
}

func getFTBAPath() string {
	switch runtime.GOOS {
	case "windows":
		return winPath
	case "linux":
		return linuxPath
	case "darwin":
		return macPath
	default:
		return ""
	}
}

func getMeta() (AppMeta, error) {
	resp, err := DoGet("https://raw.githubusercontent.com/FTBTeam/FTB-App/refs/heads/main/subprocess/meta-template.json")
	if err != nil {
		return AppMeta{}, err
	}
	if resp.StatusCode != 200 {
		return AppMeta{}, errors.New("Error: " + resp.Status)
	}
	defer resp.Body.Close()

	var meta AppMeta

	err = json.NewDecoder(resp.Body).Decode(&meta)
	if err != nil {
		return AppMeta{}, err
	}

	return meta, nil
}

func installRuntime() error {
	ftbaPath := getFTBAPath()
	if ftbaPath == "" {
		return errors.New("unsupported OS")
	}

	meta, err := getMeta()
	if err != nil {
		pterm.Error.Println("Failed to get meta")
		return err
	}

	java, err := GetJava(meta.Runtime.Version)
	if err != nil {
		pterm.Error.Println("Failed to get java")
		return err
	}

	resp, err := grab.Get(filepath.Join(ftbaPath, "runtime", java.Name), java.Url)
	if err != nil {
		return err
	}
	pterm.Success.Println("Downloaded runtime:", resp.Filename)

	var shift = func(path string) string {
		// Apparently zips in windows can use / instead of \
		// So we need to check if the path is using / or \
		sep := filepath.Separator
		if len(strings.Split(path, "\\")) > 1 {
			sep = '\\'
		} else if len(strings.Split(path, "/")) > 1 {
			sep = '/'
		}

		parts := strings.Split(path, string(sep))
		parts = parts[1:]
		join := strings.Join(parts, string(sep))
		return join
	}

	javaFile, err := os.Open(filepath.Join(ftbaPath, "runtime", java.Name))
	if err != nil {
		pterm.Fatal.Println("Error opening java archive", err.Error())
	}
	javaPkg := bufio.NewReader(javaFile)

	err = extract.Archive(context.TODO(), javaPkg, filepath.Join(ftbaPath, "runtime"), shift)
	if err != nil {
		pterm.Fatal.Println("Error extracting java archive:", err.Error())
	}
	javaVersion := []byte(meta.Runtime.Version)
	err = os.WriteFile(filepath.Join(ftbaPath, "runtime", ".java-version"), javaVersion, 0644)
	if err != nil {
		pterm.Fatal.Println("Error writing to file:", err.Error())
	}

	javaFile.Close()
	err = os.Remove(filepath.Join(ftbaPath, "runtime", java.Name))
	if err != nil {
		pterm.Warning.Println("Error removing java archive:", err.Error())
	}

	return nil
}
