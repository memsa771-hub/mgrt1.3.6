//go:build darwin && !ios && !ios_target

package sysinfo

import (
	"errors"
	"io"
	"os"
	"path/filepath"
)

func IsAdmin() bool {
	return os.Getuid() == 0
}

func Elevation() string {
	if os.Getuid() == 0 {
		return "admin"
	}
	return ""
}

func DarwinPermissions() map[string]bool {
	return map[string]bool{
		"screenRecording": darwinScreenRecordingPermission(),
		"accessibility":   darwinAccessibilityPermission(),
		"fullDiskAccess":  darwinFullDiskAccessPermission(),
		"root":            os.Getuid() == 0,
	}
}

func darwinFullDiskAccessPermission() bool {
	home, _ := os.UserHomeDir()
	candidates := []string{
		"/Library/Application Support/com.apple.TCC/TCC.db",
	}
	if home != "" {
		candidates = append(candidates,
			filepath.Join(home, "Library", "Mail"),
			filepath.Join(home, "Library", "Messages"),
			filepath.Join(home, "Library", "Safari"),
			filepath.Join(home, "Library", "Calendars"),
			filepath.Join(home, "Library", "Application Support", "AddressBook"),
		)
	}

	for _, path := range candidates {
		if canOpenProtectedPath(path) {
			return true
		}
	}
	return false
}

func canOpenProtectedPath(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return false
	}
	if !info.IsDir() {
		return true
	}
	_, err = f.Readdirnames(1)
	return err == nil || errors.Is(err, io.EOF)
}
