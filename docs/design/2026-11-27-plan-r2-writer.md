# Plan R2: the writer, the posture package and the R1 follow-ups

Approved by the owner, 2026-11-27. R2 is the second build step of the rendering work (design
note 2026-11-15, sections 8 to 11; consolidation draft
[2026-11-17-container-description.md](2026-11-17-container-description.md), sections 2, 3, 5,
6 and 8). It adds the one writer, the posture data and deny lists, and the checks the owner
settled after R1. It renders no artifact and changes no tool, no golden and no tag.

Executor: Sonnet agents launched by this session. This session reviews the diff. The owner
runs every gate.

## Why this step second

R1 shipped the checks. Every renderer (compose, quadlet, Kubernetes) needs the same three
things before it can be written: a way to write a quoted value for its sink, the posture lines
and deny lists it must not widen, and the last checks that waited on an answer. None of them
needs a `Spec`, so they land together, still without a renderer. R3 onward (`Spec`, compose,
quadlet, Kubernetes) then only assemble what R1 and R2 hold.

## Part 1: package `containerart/writer`

Standard library only, as the draft's import graph says. No package variable holds state; a
writer is a value built by a factory function and keeps its state in the closure (S4b). It
writes the values it is given. It knows no `Spec`, no field name of a tool and no sink type of
`workload`; the YAML dialect is its own small type. It writes nothing to disk and runs nothing.

Files and what they hold:

| File | Contents |
| --- | --- |
| `quote.go` | `Quote(s) string`, `QuoteDollar(s) string` (N25: the only two names a tool needs). `Quote` is `strconv.Quote`; `QuoteDollar` doubles every `$` first. Same bytes as today's `compose.Quote`, which is `QuoteDollar`. Both return a string and cannot refuse: for invalid UTF-8 the doc says what `compose.Quote` says today (a raw byte comes out as `\xNN`, a different code point), and the node writer below refuses it |
| `node.go` | `Node` (sealed interface), `Dialect` (`Compose`, `Kubernetes`), `Entry`, constructors `Str`, `StrDollar`, `Int`, `Bool`, `Word(list, w)`, `Interp(name)`, `Block(text)`, `Flow(items...)`, `Map(entries...)`, `Seq(items...)`, `EmptyMap()`, `KV(key, v)` (keyword key, bare), `DataKV(key, v)` (data key, quoted, S5) |
| `yaml.go` | `Marshal(root Node, d Dialect) ([]byte, error)` and `Join(docs ...[]byte) []byte` (one `---` line between documents, one final newline) |
| `unit.go` | the unit-line writer: `NewUnit() *Unit` with methods `Comment`, `Section`, `Bare`, `Quoted`, `Extra`, `Bytes() ([]byte, error)`; private `escapeQuoted`, `percent`. Line helpers stay unexported (N25); renderers in other packages use the `Node` and `Unit` APIs, and tools use only the two quotes |
| `error.go` | `Error{Path, Rule}`; `Error()` reads `path: rule`; the rule is a fixed sentence; no value is ever stored |

Rules the writer holds (draft 5.1, 5.2, S5, Q4, C10, K13):

- Layout: two spaces per level, a list under a key indented two further, block sequences only,
  flow lists only through `Flow` (compose `entrypoint`, `command`, `healthcheck.test`), LF only,
  exactly one final newline per document.
- Text values: a double-quoted scalar. `Str` on the compose dialect doubles every `$`. On the
  Kubernetes dialect `Str` does not, and `StrDollar` does (env values, `command`, `args`, probe
  commands; K13). `Interp(name)` writes `"${NAME:-}"` with the `$` single and takes a name that
  `envref` already vets; the writer re-checks only the name charset. `Int` and `Bool` are bare.
  `Word(list, w)` is bare, and refused unless `w` is in the closed list the caller passes.
- Keys (S5): `DataKV` quotes the key. Compose refuses a data key with `$`, a space or a control
  character. Kubernetes accepts a bare `KV` key only if it is plain letters and digits and is
  not a YAML 1.1 boolean or null word (`true false yes no on off y n null`, any case).
- Empties: an empty map or list is an error in a tool-built tree, except `EmptyMap()`, which
  writes `{}`. A key never holds null. An empty string is `"KEY": ""`.
- Invalid UTF-8 and the controls CheckInlineText refuses (CR, U+0085, U+2028, U+2029, other
  C0 controls but tab and LF, DEL, C1, U+FEFF) are refused in `Block` text. In `Str` a control
  character is written as a `strconv.Quote` escape, as today. The writer cannot import
  `workload` (it is stdlib only), so it holds its own copy of the block-text set and a test
  pins that it agrees with `workload.CheckInlineText`.
- Block text, compose (C10): the bytes as given, with the marker the text needs: `|-` for no
  final newline, `|` for one, `|+` for several. An indent indicator (`|2`, `|2-`, `|2+`) when the
  first non-empty line starts with a space. Empty lines carry no indent.
- Block text, Kubernetes: never `|+`. Text with two or more final newlines is written as a
  quoted string (`Quote`, with `\n` escapes). Otherwise `|` or `|-` as for compose, with the
  same indent indicator. The end of a block follows its content: no newline is added.
- Quadlet lines (Q3, Q4): `Quoted(key, value)` writes `Key="value"` with `\`, `"`, `%`, `$` escaped
  and control characters as C escapes (`\n`, `\r`, `\xNN`; tab kept), the same bytes as today's
  `quadlet.Environment`. `Bare` writes `Key=value` with `%` doubled and refuses (corrected at
  review: also `$` and every control character) a value with a
  line break. `Extra` is `Bare` plus the Q4 refusal of whitespace, a quote, a backslash or `$` in
  the value, and a key holding `=`, a space or a line break. `Section` order and the single blank
  line between sections are the writer's (Q3); line order inside a section is the renderer's.
  The Q4 fallback (`Label` bare under a no-space grammar) is one switch in the renderer, not here.

Errors (S8): `*Error` carries the path built from keys (`env["JAVA_TOOL_OPTIONS"]`, keys are
names and may be echoed) and a fixed rule. It never carries a text value or an offset into one
beyond what CheckInlineText already names. `Marshal` and `Bytes` collect every problem found
and return `errors.Join`; nothing partial is returned.

## Part 2: module-root `internal/posture`

Standard library only. At the module root, not under `containerart`, because the top-level
`artifacttest` of N37 could not import `containerart/internal/posture` (a draft inference; N25
says only "an `internal/` package"). Constants are data, not state; lists come from functions
as fresh slices.

| Name | Contents |
| --- | --- |
| `ComposeLines`, `QuadletLines` | the posture text, byte for byte the two `Posture` constants of `compose` and `quadlet` |
| `ComposeExtra(key) (reason string, denied bool)` | S1 deny list for compose service extras, with the 2026-11-17 additions and `pull_policy`: the keys M writes, host and cross-container keys, keys that pull in files or services, plus `profiles`, `scale`, `annotations`, `logging`, `oom_score_adj`, `cpu_rt_*`, `credential_spec`. Exact, lower-case match. P4 (2026-11-27): `entrypoint`, `command` and `tmpfs` leave S1's group of keys M writes, since M no longer writes them; A passes them as extras |
| `QuadletExtra(key) (reason string, denied bool)` | S1 deny list for `[Container]` extras: `PodmanArgs`, `GlobalArgs`, `ContainersConfModule`, the privilege keys, every key M writes (including `Pull`), `AutoUpdate`, `Annotation`, `CgroupsMode`, `LogDriver`, `LogOpt`, `HealthLogDestination`. Case-insensitive; a key with `=`, a space or a line break is denied with its own reason |
| `ComposeExtraKeys()`, `QuadletExtraKeys()` | the lists, for docs and tests |
| `ComposeCheckedAgainst`, `QuadletCheckedAgainst` | the engine versions the lists were checked against (S1: one test names them) |

The reason string is a fixed sentence per group ("keys M writes itself", "reaches the host or
another container", "pulls in files or services"). The key is echoed by the renderer, since a
key is a name (S8). `internal/posture` imports no `FieldError`; the renderer wraps the result.
The Kubernetes lists (pod, container, label prefixes) and the Kubernetes posture lines (K6)
wait for the Kubernetes step, which owns the Kubernetes constants.

How this relates to `artifacttest`: its `Widening()` and `HostNetwork()` are substrings searched
in rendered text, a different shape from key names in extras. They stay as they are, and are
not moved in this step. Each deny list entry carries its group, so the later N37 move can build
the widening check from the posture data. A test in `artifacttest_test.go` pins the overlap:
every `Widening()` token that is a compose or quadlet key name is denied by `posture`.

`compose.Posture` and `quadlet.Posture` become `= posture.ComposeLines` and
`= posture.QuadletLines`: one source, same bytes, same exported names, so no tool or golden
changes. Making them private is the renderer step's job (N25).

## Part 3: the R1 follow-ups (package `workload`)

All return `*FieldError`, never echo free text (S8). Edits are in `words.go` and `limits.go`.

| Name | Signature | Refuses |
| --- | --- | --- |
| `ValidMemory` (changed) | `(sink Sink, s string) *FieldError` | Kubernetes now takes the quantity grammar (E12): digits with an optional single fraction, then nothing, `k M G T P E`, `Ki Mi Gi Ti Pi Ei`, or an exponent `e<digits>`. A bare `m` (milli-bytes), `u` and `n` are refused with a message that says "write Mi or M, not m". Compose and quadlet unchanged. `ValidShmSize` still refuses Kubernetes (no field) |
| `ValidRestart` (changed) | `(sink Sink, w string) *FieldError` | Kubernetes now takes the Kubernetes spellings `Always`, `OnFailure`, `Never` as a union (the word is a member of some kind's list); lower-case `always` from R1 is refused. Empty allowed: writes nothing. Compose and quadlet unchanged |
| `ValidKubeRestart` (new) | `(kind string, w string) *FieldError` | the per-kind rule (P9): Deployment, StatefulSet, ReplicaSet, DaemonSet `Always` only; Job, CronJob `OnFailure`, `Never`; Pod all three; empty allowed. An unknown kind is refused naming the seven kinds |
| `ValidPullPolicy` (new) | `(sink Sink, w string) *FieldError` | compose: `always`, `missing`, `never`, `daily`, `weekly`, `if_not_present`, `every_<duration>` (one or more digit-and-unit pairs, units `w d h m s`); `build` refused with its own sentence ("M never builds"); quadlet: `always`, `missing`, `never`, `newer`; Kubernetes: `Always`, `IfNotPresent`, `Never`; empty allowed; exact case |
| `ValidServiceType` (new) | `(w string) *FieldError` | `ClusterIP`, `NodePort`, `LoadBalancer`, exact case; `ExternalName` refused with its own sentence; empty is an error ("M writes no default"); any other word refused, not echoed |

How the kind reaches the restart check (decided here, for the owner's approval): the kind is a
plain word the renderer passes to `ValidKubeRestart`. `ValidRestart(SinkKubernetes, w)` stays
the kind-free word check that `Spec.Check` can run when it does not know the workload kind; the
typed Deployment, StatefulSet and the other kinds call `ValidKubeRestart` with their own kind.
Reason: `Spec` has no kind, the renderer does, and a string with a closed list needs no new
type. The word lists stay in functions that return fresh slices.

## Part 4: P11 settled in this plan (for the owner's approval)

(a) Kubernetes identity (E22). Recommendation: reuse `workload.RunAs`, add no new type.
- `Spec.RunAs` (uid:gid) feeds the container block: `runAsUser` = UID, `runAsGroup` = GID.
- `kubernetes.Pod` carries a second `workload.RunAs`, built with `RunAsIDs(uid, fsGroup)`: its
  UID is the pod `runAsUser`, its GID is the pod `fsGroup`. The doc names the mismatch.
- The word `default` is the zero `RunAs` in all four fields: nothing is written. Unset is the
  tool's default, filled in before the call (A `1000:0`; B 1000001 and 1000002 on Kubernetes).
- `runAsNonRoot: true` follows a stated container uid other than 0; a stated 0 writes none.
  B's rule that a stated 0 is refused on Kubernetes stays in B (its CRD reads 0 as default).
Reasons: same grammar, same refusals and one parse for all four fields; compose and quadlet have
no pod level, so the pod pair belongs to the Kubernetes package; no new exported name. Nothing is
built in R2: it lands with `Spec` and the `kubernetes` package.

(b) `FieldError` context. Recommendation: each renderer prefixes its messages; `FieldError`
gains no sink or object member.
- A renderer wraps with `fmt.Errorf("compose: service %q: %w", name, err)`, so `errors.As` still
  reaches the `*FieldError` and the draft's 5.5 example reads as written.
- Reasons: the checks are shared with the tools at load, where there is no sink or object;
  adding members would make every check take context it does not have; `At` already renames the
  field; the writer's `*Error` follows the same rule (path and rule only).

## Tests

Test files mirror the source files. Table tests, the R1 naming style. Every refused value in
these tests is built around the `hunter2` marker, and each refusal row asserts the error text
does not contain the marker or the whole input (S8). Free-text and key refusals are checked the
same way; an echoed name or fixed-list word is pinned where echoed on purpose.

`containerart/writer/quote_test.go`: `TestQuote` (escapes, quotes, tabs, `\u` for non-printing
runes), `TestQuoteDollar` (every `$` doubled, none single), `TestQuoteMatchesComposeQuote` (a
fixed table of vectors copied from today's `compose.Quote` goldens, not an import).

`containerart/writer/node_test.go`: `TestNodeScalars` (Str, StrDollar, Int, Bool per dialect),
`TestWord` (closed list, refusal not echoed), `TestInterp` (single `$`, bad name refused),
`TestDataKeyQuoting` (S5, `123`, `true`, `NO`, `1e3`, M's own label keys), `TestKeywordKeyKubernetes`
(plain letters only; `yes`, `null`, `Y` refused), `TestComposeKeyRefusals` (`$`, space, control),
`TestEmptyNodes` (empty map, list and string; `EmptyMap`), `TestInvalidUTF8Refused`.

`containerart/writer/yaml_test.go`: `TestMarshalLayout` (indent, list under key, map in list,
flow list, one final newline, LF only), `TestMarshalIsDeterministic`, `TestJoin` (one `---`,
one final newline), `TestMarshalCollectsEveryProblem` (several bad nodes, one joined error,
nothing partial), `TestErrorPaths` (path from keys, no value).

`containerart/writer/block_test.go`: `TestBlockComposeMarkers` (`|-`, `|`, `|+`),
`TestBlockIndentIndicator` (first non-empty line starts with a space, with each marker),
`TestBlockEmptyLinesCarryNoIndent`, `TestBlockKubernetesNoKeep` (two or more final newlines
becomes a quoted string; `|` and `|-` as compose), `TestBlockRefusals` (each CheckInlineText
character, with an offset-free message), `TestBlockRefusalsMatchCheckInlineText` (an external
test that imports `workload` and walks both sets).

`containerart/writer/unit_test.go`: `TestUnitQuoted` (matches today's `quadlet.Environment` and
`quadlet.Escape` vectors), `TestUnitBare` (`%` doubled; line break refused), `TestUnitExtra`
(whitespace, quote, backslash, `$`, bad key; the key is named, the value is not), `TestUnitSections`
(order, one blank line, none after the last, `[Service]` always, header comment first),
`TestUnitCollectsEveryProblem`.

`internal/posture/posture_test.go`: `TestComposeLinesMatchRenderer` and `TestQuadletLinesMatchRenderer`
(the strings equal `compose.Posture` and `quadlet.Posture`, which now alias them),
`TestComposeExtraDenied` (every listed key, with its group), `TestComposeExtraAllowed` (keys not
on the list pass, including `shm_size` and `ulimits` of B and A's `entrypoint`, `command`
and `tmpfs` (P4); corrected at review: `cpuset` is denied, as S1 group 1 says), `TestQuadletExtraDenied`,
`TestQuadletExtraCase` (case-insensitive), `TestQuadletExtraKeyGrammar` (`=`, space, line break),
`TestExtraListsAreFresh` (a caller's edit does not change the next call), `TestCheckedAgainstNamesVersions`.

`containerart/artifacttest/artifacttest_test.go`: `TestWideningKeysAreDenied` (the overlap
pin from Part 2).

`containerart/workload/limits_test.go` (changed and new rows): `TestValidMemory` gains the
Kubernetes accept rows (`512Mi`, `2G`, `1.5Gi`, `129e6`, `4096`) and the refusal rows (`512m`,
`1m`, `5u`, `5n`, `Mi`, `1Ti5`, empty-not-allowed as today, the marker); the R1 Kubernetes
"not supported yet" row goes. `TestValidMemoryKubernetesRefusesBareM` pins the message text.
`TestValidShmSize` keeps its Kubernetes refusal.

`containerart/workload/words_test.go`: `TestValidRestart` (R1 Kubernetes rows replaced: `Always`,
`OnFailure`, `Never` pass, `always`, `no`, `on-failure` refused), `TestValidKubeRestart` (every
kind against every word, the unknown kind, empty), `TestValidPullPolicy` (every word per sink,
`every_` grammar boundaries, `build`, wrong case, invalid sink), `TestValidPullPolicyEchoes`
(`build` echoed as a fixed word; any other refused word is not), `TestValidServiceType` (the
three accepted, `ExternalName`, `clusterip`, empty, the marker).

## Coverage

Last recorded total: 99.8% (R1). Every new statement is reached by the tests above, including each
error path and each refusal. Expected: the total stays at 99.8% or rises. A branch that cannot
be reached (a defensive guard) is not written. If the total drops, the report names the number
and the cause.

## Docs in the same step

- `docs/test.md`: new sections for `containerart/writer` and `internal/posture` with every test
  and its source (`new`, or the tool vector it copies), the changed and new rows for
  `containerart/workload` and `artifacttest`, the summary line and package table totals (about
  five new test files and thirty new tests; the agent counts exactly), and the Coverage line
  after the owner reports the total.
- `docs/developer.md`: an R2 row in the build table; `writer` and `internal/posture` under the
  package layout; the "Rules for shared code" text on quoting (the writer's two exported quotes,
  invalid UTF-8) and on the deny lists; the note that `Posture` now comes from `internal/posture`.
- `docs/design/2026-11-17-container-description.md`: the status line says R2 shipped the writer,
  posture and the follow-ups; section 3 table marks `ValidPullPolicy`, `ValidServiceType`,
  `ValidMemory` and `ValidRestart` as R2; P9 and P11 move to "settled" with the owner's answer
  (after approval); a `ValidKubeRestart` row.
- Design note 2026-11-15 section 11: a pointer to this plan.
- `docs/consuming.md` and the tools' parity doc: no change.

## Not in this step

`Spec`, `Identity`, `Mount`, `Secret`, `Config`, `Label`, `Health` assembly, `CheckTargets`,
`CheckLabelKey` and the reserved prefixes, any renderer, `compose.File`, `quadlet.Unit`, the
Kubernetes package and its pod, container and label deny lists and posture lines, `outfile`
(S11, Q5), the `artifacttest` move (N37), making the renderer helpers private (N25), the
`workload` identity use of P11(a) (named above, built with `Spec`), Kubernetes object size
(K16), a Kubernetes hand check of `$` doubling (K13, before the tag).

## Effect on the tools

None in this step: no tool imports a new name, and `compose.Posture` and `quadlet.Posture` keep
their text. Later, A and B stop calling `compose.Quote`, `quadlet.Escape` and the other helpers
(they become private in the renderer step); a tool that still writes its own text calls
`writer.Quote` or `writer.QuoteDollar`. Both tools call `ValidMemory`, `ValidRestart`,
`ValidPullPolicy` and `ValidServiceType` at load, and pass the identity words per P11(a).

## Verification (the owner runs them)

1. `./scripts/dev.sh all scan`
2. `./scripts/dev.sh cov`; compare with 99.8%.

Commit headline when the gates pass: `add the writer, the posture package and the Kubernetes
memory, restart, pull policy and service type checks`
