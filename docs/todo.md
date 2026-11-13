# To do

> [!WARNING]
> **Not a supported Solace product.** `go-deploy-utility` was created by Solace
> Professional Services and is supported only by Solace Professional Services --
> not by Solace Support. For help with this module, contact your Solace
> Professional Services representative rather than opening a Solace Support
> case. This notice covers this module only, not the Solace products that the
> tools built on it deploy.

Planned work that no adoption step in [developer.md](developer.md) covers yet. Each item
states the goal, what happens today, the scope and what is still undecided. Nothing here is
built.

## 1. Ownership labels on every artifact

Raised 2026-11-13 from the solmq-conn-util review of its namespace removal.

### Goal

Every object a tool creates carries labels that name the tool and the instance that created
it. A destructive step reads those labels back. It no longer decides ownership from names
alone.

### Today

- Kubernetes: `nsgate.Occupants` discounts an object when its kind and name match the
  caller's `owned` list. No label is read.
- `nsgate.List` asks for `all,persistentvolumeclaims,secrets,configmaps`. Roles,
  RoleBindings, NetworkPolicies, Ingresses, ServiceAccounts, PodDisruptionBudgets and custom
  resources are never listed, so the gate cannot see them.
- A namespace a tool created cannot be told from one it found. The namespace create records
  nothing.
- Compose: the deploy-digest label is the only label the module defines.
- Podman: `podman.CreateSecret` accepts labels and each tool passes its own. Quadlet
  `Label=` lines are written by each tool with its own keys.

### Scope

- One label vocabulary in the module: the key names, the value grammar and a helper that
  renders them for each sink.
- Kubernetes: the namespace is labelled when a tool creates it. Every manifest object and
  every Secret and ConfigMap a tool creates carries the labels.
- `nsgate`: the gate reads the namespace's label to learn whether the tool created it. It
  treats a labelled object as owned and anything else as an occupant.
- `nsgate.List`: widen the listing beyond the `all` category, for example from
  `kubectl api-resources --verbs=list --namespaced -o name`.
- Compose: services, networks, volumes, configs and secrets carry the labels.
- Podman: the quadlet unit, its volumes and every secret carry the labels.
- `containerart/artifacttest`: a check that fails when a rendered object lacks the labels.
- Both tools adopt the labels in one step, and the parity body gains a row for them.

### Undecided

- Key names: the `app.kubernetes.io/managed-by` family, or a prefix of our own.
- What a destructive step does with an unlabelled object: refuse, or ask.
- Whether `--no-prompt` may delete a namespace that carries no label.
- Deployments made before the labels exist: how a tool recognises and relabels them.
- Whether the wider listing needs extra permissions, and what the gate does when a kind
  cannot be listed.

### Until then

solmq-conn-util keeps its namespace removal as it is (decided 2026-11-13). Its listing covers
the kinds named above and nothing else.

## 2. One guideline for expansion in config values

Raised 2026-11-13 from the solmq-conn-util review of its `${VAR}` handling.

### Goal

One written rule states what a tool expands in a config value and what it passes on as text.
Both tools follow it, and the module holds the code they share for it.

### Today

- No rule is written down in this module, and it has no expansion helper.
- solace-util expands a leading `~` in every path key and in every `command` token but the
  first. A `$VAR` or `${VAR}` in a value is literal text.
- solmq-conn-util expands `${VAR}` and `${VAR:default}` in non-credential string fields. The
  expansion reaches validate and the rendered files but not the engine calls.
- Secrets never expand in either tool. A reference key names the variable, resolved by
  `envref`.
- `compose.Escape` doubles every `$`, so compose interpolates nothing a tool did not write.

### Decided

solmq-conn-util follows solace-util (decided 2026-11-13): it stops expanding `${VAR}`, expands
only a leading `~`, and passes a `${VAR}` in a value on as written.

### Scope

- Write the rule under "Rules for shared code" in [developer.md](developer.md) and add a row
  to the parity body.
- Lift the tilde expansion from solace-util (`expandTilde`, `expandTildeToken`) into the
  module, so both tools share one implementation and one set of tests.
- State, per sink, what happens to a `$` that is passed on: the compose file, the quadlet
  unit, the manifest and the application's own config file.

### Undecided

- Whether a `${` in a value earns a warning at load, since the tool no longer resolves it.
- Which keys count as path keys in each tool, and whether the list lives in the module.
- How solmq-conn-util announces the change to env files that rely on `${VAR}` today.

## 3. One step order for deploy and remove on every platform

Raised 2026-11-13 from the solmq-conn-util review of its platform pipelines.

### Goal

Deploy and remove run their steps in one stated order on kubernetes, docker and podman, in
both tools. The same mistake then reports the same error on every platform.

### Today

- solmq-conn-util sequences each platform by hand in its command package. Docker resolves
  credentials before the engine preflight, and podman resolves them after it.
- An unreachable engine together with an unset variable reports a different error on each
  platform.
- The parity body states the preflight, redeploy and prompt rules. It states no order for
  the steps around them.

### Decided

solmq-conn-util aligns its three pipelines to one order first (decided 2026-11-13): validate
and resolve, then preflight, then act.

### Scope

- Add the order to the parity body as a rule, and check both tools against it.
- Compare solace-util's deploy and remove sequence with that rule and list the differences.
- Decide whether the module gains a pipeline that runs the steps in order. It would take a
  `runner.Runner` and a per-platform set of step functions, and import no tool type.

### Undecided

- Where host checks and permission probes sit in the order.
- Whether remove resolves anything at all, since it needs names and no credentials.
- Whether a shared pipeline earns its place, or the rule and two aligned tools are enough.

## 4. Builders refuse a name outside its grammar

Raised 2026-11-13 from the solmq-conn-util review of the names its `status`, `logs` and `cli`
verbs put on an engine argv.

### Goal

A builder that places a pod, container, deployment or unit name on an argv refuses a name
outside that name's grammar. A tool that forgets its own check still cannot pass a name that
starts with `-` to an engine.

### Today

- The grammars exist: `kube.ValidLabel` and `containercli.ValidName`.
- Only the namespace helper in `kube` checks its argument. `kube.Logs`, `kube.Exec`,
  `kube.Scale`, `kube.RolloutRestart`, `kube.RolloutStatus`, `containercli.Logs` and
  `containercli.Exec` take the name as given.
- solmq-conn-util checked lifecycle names with a private copy of the DNS-1123 rule, and its
  `status`, `logs` and `cli` verbs allowed a leading `-`.

### Decided

solmq-conn-util calls the shared grammars at its own boundary and drops the private copy
(decided 2026-11-13).

### Scope

- Each builder that takes a name checks it and returns an error naming the field, the value
  and the rule.
- Builders that return only a `runner.Cmd` today gain an error result, so both tools adapt
  their call sites in the same step.
- A test per builder that a dash-led name is refused and no command is built.
- Quadlet lines: a builder writes each `Image=`, `ContainerName=`, `Label=` and `Volume=`
  line and escapes `%` itself. systemd expands `%` specifiers in a unit, so a host path that
  holds one mounts another directory today. `quadlet.EscapePercent` exists, solmq-conn-util
  never calls it, and `execguard.UnsafeChars` does not include `%`. solmq-conn-util changes
  nothing until this builder exists (decided 2026-11-13).

### Undecided

- Whether `%` joins the unsafe set, or stays legal and is always escaped.
- The grammar for a kubernetes pod name passed as `kind/name`, and for a podman unit name.
- Whether the signature change ships as one breaking v0 step or builder by builder.

## 5. One diff helper for golden tests

Raised 2026-11-13 from the solmq-conn-util review of its duplicated test helpers.

### Goal

Both tools report a golden mismatch through one helper, so every golden test fails with the
same output.

### Today

- solmq-conn-util defines `lineDiff` in three test files.
- solace-util defines `firstDiff` in two test files.
- This module has `firstDifference` in `paritydoc`, unexported.

### Decided

solmq-conn-util keeps one local copy until the shared helper exists (decided 2026-11-13).

### Scope

- A test-support package beside `artifacttest` and `noticetest` with one diff function.
- Both tools replace their copies in the same step.

### Undecided

- The output shape: the first differing line with context, or a full line diff.
- Whether `paritydoc` uses the same function for its own message.

## 6. Every artifact states its network mode

Raised 2026-11-13 from the solmq-conn-util review of its container posture.

### Goal

A rendered compose file, quadlet unit and manifest states its network mode. The posture check
fails an artifact that leaves the mode to the engine default.

### Today

- `artifacttest` forbids the host-network spellings unless a tool asks for them
  (`HostNetwork`). It does not require a stated mode.
- solmq-conn-util states no network mode and no root filesystem mode in any artifact.
- solace-util writes `network_mode: host` and `Network=host` when asked. It refuses a
  read-only root by name, because the broker needs a writable one.

### Decided

solmq-conn-util adds config keys for the network mode and the root filesystem, defaulting to
bridge and read-only, and writes both into every artifact (decided 2026-11-13).

### Scope

- `artifacttest` requires a stated network mode in compose and quadlet output, and the
  equivalent field in a manifest.
- The parity body gains a row for the network mode.
- A read-only root stays a choice of each tool. The posture check accepts either a stated
  read-only root or a tool's recorded reason for a writable one.

### Undecided

- The spelling that counts as stated bridge networking in a quadlet unit.
- Whether the read-only root check belongs in `artifacttest` at all, given the two tools
  differ.

## 7. One atomic write for files a tool installs

Raised 2026-11-13 from the solmq-conn-util review of its podman deploy.

### Goal

A unit file, and any other file a tool installs on the host, is written to a temporary file
and renamed over the target. A failed or interrupted write leaves the previous file in place.

### Today

- The module has no file-write helper.
- solmq-conn-util writes the quadlet unit in place, which truncates it first.
- Both tools write quadlet units.

### Decided

solmq-conn-util writes its unit through a temporary file and a rename, and removes stale
secrets only after the unit is written and reloaded (decided 2026-11-13).

### Scope

- One helper that takes a path, the content and a mode, and writes through a temporary file
  in the same directory.
- Both tools use it for quadlet units in the same step.

### Undecided

- The package that holds it: `systemd`, `podman` or a new one.
- Behaviour on Windows, where a rename over an open file fails.

## 8. One gate for the test catalogue

Raised 2026-11-13 from the solmq-conn-util review of its `docs/test.md` gate.

### Goal

Each repo's `docs/test.md` is checked against the test functions that exist, by name. A
renamed or removed test fails the gate.

### Today

- This module and both tools keep a `docs/test.md` that lists every test.
- Only solmq-conn-util gates it, and its gate compares three totals and no names.

### Decided

solmq-conn-util's gate compares names (decided 2026-11-13).

### Scope

- A test-support package that collects `func Test` names from a module and compares them
  with the first column of the catalogue.
- This module and both tools run it.

### Undecided

- The catalogue format the check reads, since the three files differ today.

## 9. Remove is idempotent on every platform

Raised 2026-11-13 from the solmq-conn-util review of its podman removal.

### Goal

A second `remove` of something already gone exits 0 and says nothing was installed, on
kubernetes, docker and podman, in both tools.

### Today

- Kubernetes deletes with `--ignore-not-found`, and `compose down` succeeds on an absent
  project.
- Podman removal runs `systemctl stop` first and returns on any error. A stop for a unit that
  is not loaded is expected to fail, which would make a repeat removal report failure. This
  has not been run on a host.
- The parity body has no rule for a repeat removal.

### Decided

solmq-conn-util skips the stop when the unit file is absent and the unit is not loaded, then
removes what is left and exits 0 (decided 2026-11-13).

### Scope

- Add the rule to the parity body and check both tools against it.
- Decide whether the podman removal sequence moves into the module, since both tools run it.
- Record the host check `systemctl --user stop <absent>.service; echo $?` in
  [smoke.md](smoke.md).

### Undecided

- What a repeat removal prints, and whether secrets left behind by a failed run are removed.

## 10. One deploy digest for podman units

Raised 2026-11-13 from the solace-util review of its podman redeploy (solace-util CR-0014).

### Goal

A podman redeploy learns from the running container whether it matches what would be
deployed, as a compose redeploy does. A change that was written but never applied is still
seen on the next deploy.

### Today

- Compose: `containercli.ComposeDigest` covers the file and every value the compose child
  receives. The tool compares it with the running container's label.
- Podman: each tool compares the unit on disk and each secret's stored value with the new
  ones (the "Redeploy over a running service" decision in `paritydoc/body.md`). The running
  container is never asked.
- In solace-util a deploy writes the secrets and the unit, and a later step can fail before
  the restart, such as the daemon reload. The next deploy finds the unit and the secrets
  equal to the new ones, reports the service as unchanged and exits 0. The old container
  keeps running on the old image or password.
- The module has no digest for a quadlet unit. `inspect.Container.Labels` already decodes
  the labels of a running container.
- solmq-conn-util makes the same two comparisons (`cmd/solmq-conn-util/main.go`). Whether a
  step that fails after the write leaves it in the same state was not followed.

### Decided

solace-util stamps the unit with a deploy-digest label and compares it with the running
container's label, as it does on docker (decided 2026-11-13).

### Scope

- A digest function beside `ComposeDigest` for a quadlet unit and the values it depends on,
  and one label key for both sinks.
- Rewrite the "Redeploy over a running service" decision in the parity body. Both tools
  adopt it in one step.
- Every quadlet golden changes, since the unit gains a `Label=` line.

### Undecided

- Whether the digest covers secret values, as the compose digest does, or only the unit,
  with the stored secret values still compared as today. A label is readable by anyone who
  can inspect the container.
- How the digest is computed over a unit that carries the digest line itself.
- Whether the label vocabulary of item 1 names this label.

## 11. Resolve and pin the kube context

Raised 2026-11-13 from the solace-util review of its destructive prompts (solace-util
CR-0026 and CR-0082).

### Goal

A run names the cluster it acts on once, and every kubectl call of that run reaches that
cluster. A context switched in another terminal between the question and the delete changes
nothing.

### Today

- `kube` has no notion of a context. Every builder takes the command tokens as given.
- solace-util reads `kubectl config current-context` once, prints it and repeats it in each
  destructive prompt. No call is pinned to it, so each delete runs against whatever context
  is current when it executes.
- When `kubernetes.command` carries its own `--context`, the prompt still names the
  kubeconfig's current context, which is not the cluster the calls reach.
- solmq-conn-util was not read for this item.

### Decided

solace-util pins the context, and the helper lives in this module (decided 2026-11-13).

### Scope

- A helper that finds the effective context: a `--context` in the command tokens wins,
  otherwise the answer of `config current-context`.
- A helper that returns the command tokens with `--context <resolved>` added when they
  carry none.
- The resolved name is kubectl output. It is checked before it is placed on an argv, and a
  name that starts with `-` is refused.
- A parity rule for the context, adopted by both tools in one step.

### Undecided

- Whether tokens that carry `--kubeconfig` are pinned too.
- The spellings to recognise: `--context X` and `--context=X`.
- The grammar of a context name, since cloud providers write `:` and `/` into it.
- What a run does when the kubeconfig has no current context: fail before the first call, or
  let kubectl report it.

## 12. A bound on every delete

Raised 2026-11-13 from the solace-util review of `broker remove` on Kubernetes (solace-util
CR-0041).

### Goal

A delete that cannot finish ends with an error that names what is stuck. It does not hang.

### Today

- `kube.Delete` builds `delete -f - --ignore-not-found`, and `kube.DeleteNamespace` builds
  `delete namespace <ns> --ignore-not-found`. Neither takes a timeout.
- kubectl waits for finalizers, and with no `--timeout` that wait is 168h.
- `kube.RolloutStatus` is the only builder with a bound (`RolloutTimeout`, 300s).
- solace-util's custom-resource delete and its PVC delete block for as long as a finalizer
  holds the object, for example a claim still mounted by a pod on a lost node.

### Decided

solace-util bounds its deletes and words the timeout (decided 2026-11-13). Its PVC delete,
which it builds itself, is bounded in the tool.

### Scope

- `Delete` and `DeleteNamespace` write `--timeout=<t>`, with a default constant as
  `RolloutStatus` has.
- Rule 1 of the parity body names the flag.
- Both tools adopt it in one step.

### Undecided

- The default bound.
- Whether the timeout arrives as a parameter or in an options struct.
- What the namespace gate does with a delete that timed out. Rule 9 keeps the namespace with
  a warning after a failed delete.

## 13. A watch loop stops on an error no retry can fix

Raised 2026-11-13 from the solace-util review of `--watch` (solace-util CR-0043).

### Goal

A refresh error that cannot succeed on a retry ends the loop with that error and its exit
code. Only a failure that may pass is retried.

### Today

- `watch.Loop` and `watch.LoopFrames` send every refresh error to `Warn` and try again
  (rule 13).
- solace-util's `broker status --watch` with a `--pod` value it refuses warns on every
  interval and exits 0 on Ctrl-C. The usage error never reaches exit 2.

### Decided

solace-util checks `--pod` before the loop starts (decided 2026-11-13). The loop itself
changes here.

### Scope

- The loop returns a refresh error that is a usage error, and retries the rest.
- Rule 13 of the parity body says which errors end the loop.
- A test that such an error ends the loop after one pass, and that others are retried.

### Undecided

- Whether the loop asks `exitcode.IsUsage` itself, or `Options` gains a function that says
  which errors are final.

## 14. One check for the log tail count

Raised 2026-11-13 from the solace-util review of `logs --tail` (solace-util CR-0058).

### Goal

`--tail` takes a count on every platform. A value that an engine would reject, or read
differently from another engine, is refused before any engine runs.

### Today

- `kube.LogOptions.Tail` and `containercli.LogOptions.Tail` are passed as given, and each
  comment leaves the check to the tool.
- The engines disagree on the grammar: docker accepts `all`, kubectl and podman take an
  integer.
- solace-util passed the value unchecked, and its help offered `all`.

### Decided

solace-util accepts a count only, and leaving the flag out means every line (decided
2026-11-13).

### Scope

- One function that checks a tail value, called by both tools at their boundary.
- Rule 6 of the parity body states the grammar.

### Undecided

- Whether the `Logs` builders refuse a bad value themselves. They return only a
  `runner.Cmd` today, which is the signature change item 4 already weighs.
- Whether `Since` gets the same shared check.

## 15. Dev scripts build the pinned module unless asked for the checkout

Raised 2026-11-13 from the solace-util review of its local gates (solace-util CR-0051).

### Goal

A plain local gate run tests what CI and a release build: the pinned version of this
module. Building against the checkout is asked for by name.

### Today

- A tool with a `go.work` builds against this checkout on every local run, and `GOWORK=off`
  selects the pin ([consuming.md](consuming.md)).
- CI and the tag pipeline check out one repo and have no `go.work`, so they build the pin. A
  green local run can cover code the pin does not contain.
- `dev-all` hands a repo only the words its dev script knows as a task or an aggregate, and
  refuses any other word before anything runs.
- The `dev-all` notice line decides between checkout and proxy from `GOWORK` and the
  presence of `go.work`.
- `dev-all pin` before a tag relies on each tool building through its `go.work`.

### Decided

solace-util's dev scripts gain a `--local` flag. Without it a run builds the pinned module.
With it the run builds against `../go-deploy-utility` and says so in one line. `go.work`
stays uncommitted (decided 2026-11-13).

### Scope

- `dev-all.sh` and `dev-all.ps1` accept `--local` and hand it to each tool's script.
- The notice line follows the flag as well as `GOWORK` and `go.work`.
- "Develop against a local checkout" and steps 2 to 4 and 7 of "Release and merge order" in
  [consuming.md](consuming.md) are rewritten for the flag.
- solmq-conn-util's dev scripts gain the same flag in the same step.
- The init-devops dev-script contract records the flag.

### Undecided

- How a default run selects the pin: `GOWORK=off` set by the script, and whether a `GOWORK`
  the caller already set wins.
- Whether `dev-all` passes `--local` on its own, since its purpose is to test the checkout.
- What a default run prints when the required version is not tagged yet and the proxy
  cannot find it.
- Whether `--local` uses the tool's own `go.work`, with the toolchain trap consuming.md
  describes, or writes a temporary one.

## 16. One behaviour for stale podman secrets

Raised 2026-11-13 from the solace-util review of its podman secrets (solace-util CR-0042).

### Goal

Rule 5 of the parity body is true for both tools, and both remove a stale secret the same
way.

### Today

- Rule 5 says stale secrets go with `podman secret rm --ignore`, in the section both tools
  are said to follow today.
- solmq-conn-util removes stale file secrets on deploy (`cmd/solmq-conn-util/main.go`). It
  lists nothing from the store. The candidates are a fixed list of the secret names an
  instance can own (`gen.PodmanFileSecretNames`) and the `Secret=` lines of the unit the
  last deploy wrote (`gen.PodmanUnitStoreSecrets`), read before the new unit replaces it.
  Whatever the new spec no longer mounts is removed. Remove deletes the credentials, the
  fixed list and the previous unit's secrets.
- solace-util creates and removes only the secrets its current env file names. When `tls.*`,
  the HA pre-shared key, a passphrase or an additional user is dropped from the env file,
  that secret stays in the store through the next deploy and through `broker remove`. The
  server-certificate secret holds a private key.
- `podman.RemoveSecrets` removes by name. The module cannot list the secrets a deployment
  owns.

### Decided

On podman the removal goes by ownership labels, as on Kubernetes in item 18. Every secret a
tool creates carries the labels of item 1. A deploy or a remove finds the instance's
labelled secrets and deletes the ones the current config no longer names. An unlabelled
secret is never deleted. Both tools are to be aligned on it, so solmq-conn-util moves from
its old-unit read to the labels. This replaces solace-util's two earlier decisions of the
same day, docs only and then the old-unit read (decided 2026-11-13).

### Scope

- Both tools pass the labels to `podman.CreateSecret`, which accepts them already. The
  labels name the tool and the instance and nothing about a value, since a label rides
  `secret create`'s argv.
- Builders to find the labelled secrets. `secret ls` filters by name and id only, so the
  tool lists the names, reads `Spec.Labels` with `secret inspect` for several names in one
  call, and filters in Go.
- The delta is the labelled names minus the names the current config renders. The tool
  names each secret it removes.
- A label is set only at create. A secret the config still names and that carries no label
  is created again once with the same value, so it gains the labels.
- Rule 5 of the parity body states the method. Both tools adopt it in one step, with item
  1's label vocabulary first or in the same step.
- `secret create --label` is in podman 4.4, the release quadlet needs, so no version floor
  moves. This is read from the podman documentation and has not been run on a host.
- Questions come first and actions after (decided 2026-11-13). Every question a deploy or a
  remove asks is answered before its first write. The write that adds the labels, the secret
  writes, the unit and the removals all follow the answer. Among the writes the removal runs
  last, after the unit is written and reloaded (item 7).
- solmq-conn-util drops its old-unit read (`previousStoreSecrets`,
  `gen.PodmanUnitStoreSecrets`) in the same step. The labels are the one source in both
  tools (decided 2026-11-13).
- A step that needs no question just runs. Questions are for destructive steps only, the
  ones already on record in the parity body and each tool's operations guide. A deploy whose
  only change is the write that adds the labels asks nothing. Removing a stale secret asks
  nothing either, and the tool names what it removed (decided 2026-11-13).
- Known limit, as in item 18: a secret dropped before the labels exist stays unlabelled and
  is never found, so it is removed by hand.

### Undecided

- Nothing in this item. The label keys are item 1's.

## 17. A readiness wait for OnDelete StatefulSets

Raised 2026-11-13 from the solace-util review of `broker start` and `broker restart`
(solace-util CR-0083). Suspected, not confirmed: it waits on a check against a cluster.

### Goal

The wait after a start or a restart reports readiness for every workload kind a tool
manages.

### Today

- Rule 16 waits with `kubectl rollout status <kind>/<name> --timeout=300s`, built by
  `kube.RolloutStatus`.
- solace-util waits on the broker StatefulSets, which the operator creates. kubectl refuses
  `rollout status` for a StatefulSet whose update strategy is `OnDelete`. If the operator
  writes that strategy, every start and restart ends in an error after the change is
  applied.
- The check that settles it has not been run:
  `kubectl get sts <name>-pubsubplus-p -n <ns> -o jsonpath='{.spec.updateStrategy.type}'`.

### Decided

Nothing until the check is run. An answer of `RollingUpdate` closes this item.

### Scope

- On `OnDelete`: a builder for `kubectl wait --for=condition=Ready pod/<pod> -n <ns>
  --timeout=<t>`.
- Rule 16 states the solace-util wait as a difference.

### Undecided

- Whether a builder with one consumer belongs in the module, or solace-util builds this
  call itself.

## 18. Kubernetes objects a config no longer names are removed

Raised 2026-11-13 from the solace-util decision in item 16, carried over to Kubernetes.

### Goal

A deploy or a remove deletes the objects an earlier deploy of the same instance created and
the current config no longer names. It never deletes an object the tool did not create.

### Today

- Both tools apply the objects the current config renders, and on remove delete that same
  set. Neither reads what an earlier deploy created.
- solace-util builds up to four Secrets, each only while the env file supplies its material:
  admin credentials, TLS, image pull and additional users (`ownedSecretNames` in
  `internal/k8s/prep.go`). When the material is dropped, the Secret stays through the next
  deploy and through `broker remove`. The TLS Secret holds a private key.
- solmq-conn-util renders one manifest (`internal/deploy/deploy.go`). Its image-pull Secret,
  its libs PV and PVC and its Service are rendered only while env.yaml asks for them, for a
  deploy and for a teardown alike. One that is dropped stays, and a Service that was switched
  off keeps routing.
- In both tools a config can name a Secret the operator made, and the workload references
  it exactly as it references one the tool built. solace-util's Secrets carry no label, so
  nothing on a Secret says who created it.
- The namespace gate names such a leftover as an occupant and keeps the namespace (rule 9).
  Nothing removes the leftover.
- Docker needs no such step. A compose secret is sourced from the environment and copied
  into the container, and nothing is stored outside it. A config change recreates the
  container, and the old container goes with what it held.

### Decided

On Kubernetes the removal goes by ownership labels, not by reading the last deploy. Every
object a tool creates carries the labels of item 1. A deploy or a remove lists the
instance's labelled objects and deletes the ones the current config no longer names. An
unlabelled object is never deleted. Docker needs no step. Both tools are to be aligned on it
(decided 2026-11-13).

### Scope

- The listing is by label selector in the instance's namespace, for the kinds the tool
  creates. The labels name the tool and the instance, so two instances in one namespace
  never see each other's objects.
- Item 1's label vocabulary lands first or in the same step. From then on every object a
  tool creates carries the labels.
- The delta is the listed names minus the names the current config renders. The tool names
  each object it removes.
- A parity rule for the removal, which also states that docker has nothing to remove. Both
  tools adopt it in one step.
- `delete` on each kind is probed before the first delete (rule 2). The listing is a read
  and is not probed.
- Questions come first and actions after (decided 2026-11-13). Every question a deploy or a
  remove asks is answered before its first apply or delete. Deleting a leftover asks nothing
  of its own, and the tool names what it deleted (decided 2026-11-13).
- solmq-conn-util's teardown deletes the listed leftovers as well as the current set.
- Known limit: an object of a deployment made before the labels exist gains them at the
  next deploy only while the config still names it. One dropped before that stays unlabelled
  and is never found, so it is removed by hand.

### Undecided

- Which kinds the removal covers beyond Secrets. solmq-conn-util's Service is a clear case.
  A PVC holds data and may follow the data rule, never the delta. Its PV is cluster-scoped,
  so listing it needs a wider read.
- What a deploy does when the listing is refused or fails: warn and delete nothing, or stop.
- Whether the tool lists and deletes by name, or hands the delta to kubectl's own
  `apply --prune`. Listing lets the tool name what it removes before it does.
- The wording of the line that names a deleted leftover. It is settled when the work is
  built.

## Inputs from the solace-util review to items 1 to 4 and 9

Decided in solace-util on 2026-11-13. These add no item. The IDs are from solace-util's
`docs/review/findings.json`.

- Item 1: solace-util parks its namespace-gate finding (CR-0006) on the wider `nsgate.List`
  and changes nothing in the tool until it exists.
- Item 2: solace-util's `convert` copies the values of a bash env file as written, with no
  expansion and no warning (CR-0015). That is one tool's answer to the undecided warning
  point. solace-util also folds `Load`'s `~` expansion and `convert`'s output check into one
  prepare step (CR-0036). That lands before the lift, so the lift moves one call site.
- Item 3: solace-util's container deploy changes the host before it asks the redeploy
  question: sysctl, linger, the data directory's owner and the registry login. That is
  against the "asks before writing anything" decision in the parity body. Its fix is
  read-only checks, then the question, then every host change (CR-0031). That order is its
  answer to where host checks sit. The owner states it as a rule for every command and both
  tools: questions are always first, actions after (2026-11-13).
- Item 4, names: mate pod names from `replication.sites[].via.kubernetes` reach `kube.Exec`
  checked only for being non-empty (CR-0021). solace-util parks its side with its
  replication work, and the builder refusal covers it.
- Item 4, quadlet lines: solace-util escapes `%` in the `Volume=` source and refuses `:` in
  it at load (CR-0030). Until the builder exists it calls `quadlet.EscapePercent`. Its
  answer to the undecided point is that `%` stays legal and is escaped. The builder should
  refuse `:` in a mount source.
- Item 4, operands: `kube.Get` takes its operands as given. solace-util passes a
  StorageClass name to a `get` it builds itself and adds its own check (CR-0003). A builder
  that refuses a dash-led operand would back that up.
- Item 9: on docker, `compose down` exits 0 for a project it does not find, for example
  when `COMPOSE_PROJECT_NAME` was set at deploy and not at remove. solace-util then reported
  the broker removed while it kept running, and emptied its data directory. Its fix asks the
  engine whether the container is still there after the down, and after the unit stop on
  podman (CR-0005). The same check may apply to solmq-conn-util's remove, which was not read
  here.
- Parity body, Host limits: solace-util adds a cgroup v2 check for rootless podman
  (CR-0089). The check stays in the tool, and the list under Inherent differences gains it.
