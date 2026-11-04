//go:build windows

package watch

import (
	"os"
	"syscall"
)

// enableVT asks f's console to interpret escape sequences, which a Windows console
// (conhost, which PowerShell 5.1 runs in) does only once a program turns it on, and
// reports whether it now does. A handle that is not a console, or a console too old
// for it, reports false, and the loop appends instead of printing the escape bytes
// as text. It uses syscall's lazy DLL loading rather than golang.org/x/sys/windows,
// to keep this module on the standard library for one call.
func enableVT(f *os.File) bool {
	const enableVirtualTerminalProcessing = 0x0004
	h := syscall.Handle(f.Fd())
	var mode uint32
	if err := syscall.GetConsoleMode(h, &mode); err != nil {
		return false
	}
	if mode&enableVirtualTerminalProcessing != 0 {
		return true
	}
	setConsoleMode := syscall.NewLazyDLL("kernel32.dll").NewProc("SetConsoleMode")
	r, _, _ := setConsoleMode.Call(uintptr(h), uintptr(mode|enableVirtualTerminalProcessing))
	return r != 0
}
