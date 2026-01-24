package main

import (
	"os/exec"
	"strings"
)

// GetClipboardText retrieves the current clipboard text using xclip
func GetClipboardText() (string, error) {
	cmd := exec.Command("xclip", "-selection", "clipboard", "-o")
	output, err := cmd.Output()
	if err != nil {
		// Empty clipboard or xclip not available
		if exitErr, ok := err.(*exec.ExitError); ok {
			// Exit code 1 typically means empty clipboard
			if exitErr.ExitCode() == 1 {
				return "", nil
			}
		}
		return "", err
	}
	return string(output), nil
}

// SetClipboardText sets the clipboard text using xclip
func SetClipboardText(text string) error {
	cmd := exec.Command("xclip", "-selection", "clipboard")
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}

// CheckXclipAvailable checks if xclip is installed
func CheckXclipAvailable() bool {
	_, err := exec.LookPath("xclip")
	return err == nil
}

// CheckSmbclientAvailable checks if smbclient is installed
func CheckSmbclientAvailable() bool {
	_, err := exec.LookPath("smbclient")
	return err == nil
}
