# Graph Report - go-deploy-utility  (2026-11-14)

## Corpus Check
- 78 files · ~171,093 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 3 file(s) not represented in the graph (top: (none) 3)

## Summary
- 1188 nodes · 3670 edges · 66 communities (61 shown, 5 thin omitted)
- Extraction: 69% EXTRACTED · 31% INFERRED · 0% AMBIGUOUS · INFERRED: 1146 edges (avg confidence: 0.92)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `ff6f9e7e`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- dev.sh
- Appendix C: deferred code stages (not in this plan, separate approval)
- testing.T
- Package layout and build order
- dev.ps1
- noticetest_test.go
- prompter_test.go
- abbrev_test.go
- paritydoc_test.go
- handoff/README.md
- github.com/hafio/go-deploy-utility
- exitcode_test.go
- rootless.go
- Handoff: aligning solmq-conn-util and solace-util on Docker, Podman and Kubernetes
- execguard_test.go
- envref_test.go
- artifacttest_test.go
- nsgate.go
- .Run
- containercli.go
- kube.go
- Smoke checks
- inspect.go
- watch_test.go
- Plan review findings, 2026-11-13
- rbac.go
- dev-all.sh
- go_pkg_strings
- compose_test.go
- containercli_test.go
- dev-all.ps1
- artifacttest.go
- Findings by area
- echo_test.go
- Echo
- Cmd
- systemd_test.go
- Decisions on the 2026-11-11 critical review
- Appendix C: deferred code stages (not in this plan, separate approval)
- Plan review, 2026-11-13
- tests
- ComposeDigest
- Critical review - go-deploy-utility
- 2. What to fix in the plans before building
- InfoError
- build
- 11. Resolve and pin the kube context
- To do
- 14. One check for the log tail count
- 15. Dev scripts build the pinned module unless asked for the checkout
- 17. A readiness wait for OnDelete StatefulSets
- 18. Kubernetes objects a config no longer names are removed
- 19. One shape for the platform sections of the env file
- 3. One step order for deploy and remove on every platform
- 5. One diff helper for golden tests
- 6. Every artifact states its network mode
- 7. One atomic write for files a tool installs
- 8. One gate for the test catalogue
- 9. Remove is idempotent on every platform
- 2. One guideline for expansion in config values
- 4. How to make the changes
- 4. Builders refuse a name outside its grammar
- CR-0001 [minor] [new] Exported API with no consumer in either tool
- github.com/hafio/go-deploy-utility/docs/review

## God Nodes (most connected - your core abstractions)
1. `6.3 Inventory` - 158 edges
2. `Package layout and build order` - 86 edges
3. `Cmd` - 64 edges
4. `runner, runner_test.go` - 54 edges
5. `Runner` - 48 edges
6. `Echo` - 41 edges
7. `helperArgv()` - 36 edges
8. `helperOn()` - 36 edges
9. `Check()` - 31 edges
10. `execguard, execguard_test.go` - 29 edges

## Surprising Connections (you probably didn't know these)
- `Today` --references--> `HostNetwork()`  [INFERRED]
  docs/todo.md → containerart/artifacttest/artifacttest.go
- `Undecided` --references--> `CheckQuadletKeys()`  [INFERRED]
  docs/todo.md → containerart/artifacttest/artifacttest.go
- `CR-0011 [minor] [new] CheckSecretTargets passes a compose secrets block it does not recognise` --references--> `AssertSecretTargets()`  [INFERRED]
  docs/review/2026-11-11-critical-review.md → containerart/artifacttest/artifacttest.go
- `4.1 Order` --references--> `EscapePercent()`  [INFERRED]
  docs/review/2026-11-13-plan-review.md → containerart/quadlet/quadlet.go
- `CR-0024 [nit] [new] inspect.Printable keeps invisible format characters that execguard refuses` --references--> `Printable()`  [INFERRED]
  docs/review/2026-11-11-critical-review.md → containercli/inspect/inspect.go

## Import Cycles
- None detected.

## Communities (66 total, 5 thin omitted)

### Community 0 - "dev.sh"
Cohesion: 0.19
Nodes (19): c(), expand(), finish(), log_begin(), NO_COLOR, run(), dev.sh script, die() (+11 more)

### Community 1 - "Appendix C: deferred code stages (not in this plan, separate approval)"
Cohesion: 0.12
Nodes (13): A.1 Facts that shape the answer, A.2 Enforcement options, weakest to strongest, A.3 Key format and verification (options 1 and 2), A.4 Where the key comes from, A.5 Where the gate plugs in, A.6 Shared module, A.7 Code the two repos could share (synergy), A.8 Shared module: layout and GitHub mechanics (+5 more)

### Community 2 - "testing.T"
Cohesion: 0.08
Nodes (82): runner, runner_test.go, runner/runnertest, runnertest_test.go, Runner seams, Combined(), Output(), Split(), attachFiles(), captureStdout() (+74 more)

### Community 3 - "Package layout and build order"
Cohesion: 0.09
Nodes (42): Printable(), Adding a package, Adoption order, Dev tasks, Developer guide, Package layout and build order, Rules for shared code, CR-0023 [nit] [new] Builders in one package disagree on whether an empty name is refused (+34 more)

### Community 4 - "dev.ps1"
Cohesion: 0.22
Nodes (12): Get-Log(), Get-Now(), Invoke-Logged(), Task-build(), Task-cov(), Task-graphify(), Task-scan(), Task-test() (+4 more)

### Community 5 - "noticetest_test.go"
Cohesion: 0.09
Nodes (57): 3. House rules that bind the next session, Files, parity.md content, Stage 0: capture (both repos, docs only), Verification, CR-0013 [nit] [new] CheckDocs passes when Exclude removes every selected doc, internal/build, build_test.go, notice, notice_test.go (+49 more)

### Community 6 - "prompter_test.go"
Cohesion: 0.08
Nodes (39): countingReader, Form, Line(), TestLineReusesBufferedReader(), TestPromptYes(), TestPromptYesNo(), Yes(), YesNo() (+31 more)

### Community 7 - "abbrev_test.go"
Cohesion: 0.15
Nodes (21): checkShort(), New(), mustPanic(), sample(), TestAccessorsReturnCopies(), TestDigitsAreAllowed(), TestEntriesCloneKeepsNilShort(), TestExpandResolvesBothSpellings() (+13 more)

### Community 8 - "paritydoc_test.go"
Cohesion: 0.13
Nodes (27): Conventions, go-deploy-utility, Process, Decisions on the findings, paritydoc, paritydoc_test.go, Contract, Decisions, Divergences (+19 more)

### Community 12 - "exitcode_test.go"
Cohesion: 0.11
Nodes (26): exitcode, exitcode_test.go, childExitError, AsUsage(), ChildExit(), ChildStatus(), Code(), Exited(), IsUsage() (+18 more)

### Community 13 - "rootless.go"
Cohesion: 0.11
Nodes (33): rootless, rootless_test.go, CheckCore(), CheckIDMap(), CheckLinger(), covers(), Decide(), HardLimit(), LimitText() (+25 more)

### Community 14 - "Handoff: aligning solmq-conn-util and solace-util on Docker, Podman and Kubernetes"
Cohesion: 0.12
Nodes (15): 1. What this is, 2. State at export, 4. Decisions already taken, 5.1 Same in both (the contract), 5.2 Different by policy, 5. Exact similarities and differences, 6.1 Repo A, solmq-conn-util, 6.3 Not scanned or not verified (+7 more)

### Community 16 - "execguard_test.go"
Cohesion: 0.14
Nodes (44): execguard, execguard_test.go, Approvals, Approve(), base(), Check(), checkBinary(), checkFlagShape(), checkToken() (+36 more)

### Community 17 - "envref_test.go"
Cohesion: 0.20
Nodes (16): envref, envref_test.go, env(), ref(), TestCheckBothSet(), TestCheckInvalidName(), TestCheckReferenceForm(), TestMessagesNeverCarryTheValue(), TestRefWiringErrors() (+8 more)

### Community 18 - "artifacttest_test.go"
Cohesion: 0.30
Nodes (18): recorder, CheckAbsent(), HostNetwork(), composeFile(), TestAssertPassesQuietly(), TestCheckAbsent(), TestCheckComposePosture(), TestCheckQuadletKeys() (+10 more)

### Community 19 - "nsgate.go"
Cohesion: 0.15
Nodes (30): kube/nsgate, nsgate_test.go, 1. Ownership labels on every artifact, Decided, Goal, Scope, Today, Undecided, Until then (+22 more)

### Community 20 - ".Run"
Cohesion: 0.31
Nodes (5): Exec, applyEnv(), CheckAttach(), CheckRun(), Resolve()

### Community 21 - "containercli.go"
Cohesion: 0.14
Nodes (18): ByImage(), composeCmd(), ComposeDown(), ComposeUp(), HasName(), Info(), List(), Logs() (+10 more)

### Community 22 - "kube.go"
Cohesion: 0.10
Nodes (44): 2.4 gdu 2, 3, 4, kube, kube_test.go, 12. A bound on every delete, Decided, Goal, Scope, Today, Today (+36 more)

### Community 23 - "Smoke checks"
Cohesion: 0.09
Nodes (21): Consuming go-deploy-utility, Develop against a local checkout, Pin an exact version, Release and merge order, Environment, Runs, S7: redeploy asks before any write (not yet run), S8: the shared watch loop (not yet run) (+13 more)

### Community 24 - "inspect.go"
Cohesion: 0.17
Nodes (18): Containers(), Decode(), health(), TestContainersArgv(), TestDecodeDocker(), TestDecodeHealthSpellings(), TestDecodeNeverReadsEnv(), TestDecodePartial() (+10 more)

### Community 25 - "watch_test.go"
Cohesion: 0.14
Nodes (30): What each version adds, watch, watch_test.go, Today, Options, CheckInterval(), Loop(), loop(), LoopFrames() (+22 more)

### Community 26 - "Plan review findings, 2026-11-13"
Cohesion: 0.06
Nodes (31): Answers, Answers, Answers, Answers, Answers, Answers, Answers, Dev scripts, CI and init-devops; gdu 15 (+23 more)

### Community 27 - "rbac.go"
Cohesion: 0.19
Nodes (16): kube/rbac, rbac_test.go, Answer(), CanI(), Check(), ClusterScoped(), Probe, Refused(), TestAnswer() (+8 more)

### Community 28 - "dev-all.sh"
Cohesion: 0.41
Nodes (12): is_tag(), known(), name(), notice(), notice_end(), pin(), required(), resolve() (+4 more)

### Community 29 - "go_pkg_strings"
Cohesion: 0.18
Nodes (18): Environment(), Escape(), EscapePercent(), Secret(), TestEnvironmentIsAlwaysQuoted(), TestEscapePercent(), TestEscapeWritesControlsAsCEscapes(), TestPostureBytes() (+10 more)

### Community 30 - "compose_test.go"
Cohesion: 0.21
Nodes (18): Secret, Escape(), FlowList(), Quote(), ServiceSecrets(), goldenSecrets(), TestComposeQuoteEscapes(), TestEscape() (+10 more)

### Community 31 - "containercli_test.go"
Cohesion: 0.18
Nodes (19): Exec(), ParseList(), TestByImage(), TestComposeDigest(), TestComposeUpAndDown(), TestExec(), TestExecRefuses(), TestHasNameIsExact() (+11 more)

### Community 32 - "dev-all.ps1"
Cohesion: 0.38
Nodes (9): Get-Name(), Get-NoticeEnd(), Get-Required(), Invoke-Pin(), Invoke-Resolve(), Die(), Sync-Parity(), Show-Notice() (+1 more)

### Community 33 - "artifacttest.go"
Cohesion: 0.26
Nodes (17): unitLine, AssertAbsent(), AssertComposePosture(), AssertQuadletKeys(), AssertQuadletPosture(), AssertSecretTargets(), CheckComposePosture(), checkComposeSecret() (+9 more)

### Community 34 - "Findings by area"
Cohesion: 0.12
Nodes (17): cicd, conventions, correctness, CR-0003 [major] [new] Namespace gate lists only all, PVCs, Secrets and ConfigMaps, so a namespace holding other kinds is offered for deletion, CR-0007 [nit] [new] cmd, stdin, shQuote and Upload are duplicated across the builder packages, CR-0008 [nit] [new] dev-all discovers sibling tasks by scraping their dev scripts with two different regexes, CR-0015 [minor] [new] No gate runs before a tag exists, and for a Go module the tag is the release, CR-0016 [minor] [new] CI never runs the tests under the race detector (+9 more)

### Community 35 - "echo_test.go"
Cohesion: 0.25
Nodes (14): runner, echo_test.go, TestEchoDefaultWriter(), TestEchoOutput(), TestEchoOutputInput(), TestEchoRefusesWhatExecRefuses(), TestEchoRun(), TestEchoRunEnv(), TestEchoRunEnvNoEnv() (+6 more)

### Community 36 - "Echo"
Cohesion: 0.27
Nodes (10): 5.3 Runner seams (why nothing is shared yet), 6.2 Repo B, solace-util, A.7.1 Runtime layer: docker, podman, kubectl, Stage 1: B runner seam, minimum for stdin plus env, Echo, MaskEnv(), printable(), Quote() (+2 more)

### Community 37 - "Cmd"
Cohesion: 0.35
Nodes (12): Cmd, DaemonReload(), EnableLinger(), IsActive(), NRestarts(), Restart(), ShowLinger(), Start() (+4 more)

### Community 38 - "systemd_test.go"
Cohesion: 0.26
Nodes (10): systemd, systemd_test.go, Active(), ParseLinger(), ParseNRestarts(), State(), Stopped(), TestNRestarts(), TestParseLinger() (+2 more)

### Community 39 - "Decisions on the 2026-11-11 critical review"
Cohesion: 0.20
Nodes (10): Checks the owner runs, Decisions on the 2026-11-11 critical review, Decisions on the plan review, Labels (gdu 1, 16, 18), My readings, to correct if wrong, Other items, Releases, Review points carried into the items without a separate decision (+2 more)

### Community 40 - "Appendix C: deferred code stages (not in this plan, separate approval)"
Cohesion: 0.22
Nodes (9): Appendix C: deferred code stages (not in this plan, separate approval), Open decisions (recorded in parity.md, not settled here), Stage 2: B compose on stdin, like A, Stage 3: B non-TTY confirmations fail fast, Stage 4: A podman session and host parity, Stage 5: A kubernetes parity, Stage 6: podman secrets, both, Stage 7: design spike, root deploys for a declared rootless user (both) (+1 more)

### Community 41 - "Plan review, 2026-11-13"
Cohesion: 0.25
Nodes (5): 1. Facts this review settled, 3. The questions asked, 5. Not checked, Plan review, 2026-11-13, Verdict

### Community 42 - "tests"
Cohesion: 0.29
Nodes (7): squeeze(), CR-0009 [major] [new] Real-engine smoke checks for S7 to S9 are recorded as not run for a tagged release, CR-0010 [minor] [new] Widening and HostNetwork miss the podman flag spellings that ride PodmanArgs, CR-0011 [minor] [new] CheckSecretTargets passes a compose secrets block it does not recognise, CR-0012 [minor] [new] The Windows console branch of CanRedraw is never executed by a test, CR-0014 [nit] [new] TestBodyShape pins the parity body to a line count, tests

### Community 43 - "ComposeDigest"
Cohesion: 0.33
Nodes (7): ComposeDigest(), 10. One deploy digest for podman units, Decided, Goal, Scope, Today, Undecided

### Community 44 - "Critical review - go-deploy-utility"
Cohesion: 0.29
Nodes (7): Accepted (not re-litigated), Coverage, Critical review - go-deploy-utility, Dynamic checks, Fixed since last report, Summary, Top risks

### Community 45 - "2. What to fix in the plans before building"
Cohesion: 0.29
Nodes (7): 2.1 Across the three plans, 2.2 Labels: gdu 1, 16, 18, 2.3 gdu 10, deploy digest, 2.5 gdu 6, 7, 9, 11 to 14, 17, 2.6 Dev scripts and init-devops, 2.8 B's file, 2. What to fix in the plans before building

### Community 46 - "InfoError"
Cohesion: 0.67
Nodes (5): EngineHint(), InfoError(), NoPrevious(), TestInfoErrorPerEngine(), Engine

### Community 47 - "build"
Cohesion: 0.33
Nodes (6): build, CR-0018 [minor] [new] cov truncates cov.log before the previous total can be compared, CR-0019 [minor] [new] dev.ps1 and dev.sh disagree on cov exit status, task-name matching and timestamp culture, CR-0020 [nit] [new] dev-all pin edits both tools with no clean-tree or branch check and can stop half way, CR-0021 [nit] [new] dev-all.sh needs bash 4 but does not check for it, CR-0022 [nit] [new] The module zip carries graphify-out and docs/handoff, 77 percent of its bytes

### Community 48 - "11. Resolve and pin the kube context"
Cohesion: 0.33
Nodes (6): 11. Resolve and pin the kube context, Decided, Goal, Scope, Today, Undecided

### Community 49 - "To do"
Cohesion: 0.33
Nodes (6): 13. A watch loop stops on an error no retry can fix, Decided, Goal, Scope, To do, Undecided

### Community 50 - "14. One check for the log tail count"
Cohesion: 0.33
Nodes (6): 14. One check for the log tail count, Decided, Goal, Scope, Today, Undecided

### Community 51 - "15. Dev scripts build the pinned module unless asked for the checkout"
Cohesion: 0.33
Nodes (6): 15. Dev scripts build the pinned module unless asked for the checkout, Decided, Goal, Scope, Today, Undecided

### Community 52 - "17. A readiness wait for OnDelete StatefulSets"
Cohesion: 0.33
Nodes (6): 17. A readiness wait for OnDelete StatefulSets, Decided, Goal, Scope, Today, Undecided

### Community 53 - "18. Kubernetes objects a config no longer names are removed"
Cohesion: 0.33
Nodes (6): 18. Kubernetes objects a config no longer names are removed, Decided, Goal, Scope, Today, Undecided

### Community 54 - "19. One shape for the platform sections of the env file"
Cohesion: 0.33
Nodes (6): 19. One shape for the platform sections of the env file, Decided, Goal, Scope, Today, Undecided

### Community 55 - "3. One step order for deploy and remove on every platform"
Cohesion: 0.33
Nodes (6): 3. One step order for deploy and remove on every platform, Decided, Goal, Scope, Today, Undecided

### Community 56 - "5. One diff helper for golden tests"
Cohesion: 0.33
Nodes (6): 5. One diff helper for golden tests, Decided, Goal, Scope, Today, Undecided

### Community 57 - "6. Every artifact states its network mode"
Cohesion: 0.33
Nodes (6): 6. Every artifact states its network mode, Decided, Goal, Scope, Today, Undecided

### Community 58 - "7. One atomic write for files a tool installs"
Cohesion: 0.33
Nodes (6): 7. One atomic write for files a tool installs, Decided, Goal, Scope, Today, Undecided

### Community 59 - "8. One gate for the test catalogue"
Cohesion: 0.33
Nodes (6): 8. One gate for the test catalogue, Decided, Goal, Scope, Today, Undecided

### Community 60 - "9. Remove is idempotent on every platform"
Cohesion: 0.33
Nodes (6): 9. Remove is idempotent on every platform, Decided, Goal, Scope, Today, Undecided

### Community 61 - "2. One guideline for expansion in config values"
Cohesion: 0.40
Nodes (5): 2. One guideline for expansion in config values, Decided, Goal, Scope, Undecided

### Community 62 - "4. How to make the changes"
Cohesion: 0.50
Nodes (4): 4.1 Order, 4.2 Each step, 4.3 Checks only you can run, 4. How to make the changes

### Community 63 - "4. Builders refuse a name outside its grammar"
Cohesion: 0.50
Nodes (4): 4. Builders refuse a name outside its grammar, Goal, Scope, Undecided

### Community 64 - "CR-0001 [minor] [new] Exported API with no consumer in either tool"
Cohesion: 0.67
Nodes (3): CR-0001 [minor] [new] Exported API with no consumer in either tool, CR-0002 [minor] [new] execguard accepts a flag value that makes the allowed binary run a file beside the env file, design

## Knowledge Gaps
- **3 isolated node(s):** `github.com/hafio/go-deploy-utility/docs/review`, `github.com/hafio/go-deploy-utility`, `NO_COLOR`
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 225 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **5 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `6.3 Inventory` connect `prompter_test.go` to `testing.T`, `Package layout and build order`, `noticetest_test.go`, `abbrev_test.go`, `paritydoc_test.go`, `exitcode_test.go`, `rootless.go`, `execguard_test.go`, `envref_test.go`, `artifacttest_test.go`, `nsgate.go`, `.Run`, `containercli.go`, `kube.go`, `inspect.go`, `watch_test.go`, `rbac.go`, `go_pkg_strings`, `compose_test.go`, `containercli_test.go`, `artifacttest.go`, `Echo`, `Cmd`, `systemd_test.go`, `ComposeDigest`, `InfoError`?**
  _High betweenness centrality (0.177) - this node is a cross-community bridge._
- **Are the 157 inferred relationships involving `6.3 Inventory` (e.g. with `Entry` and `.Entries()`) actually correct?**
  _`6.3 Inventory` has 157 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/hafio/go-deploy-utility/docs/review`, `github.com/hafio/go-deploy-utility`, `NO_COLOR` to the rest of the system?**
  _3 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Appendix C: deferred code stages (not in this plan, separate approval)` be split into smaller, more focused modules?**
  _Cohesion score 0.125 - nodes in this community are weakly interconnected._
- **Why does `To do` connect `To do` to `Package layout and build order`, `nsgate.go`, `kube.go`, `Smoke checks`, `go_pkg_strings`, `ComposeDigest`, `11. Resolve and pin the kube context`, `14. One check for the log tail count`, `15. Dev scripts build the pinned module unless asked for the checkout`, `17. A readiness wait for OnDelete StatefulSets`, `18. Kubernetes objects a config no longer names are removed`, `19. One shape for the platform sections of the env file`, `3. One step order for deploy and remove on every platform`, `5. One diff helper for golden tests`, `6. Every artifact states its network mode`, `7. One atomic write for files a tool installs`, `8. One gate for the test catalogue`, `9. Remove is idempotent on every platform`, `2. One guideline for expansion in config values`, `4. Builders refuse a name outside its grammar`?**
  _High betweenness centrality (0.075) - this node is a cross-community bridge._
- **Are the 85 inferred relationships involving `Package layout and build order` (e.g. with `Prompter` and `.Ask()`) actually correct?**
  _`Package layout and build order` has 85 INFERRED edges - model-reasoned connections that need verification._
- **Should `testing.T` be split into smaller, more focused modules?**
  _Cohesion score 0.08282828282828283 - nodes in this community are weakly interconnected._