// Package execguard decides whether a command read from config may run.
//
// An env file is executable content. A command field names a binary this process
// runs on the operator's own machine, and env files travel -- repos, pull
// requests, shared archives -- so the person who wrote one is routinely not the
// person who runs it. Everything here exists to keep the file from choosing what
// executes.
//
// What it defends: an unlisted binary; a path form pointing at a file shipped
// beside the config; a subcommand smuggled ahead of the one the tool appends; a
// privilege-escalation wrapper in front of the tool's own commands.
//
// What it does not defend, by design: a trojan already installed on the
// operator's PATH (the host is compromised before this code runs), and config
// that is malicious but entirely legitimate in form, such as a real kubectl aimed
// at the wrong cluster. Both are review problems for the tool's trust-model docs.
//
// The whole check is one function, Check. A tool calls it from its validator AND
// from every executor immediately before argv is built, so a hostile env file is
// inert even if nobody validated first, and the two enforcement points cannot
// drift because there is only one definition to drift from.
//
// The honest limit: a flag's VALUE is an unvalidatable position. Argument arity
// is unknowable without modelling every flag of every allowed binary, so the token
// after `--kubeconfig` is accepted as that flag's value whatever it says, charset
// aside. The hard guarantee covers argv[0] and every bare token -- what runs, and
// what could act as a subcommand -- not the contents of a flag value.
//
// That limit is wider than it first reads, and the width is accepted rather than
// closed. The check cannot tell a flag that takes a value from a zero-arity
// boolean flag, so a bare word after a boolean reaches subcommand position:
//
//	kubectl --insecure-skip-tls-verify delete
//
// passes (at i=1 the token starts with `-`; at i=2 `delete` is taken for its
// value), and the real argv becomes `kubectl --insecure-skip-tls-verify delete
// apply -f -`. The container equivalent is `docker-compose --verbose down`.
// Closing it would mean requiring `--flag=value` everywhere, which breaks the
// ordinary `--context foo` form, or carrying a per-binary table of which flags
// take values, which rots as those CLIs add flags. Reaching it needs write access
// to the env file, and anyone with that can already aim an allowed binary at any
// cluster. TestFlagValuePositionIsNotGuaranteed pins it as a known limit.
//
// A flag value may also be a path, and some flags make the allowed binary load and
// run what that path holds. `docker --config ./d` makes docker run
// `./d/cli-plugins/docker-compose` for a compose command, and
// `kubectl --kubeconfig ./kc` makes kubectl run the exec credential command that
// file names, resolved beside it. So an allowed binary can still be made to run a
// file shipped beside the env file, which is the second item on the list above,
// reached through a flag value. It is accepted for the same reasons as the case
// before it: closing it needs a table of which flags load code, which rots, and
// whoever writes the env file is already trusted with what the binary reaches.
// Each tool's trust-model docs say so. TestFlagValueCanNameAFileToLoad pins it.
package execguard

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// ApproveFlag is the one way to widen an allowlist: the operator approves a
// binary for a single run on the command line. Both tools register it under this
// name. It is never a config key, because the authority to run something unusual
// belongs to the person at the keyboard, who can see what they are approving,
// and never to the config author, who may be a stranger.
const ApproveFlag = "--allow-command"

// UnsafeChars are the characters no command token may carry. Under argv exec none
// of them is an injection -- exec never involves a shell, so ';' is an ordinary
// filename character -- but a token holding one is inert only as long as it stays
// in an argv. The same tokens reach log lines, echoed output pasted into tickets,
// and rendered compose and quadlet artifacts, so they are refused at the boundary
// instead of being escaped correctly by every consumer forever.
//
// Backslash is included, which means a Windows path cannot appear in any token.
// That is not a loss for argv[0], which may not be a path at all, and a flag value
// takes forward slashes: kubectl, docker and podman all accept
// `C:/Users/you/.kube/config`. The error message says so.
//
// Tilde is deliberately not included, and the reason is positional. A leading
// tilde is refused by its own rule, because nothing downstream expands one: exec
// involves no shell, so `--kubeconfig ~/kc` would hand the CLI a directory
// literally named `~`. An embedded tilde has to be allowed, because a real
// Windows home directory is routinely an 8.3 short name (`C:/Users/RUNNER~1/...`),
// which is also what a tool's own leading-tilde expansion produces there.
//
// A tool's host-path gate may derive its own set from this one, so the charset
// keeps a single definition.
const UnsafeChars = "\"'`$;|&<>()*?[]{}#!\\"

// PathSeparators are the characters that make a token a path rather than a bare
// name. Both are rejected on every platform: an env file that ships alongside a
// `./kubectl` must not be able to point at it, and a Windows-only check would let
// the same file do exactly that when carried to Linux.
const PathSeparators = `/\`

// Rules describe what one command field may contain. A tool builds one Rules
// value per field it executes, so the per-field differences live in one place
// rather than at the call sites.
type Rules struct {
	// Field is the schema path the messages name, such as "kubernetes.command".
	Field string

	// Scope names the allowlist in the messages: "allowed on k8s are kubectl,
	// oc". Empty drops the "on k8s" part.
	Scope string

	// Allowed are the bare binary names this field may run, in the order the
	// messages list them. An escalation wrapper listed here is removed again.
	Allowed []string

	// Subword is the one bare word this field may carry, and only as its last
	// token, such as "compose" for `docker compose`. Empty allows none.
	Subword string
}

// Approvals are the operator's ApproveFlag values for one run, built by Approve.
// The zero value approves nothing. An Approvals is never changed after Approve
// returns it, so copying one is safe.
type Approvals struct {
	names map[string]bool
}

// Check is the single definition of what an executable command may be. It runs
// the layers in order: the rules themselves are usable; every token passes the
// charset; argv[0] is a bare, allowed binary name; every later token is a flag, a
// flag's value, another allowed binary, or the field's one subword in last place.
// a carries the operator's approvals for this run and is the zero value on a
// config-only path.
func Check(r Rules, cmd []string, a Approvals) error {
	if err := r.wiring(); err != nil {
		return err
	}
	if len(cmd) == 0 {
		return fmt.Errorf("%s is empty: it must name the binary to run (one of: %s)", r.Field, r.list())
	}
	for i, tok := range cmd {
		if err := checkToken(r.Field, i, tok); err != nil {
			return err
		}
	}
	ok := r.allowed(a)
	if err := checkBinary(r, cmd[0], ok); err != nil {
		return err
	}
	return checkFlagShape(r, cmd, ok)
}

// Approve records the operator's ApproveFlag values for one run and is the ONLY
// way to widen an allowlist. tool is the binary name the operator runs, which the
// escalation message names as the supported alternative.
//
// Each value gets the same charset check as a command token and must be a bare
// name, so an approval cannot itself become the path form Check refuses. A
// privilege-escalation wrapper is never approved. Any bad value fails the whole
// call and nothing is recorded.
func Approve(tool string, names []string) (Approvals, error) {
	if tool == "" {
		return Approvals{}, errors.New("execguard.Approve: tool is empty; pass the name of the binary the " +
			"operator runs, which the escalation message names")
	}
	var set map[string]bool
	for _, name := range names {
		if err := checkToken(ApproveFlag, 0, name); err != nil {
			// Re-phrased for a flag rather than a schema field: the index means
			// nothing to someone who typed a value on the command line.
			return Approvals{}, fmt.Errorf("invalid %s value: %s", ApproveFlag,
				strings.TrimPrefix(strings.TrimPrefix(err.Error(), ApproveFlag+"[0] "), "= "))
		}
		if j := strings.IndexAny(name, PathSeparators); j >= 0 {
			return Approvals{}, fmt.Errorf("invalid %s value %q: it must be a bare binary name, not a path "+
				"(found %q); approve it as %q and let PATH resolve it", ApproveFlag, name, string(name[j]), base(name))
		}
		if esc := escalator(name); esc != "" {
			return Approvals{}, fmt.Errorf("%s %s is never permitted: %s would elevate every command this tool "+
				"issues, for the whole life of an env file -- elevate the tool instead, at the moment you run it "+
				"(%s %s ...), so the privilege belongs to one invocation you chose", ApproveFlag, name, esc, esc, tool)
		}
		if set == nil {
			set = make(map[string]bool, len(names))
		}
		set[base(name)] = true
	}
	return Approvals{names: set}, nil
}

// IsControl reports the control characters no token may carry: C0, DEL and C1.
// NEL (U+0085) is left to the whitespace check, which explains it better. C1
// matters beyond tidiness: U+009B is a CSI to a terminal that honours C1, so a
// token carrying one can rewrite the log line it is printed in.
func IsControl(r rune) bool {
	return r < 0x20 || r == 0x7f || (r >= 0x80 && r <= 0x9f && r != 0x85)
}

// IsInvisible reports the Unicode format characters (category Cf): zero-width
// spaces and joiners, bidirectional overrides, the soft hyphen and their
// relatives. They split nothing under argv exec; what makes them unacceptable is
// that they are invisible. A token that renders identically to a legitimate one,
// in a review, a log line or a transcript pasted into a ticket, defeats the
// reading this package asks an operator to do. Nothing legitimate needs one.
func IsInvisible(r rune) bool { return unicode.Is(unicode.Cf, r) }

// isSpace reports the whitespace a single argument may not contain, over the
// whole Unicode White_Space property. A command written as one string is split
// with strings.Fields, which is Unicode-aware, so an ASCII-only check here would
// accept through the list form exactly what the string form splits apart. Tab
// and the vertical whitespace are caught earlier as control characters, which
// gives them a better message.
func isSpace(r rune) bool { return unicode.IsSpace(r) }

// wiring rejects Rules a tool cannot have meant, so a wiring mistake fails loudly
// on the first check rather than refusing every command with a puzzling message.
func (r Rules) wiring() error {
	if r.Field == "" {
		return errors.New("execguard: Rules.Field is empty; set it to the schema path of the command field, " +
			"such as \"kubernetes.command\"")
	}
	if len(r.Allowed) == 0 {
		return fmt.Errorf("execguard: Rules.Allowed is empty for %s; list the binaries this field may run", r.Field)
	}
	// The subword is accepted in last place before the escalator branch runs, so
	// it must never name one, and must be a plain token like any other.
	if r.Subword != "" {
		if escalator(r.Subword) != "" || strings.ContainsAny(r.Subword, PathSeparators) ||
			checkToken(r.Field, 0, r.Subword) != nil {
			return fmt.Errorf("execguard: Rules.Subword %q for %s is not a plain subcommand word; set it to the "+
				"one word the field may end with, such as \"compose\"", r.Subword, r.Field)
		}
	}
	return nil
}

// allowed builds the lookup for one check: the field's allowlist plus the
// operator's approvals. Rebuilt per call rather than cached -- the lists hold a
// handful of entries, and a cache would be one more thing that could disagree
// between validator and executor.
func (r Rules) allowed(a Approvals) map[string]bool {
	set := make(map[string]bool, len(r.Allowed)+len(a.names))
	for _, name := range r.Allowed {
		set[name] = true
	}
	for name := range a.names {
		set[name] = true
	}
	// Belt and braces on the escalation rule. Approve already refuses these with a
	// message that explains the alternative, which is where an operator meets the
	// rule; this second pass makes the outcome structural, so a Rules value that
	// lists one, or Approvals built some other way, still cannot put a
	// privilege-escalation wrapper in front of the tool's commands.
	//
	// It sweeps the SET rather than deleting known names: the two are only the
	// same while every key is lowercase, and `Sudo` or `SUDO.exe` would survive a
	// delete by name. Asking escalator about each key folds case the same way as
	// the other enforcement point, by construction.
	for name := range set {
		if escalator(name) != "" {
			delete(set, name)
		}
	}
	return set
}

// list renders the field's allowlist for a message, in declaration order, without
// the operator's approvals -- naming those back would suggest the env file could
// have asked for them.
func (r Rules) list() string { return strings.Join(r.Allowed, ", ") }

// scope renders the " on <scope>" part of the allowlist message.
func (r Rules) scope() string {
	if r.Scope == "" {
		return ""
	}
	return " on " + r.Scope
}

// checkToken is the token charset. Emptiness and control characters keep their
// own messages, since an empty argument and a converted-from-bash newline are the
// two mistakes a well-meaning env file actually makes.
func checkToken(field string, i int, tok string) error {
	if tok == "" {
		return fmt.Errorf("%s[%d] is an empty argument; remove it or quote the intended value", field, i)
	}
	if j := invalidUTF8(tok); j >= 0 {
		return fmt.Errorf("%s[%d] = %q is not valid UTF-8 (bad byte at offset %d); a command token must be "+
			"text, so re-save the env file as UTF-8", field, i, tok, j)
	}
	if j := strings.IndexFunc(tok, IsControl); j >= 0 {
		c, _ := utf8.DecodeRuneInString(tok[j:])
		return fmt.Errorf("%s[%d] contains a control character (0x%02x) at offset %d: %q", field, i, c, j, tok)
	}
	if strings.IndexFunc(tok, isSpace) >= 0 {
		return fmt.Errorf("%s[%d] = %q contains whitespace inside one argument; write each argument as its own "+
			"list entry, or use the scalar form, which splits on whitespace", field, i, tok)
	}
	if j := strings.IndexFunc(tok, IsInvisible); j >= 0 {
		c, _ := utf8.DecodeRuneInString(tok[j:])
		return fmt.Errorf("%s[%d] = %q contains an invisible formatting character (U+%04X) at offset %d; "+
			"it cannot be seen in a review or a log line, so it is not allowed in a command token", field, i, tok, c, j)
	}
	if j := strings.IndexAny(tok, UnsafeChars); j >= 0 {
		return fmt.Errorf("%s[%d] = %q contains %q, which is not allowed in a command token; "+
			"a Windows path works with forward slashes (C:/Users/you/.kube/config)",
			field, i, tok, string(tok[j]))
	}
	// A LEADING tilde, refused at every position but for two different reasons,
	// which is why the message differs. Only a leading one is refused: an embedded
	// tilde is an 8.3 short name (UnsafeChars has the argument).
	if strings.HasPrefix(tok, "~") {
		if i == 0 {
			return fmt.Errorf("%s[0] = %q starts with '~': argv[0] names the binary to run and is resolved "+
				"through the operator's own PATH, so it must be a bare name (one of the allowed binaries) "+
				"rather than a path of any kind -- a home directory cannot be part of it", field, tok)
		}
		return fmt.Errorf("%s[%d] = %q starts with '~', which nothing downstream expands: exec never involves "+
			"a shell, so this would name a directory literally called '~' -- write the path out in full",
			field, i, tok)
	}
	return nil
}

// invalidUTF8 returns the offset of the first byte that is not valid UTF-8, or
// -1. A literal U+FFFD is valid text and decodes three bytes wide; only a bad
// byte decodes as RuneError one byte wide. Without this, a raw 0x9b byte would
// decode as U+FFFD and pass every rune check, yet it is a CSI to an 8-bit
// terminal.
func invalidUTF8(s string) int {
	for i, c := range s {
		if c == utf8.RuneError {
			if _, n := utf8.DecodeRuneInString(s[i:]); n == 1 {
				return i
			}
		}
	}
	return -1
}

// checkBinary is argv[0]: a bare name from the allowlist. The bare-name rule is
// what stops `command: ./kubectl` -- a relative or absolute path would run a file
// the env file chose, which is the whole attack; a bare name can only be resolved
// through the operator's own PATH.
func checkBinary(r Rules, tok string, allowed map[string]bool) error {
	if j := strings.IndexAny(tok, PathSeparators); j >= 0 {
		return fmt.Errorf("%s[0] = %q must be a bare binary name, not a path: %q would run a file named by the "+
			"env file rather than the one on your PATH -- write it as %q and let PATH resolve it",
			r.Field, tok, string(tok[j]), base(tok))
	}
	if allowed[base(tok)] {
		return nil
	}
	if esc := escalator(tok); esc != "" {
		return neverPermitted(r.Field, 0, tok, esc)
	}
	return fmt.Errorf("%s[0] = %q is not a binary this tool runs: allowed%s are %s -- "+
		"correct the env file, or approve this one for a single run with %s %s",
		r.Field, tok, r.scope(), r.list(), ApproveFlag, base(tok))
}

// checkFlagShape is everything after argv[0]. A token passes if it is a flag,
// another allowed binary (a chained runner such as an approved `lima nerdctl`),
// the field's one subword in last place, or the value of the flag before it.
// Anything else is a bare word sitting exactly where the tool appends its own
// subcommand, so it is refused.
func checkFlagShape(r Rules, cmd []string, allowed map[string]bool) error {
	for i := 1; i < len(cmd); i++ {
		tok := cmd[i]
		switch {
		case tok == "--":
			// End-of-flags would reopen everything the bare-word rule closes: the
			// next token stops being a subcommand and starts being a positional the
			// allowed binary happily accepts.
			return fmt.Errorf("%s[%d] = \"--\" is not allowed: end-of-flags would let the env file smuggle a "+
				"positional argument ahead of the subcommand this tool appends -- remove it", r.Field, i)
		case strings.HasPrefix(tok, "-"):
		case allowed[base(tok)]:
		case r.Subword != "" && tok == r.Subword && i == len(cmd)-1:
			// The field's one permitted subcommand, and only LAST. Reaching index i
			// means every token before it passed as a flag, a flag's value or an
			// allowed binary, because this loop returns on the first that did not --
			// so "last" is the whole rule, and `docker compose up` still has a token
			// the tool did not append.
			//
			// The PRECEDING token is deliberately not constrained. A compose command
			// derived from `docker --context prod` is `docker --context prod
			// compose`, whose preceding token is the flag value `prod`; requiring an
			// allowed binary there would refuse the tool's own derived default.
		case strings.HasPrefix(cmd[i-1], "-") && !strings.Contains(cmd[i-1], "="):
			// The value of the preceding flag. Unvalidatable beyond the charset (see
			// the package doc).
		default:
			if esc := escalator(tok); esc != "" {
				return neverPermitted(r.Field, i, tok, esc)
			}
			return fmt.Errorf("%s[%d] = %q is not allowed in subcommand position: this tool appends its own "+
				"subcommand, so a bare word here would run ahead of it -- only flags, their values, and "+
				"allowed binaries (%s) may follow; approve a chained runner for a single run with %s %s",
				r.Field, i, tok, r.list(), ApproveFlag, base(tok))
		}
	}
	return nil
}

// neverPermitted is Check's message for an escalation wrapper in a position that
// would run it. It does not offer ApproveFlag, which Approve would refuse.
func neverPermitted(field string, i int, tok, esc string) error {
	return fmt.Errorf("%s[%d] = %q is a privilege-escalation wrapper (%s), which is never permitted in a "+
		"command, not even with %s: it would elevate every command this tool issues for the whole life of the "+
		"env file -- remove it, and run the tool itself under %s instead, so the privilege belongs to one "+
		"invocation you chose", field, i, tok, esc, ApproveFlag, esc)
}

// base strips one optional .exe suffix, in any case, so a Windows operator may
// write `kubectl.exe` and match the same allowlist entry. Only the suffix is
// folded -- the name itself is compared exactly, because on a case-sensitive
// filesystem `KUBECTL` and `kubectl` are different files.
func base(tok string) string {
	if len(tok) > 4 && strings.EqualFold(tok[len(tok)-4:], ".exe") {
		return tok[:len(tok)-4]
	}
	return tok
}

// escalator reports the privilege-escalation wrapper tok names, or "" for anything
// else. Each of these runs a command as another user or group, as root, in another
// namespace or root directory, or with altered capabilities. The reason is not that
// escalating is wrong -- rootful podman genuinely needs root -- but that escalating
// HERE is the wrong place for it. `sudo <tool> deploy` elevates one process the
// operator chose, visibly, at the moment they typed it. A `command: sudo podman`
// elevates every command the tool issues for the lifetime of an env file, decided
// by whoever wrote that file. The category is refused rather than the word: a list
// that blocked only `sudo` while allowing `doas` would be a control in name only.
//
// This is the ONE case-insensitive name comparison in the package (base folds
// only the .exe suffix), and the asymmetry is deliberate. The allowlist is matched exactly, because loosening a positive match
// adds binaries nobody approved. This deny list must fold, because the filesystems
// that decide whether the name resolves do: macOS is case-insensitive by default
// and Windows always is, so `Sudo` there resolves to exactly the binary `sudo`
// names, and Windows has shipped sudo.exe since 2024. Tightening a deny list can
// only refuse more, which is the safe direction to be wrong in.
func escalator(tok string) string {
	switch name := strings.ToLower(base(tok)); name {
	case "sudo", "sudoedit", "doas", "su", "pkexec", "run0", "runuser", "sg",
		"systemd-run", "machinectl", "setpriv", "capsh", "unshare", "nsenter", "chroot",
		"gosu", "su-exec", "setuidgid", "s6-setuidgid", "chpst",
		"runas", "gsudo": // the last two are Windows
		return name
	}
	return ""
}
