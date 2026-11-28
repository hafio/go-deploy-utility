package writer

// Error is one refused node or line. Path says where, built from keys, such as
// `env["JAVA_TOOL_OPTIONS"]` or `Container.Image`, and Rule is a fixed sentence
// chosen in this package. Neither holds a text value (S8): a key is echoed only
// when it is a valid name, and a refused key is shown as ?. Marshal and Bytes
// return every Error they find joined with errors.Join, so errors.As reaches each.
type Error struct {
	// Path locates the refused node: keys from the document root, and list
	// positions as [N]. "document" names the root itself.
	Path string

	// Rule says what failed and what to write instead.
	Rule string
}

// Error reads `path: rule`.
func (e *Error) Error() string { return e.Path + ": " + e.Rule }
