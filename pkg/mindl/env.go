package mindl

import (
	"os"
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
