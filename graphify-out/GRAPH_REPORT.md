# Graph Report - go-deploy-utility  (2026-11-04)

## Corpus Check
- 61 files · ~84,428 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 3 file(s) not represented in the graph (top: (none) 3)

## Summary
- 806 nodes · 2663 edges · 28 communities (24 shown, 4 thin omitted)
- Extraction: 73% EXTRACTED · 27% INFERRED · 0% AMBIGUOUS · INFERRED: 728 edges (avg confidence: 0.9)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `889ca6b3`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- dev.sh
- podman_test.go
- testing.T
- Appendix C: deferred code stages (not in this plan, separate approval)
- dev.ps1
- noticetest_test.go
- testing.T
- abbrev_test.go
- paritydoc_test.go
- handoff/README.md
- github.com/hafio/go-deploy-utility
- exitcode_test.go
- rootless.go
- echo_test.go
- execguard_test.go
- envref_test.go
- artifacttest.go
- compose_test.go
- runnertest_test.go
- containercli.go
- systemd.go
- watch_test.go
- inspect.go
- Runner
- Package layout and build order
- .Run

## God Nodes (most connected - your core abstractions)
1. `Package layout and build order` - 65 edges
2. `runner, runner_test.go` - 55 edges
3. `Cmd` - 49 edges
4. `Runner` - 43 edges
5. `Echo` - 37 edges
6. `helperArgv()` - 36 edges
7. `helperOn()` - 36 edges
8. `Check()` - 30 edges
9. `execguard, execguard_test.go` - 28 edges
10. `Notice` - 27 edges

## Surprising Connections (you probably didn't know these)
- `7. Next stages` --references--> `Echo`  [INFERRED]
  docs/handoff/HANDOFF.md → runner/echo.go
- `Package layout and build order` --references--> `Prompter`  [INFERRED]
  docs/developer.md → confirm/prompter.go
- `Package layout and build order` --references--> `Widening()`  [INFERRED]
  docs/developer.md → containerart/artifacttest/artifacttest.go
- `Package layout and build order` --references--> `HostNetwork()`  [INFERRED]
  docs/developer.md → containerart/artifacttest/artifacttest.go
- `Package layout and build order` --references--> `quadletKeySince()`  [INFERRED]
  docs/developer.md → containerart/artifacttest/artifacttest.go

## Import Cycles
- None detected.

## Communities (28 total, 4 thin omitted)

### Community 0 - "dev.sh"
Cohesion: 0.19
Nodes (19): c(), expand(), finish(), log_begin(), NO_COLOR, run(), dev.sh script, die() (+11 more)

### Community 1 - "podman_test.go"
Cohesion: 0.18
Nodes (22): podman, podman_test.go, cmd(), CreateSecret(), IDMap(), QuadletDir(), RemoveSecrets(), SupportsReplace(), env() (+14 more)

### Community 2 - "testing.T"
Cohesion: 0.14
Nodes (60): runner, runner_test.go, Combined(), Split(), attachFiles(), captureStdout(), helperArgv(), helperOn(), readFile() (+52 more)

### Community 3 - "Appendix C: deferred code stages (not in this plan, separate approval)"
Cohesion: 0.04
Nodes (43): 2. State at export, 3. House rules that bind the next session, 4. Decisions already taken, 5.1 Same in both (the contract), 5.2 Different by policy, 5. Exact similarities and differences, 6.1 Repo A, solmq-conn-util, 6.3 Not scanned or not verified (+35 more)

### Community 4 - "dev.ps1"
Cohesion: 0.22
Nodes (12): Get-Log(), Get-Now(), Invoke-Logged(), Task-build(), Task-cov(), Task-graphify(), Task-scan(), Task-test() (+4 more)

### Community 5 - "noticetest_test.go"
Cohesion: 0.11
Nodes (46): notice, notice_test.go, notice/noticetest, noticetest_test.go, checkCovers(), checkKind(), checkName(), Notice, New(), Short() (+38 more)

### Community 6 - "testing.T"
Cohesion: 0.06
Nodes (39): checkShort(), countingReader, Form, Line(), TestLineReusesBufferedReader(), TestPromptYes(), TestPromptYesNo(), Yes() (+31 more)

### Community 7 - "abbrev_test.go"
Cohesion: 0.16
Nodes (20): New(), mustPanic(), sample(), TestAccessorsReturnCopies(), TestDigitsAreAllowed(), TestEntriesCloneKeepsNilShort(), TestExpandResolvesBothSpellings(), TestKind() (+12 more)

### Community 8 - "paritydoc_test.go"
Cohesion: 0.08
Nodes (36): Conventions, go-deploy-utility, Process, Consuming go-deploy-utility, Develop against a local checkout, Pin an exact version, Release and merge order, paritydoc, paritydoc_test.go (+28 more)

### Community 12 - "exitcode_test.go"
Cohesion: 0.15
Nodes (25): Coverage, exitcode, exitcode_test.go, Running, Summary, Tests, Tests by package, childExitError, AsUsage() (+17 more)

### Community 13 - "rootless.go"
Cohesion: 0.11
Nodes (34): rootless, rootless_test.go, CheckCore(), CheckIDMap(), CheckLinger(), covers(), Decide(), HardLimit(), LimitText() (+26 more)

### Community 14 - "echo_test.go"
Cohesion: 0.23
Nodes (15): runner, echo_test.go, TestEchoDefaultWriter(), TestEchoOutput(), TestEchoOutputInput(), TestEchoRefusesWhatExecRefuses(), TestEchoRun(), TestEchoRunEnv(), TestEchoRunEnvNoEnv() (+7 more)

### Community 16 - "execguard_test.go"
Cohesion: 0.14
Nodes (43): execguard, execguard_test.go, Approvals, Approve(), base(), Check(), checkBinary(), checkFlagShape(), checkToken() (+35 more)

### Community 17 - "envref_test.go"
Cohesion: 0.25
Nodes (15): envref, envref_test.go, env(), ref(), TestCheckBothSet(), TestCheckInvalidName(), TestCheckReferenceForm(), TestMessagesNeverCarryTheValue(), TestRefWiringErrors() (+7 more)

### Community 18 - "artifacttest.go"
Cohesion: 0.11
Nodes (48): recorder, unitLine, AssertAbsent(), AssertComposePosture(), AssertQuadletKeys(), AssertQuadletPosture(), AssertSecretTargets(), CheckAbsent() (+40 more)

### Community 19 - "compose_test.go"
Cohesion: 0.24
Nodes (16): Secret, Escape(), FlowList(), Quote(), ServiceSecrets(), goldenSecrets(), TestComposeQuoteEscapes(), TestEscape() (+8 more)

### Community 20 - "runnertest_test.go"
Cohesion: 0.18
Nodes (20): runner/runnertest, runnertest_test.go, cloneStrings(), FailOn(), Seq(), cmd(), files(), TestCallHasAndString(), TestFakeAttach() (+12 more)

### Community 21 - "containercli.go"
Cohesion: 0.10
Nodes (43): ByImage(), cmd(), composeCmd(), ComposeDown(), ComposeUp(), EngineHint(), Exec(), HasName() (+35 more)

### Community 22 - "systemd.go"
Cohesion: 0.18
Nodes (21): systemd, systemd_test.go, Active(), DaemonReload(), EnableLinger(), IsActive(), NRestarts(), ParseLinger(), ParseNRestarts() (+13 more)

### Community 23 - "watch_test.go"
Cohesion: 0.19
Nodes (17): watch, watch_test.go, Options, CheckInterval(), Loop(), section(), Sleep(), clock(), run() (+9 more)

### Community 24 - "inspect.go"
Cohesion: 0.17
Nodes (18): Decode(), health(), Printable(), TestContainersArgv(), TestDecodeDocker(), TestDecodeHealthSpellings(), TestDecodeNeverReadsEnv(), TestDecodePartial() (+10 more)

### Community 25 - "Runner"
Cohesion: 0.22
Nodes (14): 1. What this is, 5.3 Runner seams (why nothing is shared yet), 6.2 Repo B, solace-util, A.7.1 Runtime layer: docker, podman, kubectl, Stage 1: B runner seam, minimum for stdin plus env, Runner seams, Echo, MaskEnv() (+6 more)

### Community 26 - "Package layout and build order"
Cohesion: 0.25
Nodes (11): List(), Containers(), Adding a package, Dev tasks, Developer guide, Package layout and build order, Rules for shared code, CheckAttach() (+3 more)

### Community 27 - ".Run"
Cohesion: 0.46
Nodes (3): Exec, applyEnv(), Resolve()

## Knowledge Gaps
- **2 isolated node(s):** `github.com/hafio/go-deploy-utility`, `NO_COLOR`
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 85 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **4 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Package layout and build order` connect `Package layout and build order` to `podman_test.go`, `testing.T`, `noticetest_test.go`, `testing.T`, `paritydoc_test.go`, `exitcode_test.go`, `rootless.go`, `execguard_test.go`, `envref_test.go`, `artifacttest.go`, `compose_test.go`, `runnertest_test.go`, `containercli.go`, `systemd.go`, `watch_test.go`, `inspect.go`, `Runner`?**
  _High betweenness centrality (0.100) - this node is a cross-community bridge._
- **Are the 64 inferred relationships involving `Package layout and build order` (e.g. with `Prompter` and `HostNetwork()`) actually correct?**
  _`Package layout and build order` has 64 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/hafio/go-deploy-utility`, `NO_COLOR` to the rest of the system?**
  _2 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `testing.T` be split into smaller, more focused modules?**
  _Cohesion score 0.13690476190476192 - nodes in this community are weakly interconnected._
- **Why does `Echo` connect `Runner` to `podman_test.go`, `Appendix C: deferred code stages (not in this plan, separate approval)`, `testing.T`, `rootless.go`, `execguard_test.go`, `runnertest_test.go`, `containercli.go`, `systemd.go`, `inspect.go`, `Package layout and build order`?**
  _High betweenness centrality (0.064) - this node is a cross-community bridge._
- **Are the 54 inferred relationships involving `runner, runner_test.go` (e.g. with `stdin()` and `.echo()`) actually correct?**
  _`runner, runner_test.go` has 54 INFERRED edges - model-reasoned connections that need verification._
- **Should `Appendix C: deferred code stages (not in this plan, separate approval)` be split into smaller, more focused modules?**
  _Cohesion score 0.043478260869565216 - nodes in this community are weakly interconnected._