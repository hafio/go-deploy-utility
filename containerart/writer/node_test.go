package writer_test

import (
	"testing"

	"github.com/hafio/go-deploy-utility/containerart/writer"
)

// render writes root in dialect d and fails the test on a refusal.
func render(t *testing.T, root writer.Node, d writer.Dialect) string {
	t.Helper()
	out, err := writer.Marshal(root, d)
	if err != nil {
		t.Fatalf("Marshal refused: %v", err)
	}
	return string(out)
}

// field wraps one node as the value of the key k, the shape most tests write.
func field(v writer.Node) writer.Node { return writer.Map(writer.KV("k", v)) }

// TestNodeScalars: Str, StrDollar, Int and Bool per dialect. Str doubles '$' on
// compose and not on Kubernetes; StrDollar doubles on both; an empty string is
// written, not left out; a control character is a quote escape.
func TestNodeScalars(t *testing.T) {
	for _, c := range []struct {
		name         string
		node         writer.Node
		compose, k8s string
	}{
		{"Str", writer.Str("a$b"), "k: \"a$$b\"\n", "k: \"a$b\"\n"},
		{"StrDollar", writer.StrDollar("a$b ${C}"), "k: \"a$$b $${C}\"\n", "k: \"a$$b $${C}\"\n"},
		{"Str empty", writer.Str(""), "k: \"\"\n", "k: \"\"\n"},
		{"Str control", writer.Str("a\x01b\n"), "k: \"a\\x01b\\n\"\n", "k: \"a\\x01b\\n\"\n"},
		{"Str quotes", writer.Str(`a"b\c`), "k: \"a\\\"b\\\\c\"\n", "k: \"a\\\"b\\\\c\"\n"},
		{"Int", writer.Int(8080), "k: 8080\n", "k: 8080\n"},
		{"Int negative", writer.Int(-1), "k: -1\n", "k: -1\n"},
		{"Bool true", writer.Bool(true), "k: true\n", "k: true\n"},
		{"Bool false", writer.Bool(false), "k: false\n", "k: false\n"},
	} {
		if got := render(t, field(c.node), writer.Compose); got != c.compose {
			t.Errorf("%s on compose = %q, want %q", c.name, got, c.compose)
		}
		if got := render(t, field(c.node), writer.Kubernetes); got != c.k8s {
			t.Errorf("%s on Kubernetes = %q, want %q", c.name, got, c.k8s)
		}
	}
}

// TestWord: a bare word is written only when it is in the closed list the caller
// passes, and a refused word is not echoed.
func TestWord(t *testing.T) {
	list := []string{"ALL", "no"}
	for _, w := range []string{"ALL", "no"} {
		if got, want := render(t, field(writer.Word(list, w)), writer.Compose), "k: "+w+"\n"; got != want {
			t.Errorf("Word(%q) = %q, want %q", w, got, want)
		}
	}
	for _, w := range []string{"", "all", "ALL " + marker, marker, "yes"} {
		_, err := writer.Marshal(field(writer.Word(list, w)), writer.Compose)
		expectRefusal(t, "Word", w, err)
	}
	_, err := writer.Marshal(field(writer.Word(nil, "ALL")), writer.Kubernetes)
	expectRefusal(t, "Word with no list", "ALL", err)
}

// TestInterp: the '$' stays single in "${NAME:-}", the name is held to the variable
// charset and not echoed, and Kubernetes has no such form.
func TestInterp(t *testing.T) {
	for _, name := range []string{"DEPLOY_DIGEST", "_x1", "a"} {
		if got, want := render(t, field(writer.Interp(name)), writer.Compose), "k: \"${"+name+":-}\"\n"; got != want {
			t.Errorf("Interp(%q) = %q, want %q", name, got, want)
		}
	}
	for _, name := range []string{"", "1ABC", "A-B", "A B", "A:-" + marker, "${" + marker + "}", marker + "\n", "caf\u00e9" + marker} {
		_, err := writer.Marshal(field(writer.Interp(name)), writer.Compose)
		expectRefusal(t, "Interp", name, err)
	}
	_, err := writer.Marshal(field(writer.Interp("DEPLOY_DIGEST")), writer.Kubernetes)
	expectRefusal(t, "Interp on Kubernetes", "DEPLOY_DIGEST", err)
}

// TestDataKeyQuoting: S5. A key that is data is always quoted, so a reader cannot
// take it for a number or a boolean; M's own label keys are quoted too.
func TestDataKeyQuoting(t *testing.T) {
	for _, key := range []string{
		"123", "true", "NO", "1e3", "TZ", "JAVA_TOOL_OPTIONS", "deploy.solace.com/instance",
		"app.kubernetes.io/managed-by", "services.solace.com/solution", "a\"b", "caf\u00e9",
	} {
		want := writer.Quote(key) + ": \"v\"\n"
		for _, d := range []writer.Dialect{writer.Compose, writer.Kubernetes} {
			root := writer.Map(writer.DataKV(key, writer.Str("v")))
			if got := render(t, root, d); got != want {
				t.Errorf("DataKV(%q) in dialect %d = %q, want %q", key, d, got, want)
			}
		}
	}
	// Kubernetes data keys may hold a space or '$': the file names of a ConfigMap
	// are not interpolated.
	for _, key := range []string{"a b", "a$b"} {
		root := writer.Map(writer.DataKV(key, writer.Str("v")))
		if got, want := render(t, root, writer.Kubernetes), writer.Quote(key)+": \"v\"\n"; got != want {
			t.Errorf("DataKV(%q) on Kubernetes = %q, want %q", key, got, want)
		}
	}
}

// TestKeywordKeyKubernetes: a bare key on Kubernetes is plain letters and digits,
// starts with a letter, and is not a YAML 1.1 boolean or null word.
func TestKeywordKeyKubernetes(t *testing.T) {
	for _, key := range []string{"apiVersion", "kind", "readOnlyRootFilesystem", "runAsUser", "x1", "Yes1", "nope", "ony"} {
		if got, want := render(t, writer.Map(writer.KV(key, writer.Str("v"))), writer.Kubernetes), key+": \"v\"\n"; got != want {
			t.Errorf("KV(%q) = %q, want %q", key, got, want)
		}
	}
	for _, key := range []string{
		"yes", "no", "on", "off", "y", "n", "null", "true", "false", "Y", "N", "NULL", "True", "OFF", "123", "1e3",
		"a-b", "a.b", "a_b", "a b", "", "a:b", "\xff", marker + "-x", "x " + marker, "caf\u00e9" + marker,
	} {
		_, err := writer.Marshal(writer.Map(writer.KV(key, writer.Str("v"))), writer.Kubernetes)
		expectRefusal(t, "KV on Kubernetes", key, err)
	}
}

// TestKeywordKeyCompose: a bare key on compose is a letter, then letters, digits,
// '_', '-' or '.', and is not a boolean or null word.
func TestKeywordKeyCompose(t *testing.T) {
	for _, key := range []string{"image", "container_name", "mem_limit", "x-extension", "x.y", "a1", "Yes1", "no-new-privileges"} {
		if got, want := render(t, writer.Map(writer.KV(key, writer.Int(1))), writer.Compose), key+": 1\n"; got != want {
			t.Errorf("KV(%q) = %q, want %q", key, got, want)
		}
	}
	for _, key := range []string{"yes", "off", "NULL", "n", "", "1a", "_a", "-a", "a b", "a$b", "a:b", marker + "$", "x " + marker} {
		_, err := writer.Marshal(writer.Map(writer.KV(key, writer.Int(1))), writer.Compose)
		expectRefusal(t, "KV on compose", key, err)
	}
}

// TestComposeKeyRefusals: S5. A data key is refused when it is empty, invalid
// UTF-8 or holds a control character, on both dialects; compose also refuses '$'
// and a space. No refusal echoes the key.
func TestComposeKeyRefusals(t *testing.T) {
	both := []string{"", "tab\t" + marker, "a\nb " + marker, "\x7f" + marker, "\u0085" + marker, "\x00" + marker, "\xff" + marker}
	composeOnly := []string{"a$b" + marker, "$" + marker, "a b" + marker, " " + marker}
	for _, key := range both {
		for _, d := range []writer.Dialect{writer.Compose, writer.Kubernetes} {
			_, err := writer.Marshal(writer.Map(writer.DataKV(key, writer.Str("v"))), d)
			expectRefusal(t, "DataKV", key, err)
		}
	}
	for _, key := range composeOnly {
		_, err := writer.Marshal(writer.Map(writer.DataKV(key, writer.Str("v"))), writer.Compose)
		expectRefusal(t, "DataKV on compose", key, err)
	}
}

// TestEmptyNodes: an empty map or list in a tool-built tree is an error, and a key
// never holds null; EmptyMap writes {} and an empty string is written as "".
func TestEmptyNodes(t *testing.T) {
	for _, c := range []struct {
		name string
		node writer.Node
	}{
		{"Map", writer.Map()},
		{"Seq", writer.Seq()},
		{"Flow", writer.Flow()},
		{"nil", nil},
		{"nil in Seq", writer.Seq(nil)},
	} {
		_, err := writer.Marshal(field(c.node), writer.Compose)
		expectRefusal(t, "empty "+c.name, c.name, err)
	}
	_, err := writer.Marshal(writer.Map(), writer.Compose)
	expectRefusal(t, "empty root", "root", err)
	if got, want := render(t, field(writer.EmptyMap()), writer.Compose), "k: {}\n"; got != want {
		t.Errorf("EmptyMap() = %q, want %q", got, want)
	}
	if got, want := render(t, writer.Map(writer.DataKV("KEY", writer.Str(""))), writer.Compose), "\"KEY\": \"\"\n"; got != want {
		t.Errorf("an empty environment value = %q, want %q", got, want)
	}
	if got, want := render(t, writer.Map(writer.KV("l", writer.Seq(writer.EmptyMap()))), writer.Compose), "l:\n  - {}\n"; got != want {
		t.Errorf("EmptyMap in a list = %q, want %q", got, want)
	}
}

// TestMisplacedNodesRefused: a list in a list, a block as a list item, a flow list
// holding a non-scalar, and a key written twice are refused with a path.
func TestMisplacedNodesRefused(t *testing.T) {
	for _, c := range []struct {
		name string
		root writer.Node
	}{
		{"list in list", field(writer.Seq(writer.Seq(writer.Str("a"))))},
		{"block in list", field(writer.Seq(writer.Block("a\n")))},
		{"map in flow", field(writer.Flow(writer.Map(writer.KV("a", writer.Int(1)))))},
		{"nil in flow", field(writer.Flow(nil))},
		{"flow in flow", field(writer.Flow(writer.Flow(writer.Int(1))))},
		{"bad item in flow", field(writer.Flow(writer.Str("a"), writer.Word(nil, marker)))},
		{"duplicate key", writer.Map(writer.KV("k", writer.Int(1)), writer.KV("k", writer.Int(2)))},
		{"duplicate in item", field(writer.Seq(writer.Map(writer.DataKV("A", writer.Int(1)), writer.DataKV("A", writer.Int(2)))))},
	} {
		_, err := writer.Marshal(c.root, writer.Compose)
		expectRefusal(t, c.name, c.name, err)
	}
	// A YAML reader takes image and "image" for one key, so the pair is refused too.
	root := writer.Map(writer.KV("image", writer.Int(1)), writer.DataKV("image", writer.Int(2)))
	_, err := writer.Marshal(root, writer.Compose)
	expectRefusal(t, "keyword and data key", "image", err)
}

// TestInvalidUTF8Refused: text that is not valid UTF-8 cannot be quoted safely, so
// every way to write it is refused on both dialects without echoing it.
func TestInvalidUTF8Refused(t *testing.T) {
	bad := "\xff" + marker
	for _, c := range []struct {
		name string
		node writer.Node
	}{
		{"Str", writer.Str(bad)},
		{"StrDollar", writer.StrDollar(bad)},
		{"Flow", writer.Flow(writer.Str("ok"), writer.Str(bad))},
		{"Block", writer.Block(bad)},
	} {
		for _, d := range []writer.Dialect{writer.Compose, writer.Kubernetes} {
			_, err := writer.Marshal(field(c.node), d)
			expectRefusal(t, c.name, bad, err)
		}
	}
}
