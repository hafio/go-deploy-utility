# Graph Report - go-deploy-utility  (2026-11-05)

## Corpus Check
- 67 files · ~93,725 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 3 file(s) not represented in the graph (top: (none) 3)

## Summary
- 896 nodes · 2976 edges · 24 communities (20 shown, 4 thin omitted)
- Extraction: 73% EXTRACTED · 27% INFERRED · 0% AMBIGUOUS · INFERRED: 800 edges (avg confidence: 0.9)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `2a0ff850`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- dev.sh
- podman_test.go
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
- Echo
- execguard_test.go
- envref_test.go
- artifacttest.go
- nsgate.go
- Runner
- containercli.go
- Package layout and build order
- inspect.go

## God Nodes (most connected - your core abstractions)
1. `Package layout and build order` - 80 edges
2. `Cmd` - 63 edges
3. `runner, runner_test.go` - 54 edges
4. `Runner` - 46 edges
5. `Echo` - 40 edges
6. `helperArgv()` - 36 edges
7. `helperOn()` - 36 edges
8. `Check()` - 30 edges
9. `execguard, execguard_test.go` - 28 edges
10. `Notice` - 27 edges

## Surprising Connections (you probably didn't know these)
- `3. House rules that bind the next session` --references--> `TestEveryDocCarriesTheSupportNotice()`  [INFERRED]
  docs/handoff/HANDOFF.md → docs_test.go
- `Stage 2: B compose on stdin, like A` --references--> `Delete()`  [INFERRED]
  docs/handoff/PLAN.md → kube/kube.go
- `Package layout and build order` --references--> `Prompter`  [INFERRED]
  docs/developer.md → confirm/prompter.go
- `Package layout and build order` --references--> `Widening()`  [INFERRED]
  docs/developer.md → containerart/artifacttest/artifacttest.go
- `Package layout and build order` --references--> `HostNetwork()`  [INFERRED]
  docs/developer.md → containerart/artifacttest/artifacttest.go

## Import Cycles
- None detected.

## Communities (24 total, 4 thin omitted)

### Community 0 - "dev.sh"
Cohesion: 0.19
Nodes (19): c(), expand(), finish(), log_begin(), NO_COLOR, run(), dev.sh script, die() (+11 more)

### Community 1 - "podman_test.go"
Cohesion: 0.21
Nodes (22): podman, podman_test.go, cmd(), CreateSecret(), IDMap(), QuadletDir(), RemoveSecrets(), SupportsReplace(), env() (+14 more)

### Community 2 - "testing.T"
Cohesion: 0.13
Nodes (61): runner, runner_test.go, Combined(), Output(), Split(), attachFiles(), captureStdout(), helperArgv(), helperOn() (+53 more)

### Community 3 - "Appendix C: deferred code stages (not in this plan, separate approval)"
Cohesion: 0.08
Nodes (22): A.1 Facts that shape the answer, A.2 Enforcement options, weakest to strongest, A.3 Key format and verification (options 1 and 2), A.4 Where the key comes from, A.5 Where the gate plugs in, A.6 Shared module, A.7 Code the two repos could share (synergy), A.8 Shared module: layout and GitHub mechanics (+14 more)

### Community 4 - "dev.ps1"
Cohesion: 0.22
Nodes (12): Get-Log(), Get-Now(), Invoke-Logged(), Task-build(), Task-cov(), Task-graphify(), Task-scan(), Task-test() (+4 more)

### Community 5 - "noticetest_test.go"
Cohesion: 0.09
Nodes (52): Files, parity.md content, Stage 0: capture (both repos, docs only), Verification, notice, notice_test.go, notice/noticetest, noticetest_test.go, root, docs_test.go, TestEveryDocCarriesTheSupportNotice() (+44 more)

### Community 6 - "prompter_test.go"
Cohesion: 0.08
Nodes (35): countingReader, Form, Line(), TestLineReusesBufferedReader(), TestPromptYes(), TestPromptYesNo(), Yes(), YesNo() (+27 more)

### Community 7 - "abbrev_test.go"
Cohesion: 0.15
Nodes (21): checkShort(), New(), mustPanic(), sample(), TestAccessorsReturnCopies(), TestDigitsAreAllowed(), TestEntriesCloneKeepsNilShort(), TestExpandResolvesBothSpellings() (+13 more)

### Community 8 - "paritydoc_test.go"
Cohesion: 0.08
Nodes (36): Conventions, go-deploy-utility, Process, Consuming go-deploy-utility, Develop against a local checkout, Pin an exact version, Release and merge order, paritydoc, paritydoc_test.go (+28 more)

### Community 12 - "exitcode_test.go"
Cohesion: 0.15
Nodes (25): Coverage, exitcode, exitcode_test.go, Running, Summary, Tests, Tests by package, childExitError, AsUsage() (+17 more)

### Community 13 - "rootless.go"
Cohesion: 0.11
Nodes (34): rootless, rootless_test.go, CheckCore(), CheckIDMap(), CheckLinger(), covers(), Decide(), HardLimit(), LimitText() (+26 more)

### Community 14 - "Echo"
Cohesion: 0.07
Nodes (41): 1. What this is, 2. State at export, 3. House rules that bind the next session, 4. Decisions already taken, 5.1 Same in both (the contract), 5.2 Different by policy, 5.3 Runner seams (why nothing is shared yet), 5. Exact similarities and differences (+33 more)

### Community 16 - "execguard_test.go"
Cohesion: 0.14
Nodes (43): execguard, execguard_test.go, Approvals, Approve(), base(), Check(), checkBinary(), checkFlagShape(), checkToken() (+35 more)

### Community 17 - "envref_test.go"
Cohesion: 0.20
Nodes (16): envref, envref_test.go, env(), ref(), TestCheckBothSet(), TestCheckInvalidName(), TestCheckReferenceForm(), TestMessagesNeverCarryTheValue(), TestRefWiringErrors() (+8 more)

### Community 18 - "artifacttest.go"
Cohesion: 0.08
Nodes (64): recorder, unitLine, Secret, AssertAbsent(), AssertComposePosture(), AssertQuadletKeys(), AssertQuadletPosture(), AssertSecretTargets() (+56 more)

### Community 19 - "nsgate.go"
Cohesion: 0.10
Nodes (38): kube/nsgate, nsgate_test.go, kube/rbac, rbac_test.go, DeleteProbe(), discounted(), List(), Occupants(), ownedBy(), Protected() (+30 more)

### Community 20 - "Runner"
Cohesion: 0.07
Nodes (45): runner/runnertest, runnertest_test.go, watch, watch_test.go, Exec, Runner, applyEnv(), classify(), Resolve(), cloneStrings() (+37 more)

### Community 21 - "containercli.go"
Cohesion: 0.10
Nodes (44): ByImage(), cmd(), composeCmd(), ComposeDown(), ComposeUp(), EngineHint(), Exec(), HasName() (+36 more)

### Community 22 - "Package layout and build order"
Cohesion: 0.07
Nodes (72): Containers(), Printable(), Adding a package, Adoption order, Dev tasks, Developer guide, Package layout and build order, Rules for shared code (+64 more)

### Community 24 - "inspect.go"
Cohesion: 0.18
Nodes (16): Decode(), health(), TestContainersArgv(), TestDecodeDocker(), TestDecodeHealthSpellings(), TestDecodeNeverReadsEnv(), TestDecodePartial(), TestDecodePodman() (+8 more)

## Knowledge Gaps
- **2 isolated node(s):** `github.com/hafio/go-deploy-utility`, `NO_COLOR`
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 82 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **4 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Package layout and build order` connect `Package layout and build order` to `podman_test.go`, `testing.T`, `noticetest_test.go`, `prompter_test.go`, `paritydoc_test.go`, `exitcode_test.go`, `rootless.go`, `Echo`, `execguard_test.go`, `envref_test.go`, `artifacttest.go`, `nsgate.go`, `Runner`, `containercli.go`, `inspect.go`?**
  _High betweenness centrality (0.106) - this node is a cross-community bridge._
- **Are the 79 inferred relationships involving `Package layout and build order` (e.g. with `Prompter` and `.Ask()`) actually correct?**
  _`Package layout and build order` has 79 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/hafio/go-deploy-utility`, `NO_COLOR` to the rest of the system?**
  _2 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `testing.T` be split into smaller, more focused modules?**
  _Cohesion score 0.13333333333333333 - nodes in this community are weakly interconnected._
- **Why does `Echo` connect `Echo` to `podman_test.go`, `noticetest_test.go`, `prompter_test.go`, `rootless.go`, `execguard_test.go`, `artifacttest.go`, `nsgate.go`, `Runner`, `containercli.go`, `Package layout and build order`, `inspect.go`?**
  _High betweenness centrality (0.055) - this node is a cross-community bridge._
- **Are the 3 inferred relationships involving `Cmd` (e.g. with `Conventions` and `Package layout and build order`) actually correct?**
  _`Cmd` has 3 INFERRED edges - model-reasoned connections that need verification._
- **Should `Appendix C: deferred code stages (not in this plan, separate approval)` be split into smaller, more focused modules?**
  _Cohesion score 0.08 - nodes in this community are weakly interconnected._