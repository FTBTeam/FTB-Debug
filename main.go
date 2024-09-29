package main

import (
	"flag"
	"fmt"
	ftbdbg "ftb-debug/v2/dbg"
	"ftb-debug/v2/fixes"
	"ftb-debug/v2/shared"
	"github.com/eiannone/keyboard"
	"github.com/pterm/pterm"
	"github.com/pterm/pterm/putils"
	"time"
)

func init() {
	verboseLogging := flag.Bool("v", false, "Enable verbose logging")
	noColours := flag.Bool("no-colours", false, "Disable colours in output")
	flag.Parse()

	if *verboseLogging {
		pterm.EnableDebugMessages()
	}
	if *noColours {
		pterm.DisableColor()
	}

	pterm.Debug.Prefix = pterm.Prefix{
		Text:  "DEBUG",
		Style: pterm.NewStyle(pterm.BgLightMagenta, pterm.FgBlack),
	}
	pterm.Debug.MessageStyle = pterm.NewStyle(98)

	if shared.GitCommit == "" {
		shared.GitCommit = "Dev"
	}
	if shared.Version == "" {
		shared.Version = "0.0.0"
	}

	logo, _ := pterm.DefaultBigText.WithLetters(
		putils.LettersFromStringWithStyle("F", pterm.NewStyle(pterm.FgCyan)),
		putils.LettersFromStringWithStyle("T", pterm.NewStyle(pterm.FgGreen)),
		putils.LettersFromStringWithStyle("B", pterm.NewStyle(pterm.FgRed))).Srender()
	pterm.DefaultCenter.Println(logo)
	pterm.DefaultCenter.WithCenterEachLineSeparately().Println(fmt.Sprintf("Version: %s-%s\n%s", shared.Version, shared.GitCommit, time.Now().UTC().Format(time.RFC1123)))
	pterm.Debug.Println("Verbose logging enabled")
}

func main() {
	options := []string{
		"Run FTB App diagnostics",
		"Fix common FTB App issues",
	}

	// Use PTerm's interactive select feature to present the options to the user and capture their selection
	// The Show() method displays the options and waits for the user's input
	selectedOption, _ := pterm.DefaultInteractiveSelect.
		WithOptions(options).
		WithFilter(false).
		Show()

	switch selectedOption {
	case "Run FTB App diagnostics":
		ftbdbg.RunDebug()
	case "Fix common FTB App issues":
		fixes.FixCommonIssues()
	default:
		pterm.Error.Println("Invalid option selected")
	}

	pterm.Println(pterm.LightCyan("Press ESC to exit..."))

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
