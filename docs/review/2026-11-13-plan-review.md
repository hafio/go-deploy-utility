# Plan review, 2026-11-13

Reviewed: `docs/todo.md` in go-deploy-utility (M, items 1-18, "gdu N"),
`docs/review/2026-11-13-decisions.md` in solmq-conn-util (A) and in solace-util (B).
Nothing was built, run or edited. Nine read-only Sonnet checkers compared the plans
with the three repos, the init-devops contract and upstream documentation; the lines
behind every point in sections 1 to 3 were opened again afterwards. The checkers'
full output is in `2026-11-13-plan-review-findings.md` beside this file.

Counts: 231 factual statements checked: 165 hold, 48 hold in part, 10 are wrong,
8 could not be checked. 182 problems raised (33 high, 92 medium, 57 low), with
overlap between checkers.

## Verdict

The direction is sound and most "Today" statements are right. The plans are not
ready to build from. Three things stand in the way:

1. The label work (gdu 1, 10, 16, 18) contradicts itself and existing behaviour in
   several places, and its migration cost is not written down.
2. Work that one tool's review created for the other tool is in neither tool's file.
3. This repository's own 25 decisions are in no file at all.

## 1. Facts this review settled

These were marked suspected, unverified or "not read" in the plans.

| # | Subject | Result | Source |
|---|---|---|---|
| F1 | M CR-0005 | Confirmed. Podman reports a container's secrets under `Config.Secrets`; there is no top-level key. Each entry has Name and ID. | libpod `define/container_inspect.go` |
| F2 | M CR-0006 | Confirmed. Quadlet copies `Environment=`, `Label=`, `Volume=` and `Image=` values into `ExecStart` and escapes neither `$` nor `%`. systemd then expands `${NAME}` (unknown names become empty) and `%` specifiers. A literal dollar is `$$`. | quadlet source, `systemd.service` |
| F3 | gdu 17 | Very likely true. The operator sets `OnDelete` on its StatefulSets on every update pass, whatever the CR's `updateStrategy` says, and `kubectl rollout status` then exits 1 with "rollout status is only available for RollingUpdate strategy type". | operator `controllers/statefulset.go`, kubectl `rollout_status.go` |
| F4 | B certificate renewal | The operator hashes the TLS Secret into a pod-template annotation and rolls the pods under `automatedRolling`. Under `manualPodRestart` it does not. B's own comment says the opposite (B `internal/cli/ops_k8s.go:116-122`), and B bundles operator 1.4.2. | operator user guide and controller source, main branch |
| F5 | Labels on the CR | Not copied to the StatefulSets, Services, Secrets or PVCs. The operator labels its children itself: `app.kubernetes.io/instance=<CR name>`, `name=pubsubpluseventbroker`, `managed-by=solace-pubsubplus-operator`. | operator `controllers/namings.go` |
| F6 | podman secrets | `--label` since 4.3. `--replace` since 4.7; it assigns a new secret ID and sets the labels to exactly those passed, so a replace without labels clears them. `secret ls` filters by name and id only. `secret inspect` takes several names; a missing one gives partial output and a non-zero exit. | podman release notes, `containers/common` secrets source |
| F7 | kubectl | `delete` waits 168h without `--timeout`. `get t1,t2` with one refused type prints the rest and exits 1. `api-resources` prints what it found and exits non-zero when an aggregated API is down. `all` is 10 types. `apply --prune` is still alpha. | kubectl source |
| F8 | compose | `down` on a project it cannot find exits 0 with a warning on stderr. `COMPOSE_PROJECT_NAME` beats the file's `name:`; only `-p` beats the variable. Labels on a top-level secret or config reach no engine object. | compose `down.go`, compose docs and schema |
| F9 | systemd | `stop` on a unit that is not loaded exits 5. `is-active` exits 4 for a unit that is not found and 3 for an inactive one. | systemctl source |
| F10 | "A was not read", gdu 10 | A has the same defect. Both tools can fail after the unit and secrets are written and before the restart, and the next deploy reports "unchanged". | A `cmd/solmq-conn-util/main.go:955-998`, `internal/runner/runner.go:733-775` |
| F11 | "A was not read", gdu 11 | A never reads, prints or pins a context. Its one destructive prompt names no cluster, so the wrong-name half (B CR-0082) is absent and the no-pin half (B CR-0026) is present. | A `internal/runner/runner.go`, `main.go:1642-1675` |
| F12 | "A was not read", gdu 9 input | A's docker remove has the false success (prints "ok" after a down that found nothing) and no data loss, since A purges nothing. A's podman remove deletes the secrets even when the stop failed. | A `main.go:716-791`, `main.go:1038-1054` |

Consequences:

- M CR-0005 and CR-0006 no longer wait on a host. Both can be fixed now; the host run
  becomes a smoke row.
- The quadlet line builder of gdu 4 must double `$` as well as `%`.
- gdu 17 should be planned as real work. The cluster check stays, on `-p`, `-b` and
  `-m`, for both update strategies.
- B's decision to drop the certificate restart holds only for `automatedRolling`, and
  only if operator 1.4.2 already behaves as main does.

## 2. What to fix in the plans before building

### 2.1 Across the three plans

- **P1. M has no decisions file.** The 25 decisions on M's own review exist only in a
  conversation. Write `docs/review/2026-11-13-decisions.md` in M in the format A and B
  use. Links to carry: CR-0003 -> gdu 1; CR-0004 and CR-0023 -> gdu 4; CR-0005 -> gdu
  10 (see P7); CR-0006 -> gdu 2 and gdu 4; CR-0009 -> the smoke rows gdu 9, 16 and 17
  ask for; CR-0017 clause 4 and CR-0020 -> gdu 15; CR-0025 statement 2 -> gdu 4.
  Hold back the removal of `runnertest.FailOn` (CR-0001): B CR-0047 wants error-path
  tests that prove no later step ran, which is what it is for.
- **P2. Cross-tool work is missing from both files.** A's file says the todo "holds
  nine items" and has no row for gdu 10, 11, 12, 14 or 15, each of which says both
  tools adopt in one step. A needs: the digest label and forced reload (10), a context
  line and pin (11), bounded deletes (12), a decision on `--tail all` (14), `--local`
  (15). B's file lists no gdu 5, 6, 7 or 8. gdu 6 is the sharp one: a shared check that
  requires a stated network mode fails B's tests the moment it lands.
- **P3. "Questions first, actions after" is written as absolute and is not.** Both
  tools ask the namespace question after the deletes, by design, and parity rule 9
  says so (A `main.go:700-711`, `namespace.go:129`; B `ops_k8s.go:727-812`). B also
  asks its certificate restart question after writing. Either name the namespace offer
  as the one exception in the new parity rule, or move the question before the deletes
  (list first with the owned set discounted, verify again after). Decide before the
  parity row is written.
- **P4. A parity rule can only enter "Contract" when both tools follow it**, because
  that section says both do today and `paritydoc.Check` ties each tool to the body.
  Thirteen of the 18 items add or change a rule. Stage a rule as an open row under
  Divergences until the tag in which both tools comply.

### 2.2 Labels: gdu 1, 16, 18

- **P5. "A labelled object is owned, anything else is an occupant" is unsafe.** It
  must mean this tool and this instance: two instances can share a namespace. The
  operator's children carry no tool label (F5), so a label-only rule makes every one
  an occupant and B's namespace is never offered. Keep the by-name roots and the
  ownerReference chain, and add the label as one more discount. The CR kind and the
  operator's PodDisruptionBudget must be in the wider listing.
- **P6. The first deploy after adoption is not silent.** A new `Label=` line changes
  the unit, and a changed unit asks the redeploy question and restarts (A
  `main.go:959-960`; B `manager.go:796-806`). A new compose label changes the digest
  and recreates the container. A label in a pod template rolls the pods. Item 16's
  "asks nothing" holds only for secrets, only on podman 4.7 and later, and only when
  the label write sits outside the "changed" test. Write the one-time restart into
  both Consequences lists, and land every unit-changing item (1, 6, 10) in one release
  so it happens once. Labels never go into a selector, a pod template or the CR's
  `podLabels`.
- **P7. Namespace labels cannot be back-filled.** Neither tool touches a namespace
  that exists, and labelling a found one would be false. Under item 1 as written, the
  namespace of every existing deployment stops being offered. Recommended: a labelled
  namespace may be deleted under `--no-prompt`; an unlabelled one is offered only at a
  terminal and never under `--no-prompt`. The label records "a tool created this", not
  an instance. `generate` output must not carry a labelled Namespace document.
- **P8. B's admin Secret must not be in the delta.** B's remove keeps it on purpose,
  because kept data needs it (B `ops_k8s.go:748-757`). Define the keep set as the names
  the config references, not the names it builds, and tie the admin Secret to the data
  rule.
- **P9. Dropping the old paths in the same step loses secrets.** A's credential secret
  names come from the workflows (`main.go:1044-1047`), so a names-only remove (A
  CR-0001) cannot compute them, and deployments made before the labels carry none. Keep
  A's old-unit read and fixed list, and B's by-name sweep of the certificate secret
  (fatal on failure today, B `ops_container.go:434-435`), for one release beside the
  labelled sweep. Remove deletes the labelled set and the by-name current set.
- **P10. Smaller points to write into the items.** Every create path passes the labels
  (F6). The listing uses `-o name` or metadata only, never `-o json` on Secrets. Any
  error or partial answer is a failed listing. The delete probe covers the kinds that
  have leftovers and sits in the existing preflight. State the failure policy: a
  deploy-time removal warns; a remove fails when a key-bearing secret survives. Compose
  needs labels on services only (no top-level volumes or networks exist in either
  tool; F8). Item 1's "Today" is wrong in two places: M defines no label key, and
  neither tool passes labels to `CreateSecret`. Item 1's "Undecided" and "Until then"
  predate the decisions in 16 and 18.

### 2.3 gdu 10, deploy digest

- **P11. The item cannot meet its goal as scoped.** A digest over the unit alone does
  not see a secret that was written and never applied. A digest that covers the
  secrets puts a secret-derived hash into a `Label=` line, which quadlet turns into
  podman's argv and which A writes to a 0644 file (`main.go:993`). That is what the
  2026-11-06 decision ruled out. An option that avoids both: keep the digest to the
  unit text, and detect an unapplied secret by comparing the secret ID the running
  container recorded (`Config.Secrets[].ID`, F1) with the store's ID for that name,
  which changes on every replace (F6). It needs M CR-0005 fixed and one host check.
- **P12. A mismatch must force `daemon-reload` before the restart.** Both tools reload
  only after a write, so after a failed reload systemd still holds the old unit.
- **P13. "One label key for both sinks" should read one key per tool,** passed by the
  caller. A uses `solace-connector/deploy-sha256` and B `solace-util.deploy-sha256`,
  and renaming either recreates every docker container.

### 2.4 gdu 2, 3, 4

- **P14. gdu 4's builder list is incomplete and its grammar is wrong for several
  operands.** `kube.DeleteNamespace` places `ns` as a bare positional and checks
  nothing (`kube/kube.go:124-126`); it is not in the list. Nor are `Get`,
  `GetNamespace`, the `containercli` lifecycle verbs, `podman.*`, `systemd.*` or
  `rbac.CanI`. B passes `pod/<name>` to `kube.Logs` (B `internal/k8s/ops.go:43`), and
  Secret, pod and StorageClass names are DNS subdomains, so `ValidLabel` and
  `ValidName` would refuse working calls. Split the rule: a typed grammar for
  bare-name builders, and for operands only "not empty, no leading `-`, no control
  character". Add a subdomain grammar. Size: 21 non-test call sites in A, 37 in B,
  about 50 direct calls in M's tests. Ship it as one breaking tag.
- **P15. "Pass a `${VAR}` on as written" is not what will happen in A.** Fields held
  to a grammar refuse `$`, so the documented `-Xmx${CONNECTOR_HEAP:1g}` becomes a
  validate error whose text still says the value "is expanded first" (A
  `validate.go:1496-1501`). Workflow files expand today too. Replace the sentence with
  a table of fields: error or literal. Name what replaces a per-environment value,
  since A gives the container no way to receive a variable. Say where the `~` step
  runs: A parses env.yaml at several sites and only one of them expands.
- **P16. B CR-0015 "expands nothing" removes more than intended.** `convert` resolves
  references to earlier assignments in the same file (B `internal/convert/parse.go:
  362-372`), which is how the bash files work. Decide whether that stays.
- **P17. gdu 2 contradicts M's developer guide,** which says `~` expansion stays in
  each tool (`docs/developer.md:96-98`). List the edit and name the receiving package.
- **P18. gdu 3's single order does not fit remove,** whose probes depend on an answer.
  State the order per command, and close "does remove resolve anything" as no.

### 2.5 gdu 6, 7, 9, 11 to 14, 17

- **P19. gdu 6: "bridge" is not neutral.** Rootless podman defaults to pasta, and
  compose `network_mode: bridge` leaves the project network. Check each spelling per
  engine on a host first. Add a separate check so both tools are not broken at once.
  A's read-only root needs its quadlet key in the release table, or `PodmanArgs`.
- **P20. gdu 7 adds the module's first code that writes a file.** Say so in the item
  and in the developer guide. Lift B's sequence (temp file in the target directory,
  chmod, write, sync, rename).
- **P21. gdu 9 needs three things it does not list.** A builder for the unit's load
  state (F9), since "absent and not loaded" cannot be told today. `-p <project>` on
  every compose call (F8). The order "secrets go only after the unit is confirmed
  gone", which A breaks today (F12). Share these and the parity rule, not a sequence
  that executes.
- **P22. gdu 11.** For A this is new behaviour with about three test files to
  rescript. Resolve per token set, so `--kubeconfig` is honoured and a run that reaches
  a second cluster pins each. With no current context: no pin, and say so. Write
  `--context=<name>`. A name is an arbitrary string (EKS writes `:` and `/`).
- **P23. gdu 12.** The tool-built deletes in B are not covered. Measure a real broker
  delete before choosing the default. Word a timeout as "requested and still in
  progress".
- **P24. gdu 14.** A accepts and documents `--tail all` (A `logs.go:21-41`). "A count
  only" changes A and no decision covers it.
- **P25. gdu 17.** A readiness wait right after a pod delete meets NotFound. State how
  that is handled.

### 2.6 Dev scripts and init-devops

- **P26. With the cov-queued skip, CI never runs the race detector** unless `cov`
  itself races. B sets `COVERMODE="count"` and says "cov never races" (B
  `scripts/dev.sh:35-36`); Go refuses `count` under `-race`. Decide in the contract.
- **P27. Removing the GOTOOLCHAIN export opens the go.work trap for `--local`.** Both
  tools' `go.work` have no `toolchain` line, both `go.mod` pin go1.27.4, the installed
  Go is 1.27.3, and M pins go1.27.3. A's script has no export, so A's gates already run
  on the installed Go. `--local` should refuse a `go.work` whose `toolchain` line is
  missing or differs from `go.mod`.
- **P28. `dev-all` cannot take `--local` today.** Its word check refuses it
  (`scripts/dev-all.sh:219-229`), `pin` takes exactly one argument, and the notice
  would name the wrong mode. `--local` goes to the tools, never to M, and never with
  `pin` or `resolve`. Release steps 2 to 4 use it; step 7 does not.
- **P29. M CR-0020 as decided blocks the next cycle.** Neither tool has an
  `adopt-go-deploy-utility` branch; both were merged and deleted. `pin` also modifies
  the two files it would then refuse on a second run. Add "create the branch" before
  `pin` in consuming.md, and define "modified" as anything beyond what `pin` writes.
- **P30. A CR-0038 contradicts the skill and M.** The skill says Dependabot manages
  actions only and templates ship verbatim; M's consuming guide says a module bump is
  always deliberate. A weekly gomod entry would bump M's pin without `dev-all pin` and
  the parity re-sync. Park it, or add an ignore for the module.
- **P31. Re-copy each workflow once, after the template changes.** B CR-0049 says copy
  now and B CR-0050 says copy after. B CR-0053 reads as if the template already
  compares the binary's version with the tag; it does not yet, and it is the same
  template change as A CR-0039. That check belongs in a dev-script task, since CI
  calls task names only. The PowerShell skeleton's timestamp would undo a fix
  M already has.
- **P32. A CR-0037 is superseded by gdu 15.**

### 2.7 A's file

- **P33. Order trap.** Seven decisions add validate or parse errors (CR-0005, 0010,
  0011, 0015, 0018, 0022, 0046). Until remove is independent of the spec on all three
  platforms, each new error blocks teardown. Do CR-0001 and CR-0008 first.
- **P34. CR-0005, the blocker, has no grammar** and no rule for reject against quote.
  One pattern cannot fit property keys, header names and logger names. CR-0010 lands
  with or before it, since dereferencing an alias renders subtrees that are dropped
  today.
- **P35. CR-0043 leaves a live wrong-mount bug** waiting on an unscheduled builder. B
  closes the same bug now by calling `quadlet.EscapePercent`. A can do the same.
- **P36. CR-0066 recreates B CR-0005** unless a post-remove check comes with it.
- **P37. Underspecified:** CR-0022, 0023, 0041, 0042, 0045, 0046. The findings file
  says what each lacks. The Severity column of CR-0062 to 0066 holds the confidence.

### 2.8 B's file

- **P38. CR-0009 closes half.** Without discovery, `operator deploy` can install a
  second operator and re-point the fixed-name ClusterRoleBinding. One read of that
  binding keeps the guard without "going looking".
- **P39. Two stated rules are broken by the table.** "No new questions" against
  CR-0055, which adds two. "It does not go looking" against the label listing; reword
  it to cover objects the tool's own labels prove it created.
- **P40. CR-0084 is built on a 100-byte file of six bare lines,** not on a transcript.
  Capture one failed and one clean chunk first.
- **P41. CR-0029 loads uid 0 with a warning** where the finding asked to refuse it.
  Record it as an exception with its reason, or take the patch.
- **P42. CR-0051 must wait for the first M tag B needs,** or every B change that uses
  new M API fails the default gate.
- **P43. Consequences list:** CR-0035 does not belong (those files fail today), CR-0072
  is missing, and several exit-code changes are not named.
- **P44. Certificate renewal:** see F4. Keep the restart for `manualPodRestart`, and
  say what `--restart` means on Kubernetes afterwards.

## 3. The questions asked

- **Label listing refused.** Warn, delete nothing, continue. M's least-privilege rule
  and the parity "Permission probes" decision already say so. Treat any error or
  partial answer the same way (F7). The warning names what could not be listed and the
  command the operator can run. B's by-name certificate sweep stays fatal; it is not a
  listing.
- **Kinds on Kubernetes.** Secrets plus A's Service; never a PVC or a PV. Two
  conditions: B's admin Secret follows the data rule (P8), and the current set is
  still deleted by name as today.
- **B CR-0085 and CR-0088.** Park both; accept neither.
  - CR-0085: if true, encrypted and replayable secrets reach logs, which the house rule
    forbids, so accepting it means changing CLAUDE.md. The captures owed for CR-0017
    and CR-0081 settle it. The fix is one masking helper and two call sites.
  - CR-0088: settled by `systemctl --user show <name>.service -p TimeoutStartUSec` on
    a deployed host. The fix is one `TimeoutStartSec=` line. Accept only if hosts
    always pre-pull.
- **Label key names.** A prefix of your own, not `app.kubernetes.io`. The operator
  writes `app.kubernetes.io/instance=<CR name>` on the Secrets it generates, which is
  the value B would use, so a selector on that key could match them. Two keys,
  `<prefix>/tool` and `<prefix>/instance`, the same string in every sink.
  `go-deploy-utility/tool` and `go-deploy-utility/instance` are valid in Kubernetes,
  docker, podman and quadlet, and A already writes slash keys. Check the value per
  sink: 63 characters with alphanumeric ends on Kubernetes only; B's `container.name`
  may be 100 long. Never truncate or hash a value. The namespace carries the tool key
  only. The deploy-digest key stays per tool.
- **Shape of the shared code.** Builders, parsers and one pure compare function; the
  tool runs them. M's rule is "builders and parsers are pure ... only code that takes
  a `runner.Runner` executes" (`docs/developer.md:29-30`), no package outside `runner`
  takes one today, and `nsgate` (a `List` builder plus the `Occupants` parser) is the
  precedent. Each tool keeps its order, probe and messages, and a parity rule states
  the order. The same answer closes the pipeline question in gdu 3 and the sequence
  question in gdu 9.
- **The three "not read" places.** Done: F10, F11, F12.
- **B CR-0069.** Refuse `--pod` with `--all` as a usage error. For one broker, narrow
  the pod and StatefulSet sections; that changes `status --detail` from the primary
  pod to every pod when `--pod` is unset, which should be named.

## 4. How to make the changes

### 4.1 Order

**Stage 0. Fix the plans.** Docs only, one session per repo: P1 to P44. Commit M's
`docs/todo.md`, which is modified and uncommitted.

**Stage 1. Work that needs no new M tag.** The three repos can move in parallel.

- M, a patch tag with no parity change: CR-0005, CR-0006 (a golden moves only where a
  value holds a `$`, which A's java-options and timezone checks refuse today), CR-0002,
  0007, 0008, 0010, 0011, 0012, 0013, 0014, 0022, 0024, and statement 1 of CR-0025.
- A: docs and repo hygiene; CR-0043 (three `EscapePercent` calls); CR-0030; CR-0061
  alone, then CR-0032; then CR-0001 and CR-0008 so remove no longer needs the spec;
  then the parse layer (CR-0018, 0046, 0010); then CR-0005 with 0009, 0011, 0016; then
  CR-0006 alone.
- B: the four blockers first (CR-0002, then 0003 and 0004; CR-0001 with 0011); then
  CR-0059 with 0071 before the other validate edits; then CR-0005 and CR-0031; the
  rest in the batches listed in the findings file.
- init-devops: pins, skeletons, contract, `ci.yml`, `tag.yml`; then one rebuild of the
  dev scripts and one re-copy of the workflows in each of M, A and B.

**Stage 2. One breaking M tag.** gdu 4 with M CR-0004 and CR-0023; the CR-0001
removals; gdu 12, 13 and 14; gdu 11; the gdu 9 pieces (load state, `-p`); the helpers
of gdu 5, 7 and 8. Then both tools adopt on a new adoption branch. The owner later
merged stages 2 and 3 into one breaking tag (decisions file, "Releases"); the review
of the exported surface in section 6 comes before that tag is cut.

**Stage 3. The label tag.** A short design note first: keys, shape, the migration
cases of P6, P7 and P9. Then gdu 1 with the wider listing (M CR-0003), gdu 16, 18 and
10, and gdu 6 in the same tag so each deployment restarts once. Run smoke S7 to S9
with the new rows before this tag (M CR-0009).

**Stage 4. Rules that need both tools aligned first.** gdu 2 after the `$` fix is in
a tag; gdu 3; the gdu 9 rule; gdu 15 once the tools hold the first tag they need;
gdu 17 after the cluster check.

### 4.2 Each step

1. Plan the step and approve it alone, as M's CLAUDE.md requires. Name who executes.
2. M first. Hand over `./scripts/dev.sh all scan`; continue from the output.
3. Create `adopt-go-deploy-utility` in each tool. `dev-all.sh pin <version>`.
4. One commit per tool. `dev-all.sh regen all scan cov` until all three pass.
5. `GOWORK=off ./scripts/dev.sh full` on Linux and Windows; the smoke rows for the
   packages touched; commit; you tag.
6. `dev-all.sh resolve`, then `GOWORK=off dev-all.sh all scan`, merge, tag the tools.

### 4.3 Checks only you can run

- Operator: `kubectl get sts <name>-pubsubplus-{p,b,m} -n <ns> -o
  jsonpath='{.spec.updateStrategy.type}'` under both strategies (F3). Renew a TLS
  Secret on a test broker and watch pod age (F4).
- Podman: `podman container inspect <name>` to read `Config.Secrets[].ID` before and
  after `secret create --replace` (P11). `podman secret create --label` and `secret
  inspect` on the oldest supported podman.
- Networking: what each candidate "bridge" spelling does on docker, rootful podman and
  rootless podman (P19).
- Time one real broker delete (P23).

## 5. Not checked

- No gate, engine or cluster was run. External behaviour is from documentation and
  upstream source on main branches; a given release may differ.
- The operator facts are from main, not from 1.4.2.
- Low-severity problems were not re-opened; they are in the findings file as returned.
- B's DR findings are parked and were not read.

## 6. Review of the exported surface (to do)

Added 2026-11-14 at the owner's request: review every name this module exports for the
two tools, and decide for each whether it belongs. Proposed timing: before the breaking
tag is cut, so that whatever is unexported, removed, moved or renamed rides that one tag
and both tools adapt once.

### 6.1 Counts

243 exported names in 22 packages, `internal/build` not included: 148 functions, 32
methods, 37 types and 26 constants. Of the 211 that are not methods, both tools use 136,
solmq-conn-util (A) alone 9, solace-util (B) alone 28, and 38 are named by neither.

Counted on 2026-11-14 by matching `package.Name` (and `.Method(` for methods) in every
Go file of A and B, tests included. A name reached only through another name, such as a
field type or a function an `Assert` wrapper calls, counts as unused here. A name used by
a tool's tests only counts as used.

### 6.2 Questions for the review

1. The 38 names neither tool names. Which are reached indirectly and stay (the `Check*`
   functions behind the `Assert*` ones; `State`, `Health` and `Mount` as fields of
   `inspect.Container`; `confirm.Form` and `NotTerminalError` as returned values), and
   which go. The four removals of CR-0001 are decided; `quadlet.Escape`,
   `watch.CheckInterval`, `watch.Sleep`, `rbac.Answer`, `rbac.Refused`, `envref.ValidName`,
   `confirm.Line`, `Yes` and `YesNo` are open.
2. The 37 names one tool uses, against the rule that a package needs two consumers. For
   each: the other tool adopts it (`kube.ValidLabel` in B, `quadlet.EscapePercent` in A,
   both already decided elsewhere), it moves back into its one tool, or it is recorded as
   an exception with its reason. The largest groups: B's limit checks in `rootless`, the
   capture helpers in `runner` (`Combined` and `Split` in A, `Output`, `Quote`, `MaskEnv`
   and `Resolve` in B), and the child-exit trio in `exitcode`.
3. The names the breaking tag changes anyway: the builders of gdu 4, the delete builders
   of gdu 12, the secret emitters of CR-0004, and the new builders of gdu 9, 10 and 11.
   Review their signatures together, so one convention holds across packages.
4. What is missing: behaviour both tools still implement themselves, which the to-do
   already plans to share (gdu 2, 5, 7 and 8), and anything else found in the review.
5. Placement: whether each name sits in the right package.

### 6.3 Inventory

A name with no mark is used by both tools. `(A)` is solmq-conn-util only, `(B)` is
solace-util only, `(-)` is named by neither. Methods are listed in brackets after their
type and carry no mark.

- `abbrev`
  - Types: `Entry`, `Set` [`Entries`, `Expand`, `Kind`, `List`, `Names`, `Short`, `Shorts`]
  - Functions: `New`
- `confirm`
  - Types: `Form` (-), `NotTerminalError` (-) [`Error`], `Options`, `Prompter` [`Ask`,
    `Gate`, `ReadLine`, `Terminal`]
  - Constants: `DefaultNoPromptFlag` (-), `Lenient`, `Strict` (B)
  - Functions: `IsCharDevice` (A), `IsNotTerminal` (-), `Line` (-), `New`,
    `StdinIsTerminal` (B), `Unanswered`, `Yes` (-), `YesNo` (-)
- `containerart/artifacttest`
  - Functions: `AssertAbsent`, `AssertComposePosture`, `AssertQuadletKeys`,
    `AssertQuadletPosture`, `AssertSecretTargets`, `CheckAbsent` (-),
    `CheckComposePosture` (-), `CheckQuadletKeys` (-), `CheckQuadletPosture` (-),
    `CheckSecretTargets` (-), `HostNetwork`, `Widening`
- `containerart/compose`
  - Types: `Secret`
  - Constants: `Posture`
  - Functions: `Escape`, `FlowList`, `Quote`, `ServiceSecrets`, `TopSecrets`
- `containerart/quadlet`
  - Constants: `Posture`
  - Functions: `Environment`, `Escape` (-), `EscapePercent` (B), `Secret`
- `containercli`
  - Types: `Engine`, `ExecOptions`, `LogOptions`, `Row` (B)
  - Constants: `Docker` (-), `Podman` (-)
  - Functions: `ByImage`, `ComposeDigest`, `ComposeDown`, `ComposeUp`, `EngineHint` (A),
    `Exec`, `HasName`, `Info`, `InfoError`, `List`, `Logs`, `NoPrevious`, `ParseList`,
    `Remove` (-), `Restart`, `Start`, `Stop`, `Upload`, `ValidName`
- `containercli/inspect`
  - Types: `Container`, `Health` (-), `Mount` (-), `State` (-)
  - Functions: `Containers`, `Decode`, `Printable`
- `envref`
  - Types: `Ref` [`Check`, `Resolve`], `UnsetError` (A) [`Error`]
  - Functions: `ValidName` (-)
- `execguard`
  - Types: `Approvals`, `Rules`
  - Constants: `ApproveFlag` (-), `PathSeparators` (B), `UnsafeChars`
  - Functions: `Approve`, `Check`, `IsControl`, `IsInvisible`
- `exitcode`
  - Constants: `Failure`, `OK`, `Usage`
  - Functions: `AsUsage`, `ChildExit` (B), `ChildStatus` (B), `Code`, `Exited` (B),
    `IsUsage`, `Usagef`
- `kube`
  - Types: `ExecOptions`, `GetOptions`, `LogOptions`, `NamespaceState`
  - Constants: `NamespaceAbsent`, `NamespaceExists`, `NamespaceUnknown`, `RolloutTimeout`
  - Functions: `Apply`, `Delete`, `DeleteNamespace`, `Exec`, `Forbidden`, `Get`,
    `GetNamespace`, `Logs`, `MissingNamespace`, `NamespaceManifest`, `ParseNamespace`,
    `RolloutRestart`, `RolloutStatus`, `Scale`, `Upload`, `ValidLabel` (A)
- `kube/nsgate`
  - Types: `Object`
  - Functions: `DeleteProbe`, `List`, `Occupants`, `Protected`
- `kube/rbac`
  - Types: `Probe` [`String`], `RefusedError` [`Error`]
  - Functions: `Answer` (-), `CanI`, `Check`, `ClusterScoped` (B), `Refused` (-),
    `Verdict` (B)
- `notice`
  - Types: `Notice` (B) [`Lines`, `Markdown`, `Plain`]
  - Constants: `Width` (-)
  - Functions: `New` (B), `Short` (A)
- `notice/noticetest`
  - Types: `Options`
  - Functions: `Assert` (A), `AssertASCII` (A), `AssertDocs`, `Check` (-), `CheckDocs` (-)
- `paritydoc`
  - Constants: `Title` (-)
  - Functions: `Body` (-), `Check`, `Compose` (-)
- `podman`
  - Constants: `SystemQuadletDir` (B)
  - Functions: `CreateSecret`, `IDMap`, `QuadletDir`, `RemoveSecrets`, `SecretData`,
    `SecretMatches`, `SupportsReplace`, `Version`
- `rootless`
  - Types: `Session` (-) [`Env`, `String`]
  - Constants: `Unbounded` (B)
  - Functions: `CheckCore` (B), `CheckIDMap`, `CheckLinger`, `Decide`, `HardLimit` (B),
    `LimitText` (B), `NextFreeRange`, `ParseLimit` (B), `ProcLimits` (B), `SubIDFix`,
    `UserSession`
- `runner`
  - Types: `Cmd` [`Check`], `Echo` (B) [`Attach`, `Run`], `Exec` [`Attach`, `Run`], `Runner`
  - Constants: `DefaultGrace` (-)
  - Functions: `CheckAttach` (-), `CheckRun` (-), `Combined` (A), `MaskEnv` (B),
    `Output` (B), `Quote` (B), `Resolve` (B), `Split` (A)
- `runner/runnertest`
  - Types: `Call` (-) [`Has`, `String`], `Fake` [`Attach`, `Calls`, `Run`], `Reply` (B)
  - Functions: `FailOn` (-), `Seq` (B)
- `systemd`
  - Functions: `Active`, `DaemonReload`, `EnableLinger`, `IsActive`, `NRestarts`,
    `ParseLinger`, `ParseNRestarts`, `Restart`, `ShowLinger`, `Start`, `State`, `Stop`,
    `Stopped` (B), `Systemctl` (B)
- `watch`
  - Types: `Options`
  - Constants: `ClearScreen`, `DefaultInterval`, `MaxInterval`, `MinInterval`
  - Functions: `CanRedraw`, `CheckInterval` (-), `Loop` (-), `LoopFrames`, `Sleep` (-)
