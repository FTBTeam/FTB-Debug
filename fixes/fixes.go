package fixes

import (
	"github.com/pterm/pterm"
	"os"
)

const (
	winPath   = ""
	linuxPath = ""
	macPath   = ""
)

func FixCommonIssues() {
	result, _ := pterm.DefaultInteractiveConfirm.WithDefaultText("Do you want to fix common issues?\nThis may cause instances to not start, if this happens please go here").WithDefaultValue(true).Show()

	// Print a blank line for better readability.
	pterm.Println()

	if result {
		pterm.Info.Println("Fixing common issues...")
		// Fix common issues
	} else {
		pterm.Info.Println("Not fixing common issues, exiting...")
		os.Exit(1)
	}
}
