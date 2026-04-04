package mindl

import (
	"os"
	"runtime"
	"strconv"
)

// ShouldUpdate returns true if mindl should update the sumdb during execution.
// This is enabled if the environment variable MINDL_UPDATE contains a truthy value.
func ShouldUpdate() bool {
	_, err := strconv.ParseBool(os.Getenv("MINDL_UPDATE"))
	return err == nil
}

func getenv(name, def string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return def
}

func OS() string {
	return getenv("MINDL_OD", runtime.GOOS)
}

func Arch() string {
	return getenv("MINDL_ARCH", runtime.GOARCH)
}
