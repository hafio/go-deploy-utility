package noticetest

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hafio/go-deploy-utility/notice"
)

func testNotice(t *testing.T) notice.Notice {
	t.Helper()
	n, err := notice.New("solace-util", "tool", "the Solace PubSub+ Event Broker or the EventBroker Operator that it deploys and operates")
	if err != nil {
		t.Fatal(err)
	}
	return n
}

// block is the notice without its trailing newline, as it sits in a doc.
func block(n notice.Notice) string { return strings.TrimSuffix(n.Markdown(), "\n") }

// handWritten is the shape every hand-written doc has; generated is the shape
// solmq-conn-util's generated docs have, with a comment between H1 and notice.
func handWritten(n notice.Notice) string {
	return "# Title\n\n" + block(n) + "\n\nIntro paragraph.\n\n## Section\n\nBody.\n"
}

func generated(n notice.Notice) string {
	return "# Title\n\n<!-- GENERATED -- do not edit\nby hand. -->\n\n" + block(n) + "\n\nBody.\n"
}

// TestCheckAcceptsBothToolShapes: the strict shape passes both modes; the
// generated shape passes the relaxed mode only, and strict mode says why.
func TestCheckAcceptsBothToolShapes(t *testing.T) {
	n := testNotice(t)
	for _, strict := range []bool{false, true} {
		if err := Check([]byte(handWritten(n)), n, Options{Strict: strict}); err != nil {
			t.Errorf("hand-written doc, Strict=%v: %v", strict, err)
		}
	}
	if err := Check([]byte(generated(n)), n, Options{}); err != nil {
		t.Errorf("generated doc, relaxed: %v", err)
	}
	err := Check([]byte(generated(n)), n, Options{Strict: true})
	if err == nil || !strings.Contains(err.Error(), "not directly under the H1") {
		t.Errorf("generated doc, Strict: err = %v, want it to say the notice is not directly under the H1", err)
	}
}

// TestCheckCRLF: a Windows checkout with CRLF endings passes, mixed endings too.
func TestCheckCRLF(t *testing.T) {
	n := testNotice(t)
	crlf := strings.ReplaceAll(handWritten(n), "\n", "\r\n")
	if err := Check([]byte(crlf), n, Options{Strict: true}); err != nil {
		t.Errorf("CRLF doc: %v", err)
	}
	mixed := strings.Replace(handWritten(n), "\n", "\r\n", 3)
	if err := Check([]byte(mixed), n, Options{Strict: true}); err != nil {
		t.Errorf("mixed-ending doc: %v", err)
	}
}

// TestCheckRejectsMissingNotice: the error carries the block to paste.
func TestCheckRejectsMissingNotice(t *testing.T) {
	n := testNotice(t)
	err := Check([]byte("# Title\n\nJust prose.\n"), n, Options{})
	if err == nil {
		t.Fatal("a doc with no notice passed")
	}
	if !strings.Contains(err.Error(), block(n)) || !strings.Contains(err.Error(), "directly under the H1") {
		t.Errorf("err = %v, want the block to paste and where to paste it", err)
	}
}

// TestCheckRejectsAlteredNotice: one changed word, such as another tool's
// notice pasted in, is reported by line with got and want.
func TestCheckRejectsAlteredNotice(t *testing.T) {
	n := testNotice(t)
	doc := strings.Replace(handWritten(n), "contact your Solace", "contact any Solace", 1)
	err := Check([]byte(doc), n, Options{Strict: true})
	if err == nil {
		t.Fatal("an altered notice passed")
	}
	for _, want := range []string{"differs at line 6", "contact any Solace", "contact your Solace"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("err = %v, want it to contain %q", err, want)
		}
	}
	short := "# Title\n\n" + strings.Join(strings.Split(block(n), "\n")[:3], "\n")
	if err := Check([]byte(short), n, Options{}); err == nil || !strings.Contains(err.Error(), "cut short") {
		t.Errorf("truncated notice: err = %v, want it to say the notice is cut short", err)
	}
}

// TestCheckRequiresBlankLines: without a blank line after it, the next paragraph
// renders inside the warning; without one before it, it is not its own block.
func TestCheckRequiresBlankLines(t *testing.T) {
	n := testNotice(t)
	noAfter := "# Title\n\n" + block(n) + "\nIntro folded into the quote.\n"
	if err := Check([]byte(noAfter), n, Options{}); err == nil || !strings.Contains(err.Error(), "after the support notice") {
		t.Errorf("no blank line after: err = %v", err)
	}
	noBefore := "# Title\nintro\n" + block(n) + "\n\nBody.\n"
	if err := Check([]byte(noBefore), n, Options{}); err == nil || !strings.Contains(err.Error(), "before the support notice") {
		t.Errorf("no blank line before: err = %v", err)
	}
}

// TestCheckNoticeBelowFirstSection: in relaxed mode the notice may move down,
// but never below the first section heading.
func TestCheckNoticeBelowFirstSection(t *testing.T) {
	n := testNotice(t)
	doc := "# Title\n\n## First\n\n" + block(n) + "\n\nBody.\n"
	err := Check([]byte(doc), n, Options{})
	if err == nil || !strings.Contains(err.Error(), "below the first section heading") {
		t.Errorf("err = %v, want it to report the notice below the first heading", err)
	}
}

// TestCheckRequiresH1: a doc must open with its title.
func TestCheckRequiresH1(t *testing.T) {
	n := testNotice(t)
	for _, doc := range []string{
		"## Not a title\n\n" + block(n) + "\n\n",
		"\n# Title after a blank\n\n" + block(n) + "\n\n",
		"---\nfront: matter\n---\n# Title\n\n" + block(n) + "\n\n",
		"# \n\n" + block(n) + "\n\n",
	} {
		if err := Check([]byte(doc), n, Options{}); err == nil || !strings.Contains(err.Error(), "line 1 must be the H1") {
			t.Errorf("doc %q: err = %v, want the H1 rule", doc[:12], err)
		}
	}
}

// TestCheckZeroNotice: an unbuilt Notice renders nothing, and an empty block
// would match every doc, so Check refuses it.
func TestCheckZeroNotice(t *testing.T) {
	err := Check([]byte(handWritten(testNotice(t))), notice.Notice{}, Options{})
	if err == nil || !strings.Contains(err.Error(), "notice.New") {
		t.Errorf("err = %v, want it to say to build the notice with notice.New", err)
	}
}

// writeTree creates files under a temp root and returns the root.
func writeTree(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for rel, body := range files {
		p := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

// TestCheckDocsGlobExcludesHandoff: the default glob is README.md plus docs/*.md
// and does not descend, so a history folder such as docs/handoff/ is never checked.
func TestCheckDocsGlobExcludesHandoff(t *testing.T) {
	n := testNotice(t)
	root := writeTree(t, map[string]string{
		"README.md":          handWritten(n),
		"docs/a.md":          handWritten(n),
		"docs/handoff/b.md":  "# History\n\nNo notice here.\n",
		"docs/handoff/c.txt": "ignored",
	})
	if err := CheckDocs(root, n, Options{Strict: true}); err != nil {
		t.Errorf("CheckDocs: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "docs", "a.md"), []byte("# A\n\nNo notice.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := CheckDocs(root, n, Options{Strict: true})
	if err == nil || !strings.Contains(err.Error(), "docs/a.md") {
		t.Errorf("err = %v, want it to name docs/a.md", err)
	}
}

// TestCheckDocsNoDocs: a glob that matches nothing is an error naming the
// pattern, not a silent pass.
func TestCheckDocsNoDocs(t *testing.T) {
	root := t.TempDir()
	err := CheckDocs(root, testNotice(t), Options{})
	var none noDocsError
	if !errors.As(err, &none) || !strings.Contains(err.Error(), filepath.Join(root, "docs", "*.md")) {
		t.Errorf("err = %v, want a no-docs error naming the pattern", err)
	}
}

// TestCheckDocsAllExcluded: an Exclude that removes every selected doc, from the
// glob or from an explicit list, is the same no-docs error as an empty glob, so
// AssertDocs stops instead of passing with nothing checked.
func TestCheckDocsAllExcluded(t *testing.T) {
	n := testNotice(t)
	root := writeTree(t, map[string]string{
		"README.md": "# Readme\n\nNo notice.\n",
		"docs/a.md": "# A\n\nNo notice.\n",
	})
	for name, o := range map[string]Options{
		"glob":     {Exclude: []string{"README.md", "docs/a.md"}},
		"explicit": {Files: []string{"docs/a.md"}, Exclude: []string{"docs/a.md"}},
	} {
		err := CheckDocs(root, n, o)
		var none noDocsError
		if !errors.As(err, &none) || !strings.Contains(err.Error(), "nothing was checked") {
			t.Errorf("%s: err = %v, want a no-docs error saying nothing was checked", name, err)
		}
	}
	if err := CheckDocs(root, n, Options{Exclude: []string{"docs/a.md"}}); err == nil || !strings.Contains(err.Error(), "README.md") {
		t.Errorf("one doc left: err = %v, want README.md checked and reported", err)
	}
}

// TestCheckDocsExplicitFilesAndExclude: an explicit list replaces the glob;
// Exclude skips a path; an unreadable listed file keeps its cause.
func TestCheckDocsExplicitFilesAndExclude(t *testing.T) {
	n := testNotice(t)
	root := writeTree(t, map[string]string{
		"README.md":    handWritten(n),
		"docs/test.md": "# Tests\n\nNo notice by choice.\n",
		"docs/ok.md":   handWritten(n),
	})
	if err := CheckDocs(root, n, Options{Files: []string{"README.md", "docs/ok.md"}}); err != nil {
		t.Errorf("explicit files: %v", err)
	}
	if err := CheckDocs(root, n, Options{Exclude: []string{"docs/test.md"}}); err != nil {
		t.Errorf("glob with exclude: %v", err)
	}
	err := CheckDocs(root, n, Options{Files: []string{"docs/missing.md"}})
	if err == nil || !strings.Contains(err.Error(), "read docs/missing.md") || !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("missing listed file: err = %v, want a read error wrapping fs.ErrNotExist", err)
	}
}

// TestCheckDocsReportsEveryFailure: two bad docs give one joined error naming both.
func TestCheckDocsReportsEveryFailure(t *testing.T) {
	n := testNotice(t)
	root := writeTree(t, map[string]string{
		"README.md":  "# Readme\n\nNo notice.\n",
		"docs/a.md":  "# A\n\nNo notice.\n",
		"docs/ok.md": handWritten(n),
	})
	err := CheckDocs(root, n, Options{})
	if err == nil {
		t.Fatal("two bad docs passed")
	}
	for _, want := range []string{"README.md", "docs/a.md"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("err = %v, want it to name %s", err, want)
		}
	}
	if strings.Contains(err.Error(), "docs/ok.md") {
		t.Errorf("err = %v, names a doc that passed", err)
	}
}

// recorder stands in for *testing.T so the helpers' failure paths can be
// observed without failing this test. Its Fatalf records and returns; the
// helpers return right after calling it.
type recorder struct {
	testing.TB
	errs   []string
	fatals []string
}

func (r *recorder) Helper() {}

func (r *recorder) Errorf(format string, args ...any) {
	r.errs = append(r.errs, fmt.Sprintf(format, args...))
}

func (r *recorder) Fatalf(format string, args ...any) {
	r.fatals = append(r.fatals, fmt.Sprintf(format, args...))
}

// TestAssertHelpers: a pass reports nothing, each bad doc reports once, a doc
// tree with nothing to check is fatal, and Assert reports a single bad doc.
func TestAssertHelpers(t *testing.T) {
	n := testNotice(t)
	good := writeTree(t, map[string]string{"README.md": handWritten(n), "docs/a.md": handWritten(n)})
	r := &recorder{TB: t}
	AssertDocs(r, n, good, Options{Strict: true})
	if len(r.errs)+len(r.fatals) != 0 {
		t.Errorf("good tree reported %v %v", r.errs, r.fatals)
	}

	bad := writeTree(t, map[string]string{"README.md": "# R\n\nx\n", "docs/a.md": "# A\n\nx\n"})
	r = &recorder{TB: t}
	AssertDocs(r, n, bad, Options{})
	if len(r.errs) != 2 || len(r.fatals) != 0 {
		t.Errorf("bad tree: errs %v fatals %v, want two errors and no fatal", r.errs, r.fatals)
	}

	r = &recorder{TB: t}
	AssertDocs(r, n, t.TempDir(), Options{})
	if len(r.fatals) != 1 {
		t.Errorf("empty tree: fatals %v, want one", r.fatals)
	}

	r = &recorder{TB: t}
	AssertDocs(r, n, good, Options{Files: []string{"missing.md"}})
	if len(r.errs) != 1 {
		t.Errorf("single read failure: errs %v, want one", r.errs)
	}

	r = &recorder{TB: t}
	Assert(r, "doc.md", []byte(handWritten(n)), n, Options{Strict: true})
	Assert(r, "bad.md", []byte("# B\n\nx\n"), n, Options{})
	if len(r.errs) != 1 || !strings.Contains(r.errs[0], "bad.md") {
		t.Errorf("Assert: errs %v, want one naming bad.md", r.errs)
	}

	r = &recorder{TB: t}
	AssertASCII(r, "lines", []string{"plain ascii", "ok -- too"})
	if len(r.errs) != 0 {
		t.Errorf("AssertASCII on clean lines reported %v", r.errs)
	}
}

// TestAssertASCIIRejectsControlAndNonASCII: tab, DEL and a multi-byte dash are
// each reported with their offset.
func TestAssertASCIIRejectsControlAndNonASCII(t *testing.T) {
	r := &recorder{TB: t}
	AssertASCII(r, "notice", []string{"a\tb", "c\x7fd", "e\u2014f"})
	if len(r.errs) != 5 {
		t.Fatalf("errs = %d (%v), want 5: tab, DEL and the three bytes of the dash", len(r.errs), r.errs)
	}
	for i, want := range []string{"0x09 at offset 1", "0x7f at offset 1", "0xe2 at offset 1"} {
		if !strings.Contains(r.errs[i], want) {
			t.Errorf("errs[%d] = %q, want it to contain %q", i, r.errs[i], want)
		}
	}
}
