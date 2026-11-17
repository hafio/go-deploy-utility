# Graph Report - go-deploy-utility  (2026-11-18)

## Corpus Check
- 103 files · ~360,650 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 3 file(s) not represented in the graph (top: (none) 3)

## Summary
- 1813 nodes · 5026 edges · 126 communities (117 shown, 9 thin omitted)
- Extraction: 66% EXTRACTED · 34% INFERRED · 0% AMBIGUOUS · INFERRED: 1721 edges (avg confidence: 0.92)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `813104fa`
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
- rootless_test.go
- Handoff: aligning solmq-conn-util and solace-util on Docker, Podman and Kubernetes
- execguard_test.go
- IsControl
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
- Findings by area
- Exported names
- Echo
- Cmd
- systemd_test.go
- expectRefusal
- Sink
- Tests by package
- tests
- 10. One deploy digest for podman units
- Env-file keys
- Loop
- limits.go
- go_pkg_strings
- 11. Resolve and pin the kube context
- 13. A watch loop stops on an error no retry can fix
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
- Shared deployment API (design note)
- github.com/hafio/go-deploy-utility/docs/review
- runnertest_test.go
- 2.2 containercli.Client (docker, podman, nerdctl engines and compose)
- 4. One table per field
- FieldError
- The compose file
- The quadlet unit
- Shared by all three renderers
- 1. Tool A: solmq-conn-util
- Fake
- rootless.go
- ParseImage
- 2.3 podman.Client
- 2.4 systemd.Client
- One container description and one writer (consolidation draft)
- 1. Exported surface
- 7. Coverage of what A and B write today
- CheckCore
- Decisions on the 2026-11-11 critical review
- One catalogue (consolidation draft, 2026-11-17)
- Surface review, 2026-11-15
- Smoke checks
- 2. Per client (conventions, then each client)
- Plan R1: the workload package and its core checks
- 16. One behaviour for stale podman secrets
- 1. Ownership labels on every artifact
- go-deploy-utility
- Widening
- 8. Rendering in the module
- health_test.go
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
- Critical review - go-deploy-utility
- TestParsePortRefuses
- writes

## God Nodes (most connected - your core abstractions)
1. `6.3 Inventory` - 156 edges
2. `No decision needed` - 120 edges
3. `Package layout and build order` - 87 edges
4. `Exported names` - 83 edges
5. `Cmd` - 67 edges
6. `runner, runner_test.go` - 55 edges
7. `Runner` - 54 edges
8. `Echo` - 47 edges
9. `FieldError` - 39 edges
10. `helperArgv()` - 36 edges

## Surprising Connections (you probably didn't know these)
- `N37. Where `artifacttest` lives, and the Kubernetes checks` --references--> `Widening()`  [INFERRED]
  docs/review/2026-11-15-surface-review-questions.md → containerart/artifacttest/artifacttest.go
- `Today` --references--> `HostNetwork()`  [INFERRED]
  docs/todo.md → containerart/artifacttest/artifacttest.go
- `CR-0011 [minor] [new] CheckSecretTargets passes a compose secrets block it does not recognise` --references--> `AssertSecretTargets()`  [INFERRED]
  docs/review/2026-11-11-critical-review.md → containerart/artifacttest/artifacttest.go
- `8.2 Proposal` --references--> `Quote()`  [INFERRED]
  docs/design/2026-11-15-shared-deploy-api.md → containerart/compose/compose.go
- `2.7 A's file` --references--> `EscapePercent()`  [INFERRED]
  docs/review/2026-11-13-plan-review.md → containerart/quadlet/quadlet.go

## Import Cycles
- None detected.

## Communities (126 total, 9 thin omitted)

### Community 0 - "dev.sh"
Cohesion: 0.19
Nodes (19): c(), expand(), finish(), log_begin(), NO_COLOR, run(), dev.sh script, die() (+11 more)

### Community 1 - "Tests by package"
Cohesion: 0.08
Nodes (22): A.1 Facts that shape the answer, A.2 Enforcement options, weakest to strongest, A.3 Key format and verification (options 1 and 2), A.4 Where the key comes from, A.5 Where the gate plugs in, A.6 Shared module, A.7 Code the two repos could share (synergy), A.8 Shared module: layout and GitHub mechanics (+14 more)

### Community 2 - "testing.T"
Cohesion: 0.07
Nodes (89): IsCharDevice(), StdinIsTerminal(), swapStdin(), TestIsCharDeviceDevNull(), TestIsCharDeviceReportsAPipe(), TestStdinIsTerminalFalseForPipeFileAndClosed(), N16. One home for the terminal facts, and the Windows console switch for attached sessions, N17. Keeping the tool alive on Ctrl-C while a shell is attached (+81 more)

### Community 3 - "podman_test.go"
Cohesion: 0.21
Nodes (18): N48. Does `QuadletDir` stay in `podman`, podman, podman_test.go, QuadletDir(), env(), homeAt(), TestCreateSecretEmptyValueIsNonNil(), TestCreateSecretLabelOrder(), TestCreateSecretReplaceIsOneCall() (+10 more)

### Community 4 - "dev.ps1"
Cohesion: 0.22
Nodes (12): Get-Log(), Get-Now(), Invoke-Logged(), Task-build(), Task-cov(), Task-graphify(), Task-scan(), Task-test() (+4 more)

### Community 5 - "noticetest_test.go"
Cohesion: 0.09
Nodes (56): 3. House rules that bind the next session, Files, parity.md content, Stage 0: capture (both repos, docs only), Verification, CR-0013 [nit] [new] CheckDocs passes when Exclude removes every selected doc, N63. The version command's output block, notice, notice_test.go (+48 more)

### Community 6 - "prompter_test.go"
Cohesion: 0.09
Nodes (36): countingReader, Form, Line(), TestLineReusesBufferedReader(), TestPromptYes(), TestPromptYesNo(), Yes(), YesNo() (+28 more)

### Community 7 - "abbrev_test.go"
Cohesion: 0.16
Nodes (23): checkShort(), New(), mustPanic(), sample(), TestAccessorsReturnCopies(), TestDigitsAreAllowed(), TestEntriesCloneKeepsNilShort(), TestExpandResolvesBothSpellings() (+15 more)

### Community 8 - "paritydoc_test.go"
Cohesion: 0.06
Nodes (53): Conventions, go-deploy-utility, Process, Decisions on the findings, 1. Facts this review settled, 2.1 Across the three plans, 2.2 Labels: gdu 1, 16, 18, 2.3 gdu 10, deploy digest (+45 more)

### Community 12 - "exitcode_test.go"
Cohesion: 0.16
Nodes (25): N3. What exit code means when a child is killed by a signal, N4. `exitcode.ChildStatus` stays or goes private, N5. What `Result` holds when one operation runs two commands, exitcode, exitcode_test.go, childExitError, AsUsage(), ChildExit(), ChildStatus() (+17 more)

### Community 13 - "rootless_test.go"
Cohesion: 0.23
Nodes (17): N38. Does the session repair reach podman commands too, N47. A single call that resolves the whole podman scope, rootless, rootless_test.go, Decide(), dirs(), env(), procLimits(), TestCheckLinger() (+9 more)

### Community 14 - "Handoff: aligning solmq-conn-util and solace-util on Docker, Podman and Kubernetes"
Cohesion: 0.10
Nodes (18): 1. What this is, 2. State at export, 4. Decisions already taken, 5.1 Same in both (the contract), 5.2 Different by policy, 5.3 Runner seams (why nothing is shared yet), 5. Exact similarities and differences, 6.1 Repo A, solmq-conn-util (+10 more)

### Community 16 - "execguard_test.go"
Cohesion: 0.15
Nodes (43): N13. `IsControl` and U+0085 (NEL), execguard, execguard_test.go, Approvals, Approve(), base(), Check(), checkBinary(), checkFlagShape() (+35 more)

### Community 17 - "IsControl"
Cohesion: 0.11
Nodes (24): CheckMountSource(), CheckName(), hasDriveRoot(), ValidDNSLabel(), CheckText(), TestValidName(), ValidName(), 8. Open points (+16 more)

### Community 18 - "artifacttest_test.go"
Cohesion: 0.32
Nodes (15): recorder, CheckAbsent(), composeFile(), TestCheckAbsent(), TestCheckComposePosture(), TestCheckQuadletKeys(), TestCheckQuadletKeysBadFloor(), TestCheckQuadletPosture() (+7 more)

### Community 19 - "nsgate.go"
Cohesion: 0.13
Nodes (32): N49. How the namespace check lists what is in a namespace, N52. The module refuses to delete a protected namespace itself, N55. `Object` moves to `kube`, and `Occupants` takes and returns `Object`, kube/nsgate, nsgate_test.go, kube/rbac, rbac_test.go, discounted(), Occupants(), ownedBy() (+24 more)

### Community 20 - "2. Operations"
Cohesion: 0.05
Nodes (42): 2.1 Client methods, 2.2 Gaps: common operations no catalogue method covers, 2.3 Catalogue methods neither tool needs, 2.4 Operations both tools build locally, 2.5 Operations only one tool runs, 2.6 Owner questions, 2.7 Census notes, 2.8 Verifier (+34 more)

### Community 21 - "No decision needed"
Cohesion: 0.13
Nodes (25): ComposeDigest(), Remove(), Containers(), Printable(), Adding a package, Adoption order, Dev tasks, Developer guide (+17 more)

### Community 22 - "kube.go"
Cohesion: 0.12
Nodes (40): 3.4 Catalogue, 2.4 gdu 2, 3, 4, kube, kube_test.go, Decided, Inputs from the solace-util review to items 1 to 4 and 9, Scope, Today, Today (+32 more)

### Community 23 - "consuming.md"
Cohesion: 0.21
Nodes (8): Consuming go-deploy-utility, Develop against a local checkout, Pin an exact version, Release and merge order, Coverage, Running, Summary, Tests

### Community 24 - "inspect.go"
Cohesion: 0.17
Nodes (18): Decode(), health(), TestContainersArgv(), TestDecodeDocker(), TestDecodeHealthSpellings(), TestDecodeNeverReadsEnv(), TestDecodePartial(), TestDecodePodman() (+10 more)

### Community 25 - "watch_test.go"
Cohesion: 0.21
Nodes (20): watch, watch_test.go, clock(), run(), runFrames(), stopAfter(), TestCanRedraw(), TestCheckInterval(), TestDefaultSleepHonoursCancel() (+12 more)

### Community 26 - "Plan review findings, 2026-11-13"
Cohesion: 0.06
Nodes (31): Answers, Answers, Answers, Answers, Answers, Answers, Answers, Dev scripts, CI and init-devops; gdu 15 (+23 more)

### Community 27 - "rbac.go"
Cohesion: 0.28
Nodes (11): 6.3 Inventory, N51. The delete check asks about one namespace, not all of them, DeleteProbe(), Answer(), CanI(), Check(), ClusterScoped(), Probe (+3 more)

### Community 28 - "dev-all.sh"
Cohesion: 0.41
Nodes (12): is_tag(), known(), name(), notice(), notice_end(), pin(), required(), resolve() (+4 more)

### Community 29 - "go_pkg_strings"
Cohesion: 0.22
Nodes (16): Environment(), Escape(), EscapePercent(), Secret(), TestEnvironmentIsAlwaysQuoted(), TestEscapePercent(), TestEscapeWritesControlsAsCEscapes(), TestPostureBytes() (+8 more)

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
Cohesion: 0.24
Nodes (20): unitLine, AssertAbsent(), AssertComposePosture(), AssertQuadletKeys(), AssertQuadletPosture(), AssertSecretTargets(), CheckComposePosture(), checkComposeSecret() (+12 more)

### Community 34 - "Findings by area"
Cohesion: 0.08
Nodes (25): build, cicd, conventions, correctness, CR-0001 [minor] [new] Exported API with no consumer in either tool, CR-0002 [minor] [new] execguard accepts a flag value that makes the allowed binary run a file beside the env file, CR-0003 [major] [new] Namespace gate lists only all, PVCs, Secrets and ConfigMaps, so a namespace holding other kinds is offered for deletion, CR-0007 [nit] [new] cmd, stdin, shQuote and Upload are duplicated across the builder packages (+17 more)

### Community 35 - "Exported names"
Cohesion: 0.05
Nodes (38): Exported names, Group: confirm, abbrev, watch and platform words, Group: containercli, inspect, the compose and quadlet helpers and artifacttest, Group: kube, nsgate and rbac, Group: notice, noticetest and paritydoc, Group: podman, systemd and rootless, Group: runner, exit codes, character rules and command guards, Group: test support (+30 more)

### Community 36 - "Echo"
Cohesion: 0.13
Nodes (11): A.7.1 Runtime layer: docker, podman, kubectl, Stage 1: B runner seam, minimum for stdin plus env, N7. One text form for a whole command (`Quote`, `MaskEnv`, `Call.String`, `Announce`), Echo, MaskEnv(), printable(), Quote(), quoteTok() (+3 more)

### Community 37 - "Cmd"
Cohesion: 0.16
Nodes (23): systemd, systemd_test.go, ProcLimits(), CheckAttach(), CheckRun(), Cmd, Active(), DaemonReload(), EnableLinger() (+15 more)

### Community 38 - "systemd_test.go"
Cohesion: 0.26
Nodes (12): 10. Rulings after the surface review, 7. Open points, 2. Rulings needed, in order, 4. Defects found on the way, N46. Typed unit state for `Active` and `Stopped`, N75. A drift-check helper for generated files, Settled by the rulings, Forbidden() (+4 more)

### Community 39 - "expectRefusal"
Cohesion: 0.22
Nodes (15): TestCheckComposeVar(), TestCheckMountSource(), TestCheckName(), TestValidDNSLabel(), TestValidHostName(), TestValidTimezone(), rows(), TestCheckInlineText() (+7 more)

### Community 40 - "Sink"
Cohesion: 0.20
Nodes (13): networkWords(), restartWords(), TestWordListsAreFresh(), ValidNetwork(), ValidRestart(), CheckInlineText(), checkSink(), decimal() (+5 more)

### Community 41 - "Tests by package"
Cohesion: 0.26
Nodes (16): envref, envref_test.go, internal/build, build_test.go, Tests by package, watch, vt_windows_test.go, env(), ref(), TestCheckBothSet(), TestCheckInvalidName() (+8 more)

### Community 42 - "tests"
Cohesion: 0.29
Nodes (7): squeeze(), CR-0009 [major] [new] Real-engine smoke checks for S7 to S9 are recorded as not run for a tagged release, CR-0010 [minor] [new] Widening and HostNetwork miss the podman flag spellings that ride PodmanArgs, CR-0011 [minor] [new] CheckSecretTargets passes a compose secrets block it does not recognise, CR-0012 [minor] [new] The Windows console branch of CanRedraw is never executed by a test, CR-0014 [nit] [new] TestBodyShape pins the parity body to a line count, tests

### Community 43 - "10. One deploy digest for podman units"
Cohesion: 0.33
Nodes (6): 10. One deploy digest for podman units, Decided, Goal, Scope, Today, Undecided

### Community 44 - "Env-file keys"
Cohesion: 0.08
Nodes (25): E10. B's operator pull Secret when the broker's is "existing", E11. B's admin password together with "existing", E12. Memory and CPU: one key name, two unit grammars, E13. Placement keys on Kubernetes, E14. Spread and B's always-on anti-affinity, E15. The Service block on Kubernetes, E16. A gains `docker.compose`; command keys take a list, E17. Which programs `docker.command` and `docker.compose` may name (+17 more)

### Community 45 - "Loop"
Cohesion: 0.37
Nodes (8): N23. `watch.CheckInterval`, Today, Options, CheckInterval(), Loop(), loop(), LoopFrames(), section()

### Community 46 - "limits.go"
Cohesion: 0.20
Nodes (14): sinkSize(), TestParseLimit(), TestParseLimitRefuses(), TestUlimitCheck(), TestUlimitNames(), TestValidCPUSet(), TestValidMemoryAndShmSize(), UlimitNames() (+6 more)

### Community 47 - "go_pkg_strings"
Cohesion: 0.23
Nodes (11): ParseRunAs(), RunAsIDs(), TestParseRunAs(), TestParseRunAsRefuses(), TestRunAsIDs(), TestRunAsRootIsStatedNotEmpty(), 3.1 Sink-owned inputs (not in workload), 3. The description (Go sketch) (+3 more)

### Community 48 - "11. Resolve and pin the kube context"
Cohesion: 0.33
Nodes (6): 11. Resolve and pin the kube context, Decided, Goal, Scope, Today, Undecided

### Community 49 - "13. A watch loop stops on an error no retry can fix"
Cohesion: 0.40
Nodes (5): 13. A watch loop stops on an error no retry can fix, Decided, Goal, Scope, Undecided

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
Cohesion: 0.11
Nodes (19): 2.5 kube.Client (kubectl or oc), Apply, Delete, DeleteNamed (replaces DeletePod), DeleteNamespace, Exec, ExecOutput, Attach, Get, ListLabelled (decided only: gdu 18), Logs (+11 more)

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

### Community 64 - "Shared deployment API (design note)"
Cohesion: 0.12
Nodes (17): 11. Answers to the detail questions, 1. Direction, 2. What each tool repeats today, 3.1 One client per command family, 3.2 Output and results, 3.3 Rules for an executor, 3.5 What stays in each tool, 3.6 Example (+9 more)

### Community 66 - "runnertest_test.go"
Cohesion: 0.35
Nodes (14): runner/runnertest, runnertest_test.go, cmd(), files(), TestCallHasAndString(), TestFakeAttach(), TestFakeConcurrentUse(), TestFakeCopiesItsInputs(), TestFakeFailOn() (+6 more)

### Community 67 - "2.2 containercli.Client (docker, podman, nerdctl engines and compose)"
Cohesion: 0.12
Nodes (16): 2.2 containercli.Client (docker, podman, nerdctl engines and compose), ComposeDown, ComposeUp, ComposeVersion (new, ruling 6), CopyTo, CopyFrom (new, ruling 6), Exec, ExecOutput, Attach (ruling 5), Info, Inspect (+8 more)

### Community 68 - "4. One table per field"
Cohesion: 0.14
Nodes (14): ValidHostName(), ValidTimezone(), 4. One table per field, compose.File.Project, quadlet.Unit.Description, Rootful, Header, Spec.Entrypoint and Spec.Command, Spec.HostName, Spec.Labels and Spec.Digest (+6 more)

### Community 69 - "FieldError"
Cohesion: 0.14
Nodes (22): ParseSeconds(), seconds(), limitValue(), ParseLimit(), refusedLimits(), sizeCheck(), ValidCPUSet(), ValidMemory() (+14 more)

### Community 70 - "The compose file"
Cohesion: 0.14
Nodes (14): C10. Inline text that does not end in exactly one newline, C12. Named volumes and other networks, C1. One order of keys in the file, C2. B's word "bridge" would move every deploy to docker's shared bridge, C3. Read-only root and how compose hands over secrets, C4. Does the project network carry labels, C5. A secret or config target must not sit under a mount, C6. Secret variable names: duplicates and names the child process reads (+6 more)

### Community 71 - "The quadlet unit"
Cohesion: 0.14
Nodes (14): No decision needed, Q10. First line of the unit, Q11. `WantedBy` follows rootful or rootless, Q1. How B gets its `[Service]` limit lines, Q2. The compare of the installed unit stays beside the digest label, Q3. Line order inside the unit, Q4. Which lines are written quoted, and what is escaped where, Q5. File mode of the installed unit: 0600 or 0644 (+6 more)

### Community 72 - "Shared by all three renderers"
Cohesion: 0.14
Nodes (14): S10. Check at load as well as at render, S11. How `generate` writes its output file, S12. What a host name may contain, S13. What a time zone may contain, S2. Which settings get their own field, S3. A stated root user against a forgotten one, S4. The `solution` and `developed-by` labels, S5. Quotes on keys that are data (+6 more)

### Community 73 - "1. Tool A: solmq-conn-util"
Cohesion: 0.15
Nodes (13): 1. Tool A: solmq-conn-util, 2. Tool B: solace-util, 3. Moves, 4. Questions the flatten raises, A: docker, A: image, A: kubernetes, A: podman (+5 more)

### Community 74 - "Fake"
Cohesion: 0.29
Nodes (7): CR-0016 [minor] [new] CI never runs the tests under the race detector, N72. Test helpers beyond `Fake`, FailOn(), Seq(), Call, Fake, Reply

### Community 75 - "rootless.go"
Cohesion: 0.25
Nodes (8): N41. The sub-id refusal: one function, and the range size, N42. `CheckIDMap`: typed error, and both maps in one call, CheckIDMap(), covers(), NextFreeRange(), TestCheckIDMap(), TestNextFreeRange(), TestSubIDFix()

### Community 76 - "ParseImage"
Cohesion: 0.18
Nodes (13): isAlnum(), ParseImage(), TestParseImage(), TestParseImageRefuses(), validDigest(), validRegistry(), validRepository(), validTag() (+5 more)

### Community 78 - "2.3 podman.Client"
Cohesion: 0.22
Nodes (9): 2.3 podman.Client, CreateSecret, IDMap, RemoveSecrets, Run, Output, SecretIDs (decided only: gdu 10), SecretLabels (decided only: gdu 16), SecretMatches (+1 more)

### Community 79 - "2.4 systemd.Client"
Cohesion: 0.22
Nodes (9): 2.4 systemd.Client, DaemonReload, Linger, EnableLinger, LoadState (decided only: gdu 9), NRestarts, Run, Output, Show (new, ruling 6), Start, Stop, Restart (+1 more)

### Community 80 - "One container description and one writer (consolidation draft)"
Cohesion: 0.22
Nodes (9): 1. Purpose, 5.1 YAML sinks (compose, Kubernetes), 5.2 Quadlet lines (Q3, Q4), 5.3 Compose order (C1), 5.4 Kubernetes order, 5.5 Errors (S8), 5. The writer, 6. Cross-field checks (+1 more)

### Community 81 - "1. Exported surface"
Cohesion: 0.22
Nodes (9): 1. Exported surface, Verifiers for the cli group, Verifiers for the containerart group, Verifiers for the containercli group, Verifiers for the docs group, Verifiers for the guards group, Verifiers for the kube group, Verifiers for the podman-host group (+1 more)

### Community 82 - "7. Coverage of what A and B write today"
Cohesion: 0.25
Nodes (8): 7.1 A, compose (internal/dockergen/dockergen.go), 7.2 A, quadlet (internal/podmangen/podmangen.go), 7.3 A, Kubernetes (internal/deploy/deploy.go), 7.4 B, compose (internal/render/render.go), 7.5 B, quadlet (internal/render/render.go), 7.6 B, Kubernetes (internal/k8s, internal/render), 7.7 Lines no field covers, 7. Coverage of what A and B write today

### Community 83 - "CheckCore"
Cohesion: 0.22
Nodes (10): N43. The core-limit checks: common or solace-util's own, CheckCore(), HardLimit(), LimitText(), origin(), ParseLimit(), TestCheckCore(), TestLimitText() (+2 more)

### Community 84 - "Decisions on the 2026-11-11 critical review"
Cohesion: 0.17
Nodes (12): Checks the owner runs, Decided 2026-11-15, Decisions on the 2026-11-11 critical review, Decisions on the plan review, Labels (gdu 1, 16, 18), Other items, Readings the owner confirmed (2026-11-15), Releases (+4 more)

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
Cohesion: 0.33
Nodes (6): 1. Ownership labels on every artifact, Decided, Goal, Scope, Undecided, Until then

### Community 92 - "go-deploy-utility"
Cohesion: 0.33
Nodes (6): Development, go-deploy-utility, Licence, Packages, Status, Using it

### Community 93 - "Widening"
Cohesion: 0.60
Nodes (5): HostNetwork(), Widening(), What each version adds, 3.3 Renderers, N36. `artifacttest.HostNetwork`

### Community 94 - "8. Rendering in the module"
Cohesion: 0.40
Nodes (5): 8.1 Today, 8.2 Proposal, 8.3 What stays in the tool, 8.4 Effects, 8. Rendering in the module

### Community 95 - "health_test.go"
Cohesion: 0.46
Nodes (6): healthy(), TestHealthCheck(), TestHealthQuadletRefusesShellCharacters(), TestHealthRetriesEchoesTheNumber(), TestParseSeconds(), health_test.go

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
Cohesion: 0.50
Nodes (4): 12. A bound on every delete, Goal, To do, Undecided

### Community 118 - "exitcode"
Cohesion: 0.67
Nodes (3): exitcode, exitcode: placement, exitcode: signatures

### Community 119 - "E25. The list of inherent differences"
Cohesion: 0.67
Nodes (3): E25. The list of inherent differences, No decision needed, Settled by the rulings

### Community 120 - "K18. Kinds and fields nobody uses yet"
Cohesion: 0.67
Nodes (3): K18. Kinds and fields nobody uses yet, No decision needed, Settled by the rulings

### Community 123 - "Critical review - go-deploy-utility"
Cohesion: 0.25
Nodes (7): Accepted (not re-litigated), Coverage, Critical review - go-deploy-utility, Dynamic checks, Fixed since last report, Summary, Top risks

### Community 124 - "TestParsePortRefuses"
Cohesion: 0.60
Nodes (4): TestParsePort(), TestParsePortEchoesOnlyNumbers(), TestParsePortRefuses(), port_test.go

## Knowledge Gaps
- **3 isolated node(s):** `github.com/hafio/go-deploy-utility/docs/review`, `github.com/hafio/go-deploy-utility`, `NO_COLOR`
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 541 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **9 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `To do` connect `To do` to `paritydoc_test.go`, `Handoff: aligning solmq-conn-util and solace-util on Docker, Podman and Kubernetes`, `kube.go`, `10. One deploy digest for podman units`, `11. Resolve and pin the kube context`, `13. A watch loop stops on an error no retry can fix`, `14. One check for the log tail count`, `15. Dev scripts build the pinned module unless asked for the checkout`, `17. A readiness wait for OnDelete StatefulSets`, `18. Kubernetes objects a config no longer names are removed`, `19. One shape for the platform sections of the env file`, `3. One step order for deploy and remove on every platform`, `6. Every artifact states its network mode`, `7. One atomic write for files a tool installs`, `8. One gate for the test catalogue`, `9. Remove is idempotent on every platform`, `2. One guideline for expansion in config values`, `4. Builders refuse a name outside its grammar`, `todo.md`, `16. One behaviour for stale podman secrets`, `1. Ownership labels on every artifact`?**
  _High betweenness centrality (0.173) - this node is a cross-community bridge._
- **Are the 155 inferred relationships involving `6.3 Inventory` (e.g. with `Entry` and `.Entries()`) actually correct?**
  _`6.3 Inventory` has 155 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/hafio/go-deploy-utility/docs/review`, `github.com/hafio/go-deploy-utility`, `NO_COLOR` to the rest of the system?**
  _3 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Tests by package` be split into smaller, more focused modules?**
  _Cohesion score 0.08 - nodes in this community are weakly interconnected._
- **Why does `Runner` connect `Handoff: aligning solmq-conn-util and solace-util on Docker, Podman and Kubernetes` to `testing.T`, `noticetest_test.go`, `paritydoc_test.go`, `exitcode_test.go`, `nsgate.go`, `No decision needed`, `kube.go`, `inspect.go`, `rbac.go`, `containercli.go`, `Findings by area`, `Echo`, `Cmd`, `Loop`, `Shared deployment API (design note)`, `runnertest_test.go`, `ParseImage`, `CheckCore`, `Decisions on the 2026-11-11 critical review`, `To do`?**
  _High betweenness centrality (0.165) - this node is a cross-community bridge._
- **Are the 119 inferred relationships involving `No decision needed` (e.g. with `Entry` and `.Entries()`) actually correct?**
  _`No decision needed` has 119 INFERRED edges - model-reasoned connections that need verification._
- **Should `testing.T` be split into smaller, more focused modules?**
  _Cohesion score 0.0720687079910381 - nodes in this community are weakly interconnected._