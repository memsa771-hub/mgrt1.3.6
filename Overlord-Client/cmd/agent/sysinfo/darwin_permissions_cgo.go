//go:build darwin && cgo

package sysinfo

/*
#cgo LDFLAGS: -framework ApplicationServices -framework CoreGraphics

#include <ApplicationServices/ApplicationServices.h>
#include <CoreGraphics/CoreGraphics.h>
#include <stdbool.h>

extern bool CGPreflightScreenCaptureAccess(void) __attribute__((weak_import));

static int checkAccessibilityPermission(void) {
	return AXIsProcessTrusted() ? 1 : 0;
}

static int checkScreenRecordingPermission(void) {
	if (CGPreflightScreenCaptureAccess == NULL) return 0;
	return CGPreflightScreenCaptureAccess() ? 1 : 0;
}
*/
import "C"

func darwinAccessibilityPermission() bool {
	return C.checkAccessibilityPermission() == 1
}

func darwinScreenRecordingPermission() bool {
	return C.checkScreenRecordingPermission() == 1
}
