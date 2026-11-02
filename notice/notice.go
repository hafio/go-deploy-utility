// Package notice renders the support notice that every doc of the tools and of
// this module carries directly under its title: a GitHub [!WARNING] alert saying
// the software is not a supported Solace product.
//
// The wording is one template. Only three things vary: the name (the binary or
// module), the kind word ("tool" or "module"), and the covers clause naming what
// the notice does not cover. The rendering is fixed so that every committed copy
// can be checked byte for byte (see package noticetest).
package notice

import (
	"fmt"
	"strings"
)

// Width is the maximum length of one rendered line, including the "> " prefix.
const Width = 80

// Notice is a validated support notice. Build it with New; the zero value
// renders nothing.
type Notice struct {
	name, kind, covers string
}

// New validates the three variable parts and returns the notice.
//
// name is the binary or module name, such as "solace-util": lower-case letters,
// digits and '-', starting with a letter or digit. kind is one lower-case word,
// such as "tool". covers is the clause after "not " without its final period,
// which the renderer adds; it must be printable ASCII with single spaces.
func New(name, kind, covers string) (Notice, error) {
	if err := checkName(name); err != nil {
		return Notice{}, err
	}
	if err := checkKind(kind); err != nil {
		return Notice{}, err
	}
	if err := checkCovers(covers); err != nil {
		return Notice{}, err
	}
	return Notice{name: name, kind: kind, covers: covers}, nil
}

func checkName(name string) error {
	ok := name != ""
	for i := 0; ok && i < len(name); i++ {
		c := name[i]
		ok = (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || (c == '-' && i > 0)
	}
	if !ok {
		return fmt.Errorf("notice: name %q is not a valid tool or module name "+
			"(want lower-case letters, digits and '-', starting with a letter or digit, such as \"solace-util\"); "+
			"pass the binary or module name", name)
	}
	return nil
}

func checkKind(kind string) error {
	ok := kind != ""
	for i := 0; ok && i < len(kind); i++ {
		ok = kind[i] >= 'a' && kind[i] <= 'z'
	}
	if !ok {
		return fmt.Errorf("notice: kind %q is not one lower-case word; pass \"tool\" or \"module\"", kind)
	}
	return nil
}

func checkCovers(covers string) error {
	switch {
	case covers == "":
		return fmt.Errorf("notice: covers is empty; pass the clause after \"not \", " +
			"such as \"the Solace PubSub+ Event Broker that it deploys\"")
	case strings.TrimSpace(covers) != covers || strings.Contains(covers, "  "):
		return fmt.Errorf("notice: covers %q has leading, trailing or double spaces; use single spaces between words", covers)
	case strings.HasSuffix(covers, "."):
		return fmt.Errorf("notice: covers %q ends with a period; drop it, the notice adds the final period", covers)
	}
	for i := 0; i < len(covers); i++ {
		if c := covers[i]; c < 0x20 || c > 0x7e {
			return fmt.Errorf("notice: covers %q has a non-ASCII or control byte 0x%02x at offset %d; "+
				"use printable ASCII only, since the notice reaches terminals, logs and PowerShell 5.1", covers, c, i)
		}
	}
	return nil
}

// text is the notice sentence before wrapping.
func (n Notice) text() string {
	return "**Not a supported Solace product.** `" + n.name + "` was created by Solace " +
		"Professional Services and is supported only by Solace Professional Services -- " +
		"not by Solace Support. For help with this " + n.kind + ", contact your Solace " +
		"Professional Services representative rather than opening a Solace Support case. " +
		"This notice covers this " + n.kind + " only, not " + n.covers + "."
}

// wrap fills words greedily into lines of at most Width bytes, each starting with
// prefix. A word longer than a line gets a line of its own rather than being split.
func wrap(text, prefix string) []string {
	var out []string
	cur := ""
	for _, w := range strings.Fields(text) {
		switch {
		case cur == "":
			cur = prefix + w
		case len(cur)+1+len(w) <= Width:
			cur += " " + w
		default:
			out = append(out, cur)
			cur = prefix + w
		}
	}
	if cur != "" {
		out = append(out, cur)
	}
	return out
}

// Lines is the notice as Markdown, one line per element, with no newlines: the
// "> [!WARNING]" line, then the sentence wrapped at Width. Each call returns a
// fresh slice.
func (n Notice) Lines() []string {
	if n == (Notice{}) {
		return nil
	}
	return append([]string{"> [!WARNING]"}, wrap(n.text(), "> ")...)
}

// Markdown is Lines joined by newlines, ending with one. A doc carries it
// directly under its H1 with one blank line either side.
func (n Notice) Markdown() string {
	lines := n.Lines()
	if lines == nil {
		return ""
	}
	return strings.Join(lines, "\n") + "\n"
}

// Plain is the notice reduced to its words: no quote markers, no alert tag, no
// bold or code-span markup, single spaces. It is for comparing a copy that is
// not Markdown, such as an HTML page.
func (n Notice) Plain() string {
	if n == (Notice{}) {
		return ""
	}
	return strings.Join(strings.Fields(strings.NewReplacer("**", "", "`", "").Replace(n.text())), " ")
}

// shortText is the version-command form; it names no tool, so it takes no input.
const shortText = "Not a supported Solace product. Created by Solace Professional Services and " +
	"supported only by Solace Professional Services -- not by Solace Support."

// Short is the two-line form a `version` command prints under its first line,
// wrapped at Width with no prefix. Each call returns a fresh slice.
func Short() []string {
	return wrap(shortText, "")
}
