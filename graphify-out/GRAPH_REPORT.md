# Graph Report - go-deploy-utility  (2026-11-28)

## Corpus Check
- 119 files · ~394,782 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 3 file(s) not represented in the graph (top: (none) 3)

## Summary
- 2021 nodes · 5860 edges · 131 communities (122 shown, 9 thin omitted)
- Extraction: 66% EXTRACTED · 34% INFERRED · 0% AMBIGUOUS · INFERRED: 1985 edges (avg confidence: 0.92)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `b452592f`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- dev.sh
- Tests by package
- testing.T
- podman_test.go
- dev.ps1
- noticetest_test.go
- prompter_test.go
- abbrev_test.go
- paritydoc_test.go
- handoff/README.md
- github.com/hafio/go-deploy-utility
- exitcode_test.go
- Node
- Handoff: aligning solmq-conn-util and solace-util on Docker, Podman and Kubernetes
- execguard_test.go
- 4. One table per field
- artifacttest_test.go
- nsgate.go
- 2. Operations
- No decision needed
- kube.go
- consuming.md
- inspect.go
- watch_test.go
- Plan review findings, 2026-11-13
- rbac.go
- dev-all.sh
- go_pkg_strings
- compose_test.go
- containercli.go
- dev-all.ps1
- artifacttest.go
- runnertest_test.go
- Exported names
- go_pkg_strings
- Cmd
- systemd_test.go
- expectRefusal
- 3. The description (Go sketch)
- Tests by package
- tests
- 10. One deploy digest for podman units
- Env-file keys
- Tests
- limits.go
- go_pkg_strings
- 11. Resolve and pin the kube context
- emitter
- 14. One check for the log tail count
- 15. Dev scripts build the pinned module unless asked for the checkout
- 17. A readiness wait for OnDelete StatefulSets
- 18. Kubernetes objects a config no longer names are removed
- 19. One shape for the platform sections of the env file
- 3. One step order for deploy and remove on every platform
- 2.5 kube.Client (kubectl or oc)
- 6. Every artifact states its network mode
- 7. One atomic write for files a tool installs
- 8. One gate for the test catalogue
- 9. Remove is idempotent on every platform
- 2. One guideline for expansion in config values
- The Kubernetes documents
- 4. Builders refuse a name outside its grammar
- Runner
- github.com/hafio/go-deploy-utility/docs/review
- Part 1: package `containerart/writer`
- 2.2 containercli.Client (docker, podman, nerdctl engines and compose)
- 2. One table per secret type
- FieldError
- The compose file
- The quadlet unit
- Shared by all three renderers
- 1. Tool A: solmq-conn-util
- workload.go
- block
- ParseImage
- 2.3 podman.Client
- 2.4 systemd.Client
- One container description and one writer (consolidation draft)
- 1. Exported surface
- 7. Coverage of what A and B write today
- 6.3 Inventory
- Docker delivery of tool-created TLS stores: is the script still the right way
- One catalogue (consolidation draft, 2026-11-17)
- Surface review, 2026-11-15
- Smoke checks
- 2. Per client (conventions, then each client)
- Plan R1: the workload package and its core checks
- 16. One behaviour for stale podman secrets
- 1. Ownership labels on every artifact
- go-deploy-utility
- Check
- 8. Rendering in the module
- unit_test.go
- abbrev
- confirm
- containerart/artifacttest
- containerart/compose
- containerart/quadlet
- containercli
- containercli/inspect
- envref
- execguard
- kube
- kube/nsgate
- kube/rbac
- notice
- notice/noticetest
- paritydoc
- podman
- rootless
- runner
- runner/runnertest
- systemd
- watch
- To do
- exitcode
- E25. The list of inherent differences
- K18. Kinds and fields nobody uses yet
- UnsetError
- body.md
- Critical review - go-deploy-utility
- 3. Alternatives
- ValidName
- 5. One diff helper for golden tests
- node
- .ExitCode
- 9. Reuse by other tools
- No decision needed

## God Nodes (most connected - your core abstractions)
1. `6.3 Inventory` - 154 edges
2. `No decision needed` - 118 edges
3. `Package layout and build order` - 109 edges
4. `Exported names` - 83 edges
5. `Cmd` - 68 edges
6. `Runner` - 54 edges
7. `runner, runner_test.go` - 54 edges
8. `Tests` - 48 edges
9. `Echo` - 47 edges
10. `FieldError` - 46 edges

## Surprising Connections (you probably didn't know these)
- `N64. `notice.Must`` --references--> `New()`  [INFERRED]
  docs/review/2026-11-15-surface-review-questions.md → abbrev/abbrev.go
- `N37. Where `artifacttest` lives, and the Kubernetes checks` --references--> `Widening()`  [INFERRED]
  docs/review/2026-11-15-surface-review-questions.md → containerart/artifacttest/artifacttest.go
- `Today` --references--> `HostNetwork()`  [INFERRED]
  docs/todo.md → containerart/artifacttest/artifacttest.go
- `CR-0011 [minor] [new] CheckSecretTargets passes a compose secrets block it does not recognise` --references--> `AssertSecretTargets()`  [INFERRED]
  docs/review/2026-11-11-critical-review.md → containerart/artifacttest/artifacttest.go
- `8.2 Proposal` --references--> `Quote()`  [INFERRED]
  docs/design/2026-11-15-shared-deploy-api.md → containerart/compose/compose.go

## Import Cycles
- None detected.

## Communities (131 total, 9 thin omitted)

### Community 0 - "dev.sh"
Cohesion: 0.19
Nodes (19): c(), expand(), finish(), log_begin(), NO_COLOR, run(), dev.sh script, die() (+11 more)

### Community 1 - "Tests by package"
Cohesion: 0.08
Nodes (22): A.1 Facts that shape the answer, A.2 Enforcement options, weakest to strongest, A.3 Key format and verification (options 1 and 2), A.4 Where the key comes from, A.5 Where the gate plugs in, A.6 Shared module, A.7 Code the two repos could share (synergy), A.8 Shared module: layout and GitHub mechanics (+14 more)

### Community 2 - "testing.T"
Cohesion: 0.06
Nodes (109): N17. Keeping the tool alive on Ctrl-C while a shell is attached, N2. The three capture helpers `Combined`, `Split`, `Output`, kube/rbac, rbac_test.go, rootless, rootless_test.go, runner, echo_test.go, runner, runner_test.go, TestAnswer(), TestCanIArgv() (+101 more)

### Community 3 - "podman_test.go"
Cohesion: 0.19
Nodes (18): N48. Does `QuadletDir` stay in `podman`, podman, podman_test.go, QuadletDir(), env(), homeAt(), TestCreateSecretEmptyValueIsNonNil(), TestCreateSecretLabelOrder(), TestCreateSecretReplaceIsOneCall() (+10 more)

### Community 4 - "dev.ps1"
Cohesion: 0.22
Nodes (12): Get-Log(), Get-Now(), Invoke-Logged(), Task-build(), Task-cov(), Task-graphify(), Task-scan(), Task-test() (+4 more)

### Community 5 - "noticetest_test.go"
Cohesion: 0.09
Nodes (56): 3. House rules that bind the next session, Files, parity.md content, Stage 0: capture (both repos, docs only), Verification, CR-0013 [nit] [new] CheckDocs passes when Exclude removes every selected doc, N63. The version command's output block, notice, notice_test.go (+48 more)

### Community 6 - "prompter_test.go"
Cohesion: 0.07
Nodes (43): countingReader, Form, Line(), TestLineReusesBufferedReader(), TestPromptYes(), TestPromptYesNo(), Yes(), YesNo() (+35 more)

### Community 7 - "abbrev_test.go"
Cohesion: 0.17
Nodes (22): checkShort(), New(), mustPanic(), sample(), TestAccessorsReturnCopies(), TestDigitsAreAllowed(), TestEntriesCloneKeepsNilShort(), TestExpandResolvesBothSpellings() (+14 more)

### Community 8 - "paritydoc_test.go"
Cohesion: 0.28
Nodes (15): N69. Two names nobody outside the module uses: `notice.Width` and `paritydoc.Title`, paritydoc, paritydoc_test.go, Compose(), brokerBlock(), connectorBlock(), noticeBlock(), page(), TestBodyShape() (+7 more)

### Community 12 - "exitcode_test.go"
Cohesion: 0.09
Nodes (40): Checks the owner runs, Decided 2026-11-15, Decisions on the 2026-11-11 critical review, Decisions on the plan review, Labels (gdu 1, 16, 18), Other items, Readings the owner confirmed (2026-11-15), Releases (+32 more)

### Community 13 - "Node"
Cohesion: 0.19
Nodes (40): TestBlockEmptyLinesCarryNoIndent(), TestBlockRefusals(), expectRefusal(), Block(), Bool(), DataKV(), EmptyMap(), Flow() (+32 more)

### Community 14 - "Handoff: aligning solmq-conn-util and solace-util on Docker, Podman and Kubernetes"
Cohesion: 0.12
Nodes (15): 1. What this is, 2. State at export, 4. Decisions already taken, 5.1 Same in both (the contract), 5.2 Different by policy, 5. Exact similarities and differences, 6.1 Repo A, solmq-conn-util, 6.3 Not scanned or not verified (+7 more)

### Community 16 - "execguard_test.go"
Cohesion: 0.15
Nodes (42): N13. `IsControl` and U+0085 (NEL), execguard, execguard_test.go, Approve(), base(), Check(), checkBinary(), checkFlagShape(), checkToken() (+34 more)

### Community 17 - "4. One table per field"
Cohesion: 0.09
Nodes (27): CheckComposeVar(), CheckMountSource(), CheckName(), hasDriveRoot(), ValidDNSLabel(), 4. One table per field, 8.1 Settled since the first draft, 8.2 Still open (+19 more)

### Community 18 - "artifacttest_test.go"
Cohesion: 0.26
Nodes (21): recorder, CheckAbsent(), HostNetwork(), composeFile(), TestAssertPassesQuietly(), TestCheckAbsent(), TestCheckComposePosture(), TestCheckQuadletKeys() (+13 more)

### Community 19 - "nsgate.go"
Cohesion: 0.08
Nodes (45): correctness, CR-0003 [major] [new] Namespace gate lists only all, PVCs, Secrets and ConfigMaps, so a namespace holding other kinds is offered for deletion, 1. Facts this review settled, 2.1 Across the three plans, 2.2 Labels: gdu 1, 16, 18, 2.3 gdu 10, deploy digest, 2.5 gdu 6, 7, 9, 11 to 14, 17, 2.6 Dev scripts and init-devops (+37 more)

### Community 20 - "2. Operations"
Cohesion: 0.05
Nodes (42): 2.1 Client methods, 2.2 Gaps: common operations no catalogue method covers, 2.3 Catalogue methods neither tool needs, 2.4 Operations both tools build locally, 2.5 Operations only one tool runs, 2.6 Owner questions, 2.7 Census notes, 2.8 Verifier (+34 more)

### Community 21 - "No decision needed"
Cohesion: 0.13
Nodes (25): CheckComposePosture(), ComposeDigest(), Containers(), 3.4 Catalogue, Adding a package, Adoption order, Developer guide, Package layout and build order (+17 more)

### Community 22 - "kube.go"
Cohesion: 0.12
Nodes (39): 2.4 gdu 2, 3, 4, kube, kube_test.go, 12. A bound on every delete, Decided, Goal, Scope, Today, Today (+31 more)

### Community 23 - "consuming.md"
Cohesion: 0.36
Nodes (4): Consuming go-deploy-utility, Develop against a local checkout, Pin an exact version, Release and merge order

### Community 24 - "inspect.go"
Cohesion: 0.17
Nodes (18): Decode(), health(), TestContainersArgv(), TestDecodeDocker(), TestDecodeHealthSpellings(), TestDecodeNeverReadsEnv(), TestDecodePartial(), TestDecodePodman() (+10 more)

### Community 25 - "watch_test.go"
Cohesion: 0.11
Nodes (35): healthy(), TestHealthCheck(), TestHealthQuadletRefusesShellCharacters(), TestHealthRetriesEchoesTheNumber(), TestParseSeconds(), health_test.go, watch, watch_test.go, Today (+27 more)

### Community 26 - "Plan review findings, 2026-11-13"
Cohesion: 0.06
Nodes (31): Answers, Answers, Answers, Answers, Answers, Answers, Answers, Dev scripts, CI and init-devops; gdu 15 (+23 more)

### Community 27 - "rbac.go"
Cohesion: 0.25
Nodes (10): N51. The delete check asks about one namespace, not all of them, DeleteProbe(), Answer(), CanI(), Check(), ClusterScoped(), Probe, Refused() (+2 more)

### Community 28 - "dev-all.sh"
Cohesion: 0.41
Nodes (12): is_tag(), known(), name(), notice(), notice_end(), pin(), required(), resolve() (+4 more)

### Community 29 - "go_pkg_strings"
Cohesion: 0.20
Nodes (18): Environment(), Escape(), EscapePercent(), Secret(), TestEnvironmentIsAlwaysQuoted(), TestEscapePercent(), TestEscapeWritesControlsAsCEscapes(), TestPostureBytes() (+10 more)

### Community 30 - "compose_test.go"
Cohesion: 0.24
Nodes (19): Secret, Escape(), FlowList(), Quote(), ServiceSecrets(), goldenSecrets(), TestComposeQuoteEscapes(), TestEscape() (+11 more)

### Community 31 - "containercli.go"
Cohesion: 0.08
Nodes (47): ByImage(), composeCmd(), ComposeDown(), ComposeUp(), EngineHint(), Exec(), HasName(), Info() (+39 more)

### Community 32 - "dev-all.ps1"
Cohesion: 0.38
Nodes (9): Get-Name(), Get-NoticeEnd(), Get-Required(), Invoke-Pin(), Invoke-Resolve(), Die(), Sync-Parity(), Show-Notice() (+1 more)

### Community 33 - "artifacttest.go"
Cohesion: 0.27
Nodes (17): unitLine, AssertAbsent(), AssertComposePosture(), AssertQuadletKeys(), AssertQuadletPosture(), AssertSecretTargets(), checkComposeSecret(), CheckQuadletKeys() (+9 more)

### Community 34 - "runnertest_test.go"
Cohesion: 0.08
Nodes (42): build, cicd, conventions, CR-0007 [nit] [new] cmd, stdin, shQuote and Upload are duplicated across the builder packages, CR-0008 [nit] [new] dev-all discovers sibling tasks by scraping their dev scripts with two different regexes, CR-0015 [minor] [new] No gate runs before a tag exists, and for a Go module the tag is the release, CR-0016 [minor] [new] CI never runs the tests under the race detector, CR-0017 [minor] [new] Dev scripts and ci.yml have fallen behind the dev-script contract (+34 more)

### Community 35 - "Exported names"
Cohesion: 0.05
Nodes (38): Exported names, Group: confirm, abbrev, watch and platform words, Group: containercli, inspect, the compose and quadlet helpers and artifacttest, Group: kube, nsgate and rbac, Group: notice, noticetest and paritydoc, Group: podman, systemd and rootless, Group: runner, exit codes, character rules and command guards, Group: test support (+30 more)

### Community 37 - "Cmd"
Cohesion: 0.21
Nodes (20): systemd, systemd_test.go, Cmd, Active(), DaemonReload(), EnableLinger(), IsActive(), NRestarts(), ParseLinger() (+12 more)

### Community 38 - "systemd_test.go"
Cohesion: 0.23
Nodes (13): 10. Rulings after the surface review, 7. Open points, 2. Rulings needed, in order, 4. Defects found on the way, N46. Typed unit state for `Active` and `Stopped`, N75. A drift-check helper for generated files, Settled by the rulings, Approvals (+5 more)

### Community 39 - "expectRefusal"
Cohesion: 0.12
Nodes (28): TestCheckComposeVar(), TestCheckMountSource(), TestCheckName(), TestValidDNSLabel(), TestValidHostName(), TestValidTimezone(), TestParsePort(), TestParsePortEchoesOnlyNumbers() (+20 more)

### Community 40 - "3. The description (Go sketch)"
Cohesion: 0.21
Nodes (23): everyDuration(), kubeKinds(), kubeRestartWords(), networkWords(), pullWords(), restartWords(), serviceTypes(), TestWordListsAreFresh() (+15 more)

### Community 41 - "Tests by package"
Cohesion: 0.19
Nodes (20): Coverage, envref, envref_test.go, internal/build, build_test.go, Running, Tests, Tests by package, watch, vt_windows_test.go, env() (+12 more)

### Community 42 - "tests"
Cohesion: 0.29
Nodes (7): squeeze(), CR-0009 [major] [new] Real-engine smoke checks for S7 to S9 are recorded as not run for a tagged release, CR-0010 [minor] [new] Widening and HostNetwork miss the podman flag spellings that ride PodmanArgs, CR-0011 [minor] [new] CheckSecretTargets passes a compose secrets block it does not recognise, CR-0012 [minor] [new] The Windows console branch of CanRedraw is never executed by a test, CR-0014 [nit] [new] TestBodyShape pins the parity body to a line count, tests

### Community 43 - "10. One deploy digest for podman units"
Cohesion: 0.33
Nodes (6): 10. One deploy digest for podman units, Decided, Goal, Scope, Today, Undecided

### Community 44 - "Env-file keys"
Cohesion: 0.08
Nodes (25): E10. B's operator pull Secret when the broker's is "existing", E11. B's admin password together with "existing", E12. Memory and CPU: one key name, two unit grammars, E13. Placement keys on Kubernetes, E14. Spread and B's always-on anti-affinity, E15. The Service block on Kubernetes, E16. A gains `docker.compose`; command keys take a list, E17. Which programs `docker.command` and `docker.compose` may name (+17 more)

### Community 45 - "Tests"
Cohesion: 0.22
Nodes (22): Part 2: module-root `internal/posture`, Tests, internal/posture, posture_test.go, composeEntries(), ComposeExtra(), ComposeExtraKeys(), quadletEntries(), QuadletExtra() (+14 more)

### Community 46 - "limits.go"
Cohesion: 0.15
Nodes (24): kubeQuantity(), kubeSuffixes(), limitValue(), ParseLimit(), refusedLimits(), sizeCheck(), sinkSize(), TestParseLimit() (+16 more)

### Community 47 - "go_pkg_strings"
Cohesion: 0.15
Nodes (15): ParseRunAs(), RunAsIDs(), TestParseRunAs(), TestParseRunAsRefuses(), TestRunAsIDs(), TestRunAsRootIsStatedNotEmpty(), Spec.RunAs, Coverage (+7 more)

### Community 48 - "11. Resolve and pin the kube context"
Cohesion: 0.33
Nodes (6): 11. Resolve and pin the kube context, Decided, Goal, Scope, Today, Undecided

### Community 49 - "emitter"
Cohesion: 0.22
Nodes (9): Dialect, Quote(), TestQuote(), blockProblem(), dataKeyProblem(), dotted(), keywordKeyProblem(), validVarName() (+1 more)

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

### Community 56 - "2.5 kube.Client (kubectl or oc)"
Cohesion: 0.10
Nodes (20): 2.5 kube.Client (kubectl or oc), Apply, Delete, CheckSecretKeys (new, Secrets decision 2026-11-27; working name), DeleteNamed (replaces DeletePod), DeleteNamespace, Exec, ExecOutput, Attach, Get, ListLabelled (decided only: gdu 18) (+12 more)

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
Cohesion: 0.33
Nodes (6): 2. One guideline for expansion in config values, Decided, Goal, Scope, Today, Undecided

### Community 62 - "The Kubernetes documents"
Cohesion: 0.11
Nodes (18): K10. Instance label of the operator's pull Secret, K11. Registry key when the image names none, K12. A's own affinity block, K13. A dollar sign in exec probes, K14. How the Secret file mode is written, K15. One list of what a manifest writes, K16. The size limit of one applied object, K17. Volume sources other than NFS (+10 more)

### Community 63 - "4. Builders refuse a name outside its grammar"
Cohesion: 0.50
Nodes (4): 4. Builders refuse a name outside its grammar, Goal, Scope, Undecided

### Community 64 - "Runner"
Cohesion: 0.10
Nodes (27): 11. Answers to the detail questions, 1. Direction, 2. What each tool repeats today, 3.1 One client per command family, 3.2 Output and results, 3.3 Rules for an executor, 3.5 What stays in each tool, 3.6 Example (+19 more)

### Community 66 - "Part 1: package `containerart/writer`"
Cohesion: 0.34
Nodes (6): escapeQuoted(), Unit, percent(), sectionIndex(), sectionName(), Part 1: package `containerart/writer`

### Community 67 - "2.2 containercli.Client (docker, podman, nerdctl engines and compose)"
Cohesion: 0.12
Nodes (16): 2.2 containercli.Client (docker, podman, nerdctl engines and compose), ComposeDown, ComposeUp, ComposeVersion (new, ruling 6), CopyTo, CopyFrom (new, ruling 6), Exec, ExecOutput, Attach (ruling 5), Info, Inspect (+8 more)

### Community 68 - "2. One table per secret type"
Cohesion: 0.14
Nodes (14): 1. Purpose and the proposed rule, 2.1 Kubernetes credentials (A connection credentials; B admin, monitor, psk), 2.2 Kubernetes image pull (broker or connector namespace), 2.3 B operator `regcred` (A has no operator), 2.4 TLS (A stores Secret and per-store existing Secrets; B server TLS), 2.5 B additional users, 2.6 Docker compose secrets (both), 2.7 Podman secrets (both) (+6 more)

### Community 69 - "FieldError"
Cohesion: 0.20
Nodes (13): ParseSeconds(), seconds(), ValidHostName(), ValidTimezone(), ParsePort(), portNumber(), portRange(), fail() (+5 more)

### Community 70 - "The compose file"
Cohesion: 0.14
Nodes (14): C10. Inline text that does not end in exactly one newline, C11. One service per file, C12. Named volumes and other networks, C1. One order of keys in the file, C2. B's word "bridge" would move every deploy to docker's shared bridge, C3. Read-only root and how compose hands over secrets, C4. Does the project network carry labels, C5. A secret or config target must not sit under a mount (+6 more)

### Community 71 - "The quadlet unit"
Cohesion: 0.17
Nodes (12): Q10. First line of the unit, Q1. How B gets its `[Service]` limit lines, Q2. The compare of the installed unit stays beside the digest label, Q3. Line order inside the unit, Q4. Which lines are written quoted, and what is escaped where, Q5. File mode of the installed unit: 0600 or 0644, Q6. What a health command means on podman, Q7. What a mount source may contain (+4 more)

### Community 72 - "Shared by all three renderers"
Cohesion: 0.14
Nodes (14): S10. Check at load as well as at render, S11. How `generate` writes its output file, S12. What a host name may contain, S13. What a time zone may contain, S2. Which settings get their own field, S3. A stated root user against a forgotten one, S4. The `solution` and `developed-by` labels, S5. Quotes on keys that are data (+6 more)

### Community 73 - "1. Tool A: solmq-conn-util"
Cohesion: 0.14
Nodes (14): 1. Tool A: solmq-conn-util, 2. Tool B: solace-util, 3. Moves, 4. The owner's answers (2026-11-18), 5. Still open, A: docker, A: image, A: kubernetes (+6 more)

### Community 74 - "workload.go"
Cohesion: 0.22
Nodes (9): CheckInlineText(), CheckText(), scan(), Contents, Dev tasks, N11. One exported check for a plain value (`execguard` charset check), No decision needed, IsControl() (+1 more)

### Community 75 - "block"
Cohesion: 0.21
Nodes (13): block(), TestBlockComposeMarkers(), TestBlockIndentIndicator(), TestBlockKubernetesNoKeep(), TestBlockRefusalsMatchCheckInlineText(), TestErrorText(), QuoteDollar(), TestQuoteDollar() (+5 more)

### Community 76 - "ParseImage"
Cohesion: 0.20
Nodes (13): isAlnum(), ParseImage(), TestParseImage(), TestParseImageRefuses(), validDigest(), validRegistry(), validRepository(), validTag() (+5 more)

### Community 78 - "2.3 podman.Client"
Cohesion: 0.22
Nodes (9): 2.3 podman.Client, CreateSecret, IDMap, RemoveSecrets, Run, Output, SecretIDs (decided only: gdu 10), SecretLabels (decided only: gdu 16), SecretMatches (+1 more)

### Community 79 - "2.4 systemd.Client"
Cohesion: 0.22
Nodes (9): 2.4 systemd.Client, DaemonReload, Linger, EnableLinger, LoadState (decided only: gdu 9), NRestarts, Run, Output, Show (new, ruling 6), Start, Stop, Restart (+1 more)

### Community 80 - "One container description and one writer (consolidation draft)"
Cohesion: 0.33
Nodes (6): 5.1 YAML sinks (compose, Kubernetes), 5.2 Quadlet lines (Q3, Q4), 5.3 Compose order (C1), 5.4 Kubernetes order, 5.5 Errors (S8), 5. The writer

### Community 81 - "1. Exported surface"
Cohesion: 0.22
Nodes (9): 1. Exported surface, Verifiers for the cli group, Verifiers for the containerart group, Verifiers for the containercli group, Verifiers for the docs group, Verifiers for the guards group, Verifiers for the kube group, Verifiers for the podman-host group (+1 more)

### Community 82 - "7. Coverage of what A and B write today"
Cohesion: 0.22
Nodes (10): Join(), TestJoin(), 7.1 A, compose (internal/dockergen/dockergen.go), 7.2 A, quadlet (internal/podmangen/podmangen.go), 7.3 A, Kubernetes (internal/deploy/deploy.go), 7.4 B, compose (internal/render/render.go), 7.5 B, quadlet (internal/render/render.go), 7.6 B, Kubernetes (internal/k8s, internal/render) (+2 more)

### Community 83 - "6.3 Inventory"
Cohesion: 0.17
Nodes (19): 6.3 Inventory, N41. The sub-id refusal: one function, and the range size, N42. `CheckIDMap`: typed error, and both maps in one call, N43. The core-limit checks: common or solace-util's own, N47. A single call that resolves the whole podman scope, CheckCore(), CheckIDMap(), CheckLinger() (+11 more)

### Community 84 - "Docker delivery of tool-created TLS stores: is the script still the right way"
Cohesion: 0.15
Nodes (12): 1. Purpose, 2.1 Where the store comes from, 2.2 The docker mechanism, 2.3 Who reads the stores in this connector, 2.4 The other platforms, 2.5 Facts that matter for the comparison, 2. Today, 4. Side by side (+4 more)

### Community 85 - "One catalogue (consolidation draft, 2026-11-17)"
Cohesion: 0.29
Nodes (6): 1. Purpose, 3. The 13 gap operations (findings 2.2), 4. Methods neither tool needs today (findings 2.3), 5. Host facts (ruling 7), 6. Exported names that go private, move or change, One catalogue (consolidation draft, 2026-11-17)

### Community 86 - "Surface review, 2026-11-15"
Cohesion: 0.29
Nodes (7): 1. Verdict, 3.2 Operations, 3.4 Env keys, 3. What each stream found, 5. Not verified, 6. Next work, Surface review, 2026-11-15

### Community 87 - "Smoke checks"
Cohesion: 0.29
Nodes (7): Environment, Runs, S7: redeploy asks before any write (not yet run), S8: the shared watch loop (not yet run), S9: kube, rbac and the namespace gate (not yet run), Smoke checks, v0.1.3: two fixes settled from source (not yet run)

### Community 88 - "2. Per client (conventions, then each client)"
Cohesion: 0.33
Nodes (6): 2.1.1 Method counts, 2.1 Set-up, writers, Run and Output (rulings 1 to 3), 2.6 kube/rbac and kube/nsgate, 2. Per client (conventions, then each client), nsgate.Read, rbac.Run

### Community 89 - "Plan R1: the workload package and its core checks"
Cohesion: 0.33
Nodes (6): Docs in the same step, Gates (the owner runs them), Package, Plan R1: the workload package and its core checks, Tests, Why this step first

### Community 90 - "16. One behaviour for stale podman secrets"
Cohesion: 0.33
Nodes (6): 16. One behaviour for stale podman secrets, Decided, Failed removal, Goal, Scope, Today

### Community 91 - "1. Ownership labels on every artifact"
Cohesion: 0.28
Nodes (9): 1. Ownership labels on every artifact, Decided, Goal, Inputs from the solace-util review to items 1 to 4 and 9, Scope, Today, Undecided, Until then (+1 more)

### Community 92 - "go-deploy-utility"
Cohesion: 0.33
Nodes (6): Development, go-deploy-utility, Licence, Packages, Status, Using it

### Community 93 - "Check"
Cohesion: 0.21
Nodes (9): Conventions, go-deploy-utility, Process, N70. `paritydoc.CheckFile`, N71. The parity page for a third tool, N74. The one diff function for golden and doc mismatches, Check(), clip() (+1 more)

### Community 94 - "8. Rendering in the module"
Cohesion: 0.40
Nodes (5): 8.1 Today, 8.2 Proposal, 8.3 What stays in the tool, 8.4 Effects, 8. Rendering in the module

### Community 95 - "unit_test.go"
Cohesion: 0.46
Nodes (8): NewUnit(), containerLines(), TestUnitBare(), TestUnitCollectsEveryProblem(), TestUnitExtra(), TestUnitQuoted(), TestUnitSections(), unit_test.go

### Community 96 - "abbrev"
Cohesion: 0.50
Nodes (4): abbrev, abbrev: missing, abbrev: placement, abbrev: signatures

### Community 97 - "confirm"
Cohesion: 0.50
Nodes (4): confirm, confirm: missing, confirm: placement, confirm: signatures

### Community 98 - "containerart/artifacttest"
Cohesion: 0.50
Nodes (4): containerart/artifacttest, containerart/artifacttest: missing, containerart/artifacttest: placement, containerart/artifacttest: signatures

### Community 99 - "containerart/compose"
Cohesion: 0.50
Nodes (4): containerart/compose, containerart/compose: missing, containerart/compose: placement, containerart/compose: signatures

### Community 100 - "containerart/quadlet"
Cohesion: 0.50
Nodes (4): containerart/quadlet, containerart/quadlet: missing, containerart/quadlet: placement, containerart/quadlet: signatures

### Community 101 - "containercli"
Cohesion: 0.50
Nodes (4): containercli, containercli: missing, containercli: placement, containercli: signatures

### Community 102 - "containercli/inspect"
Cohesion: 0.50
Nodes (4): containercli/inspect, containercli/inspect: missing, containercli/inspect: placement, containercli/inspect: signatures

### Community 103 - "envref"
Cohesion: 0.50
Nodes (4): envref, envref: missing, envref: placement, envref: signatures

### Community 104 - "execguard"
Cohesion: 0.50
Nodes (4): execguard, execguard: missing, execguard: placement, execguard: signatures

### Community 105 - "kube"
Cohesion: 0.50
Nodes (4): kube, kube: missing, kube: placement, kube: signatures

### Community 106 - "kube/nsgate"
Cohesion: 0.50
Nodes (4): kube/nsgate, kube/nsgate: missing, kube/nsgate: placement, kube/nsgate: signatures

### Community 107 - "kube/rbac"
Cohesion: 0.50
Nodes (4): kube/rbac, kube/rbac: missing, kube/rbac: placement, kube/rbac: signatures

### Community 108 - "notice"
Cohesion: 0.50
Nodes (4): notice, notice: missing, notice: placement, notice: signatures

### Community 109 - "notice/noticetest"
Cohesion: 0.50
Nodes (4): notice/noticetest, notice/noticetest: missing, notice/noticetest: placement, notice/noticetest: signatures

### Community 110 - "paritydoc"
Cohesion: 0.50
Nodes (4): paritydoc, paritydoc: missing, paritydoc: placement, paritydoc: signatures

### Community 111 - "podman"
Cohesion: 0.50
Nodes (4): podman, podman: missing, podman: placement, podman: signatures

### Community 112 - "rootless"
Cohesion: 0.50
Nodes (4): rootless, rootless: missing, rootless: placement, rootless: signatures

### Community 113 - "runner"
Cohesion: 0.50
Nodes (4): runner, runner: missing, runner: placement, runner: signatures

### Community 114 - "runner/runnertest"
Cohesion: 0.50
Nodes (4): runner/runnertest, runner/runnertest: missing, runner/runnertest: placement, runner/runnertest: signatures

### Community 115 - "systemd"
Cohesion: 0.50
Nodes (4): systemd, systemd: missing, systemd: placement, systemd: signatures

### Community 116 - "watch"
Cohesion: 0.50
Nodes (4): watch, watch: missing, watch: placement, watch: signatures

### Community 117 - "To do"
Cohesion: 0.40
Nodes (5): 20. Tool bugs found during the surface review, Goal, To do, Today, Undecided

### Community 118 - "exitcode"
Cohesion: 0.67
Nodes (3): exitcode, exitcode: placement, exitcode: signatures

### Community 119 - "E25. The list of inherent differences"
Cohesion: 0.67
Nodes (3): E25. The list of inherent differences, No decision needed, Settled by the rulings

### Community 120 - "K18. Kinds and fields nobody uses yet"
Cohesion: 0.67
Nodes (3): K18. Kinds and fields nobody uses yet, No decision needed, Settled by the rulings

### Community 122 - "body.md"
Cohesion: 0.21
Nodes (7): CR-0001 [minor] [new] Exported API with no consumer in either tool, CR-0002 [minor] [new] execguard accepts a flag value that makes the allowed binary run a file beside the env file, design, Contract, Decisions, Divergences, Inherent differences

### Community 123 - "Critical review - go-deploy-utility"
Cohesion: 0.18
Nodes (10): Accepted (not re-litigated), Coverage, CR-0005 [minor] [new] inspect reads podman secrets from a top-level Secrets key, CR-0006 [minor] [new] quadlet.Environment leaves a dollar sign for systemd to expand, Critical review - go-deploy-utility, Dynamic checks, Fixed since last report, Summary (+2 more)

### Community 124 - "3. Alternatives"
Cohesion: 0.25
Nodes (8): 3.1 Keep today's script (baseline), 3.2 One-shot init service decoding into a tmpfs-backed named volume, 3.3 PEM instead of PKCS12, 3.4 `secrets:` with `file:` from an in-memory host location, 3.5 Newer Docker Compose features that carry binary, 3.6 Letting Spring decode base64 itself, 3.7 Variants of the baseline (mine), 3. Alternatives

### Community 125 - "ValidName"
Cohesion: 0.33
Nodes (6): TestValidName(), ValidName(), Spec.Name, N15. `envref.ValidName`, N30. One container-name grammar or two, N59. The name of `kube.ValidLabel`, and the checks beside it

### Community 126 - "5. One diff helper for golden tests"
Cohesion: 0.33
Nodes (6): 5. One diff helper for golden tests, Decided, Goal, Scope, Today, Undecided

### Community 128 - ".ExitCode"
Cohesion: 0.40
Nodes (3): N3. What exit code means when a child is killed by a signal, N5. What `Result` holds when one operation runs two commands, exitedError

### Community 129 - "9. Reuse by other tools"
Cohesion: 0.50
Nodes (4): 9.1 Where a new tool stands, 9.2 Gaps beyond code, 9.3 Steps, in order, 9. Reuse by other tools

### Community 130 - "No decision needed"
Cohesion: 0.67
Nodes (3): No decision needed, Q11. `WantedBy` follows rootful or rootless, Settled by the rulings

## Knowledge Gaps
- **3 isolated node(s):** `github.com/hafio/go-deploy-utility/docs/review`, `github.com/hafio/go-deploy-utility`, `NO_COLOR`
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 578 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **9 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Runner` connect `Runner` to `testing.T`, `runnertest_test.go`, `go_pkg_strings`, `noticetest_test.go`, `prompter_test.go`, `Cmd`, `exitcode_test.go`, `Handoff: aligning solmq-conn-util and solace-util on Docker, Podman and Kubernetes`, `nsgate.go`, `6.3 Inventory`, `No decision needed`, `To do`, `kube.go`, `inspect.go`, `watch_test.go`, `rbac.go`, `containercli.go`?**
  _High betweenness centrality (0.149) - this node is a cross-community bridge._
- **Are the 153 inferred relationships involving `6.3 Inventory` (e.g. with `.Entries()` and `.Expand()`) actually correct?**
  _`6.3 Inventory` has 153 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/hafio/go-deploy-utility/docs/review`, `github.com/hafio/go-deploy-utility`, `NO_COLOR` to the rest of the system?**
  _3 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Tests by package` be split into smaller, more focused modules?**
  _Cohesion score 0.08 - nodes in this community are weakly interconnected._
- **Why does `To do` connect `To do` to `exitcode_test.go`, `kube.go`, `10. One deploy digest for podman units`, `11. Resolve and pin the kube context`, `14. One check for the log tail count`, `15. Dev scripts build the pinned module unless asked for the checkout`, `17. A readiness wait for OnDelete StatefulSets`, `18. Kubernetes objects a config no longer names are removed`, `19. One shape for the platform sections of the env file`, `3. One step order for deploy and remove on every platform`, `6. Every artifact states its network mode`, `7. One atomic write for files a tool installs`, `8. One gate for the test catalogue`, `9. Remove is idempotent on every platform`, `2. One guideline for expansion in config values`, `4. Builders refuse a name outside its grammar`, `Runner`, `todo.md`, `16. One behaviour for stale podman secrets`, `1. Ownership labels on every artifact`, `5. One diff helper for golden tests`?**
  _High betweenness centrality (0.141) - this node is a cross-community bridge._
- **Are the 117 inferred relationships involving `No decision needed` (e.g. with `.Entries()` and `.Expand()`) actually correct?**
  _`No decision needed` has 117 INFERRED edges - model-reasoned connections that need verification._
- **Should `testing.T` be split into smaller, more focused modules?**
  _Cohesion score 0.057566220823498555 - nodes in this community are weakly interconnected._