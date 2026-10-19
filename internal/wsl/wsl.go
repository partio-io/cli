package wsl

import (
	"os"
	"strings"
)

// IsWSL returns true when Partio is running inside a WSL environment with
// Windows interop enabled. Both conditions must hold:
//  1. /proc/sys/kernel/osrelease contains "microsoft" (case-insensitive).
//     This covers both WSL1 and WSL2.
//  2. /proc/sys/fs/binfmt_misc/WSLInterop exists. This file is present only
//     when Windows interop is enabled, which is required before calling any
//     Windows executables (tasklist.exe, cmd.exe, wslpath).
//
// A WSL environment with interop disabled is treated the same as plain Linux —
// IsWSL returns false so callers never attempt to invoke Windows executables.
func IsWSL() bool {
	data, err := os.ReadFile("/proc/sys/kernel/osrelease")
	if err != nil {
		return false
	}
	if !strings.Contains(strings.ToLower(string(data)), "microsoft") {
		return false
	}
	_, err = os.Stat("/proc/sys/fs/binfmt_misc/WSLInterop")
	return err == nil
}
