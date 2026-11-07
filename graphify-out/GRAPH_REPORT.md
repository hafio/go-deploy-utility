# Graph Report - go-deploy-utility  (2026-11-07)

## Corpus Check
- 68 files · ~96,916 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 3 file(s) not represented in the graph (top: (none) 3)

## Summary
- 914 nodes · 3025 edges · 27 communities (23 shown, 4 thin omitted)
- Extraction: 73% EXTRACTED · 27% INFERRED · 0% AMBIGUOUS · INFERRED: 827 edges (avg confidence: 0.91)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `8e118b07`
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
- runnertest_test.go
- execguard_test.go
- envref_test.go
- artifacttest.go
- nsgate.go
- Runner
- containercli.go
- Cmd
- Smoke checks
- inspect.go
- Echo
- compose_test.go

## God Nodes (most connected - your core abstractions)
1. `Package layout and build order` - 85 edges
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

## Communities (27 total, 4 thin omitted)

### Community 0 - "dev.sh"
Cohesion: 0.19
Nodes (19): c(), expand(), finish(), log_begin(), NO_COLOR, run(), dev.sh script, die() (+11 more)

### Community 1 - "podman_test.go"
Cohesion: 0.18
Nodes (26): podman, podman_test.go, cmd(), CreateSecret(), IDMap(), QuadletDir(), RemoveSecrets(), SecretData(), SecretMatches() (+18 more)

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
Cohesion: 0.07
Nodes (37): countingReader, Form, Line(), TestLineReusesBufferedReader(), TestPromptYes(), TestPromptYesNo(), Yes(), YesNo() (+29 more)

### Community 7 - "abbrev_test.go"
Cohesion: 0.15
Nodes (21): checkShort(), New(), mustPanic(), sample(), TestAccessorsReturnCopies(), TestDigitsAreAllowed(), TestEntriesCloneKeepsNilShort(), TestExpandResolvesBothSpellings() (+13 more)

### Community 8 - "paritydoc_test.go"
Cohesion: 0.07
Nodes (41): Conventions, go-deploy-utility, Process, Consuming go-deploy-utility, Develop against a local checkout, Pin an exact version, Release and merge order, Coverage (+33 more)

### Community 12 - "exitcode_test.go"
Cohesion: 0.18
Nodes (23): exitcode, exitcode_test.go, childExitError, AsUsage(), ChildExit(), ChildStatus(), Code(), Exited(), IsUsage() (+15 more)

### Community 13 - "rootless.go"
Cohesion: 0.11
Nodes (33): rootless, rootless_test.go, CheckCore(), CheckIDMap(), CheckLinger(), covers(), Decide(), HardLimit(), LimitText() (+25 more)

### Community 14 - "runnertest_test.go"
Cohesion: 0.17
Nodes (20): runner/runnertest, runnertest_test.go, cloneStrings(), FailOn(), Seq(), cmd(), files(), TestCallHasAndString(), TestFakeAttach() (+12 more)

### Community 16 - "execguard_test.go"
Cohesion: 0.14
Nodes (43): execguard, execguard_test.go, Approvals, Approve(), base(), Check(), checkBinary(), checkFlagShape(), checkToken() (+35 more)

### Community 17 - "envref_test.go"
Cohesion: 0.19
Nodes (16): envref, envref_test.go, env(), ref(), TestCheckBothSet(), TestCheckInvalidName(), TestCheckReferenceForm(), TestMessagesNeverCarryTheValue(), TestRefWiringErrors() (+8 more)

### Community 18 - "artifacttest.go"
Cohesion: 0.11
Nodes (48): recorder, unitLine, AssertAbsent(), AssertComposePosture(), AssertQuadletKeys(), AssertQuadletPosture(), AssertSecretTargets(), CheckAbsent() (+40 more)

### Community 19 - "nsgate.go"
Cohesion: 0.10
Nodes (38): kube/nsgate, nsgate_test.go, kube/rbac, rbac_test.go, DeleteProbe(), discounted(), List(), Occupants(), ownedBy(), Protected() (+30 more)

### Community 20 - "Runner"
Cohesion: 0.17
Nodes (19): Adoption order, watch, watch_test.go, Options, CanRedraw(), CheckInterval(), Loop(), section(), Sleep() (+11 more)

### Community 21 - "containercli.go"
Cohesion: 0.09
Nodes (45): ByImage(), cmd(), composeCmd(), ComposeDigest(), ComposeDown(), ComposeUp(), EngineHint(), Exec() (+37 more)

### Community 22 - "Cmd"
Cohesion: 0.06
Nodes (77): List(), Containers(), Adding a package, Dev tasks, Developer guide, Package layout and build order, Rules for shared code, kube, kube_test.go (+69 more)

### Community 23 - "Smoke checks"
Cohesion: 0.33
Nodes (5): Environment, Runs, S7: redeploy asks before any write (not yet run), S8: the shared watch loop (not yet run), Smoke checks

### Community 24 - "inspect.go"
Cohesion: 0.17
Nodes (17): Decode(), health(), Printable(), TestContainersArgv(), TestDecodeDocker(), TestDecodeHealthSpellings(), TestDecodeNeverReadsEnv(), TestDecodePartial() (+9 more)

### Community 26 - "Echo"
Cohesion: 0.07
Nodes (40): 1. What this is, 2. State at export, 3. House rules that bind the next session, 4. Decisions already taken, 5.1 Same in both (the contract), 5.2 Different by policy, 5.3 Runner seams (why nothing is shared yet), 5. Exact similarities and differences (+32 more)

### Community 31 - "compose_test.go"
Cohesion: 0.24
Nodes (16): Secret, Escape(), FlowList(), Quote(), ServiceSecrets(), goldenSecrets(), TestComposeQuoteEscapes(), TestEscape() (+8 more)

## Knowledge Gaps
- **2 isolated node(s):** `github.com/hafio/go-deploy-utility`, `NO_COLOR`
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 90 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **4 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Package layout and build order` connect `Cmd` to `podman_test.go`, `testing.T`, `noticetest_test.go`, `prompter_test.go`, `paritydoc_test.go`, `exitcode_test.go`, `rootless.go`, `runnertest_test.go`, `execguard_test.go`, `envref_test.go`, `artifacttest.go`, `nsgate.go`, `Runner`, `containercli.go`, `inspect.go`, `Echo`, `compose_test.go`?**
  _High betweenness centrality (0.108) - this node is a cross-community bridge._
- **Are the 84 inferred relationships involving `Package layout and build order` (e.g. with `Prompter` and `.Ask()`) actually correct?**
  _`Package layout and build order` has 84 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/hafio/go-deploy-utility`, `NO_COLOR` to the rest of the system?**
  _2 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `testing.T` be split into smaller, more focused modules?**
  _Cohesion score 0.13333333333333333 - nodes in this community are weakly interconnected._
- **Why does `Echo` connect `Echo` to `podman_test.go`, `noticetest_test.go`, `prompter_test.go`, `rootless.go`, `runnertest_test.go`, `execguard_test.go`, `envref_test.go`, `nsgate.go`, `containercli.go`, `Cmd`, `inspect.go`?**
  _High betweenness centrality (0.054) - this node is a cross-community bridge._
- **Are the 4 inferred relationships involving `Cmd` (e.g. with `Conventions` and `Package layout and build order`) actually correct?**
  _`Cmd` has 4 INFERRED edges - model-reasoned connections that need verification._
- **Should `Appendix C: deferred code stages (not in this plan, separate approval)` be split into smaller, more focused modules?**
  _Cohesion score 0.08 - nodes in this community are weakly interconnected._