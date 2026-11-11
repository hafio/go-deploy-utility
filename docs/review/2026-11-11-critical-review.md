# Critical review - go-deploy-utility

Date: 2026-11-11
Commit: 83fa381 on main
Previous report: none
Scope: all areas, this module only. Each finding carries a Downstream line naming the call sites it reaches in solmq-conn-util (A, `../solace-ibmmq-connector-helper`, main@298cccf) and solace-util (B, `../solace-cnt-scripts`, main@b0941b1). Neither tool was reviewed.
Stacks: Go (core + go.md), shell (bash.md), GitHub Actions (github-actions.md), containers.md runtime-posture and sink-grammar items applied to the `containerart` emitters; PowerShell with the core checklist only
Graph: reused; 956 nodes, 3158 links, 27 communities
Dynamic checks: none run by the review; the user-run gate logs of 2026-11-11 16:17 were read; six checks offered

## Summary

| Severity | New | Still open | Fixed since last | Total open |
|---|---|---|---|---|
| Blocker | 0 | 0 | 0 | 0 |
| Major | 2 | 0 | 0 | 2 |
| Minor | 11 | 0 | 0 | 11 |
| Nit | 10 | 0 | 0 | 10 |
| Suspected (uncounted) | 2 | 0 | 0 | 2 |

The module is in good shape: the guards are careful, errors say what to do, the
house rules hold (no globals beyond the documented embed, no `os.Exit`, plain
ASCII, LF), and the 16:17 gates are green at 99.7% coverage. Two things matter
most. The namespace gate judges a namespace from a listing that cannot see most
kinds, so it can offer, and under `--no-prompt` perform, a delete that destroys
other owners' objects (CR-0003). And the real-engine smoke checks for S7 to S9
are recorded as never run although v0.1.2 is tagged and both tools pin it
(CR-0009); two suspected findings are exactly what those checks would settle.
The rest is guard blind spots in the shared test helpers, script and CI drift
from the dev-script contract, and unused surface. Both tools pin v0.1.2, which
is this code, so every behaviour fix needs a new patch tag and a repin in both.
First run, so there is no trend.

## Top risks

1. CR-0003 [major] Namespace gate lists only all, PVCs, Secrets and ConfigMaps - `kube/nsgate/nsgate.go:47-57` - a namespace holding only other kinds is deleted with everything in it.
2. CR-0009 [major] Real-engine smoke checks for S7 to S9 recorded as not run - `docs/smoke.md:25` - redeploy, watch and the namespace gate ship on fake-runner evidence.

## Findings by area

### design

#### CR-0001 [minor] [new] Exported API with no consumer in either tool

Location: watch/watch.go:83-88, containercli/containercli.go:168-169, paritydoc/paritydoc.go:31-51, runner/runnertest/runnertest.go:164-173, confirm/prompter.go:145-149, execguard/execguard.go:53-58, scripts/dev-all.sh:83-100
Confidence: verified
Evidence:

    watch/watch.go:83                    func Loop(ctx context.Context, o Options, refresh func(now time.Time) error) error {
    containercli/containercli.go:169     func Remove(base []string, name string) runner.Cmd { return cmd(base, "rm", name) }
    paritydoc/paritydoc.go:32            func Body() []byte { return []byte(body) }
    paritydoc/paritydoc.go:36            func Compose(noticeBlock string) ([]byte, error) {
    runner/runnertest/runnertest.go:166  func FailOn(sub string, err error) func(int, Call) Reply {
    confirm/prompter.go:146              func IsNotTerminal(err error) bool {
    execguard/execguard.go:58            const ApproveFlag = "--allow-command"
    scripts/dev-all.sh:90                { head -n "$end" "$doc"; cat "$HERE/paritydoc/body.md"; } >"$tmp"

    grep over both tools, tests included: 0 uses of each of the seven.

Why it matters: docs/developer.md:25 says shared code exists only when both tools use it. Unused exports are surface that still has to stay compatible, and two already have a second implementation that can drift: `dev-all` composes `docs/parity.md` in bash and again in PowerShell instead of through `Compose`, and both tools spell the flag themselves while execguard's messages print `ApproveFlag`.
Downstream: A `cmd/solmq-conn-util/commands.go:27` and `main.go:1593`, B `internal/cli/flags.go:239` hard-code `--allow-command`. No call site uses the other six.

Change: remove the seven, or record beside each why it stays (`docs/consuming.md:71-72` keeps `Loop` on purpose). Have each tool register its flag from `execguard.ApproveFlag`. Removal is an API change: new tag, repin both tools.

#### CR-0002 [minor] [new] execguard accepts a flag value that makes the allowed binary run a file beside the env file

Location: execguard/execguard.go:9-11, execguard/execguard.go:23-27, execguard/execguard.go:378-380
Confidence: verified
Evidence:

    9   // What it defends: an unlisted binary; a path form pointing at a file shipped
    10  // beside the config; a subcommand smuggled ahead of the one the tool appends; a
    11  // privilege-escalation wrapper in front of the tool's own commands.
    378 		case strings.HasPrefix(cmd[i-1], "-") && !strings.Contains(cmd[i-1], "="):
    379 			// The value of the preceding flag. Unvalidatable beyond the charset (see
    380 			// the package doc).

    execguard/execguard_test.go:55 pins the accept: {"kubectl", "--kubeconfig", "/tmp/kc"}.

Why it matters: `/` and `.` pass the charset for every token after argv[0], so `docker --config ./d` and `kubectl --kubeconfig ./kc` pass `Check`. A docker config directory holds `cli-plugins/docker-compose`, and a kubeconfig can name an exec credential plugin, so the allowed binary then runs a file the env file chose: the second item on the "defends" list, reached through a flag value. The package doc states the flag-value limit but names only the subcommand consequence, and `TestFlagValuePositionIsNotGuaranteed` pins only that one.
Downstream: A `internal/validate/validate.go:64-70` (rules for docker, podman, kubernetes) and `internal/runner/runner.go:107`; B `internal/config/execguard.go:46-83` (docker, docker-compose, nerdctl, podman, cluster CLI) and `:168,252,278,288,301`. Both tools' trust-model docs inherit the same gap.

Change: state this consequence in the package doc beside the boolean-flag case and pin it in the test. Consider a `Rules.DenyFlags` field so a tool can refuse flags no deployment needs (`--config` on docker, `--runtime`, `--conmon` and `--module` on podman) while `--kubeconfig` stays documented as trusted input. Nothing changes for a tool until it sets the field.

### correctness

#### CR-0003 [major] [new] Namespace gate lists only all, PVCs, Secrets and ConfigMaps, so a namespace holding other kinds is offered for deletion

Location: kube/nsgate/nsgate.go:47-57, kube/nsgate/nsgate.go:136-149, paritydoc/body.md:71-76
Confidence: verified
Evidence:

    kube/nsgate/nsgate.go
    55  	return runner.Cmd{Argv: append(argv, "get", "all,persistentvolumeclaims,secrets,configmaps", "-n", ns,
    56  		"-o", "json")}
    141 	case it.Kind == "Event" || it.Kind == "LimitRange" || it.Kind == "ResourceQuota":
    143 	case it.Kind == "ConfigMap" && m.Name == "kube-root-ca.crt",
    144 		it.Kind == "ServiceAccount" && m.Name == "default":

    paritydoc/body.md
    71  9. **An emptied namespace is offered for deletion, never assumed empty.** After a successful
    76     Anything else keeps the namespace and is named. [...]
    78     `--no-prompt` approves, and without a terminal the namespace is kept. [...]

Why it matters: `Occupants` judges only what `List` returns, and kubectl's `all` category leaves out Ingress, NetworkPolicy, Role, RoleBinding, ServiceAccount, PodDisruptionBudget and every custom resource. A namespace holding only those reads as empty: the delete is offered, `--no-prompt` approves it with no question, and deleting the namespace destroys them. Lines 141 and 144 discount four kinds this listing never returns, which shows the listing is narrower than the rule assumes; rule 9's "anything else keeps the namespace" does not hold. Reaching it needs a shared or pre-provisioned namespace and an identity that may delete namespaces. The step rests on what kubectl's `all` covers, which no test here exercises (see CR-0009).
Downstream: A `cmd/solmq-conn-util/namespace.go:98-148` (prompt skipped at `:129` under `--no-prompt`, delete at `:142`) with the listing at `internal/runner/runner.go:871` and the delete at `:279`; B `internal/cli/ops_k8s.go:790-824`, `internal/k8s/namespace.go:57,67` and `internal/k8s/prep.go:129`. Both tools ship it at v0.1.2.

Change: either list what is really there, or stop inferring ownership from emptiness.
- Wider listing: add a builder for `api-resources --verbs=list --namespaced -o name` and list that set; a type whose list is refused keeps the namespace, as a failed listing already does. `Occupants` then needs discounts for the Endpoints of an owned Service and for metrics objects, and its four dormant discounts become live.
- Ownership stamp: have `NamespaceManifest` write a label, and offer deletion only for a namespace that carries it, so a namespace the tool did not create is never offered.
Either way rewrite parity rule 9 and use an Ingress or RoleBinding as the foreign object in smoke S9 row 5. Behaviour change: new patch tag, repin both tools.

#### CR-0004 [minor] [new] Secret wiring writes target, env var and secret name into compose and quadlet unchecked

Location: containerart/compose/compose.go:98-101, containerart/compose/compose.go:113-116, containerart/quadlet/quadlet.go:64-73, containerart/quadlet/quadlet_test.go:91-102
Confidence: verified
Evidence:

    containerart/compose/compose.go
    99  		b.WriteString("      - source: " + Quote(s.Name) + "\n")
    100 		b.WriteString("        target: " + s.Target + "\n")
    115 		b.WriteString("    environment: " + s.EnvVar + "\n")

    containerart/quadlet/quadlet.go
    72  	return "Secret=" + EscapePercent(name) + ",type=mount,target=" + EscapePercent(target)

    containerart/quadlet/quadlet_test.go
    93  		{"solace-admin-password", "/mnt/secrets/username_admin_password",
    95  		{"a%b", "/run/app/%t", "Secret=a%%b,type=mount,target=/run/app/%%t"},

Why it matters: a newline, a comma or `: ` in any of the three adds a key or an option to the rendered artifact, and nothing in the module refuses it. `Environment` has an injection test (`TestEscapeWritesControlsAsCEscapes`); `Secret`, `ServiceSecrets` and `TopSecrets` have none, so the guarantee rests wholly on each tool's load-time validation and a loosened rule there is not caught here.
Downstream: A `internal/dockergen/dockergen.go:135-141` (name doubles as env var; target is a constant path plus the name) and `internal/podmangen/podmangen.go:109`; B `internal/render/render.go:997` and `:1181-1186`, where the name derives from `container.name`. B closes it today with `^[A-Za-z0-9][A-Za-z0-9._-]*$` at `internal/config/validate.go:404`. No live path was found in either tool.

Change: write the compose target through `Quote` (`artifacttest.go:285` already strips quotes), hold `EnvVar` to `envref.ValidName`, and have the quadlet line refuse a name or target with a comma, whitespace or control character. Add the injection rows to `TestSecretLine`, `TestServiceSecrets` and `TopSecrets`. The compose bytes change, so both tools' goldens move with the repin.

### resilience

No findings. Read: `runner` cancel, grace and `classify` paths, `Attach` exit handling, the `watch` retry loop, `confirm` end-of-input paths, `rootless` and `systemd` probe parsing.

### quality

#### CR-0007 [nit] [new] cmd, stdin, shQuote and Upload are duplicated across the builder packages

Location: containercli/containercli.go:47-60, containercli/containercli.go:273-293, kube/kube.go:28-41, kube/kube.go:226-248, podman/podman.go:24-29
Confidence: verified
Evidence:

    containercli/containercli.go:48  func cmd(base []string, args ...string) runner.Cmd {
    kube/kube.go:29                  func cmd(base []string, args ...string) runner.Cmd {
    podman/podman.go:25              func cmd(base []string, args ...string) runner.Cmd {
    containercli/containercli.go:274 func shQuote(s string) string { return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'" }
    kube/kube.go:227                 func shQuote(s string) string { return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'" }

Why it matters: `cmd` is identical in three packages, `stdin` and `shQuote` in two, and the two `Upload` bodies differ only in the `Exec` they call. A fix to the copy or quoting rule has to be made in each copy.
Downstream: none; the helpers are unexported.

Change: move the argv copy and the non-nil stdin copy into `runner`, and share the upload script builder.

#### CR-0008 [nit] [new] dev-all discovers sibling tasks by scraping their dev scripts with two different regexes

Location: scripts/dev-all.sh:56-62, scripts/dev-all.ps1:60-68
Confidence: verified
Evidence:

    scripts/dev-all.sh
    59    grep -oE '^task_[a-z_]+\(\)' "$script" | sed -E 's/^task_([a-z_]+)\(\)$/\1/'
    60    grep -oE '^[[:space:]]*[a-z]+\)[[:space:]].*(ALL|FULL|tasks)' "$script" |

    scripts/dev-all.ps1
    65    $words += [regex]::Matches($text, '(?m)^function Task-([a-z]+)') | ForEach-Object { $_.Groups[1].Value }
    66    $words += [regex]::Matches($text, '''([a-z]+)''\s*\{\s*\$queue') | ForEach-Object { $_.Groups[1].Value }

Why it matters: a task name with an underscore is found by bash and cut short by PowerShell, and an aggregate reformatted in a tool turns into "no repo's dev.sh knows" with no hint why.
Downstream: the `dev.sh` and `dev.ps1` of A and B are the scraped input; a reformat there breaks this script silently.

Change: document the shape the scrapers need in docs/developer.md, or have each dev script answer a `tasks` word and ask it instead.

### tests

#### CR-0009 [major] [new] Real-engine smoke checks for S7 to S9 are recorded as not run for a tagged release

Location: docs/smoke.md:25, docs/smoke.md:49, docs/smoke.md:60, docs/smoke.md:83-86
Confidence: verified
Evidence:

    11  Hand checks for each adoption step, run against real engines once the gates pass. Unit tests mock
    12  the engines, so only a real host proves this behaviour. [...]
    25  ## S7: redeploy asks before any write (not yet run)
    49  ## S8: the shared watch loop (not yet run)
    60  ## S9: kube, rbac and the namespace gate (not yet run)
    85  | Date | Host | docker / compose | podman | Steps | Result |
    86  | --- | --- | --- | --- | --- | --- |

Why it matters: v0.1.0 to v0.1.2 are tagged and both tools pin v0.1.2 on main, so the redeploy with secret comparison, the watch loop and the namespace gate ship with fake-runner evidence only, or with evidence nobody recorded. CR-0003, CR-0005 and CR-0006 all turn on what a real kubectl, podman or systemd does.
Downstream: A main@298cccf and B main@b0941b1 both carry the adopted behaviour.

Change: run S7 to S9 on a docker host, a rootless podman host and a cluster, tick the boxes and fill the Runs table, or record which rows were run and when. Add rows for a foreign Ingress or RoleBinding (CR-0003), for the secret rows of `status` on podman (CR-0005), and for a `${NAME}` in a podman environment value (CR-0006).

#### CR-0010 [minor] [new] Widening and HostNetwork miss the podman flag spellings that ride PodmanArgs

Location: containerart/artifacttest/artifacttest.go:32-50, containerart/artifacttest/artifacttest.go:56-84, containerart/artifacttest/artifacttest.go:181-182
Confidence: verified
Evidence:

    34  		"privileged: true", "--privileged", "Privileged=true",
    36  		"cap_add", "AddCapability", "SYS_ADMIN",
    37  		"devices:", "AddDevice",
    38  		"sysctls:", "Sysctl=",
    41  		"pid: host", "ipc: host",
    49  	return []string{"network_mode: host", "Network=host"}
    65  		case strings.Contains(flat, t):
    181 		{[2]int{4, 4}, []string{"Image", "ContainerName", "User", "Group", "NoNewPrivileges", "DropCapability",
    182 			"PodmanArgs", "Network", "PublishPort", "Volume", "Environment", "Label"}},

Why it matters: `PodmanArgs` is an accepted key, and the match is a case-sensitive substring. `PodmanArgs=--cap-add=NET_ADMIN`, `--device`, `--sysctl`, `--pid=host`, `--ipc=host` or `--network=host` widens the container and passes every posture assertion; so do truthy spellings such as `privileged: True`. The helper is the regression guard both tools rely on.
Downstream: A `internal/gen/gen_runas_test.go` and B `internal/render/render_test.go` call `AssertAbsent` with these lists. B writes `PodmanArgs` at `internal/render/render.go:950,953`; its cpuset value is regex-checked at load (`internal/config/validate.go:425`), so this is a guard gap, not a live widening.

Suggested patch (not applied):

```diff
--- a/containerart/artifacttest/artifacttest.go
+++ b/containerart/artifacttest/artifacttest.go
@@ -39,5 +39,7 @@ func Widening() []string {
 		"SecurityLabelDisable", "label=disable", "label:disable",
 		"SeccompProfile=", "unconfined", "Unconfined",
 		"pid: host", "ipc: host",
+		"--cap-add", "--device", "--sysctl",
+		"--pid=host", "--pid host", "--ipc=host", "--ipc host", "--userns=host", "--userns host",
 	}
 }
@@ -46,5 +48,6 @@ func Widening() []string {
 // solmq-conn-util forbids it in every artifact. solace-util writes it only when
 // the operator asks, and checks that no other shape carries it.
 func HostNetwork() []string {
-	return []string{"network_mode: host", "Network=host"}
+	return []string{"network_mode: host", "Network=host",
+		"--network=host", "--network host", "--net=host", "--net host"}
 }
```

Also add the rows to `TestCheckAbsent`, and either fold case in `squeeze` or say in the doc comment that case variants are not caught. Both tools' posture tests re-run at the repin.

#### CR-0011 [minor] [new] CheckSecretTargets passes a compose secrets block it does not recognise

Location: containerart/artifacttest/artifacttest.go:217-223, containerart/artifacttest/artifacttest.go:246-270
Confidence: verified
Evidence:

    219 	for _, item := range composeSecretItems(strings.Split(artifact, "\n")) {
    257 		case l == "    secrets:":
    258 			in, cur = true, -1
    260 		case indent <= 4:
    261 			in, cur = false, -1
    262 		case indent == 6 && strings.HasPrefix(t, "- "):

Why it matters: a block opens only on a line that is exactly four spaces and `secrets:`. A flow list (`    secrets: [name]`), a sequence at the key's own indent (closed by line 260 before any item is seen) or a CRLF file yields no items, and the function returns nil. The short syntax the check exists to refuse therefore passes in flow form, and so does any renderer whose layout shifts.
Downstream: A `internal/gen/gen_runas_test.go` and B `internal/render/render_test.go` call `AssertSecretTargets` twice each.

Change: normalise CRLF, and report any line at service indent that starts with `secrets:` but is not the recognised block form; add tests for the flow list, the same-indent sequence and CRLF.

#### CR-0012 [minor] [new] The Windows console branch of CanRedraw is never executed by a test

Location: watch/vt_windows.go:16-29, watch/watch_test.go:425-455
Confidence: verified
Evidence:

    watch/vt_windows.go
    20  	if err := syscall.GetConsoleMode(h, &mode); err != nil {
    21  		return false
    23  	if mode&enableVirtualTerminalProcessing != 0 {
    26  	setConsoleMode := syscall.NewLazyDLL("kernel32.dll").NewProc("SetConsoleMode")
    27  	r, _, _ := setConsoleMode.Call(uintptr(h), uintptr(mode|enableVirtualTerminalProcessing))

    watch/watch_test.go
    451 		if want := runtime.GOOS != "windows"; CanRedraw(null) != want {

Why it matters: the test reaches `enableVT` only with the null device, where line 20 fails. Lines 23 to 28, which decide whether a PowerShell 5.1 console gets a redraw or escape bytes as text, run in no test. The 99.7% total is measured on Linux and does not include this file, so the number hides the gap; smoke S8 is the only check and it is unrun (CR-0009).
Downstream: A `cmd/solmq-conn-util/status.go:381` and B `internal/cli/watch.go:131` pass the result straight to `Options.Redraw`.

Change: add a Windows-only test that opens `CONOUT$` and skips when there is no console, or record in docs/test.md that this branch is smoke-only and that the coverage total is Linux-only.

#### CR-0013 [nit] [new] CheckDocs passes when Exclude removes every selected doc

Location: notice/noticetest/noticetest.go:130-148
Confidence: verified
Evidence:

    117 		if len(matches) == 0 {
    118 			return noDocsError{pattern}
    135 	for _, rel := range files {
    136 		if skip[rel] {
    137 			continue
    148 	return errors.Join(errs...)

Why it matters: the package's own rule (line 100) is that a gate checking no docs proves nothing, yet the empty case is refused only for the default glob.
Downstream: A `cmd/solmq-conn-util/support_test.go` and B `internal/cli/commanddoc_test.go` call `AssertDocs`; neither excludes everything today.

Change: count the docs actually checked and return `noDocsError` when the count is zero.

#### CR-0014 [nit] [new] TestBodyShape pins the parity body to a line count

Location: paritydoc/paritydoc_test.go:57-65
Confidence: verified
Evidence:

    57  	if len(lines) != 228 {
    58  		t.Errorf("body has %d lines, want 228; update this count when the body changes on purpose", len(lines))
    63  	if lines[10] != "## Contract" {

Why it matters: a change detector. Every deliberate body edit needs a test edit, and the count guards nothing the ASCII and final-newline checks above it do not.
Downstream: none.

Change: drop the count and the fixed heading index; keep the byte checks.

### cicd

#### CR-0015 [minor] [new] No gate runs before a tag exists, and for a Go module the tag is the release

Location: .github/workflows/ci.yml:3-5, .github/workflows/ci.yml:19-24, .github/workflows/tag.yml:13-16, .github/workflows/tag.yml:38-40, docs/consuming.md:100-103
Confidence: verified
Evidence:

    .github/workflows/ci.yml
    3   # No automatic triggers by design: nothing runs on push or PR. The full stack
    19    pull_request:
    20      paths:
    21        - ".github/workflows/**"
    22        - ".github/dependabot.yml"
    23        - "scripts/dev.sh"
    24        - "scripts/dev.ps1"

    .github/workflows/tag.yml
    14  # to clean up: read the logs artifact, push a fix, tag again. A library is the
    15  # exception: its tag is fetchable the moment it is pushed, and a module proxy
    16  # caches it for good, so a library fix goes out under a NEW tag, never the same.

Why it matters: a change to Go code, `go.mod` or `dev-all` runs no pipeline until a tag is pushed, and by then the version is public and permanent. A failing gate only withholds the notes-only release, so the one blocking gate is the hand-run `full` on two operating systems in the release order. This is a documented trade-off; it is listed because the pipeline cannot stop a bad version. Line 3 also contradicts lines 19-24.
Downstream: `dev-all.sh resolve` would write such a tag into both tools' `go.sum`.

Change: add `**.go`, `go.mod`, `go.sum` and `scripts/dev-all.*` to the pull-request paths, or trigger on pushes to main, and correct the header. `ci.yml` is the init-devops template, so make the change there. If tag-only is the decision, set this finding to `accepted` in the ledger.

#### CR-0016 [minor] [new] CI never runs the tests under the race detector

Location: .github/workflows/ci.yml:50-51, .github/workflows/ci.yml:59-66, scripts/dev.sh:90-99, scripts/dev.ps1:104-112, runner/runnertest/runnertest_test.go:224
Confidence: verified
Evidence:

    scripts/dev.sh
    94    if [ "${DEV_RACE:-}" = "1" ]; then
    95      run test go test -race -count=1 ./...
    96    else
    97      run test go test -count=1 ./...

    .github/workflows/ci.yml
    50      env:
    51        NO_COLOR: "1" # keeps captured logs free of ANSI mojibake
    61          run: ./scripts/dev.sh all scan

Why it matters: `TestFakeConcurrentUse` and the runner's stream tests can catch a data race only under `-race`. As run, a lost mutex in the `Fake` both tools test against passes every gate. The contract asks for `-race` wherever gcc is on PATH, both runners included; the local 16:17 run was without it too.
Downstream: B `internal/cli/watch_test.go` and `internal/engine/runner_test.go` use the `Fake`; A uses it in `internal/runner/runner_test.go`.

Suggested patch (not applied):

```diff
--- a/.github/workflows/ci.yml
+++ b/.github/workflows/ci.yml
@@ -49,4 +49,5 @@
         os: [ubuntu-24.04, windows-2025]
     env:
       NO_COLOR: "1" # keeps captured logs free of ANSI mojibake
+      DEV_RACE: "1" # the contract expects gcc on both runner images
     steps:
```

`ci.yml` is the init-devops template; the contract's own form is to detect gcc in both scripts and say so in `test.log`.

#### CR-0017 [minor] [new] Dev scripts and ci.yml have fallen behind the dev-script contract

Location: .github/workflows/ci.yml:59-66, .github/workflows/tag.yml:8, scripts/dev.sh:19-25, scripts/dev.sh:72-78, scripts/dev.sh:93-99, scripts/dev.ps1:19-26, scripts/dev.ps1:87-93
Confidence: verified
Evidence:

    .github/workflows/tag.yml:8      #     -> gates    build vet test cov scan, ubuntu + windows
    .github/workflows/ci.yml:61              run: ./scripts/dev.sh all scan
    scripts/dev.sh:23                  t="$(sed -n 's/^toolchain //p' "$REPO_ROOT/go.mod")"
    scripts/dev.sh:24                  [ -n "$t" ] && export GOTOOLCHAIN="$t"
    scripts/dev.sh:78                  unformatted=$(gofmt -l .)

    dev-script-contract.md: "CI runs `all cov scan`"; "`test` skips the suite `cov` runs";
    "use the selected toolchain's gofmt, `$(go env GOROOT)/bin/gofmt`"; "no version-detection
    or `GOTOOLCHAIN`-style compensation logic".

Why it matters: the contract is the written spec for these files and they differ in four clauses. `cov` never runs in the pipeline although `tag.yml:8` says it does; `full` runs the suite twice because there is no `COV_QUEUED`; `gofmt` comes from PATH and can be a different release from the pinned toolchain; and both scripts export `GOTOOLCHAIN` from `go.mod`. That last one is what keeps a `go.work` without a `toolchain` line on the pinned toolchain (`docs/consuming.md:48-52`), so it is a deliberate exception nobody recorded.
Downstream: none directly. A and B carry dev scripts of the same family.

Change: reconcile through init-devops rather than by hand, and record the `GOTOOLCHAIN` exception in docs/developer.md if it stays.

### build

#### CR-0018 [minor] [new] cov truncates cov.log before the previous total can be compared

Location: scripts/dev.sh:36-39, scripts/dev.sh:101-109, scripts/dev.ps1:42-46, scripts/dev.ps1:114-126
Confidence: verified
Evidence:

    scripts/dev.sh
    38    printf '=== %s | %s ===\n' "$(now)" "$1" > "$LOG_DIR/$1.log"
    103   # The previous total in logs/cov.log is the floor (local only -- CI is a
    104   # fresh checkout with no prior log). [...]

    scripts/dev.ps1
    44    Set-Content -Path (Get-Log $Task) -Encoding utf8 `

Why it matters: the coverage floor is the last recorded total, and the run that would be compared against it erases it first. A drop is visible only to someone who read the log before starting, so an unexplained fall can pass unnoticed.
Downstream: none.

Change: in both scripts, read the last `total:` line of `cov.log` before truncating, write `previous total: <x>` into the new log and the console, and warn or fail when the new total is lower.

#### CR-0019 [minor] [new] dev.ps1 and dev.sh disagree on cov exit status, task-name matching and timestamp culture

Location: scripts/dev.ps1:121-125, scripts/dev.sh:106-108, scripts/dev.ps1:36-39, scripts/dev.ps1:167-173, scripts/dev.sh:159-162
Confidence: verified
Evidence:

    scripts/dev.ps1
    38    '{0}{1}' -f $d.ToString('yyyy-MM-ddTHH:mm:ss'), ($d.ToString('zzz') -replace ':', '')
    121   go tool cover "-html=coverage.out" -o coverage.html | Out-Null
    122   $total = (go tool cover "-func=coverage.out" | Select-Object -Last 1)
    125   return 0
    168   switch ($t) { 'all' { $queue += $All } 'full' { $queue += $Full } default { $queue += $t } }
    173   if (-not (Get-Command "Task-$task" -ErrorAction SilentlyContinue)) { Die "unknown task: $task" }

    scripts/dev.sh
    108   go tool cover -func=coverage.out | tail -n1 | tee -a "$LOG_DIR/cov.log"
    161   type "task_$task" >/dev/null 2>&1 || die "unknown task: $task"

Why it matters: the two scripts must behave identically. A failed coverage report is OK on Windows and FAILED on Linux (under `pipefail` line 108 returns the tool's status); `dev.ps1 BUILD` works where `dev.sh BUILD` dies, and a wildcard passes the PowerShell lookup; and `:` in a .NET custom format follows the current culture, so the footer is not ISO everywhere. Neither script checks the `-html` step, and both find an unknown task only after earlier tasks have run.
Downstream: none.

Change: check `$LASTEXITCODE` after each `go tool cover` in dev.ps1 and add `|| return $?` after `-html` in dev.sh; format with `InvariantCulture`; match task names exactly and validate the whole queue before the first task in both.

#### CR-0020 [nit] [new] dev-all pin edits both tools with no clean-tree or branch check and can stop half way

Location: scripts/dev-all.sh:102-140, scripts/dev-all.ps1:119-169
Confidence: verified
Evidence:

    scripts/dev-all.sh
    97    cat "$tmp" >"$doc"
    121   if ! (cd "$tool" && go mod edit -require="$MODULE@$version"); then
    122     echo "dev-all: pin: go mod edit failed in $(name "$tool"); nothing more was changed there" >&2
    123     exit 1

Why it matters: the adoption rule is one commit per tool on `adopt-go-deploy-utility`. `pin` writes into whatever branch and state it finds, and a failure in the second tool exits after the first is already pinned and re-synced.
Downstream: `go.mod` and `docs/parity.md` of A and B are the files written.

Change: before the first write, refuse a tool whose `go.mod` or `docs/parity.md` is modified (`git -C <tool> status --porcelain`), and warn when its branch is not the adoption branch.

#### CR-0021 [nit] [new] dev-all.sh needs bash 4 but does not check for it

Location: scripts/dev-all.sh:1, scripts/dev-all.sh:214-217
Confidence: verified
Evidence:

    1   #!/usr/bin/env bash
    34  set -uo pipefail
    214 declare -A KNOWN

Why it matters: on bash 3.2, which macOS ships, `declare -A` fails; with no `-e` the script carries on with a plain array and the word check misbehaves instead of stopping with a clear message.
Downstream: none.

Change: test `BASH_VERSINFO` at the top and stop with a message, or replace the associative array with two parallel indexed arrays.

#### CR-0022 [nit] [new] The module zip carries graphify-out and docs/handoff, 77 percent of its bytes

Location: .gitignore:16-19, graphify-out/graph.json:1, graphify-out/graph.html:1
Confidence: verified
Evidence:

    .gitignore
    16  # graphify dated snapshots and per-machine state (the live graph stays tracked)

    git ls-files sizes: 3,013,989 bytes tracked in total; graphify-out/ and docs/handoff/
    account for 2,329,101 of them.

Why it matters: a Go module zip holds every tracked file outside a nested module. Every download of a tag fetches and caches the graph and the planning history, and the graph changes on every refresh.
Downstream: A and B fetch it at `dev-all resolve`, on each developer machine and in their own CI.

Change: add a one-line `go.mod` inside `graphify-out/` (and `docs/handoff/` if wanted). The directory stays in git and drops out of the zip from the next tag.

### interfaces

#### CR-0023 [nit] [new] Builders in one package disagree on whether an empty name is refused

Location: containercli/containercli.go:157-169, containercli/containercli.go:254-262, containercli/containercli.go:304-319, podman/podman.go:71-83, podman/podman.go:96-101
Confidence: verified
Evidence:

    containercli/containercli.go
    160 func Start(base []string, name string) runner.Cmd { return cmd(base, "start", name) }
    256 	case name == "":
    257 		return runner.Cmd{}, errors.New("containercli: exec needs a container name")

    podman/podman.go
    71  func CreateSecret(base []string, name string, value []byte, replace bool, labels ...string) runner.Cmd {
    97  	if name == "" {
    98  		return runner.Cmd{}, errors.New("podman: reading a secret's value needs the secret's name")

Why it matters: `Exec`, `Upload` and `SecretData` return an error for an empty name; `Start`, `Stop`, `Restart`, `Remove`, `Logs`, `CreateSecret` and `RemoveSecrets` hand back a command with an empty argument. Two contracts in one package, and the signature alone does not say which a builder follows.
Downstream: A `internal/runner/runner.go:292-296,502,1139`; B `internal/container/manager.go:612,1178-1202,1401`. Both tools validate names at load, so this is a consistency point.

Change: pick one rule for name-taking builders and state it in the package doc.

### observability

#### CR-0024 [nit] [new] inspect.Printable keeps invisible format characters that execguard refuses

Location: containercli/inspect/inspect.go:180-191, execguard/execguard.go:187-193
Confidence: verified
Evidence:

    containercli/inspect/inspect.go
    186 		if unicode.IsControl(r) {
    187 			return -1

    execguard/execguard.go
    193 func IsInvisible(r rune) bool { return unicode.Is(unicode.Cf, r) }

Why it matters: execguard refuses category Cf because a bidi override or zero-width character defeats reading a log line. `Printable` exists for the same reason, for text that comes from the engine, and lets those through.
Downstream: A `internal/statusreport/parse.go:569-646` and B `internal/container/inspect.go:119-172` print engine text through `Printable`.

Suggested patch (not applied):

```diff
--- a/containercli/inspect/inspect.go
+++ b/containercli/inspect/inspect.go
@@ -184,7 +184,7 @@ func health(spellings ...*Health) Health {
 func Printable(s string) string {
 	return strings.Map(func(r rune) rune {
-		if unicode.IsControl(r) {
+		if unicode.IsControl(r) || unicode.Is(unicode.Cf, r) {
 			return -1
 		}
 		return r
 	}, s)
```

### conventions

#### CR-0025 [nit] [new] Developer guide statements that no longer match the code

Location: docs/developer.md:171-174, docs/developer.md:256-258, kube/kube.go:11-13, containerart/artifacttest/artifacttest.go:30-31
Confidence: verified
Evidence:

    docs/developer.md
    172   podman release that introduced it. The Kubernetes spellings
    173   (`allowPrivilegeEscalation: true`, `hostNetwork`, `hostPID`, `hostIPC`) stay
    174   in solmq-conn-util until B7.
    256   `-n <ns>`. They trust their names: a tool checks every namespace and object
    257   name with `kube.ValidLabel` at load, since a name starting with `-` would be

    containerart/artifacttest/artifacttest.go
    30  // Host networking is not in it (see HostNetwork), and neither are Kubernetes
    31  // spellings. Each call returns a new slice.

Why it matters: B7 is built and the Kubernetes spellings did not move, and solace-util never calls `kube.ValidLabel`. The docs rule is no drift.
Downstream: B validates names with its own `dnsLabelBodyRE` at `internal/config/validate.go:443`, a second definition of the rule `kube.ValidLabel` holds.

Change: say where the Kubernetes spellings live now, and either have B adopt `kube.ValidLabel` or word the two sentences as "each tool checks".

## Suspected (needs confirmation)

#### CR-0005 [minor] [new] inspect reads podman secrets from a top-level Secrets key

Location: containercli/inspect/inspect.go:115-117, containercli/inspect/inspect.go:161-163, containercli/inspect/inspect_test.go:57-71
Confidence: suspected
Evidence:

    containercli/inspect/inspect.go
    104 	Config       struct {
    111 	Mounts          []Mount
    115 	Secrets []struct {
    116 		Name string
    117 	}

    containercli/inspect/inspect_test.go
    57  // podmanInspect is podman's answer (solace-util's podmanInspect): no slash, the older
    70    "Secrets": [{"Name": "solace-admin-password"}, {"Name": "solace-tls-servercertificate"}]

Why it matters: podman reports a container's secrets inside `Config` (`.Config.Secrets`), not at the top of the element. If so, `Container.Secrets` is always empty on a real host while the hand-written fixture keeps the test green. The shape was lifted unchanged from A, so this would predate the module.
Downstream: A `internal/statusreport/parse.go:645` prints one `secret ... attached` row per entry, so the rows would never appear on podman. B has no non-test use.
To confirm: on a host with a container created with a secret, `podman container inspect <name> | grep -n -B3 -A6 '"Secrets"'` and read which object the key sits in.

Change, if confirmed: decode `Config.Secrets`, keep the top-level key as a fallback, and take the fixture from real output.

#### CR-0006 [minor] [new] quadlet.Environment leaves a dollar sign for systemd to expand

Location: containerart/quadlet/quadlet.go:22-62, containerart/quadlet/quadlet_test.go:43
Confidence: suspected
Evidence:

    containerart/quadlet/quadlet.go
    14  // EscapePercent doubles every '%'. systemd expands % specifiers in every
    15  // assignment of a unit, including what quadlet passes on to the service it
    16  // generates, [...]
    40  		case c == '%':
    41  			b.WriteString("%%")
    50  		default:
    51  			b.WriteByte(c)

    containerart/quadlet/quadlet_test.go
    43  		{`sh -c "test \"$X\""`, `sh -c "test \"$X\""`},

Why it matters: `%` is doubled because quadlet passes values through to the service it generates. If `$` passes through the same way, systemd substitutes `${NAME}` in the generated `ExecStart`, and the container receives a different value with no error. `compose.Escape` exists for the same hazard on the docker path; the podman path has no counterpart.
Downstream: A `internal/podmangen/podmangen.go:173-181` writes `JAVA_TOOL_OPTIONS` and `JDK_JAVA_OPTIONS` from operator Java options, plus the timezone and syslog host; B `internal/render/render.go:987` writes every `EnvPairs` value.
To confirm: write a unit with `Environment="A=x${HOME}y"`, run `/usr/libexec/podman/quadlet -dryrun -user` and read the `ExecStart` it prints; then start it and `podman exec <name> printenv A`.

Change, if confirmed: double `$` in `Escape` and add the row to `TestQuadletEscape`. The rendered bytes change only for values that hold a `$`.

## Dynamic checks

None run by the review. Read from `scripts/logs/`, user-run at 2026-11-11 16:17 on this commit with a clean tree:

- build: OK.
- vet: OK, no unformatted files listed.
- test: OK, 23 packages, `go test -count=1 ./...` without `-race`.
- cov: OK, total 99.7% of statements, measured on Linux.
- scan: OK, 0 reachable vulnerabilities; govulncheck notes 1 vulnerability in a required module that the code does not call.

## Fixed since last report

First run.

## Accepted (not re-litigated)

None.

## Coverage

- Areas reviewed: design, correctness, resilience, quality, tests, cicd, build, interfaces, observability, conventions; skipped: none.
- Read closely: all 29 non-test Go files; `scripts/dev.sh`, `dev.ps1`, `dev-all.sh`; `.github/workflows/ci.yml`, `tag.yml`, `.github/dependabot.yml`; `go.mod`, `.gitignore`, `.gitattributes`; `README.md`, `CLAUDE.md`, `docs/developer.md`, `docs/consuming.md`, `docs/smoke.md`. By a Sonnet subagent, with every cited line opened afterwards: the 26 test files against `docs/test.md` (all 318 test names agree in both directions), and both script pairs against the dev-script contract. Sampled: `paritydoc/body.md` (rules 8 and 9 and the tables), `scripts/dev-all.ps1` (the scraper and `pin`), the adoption table of `docs/developer.md`. Not read: `docs/handoff/`, `graphify-out/`, `go.sum`, `LICENSE`.
- Downstream: A and B were read only at the call sites named in the findings, from a subagent's usage map of every exported symbol. Seen while linking and not reviewed: the `go.work` in both tools has no `toolchain` line, which `docs/consuming.md:48-52` asks for.
- Stacks with core checklist only: PowerShell.
- Dynamic checks not run (offered, user-run): `DEV_RACE=1 ./scripts/dev.sh test` (CR-0016); `shellcheck scripts/dev.sh scripts/dev-all.sh` (CR-0021 and quoting); `actionlint` (both workflows); `go mod tidy -diff` (`go.mod` against imports); the podman inspect check (CR-0005); the quadlet dry-run check (CR-0006).
- Limits: CR-0003 rests on what kubectl's `all` category covers, and CR-0005 and CR-0006 on podman and systemd behaviour; none was exercised here. There are no rendered compose files or quadlet units in this tree, so the posture audit greps had nothing to read; the posture constants were read directly. Two subagent findings were dropped: `dev.ps1` reporting OK without a Go toolchain (the dispatcher's `try`/`catch` at `scripts/dev.ps1:178` turns that into a failure) and a timing concern in `TestDeadlineIsReportedAsDeadline` that could not be confirmed by reading.
