# Shared deployment API (design note)

Draft for the owner, 2026-11-15. Nothing here is built or decided. Section 8 holds the
questions that decide it.

## 1. Why

The owner's intent (2026-11-15): this module is a common deployment interface. It provides
deploy, update, create, start, stop, restart, remove and the like, and both tools call them.
A flow both tools need is written once, not once per tool.

What the module offers today is one layer below that. Its builders return a `runner.Cmd`
and its parsers read an answer. Each tool strings them into its own flows: the order, the
questions, the permission probes, the compares, the waits and the cleanup. The parity body
already makes the two tools behave the same in 16 contract rules, so the same behaviour is
written twice and kept equal by review.

Where the flows live today:

| solmq-conn-util | solace-util |
| --- | --- |
| `cmd/solmq-conn-util/main.go` (`actKubernetes`, `actDocker`, `actPodman`, `podmanDeploy`, `podmanRemove`) | `internal/container/manager.go` (1578 lines: `PrepHost`, `Deploy`, `Delete`, lifecycle, status) |
| `cmd/solmq-conn-util/lifecycle.go`, `namespace.go`, `rootless.go` | `internal/k8s/prep.go`, `deploy.go`, `ops.go`, `namespace.go`, `preflight.go` |
| `internal/runner/runner.go` (1214 lines: the tool's wrapper and its flow helpers) | `internal/container/rootless.go`, `preflight.go`, `internal/cli/ops_*.go` |

The inventories behind this note were read from both trees on 2026-11-15. Section 2 condenses
them.

## 2. What the two tools do today

"Both" means both tools run the step, with the module's builders, in the same place.

### 2.1 Docker (compose)

| Operation | Both | solmq-conn-util only | solace-util only |
| --- | --- | --- | --- |
| deploy | `info`; digest of the compose file and child env (`containercli.ComposeDigest`); find the running container (`containercli.List`, matched in Go); read the digest label (`inspect`); ask when running and changed; `compose -f - --project-directory <dir> up -d` | image entrypoint check, pulling once, before any write (created stores) | host prep before the deploy: limits, `mkdir` and `chown` of the data dir, DNS, registry login; `info` twice; `--restart` adds `--force-recreate` on an unchanged container |
| remove | remove question; `info`; `compose down` | - | `--delete-data` layer question and data purge |
| start, stop, restart | `info`; `<engine> start\|stop\|restart <name>` | - | restart asks first |
| status | `ps` listed and filtered in Go; `container inspect` decoded | status script in the container, stats, image digests | `--detail` streams the inspect |
| logs, exec | one builder each (`containercli.Logs`, `containercli.Exec`) | - | - |

### 2.2 Podman (quadlet and systemd)

| Operation | Both | solmq-conn-util only | solace-util only |
| --- | --- | --- | --- |
| deploy | scope (`rootless.Decide`, `podman.QuadletDir`) and user-session repair; `info`; linger ensured, a refusal stops; subid check; `podman version` for replace; unit compared with the file on disk; each secret compared (`secret inspect --showsecret`); `is-active`; ask when changed and active; `secret create [--replace] <n> -`; stale secrets `rm --ignore`; write the unit; `daemon-reload`; start, or restart when changed | stale store secrets read from the installed unit; mount readability warnings; unit mode 0644 | limits, user systemd check, data dir probes and `unshare chown`, DNS, registry login; legacy certificate bundle removal; `--restart`; unit mode 0600 |
| remove | remove question; `systemctl stop`; unit file removed; `daemon-reload`; `secret rm --ignore` | secrets removed even after a failed teardown | a failed stop is checked with `is-active` before it stops the teardown; one `rm` per secret, a surviving certificate secret fails; data purge |
| start, stop, restart | `info`; `systemctl [--user] start\|stop\|restart <svc>` | - | session ensured first; restart asks |
| status | as docker, restart count from `systemctl show -p NRestarts` | as docker | `systemctl status` and, with `--detail`, `systemctl cat` |

### 2.3 Kubernetes

| Operation | Both | solmq-conn-util only | solace-util only |
| --- | --- | --- | --- |
| deploy | `auth can-i` for every kind written, before the first write; namespace read, created only when absent (parity rule 15); `apply -f -` with the manifest on stdin | operator-managed Secret keys probed; `rollout status --timeout=300s`, a refused wait warns | admin password preflight; certificate digest compare and restart offer; Secrets applied separately; the custom resource read back instead of a rollout wait |
| remove | remove question; `can-i delete` per kind; `delete -f - --ignore-not-found`; namespace gate (`nsgate`, parity rule 9) | - | `--delete-data` claims; owned Secrets deleted by name; each permission probed twice |
| start, stop | `scale` to the configured replicas or 0; wait on start | probes `patch deployments` | no probe; three StatefulSets in role order |
| restart | wait after the restart | `rollout restart deployment/<name>` | pods deleted one at a time in role order (OnDelete StatefulSets, gdu 17) |
| status, logs, exec | reads only, never probed (parity rule 2) | pods, the status script | the custom resource's conditions |

### 2.4 Seams already in place

- Both tools take the module's `runner.Runner` and one `confirm.Prompter`.
- solmq-conn-util's `runner.PodmanDeploy` already takes `approve` and `write` callbacks
  (`internal/runner/runner.go:733`): the closest thing to a shared flow today.
- solace-util's `Manager` and `Cluster` take `Confirm`, `Log` and `Out` as fields
  (`internal/container/manager.go:44-84`, `internal/k8s/cluster.go:40-70`).
- Output differs: solmq-conn-util buffers each step and prints `ok:` or `error:` at the end;
  solace-util streams. This is the parity body's one "open" row.

## 3. Proposed shape

### 3.1 Two layers

- **Builders and parsers** stay as they are: pure, returning a `runner.Cmd` or data.
- **Flows** are new. A flow runs a sequence of builder calls through seams it is handed, in
  a `Session`. It never reads the environment or the filesystem itself, keeps no globals,
  registers no flags and never calls `os.Exit`. It takes no tool type: the tool passes
  values (names, command tokens, rendered artifact bytes, resolved secret values).

The flows sit in new packages, apart from the builders, so the rule stays readable: a
builder package never executes, a flow package executes only through its `Session`.
Proposed names: `lifecycle` for the shared types, and `lifecycle/composeops`,
`lifecycle/quadletops` and `lifecycle/kubeops` for the platforms.

### 3.2 The session

```go
package lifecycle

// Session is everything a flow may touch. The tool builds one per run.
type Session struct {
	Runner  runner.Runner // every command
	FS      FS            // the only file access: read, atomic write (gdu 7), remove
	Report  Reporter      // progress and warnings; child output goes to Report.Output
	Preview bool          // print what would run and write, change nothing
}

type Reporter interface {
	Step(format string, args ...any)
	Info(format string, args ...any)
	Warn(format string, args ...any)
	Output() (stdout, stderr io.Writer)
}

// Question is asked by the tool, never by a flow.
type Question struct {
	Text   string
	Strict bool // the answer must be an exact "yes"
}
```

A `Reporter` lets each tool keep its output model. solmq-conn-util hands in one that buffers
and prints at the end. solace-util hands in its `output.Sink`, which streams.

### 3.3 Plan, then apply

A deploy is two calls:

1. `Plan` runs the read-only part: the engine preflight, the permission probes, the reads
   and the compares. It returns what would change and the questions that change needs.
2. The tool asks those questions together with its own, then runs its own host changes.
3. `Apply` takes the plan and writes.

This keeps the step order of gdu 3 (validate and resolve, read-only preflight, compare, ask,
act) and the rule that every question is answered before the first write, while a tool's own
steps still fit between the questions and the writes. A flow never asks: it returns
questions, and the tool asks them through the shared `confirm.Prompter`.

A remove asks first and contacts nothing before the answer (gdu 3), so it is one call after
the tool's question: `Remove` runs the preflight for what was answered, then acts. The
namespace offer, the named exception, follows as `PlanNamespace` and `DeleteNamespace`.

Start, stop and restart are one call each: preflight, act, and the wait where there is one.

### 3.4 Catalogue

| Package | Functions |
| --- | --- |
| `lifecycle` | `Session`, `Reporter`, `FS` (with the atomic-write default of gdu 7), `Question`, `AskAll(prompter, noPrompt, questions)` |
| `composeops` | `Plan`, `Apply`, `Remove` (with `-p`, and the post-check that the container is gone, gdu 9), `Start`, `Stop`, `Restart`, `Status` (decoded containers), `Logs`, `Exec` |
| `quadletops` | `Scope` (decide, session repair), `PrepareHost` (linger, subid check), `Plan` (unit digest and secret IDs, gdu 10; stale secrets by label, gdu 16), `Apply`, `Remove` (stop, load-state check, unit, reload, then secrets, gdu 9 and 16), `Start`, `Stop`, `Restart`, `Status`, `Logs`, `Exec` |
| `kubeops` | `ResolveContext` (once per run, pinned, gdu 11), `Plan` (probes per kind, namespace state, labelled leftovers, gdu 18), `Apply` (namespace create with labels, apply, wait, then leftovers), `Remove`, `PlanNamespace`, `DeleteNamespace` (gdu 1), `Scale`, `RolloutRestart`, `Wait`, `Logs`, `Exec` |

Status returns data, not a report. Each tool's report is about its own workload (the
connector's status script, the broker's conditions) and stays in the tool.

### 3.5 An example

```go
plan, err := kubeops.Plan(ctx, s, kubeops.Deploy{
	Base:      base,     // guarded command tokens, the context pinned
	Namespace: ns,
	Manifest:  manifest, // rendered by the tool
	Writes:    written,  // kinds and names, for the probes
	Labels:    labels.For(tool, instance, toolVersion),
	Rollout:   []kubeops.Target{{Kind: "deployment", Name: name}},
})
if err != nil {
	return err
}
// The tool asks plan.Questions with its own questions, then runs its own steps.
return kubeops.Apply(ctx, s, plan)
```

solace-util passes no rollout target on deploy and reads its custom resource back after
`Apply`, as today.

### 3.6 What stays in each tool

- Config loading and validation, the CLI surface, flags and wording.
- Rendering: the compose file, the quadlet unit, the manifests and the custom resource.
- Resolving credentials: the flows receive values, never variable names.
- Workload steps:
  - solmq-conn-util: store building, the image entrypoint check, library download, the
    status script, the operator Secret key probe.
  - solace-util: host limits, data dir, DNS, registry login, HA role detection, the
    operator, the admin password checks, the certificate restart, broker configuration,
    and the ordered pod restart of gdu 17.
- The report each status command prints.

### 3.7 Labels and the lib annotation

The flows apply the labels where they write: on each podman secret, on a namespace they
create, and in the leftover listing. Manifests and units are rendered by the tool, so it
writes the labels there from the module's label helper (gdu 1). That helper adds
`deploy.solace.com/lib` itself, reading the module's version with
`runtime/debug.ReadBuildInfo`; a workspace build writes `go-deploy-devel`. No tool passes the
lib value. Adding it inside `Apply` instead would mean parsing YAML in the module, and the
`generate` output would lack it.

## 4. Differences a shared flow has to settle

Each line gives today's state and a recommendation.

1. **Which lifecycle verbs ask.** solace-util asks before restart on every platform and
   before nothing else; solmq-conn-util asks before none. Recommended: the flows ask nothing,
   and each tool keeps its own question before the call. Only a deploy and a remove carry
   questions in the flow.
2. **Probes on start and stop.** solmq-conn-util probes `patch deployments`; solace-util
   probes nothing, against parity rule 2. Recommended: `kubeops.Scale` probes, in both.
3. **A failed podman stop on remove.** solace-util checks the state before giving up;
   solmq-conn-util stops. Recommended: solace-util's check, which is the load-state step of
   gdu 9.
4. **The user session on podman lifecycle verbs.** solace-util repairs it first;
   solmq-conn-util does not. Recommended: always, since `systemctl --user` needs it. Linger
   is ensured on deploy only.
5. **Probing twice.** solace-util probes the same permissions in its preflight and again in
   the step. Recommended: once, in `Plan` or at the start of `Remove`.
6. **Recreate an unchanged service.** solace-util's `--restart`; solmq-conn-util has none.
   Recommended: a `RestartUnchanged` field in the deploy request, unused by solmq-conn-util.
7. **Legacy cleanup.** solace-util removes an old certificate bundle; solmq-conn-util reads
   stale store secrets from the installed unit. Both are migration code. Recommended: drop
   both (migration is out of scope; the labels of gdu 16 replace the second).
8. **Unit file mode.** 0644 against 0600. Recommended: 0600, since a unit can carry
   environment values and only its owner and systemd need to read it.
9. **A wait after deploy on docker and podman.** Neither tool waits. Recommended: no change
   in this step.

## 5. Effect on what is decided

| Item | Change |
| --- | --- |
| gdu 1 | Unchanged decisions. The namespace create, the label discount in the gate and the label on secrets move into `kubeops` and `quadletops`. The lib annotation is the label helper's (3.7). |
| gdu 3 | The order is built into `Plan` and `Apply`, not repeated in each tool. The parity rule stays. "The module gets no pipeline" is replaced. |
| gdu 7 | The atomic write is the default `FS`. |
| gdu 9 | The remove sequence (`-p`, load state, post-check, secrets after the unit) is built once, in `composeops` and `quadletops`. |
| gdu 10 | The unit digest and secret ID compare are `quadletops.Plan`. |
| gdu 11 | `kubeops.ResolveContext`, once per run. |
| gdu 12 | The delete timeout sits in `kubeops`. |
| gdu 16, 18 | The listing, the compare, the delete and the failure policy are built in the flows, not handed to each tool as builders. |
| gdu 17 | Stays in solace-util. |
| gdu 2, 4, 5, 6, 8, 13, 14, 15, 19 | Unaffected. |
| CR-0001 and the surface review | The review also asks which builders a tool still calls once the flows exist; those stay exported for the tools' own steps, the rest may go internal. |

## 6. Rules and docs that change if this is accepted

- `docs/developer.md` "Rules for shared code" (lines 29-30) and `CLAUDE.md` (lines 10-11):
  builders and parsers stay pure; flows execute only through a `Session`.
- The package docs that say the tool runs the command and owns the flow: `kube/kube.go:13-15`,
  and the same sentence in `podman`, `systemd` and `containercli`.
- `docs/todo.md`: the intro rule and item 3.
- The parity body: each contract rule names the module's flow instead of two tool files.
- `docs/test.md`: the flows' tests, and the tool tests they lift, under their own names.

## 7. Release and adoption

The owner's release plan stands: one breaking tag, both tools adopting it in one round, their
adoption work on a branch until the tag (decisions, "Releases").

Recommended order inside that round, each step planned, approved and gated on its own:

1. `lifecycle` and `kubeops`. Kubernetes has the most shared steps (probes, namespace, apply,
   wait, gate) and the most decided items waiting on it (gdu 1, 11, 12, 18).
2. `quadletops` (gdu 9, 10, 16).
3. `composeops` (gdu 9).
4. Each tool moves platform by platform on its branch, through its `go.work`, and keeps its
   gates green after every move.
5. The tag, once both tools pass their complete gate against the checkout.

The flows are new packages, so they break nothing on their own; the tag is breaking because
of the builder changes already planned (gdu 4, 12, CR-0001).

Tests: every flow runs against `runnertest.Fake`, a fake `FS` and a recording `Reporter`,
and pins its order of commands, its questions and every refusal path. Lifted tool tests keep
their names. The module's coverage floor of 99.7% holds.

## 8. Questions for the owner

1. **Shape.** Plan, then apply, with the tool asking between them (3.3)? Or one call per
   operation with hooks the tool fills in?
2. **Packages.** New `lifecycle/*` packages beside the builders (3.1)? Or the flows inside
   the existing `containercli`, `podman` and `kube` packages?
3. **Output.** A `Reporter` so each tool keeps its own model (3.2)? Or align the two models
   now and close the open parity row?
4. **Status.** The flows return data and each tool keeps its report (3.4)? Or a shared
   status report too?
5. **The nine differences in section 4.** Take the recommendations as a set, or one by one?
6. **Preview.** solace-util previews through `runner.Echo`; solmq-conn-util has no preview.
   Should every flow honour `Session.Preview`?
7. **Order of work.** The order in section 7, inside the one breaking tag?
