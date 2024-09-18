package main

import (
	"flag"
	ftbdbg "ftb-debug/v2/dbg"
	"ftb-debug/v2/fixes"
	"github.com/pterm/pterm"
	"os"
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
}

func main() {
	options := []string{
		"Debug checks",
		"Fix common issues",
	}

	// Use PTerm's interactive select feature to present the options to the user and capture their selection
	// The Show() method displays the options and waits for the user's input
	selectedOption, _ := pterm.DefaultInteractiveSelect.WithOptions(options).Show()

	if selectedOption == "Debug checks" {
		// Run dbg checks
		ftbdbg.RunDebug()
	}

	if selectedOption == "Fix common issues" {
		// Fix common issues
		fixes.FixCommonIssues()
	}

	pterm.Error.Println("Invalid option selected")
	os.Exit(1)
}
