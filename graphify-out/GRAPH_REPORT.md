# Graph Report - go-deploy-utility  (2026-11-04)

## Corpus Check
- 53 files · ~75,909 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 3 file(s) not represented in the graph (top: (none) 3)

## Summary
- 709 nodes · 2347 edges · 21 communities (17 shown, 4 thin omitted)
- Extraction: 72% EXTRACTED · 28% INFERRED · 0% AMBIGUOUS · INFERRED: 649 edges (avg confidence: 0.91)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `e013f92e`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- dev.sh
- Package layout and build order
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
- Echo
- execguard_test.go
- envref_test.go
- artifacttest.go
- compose_test.go
- runnertest_test.go

## God Nodes (most connected - your core abstractions)
1. `Package layout and build order` - 55 edges
2. `runner, runner_test.go` - 54 edges
3. `Runner` - 41 edges
4. `helperArgv()` - 36 edges
5. `helperOn()` - 36 edges
6. `Echo` - 35 edges
7. `Cmd` - 35 edges
8. `Check()` - 30 edges
9. `execguard, execguard_test.go` - 28 edges
10. `Notice` - 27 edges

## Surprising Connections (you probably didn't know these)
- `Package layout and build order` --references--> `Prompter`  [INFERRED]
  docs/developer.md → confirm/prompter.go
- `Package layout and build order` --references--> `Widening()`  [INFERRED]
  docs/developer.md → containerart/artifacttest/artifacttest.go
- `Package layout and build order` --references--> `HostNetwork()`  [INFERRED]
  docs/developer.md → containerart/artifacttest/artifacttest.go
- `Package layout and build order` --references--> `quadletKeySince()`  [INFERRED]
  docs/developer.md → containerart/artifacttest/artifacttest.go
- `Package layout and build order` --references--> `report()`  [INFERRED]
  docs/developer.md → containerart/artifacttest/artifacttest.go

## Import Cycles
- None detected.

## Communities (21 total, 4 thin omitted)

### Community 0 - "dev.sh"
Cohesion: 0.19
Nodes (19): c(), expand(), finish(), log_begin(), NO_COLOR, run(), dev.sh script, die() (+11 more)

### Community 1 - "Package layout and build order"
Cohesion: 0.07
Nodes (59): Adding a package, Adoption order, Dev tasks, Developer guide, Package layout and build order, Rules for shared code, 6.2 Repo B, solace-util, Stage 2: B compose on stdin, like A (+51 more)

### Community 2 - "testing.T"
Cohesion: 0.13
Nodes (61): runner, runner_test.go, Combined(), Output(), Split(), attachFiles(), captureStdout(), helperArgv(), helperOn() (+53 more)

### Community 3 - "Appendix C: deferred code stages (not in this plan, separate approval)"
Cohesion: 0.10
Nodes (19): A.1 Facts that shape the answer, A.2 Enforcement options, weakest to strongest, A.3 Key format and verification (options 1 and 2), A.4 Where the key comes from, A.5 Where the gate plugs in, A.6 Shared module, A.8 Shared module: layout and GitHub mechanics, Align docker, podman and kubernetes behaviour across solmq-conn-util and solace-util (+11 more)

### Community 4 - "dev.ps1"
Cohesion: 0.22
Nodes (12): Get-Log(), Get-Now(), Invoke-Logged(), Task-build(), Task-cov(), Task-graphify(), Task-scan(), Task-test() (+4 more)

### Community 5 - "noticetest_test.go"
Cohesion: 0.09
Nodes (53): 3. House rules that bind the next session, Files, parity.md content, Stage 0: capture (both repos, docs only), Verification, notice, notice_test.go, notice/noticetest, noticetest_test.go, root, docs_test.go (+45 more)

### Community 6 - "testing.T"
Cohesion: 0.08
Nodes (35): countingReader, Form, Line(), TestLineReusesBufferedReader(), TestPromptYes(), TestPromptYesNo(), Yes(), YesNo() (+27 more)

### Community 7 - "abbrev_test.go"
Cohesion: 0.15
Nodes (21): checkShort(), New(), mustPanic(), sample(), TestAccessorsReturnCopies(), TestDigitsAreAllowed(), TestEntriesCloneKeepsNilShort(), TestExpandResolvesBothSpellings() (+13 more)

### Community 8 - "paritydoc_test.go"
Cohesion: 0.08
Nodes (37): Conventions, go-deploy-utility, Process, Consuming go-deploy-utility, Develop against a local checkout, Pin an exact version, Release and merge order, Coverage (+29 more)

### Community 12 - "exitcode_test.go"
Cohesion: 0.19
Nodes (20): exitcode, exitcode_test.go, childExitError, AsUsage(), ChildExit(), ChildStatus(), Code(), IsUsage(), childStatusError() (+12 more)

### Community 13 - "rootless.go"
Cohesion: 0.11
Nodes (34): rootless, rootless_test.go, CheckCore(), CheckIDMap(), CheckLinger(), covers(), Decide(), HardLimit(), LimitText() (+26 more)

### Community 14 - "Echo"
Cohesion: 0.05
Nodes (45): 1. What this is, 2. State at export, 4. Decisions already taken, 5.1 Same in both (the contract), 5.2 Different by policy, 5.3 Runner seams (why nothing is shared yet), 5. Exact similarities and differences, 6.1 Repo A, solmq-conn-util (+37 more)

### Community 16 - "execguard_test.go"
Cohesion: 0.14
Nodes (43): execguard, execguard_test.go, Approvals, Approve(), base(), Check(), checkBinary(), checkFlagShape(), checkToken() (+35 more)

### Community 17 - "envref_test.go"
Cohesion: 0.21
Nodes (16): envref, envref_test.go, env(), ref(), TestCheckBothSet(), TestCheckInvalidName(), TestCheckReferenceForm(), TestMessagesNeverCarryTheValue(), TestRefWiringErrors() (+8 more)

### Community 18 - "artifacttest.go"
Cohesion: 0.11
Nodes (48): recorder, unitLine, AssertAbsent(), AssertComposePosture(), AssertQuadletKeys(), AssertQuadletPosture(), AssertSecretTargets(), CheckAbsent() (+40 more)

### Community 19 - "compose_test.go"
Cohesion: 0.22
Nodes (16): Secret, Escape(), FlowList(), Quote(), ServiceSecrets(), goldenSecrets(), TestComposeQuoteEscapes(), TestEscape() (+8 more)

### Community 20 - "runnertest_test.go"
Cohesion: 0.19
Nodes (20): runner/runnertest, runnertest_test.go, cloneStrings(), FailOn(), Seq(), cmd(), files(), TestCallHasAndString(), TestFakeAttach() (+12 more)

## Knowledge Gaps
- **2 isolated node(s):** `github.com/hafio/go-deploy-utility`, `NO_COLOR`
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 80 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **4 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Package layout and build order` connect `Package layout and build order` to `testing.T`, `noticetest_test.go`, `testing.T`, `paritydoc_test.go`, `exitcode_test.go`, `rootless.go`, `Echo`, `execguard_test.go`, `envref_test.go`, `artifacttest.go`, `compose_test.go`, `runnertest_test.go`?**
  _High betweenness centrality (0.099) - this node is a cross-community bridge._
- **Are the 54 inferred relationships involving `Package layout and build order` (e.g. with `Prompter` and `HostNetwork()`) actually correct?**
  _`Package layout and build order` has 54 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/hafio/go-deploy-utility`, `NO_COLOR` to the rest of the system?**
  _2 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Package layout and build order` be split into smaller, more focused modules?**
  _Cohesion score 0.07063063063063063 - nodes in this community are weakly interconnected._
- **Why does `Echo` connect `Echo` to `Package layout and build order`, `noticetest_test.go`, `testing.T`, `exitcode_test.go`, `rootless.go`, `execguard_test.go`, `compose_test.go`, `runnertest_test.go`?**
  _High betweenness centrality (0.059) - this node is a cross-community bridge._
- **Are the 53 inferred relationships involving `runner, runner_test.go` (e.g. with `.echo()` and `Runner`) actually correct?**
  _`runner, runner_test.go` has 53 INFERRED edges - model-reasoned connections that need verification._
- **Should `testing.T` be split into smaller, more focused modules?**
  _Cohesion score 0.13116236996834013 - nodes in this community are weakly interconnected._