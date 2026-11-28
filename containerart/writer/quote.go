// Package writer writes the text of the renderers' artifacts: quoted scalars, a
// small YAML tree for compose and Kubernetes documents, and the lines of a quadlet
// unit. It imports only the standard library and knows no Spec, no field name of a
// tool and no sink type of workload.
//
// A writer is a value built by a factory function and keeps its state in its
// closure or its fields; no package variable holds state. It writes the values it
// is given and writes nothing to disk and runs nothing. Every writing function
// that can refuse returns an *Error, which carries a path built from keys and a
// fixed rule sentence, never a text value (S8).
//
// A tool needs two names from this package: Quote and QuoteDollar, for the text it
// still writes itself. The node and unit APIs are for the renderers.
package writer

import (
	"strconv"
	"strings"
)

// Quote renders s as a double-quoted scalar. It is strconv.Quote, which is what
// fmt's %q produces, and it leaves '$' alone.
//
// For valid UTF-8 every escape strconv.Quote writes is a YAML escape with the same
// meaning: the single-letter ones, \xNN for a control character, and \u or \U for
// any other rune that does not print. Invalid UTF-8 is the exception: a raw byte
// comes out as \xNN, which YAML reads as the code point U+00NN, a different value.
// Quote cannot refuse, so a caller checks its text first; the node writer refuses
// invalid UTF-8.
func Quote(s string) string { return strconv.Quote(s) }

// QuoteDollar renders s as a double-quoted scalar with every '$' doubled first.
// Docker Compose interpolates the whole document and turns "$$" back into one "$",
// and the kubelet does the same for a container's command, args and env, so an
// escaped value reaches the container exactly as it was written. It is the bytes
// compose.Quote writes today, and it has the same invalid UTF-8 caveat as Quote.
//
// A value that is meant to be interpolated, such as a digest label written as
// "${VAR:-}", is written by Interp and never goes through QuoteDollar.
func QuoteDollar(s string) string { return strconv.Quote(strings.ReplaceAll(s, "$", "$$")) }
