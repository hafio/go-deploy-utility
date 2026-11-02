// Package confirm asks the operator before a destructive command runs.
//
// The rule is the one decided for both tools: a destructive command whose stdin
// is not a terminal, run without --no-prompt, fails at once with an error that
// names the flag, and piped answers are never read. A script or CI job therefore
// says what it wants on its command line rather than through a pipe that might be
// stale.
//
// "Is a terminal" means stdin is a character device (StdinIsTerminal). That test
// cannot tell a real terminal from /dev/null, so `cmd </dev/null` and cron runs
// are asked, read end of file at once, and fail with an error that also names
// --no-prompt, with "reading input: EOF" as its cause.
// Tests must therefore always inject Options.IsTerminal: `go test` itself runs
// with stdin on /dev/null.
//
// Keep one Prompter per run. It holds the run's single buffered reader, so two
// questions answered by one write ("y\nyes\n") both get their answer.
package confirm
