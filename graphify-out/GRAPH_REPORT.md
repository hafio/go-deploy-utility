# Graph Report - go-deploy-utility  (2026-11-03)

## Corpus Check
- 41 files · ~59,953 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 3 file(s) not represented in the graph (top: (none) 3)

## Summary
- 550 nodes · 1780 edges · 19 communities (15 shown, 4 thin omitted)
- Extraction: 74% EXTRACTED · 26% INFERRED · 0% AMBIGUOUS · INFERRED: 466 edges (avg confidence: 0.91)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `28372be4`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- dev.sh
- Developer guide
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
- Runner
- runnertest_test.go
- execguard_test.go
- envref_test.go
- echo_test.go

## God Nodes (most connected - your core abstractions)
1. `runner, runner_test.go` - 55 edges
2. `Runner` - 37 edges
3. `helperArgv()` - 36 edges
4. `helperOn()` - 36 edges
5. `Echo` - 31 edges
6. `Check()` - 30 edges
7. `execguard, execguard_test.go` - 28 edges
8. `Notice` - 27 edges
9. `Package layout and build order` - 25 edges
10. `Combined()` - 23 edges

## Surprising Connections (you probably didn't know these)
- `3. House rules that bind the next session` --references--> `TestEveryDocCarriesTheSupportNotice()`  [INFERRED]
  docs/handoff/HANDOFF.md → docs_test.go
- `Pin an exact version` --references--> `Check()`  [INFERRED]
  docs/consuming.md → paritydoc/paritydoc.go
- `Adoption order` --references--> `Quote()`  [INFERRED]
  docs/developer.md → runner/echo.go
- `confirm, terminal_test.go, line_test.go, prompter_test.go` --references--> `TestPromptYesNo()`  [INFERRED]
  docs/test.md → confirm/line_test.go
- `confirm, terminal_test.go, line_test.go, prompter_test.go` --references--> `TestPromptYes()`  [INFERRED]
  docs/test.md → confirm/line_test.go

## Import Cycles
- None detected.

## Communities (19 total, 4 thin omitted)

### Community 0 - "dev.sh"
Cohesion: 0.19
Nodes (19): c(), expand(), finish(), log_begin(), NO_COLOR, run(), dev.sh script, die() (+11 more)

### Community 1 - "Developer guide"
Cohesion: 0.12
Nodes (15): Consuming go-deploy-utility, Develop against a local checkout, Pin an exact version, Release and merge order, Adding a package, Adoption order, Dev tasks, Developer guide (+7 more)

### Community 2 - "testing.T"
Cohesion: 0.12
Nodes (63): runner, runner_test.go, classify(), Combined(), Output(), Resolve(), Split(), attachFiles(), captureStdout() (+55 more)

### Community 3 - "Appendix C: deferred code stages (not in this plan, separate approval)"
Cohesion: 0.08
Nodes (22): A.1 Facts that shape the answer, A.2 Enforcement options, weakest to strongest, A.3 Key format and verification (options 1 and 2), A.4 Where the key comes from, A.5 Where the gate plugs in, A.6 Shared module, A.7 Code the two repos could share (synergy), A.8 Shared module: layout and GitHub mechanics (+14 more)

### Community 4 - "dev.ps1"
Cohesion: 0.22
Nodes (12): Get-Log(), Get-Now(), Invoke-Logged(), Task-build(), Task-cov(), Task-graphify(), Task-scan(), Task-test() (+4 more)

### Community 5 - "noticetest_test.go"
Cohesion: 0.09
Nodes (52): Files, parity.md content, Stage 0: capture (both repos, docs only), Verification, notice, notice_test.go, notice/noticetest, noticetest_test.go, root, docs_test.go, TestEveryDocCarriesTheSupportNotice() (+44 more)

### Community 6 - "testing.T"
Cohesion: 0.12
Nodes (25): countingReader, IsNotTerminal(), New(), TestAskNonTerminalNeverWrites(), TestAskUnansweredIsNotAnError(), TestGateCustomNoPromptFlagAndDoing(), TestGateNonTerminalFailsNamingTheFlag(), TestGateNoPromptSkipsEverything() (+17 more)

### Community 7 - "abbrev_test.go"
Cohesion: 0.14
Nodes (21): checkShort(), New(), mustPanic(), sample(), TestAccessorsReturnCopies(), TestDigitsAreAllowed(), TestEntriesCloneKeepsNilShort(), TestExpandResolvesBothSpellings() (+13 more)

### Community 8 - "paritydoc_test.go"
Cohesion: 0.14
Nodes (26): Conventions, go-deploy-utility, Process, paritydoc, paritydoc_test.go, Contract, Decisions, Divergences, Inherent differences (+18 more)

### Community 12 - "exitcode_test.go"
Cohesion: 0.16
Nodes (25): Coverage, exitcode, exitcode_test.go, Running, Summary, Tests, Tests by package, childExitError, AsUsage() (+17 more)

### Community 13 - "Runner"
Cohesion: 0.06
Nodes (38): Form, Line(), TestLineReusesBufferedReader(), TestPromptYes(), TestPromptYesNo(), Yes(), YesNo(), NotTerminalError (+30 more)

### Community 14 - "runnertest_test.go"
Cohesion: 0.13
Nodes (24): runner/runnertest, runnertest_test.go, applyEnv(), CheckAttach(), CheckRun(), Cmd, cloneStrings(), FailOn(), Seq() (+16 more)

### Community 16 - "execguard_test.go"
Cohesion: 0.13
Nodes (44): Package layout and build order, execguard, execguard_test.go, Approvals, Approve(), base(), Check(), checkBinary(), checkFlagShape() (+36 more)

### Community 17 - "envref_test.go"
Cohesion: 0.21
Nodes (16): envref, envref_test.go, env(), ref(), TestCheckBothSet(), TestCheckInvalidName(), TestCheckReferenceForm(), TestMessagesNeverCarryTheValue(), TestRefWiringErrors() (+8 more)

### Community 18 - "echo_test.go"
Cohesion: 0.21
Nodes (15): runner, echo_test.go, TestEchoDefaultWriter(), TestEchoOutput(), TestEchoOutputInput(), TestEchoRefusesWhatExecRefuses(), TestEchoRun(), TestEchoRunEnv(), TestEchoRunEnvNoEnv() (+7 more)

## Knowledge Gaps
- **2 isolated node(s):** `github.com/hafio/go-deploy-utility`, `NO_COLOR`
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 79 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **4 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Echo` connect `Runner` to `execguard_test.go`, `runnertest_test.go`, `abbrev_test.go`?**
  _High betweenness centrality (0.078) - this node is a cross-community bridge._
- **Are the 54 inferred relationships involving `runner, runner_test.go` (e.g. with `env()` and `.echo()`) actually correct?**
  _`runner, runner_test.go` has 54 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/hafio/go-deploy-utility`, `NO_COLOR` to the rest of the system?**
  _2 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Developer guide` be split into smaller, more focused modules?**
  _Cohesion score 0.11695906432748537 - nodes in this community are weakly interconnected._
- **Why does `Runner` connect `Runner` to `Developer guide`, `testing.T`, `testing.T`, `abbrev_test.go`, `exitcode_test.go`, `runnertest_test.go`, `execguard_test.go`, `echo_test.go`?**
  _High betweenness centrality (0.069) - this node is a cross-community bridge._
- **Are the 9 inferred relationships involving `Runner` (e.g. with `Package layout and build order` and `Rules for shared code`) actually correct?**
  _`Runner` has 9 INFERRED edges - model-reasoned connections that need verification._
- **Should `testing.T` be split into smaller, more focused modules?**
  _Cohesion score 0.12422360248447205 - nodes in this community are weakly interconnected._