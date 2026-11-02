package confirm

import "os"

// IsCharDevice reports whether f is a character device: a terminal, or a device
// such as /dev/null. A nil or closed file is not.
func IsCharDevice(f *os.File) bool {
	fi, err := f.Stat()
	return err == nil && fi.Mode()&os.ModeCharDevice != 0
}

// StdinIsTerminal is the default answer to "is the operator here to answer":
// stdin is a character device.
func StdinIsTerminal() bool { return IsCharDevice(os.Stdin) }
