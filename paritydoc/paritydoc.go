// Package paritydoc holds the canonical body of docs/parity.md, the Docker,
// Podman and Kubernetes contract that solmq-conn-util and solace-util share, and
// the check each tool runs on its own copy.
//
// Each tool's docs/parity.md is the Title, a blank line, the tool's own support
// notice, a blank line, then Body unchanged. The body is owned here: a change is
// made in body.md, released, and picked up by both tools in the same step.
package paritydoc

import (
	_ "embed"
	"errors"
	"fmt"
	"strings"
)

// Title is the H1 both tools' docs/parity.md carry.
const Title = "# Platform parity: solmq-conn-util and solace-util"

// body is the canonical text. go:embed needs a package-level variable; this is
// the module's one exception to "no globals". It is unexported, read only through
// Body and Check, and never assigned.
//
//go:embed body.md
var body string

// remedy ends every body mismatch, so the error says what to do next.
const remedy = "The body below the notice is shared and owned by go-deploy-utility (paritydoc/body.md): " +
	"do not edit one tool's copy. Change paritydoc/body.md, tag a release, and bump both tools."

// Body returns a copy of the canonical body.
func Body() []byte { return []byte(body) }

// Compose builds a whole docs/parity.md from a tool's notice block (as
// notice.Markdown renders it), so a tool can regenerate its page.
func Compose(noticeBlock string) ([]byte, error) {
	nb := strings.TrimSuffix(noticeBlock, "\n")
	if nb == "" {
		return nil, errors.New("paritydoc: notice block is empty; pass notice.Markdown() for the tool")
	}
	lines := strings.Split(nb, "\n")
	if lines[0] != "> [!WARNING]" {
		return nil, fmt.Errorf("paritydoc: notice block line 1 is %q, want \"> [!WARNING]\"; pass notice.Markdown() for the tool", lines[0])
	}
	for i, ln := range lines {
		if !strings.HasPrefix(ln, ">") {
			return nil, fmt.Errorf("paritydoc: notice block line %d is %q, not a quoted line; pass notice.Markdown() for the tool", i+1, ln)
		}
	}
	return []byte(Title + "\n\n" + nb + "\n\n" + body), nil
}

// Check verifies a tool's whole docs/parity.md: the Title on line 1, a blank line,
// a quoted notice starting "> [!WARNING]" on line 3, exactly one blank line, then
// Body byte for byte to the end of the file. CRLF endings are normalised first.
// The notice's wording is not judged here; package noticetest owns that.
func Check(doc []byte) error {
	text := strings.ReplaceAll(string(doc), "\r\n", "\n")
	if text == "" {
		return errors.New("paritydoc: docs/parity.md is empty; regenerate it with paritydoc.Compose(notice.Markdown())")
	}
	lines := strings.Split(text, "\n")
	if lines[0] != Title {
		return fmt.Errorf("paritydoc: docs/parity.md line 1 is the title:\n  got:  %q\n  want: %q", clip(lines[0]), Title)
	}
	if len(lines) < 2 || lines[1] != "" {
		return errors.New("paritydoc: docs/parity.md line 2 must be blank, between the title and the support notice")
	}
	if len(lines) < 3 || lines[2] != "> [!WARNING]" {
		return errors.New("paritydoc: docs/parity.md line 3 must open the support notice with \"> [!WARNING]\"; " +
			"regenerate the page with paritydoc.Compose(notice.Markdown())")
	}
	end := 2
	for end < len(lines) && strings.HasPrefix(lines[end], ">") {
		end++
	}
	// lines[end] is the first line after the notice (0-based); it must be blank,
	// and the body starts on the line after it.
	if end >= len(lines) || lines[end] != "" {
		return fmt.Errorf("paritydoc: docs/parity.md line %d must be blank, closing the support notice that runs from line 3", end+1)
	}
	start := end + 1
	if start < len(lines) && lines[start] == "" {
		return fmt.Errorf("paritydoc: docs/parity.md has more than one blank line after the support notice (line %d); keep exactly one", start+1)
	}
	got := strings.Join(lines[start:], "\n")
	if got == body {
		return nil
	}
	if got+"\n" == body {
		return errors.New("paritydoc: docs/parity.md is missing its final newline. " + remedy)
	}
	// Compare without the final newline so line numbers name real lines: a split
	// of text that ends in "\n" has an empty last element that is not a line.
	return firstDifference(strings.Split(strings.TrimSuffix(got, "\n"), "\n"),
		strings.Split(strings.TrimSuffix(body, "\n"), "\n"), start)
}

// firstDifference reports the first line where the doc's body region (got,
// starting at 0-based doc line offset) departs from the canonical body (want).
// Check calls it only when the two differ, so one of the three cases applies.
func firstDifference(got, want []string, offset int) error {
	k := 0
	for k < len(got) && k < len(want) && got[k] == want[k] {
		k++
	}
	docLine, bodyLine := offset+k+1, k+1
	switch {
	case k >= len(got):
		return fmt.Errorf("paritydoc: docs/parity.md ends after body line %d; next expected line %d: %q. %s",
			k, bodyLine, clip(want[k]), remedy)
	case k >= len(want):
		return fmt.Errorf("paritydoc: docs/parity.md has an extra line %d after the end of the body: %q. %s",
			docLine, clip(got[k]), remedy)
	default:
		return fmt.Errorf("paritydoc: docs/parity.md differs from the canonical body at doc line %d (body line %d):\n"+
			"  got:  %q\n  want: %q\n%s", docLine, bodyLine, clip(got[k]), clip(want[k]), remedy)
	}
}

// clip shortens a line for an error message only.
func clip(s string) string {
	if len(s) > 120 {
		return s[:120] + "..."
	}
	return s
}
