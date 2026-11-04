# Graph Report - go-deploy-utility  (2026-11-04)

## Corpus Check
- 47 files · ~67,259 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 3 file(s) not represented in the graph (top: (none) 3)

## Summary
- 622 nodes · 2042 edges · 20 communities (16 shown, 4 thin omitted)
- Extraction: 74% EXTRACTED · 26% INFERRED · 0% AMBIGUOUS · INFERRED: 541 edges (avg confidence: 0.91)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `dec009d5`
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
- Handoff: aligning solmq-conn-util and solace-util on Docker, Podman and Kubernetes
- Runner
- execguard_test.go
- envref_test.go
- artifacttest.go
- compose_test.go

## God Nodes (most connected - your core abstractions)
1. `runner, runner_test.go` - 55 edges
2. `Runner` - 37 edges
3. `helperArgv()` - 36 edges
4. `helperOn()` - 36 edges
5. `Package layout and build order` - 35 edges
6. `Echo` - 31 edges
7. `Check()` - 30 edges
8. `execguard, execguard_test.go` - 28 edges
9. `Notice` - 27 edges
10. `Combined()` - 23 edges

## Surprising Connections (you probably didn't know these)
- `Adoption order` --references--> `Environment()`  [INFERRED]
  docs/developer.md → containerart/quadlet/quadlet.go
- `3. House rules that bind the next session` --references--> `TestEveryDocCarriesTheSupportNotice()`  [INFERRED]
  docs/handoff/HANDOFF.md → docs_test.go
- `root, docs_test.go` --references--> `TestEveryDocCarriesTheSupportNotice()`  [INFERRED]
  docs/test.md → docs_test.go
- `Pin an exact version` --references--> `Check()`  [INFERRED]
  docs/consuming.md → paritydoc/paritydoc.go
- `Package layout and build order` --references--> `Prompter`  [INFERRED]
  docs/developer.md → confirm/prompter.go

## Import Cycles
- None detected.

## Communities (20 total, 4 thin omitted)

### Community 0 - "dev.sh"
Cohesion: 0.19
Nodes (19): c(), expand(), finish(), log_begin(), NO_COLOR, run(), dev.sh script, die() (+11 more)

### Community 1 - "Developer guide"
Cohesion: 0.09
Nodes (19): Consuming go-deploy-utility, Develop against a local checkout, Pin an exact version, Release and merge order, Adding a package, Adoption order, Dev tasks, Developer guide (+11 more)

### Community 2 - "testing.T"
Cohesion: 0.14
Nodes (59): runner, runner_test.go, Combined(), attachFiles(), captureStdout(), helperArgv(), helperOn(), readFile(), regularFile() (+51 more)

### Community 3 - "Appendix C: deferred code stages (not in this plan, separate approval)"
Cohesion: 0.08
Nodes (22): A.1 Facts that shape the answer, A.2 Enforcement options, weakest to strongest, A.3 Key format and verification (options 1 and 2), A.4 Where the key comes from, A.5 Where the gate plugs in, A.6 Shared module, A.7 Code the two repos could share (synergy), A.8 Shared module: layout and GitHub mechanics (+14 more)

### Community 4 - "dev.ps1"
Cohesion: 0.22
Nodes (12): Get-Log(), Get-Now(), Invoke-Logged(), Task-build(), Task-cov(), Task-graphify(), Task-scan(), Task-test() (+4 more)

### Community 5 - "noticetest_test.go"
Cohesion: 0.09
Nodes (51): Files, parity.md content, Stage 0: capture (both repos, docs only), Verification, notice, notice_test.go, notice/noticetest, noticetest_test.go, TestEveryDocCarriesTheSupportNotice(), checkCovers() (+43 more)

### Community 6 - "testing.T"
Cohesion: 0.08
Nodes (35): countingReader, Form, Line(), TestLineReusesBufferedReader(), TestPromptYes(), TestPromptYesNo(), Yes(), YesNo() (+27 more)

### Community 7 - "abbrev_test.go"
Cohesion: 0.15
Nodes (21): checkShort(), New(), mustPanic(), sample(), TestAccessorsReturnCopies(), TestDigitsAreAllowed(), TestEntriesCloneKeepsNilShort(), TestExpandResolvesBothSpellings() (+13 more)

### Community 8 - "paritydoc_test.go"
Cohesion: 0.13
Nodes (26): Conventions, go-deploy-utility, Process, paritydoc, paritydoc_test.go, Contract, Decisions, Divergences, Inherent differences (+18 more)

### Community 12 - "exitcode_test.go"
Cohesion: 0.17
Nodes (21): exitcode, exitcode_test.go, Summary, childExitError, AsUsage(), ChildExit(), ChildStatus(), Code(), IsUsage() (+13 more)

### Community 13 - "Handoff: aligning solmq-conn-util and solace-util on Docker, Podman and Kubernetes"
Cohesion: 0.13
Nodes (14): 2. State at export, 3. House rules that bind the next session, 4. Decisions already taken, 5.1 Same in both (the contract), 5.2 Different by policy, 5. Exact similarities and differences, 6.1 Repo A, solmq-conn-util, 6.3 Not scanned or not verified (+6 more)

### Community 14 - "Runner"
Cohesion: 0.07
Nodes (59): Package layout and build order, Rules for shared code, 1. What this is, 5.3 Runner seams (why nothing is shared yet), 6.2 Repo B, solace-util, 7. Next stages, A.7.1 Runtime layer: docker, podman, kubectl, Stage 1: B runner seam, minimum for stdin plus env (+51 more)

### Community 16 - "execguard_test.go"
Cohesion: 0.13
Nodes (43): execguard, execguard_test.go, Approvals, Approve(), base(), Check(), checkBinary(), checkFlagShape(), checkToken() (+35 more)

### Community 17 - "envref_test.go"
Cohesion: 0.21
Nodes (16): envref, envref_test.go, env(), ref(), TestCheckBothSet(), TestCheckInvalidName(), TestCheckReferenceForm(), TestMessagesNeverCarryTheValue(), TestRefWiringErrors() (+8 more)

### Community 18 - "artifacttest.go"
Cohesion: 0.11
Nodes (48): recorder, unitLine, AssertAbsent(), AssertComposePosture(), AssertQuadletKeys(), AssertQuadletPosture(), AssertSecretTargets(), CheckAbsent() (+40 more)

### Community 19 - "compose_test.go"
Cohesion: 0.24
Nodes (16): Secret, Escape(), FlowList(), Quote(), ServiceSecrets(), goldenSecrets(), TestComposeQuoteEscapes(), TestEscape() (+8 more)

## Knowledge Gaps
- **2 isolated node(s):** `github.com/hafio/go-deploy-utility`, `NO_COLOR`
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 80 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **4 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Package layout and build order` connect `Runner` to `Developer guide`, `testing.T`, `noticetest_test.go`, `testing.T`, `paritydoc_test.go`, `exitcode_test.go`, `execguard_test.go`, `envref_test.go`, `artifacttest.go`, `compose_test.go`?**
  _High betweenness centrality (0.086) - this node is a cross-community bridge._
- **Are the 54 inferred relationships involving `runner, runner_test.go` (e.g. with `env()` and `.echo()`) actually correct?**
  _`runner, runner_test.go` has 54 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/hafio/go-deploy-utility`, `NO_COLOR` to the rest of the system?**
  _2 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Developer guide` be split into smaller, more focused modules?**
  _Cohesion score 0.09486166007905138 - nodes in this community are weakly interconnected._
- **Why does `Echo` connect `Runner` to `execguard_test.go`, `artifacttest.go`, `exitcode_test.go`, `noticetest_test.go`?**
  _High betweenness centrality (0.068) - this node is a cross-community bridge._
- **Are the 9 inferred relationships involving `Runner` (e.g. with `Package layout and build order` and `Rules for shared code`) actually correct?**
  _`Runner` has 9 INFERRED edges - model-reasoned connections that need verification._
- **Should `testing.T` be split into smaller, more focused modules?**
  _Cohesion score 0.1382488479262673 - nodes in this community are weakly interconnected._