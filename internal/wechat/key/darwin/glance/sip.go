package glance

import (
	"os/exec"
	"strings"
)

// IsSIPDisabled checks if System Integrity Protection (SIP) is disabled on macOS.
// Returns true if SIP is disabled, false if it's enabled or if the status cannot be determined.
func IsSIPDisabled() bool {
	// Run the csrutil status command to check SIP status
	cmd := exec.Command("csrutil", "status")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// If there's an error running the command, assume SIP is enabled
		return false
	}

	// Convert output to string and check if SIP is disabled
	outputStr := strings.ToLower(string(output))

	// $ csrutil status
	// System Integrity Protection status: disabled.

	// If the output contains "disabled", SIP is disabled
	if strings.Contains(outputStr, "system integrity protection status: disabled") {
		return true
	}

	// Check for partial SIP disabling - some configurations might have specific protections disabled
	if strings.Contains(outputStr, "disabled") && strings.Contains(outputStr, "debugging") {
		return true
	}

	// By default, assume SIP is enabled
	return false
}
