//go:build !windows

package watch

import "os"

// enableVT reports whether f's terminal interprets escape sequences, which every
// terminal outside Windows does.
func enableVT(*os.File) bool { return true }
