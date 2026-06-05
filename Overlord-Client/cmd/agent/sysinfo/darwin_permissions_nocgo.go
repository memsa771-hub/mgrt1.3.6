//go:build darwin && !cgo

package sysinfo

func darwinAccessibilityPermission() bool {
	return false
}

func darwinScreenRecordingPermission() bool {
	return false
}
