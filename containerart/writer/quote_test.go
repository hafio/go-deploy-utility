package writer_test

import (
	"strings"
	"testing"

	"github.com/hafio/go-deploy-utility/containerart/writer"
)

// TestQuote: strconv.Quote with '$' left alone. Escapes, quotes, tabs, and \u for a
// rune that does not print; printable non-ASCII is kept.
func TestQuote(t *testing.T) {
	for _, c := range []struct{ in, want string }{
		{"", `""`},
		{"plain", `"plain"`},
		{`a"b`, `"a\"b"`},
		{`a\b`, `"a\\b"`},
		{"a\tb", `"a\tb"`},
		{"a\nb", `"a\nb"`},
		{"$x ${Y}", `"$x ${Y}"`},
		{"\x1b[31m", `"\x1b[31m"`},
		{"a\u2028b", `"a\u2028b"`},
		{"\ufeffname", `"\ufeffname"`},
		{"caf\u00e9", "\"caf\u00e9\""},
		{"\xff", `"\xff"`},
	} {
		if got := writer.Quote(c.in); got != c.want {
			t.Errorf("Quote(%q) = %s, want %s", c.in, got, c.want)
		}
	}
}

// TestQuoteDollar: every '$' is doubled, so none is left single.
func TestQuoteDollar(t *testing.T) {
	for _, c := range []struct{ in, want string }{
		{"-Xmx1g", `"-Xmx1g"`},
		{"-Dx=${Y}", `"-Dx=$${Y}"`},
		{`-Dx="a"`, `"-Dx=\"a\""`},
		{`-Dpath=C:\tmp`, `"-Dpath=C:\\tmp"`},
		{"$$", `"$$$$"`},
		{"echo $(id -u)", `"echo $$(id -u)"`},
		{"a$b${c}d", `"a$$b$${c}d"`},
	} {
		got := writer.QuoteDollar(c.in)
		if got != c.want {
			t.Errorf("QuoteDollar(%q) = %s, want %s", c.in, got, c.want)
		}
		if strings.Contains(strings.ReplaceAll(got, "$$", ""), "$") {
			t.Errorf("QuoteDollar(%q) = %s leaves a single $", c.in, got)
		}
	}
}

// TestQuoteMatchesComposeQuote: QuoteDollar writes the bytes compose.Quote writes
// today. The vectors are copied from compose_test.go (TestComposeQuoteEscapes,
// TestQuoteMatchesBothTools, TestQuoteEscapesUnprintable), not imported.
func TestQuoteMatchesComposeQuote(t *testing.T) {
	for _, c := range []struct{ in, want string }{
		{"", `""`},
		{"solace", `"solace"`},
		{"My.Broker", `"My.Broker"`},
		{"1000001:0", `"1000001:0"`},
		{"8080:8080", `"8080:8080"`},
		{"/var/lib/solace:/var/lib/solace:Z", `"/var/lib/solace:/var/lib/solace:Z"`},
		{"Europe/London", `"Europe/London"`},
		{`-Dx="a" -Dpath=C:\tmp`, `"-Dx=\"a\" -Dpath=C:\\tmp"`},
		{"a$b${c}d", `"a$$b$${c}d"`},
		{"curl -fs http://localhost:5550/health-check/readiness", `"curl -fs http://localhost:5550/health-check/readiness"`},
		{"yes", `"yes"`},
		{"off", `"off"`},
		{"0123", `"0123"`},
		{"1.5", `"1.5"`},
		{"a\nb", `"a\nb"`},
		{"a\tb", `"a\tb"`},
		{"\x1b[31m", `"\x1b[31m"`},
		{"a\xe2\x80\xa8b", `"a\u2028b"`},
		{"caf\xc3\xa9", "\"caf\xc3\xa9\""},
		{"\xef\xbb\xbfname", `"\ufeffname"`},
	} {
		if got := writer.QuoteDollar(c.in); got != c.want {
			t.Errorf("QuoteDollar(%q) = %s, compose.Quote writes %s", c.in, got, c.want)
		}
	}
}
