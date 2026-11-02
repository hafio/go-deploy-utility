package paritydoc

import (
	"strings"
	"testing"

	"github.com/hafio/go-deploy-utility/notice"
)

// noticeBlock renders a tool's notice the way its docs/parity.md carries it.
func noticeBlock(t *testing.T, name, covers string) string {
	t.Helper()
	n, err := notice.New(name, "tool", covers)
	if err != nil {
		t.Fatal(err)
	}
	return n.Markdown()
}

func connectorBlock(t *testing.T) string {
	return noticeBlock(t, "solmq-conn-util", "the Solace PubSub+ Connector for IBM MQ that it configures and deploys")
}

func brokerBlock(t *testing.T) string {
	return noticeBlock(t, "solace-util", "the Solace PubSub+ Event Broker or the EventBroker Operator that it deploys and operates")
}

// page is a valid docs/parity.md for the connector tool.
func page(t *testing.T) string {
	t.Helper()
	b, err := Compose(connectorBlock(t))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// TestBodyShape guards the embedded file against an editor adding a BOM, CRLF
// endings, a tab, non-ASCII text or a trailing blank line, any of which would
// break every tool's parity check at once.
func TestBodyShape(t *testing.T) {
	b := Body()
	if len(b) == 0 {
		t.Fatal("Body() is empty")
	}
	for i, c := range b {
		if c != '\n' && (c < 0x20 || c > 0x7e) {
			t.Fatalf("body has byte 0x%02x at offset %d; it must be plain ASCII with LF endings", c, i)
		}
	}
	s := string(b)
	if !strings.HasSuffix(s, "\n") || strings.HasSuffix(s, "\n\n") {
		t.Errorf("body must end with exactly one newline")
	}
	lines := strings.Split(strings.TrimSuffix(s, "\n"), "\n")
	if len(lines) != 157 {
		t.Errorf("body has %d lines, want 157; update this count when the body changes on purpose", len(lines))
	}
	if lines[0] == "" || strings.HasPrefix(lines[0], "#") || strings.HasPrefix(lines[0], ">") {
		t.Errorf("body line 1 = %q, want the intro paragraph", lines[0])
	}
	if lines[10] != "## Contract" {
		t.Errorf("body line 11 = %q, want the first section heading \"## Contract\"", lines[10])
	}
}

// TestBodyIsACopy: a caller that edits the returned bytes cannot change the next caller's.
func TestBodyIsACopy(t *testing.T) {
	b := Body()
	b[0] = 'X'
	if Body()[0] == 'X' {
		t.Error("Body() shares its backing array between calls")
	}
}

// TestCheckAcceptsEitherToolsNotice: the tool-specific notice sits outside the
// compared region, so both tools' pages pass with their own wording.
func TestCheckAcceptsEitherToolsNotice(t *testing.T) {
	for name, block := range map[string]string{"solmq-conn-util": connectorBlock(t), "solace-util": brokerBlock(t)} {
		doc, err := Compose(block)
		if err != nil {
			t.Fatalf("%s: Compose: %v", name, err)
		}
		if err := Check(doc); err != nil {
			t.Errorf("%s: Check: %v", name, err)
		}
	}
	// A longer notice is still a notice: its wording belongs to noticetest.
	longer := Title + "\n\n" + strings.TrimSuffix(connectorBlock(t), "\n") + "\n> one more quoted line\n\n" + string(Body())
	if err := Check([]byte(longer)); err != nil {
		t.Errorf("longer notice: %v", err)
	}
}

// TestCheckCRLF: a Windows checkout passes.
func TestCheckCRLF(t *testing.T) {
	if err := Check([]byte(strings.ReplaceAll(page(t), "\n", "\r\n"))); err != nil {
		t.Errorf("CRLF page: %v", err)
	}
}

// TestCheckTitle: the H1 is shared too, so a changed or padded title fails with
// both versions shown.
func TestCheckTitle(t *testing.T) {
	for _, title := range []string{"# Platform parity", Title + " "} {
		doc := strings.Replace(page(t), Title, title, 1)
		err := Check([]byte(doc))
		if err == nil || !strings.Contains(err.Error(), "line 1 is the title") || !strings.Contains(err.Error(), Title) {
			t.Errorf("title %q: err = %v, want got and want titles", title, err)
		}
	}
}

// TestCheckBoundaryStructure: each way the frame around the body can break is
// reported with what to fix.
func TestCheckBoundaryStructure(t *testing.T) {
	block := strings.TrimSuffix(connectorBlock(t), "\n")
	body := string(Body())
	for _, tc := range []struct {
		name, doc, want string
	}{
		{"title only", Title, "line 2 must be blank"},
		{"no blank line 2", Title + "\nprose\n" + block + "\n\n" + body, "line 2 must be blank"},
		{"line 3 is not the alert", Title + "\n\n> [!NOTE]\n> x\n\n" + body, "line 3 must open"},
		{"no notice at all", Title + "\n\n" + body, "line 3 must open"},
		{"notice runs to the end", Title + "\n\n" + block, "line 10 must be blank"},
		{"no blank line after the notice", Title + "\n\n" + block + "\n" + body, "line 10 must be blank"},
		{"two blank lines after the notice", Title + "\n\n" + block + "\n\n\n" + body, "more than one blank line"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := Check([]byte(tc.doc))
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Errorf("err = %v, want it to contain %q", err, tc.want)
			}
		})
	}
}

// TestCheckReportsFirstDifference: a body mismatch names the doc line and the
// body line, shows both texts, and says to change the shared copy instead.
func TestCheckReportsFirstDifference(t *testing.T) {
	good := page(t)
	lines := strings.Split(strings.TrimSuffix(good, "\n"), "\n")
	// Doc lines 1-10 are title, blank, the 7-line notice and a blank; body line N
	// is doc line N+10.
	edit := func(docLine int, text string) string {
		c := append([]string(nil), lines...)
		c[docLine-1] = text
		return strings.Join(c, "\n") + "\n"
	}
	for _, tc := range []struct {
		name string
		doc  string
		want []string
	}{
		{"a changed line", edit(41, "changed by hand"),
			[]string{"doc line 41", "body line 31", "changed by hand", "Change paritydoc/body.md"}},
		{"the last line changed", edit(167, "changed last line"),
			[]string{"doc line 167", "body line 157", "changed last line"}},
		{"the last line deleted", strings.Join(lines[:len(lines)-1], "\n") + "\n",
			[]string{"ends after body line 156", "next expected line 157"}},
		{"a line appended", good + "extra\n",
			[]string{"extra line 168", "extra"}},
		{"a blank line appended", good + "\n",
			[]string{"extra line 168"}},
		{"the final newline dropped", strings.TrimSuffix(good, "\n"),
			[]string{"missing its final newline"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := Check([]byte(tc.doc))
			if err == nil {
				t.Fatal("a page with a changed body passed")
			}
			for _, w := range tc.want {
				if !strings.Contains(err.Error(), w) {
					t.Errorf("err = %v, want it to contain %q", err, w)
				}
			}
		})
	}
	// A long line is clipped in the message only.
	long := strings.Repeat("y", 200)
	err := Check([]byte(edit(41, long)))
	if err == nil || strings.Contains(err.Error(), long) || !strings.Contains(err.Error(), "...") {
		t.Errorf("long line: err = %v, want the line clipped with ...", err)
	}
}

// TestCheckEmptyAndNil: no input panics; each is reported.
func TestCheckEmptyAndNil(t *testing.T) {
	for name, doc := range map[string][]byte{"nil": nil, "empty": {}, "only a newline": []byte("\n")} {
		if err := Check(doc); err == nil {
			t.Errorf("%s: Check accepted it", name)
		}
	}
}

// TestComposeRoundTrip: Compose output always passes Check, and Compose refuses
// a block that is not a quoted notice.
func TestComposeRoundTrip(t *testing.T) {
	doc, err := Compose(brokerBlock(t))
	if err != nil {
		t.Fatal(err)
	}
	if err := Check(doc); err != nil {
		t.Errorf("Check(Compose(block)): %v", err)
	}
	for _, tc := range []struct{ block, want string }{
		{"", "empty"},
		{"\n", "empty"},
		{"> [!NOTE]\n> x\n", "line 1"},
		{"> [!WARNING]\nnot quoted\n", "line 2"},
	} {
		if _, err := Compose(tc.block); err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("Compose(%q): err = %v, want it to mention %q", tc.block, err, tc.want)
		}
	}
}
