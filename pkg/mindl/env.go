package mindl

import (
	"os"
	"runtime"
	"strconv"
)

// ShouldUpdate returns true if mindl should update the sumdb during execution.
// This is enabled if the environment variable MINDL_UPDATE contains a truthy value.
func ShouldUpdate() bool {
	val, err := strconv.ParseBool(os.Getenv("MINDL_UPDATE"))
	return err == nil && val
}

func getenv(name, def string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return def
}

// OS returns the value of MINDL_OS or runtime.GOOS.
func OS() string {
	return getenv("MINDL_OS", runtime.GOOS)
}

// Arch returns the value of MINDL_ARCH or runtime.GOARCH.
func Arch() string {
	return getenv("MINDL_ARCH", runtime.GOARCH)
}
