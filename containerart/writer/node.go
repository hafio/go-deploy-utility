package writer

// Dialect names the YAML dialect a tree is written in. The zero value is invalid,
// so a caller that did not choose one is refused.
type Dialect int

const (
	// Compose is the docker compose file: Str and Block double '$', and keys
	// are held to compose's grammar.
	Compose Dialect = iota + 1

	// Kubernetes is a Kubernetes manifest: Str leaves '$' alone and StrDollar
	// doubles it, and block text never uses |+.
	Kubernetes
)

// kind tells the node types apart. The scalar kinds come first, so a flow list can
// ask whether an item is one.
type kind int

const (
	kStr kind = iota
	kStrDollar
	kInt
	kBool
	kWord
	kInterp
	kEmptyMap
	kFlow
	kBlock
	kMap
	kSeq
)

// Node is a value in a tree: a scalar, a block of text, a flow list, a map or a
// list. The interface is sealed: build nodes with the constructors of this package.
type Node interface{ isNode() }

// node is the one concrete Node. Which fields are read depends on kind.
type node struct {
	kind    kind
	s       string
	i       int64
	b       bool
	words   []string
	items   []Node
	entries []Entry
}

func (*node) isNode() {}

// scalar reports whether the node is written on one line as a plain scalar.
func (n *node) scalar() bool { return n.kind <= kInterp }

// Entry is one key and its value in a Map. Build one with KV or DataKV.
type Entry struct {
	key   string
	data  bool
	value Node
}

// KV is an entry whose key is a keyword of the file, such as image or apiVersion,
// written bare. A keyword key is plain: on compose letters, digits, '_', '-' and
// '.', on Kubernetes letters and digits, starting with a letter on both, and never
// one of the YAML 1.1 boolean or null words (true, false, yes, no, on, off, y, n,
// null in any case), which a reader would not take as text. A key that is data
// takes DataKV.
func KV(key string, v Node) Entry { return Entry{key: key, value: v} }

// DataKV is an entry whose key is data, such as an environment name, a label key or
// the key of a ConfigMap (S5). It is always written quoted, so `123`, `true`, `NO`
// and `1e3` stay text. Compose also refuses a data key that holds '$', a space or a
// control character. Both dialects refuse an empty key, invalid UTF-8 and a control
// character.
func DataKV(key string, v Node) Entry { return Entry{key: key, data: true, value: v} }

// Str is a double-quoted text value. On the Compose dialect every '$' is doubled,
// since compose interpolates the whole document; on Kubernetes it is not (see
// StrDollar). A control character is written as a strconv.Quote escape. Invalid
// UTF-8 is refused when the tree is written.
func Str(s string) Node { return &node{kind: kStr, s: s} }

// StrDollar is a double-quoted text value with every '$' doubled on both dialects.
// Kubernetes needs it for env values, command, args and probe commands, where the
// kubelet expands $(NAME) and reads $$ as one $ (K13).
func StrDollar(s string) Node { return &node{kind: kStrDollar, s: s} }

// Int is a bare integer.
func Int(n int64) Node { return &node{kind: kInt, i: n} }

// Bool is a bare true or false.
func Bool(b bool) Node { return &node{kind: kBool, b: b} }

// Word is a bare word the caller chose, such as ALL in cap_drop. It is refused when
// w is not in list, the closed list the caller passes, so a word that came from
// operator input cannot be written bare. The word is not echoed.
func Word(list []string, w string) Node { return &node{kind: kWord, s: w, words: list} }

// Interp is the interpolation "${NAME:-}" with the '$' single, for the two values M
// asks compose to fill in itself: a label that reads its value from the
// environment, and the deploy-digest label. name is checked against the variable
// name charset (a letter or '_', then letters, digits or '_', ASCII), which is all
// this package re-checks of what envref already vets. It is a compose form and is
// refused on Kubernetes. The name is not echoed.
func Interp(name string) Node { return &node{kind: kInterp, s: name} }

// Block is multi-line text written as a block scalar. The text must pass the
// inline-text rule of workload.CheckInlineText, which this package re-states: no
// invalid UTF-8, CR, NEL, LS, PS, other control character but tab and newline, DEL,
// C1 character or BOM. Compose writes the bytes as given (C10) under the marker the
// text needs: |- for no final newline, | for one, |+ for several; and the text has
// every '$' doubled, since compose interpolates inline content too. Kubernetes
// never writes |+: text that needs it is written as a quoted string instead (P8),
// and '$' is left alone. A first non-empty line that starts with a space adds an
// indent indicator, as in |2-. Empty lines carry no indent. A Block is the value of
// a key and is not a list item.
func Block(text string) Node { return &node{kind: kBlock, s: text} }

// Flow is a one-line list of scalars, written [a, b]. It is the shape of compose
// entrypoint, command and healthcheck.test. Every item is a scalar node: Str,
// StrDollar, Int, Bool, Word or Interp. A flow list with no item is refused.
func Flow(items ...Node) Node { return &node{kind: kFlow, items: items} }

// Map is a block map of entries, written in the order given. A map with no entry
// is refused when the tree is written: a tool leaves the key out, or uses EmptyMap
// where the file must read {}.
func Map(entries ...Entry) Node { return &node{kind: kMap, entries: entries} }

// Seq is a block list, written one "- " item per element, indented two under its
// key. A list with no item is refused, and so is a list inside a list.
func Seq(items ...Node) Node { return &node{kind: kSeq, items: items} }

// EmptyMap is the one empty node a tree may hold: it is written {}.
func EmptyMap() Node { return &node{kind: kEmptyMap} }
