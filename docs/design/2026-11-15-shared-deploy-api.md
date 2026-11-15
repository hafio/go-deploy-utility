# Shared deployment API (design note)

Draft for the owner, 2026-11-15. Nothing is built. Section 7 records the owner's answers and
what is still open.

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
6. A preview works through the runner, as today: under `runner.Echo` a read answers
   nothing, and its parser reports "unknown".

### 3.4 Catalogue

What the two tools run today, plus what the decided items add:

| Client | Methods |
| --- | --- |
| `containercli.Client` (engine and compose tokens) | `Info` (error carries the platform hint), `List`, `Inspect` (decoded), `ComposeUp`, `ComposeDown` (with `-p`, gdu 9), `Start`, `Stop`, `Restart`, `Logs`, `Exec` (returns the child's exit code), `Upload` |
| `podman.Client` | `Version` (read once per run), `CreateSecret` (labels; the client picks replace or the fallback), `RemoveSecrets`, `SecretMatches` (the value never printed), `SecretLabels` and `SecretIDs` (gdu 10, 16), `IDMap` |
| `systemd.Client` (user mode and session env) | `DaemonReload`, `Start`, `Stop`, `Restart`, `State` (is-active), `LoadState` (gdu 9), `NRestarts`, `Linger`, `EnableLinger` |
| `kube.Client` | `CurrentContext` (gdu 11), `Apply`, `Delete` (timeout, gdu 12), `NamespaceState`, `DeleteNamespace` (timeout), `Scale`, `RolloutRestart`, `RolloutStatus`, `DeletePod`, `Get`, `ListLabelled` (gdu 18), `Logs`, `Exec`, `Upload` |
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
| gdu 11 | `kube.Client.CurrentContext` reads the context. Each tool resolves it once and puts `--context` into the client's `Base`. |
| gdu 12 | The delete executors take the timeout. |
| gdu 16, 18 | The shape decided on 2026-11-14 stands: builders, parsers and one pure compare function, which each tool runs. The listing and the deletes also get executors. |
| gdu 1 | The namespace create, the label on secrets and the labelled listing use executors. The lib annotation is question 5. |
| The surface review | Once the clients exist, the review also decides which builders stay exported: a tool's own steps still build commands, such as solmq-conn-util's image inspect. |

## 5. Rules and docs that change if this is accepted

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
and both tools adopt them in the same round (decided 2026-11-15).

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
7. Rendering moves into the module, all three layers of section 8 (2026-11-15).

Open:

6. **Builders.** Which builders stay exported once the clients exist: settled in the surface
   review.
7. **Rendering, timing.** In the breaking tag, or as its own step after it?
8. **The consumer rule.** "A package needs two consumers" blocks the workload kinds only
   solmq-conn-util renders today (Deployment, Service, ConfigMap, PVC). Relax it for kinds
   the reuse aim needs (section 9)?
9. **Licence.** GPLv3 binds every distributed tool that imports the module to a compatible
   licence. Keep it, or move to a licence a non-GPL tool can use? The owner's call, and
   possibly legal's.
10. **Module path.** Move from the personal `github.com/hafio/...` to an organisation path
    before more tools import it?
11. **v1.** When is the API declared stable for tools beyond the two?
12. **New-tool guide.** When is the guide of section 9.3 written, with or without an example
    tool?

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
| Rendering: Deployment, Service, ConfigMap, PV, PVC | decided (section 8), not built; open point 8 | all of it today |
| The order of operations: compare, ask, apply, wait, clean up | stays in each tool (section 1) | all of it, guided by 9.3 |

### 9.2 Gaps beyond code

- **No guide for a new tool.** The order of each operation, the rules a tool keeps (questions
  first, least privilege, what to probe, what a refusal means) and how to pin the module sit
  in the parity body, which names exactly two tools.
- **Built for two tools.** "A package needs two consumers", `paritydoc`, `dev-all` and the
  release gate ("both tools pass") assume a pair.
- **Licence.** GPLv3: a distributed tool that imports the module must carry a compatible
  licence.
- **Module path.** `github.com/hafio/go-deploy-utility` is a personal account. Moving it later
  changes every import in every consumer.
- **Stability.** The module is v0.x, with a breaking tag ahead.
- **Solace values.** The `services.solace.com/*` label values (solution, developed-by) suit
  Solace Professional Services accelerators, not tools outside them.

### 9.3 Steps, in order

1. The breaking tag as planned: executors, labels, network mode, name checks, atomic write,
   context pinning.
2. Rendering, section 8.
3. A guide for building a tool on the module: the operations and their order, the rules, how
   to pin, and the tests a tool runs (`artifacttest`, `noticetest`, `runnertest`). An example
   tool kept as its own module inside the repository would stay out of the published module.
4. The licence, the module path and v1 settled before other teams adopt it.

After step 2 a new tool writes its config, its application's content and its own sequence of
operations, the last guided by step 3. Open points 7 to 12 decide these steps.
