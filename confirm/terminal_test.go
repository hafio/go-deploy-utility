package confirm

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// swapStdin points os.Stdin at f for the rest of the test.
func swapStdin(t *testing.T, f *os.File) {
	t.Helper()
	old := os.Stdin
	os.Stdin = f
	t.Cleanup(func() { os.Stdin = old })
}

// TestIsCharDeviceReportsAPipe: neither end of a pipe is a terminal, which is
// what a script or CI job hands a tool.
func TestIsCharDeviceReportsAPipe(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	defer w.Close()
	if IsCharDevice(r) || IsCharDevice(w) {
		t.Error("a pipe was reported as a character device")
	}
	if IsCharDevice(nil) {
		t.Error("a nil file was reported as a character device")
	}
}

// TestStdinIsTerminalFalseForPipeFileAndClosed covers the three non-terminal
// stdins a tool meets. It never asserts true: under `go test` the real stdin is
// /dev/null, which is a character device, so a true answer would be host-shaped.
func TestStdinIsTerminalFalseForPipeFileAndClosed(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	defer r.Close()
	file, err := os.Create(filepath.Join(t.TempDir(), "stdin.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	closed, err := os.Create(filepath.Join(t.TempDir(), "closed.txt"))
	if err != nil {
		t.Fatal(err)
	}
	closed.Close()

	for name, f := range map[string]*os.File{"pipe": r, "regular file": file, "closed file": closed} {
		swapStdin(t, f)
		if StdinIsTerminal() {
			t.Errorf("StdinIsTerminal() with stdin on a %s = true, want false", name)
		}
	}
}

// TestIsCharDeviceDevNull documents why the default detector must always be
// replaced in tests: /dev/null counts as a character device.
func TestIsCharDeviceDevNull(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the Unix case documents the point; NUL's mode is reported differently on Windows")
	}
	f, err := os.Open(os.DevNull)
	if err != nil {
		t.Skipf("cannot open %s: %v", os.DevNull, err)
	}
	defer f.Close()
	if !IsCharDevice(f) {
		t.Errorf("%s is not reported as a character device", os.DevNull)
	}
}
