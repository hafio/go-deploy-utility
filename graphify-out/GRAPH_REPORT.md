# Graph Report - go-deploy-utility  (2026-11-02)

## Corpus Check
- 35 files · ~46,542 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 3 file(s) not represented in the graph (top: (none) 3)

## Summary
- 409 nodes · 1112 edges · 18 communities (14 shown, 4 thin omitted)
- Extraction: 75% EXTRACTED · 25% INFERRED · 0% AMBIGUOUS · INFERRED: 278 edges (avg confidence: 0.9)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `3d0c0e1f`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- dev.sh
- Developer guide
- Handoff: aligning solmq-conn-util and solace-util on Docker, Podman and Kubernetes
- Appendix C: deferred code stages (not in this plan, separate approval)
- dev.ps1
- noticetest_test.go
- testing.T
- abbrev_test.go
- paritydoc_test.go
- handoff/README.md
- github.com/hafio/go-deploy-utility
- exitcode_test.go
- Prompter
- Notice
- execguard_test.go
- envref_test.go

## God Nodes (most connected - your core abstractions)
1. `Check()` - 30 edges
2. `execguard, execguard_test.go` - 28 edges
3. `Notice` - 27 edges
4. `Prompter` - 22 edges
5. `confirm, terminal_test.go, line_test.go, prompter_test.go` - 21 edges
6. `clusterRules()` - 18 edges
7. `testNotice()` - 18 edges
8. `Approve()` - 16 edges
9. `exitcode, exitcode_test.go` - 15 edges
10. `abbrev, abbrev_test.go` - 15 edges

## Surprising Connections (you probably didn't know these)
- `6.2 Repo B, solace-util` --references--> `Compose()`  [INFERRED]
  docs/handoff/HANDOFF.md → paritydoc/paritydoc.go
- `Pin an exact version` --references--> `Check()`  [INFERRED]
  docs/consuming.md → paritydoc/paritydoc.go
- `Package layout and build order` --references--> `Prompter`  [INFERRED]
  docs/developer.md → confirm/prompter.go
- `3. House rules that bind the next session` --references--> `TestEveryDocCarriesTheSupportNotice()`  [INFERRED]
  docs/handoff/HANDOFF.md → docs_test.go
- `Files` --references--> `TestEveryDocCarriesTheSupportNotice()`  [INFERRED]
  docs/handoff/PLAN.md → docs_test.go

## Import Cycles
- None detected.

## Communities (18 total, 4 thin omitted)

### Community 0 - "dev.sh"
Cohesion: 0.19
Nodes (19): c(), expand(), finish(), log_begin(), NO_COLOR, run(), dev.sh script, die() (+11 more)

### Community 1 - "Developer guide"
Cohesion: 0.12
Nodes (15): Consuming go-deploy-utility, Develop against a local checkout, Pin an exact version, Release and merge order, Adding a package, Adoption order, Dev tasks, Developer guide (+7 more)

### Community 2 - "Handoff: aligning solmq-conn-util and solace-util on Docker, Podman and Kubernetes"
Cohesion: 0.07
Nodes (29): 1. What this is, 2. State at export, 3. House rules that bind the next session, 4. Decisions already taken, 5.1 Same in both (the contract), 5.2 Different by policy, 5.3 Runner seams (why nothing is shared yet), 5. Exact similarities and differences (+21 more)

### Community 3 - "Appendix C: deferred code stages (not in this plan, separate approval)"
Cohesion: 0.07
Nodes (24): A.1 Facts that shape the answer, A.2 Enforcement options, weakest to strongest, A.3 Key format and verification (options 1 and 2), A.4 Where the key comes from, A.5 Where the gate plugs in, A.6 Shared module, A.7.1 Runtime layer: docker, podman, kubectl, A.7 Code the two repos could share (synergy) (+16 more)

### Community 4 - "dev.ps1"
Cohesion: 0.22
Nodes (12): Get-Log(), Get-Now(), Invoke-Logged(), Task-build(), Task-cov(), Task-graphify(), Task-scan(), Task-test() (+4 more)

### Community 5 - "noticetest_test.go"
Cohesion: 0.14
Nodes (30): checkShort(), notice/noticetest, noticetest_test.go, Assert(), AssertASCII(), AssertDocs(), Check(), CheckDocs(), describe() (+22 more)

### Community 6 - "testing.T"
Cohesion: 0.15
Nodes (26): TestLineReusesBufferedReader(), TestPromptYes(), TestPromptYesNo(), IsNotTerminal(), New(), TestAskNonTerminalNeverWrites(), TestAskUnansweredIsNotAnError(), TestGateCustomNoPromptFlagAndDoing() (+18 more)

### Community 7 - "abbrev_test.go"
Cohesion: 0.16
Nodes (20): New(), mustPanic(), sample(), TestAccessorsReturnCopies(), TestDigitsAreAllowed(), TestEntriesCloneKeepsNilShort(), TestExpandResolvesBothSpellings(), TestKind() (+12 more)

### Community 8 - "paritydoc_test.go"
Cohesion: 0.13
Nodes (27): Conventions, go-deploy-utility, Process, paritydoc, paritydoc_test.go, Contract, Decisions, Divergences, Inherent differences (+19 more)

### Community 12 - "exitcode_test.go"
Cohesion: 0.18
Nodes (20): exitcode, exitcode_test.go, childExitError, AsUsage(), ChildExit(), ChildStatus(), Code(), IsUsage(), childStatusError() (+12 more)

### Community 13 - "Prompter"
Cohesion: 0.17
Nodes (9): countingReader, Form, Line(), Yes(), YesNo(), NotTerminalError, Options, Prompter (+1 more)

### Community 14 - "Notice"
Cohesion: 0.25
Nodes (17): notice, notice_test.go, checkCovers(), checkKind(), checkName(), Notice, New(), Short(), mustNew() (+9 more)

### Community 16 - "execguard_test.go"
Cohesion: 0.13
Nodes (44): Package layout and build order, execguard, execguard_test.go, Approvals, Approve(), base(), Check(), checkBinary(), checkFlagShape() (+36 more)

### Community 17 - "envref_test.go"
Cohesion: 0.21
Nodes (16): envref, envref_test.go, env(), ref(), TestCheckBothSet(), TestCheckInvalidName(), TestCheckReferenceForm(), TestMessagesNeverCarryTheValue(), TestRefWiringErrors() (+8 more)

## Knowledge Gaps
- **2 isolated node(s):** `github.com/hafio/go-deploy-utility`, `NO_COLOR`
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 88 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **4 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Compose()` connect `paritydoc_test.go` to `Handoff: aligning solmq-conn-util and solace-util on Docker, Podman and Kubernetes`, `Appendix C: deferred code stages (not in this plan, separate approval)`?**
  _High betweenness centrality (0.099) - this node is a cross-community bridge._
- **Are the 22 inferred relationships involving `Check()` (e.g. with `Package layout and build order` and `TestAllowedBinaryIsNotGloballyAllowed()`) actually correct?**
  _`Check()` has 22 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/hafio/go-deploy-utility`, `NO_COLOR` to the rest of the system?**
  _2 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Developer guide` be split into smaller, more focused modules?**
  _Cohesion score 0.11695906432748537 - nodes in this community are weakly interconnected._
- **Why does `TestEveryDocCarriesTheSupportNotice()` connect `Handoff: aligning solmq-conn-util and solace-util on Docker, Podman and Kubernetes` to `noticetest_test.go`, `testing.T`, `Notice`?**
  _High betweenness centrality (0.089) - this node is a cross-community bridge._
- **Are the 27 inferred relationships involving `execguard, execguard_test.go` (e.g. with `Approve()` and `IsControl()`) actually correct?**
  _`execguard, execguard_test.go` has 27 INFERRED edges - model-reasoned connections that need verification._
- **Should `Handoff: aligning solmq-conn-util and solace-util on Docker, Podman and Kubernetes` be split into smaller, more focused modules?**
  _Cohesion score 0.06896551724137931 - nodes in this community are weakly interconnected._