package writer

import (
	"errors"
	"slices"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Marshal writes the tree under root as one YAML document in the dialect d. The
// root is a Map. Layout is fixed (draft 5.1): two spaces per level, a list under a
// key indented two further, a map inside a list item nested under its first key,
// block sequences only (a flow list only through Flow), LF line ends, and every
// line, so the document, ends in a newline. A block scalar that keeps its final
// newlines (|+) ends the document with those lines, which are part of its text.
//
// Every problem found is collected and returned joined with errors.Join, each an
// *Error with a path built from keys; nothing partial is returned. The output
// depends only on the tree and the dialect.
func Marshal(root Node, d Dialect) ([]byte, error) {
	e := &emitter{d: d}
	n, _ := root.(*node)
	switch {
	case d != Compose && d != Kubernetes:
		e.fail("", "the dialect is not set: pass Compose or Kubernetes")
	case n == nil || n.kind != kMap:
		e.fail("", "the document root must be a Map")
	default:
		e.mapping(n.entries, 0, "", false)
	}
	if len(e.errs) > 0 {
		return nil, errors.Join(e.errs...)
	}
	return []byte(e.b.String()), nil
}

// Join joins documents into one stream: one "---" line between documents and a
// final newline after each. A document that is empty is left out. Join with no
// document returns nothing.
func Join(docs ...[]byte) []byte {
	var out []byte
	for _, doc := range docs {
		if len(doc) == 0 {
			continue
		}
		if len(out) > 0 {
			out = append(out, "---\n"...)
		}
		out = append(out, doc...)
		if doc[len(doc)-1] != '\n' {
			out = append(out, '\n')
		}
	}
	return out
}

// emitter holds one Marshal call: the dialect, the text so far and the problems.
type emitter struct {
	d    Dialect
	b    strings.Builder
	errs []error
}

// fail records one problem at path. The root has no path and is named document.
func (e *emitter) fail(path, rule string) {
	if path == "" {
		path = "document"
	}
	e.errs = append(e.errs, &Error{Path: path, Rule: rule})
}

// dotted joins a keyword key to its parent's path.
func dotted(path, key string) string {
	if path == "" {
		return key
	}
	return path + "." + key
}

// key returns the text of an entry's key, the path of what it holds, and whether
// the key is writable. A refused key is shown as ? in the path and never echoed.
func (e *emitter) key(en Entry, path string) (text, child string, ok bool) {
	if en.data {
		if p := dataKeyProblem(e.d, en.key); p != "" {
			child = path + "[?]"
			e.fail(child, p)
			return "", child, false
		}
		text = strconv.Quote(en.key)
		return text, path + "[" + text + "]", true
	}
	if p := keywordKeyProblem(e.d, en.key); p != "" {
		child = dotted(path, "?")
		e.fail(child, p)
		return "", child, false
	}
	return en.key, dotted(path, en.key), true
}

// mapping writes entries with their keys at column col. With dash, the first key
// follows a "- " that sits two columns left, as the first key of a list item does.
func (e *emitter) mapping(entries []Entry, col int, path string, dash bool) {
	if len(entries) == 0 {
		e.fail(path, "an empty map is not written: leave the key out, or use EmptyMap where the file must read {}")
		return
	}
	seen := make(map[string]bool, len(entries))
	for i, en := range entries {
		pad := strings.Repeat(" ", col)
		if dash && i == 0 {
			pad = strings.Repeat(" ", col-2) + "- "
		}
		text, child, ok := e.key(en, path)
		if ok {
			// Compared as written by the caller, not as quoted: a YAML reader takes
			// image and "image" for one key.
			if seen[en.key] {
				e.fail(child, "the key is written twice in one map")
			}
			seen[en.key] = true
		}
		head := pad + text
		v, _ := en.value.(*node)
		switch {
		case v == nil:
			e.fail(child, "the value is missing, so the key would hold null: give it a node or leave the key out")
		case v.kind == kMap:
			e.b.WriteString(head + ":\n")
			e.mapping(v.entries, col+2, child, false)
		case v.kind == kSeq:
			e.b.WriteString(head + ":\n")
			e.sequence(v.items, col+2, child)
		case v.kind == kBlock:
			e.block(head, v.s, col+2, child)
		default:
			if s, ok := e.inline(v, child); ok {
				e.b.WriteString(head + ": " + s + "\n")
			}
		}
	}
}

// sequence writes items as a block list with its dashes at column col.
func (e *emitter) sequence(items []Node, col int, path string) {
	if len(items) == 0 {
		e.fail(path, "an empty list is not written: leave the key out")
		return
	}
	for i, it := range items {
		p := path + "[" + strconv.Itoa(i) + "]"
		n, _ := it.(*node)
		switch {
		case n == nil:
			e.fail(p, "the item is missing, so the list would hold null")
		case n.kind == kMap:
			e.mapping(n.entries, col+2, p, true)
		case n.kind == kSeq:
			e.fail(p, "a list inside a list is not written")
		case n.kind == kBlock:
			e.fail(p, "a block text is the value of a key, not a list item")
		default:
			if s, ok := e.inline(n, p); ok {
				e.b.WriteString(strings.Repeat(" ", col) + "- " + s + "\n")
			}
		}
	}
}

// inline returns the one-line text of a scalar, an empty map or a flow list.
func (e *emitter) inline(n *node, path string) (string, bool) {
	switch n.kind {
	case kStr, kStrDollar:
		if !utf8.ValidString(n.s) {
			e.fail(path, "the text is not valid UTF-8, which cannot be quoted safely; the value is not shown")
			return "", false
		}
		if n.kind == kStrDollar || e.d == Compose {
			return QuoteDollar(n.s), true
		}
		return Quote(n.s), true
	case kInt:
		return strconv.FormatInt(n.i, 10), true
	case kBool:
		return strconv.FormatBool(n.b), true
	case kWord:
		if !slices.Contains(n.words, n.s) {
			e.fail(path, "the bare word is not in the closed list the caller passed; the value is not shown")
			return "", false
		}
		return n.s, true
	case kInterp:
		switch {
		case e.d != Compose:
			e.fail(path, "the interpolation \"${NAME:-}\" is compose's; Kubernetes has no form for it")
		case !validVarName(n.s):
			e.fail(path, "the interpolation name is not a plain variable name (a letter or '_', then letters, digits or '_'); the name is not shown")
		default:
			return `"${` + n.s + `:-}"`, true
		}
		return "", false
	case kEmptyMap:
		return "{}", true
	}
	// kFlow is the only kind left: the callers send blocks, maps and lists apart.
	if len(n.items) == 0 {
		e.fail(path, "an empty list is not written: leave the key out")
		return "", false
	}
	parts := make([]string, 0, len(n.items))
	good := true
	for i, it := range n.items {
		p := path + "[" + strconv.Itoa(i) + "]"
		c, _ := it.(*node)
		if c == nil || !c.scalar() {
			e.fail(p, "a flow list holds scalar items only")
			good = false
			continue
		}
		s, ok := e.inline(c, p)
		good = good && ok
		parts = append(parts, s)
	}
	return "[" + strings.Join(parts, ", ") + "]", good
}

// block writes text as a block scalar under head, with its lines at column col.
func (e *emitter) block(head, text string, col int, path string) {
	if blockProblem(text) {
		e.fail(path, "the text holds invalid UTF-8, a line break other than newline or a control character; "+
			"the text is not shown, remove it")
		return
	}
	if e.d == Compose {
		text = strings.ReplaceAll(text, "$", "$$")
	}
	trail := len(text) - len(strings.TrimRight(text, "\n"))
	keep := trail >= 2 || (trail > 0 && trail == len(text))
	if keep && e.d == Kubernetes {
		e.b.WriteString(head + ": " + Quote(text) + "\n")
		return
	}
	var lines []string
	if text != "" {
		// Each element is one line of the file, so a text that ends in newlines
		// ends in empty lines, which a |+ block keeps.
		lines = strings.Split(strings.TrimSuffix(text, "\n"), "\n")
	}
	indicator := ""
	for _, l := range lines {
		if l != "" {
			if l[0] == ' ' {
				indicator = "2"
			}
			break
		}
	}
	chomp := "-"
	switch {
	case keep:
		chomp = "+"
	case trail == 1:
		chomp = ""
	}
	e.b.WriteString(head + ": |" + indicator + chomp + "\n")
	pad := strings.Repeat(" ", col)
	for _, l := range lines {
		if l != "" {
			e.b.WriteString(pad + l)
		}
		e.b.WriteString("\n")
	}
}

// blockProblem reports whether text holds what workload.CheckInlineText refuses.
// The writer imports only the standard library, so it holds its own copy of the set
// and a test pins that the two agree.
func blockProblem(text string) bool {
	for i := 0; i < len(text); {
		r, n := utf8.DecodeRuneInString(text[i:])
		if r == utf8.RuneError && n == 1 {
			return true
		}
		switch {
		case r == '\n' || r == '\t':
		case r == '\r' || r == 0x7f || r == 0x85 || r == 0x2028 || r == 0x2029 || r == 0xfeff:
			return true
		case r < 0x20 || (r >= 0x80 && r <= 0x9f):
			return true
		}
		i += n
	}
	return false
}

// validVarName reports whether name is a plain variable name: a letter or '_', then
// letters, digits or '_', ASCII only.
func validVarName(name string) bool {
	if name == "" {
		return false
	}
	for i := 0; i < len(name); i++ {
		c := name[i]
		letter := c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c == '_'
		if !letter && (i == 0 || c < '0' || c > '9') {
			return false
		}
	}
	return true
}

// keywordKeyProblem says why key cannot be written bare as a keyword, or "".
func keywordKeyProblem(d Dialect, key string) string {
	for i := 0; i < len(key); i++ {
		c := key[i]
		letter := c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
		digit := c >= '0' && c <= '9'
		extra := d == Compose && (c == '_' || c == '-' || c == '.')
		if !letter && !(i > 0 && (digit || extra)) {
			return "a bare keyword key must start with a letter and hold only letters and digits " +
				"(on compose also '_', '-' and '.'); write a key that is data with DataKV; the key is not shown"
		}
	}
	switch strings.ToLower(key) {
	case "", "true", "false", "yes", "no", "on", "off", "y", "n", "null":
		return "a bare key must not be empty or a YAML 1.1 boolean or null word, which a reader would not take as text; " +
			"write a key that is data with DataKV"
	}
	return ""
}

// dataKeyProblem says why key cannot be written as a data key in dialect d, or "".
// The text never holds the key.
func dataKeyProblem(d Dialect, key string) string {
	if key == "" {
		return "a data key is empty"
	}
	if !utf8.ValidString(key) {
		return "a data key is not valid UTF-8; the key is not shown"
	}
	for _, r := range key {
		switch {
		case unicode.IsControl(r):
			return "a data key holds a control character; the key is not shown"
		case d == Compose && r == '$':
			return "a data key holds '$', which compose would interpolate; the key is not shown"
		case d == Compose && r == ' ':
			return "a data key holds a space; the key is not shown"
		}
	}
	return ""
}
