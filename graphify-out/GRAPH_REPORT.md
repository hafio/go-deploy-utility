# Graph Report - go-deploy-utility  (2026-11-06)

## Corpus Check
- 67 files · ~95,958 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 3 file(s) not represented in the graph (top: (none) 3)

## Summary
- 906 nodes · 3010 edges · 28 communities (24 shown, 4 thin omitted)
- Extraction: 73% EXTRACTED · 27% INFERRED · 0% AMBIGUOUS · INFERRED: 820 edges (avg confidence: 0.91)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `7b533773`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- dev.sh
- compose_test.go
- testing.T
- Appendix C: deferred code stages (not in this plan, separate approval)
- dev.ps1
- noticetest_test.go
- prompter_test.go
- abbrev_test.go
- paritydoc_test.go
- handoff/README.md
- github.com/hafio/go-deploy-utility
- exitcode_test.go
- rootless.go
- runnertest_test.go
- execguard_test.go
- envref_test.go
- artifacttest.go
- nsgate.go
- watch_test.go
- containercli.go
- Cmd
- runnertest.go
- inspect.go
- Runner
- Handoff: aligning solmq-conn-util and solace-util on Docker, Podman and Kubernetes
- Echo

## God Nodes (most connected - your core abstractions)
1. `Package layout and build order` - 84 edges
2. `Cmd` - 65 edges
3. `runner, runner_test.go` - 54 edges
4. `Runner` - 46 edges
5. `Echo` - 40 edges
6. `helperArgv()` - 36 edges
7. `helperOn()` - 36 edges
8. `Check()` - 30 edges
9. `execguard, execguard_test.go` - 28 edges
10. `Notice` - 27 edges

## Surprising Connections (you probably didn't know these)
- `Stage 2: B compose on stdin, like A` --references--> `Delete()`  [INFERRED]
  docs/handoff/PLAN.md → kube/kube.go
- `7. Next stages` --references--> `Echo`  [INFERRED]
  docs/handoff/HANDOFF.md → runner/echo.go
- `Package layout and build order` --references--> `Prompter`  [INFERRED]
  docs/developer.md → confirm/prompter.go
- `Package layout and build order` --references--> `Widening()`  [INFERRED]
  docs/developer.md → containerart/artifacttest/artifacttest.go
- `Package layout and build order` --references--> `HostNetwork()`  [INFERRED]
  docs/developer.md → containerart/artifacttest/artifacttest.go

## Import Cycles
- None detected.

## Communities (28 total, 4 thin omitted)

### Community 0 - "dev.sh"
Cohesion: 0.19
Nodes (19): c(), expand(), finish(), log_begin(), NO_COLOR, run(), dev.sh script, die() (+11 more)

### Community 1 - "compose_test.go"
Cohesion: 0.24
Nodes (16): Secret, Escape(), FlowList(), Quote(), ServiceSecrets(), goldenSecrets(), TestComposeQuoteEscapes(), TestEscape() (+8 more)

### Community 2 - "testing.T"
Cohesion: 0.10
Nodes (76): runner, echo_test.go, runner, runner_test.go, TestEchoDefaultWriter(), TestEchoOutput(), TestEchoOutputInput(), TestEchoRefusesWhatExecRefuses(), TestEchoRun(), TestEchoRunEnv() (+68 more)

### Community 3 - "Appendix C: deferred code stages (not in this plan, separate approval)"
Cohesion: 0.08
Nodes (22): A.1 Facts that shape the answer, A.2 Enforcement options, weakest to strongest, A.3 Key format and verification (options 1 and 2), A.4 Where the key comes from, A.5 Where the gate plugs in, A.6 Shared module, A.7 Code the two repos could share (synergy), A.8 Shared module: layout and GitHub mechanics (+14 more)

### Community 4 - "dev.ps1"
Cohesion: 0.22
Nodes (12): Get-Log(), Get-Now(), Invoke-Logged(), Task-build(), Task-cov(), Task-graphify(), Task-scan(), Task-test() (+4 more)

### Community 5 - "noticetest_test.go"
Cohesion: 0.09
Nodes (53): 3. House rules that bind the next session, Files, parity.md content, Stage 0: capture (both repos, docs only), Verification, notice, notice_test.go, notice/noticetest, noticetest_test.go, root, docs_test.go (+45 more)

### Community 6 - "prompter_test.go"
Cohesion: 0.07
Nodes (37): countingReader, Form, Line(), TestLineReusesBufferedReader(), TestPromptYes(), TestPromptYesNo(), Yes(), YesNo() (+29 more)

### Community 7 - "abbrev_test.go"
Cohesion: 0.15
Nodes (21): checkShort(), New(), mustPanic(), sample(), TestAccessorsReturnCopies(), TestDigitsAreAllowed(), TestEntriesCloneKeepsNilShort(), TestExpandResolvesBothSpellings() (+13 more)

### Community 8 - "paritydoc_test.go"
Cohesion: 0.08
Nodes (36): Conventions, go-deploy-utility, Process, Consuming go-deploy-utility, Develop against a local checkout, Pin an exact version, Release and merge order, paritydoc, paritydoc_test.go (+28 more)

### Community 12 - "exitcode_test.go"
Cohesion: 0.13
Nodes (28): Coverage, exitcode, exitcode_test.go, Running, Summary, Tests, Tests by package, childExitError, AsUsage() (+20 more)

### Community 13 - "rootless.go"
Cohesion: 0.11
Nodes (33): rootless, rootless_test.go, CheckCore(), CheckIDMap(), CheckLinger(), covers(), Decide(), HardLimit(), LimitText() (+25 more)

### Community 14 - "runnertest_test.go"
Cohesion: 0.35
Nodes (14): runner/runnertest, runnertest_test.go, cmd(), files(), TestCallHasAndString(), TestFakeAttach(), TestFakeConcurrentUse(), TestFakeCopiesItsInputs(), TestFakeFailOn() (+6 more)

### Community 16 - "execguard_test.go"
Cohesion: 0.14
Nodes (43): execguard, execguard_test.go, Approvals, Approve(), base(), Check(), checkBinary(), checkFlagShape(), checkToken() (+35 more)

### Community 17 - "envref_test.go"
Cohesion: 0.20
Nodes (16): envref, envref_test.go, env(), ref(), TestCheckBothSet(), TestCheckInvalidName(), TestCheckReferenceForm(), TestMessagesNeverCarryTheValue(), TestRefWiringErrors() (+8 more)

### Community 18 - "artifacttest.go"
Cohesion: 0.11
Nodes (48): recorder, unitLine, AssertAbsent(), AssertComposePosture(), AssertQuadletKeys(), AssertQuadletPosture(), AssertSecretTargets(), CheckAbsent() (+40 more)

### Community 19 - "nsgate.go"
Cohesion: 0.10
Nodes (39): kube/nsgate, nsgate_test.go, kube/rbac, rbac_test.go, DeleteProbe(), discounted(), List(), Occupants(), ownedBy(), Protected() (+31 more)

### Community 20 - "watch_test.go"
Cohesion: 0.20
Nodes (17): watch, watch_test.go, Options, CheckInterval(), Loop(), section(), Sleep(), clock(), run() (+9 more)

### Community 21 - "containercli.go"
Cohesion: 0.09
Nodes (45): ByImage(), cmd(), composeCmd(), ComposeDown(), ComposeUp(), EngineHint(), Exec(), HasName() (+37 more)

### Community 22 - "Cmd"
Cohesion: 0.05
Nodes (93): ComposeDigest(), Printable(), Adding a package, Adoption order, Dev tasks, Developer guide, Package layout and build order, Rules for shared code (+85 more)

### Community 23 - "runnertest.go"
Cohesion: 0.23
Nodes (6): cloneStrings(), FailOn(), Seq(), Call, Fake, Reply

### Community 24 - "inspect.go"
Cohesion: 0.17
Nodes (17): Containers(), Decode(), health(), TestContainersArgv(), TestDecodeDocker(), TestDecodeHealthSpellings(), TestDecodeNeverReadsEnv(), TestDecodePartial() (+9 more)

### Community 25 - "Runner"
Cohesion: 0.23
Nodes (9): Runner seams, Exec, Runner, applyEnv(), CheckAttach(), CheckRun(), classify(), Resolve() (+1 more)

### Community 26 - "Handoff: aligning solmq-conn-util and solace-util on Docker, Podman and Kubernetes"
Cohesion: 0.12
Nodes (15): 1. What this is, 2. State at export, 4. Decisions already taken, 5.1 Same in both (the contract), 5.2 Different by policy, 5. Exact similarities and differences, 6.1 Repo A, solmq-conn-util, 6.3 Not scanned or not verified (+7 more)

### Community 27 - "Echo"
Cohesion: 0.27
Nodes (9): 5.3 Runner seams (why nothing is shared yet), 6.2 Repo B, solace-util, A.7.1 Runtime layer: docker, podman, kubectl, Stage 1: B runner seam, minimum for stdin plus env, Echo, MaskEnv(), printable(), Quote() (+1 more)

## Knowledge Gaps
- **2 isolated node(s):** `github.com/hafio/go-deploy-utility`, `NO_COLOR`
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 85 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **4 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Package layout and build order` connect `Cmd` to `compose_test.go`, `testing.T`, `noticetest_test.go`, `prompter_test.go`, `paritydoc_test.go`, `exitcode_test.go`, `rootless.go`, `execguard_test.go`, `envref_test.go`, `artifacttest.go`, `nsgate.go`, `watch_test.go`, `containercli.go`, `runnertest.go`, `inspect.go`, `Runner`, `Echo`?**
  _High betweenness centrality (0.110) - this node is a cross-community bridge._
- **Are the 83 inferred relationships involving `Package layout and build order` (e.g. with `Prompter` and `.Ask()`) actually correct?**
  _`Package layout and build order` has 83 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/hafio/go-deploy-utility`, `NO_COLOR` to the rest of the system?**
  _2 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `testing.T` be split into smaller, more focused modules?**
  _Cohesion score 0.1002710027100271 - nodes in this community are weakly interconnected._
- **Why does `Echo` connect `Echo` to `noticetest_test.go`, `prompter_test.go`, `rootless.go`, `runnertest_test.go`, `execguard_test.go`, `nsgate.go`, `containercli.go`, `Cmd`, `runnertest.go`, `inspect.go`, `Runner`, `Handoff: aligning solmq-conn-util and solace-util on Docker, Podman and Kubernetes`?**
  _High betweenness centrality (0.055) - this node is a cross-community bridge._
- **Are the 4 inferred relationships involving `Cmd` (e.g. with `Conventions` and `Package layout and build order`) actually correct?**
  _`Cmd` has 4 INFERRED edges - model-reasoned connections that need verification._
- **Should `Appendix C: deferred code stages (not in this plan, separate approval)` be split into smaller, more focused modules?**
  _Cohesion score 0.08 - nodes in this community are weakly interconnected._