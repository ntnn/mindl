package mindl

import "os"

const executableMask = 0111

// MakeExecutable marks the file at the given path executable without changing other bits.
func MakeExecutable(target string) error {
	info, err := os.Stat(target)
	if err != nil {
		return err
	}
	return os.Chmod(target, info.Mode()|executableMask)
}
