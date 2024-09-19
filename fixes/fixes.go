package fixes

import (
	"github.com/pterm/pterm"
	"os"
	"path/filepath"
	"runtime"
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
