package writer_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/hafio/go-deploy-utility/containerart/writer"
)

// layoutTree is a compose service in the shapes the writer lays out: a flow list, a
// list of words, a list of maps, a map inside a list item whose first key holds a
// list or a map, a data-key map and an empty map.
func layoutTree() writer.Node {
	return writer.Map(
		writer.KV("name", writer.Str("app")),
		writer.KV("services", writer.Map(
			writer.DataKV("app", writer.Map(
				writer.KV("image", writer.Str("registry.example/app:1.0")),
				writer.KV("cap_drop", writer.Seq(writer.Word([]string{"ALL"}, "ALL"))),
				writer.KV("entrypoint", writer.Flow(writer.Str("CMD"), writer.Str("curl"), writer.Int(5))),
				writer.KV("secrets", writer.Seq(
					writer.Map(writer.KV("source", writer.Str("s1")), writer.KV("target", writer.Str("/mnt/s1"))),
					writer.Map(writer.KV("source", writer.Str("s2")), writer.KV("target", writer.Str("/mnt/s2"))),
				)),
				writer.KV("volumes", writer.Seq(
					writer.Map(writer.KV("opts", writer.Seq(writer.Str("ro"))), writer.KV("source", writer.Str("x"))),
					writer.Map(writer.KV("a", writer.Map(writer.KV("b", writer.Str("c")))), writer.KV("d", writer.Bool(false))),
				)),
				writer.KV("ports", writer.Seq(writer.Str("8080:8080"), writer.Int(9090))),
				writer.KV("read_only", writer.Bool(true)),
				writer.KV("healthcheck", writer.Map(
					writer.KV("test", writer.Flow(writer.Str("CMD"), writer.Str("sh"))),
					writer.KV("retries", writer.Int(3)),
				)),
				writer.KV("labels", writer.Map(writer.DataKV("a.b/c", writer.Str("v")))),
				writer.KV("empty", writer.EmptyMap()),
			)),
		)),
	)
}

// TestMarshalLayout: two spaces per level, a list under a key indented two further,
// a map in a list item nested under its first key, a flow list on one line, LF only
// and one final newline.
func TestMarshalLayout(t *testing.T) {
	want := "name: \"app\"\n" +
		"services:\n" +
		"  \"app\":\n" +
		"    image: \"registry.example/app:1.0\"\n" +
		"    cap_drop:\n" +
		"      - ALL\n" +
		"    entrypoint: [\"CMD\", \"curl\", 5]\n" +
		"    secrets:\n" +
		"      - source: \"s1\"\n" +
		"        target: \"/mnt/s1\"\n" +
		"      - source: \"s2\"\n" +
		"        target: \"/mnt/s2\"\n" +
		"    volumes:\n" +
		"      - opts:\n" +
		"          - \"ro\"\n" +
		"        source: \"x\"\n" +
		"      - a:\n" +
		"          b: \"c\"\n" +
		"        d: false\n" +
		"    ports:\n" +
		"      - \"8080:8080\"\n" +
		"      - 9090\n" +
		"    read_only: true\n" +
		"    healthcheck:\n" +
		"      test: [\"CMD\", \"sh\"]\n" +
		"      retries: 3\n" +
		"    labels:\n" +
		"      \"a.b/c\": \"v\"\n" +
		"    empty: {}\n"
	got := render(t, layoutTree(), writer.Compose)
	if got != want {
		t.Errorf("compose layout =\n%s\nwant\n%s", got, want)
	}
	if strings.Contains(got, "\r") || !strings.HasSuffix(got, "\n") || strings.HasSuffix(got, "\n\n") {
		t.Errorf("compose output is not LF with exactly one final newline: %q", got)
	}

	// The same layout on Kubernetes, whose keyword keys are letters and digits only,
	// so the tree is a pod spec's shape rather than a compose service's.
	k8s := writer.Map(
		writer.KV("apiVersion", writer.Str("apps/v1")),
		writer.KV("spec", writer.Map(writer.KV("containers", writer.Seq(writer.Map(
			writer.KV("name", writer.Str("app")),
			writer.KV("args", writer.Seq(writer.StrDollar("$x"), writer.Str("$y"))),
			writer.KV("resources", writer.Map(writer.KV("limits", writer.Map(writer.DataKV("memory", writer.Str("512Mi")))))),
		))))),
	)
	wantK8s := "apiVersion: \"apps/v1\"\n" +
		"spec:\n" +
		"  containers:\n" +
		"    - name: \"app\"\n" +
		"      args:\n" +
		"        - \"$$x\"\n" +
		"        - \"$y\"\n" +
		"      resources:\n" +
		"        limits:\n" +
		"          \"memory\": \"512Mi\"\n"
	if got := render(t, k8s, writer.Kubernetes); got != wantK8s {
		t.Errorf("Kubernetes layout =\n%s\nwant\n%s", got, wantK8s)
	}
}

// TestMarshalIsDeterministic: the same tree writes the same bytes every time, and
// writing does not change the tree.
func TestMarshalIsDeterministic(t *testing.T) {
	tree := layoutTree()
	first := render(t, tree, writer.Compose)
	for i := 0; i < 5; i++ {
		if got := render(t, tree, writer.Compose); got != first {
			t.Fatalf("write %d differs from the first:\n%s\nvs\n%s", i, got, first)
		}
		if got := render(t, layoutTree(), writer.Compose); got != first {
			t.Fatalf("a rebuilt tree differs from the first:\n%s\nvs\n%s", got, first)
		}
	}
}

// TestJoin: one --- line between documents, one final newline, and nothing for an
// empty document or no document.
func TestJoin(t *testing.T) {
	for _, c := range []struct {
		name string
		docs [][]byte
		want string
	}{
		{"two", [][]byte{[]byte("a: 1\n"), []byte("b: 2\n")}, "a: 1\n---\nb: 2\n"},
		{"three", [][]byte{[]byte("a: 1\n"), []byte("b: 2\n"), []byte("c: 3\n")}, "a: 1\n---\nb: 2\n---\nc: 3\n"},
		{"one", [][]byte{[]byte("a: 1\n")}, "a: 1\n"},
		{"no final newline", [][]byte{[]byte("a: 1"), []byte("b: 2")}, "a: 1\n---\nb: 2\n"},
		{"empty document", [][]byte{[]byte("a: 1\n"), nil, []byte("b: 2\n")}, "a: 1\n---\nb: 2\n"},
		{"leading empty document", [][]byte{nil, []byte("a: 1\n")}, "a: 1\n"},
		{"none", nil, ""},
	} {
		if got := string(writer.Join(c.docs...)); got != c.want {
			t.Errorf("Join(%s) = %q, want %q", c.name, got, c.want)
		}
	}
}

// TestMarshalCollectsEveryProblem: several bad nodes give one joined error, each an
// *Error, and no partial output; no refusal echoes a value or a key.
func TestMarshalCollectsEveryProblem(t *testing.T) {
	root := writer.Map(
		writer.KV("a", writer.Str("\xff"+marker)),
		writer.KV("b", writer.Seq()),
		writer.DataKV("c$"+marker, writer.Str("v")),
		writer.KV("d", writer.Interp("1"+marker)),
		writer.KV("e", writer.Block("x\r"+marker)),
		writer.KV("f", nil),
	)
	out, err := writer.Marshal(root, writer.Compose)
	if out != nil || err == nil {
		t.Fatalf("Marshal = %q, %v; want no output and an error", out, err)
	}
	msg := err.Error()
	if strings.Contains(msg, marker) {
		t.Errorf("the joined error echoed the marker: %q", msg)
	}
	if got := strings.Count(msg, "\n") + 1; got != 6 {
		t.Errorf("the joined error holds %d problems, want 6: %q", got, msg)
	}
	for _, path := range []string{"a: ", "b: ", "[?]: ", "d: ", "e: ", "f: "} {
		if !strings.Contains(msg, path) {
			t.Errorf("the joined error lacks a problem at %q: %q", path, msg)
		}
	}
	var joined interface{ Unwrap() []error }
	if !errors.As(err, &joined) || len(joined.Unwrap()) != 6 {
		t.Errorf("the error does not unwrap into the 6 problems: %v", err)
	}
}

// TestErrorPaths: the path is built from keys, a data key as ["KEY"], a list position
// as [N], and the root as document. A key that is a name is echoed; a text value is
// never.
func TestErrorPaths(t *testing.T) {
	deep := writer.Map(writer.KV("services", writer.Map(writer.DataKV("app", writer.Map(
		writer.KV("environment", writer.Map(writer.DataKV("JAVA_TOOL_OPTIONS", writer.Str("\xff"+marker)))),
	)))))
	for _, c := range []struct {
		name string
		root writer.Node
		path string
	}{
		{"nested keys", deep, `services["app"].environment["JAVA_TOOL_OPTIONS"]`},
		{"list position", writer.Map(writer.KV("ports", writer.Seq(writer.Str("a"), writer.Str("\xff")))), "ports[1]"},
		{"item key", writer.Map(writer.KV("l", writer.Seq(writer.Map(writer.KV("k", writer.Str("\xff")))))), "l[0].k"},
		{"flow item", writer.Map(writer.KV("t", writer.Flow(writer.Str("a"), writer.Str("\xff")))), "t[1]"},
		{"refused data key", writer.Map(writer.KV("labels", writer.Map(writer.DataKV("a b", writer.Str("v"))))), "labels[?]"},
		{"refused keyword key", writer.Map(writer.KV("labels", writer.Map(writer.KV("a b", writer.Str("v"))))), "labels.?"},
		{"root", writer.Str("x"), "document"},
	} {
		_, err := writer.Marshal(c.root, writer.Compose)
		var we *writer.Error
		if !errors.As(err, &we) {
			t.Errorf("%s: %v is not a *writer.Error", c.name, err)
			continue
		}
		if we.Path != c.path {
			t.Errorf("%s: path = %q, want %q", c.name, we.Path, c.path)
		}
		if strings.Contains(err.Error(), marker) || !strings.HasPrefix(err.Error(), c.path+": ") {
			t.Errorf("%s: error %q must start with the path and hold no value", c.name, err)
		}
	}
}

// TestMarshalRootAndDialect: the dialect must be one of the two and the root a Map
// with an entry.
func TestMarshalRootAndDialect(t *testing.T) {
	ok := writer.Map(writer.KV("k", writer.Int(1)))
	for _, d := range []writer.Dialect{0, 3, -1} {
		out, err := writer.Marshal(ok, d)
		if out != nil || err == nil || !strings.Contains(err.Error(), "dialect") {
			t.Errorf("Marshal with dialect %d = %q, %v; want a dialect refusal", d, out, err)
		}
	}
	for name, root := range map[string]writer.Node{"nil": nil, "scalar": writer.Int(1), "list": writer.Seq(writer.Int(1)), "empty map": writer.EmptyMap()} {
		out, err := writer.Marshal(root, writer.Compose)
		if out != nil || err == nil || !strings.Contains(err.Error(), "Map") {
			t.Errorf("Marshal with a %s root = %q, %v; want a root refusal", name, out, err)
		}
	}
	if out, err := writer.Marshal(ok, writer.Kubernetes); err != nil || !bytes.Equal(out, []byte("k: 1\n")) {
		t.Errorf("Marshal(ok) = %q, %v; want k: 1", out, err)
	}
}
