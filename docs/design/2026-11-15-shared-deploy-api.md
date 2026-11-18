# Shared deployment API (design note)

Draft for the owner, 2026-11-15. Nothing is built. Section 7 records the owner's answers and
what is still open. Section 10 holds the rulings that followed the surface review of the
same day; where it differs from sections 3 to 8, section 10 holds.

## 1. Direction

The owner's rule (2026-11-15): this module executes single deployment operations, one call
per operation. Chaining and orchestration stay in the tools. The artifacts each tool deploys
differ too much for a shared check-then-deploy, and the same holds for stop, remove and
restart. An earlier draft of this note proposed shared flows (plan, then apply); the owner
turned that down.

So the shared code is four kinds of thing:

- Builders: return a `runner.Cmd`. Pure, as today.
- Parsers: read an answer. Pure, as today.
- Pure functions: compares, digests, the namespace gate's judgement. As today.
- **Executors, new**: run one operation through a `runner.Runner` the tool hands in, and
  return its result or a typed error.

An executor never asks a question, never runs a second operation after the first, and never
decides whether a failure is a warning. The tool does all three.

## 2. What each tool repeats today

Each tool wraps every builder in its own code that runs it, captures what it needs, parses
it, classifies a refusal and words the error. The same operation is wrapped twice:

| Operation | solmq-conn-util (`internal/runner/runner.go`) | solace-util |
| --- | --- | --- |
| engine check with its hint | `Preflight` (376) | `Manager.Preflight` (`internal/container/preflight.go:28`) |
| permission probe | `Preflight`, `CheckProbe` (376, 407) | `Cluster.Preflight`, `preflightOne` (`internal/k8s/preflight.go:30, 56`) |
| apply or delete a manifest | `Kubernetes` (136) | `Cluster.apply` (`internal/k8s/cluster.go:189`) |
| namespace state | `KubernetesNamespaceState` (268) | `Cluster.readNamespace` (`internal/k8s/prep.go:51`) |
| rollout wait, refused told apart | `KubernetesRolloutStatus` (257) | `Cluster.waitRollout` (`internal/k8s/ops.go:157`) |
| scale, restart | `KubernetesLifecycle` (233) | `ReplicasStart`, `RestartPod` (`internal/k8s/ops.go:178, 135`) |
| list containers | `EngineList` (973) | `Manager.containerRunningKnown` (`internal/container/manager.go:984`) |
| inspect and decode | `EngineInspectJSON` (915) | `Manager.deployedDigest` (`manager.go:930`) |
| podman version | `PodmanVersion` (475) | `Manager.podmanReplace` (`manager.go:568`) |
| compare a stored secret | `PodmanSecretMatches` (545) | `Manager.storedMatches` (`manager.go:550`) |
| create or replace a secret | `PodmanSecretCreate` (496) | `Manager.writePodmanSecrets` (`manager.go:602`) |
| remove secrets | `PodmanSecretRemove` (523) | `Manager.removePodmanSecrets` (`manager.go:1080`) |
| unit state | inside `PodmanDeploy` (733) | `Manager.serviceActive`, `serviceState` (`manager.go:948, 967`) |
| restart count | `SystemctlNRestarts` (992) | `internal/container/inspect.go` |
| linger | `EnsureLinger`, `ProbeLinger` (679, 700) | `Manager.ensureLinger` (`internal/container/rootless.go:338`) |
| subid mapping | `PodmanIDMap` (561) | `Manager.checkIDMapping` (`rootless.go:193`) |

The chaining around these calls differs, and stays in each tool: what is compared before a
deploy, what is asked, in which order, what a refusal means, what is cleaned up.

## 3. Shape

### 3.1 One client per command family

Set-up is by constructor, not struct literal: see section 10, ruling 3.

Each package that builds a command family gains a small client type. The tool sets it up
once at the start of the run, with what every call of that family needs, and then calls one
method per operation (decided 2026-11-15):

```go
// once, at the start of the run
kc := kube.Client{
	Runner: r,
	Base:   []string{"kubectl", "--context", "prod"}, // guarded tokens, context pinned (gdu 11)
	Output: show,                                     // the tool's output hook (3.2)
}

// then, anywhere in the tool
res, err := kc.Apply(ctx, manifest)
state, res, err := kc.NamespaceState(ctx, ns)
res, err = kc.RolloutStatus(ctx, "deployment", name, ns, kube.RolloutTimeout)
```

The method sits beside the builder it runs (`kube.Apply` builds, `kube.Client.Apply`
executes), so a reader finds both in one place. A struct the tool fills in is an explicit
context, not global state, and tests fill it with `runnertest.Fake`.

### 3.2 Output and results

Replaced by section 10, ruling 1: two writers the tool supplies, and an answer's stdout is
returned, not written to them. The result type below stands.

Every executor hands each chunk of the command's output to the client's hook as it arrives,
unmodified and marked stdout or stderr, and returns a result beside its typed error
(decided 2026-11-15):

```go
type Result struct {
	Argv     []string      // the command as run; stdin as a byte count, env values masked
	ExitCode int           // -1 when it never started
	Duration time.Duration
}

// the tool's hook
func show(s runner.Stream, chunk []byte) {
	// solmq-conn-util: append to a buffer and print at the end
	// solace-util:     write to the terminal at once
}
```

The executor keeps its own copy of what it must parse, such as a namespace state, and the
tool still sees every byte. Each tool keeps its output model, so the parity body's open
Output row is untouched. An interactive session hands over the terminal instead, and its
result carries the child's exit code.

### 3.3 Rules for an executor

1. It runs one operation, and the module decides which command or commands that takes
   (decided 2026-11-15). The tool calls the same method on every engine version: for a
   podman secret, the client reads the version once, keeps it for the run, and runs
   `secret create --replace` on 4.7 and later, or `secret rm --ignore` and then
   `secret create` below. The method's doc names the fallback.
2. It returns data or a typed error the tool can branch on: refused (Forbidden, `can-i`
   no), not found, unanswered. It never turns a refusal into a warning; the tool chooses.
3. Its output reaches the tool through the hook (3.2).
4. It never asks, never sleeps, never retries, and never reads the environment or a file.
5. Its errors name the operation and the offending value, never a secret, and wrap the
   cause.
6. Dropped on 2026-11-15 (section 10, ruling 12). It read: a preview works through the
   runner, as today: under `runner.Echo` a read answers nothing, and its parser reports
   "unknown".

### 3.4 Catalogue

A first list. The surface review's operations map corrects and extends it (section 10,
rulings 2, 5, 6 and 13); the merged catalogue is the next design step.

What the two tools run today, plus what the decided items add:

| Client | Methods |
| --- | --- |
| `containercli.Client` (engine and compose tokens) | `Info` (error carries the platform hint), `List`, `Inspect` (decoded), `ComposeUp`, `ComposeDown` (with `-p`, gdu 9), `Start`, `Stop`, `Restart`, `Logs`, `Exec` (returns the child's exit code), `Upload` |
| `podman.Client` | `Version` (read once per run), `CreateSecret` (labels; the client picks replace or the fallback), `RemoveSecrets`, `SecretMatches` (the value never printed), `SecretLabels` and `SecretIDs` (gdu 10, 16), `IDMap` |
| `systemd.Client` (user mode and session env) | `DaemonReload`, `Start`, `Stop`, `Restart`, `State` (is-active), `LoadState` (gdu 9), `NRestarts`, `Linger`, `EnableLinger` |
| `kube.Client` | `CurrentContext` and `WhoAmI` (gdu 11), `Apply`, `Delete` (timeout, gdu 12), `NamespaceState`, `DeleteNamespace` (timeout), `Scale`, `RolloutRestart`, `RolloutStatus`, `DeletePod`, `Get`, `ListLabelled` (gdu 18), `Logs`, `Exec`, `Upload` |
| `kube/rbac` | `Run`: one probe through a `kube.Client`, returning `*RefusedError` |
| `kube/nsgate` | `Read`: the namespace listing through a `kube.Client`; `Occupants` stays pure |

### 3.5 What stays in each tool

- Every sequence: compare, then ask, then write; stop, then remove the unit, then the
  secrets; apply, then wait, then the leftovers.
- Every question, and the policy on each outcome: a refused wait warns, a refused write
  stops, a surviving secret fails a remove.
- Config, credential resolution, and the workload's own steps.
- What goes into its artifacts. The module renders them (section 8).

### 3.6 Example

The rollout wait in both tools today, after the change:

```go
_, err := kc.RolloutStatus(ctx, "deployment", name, ns, kube.RolloutTimeout)
switch {
case errors.Is(err, kube.ErrForbidden):
	warn("the wait for deployment/%s was refused; the change is applied", name)
case err != nil:
	return err
}
```

The executor runs and classifies. The tool decides that a refusal is a warning (parity rule
16) and words it.

## 4. Effect on what is decided

| Item | Change |
| --- | --- |
| gdu 3 | Stands: the module gets no pipeline, and each tool runs the order itself. |
| gdu 9 | `ComposeDown` with `-p` and `systemd.Client.LoadState` are executors. The sequence and the post-check stay in each tool. |
| gdu 10 | `podman.Client.SecretIDs` and the decoded `Config.Secrets` IDs are read by executors. The compare is a pure function. Each tool chains them. |
| gdu 11 | `kube.Client.CurrentContext` reads the context. Each tool resolves it once and puts `--context` into the client's `Base`. With no context anywhere, `kube.Client.WhoAmI` reports the identity the run acts as. |
| gdu 12 | The delete executors take the timeout. |
| gdu 16, 18 | The shape decided on 2026-11-14 stands: builders, parsers and one pure compare function, which each tool runs. The listing and the deletes also get executors. |
| gdu 1 | The namespace create, the label on secrets and the labelled listing use executors. The lib annotation is question 5. |
| The surface review | A builder a client covers is unexported (section 7, point 6); the review settles the exceptions. A command the module does not offer, such as solmq-conn-util's image inspect, the tool still builds itself. |

## 5. Rules and docs that change if this is accepted

- The consumer rule ("A package exists only when both tools use it" in `docs/developer.md`,
  "A package needs two consumers" in `CLAUDE.md`) becomes: common artifacts and operations are
  handled by the module; any customisation or orchestration stays in the tools.
- `docs/developer.md` "Rules for shared code" (lines 29-30) already says only code taking a
  `runner.Runner` executes. It gains the executor rules of 3.3.
- `CLAUDE.md` (lines 10-11) gains the executors beside the builders and parsers.
- The package docs that say nothing in the package executes (`kube/kube.go:13-15`, and
  `podman`, `systemd`, `containercli`) gain the client. "The flows stay in the tool" remains
  true.
- `docs/test.md`: the executors' tests.
- The parity body: a contract rule may name the module's executor instead of each tool's
  wrapper.

## 6. Release and tests

The clients are new types and methods, so adding them breaks nothing. The tools change their
call sites anyway in the breaking tag (gdu 4, 12, CR-0001), so the clients land in that tag
and both tools adopt them in the same round (decided 2026-11-15). Rendering lands in the same
tag: the labels already change every artifact, so each service restarts once, not twice.
The next tag ships everything in the plan; the module stays on v0.x.

Each executor is tested against `runnertest.Fake`: its argv, what it captures, what it writes
to the client's writers, and every typed error. The module's coverage floor of 99.7% holds.

## 7. Answers and open points

Answered 2026-11-15:

1. Form: a client per command family, set up once per run (3.1).
2. Output: the hook and the result (3.2).
3. Fallbacks: the module decides the commands for one operation (3.3, rule 1).
4. Release: the clients land in the breaking tag, in the same adoption round (section 6).
5. The lib annotation: the module writes it, with the tool's version annotation, in every
   object's metadata when it renders (section 8.4), so deploy and `generate` both carry it.
   A second `kubectl annotate` call was considered and dropped: it needs `patch` on every
   kind, and `generate` output would lack it.
6. A builder that a client method covers is no longer exported; tools run that operation
   through the client. A builder no client covers stays exported. Exceptions are decided
   name by name in the surface review (2026-11-15).
7. Rendering moves into the module, all three layers of section 8, in the breaking tag
   (2026-11-15).
8. The consumer rule is discarded. The rule is now: common artifacts and operations are
   handled by the module; any customisation or orchestration stays in the tools
   (2026-11-15).
10. The module path stays `github.com/hafio/go-deploy-utility` (2026-11-15).
9. The licence stays GPLv3 (2026-11-15): only GPL-compatible tools can adopt the module.
11. The module stays on v0.x; v1 is decided later. The next tag ships everything in the plan
    (2026-11-15).
12. The new-tool guide and an example tool ship in the breaking tag (2026-11-15). The example
    tool is its own module inside the repository, outside the published module.

Open: none. What remains is the surface review (plan review section 6), the plan for the
breaking tag, and the owner checks.

## 8. Rendering in the module

### 8.1 Today

Each tool renders its artifacts by hand, line by line, and the module holds pieces only:

| Artifact | solmq-conn-util | solace-util | Module |
| --- | --- | --- | --- |
| Kubernetes Secrets: Opaque, TLS, image pull | inline in `internal/deploy/deploy.go:374-418` | `secretManifest.render()` (`internal/k8s/secrets.go:35`), own registry payload (`secrets.go:249`) | nothing |
| Metadata block, `---` joining | inline | inline | nothing |
| Namespace | `kube.NamespaceManifest` | `kube.NamespaceManifest` | the one whole document |
| Deployment, Service, ConfigMap, PV, PVC | `internal/deploy` | - | nothing |
| Operator custom resource | - | `render.BrokerCR` (`internal/render/render.go:70`) | nothing |
| Compose file | `internal/dockergen` | `render.Compose` (`render.go:1043`) | quoting, hardening lines, `secrets:` block |
| Quadlet unit | `internal/podmangen` | `render.Quadlet` (`render.go:888`) | `Environment=` and `Secret=` lines, escaping, hardening lines |
| Writer | `internal/yamlwriter` | `fmt.Fprintf` | - |

The split comes from the original adoption rule, written in the `compose` and `quadlet`
package docs: each tool keeps its own renderer, its field order and its product settings.
Only the pieces both tools wrote identically were lifted.

### 8.2 Proposal

Decided 2026-11-15: the module renders the platform artifacts, and the tool describes its
application. A renderer is a pure function, a typed value in and bytes out, so it fits the module's rules
unchanged.

- The tool supplies what to deploy: image, names, user, environment, secrets and their
  targets, mounts, ports, the network word, resources, replicas, probes, restart policy, the
  instance and its own version.
- The module produces the artifact: the text, field order, quoting and escaping for each
  sink, the hardening lines, the network line (gdu 6), the labels and annotations (gdu 1,
  with `lib` filled in from the module's own version), secrets as files at absolute targets,
  and the joining of documents.

Layers, each usable on its own:

1. **Writer.** One YAML and line writer with the quoting rule `compose.Quote` already holds.
   No YAML library: the standard library only, and output that stays byte for byte stable
   for golden tests.
2. **Kubernetes documents.** `Meta` (name, namespace, labels, annotations), written by the
   module for every object. `Secret` (Opaque, TLS, image pull, with the registry payload),
   `Namespace`, `ConfigMap`, `Service`, `PersistentVolume`, `PersistentVolumeClaim`, and
   `Deployment` with its container spec (image, environment, secret mounts, the security
   context, probes, resources). `Custom` wraps a custom resource whose body only the tool
   knows, such as the broker CR, in the module's `Meta`.
3. **Compose file and quadlet unit.** One service and one unit model, with a slot for a
   workload's extra settings such as solace-util's ulimits.

### 8.3 What stays in the tool

- The content: every value in 8.2's first bullet.
- The body of a custom resource only its operator understands.
- The application's own config files (solmq-conn-util's application.yml, solace-util's
  broker configuration), which are not platform artifacts.

### 8.4 Effects

- The lib annotation (open point 5) is settled by `Meta`: deploy and `generate` both carry
  it, with no second call and no extra permission.
- The hardening lines, the network line and the labels are right by construction.
  `artifacttest` stays as the check on what a tool ships.
- Every golden in both tools changes once. The first deploy after adoption asks and
  recreates on docker and restarts a podman unit, which "migration is out of scope"
  accepts.
- The `compose` and `quadlet` package docs, and `docs/developer.md`, drop the rule that each
  tool keeps its own renderer.

## 9. Reuse by other tools

The owner's aim (2026-11-15): other tools adopt this module to deploy other applications on
docker, podman and Kubernetes.

### 9.1 Where a new tool stands

| Layer | State | What a new tool writes itself |
| --- | --- | --- |
| Safety: command guard, runner, env references, prompts, exit codes | ready | nothing |
| Host and engine checks: rootless podman, quadlet dir, systemd, engine info | ready | nothing |
| Builders and parsers for docker, compose, podman, systemd, kubectl | ready | nothing |
| Kubernetes permission probes, namespace gate | ready | the list of objects it owns |
| Executors (one operation each) | decided, in the breaking tag | nothing, once built |
| Labels, network mode, name checks, atomic write, context pinning | decided, in the breaking tag | nothing, once built |
| Rendering: compose, quadlet, Secrets, Namespace | decided (section 8), not built | all of it today: about 1,370 lines in solmq-conn-util and 1,300 in solace-util |
| Rendering: Deployment, Service, ConfigMap, PV, PVC | decided (section 8), not built | all of it today |
| The order of operations: compare, ask, apply, wait, clean up | stays in each tool (section 1) | all of it, guided by 9.3 |

### 9.2 Gaps beyond code

- **No guide for a new tool.** The order of each operation, the rules a tool keeps (questions
  first, least privilege, what to probe, what a refusal means) and how to pin the module sit
  in the parity body, which names exactly two tools.
- **Built for two tools.** `paritydoc`, `dev-all` and the release gate ("both tools pass")
  assume a pair. The consumer rule that did too is replaced (section 7, point 8).
- **Licence.** GPLv3, kept (section 7, point 9): a distributed tool that imports the module
  must carry a compatible licence.
- **Module path.** `github.com/hafio/go-deploy-utility`, a personal account, is kept (section 7,
  point 10).
- **Stability.** The module stays on v0.x until the owner decides on v1 (section 7, point 11).
- **Solace values.** The `services.solace.com/*` label values (solution, developed-by) suit
  Solace Professional Services accelerators, not tools outside them.

### 9.3 Steps, in order

1. The breaking tag as planned: executors, labels, network mode, name checks, atomic write,
   context pinning.
2. Rendering, section 8.
3. A guide for building a tool on the module: the operations and their order, the rules, how
   to pin, and the tests a tool runs (`artifacttest`, `noticetest`, `runnertest`). With it, an
   example tool kept as its own module inside the repository, so it stays out of the
   published module. Both ship in the breaking tag.
4. v1 decided before other teams adopt it.

After step 2 a new tool writes its config, its application's content and its own sequence of
operations, the last guided by step 3. Open points 7 to 12 decide these steps.

## 10. Rulings after the surface review

The owner's rulings of 2026-11-15 on
[../review/2026-11-15-surface-review.md](../review/2026-11-15-surface-review.md), section 2.
Where this section differs from sections 3 to 8, this section holds.

1. **Output.** A client holds two writers the tool supplies: its real stdout and stderr for
   direct, live output, a buffer, or nothing. The tool may give one call other writers
   (`kc.With(...)`). A method that returns an answer hands its stdout back as that answer,
   unmodified, and does not also write it to the tool's writer; its stderr goes there.
   A method that only acts sends both streams to the writers. No error text embeds
   captured stdout. This replaces the callback hook of 3.2 and the sentence that the tool
   sees every byte.
2. **Generic run.** Each client has `Run` (output to the writers) and `Output` (stdout
   returned) for an argument list the tool built. The command gets the client's base
   tokens, pinned context, session environment, writers, result and typed errors; the
   module decides nothing about the command itself.
3. **Set-up.** A constructor per client takes the runner, the tool's `execguard` rules, the
   command tokens, the approvals and the writers, runs the guard once and fails at set-up.
   The tokens cannot be changed afterwards. This replaces the struct literal of 3.1.
4. **Errors.** One set for every client, defined once in `runner`: refused, not found,
   unanswered, timed out, unsupported, no context, partial listing. The error carries the
   command's stderr as a field and keeps its cause; the result comes back on failure too.
5. **Exec.** Three methods: `Exec` (output to the writers), `ExecOutput` (stdout returned,
   optional stdin bytes) and `Attach` (hands over the terminal). The child's exit code is in
   the result in all three. A non-zero exit is an error for `Exec` and `ExecOutput`; an
   attached session that ran is a success whose code the tool returns. A pod target and a
   container target share one small interface.
6. **One-tool operations.** Those with an answer, a secret or a check move into the module:
   image inspect (decoded, with not found), stats and top from solmq-conn-util; registry
   login (password on stdin), a Secret value read, compose version, server version,
   `podman unshare`, copy in and out and `systemctl show` from solace-util. Image pull,
   describe, set env, `systemctl status` and `cat`, and every in-broker command stay with
   their tool over the generic run. Section 4's line that image inspect is tool-built is
   replaced.
7. **Host facts.** The module holds one implementation of each, written in Go: the subid
   files, the runtime-directory test, the account name. No `sh`, `test` or `id` child
   process: one correct way to read each.
8. **Release order.** gdu 2 and gdu 19 ship inside the breaking tag.
9. **One container description.** The compose, quadlet and Kubernetes renderers share one
   set of types (identity and labels, run-as, mount, port, environment entry, health check,
   timezone); each renderer checks its own sink's grammar. A forgotten field never means
   writable, root or host, apart from run-as as ruled below.
   - Extras slot: any key passes unless the module writes it itself or it is a
     privilege-widening key (a deny list). The deny list is complete for each sink, and
     `artifacttest` forbids the same keys.
   - Root filesystem: the field has no default and the tool must choose. Read-only writes
     the line; writable writes nothing, so gdu 6's decision holds.
   - Ports together with host or none networking: the renderer leaves the ports out and its
     doc says so.
   - Restart: an empty field writes `always`.
   - Run-as: numeric uid and gid. An empty field means the image's own user.
   - Quoting: every text value a tool supplies is written quoted. Booleans are bare, and so
     are the numbers a sink requires as numbers (Kubernetes refuses a quoted integer).
10. **The lib annotation.** The renderer writes it from a variable in the module that each
    tool's `build` task stamps at link time (`-ldflags -X`) with the version in its
    `go.mod`. Unstamped, as in every test run, it reads `devel`, so goldens stay stable.
    The init-devops skeleton gains the flag. This replaces "reads its own version" in 8.2
    and 8.4.
11. **The env file (gdu 19).** camelCase keys in both tools: solmq-conn-util renames its
    kebab-case keys. Both layouts flatten: the instance is `<platform>.name` on all three
    platforms and the namespace is `kubernetes.namespace`. The container name is a DNS
    label of at most 63 characters in both tools.
12. **No preview.** A mode that prints the commands a deploy would run is not a
    requirement. `runner.Echo` stays for tests and debugging, and rule 6 of 3.3 is dropped.
    `generate` stays as it is.
13. **Corrections** to recorded points: gdu 13 changes solmq-conn-util too; the compose file
    states its project network (gdu 6), against gdu 1's "no top-level network"; Kubernetes
    gets no `hostNetwork` line; names are DNS labels, not up to 100 characters;
    `podman.Client.Version` leaves the catalogue; an empty answer is not read as "unknown"
    by `ParseNamespace` or `systemd.Stopped`, which the catalogue step fixes per method.

Still to settle: the detail questions in
[../review/2026-11-15-surface-review-questions.md](../review/2026-11-15-surface-review-questions.md),
which the owner reviews, and the consolidation work of the review's section 6.

## 11. Answers to the detail questions

The owner's answers, asked in batches from 2026-11-17. The full text of each is under its
"Answer:" line in
[../review/2026-11-15-surface-review-questions.md](../review/2026-11-15-surface-review-questions.md).
Where this section differs from sections 3 to 8, this section holds; section 10 still holds
over both.

- **S1 Extras.** A blacklist that denies everything known to be wrong or harmful, as far as
  possible: the full lists of the entry plus the additions in its answer, among them
  Kubernetes `initContainers` and `containers`, probe and `lifecycle` keys, the service
  account keys, compose `provider` and `develop`, quadlet `LogOpt` and
  `HealthLogDestination`, and reserved label and annotation prefixes. Each list names the
  engine versions it was checked against.
- **S2 Own fields.** Host name, memory limit, CPU list, shared-memory size and typed ulimits
  (the 16 Linux names; `rtprio`, `rttime`, `nice` refused), and on compose A's entrypoint,
  command, configs, tmpfs and tool labels. One ulimits field writes compose `ulimits:`,
  quadlet `Ulimit=` and the `[Service]` `Limit*=` lines (Q1). Refused on Kubernetes. B
  passes no extras; `PodmanArgs` stays refused.
- **Q1 `[Service]` and `[Unit]`.** No extras there; the limit lines come from S2.
- **S3 Root.** The run-as field tells "stated" from "empty". A stated `0:0` writes the root
  lines without `runAsNonRoot`; empty writes nothing and means the image's user.
- **S9 Secret files.** No mode or owner on compose and quadlet. Kubernetes writes
  `defaultMode: 0400` and refuses a mode with a bit for "other".
- **S4 to S8, S12** as recommended: the `solution` and `developed-by` labels are optional
  fields; every data key is quoted; health times are whole seconds and retries at least 1;
  an image needs a tag or digest; a render error never echoes a free-text value; host
  names follow RFC 1123.
- **S10 Checks.** M exports a core set of checks, one per rule it renders by, and the
  renderers call the same ones. Tools call them at load and may add their own on top.
- **S11 generate output.** One write in M for every `generate` output: the mode follows the
  umask like a shell redirect, a symbolic link at the path is refused, and the file is
  written to a temp file and renamed. The overwrite question stays in the tool.
- **S13 Time zone.** An IANA time zone database name only; POSIX rule strings are refused.
- **E1 camelCase in A.** Only A's own keys are renamed. Keys passed on to the application,
  the connector's property names in `solace:` and `mq:` among them, stay as written.
- **E2 Old keys.** No detectors: an old key is an unknown key. Needs A's strict decode
  (CR-0046) in or before the breaking tag.
- **E3 Image keys.** `image.registry`, `image.repository`, `image.tag` in both tools.
- **E6 Restart.** Default `always` in both tools on every platform, A's docker default
  included; B gains the two keys; the word lists move into M.
- **E5 Run-as.** Absent is the tool's stated non-zero default; the word `default` gives the
  image's own user (M writes no user line); otherwise numeric `uid:gid` only. B reads a
  numeric image user for its chown.
- **E8 Kubernetes Secrets.** `kubernetes.secrets.credentials` and `.imagePull`, each
  `create: true` or `existing: NAME`, in both tools.
- **E7 Ports.** A bare port means `N:N` in both tools; M always writes both sides.
- **E9 to E11 Secrets.** Every Secret has `create` and `existing` keys, no hidden logic: B's
  TLS Secret is `kubernetes.secrets.tls`. `existing` with `semp.adminPass` is refused, and so
  is a password or cert file with no block. B's operator `regcred` stays as today for now.
- **Secrets reach.** A's TLS stores Secret and B's additional-users Secret also get
  `create` and `existing`; B's operator `regcred` does not.
- **C1 to C12 Compose**, all as recommended except C2 (open with E4): one key order; no
  labels on the project network; `read_only` written as stated, with a smoke row before A
  turns it on; no secret or config target under a mount; no duplicate or child-read
  variable names (`PATH`, `LD_*`, `DOCKER_*`, `COMPOSE_*` ...); line-break and control
  characters refused in inline text; one drive-letter root allowed in a bind source; `name:`
  always written; inline text byte for byte; one service per file; no named volumes or
  extra networks.
- **Q2 to Q11 Quadlet**, all as recommended: the on-disk compare stays beside the digest
  label; one line order; one write form per key (the bare keys are exceptions to ruling 9);
  the installed unit is 0600; a podman health command refuses shell characters; a mount
  source is an absolute path with no `:`, `$`, quote or backslash; relabel is none, `z` or
  `Z`, forgotten none; no password matching by name; a one-line header naming the tool;
  `WantedBy` follows rootful or rootless.
- **K1 to K18 Kubernetes**, as recommended except: K4 a CRD is never deleted (B drops
  `--delete-crd`; the CRD is applied with the operator); K10 `regcred` is unlabelled and
  handled by name, like the bundle; K18 nothing is built ahead of need, except that K6
  types every workload kind (Deployment, StatefulSet, ReplicaSet, DaemonSet, Job, CronJob,
  Pod) with the hardening lines built in and `Custom` refuses them; K8 B's additional users
  keep `extraEnvVarsSecret` as the one exception the operator's CRD forces. Also typed
  now: Ingress, NetworkPolicy, PodDisruptionBudget, HorizontalPodAutoscaler,
  ServiceAccount, Role, RoleBinding, ResourceQuota, LimitRange, PriorityClass,
  StorageClass; cluster roles refused. K16 (2026-11-18): a deploy refuses an object over
  the size limit and `generate` warns.
- **E19 to E24** (2026-11-18): B's `quadletDir` and both `baseDir` keys dropped; a
  read-only root key in both tools (A true, B false, M requires a value); Kubernetes
  identity in the pod and container fields of B's convention, in both tools; every image
  key a full registry, repository and tag block; free-form command keys replaced by
  structured keys on all three platforms (design next; reshapes E16, E17, gdu 11).
- **E16 to E18** (2026-11-18): E16 and E17 are superseded by E24's structured keys
  (`docker.compose: plugin | standalone`; `cli` words per platform). E18: deploy pulls the
  image first and M names the cause of a failed pull (login, wrong image, unreachable); on
  a login failure the tool tells the operator to log in and never logs in itself. This
  takes registry login out of ruling 6's list.
- **N1 to N75** (2026-11-18): every recommendation taken; N9 and N10 reshaped by E24 (no
  `--allow-command`; M exports only the no-prompt flag pair).
- **Flatten review** (2026-11-18), on
  [../review/2026-11-17-env-layout.md](../review/2026-11-17-env-layout.md):
  - `image.name` replaces `image.repository` (E3 revised); `image.registry` stays.
  - `<platform>.name` stays one container name; HA brokers run on separate hosts. A new
    `redundancy.sameHost: true | false` (B) makes the names `<name>-primary`, `-backup` and
    `-monitor` for three brokers on one host.
  - Identity: the word `default` in both tools tells M to write none; unset means the tool's
    default (A `1000:0`).
  - A remote engine is detectable on docker (the selected context's endpoint is not
    `unix://` or `npipe://`); a bind-mount source is then refused, and the docs say so.
    Podman is local only (E24); Kubernetes writes no host path.
  - `network.mode` and `network.ports` stay as B's block in both tools (E4, E7).
  - `kubernetes.service.ports` (E15). Memory under `resources` on every platform, CPU on
    Kubernetes in A only (E12).
  - B's per-role placement and A's single Deployment are an inherent difference.
  - The time zone key is the top-level `timezone` in both tools (written as `TZ`).
  - Strict decoding is a tool convention, not module code: each tool accepts only the keys
    of its own env.yaml shape, except blocks passed on verbatim (A's connector properties,
    B's custom-resource pass-through). The guide states it for a third tool.
  - C2: absent `network.mode` is the deploy's own network (docker project network, podman
    `private`); a written `bridge` is docker's shared bridge. E14: one `spread` key, A
    default `none`, B `preferred` with a new `none`. B on Kubernetes states 1000001 and
    1000002 when the identity is unset and refuses a stated 0.
  - E13: each tool keeps its placement shapes (A Kubernetes' own, B its strings), since they
    deploy different kinds.
  - E25: the list of inherent and one-sided differences is accepted for the parity body.
- **Service types** (2026-11-18), revising K18 and the stated ClusterIP default: a new
  `kubernetes.service.type` in both tools. M owns the word list (ClusterIP, NodePort,
  LoadBalancer, exact case; ExternalName refused), checks it, and its Service object requires
  a word: M writes no default. Each tool sets its own. B defaults to LoadBalancer, written
  into the custom resource as today, accepts all three, and refuses `service.loadBalancer`
  unless the type is LoadBalancer. A: default to confirm (ClusterIP recommended). A fixed
  node port is not offered: the cluster picks one.
- **Still open** (2026-11-18): one rule for the `create` defaults of every Secret
  ([../review/2026-11-18-secrets-assessment.md](../review/2026-11-18-secrets-assessment.md));
  A's default Service type (ClusterIP recommended).
- **Consolidation drafts** (2026-11-17, Sonnet agents, open points marked):
  [2026-11-17-container-description.md](2026-11-17-container-description.md) and
  [2026-11-17-catalogue.md](2026-11-17-catalogue.md).
- **First build step**: [2026-11-17-plan-r1-workload-checks.md](2026-11-17-plan-r1-workload-checks.md),
  approved 2026-11-17.
- **E4 and the flatten of ruling 11** are reopened: the owner reviews the flattened layout,
  [../review/2026-11-17-env-layout.md](../review/2026-11-17-env-layout.md), before key
  placement is decided.
