// Package noticetest checks that docs carry the support notice from package notice.
//
// Check and CheckDocs are pure and return errors, so their rules are testable
// without a fake testing.T. Assert, AssertDocs and AssertASCII wrap them for a
// test. Nothing here owns a flag: a tool that regenerates docs under an update
// flag skips its own call while regenerating.
package noticetest

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/hafio/go-deploy-utility/notice"
)

// Options selects how strictly and where docs are checked.
type Options struct {
	// Strict requires the notice directly under the H1: title, one blank line,
	// notice, one blank line. Without it the notice may follow other lines (such
	// as a generated-file comment) but must come before the first "## " heading.
	Strict bool

	// Files lists the docs to check, as slash-separated paths relative to the
	// root. Empty means README.md plus docs/*.md, non-recursively.
	Files []string

	// Exclude lists root-relative slash paths to skip.
	Exclude []string
}

// fixHint is the remedy every placement error ends with.
const fixHint = "paste notice.Markdown() directly under the H1 with one blank line either side"

// Check verifies one doc. It normalises CRLF line endings first.
func Check(doc []byte, n notice.Notice, o Options) error {
	block := strings.TrimSuffix(n.Markdown(), "\n")
	if block == "" {
		return errors.New("zero Notice: build it with notice.New, or every doc would match an empty notice")
	}
	text := strings.ReplaceAll(string(doc), "\r\n", "\n")
	title, rest, _ := strings.Cut(text, "\n")
	if !strings.HasPrefix(title, "# ") || strings.TrimSpace(title[2:]) == "" {
		return fmt.Errorf("line 1 must be the H1 title (\"# ...\"), got %q", title)
	}
	if o.Strict {
		if strings.HasPrefix(rest, "\n"+block+"\n\n") {
			return nil
		}
		if strings.Contains(text, "\n\n"+block+"\n\n") {
			return fmt.Errorf("the support notice is present but not directly under the H1; %s", fixHint)
		}
		return describe(text, block)
	}
	i := strings.Index(text, "\n\n"+block+"\n\n")
	if i < 0 {
		return describe(text, block)
	}
	if h := strings.Index(text, "\n## "); h >= 0 && h < i {
		return fmt.Errorf("the support notice sits below the first section heading; %s", fixHint)
	}
	return nil
}

// describe explains why the notice block was not found where it belongs.
func describe(text, block string) error {
	lines := strings.Split(text, "\n")
	want := strings.Split(block, "\n")
	start := -1
	for i, ln := range lines {
		if ln == want[0] {
			start = i
			break
		}
	}
	if start < 0 {
		return fmt.Errorf("no support notice; %s:\n\n%s", fixHint, block)
	}
	for j, w := range want {
		k := start + j
		if k >= len(lines) {
			return fmt.Errorf("the support notice is cut short at line %d, want %q; %s", k+1, w, fixHint)
		}
		if lines[k] != w {
			return fmt.Errorf("the support notice differs at line %d:\n  got:  %q\n  want: %q\n%s", k+1, lines[k], w, fixHint)
		}
	}
	if start == 0 || lines[start-1] != "" {
		return fmt.Errorf("the line before the support notice (line %d) must be blank; %s", start, fixHint)
	}
	return fmt.Errorf("the line after the support notice (line %d) must be blank, or the next paragraph "+
		"renders inside the warning; %s", start+len(want)+1, fixHint)
}

// noDocsError reports that nothing was checked, which AssertDocs treats as fatal: a
// gate that checks no docs proves nothing. Either the glob found nothing, or
// Options.Exclude removed every doc that was selected; excluded counts those.
type noDocsError struct {
	pattern  string
	excluded int
}

func (e noDocsError) Error() string {
	if e.excluded > 0 {
		return fmt.Sprintf("all %d selected docs are in Options.Exclude, so nothing was checked; narrow Options.Exclude", e.excluded)
	}
	return fmt.Sprintf("no docs matched %s; pass Options.Files, or run from the repo root", e.pattern)
}

// CheckDocs checks every doc under root that Options selects and joins every
// failure, so one run reports them all. It is an error when nothing is left to
// check, whether the glob matched nothing or Options.Exclude removed every doc.
func CheckDocs(root string, n notice.Notice, o Options) error {
	files := o.Files
	if len(files) == 0 {
		pattern := filepath.Join(root, "docs", "*.md")
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return fmt.Errorf("glob %s: %w", pattern, err)
		}
		if len(matches) == 0 {
			return noDocsError{pattern: pattern}
		}
		sort.Strings(matches)
		for _, m := range matches {
			rel, err := filepath.Rel(root, m)
			if err != nil {
				return fmt.Errorf("relativise %s against %s: %w", m, root, err)
			}
			files = append(files, filepath.ToSlash(rel))
		}
		files = append(files, "README.md")
	}
	skip := make(map[string]bool, len(o.Exclude))
	for _, e := range o.Exclude {
		skip[e] = true
	}
	var errs []error
	checked := 0
	for _, rel := range files {
		if skip[rel] {
			continue
		}
		checked++
		b, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
		if err != nil {
			errs = append(errs, fmt.Errorf("read %s: %w", rel, err))
			continue
		}
		if err := Check(b, n, o); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", rel, err))
		}
	}
	if checked == 0 {
		return noDocsError{excluded: len(files)}
	}
	return errors.Join(errs...)
}

// Assert fails the test when doc does not carry the notice.
func Assert(t testing.TB, name string, doc []byte, n notice.Notice, o Options) {
	t.Helper()
	if err := Check(doc, n, o); err != nil {
		t.Errorf("%s: %v", name, err)
	}
}

// AssertDocs fails the test once per doc that does not carry the notice, and
// fatally when there was nothing to check.
func AssertDocs(t testing.TB, n notice.Notice, root string, o Options) {
	t.Helper()
	err := CheckDocs(root, n, o)
	if err == nil {
		return
	}
	var none noDocsError
	if errors.As(err, &none) {
		t.Fatalf("%v", err)
		return
	}
	if joined, ok := err.(interface{ Unwrap() []error }); ok {
		for _, e := range joined.Unwrap() {
			t.Errorf("%v", e)
		}
		return
	}
	t.Errorf("%v", err)
}

// AssertASCII fails the test for any byte outside printable ASCII (0x20-0x7e),
// naming the line and the offset. Rendered notice text reaches terminals, logs
// and PowerShell 5.1, so a tab is rejected too.
func AssertASCII(t testing.TB, what string, lines []string) {
	t.Helper()
	for _, ln := range lines {
		for i := 0; i < len(ln); i++ {
			if c := ln[i]; c < 0x20 || c > 0x7e {
				t.Errorf("%s has a non-ASCII or control byte 0x%02x at offset %d: %q", what, c, i, ln)
			}
		}
	}
}
