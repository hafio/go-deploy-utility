# Graph Report - go-deploy-utility  (2026-11-05)

## Corpus Check
- 67 files · ~94,247 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 3 file(s) not represented in the graph (top: (none) 3)

## Summary
- 901 nodes · 2996 edges · 32 communities (28 shown, 4 thin omitted)
- Extraction: 73% EXTRACTED · 27% INFERRED · 0% AMBIGUOUS · INFERRED: 813 edges (avg confidence: 0.9)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `71b78fd3`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- dev.sh
- podman_test.go
- testing.T
- Appendix A: research notes (licensing and shared module)
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
- Runner
- containercli.go
- systemd.go
- compose_test.go
- inspect.go
- Appendix C: deferred code stages (not in this plan, separate approval)
- Developer guide
- TestEveryDocCarriesTheSupportNotice
- Tests by package
- go-deploy-utility
- Consuming go-deploy-utility
- PLAN.md

## God Nodes (most connected - your core abstractions)
1. `Package layout and build order` - 82 edges
2. `Cmd` - 64 edges
3. `runner, runner_test.go` - 54 edges
4. `Runner` - 46 edges
5. `Echo` - 40 edges
6. `helperArgv()` - 36 edges
7. `helperOn()` - 36 edges
8. `Check()` - 30 edges
9. `execguard, execguard_test.go` - 28 edges
10. `Notice` - 27 edges

## Surprising Connections (you probably didn't know these)
- `root, docs_test.go` --references--> `TestEveryDocCarriesTheSupportNotice()`  [INFERRED]
  docs/test.md → docs_test.go
- `Stage 2: B compose on stdin, like A` --references--> `Delete()`  [INFERRED]
  docs/handoff/PLAN.md → kube/kube.go
- `Pin an exact version` --references--> `Check()`  [INFERRED]
  docs/consuming.md → paritydoc/paritydoc.go
- `Package layout and build order` --references--> `Prompter`  [INFERRED]
  docs/developer.md → confirm/prompter.go
- `Package layout and build order` --references--> `Widening()`  [INFERRED]
  docs/developer.md → containerart/artifacttest/artifacttest.go

## Import Cycles
- None detected.

## Communities (32 total, 4 thin omitted)

### Community 0 - "dev.sh"
Cohesion: 0.19
Nodes (19): c(), expand(), finish(), log_begin(), NO_COLOR, run(), dev.sh script, die() (+11 more)

### Community 1 - "podman_test.go"
Cohesion: 0.21
Nodes (22): podman, podman_test.go, cmd(), CreateSecret(), IDMap(), QuadletDir(), RemoveSecrets(), SupportsReplace(), env() (+14 more)

### Community 2 - "testing.T"
Cohesion: 0.10
Nodes (76): runner, echo_test.go, runner, runner_test.go, TestEchoDefaultWriter(), TestEchoOutput(), TestEchoOutputInput(), TestEchoRefusesWhatExecRefuses(), TestEchoRun(), TestEchoRunEnv() (+68 more)

### Community 3 - "Appendix A: research notes (licensing and shared module)"
Cohesion: 0.17
Nodes (10): A.1 Facts that shape the answer, A.2 Enforcement options, weakest to strongest, A.3 Key format and verification (options 1 and 2), A.4 Where the key comes from, A.5 Where the gate plugs in, A.6 Shared module, A.7 Code the two repos could share (synergy), A.8 Shared module: layout and GitHub mechanics (+2 more)

### Community 4 - "dev.ps1"
Cohesion: 0.22
Nodes (12): Get-Log(), Get-Now(), Invoke-Logged(), Task-build(), Task-cov(), Task-graphify(), Task-scan(), Task-test() (+4 more)

### Community 5 - "noticetest_test.go"
Cohesion: 0.10
Nodes (47): notice, notice_test.go, notice/noticetest, noticetest_test.go, UnsetError, checkCovers(), checkKind(), checkName(), Notice, New() (+39 more)

### Community 6 - "prompter_test.go"
Cohesion: 0.08
Nodes (35): countingReader, Form, Line(), TestLineReusesBufferedReader(), TestPromptYes(), TestPromptYesNo(), Yes(), YesNo() (+27 more)

### Community 7 - "abbrev_test.go"
Cohesion: 0.15
Nodes (21): checkShort(), New(), mustPanic(), sample(), TestAccessorsReturnCopies(), TestDigitsAreAllowed(), TestEntriesCloneKeepsNilShort(), TestExpandResolvesBothSpellings() (+13 more)

### Community 8 - "paritydoc_test.go"
Cohesion: 0.14
Nodes (26): Conventions, go-deploy-utility, Process, paritydoc, paritydoc_test.go, Contract, Decisions, Divergences, Inherent differences (+18 more)

### Community 12 - "exitcode_test.go"
Cohesion: 0.18
Nodes (23): exitcode, exitcode_test.go, childExitError, AsUsage(), ChildExit(), ChildStatus(), Code(), Exited(), IsUsage() (+15 more)

### Community 13 - "rootless.go"
Cohesion: 0.11
Nodes (34): rootless, rootless_test.go, CheckCore(), CheckIDMap(), CheckLinger(), covers(), Decide(), HardLimit(), LimitText() (+26 more)

### Community 14 - "runnertest_test.go"
Cohesion: 0.17
Nodes (20): runner/runnertest, runnertest_test.go, cloneStrings(), FailOn(), Seq(), cmd(), files(), TestCallHasAndString(), TestFakeAttach() (+12 more)

### Community 16 - "execguard_test.go"
Cohesion: 0.14
Nodes (43): execguard, execguard_test.go, Approvals, Approve(), base(), Check(), checkBinary(), checkFlagShape(), checkToken() (+35 more)

### Community 17 - "envref_test.go"
Cohesion: 0.25
Nodes (15): envref, envref_test.go, env(), ref(), TestCheckBothSet(), TestCheckInvalidName(), TestCheckReferenceForm(), TestMessagesNeverCarryTheValue(), TestRefWiringErrors() (+7 more)

### Community 18 - "artifacttest.go"
Cohesion: 0.11
Nodes (48): recorder, unitLine, AssertAbsent(), AssertComposePosture(), AssertQuadletKeys(), AssertQuadletPosture(), AssertSecretTargets(), CheckAbsent() (+40 more)

### Community 19 - "nsgate.go"
Cohesion: 0.10
Nodes (39): kube/nsgate, nsgate_test.go, kube/rbac, rbac_test.go, DeleteProbe(), discounted(), List(), Occupants(), ownedBy(), Protected() (+31 more)

### Community 20 - "Runner"
Cohesion: 0.05
Nodes (52): 1. What this is, 2. State at export, 4. Decisions already taken, 5.1 Same in both (the contract), 5.2 Different by policy, 5.3 Runner seams (why nothing is shared yet), 5. Exact similarities and differences, 6.1 Repo A, solmq-conn-util (+44 more)

### Community 21 - "containercli.go"
Cohesion: 0.06
Nodes (85): ByImage(), cmd(), composeCmd(), ComposeDown(), ComposeUp(), EngineHint(), Exec(), HasName() (+77 more)

### Community 22 - "systemd.go"
Cohesion: 0.18
Nodes (21): systemd, systemd_test.go, Active(), DaemonReload(), EnableLinger(), IsActive(), NRestarts(), ParseLinger(), ParseNRestarts() (+13 more)

### Community 23 - "compose_test.go"
Cohesion: 0.24
Nodes (16): Secret, Escape(), FlowList(), Quote(), ServiceSecrets(), goldenSecrets(), TestComposeQuoteEscapes(), TestEscape() (+8 more)

### Community 24 - "inspect.go"
Cohesion: 0.15
Nodes (19): Containers(), Decode(), health(), Printable(), TestContainersArgv(), TestDecodeDocker(), TestDecodeHealthSpellings(), TestDecodeNeverReadsEnv() (+11 more)

### Community 25 - "Appendix C: deferred code stages (not in this plan, separate approval)"
Cohesion: 0.22
Nodes (9): Appendix C: deferred code stages (not in this plan, separate approval), Open decisions (recorded in parity.md, not settled here), Stage 2: B compose on stdin, like A, Stage 3: B non-TTY confirmations fail fast, Stage 4: A podman session and host parity, Stage 5: A kubernetes parity, Stage 6: podman secrets, both, Stage 7: design spike, root deploys for a declared rootless user (both) (+1 more)

### Community 26 - "Developer guide"
Cohesion: 0.32
Nodes (4): Adding a package, Dev tasks, Developer guide, Rules for shared code

### Community 27 - "TestEveryDocCarriesTheSupportNotice"
Cohesion: 0.40
Nodes (6): 3. House rules that bind the next session, Files, parity.md content, Stage 0: capture (both repos, docs only), Verification, TestEveryDocCarriesTheSupportNotice()

### Community 28 - "Tests by package"
Cohesion: 0.33
Nodes (6): Coverage, root, docs_test.go, Running, Summary, Tests, Tests by package

### Community 29 - "go-deploy-utility"
Cohesion: 0.33
Nodes (6): Development, go-deploy-utility, Licence, Packages, Status, Using it

### Community 30 - "Consuming go-deploy-utility"
Cohesion: 0.50
Nodes (4): Consuming go-deploy-utility, Develop against a local checkout, Pin an exact version, Release and merge order

### Community 31 - "PLAN.md"
Cohesion: 0.50
Nodes (3): Align docker, podman and kubernetes behaviour across solmq-conn-util and solace-util, Appendix B: B tests and doc lines pinned to the on-disk compose file (Stage 2 input), Context

## Knowledge Gaps
- **2 isolated node(s):** `github.com/hafio/go-deploy-utility`, `NO_COLOR`
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 83 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **4 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Package layout and build order` connect `containercli.go` to `podman_test.go`, `testing.T`, `noticetest_test.go`, `prompter_test.go`, `paritydoc_test.go`, `exitcode_test.go`, `rootless.go`, `runnertest_test.go`, `execguard_test.go`, `envref_test.go`, `artifacttest.go`, `nsgate.go`, `Runner`, `systemd.go`, `compose_test.go`, `inspect.go`, `Developer guide`?**
  _High betweenness centrality (0.108) - this node is a cross-community bridge._
- **Are the 81 inferred relationships involving `Package layout and build order` (e.g. with `Prompter` and `.Ask()`) actually correct?**
  _`Package layout and build order` has 81 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/hafio/go-deploy-utility`, `NO_COLOR` to the rest of the system?**
  _2 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `testing.T` be split into smaller, more focused modules?**
  _Cohesion score 0.1002710027100271 - nodes in this community are weakly interconnected._
- **Why does `Echo` connect `Runner` to `podman_test.go`, `noticetest_test.go`, `prompter_test.go`, `rootless.go`, `runnertest_test.go`, `execguard_test.go`, `nsgate.go`, `containercli.go`, `systemd.go`, `inspect.go`?**
  _High betweenness centrality (0.055) - this node is a cross-community bridge._
- **Are the 4 inferred relationships involving `Cmd` (e.g. with `Conventions` and `Package layout and build order`) actually correct?**
  _`Cmd` has 4 INFERRED edges - model-reasoned connections that need verification._
- **Should `noticetest_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.09883449883449884 - nodes in this community are weakly interconnected._