//go:build windows

package watch

import (
	"os"
	"syscall"
	"testing"
)

// TestEnableVTOnAConsole runs the console branch of enableVT that TestCanRedraw
// cannot reach with the null device: on a real console the mode is read, escape
// processing is switched on when it is off, and the answer matches the mode the
// console is left in. It skips where the process has no console, as on a hosted
// runner, and restores the console's mode afterwards.
func TestEnableVTOnAConsole(t *testing.T) {
	const enableVirtualTerminalProcessing = 0x0004
	f, err := os.OpenFile("CONOUT$", os.O_RDWR, 0)
	if err != nil {
		t.Skipf("no console to test against: %v", err)
	}
	t.Cleanup(func() { f.Close() })
	h := syscall.Handle(f.Fd())
	var before uint32
	if err := syscall.GetConsoleMode(h, &before); err != nil {
		t.Skipf("CONOUT$ opened but is not a console: %v", err)
	}
	setConsoleMode := syscall.NewLazyDLL("kernel32.dll").NewProc("SetConsoleMode")
	t.Cleanup(func() { setConsoleMode.Call(uintptr(h), uintptr(before)) })

	got := enableVT(f)
	var after uint32
	if err := syscall.GetConsoleMode(h, &after); err != nil {
		t.Fatalf("reading the console mode back: %v", err)
	}
	if want := after&enableVirtualTerminalProcessing != 0; got != want {
		t.Errorf("enableVT = %v, but the console's escape processing is now %v (mode 0x%x, was 0x%x)",
			got, want, after, before)
	}
	if fi, err := f.Stat(); err == nil && fi.Mode()&os.ModeCharDevice != 0 {
		if redraw := CanRedraw(f); redraw != got {
			t.Errorf("CanRedraw = %v on the console, enableVT = %v; they must agree", redraw, got)
		}
	}
}
