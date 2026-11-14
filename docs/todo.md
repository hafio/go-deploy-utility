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

The owner decided the open points on 2026-11-14. Each decision is written into its item
below, and [review/2026-11-13-decisions.md](review/2026-11-13-decisions.md) holds them all
beside the decisions on this module's own review. Three of them apply to every item:

- Migration is out of scope. A deployment made before a change is not carried over, and a
  restart on the first deploy after a change is accepted.
- The shared code is builders, parsers and pure functions. Each tool runs them in its own
  order. The module gains no code that runs a command.
- Order of work: a patch tag first, which holds no item of this file. Then one breaking tag
  for items 1, 4 to 14, 16 and 18. Items 2, 3, 15 and 19 follow as rules both tools align
  on. Item 17 moved to solace-util.

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
- Compose: the module defines no label key. Each tool has its own deploy-digest key
  (`solace-connector/deploy-sha256`, `solace-util.deploy-sha256`), and solmq-conn-util also
  writes `solace-connector/le-mode`, `/role` and `/stores-sha256`.
- Podman: `podman.CreateSecret` accepts labels and neither tool passes any. solmq-conn-util
  writes two quadlet `Label=` lines (`le-mode`, `role`); solace-util writes none.
- The Solace operator labels the objects it generates with `app.kubernetes.io/instance=<CR
  name>`, `/name` and `/managed-by`. It copies no label from the custom resource to them.
  They are tied to the custom resource by ownerReferences, except the PVCs.

### Decided

Decided 2026-11-14.

- Labels, on every sink: `app.kubernetes.io/managed-by=<tool>`,
  `services.solace.com/product=<tool>`, `services.solace.com/solution=accelerator`,
  `services.solace.com/developed-by=apac-psg` and `deploy.solace.com/instance=<instance>`.
- Annotations, on Kubernetes only and for information only:
  `deploy.solace.com/version=<the tool's version>` and
  `deploy.solace.com/lib=go-deploy-v<this module's version>`. Neither is written to a unit,
  a compose file or a podman secret, so a tool upgrade alone restarts nothing.
- A destructive step selects on `services.solace.com/product` and
  `deploy.solace.com/instance` only. `managed-by` is a key other software writes too, so it
  is written and never read.
- The gate keeps its by-name roots and its ownerReference chain. A label is one more
  discount, and only for this tool and this instance. Operator-generated objects carry no
  tool label and stay discounted through their owner.
- The listing is widened (`nsgate.List`), and that is what protects a namespace. Any error
  or partial answer keeps the namespace with a warning, as a failed listing does today.
- A namespace the tool creates is labelled. One without the label is deleted only after a
  yes and then the typed namespace name. `--no-prompt` alone never deletes a namespace.
  `--no-prompt --delete-namespace` does, labelled or not. Without a terminal and without
  both flags the namespace is kept.
- Labels never go into a selector, a pod template or the custom resource's `podLabels`.
- Existing deployments are not relabelled (migration is out of scope). Their namespace
  takes the unlabelled path, and their first deploy after adoption restarts.

### Scope

- One label vocabulary in the module: the key names, the value grammar and a helper that
  renders them for each sink.
- Kubernetes: the namespace is labelled when a tool creates it. Every manifest object and
  every Secret and ConfigMap a tool creates carries the labels.
- `nsgate`: the gate reads the namespace's label to choose between one question and two. It
  discounts an object labelled for this tool and this instance, beside the by-name and
  ownerReference discounts it has.
- `nsgate.List`: widen the listing beyond the `all` category, from
  `kubectl api-resources --verbs=list --namespaced -o name`. The custom resource kind and
  the operator's PodDisruptionBudget then show. New discounts are needed for objects that
  come with every namespace or every owned Service, such as the Endpoints of an owned
  Service. An identity that may not list one of the types is never offered the namespace.
- `confirm`: a question that is answered by typing a given name.
- Both tools gain `--delete-namespace`, valid on a Kubernetes remove only. Rule 9 of the
  parity body is rewritten for it and for the two questions.
- Compose: the service carries the labels. Neither tool renders a top-level network or
  volume, and a label on a compose secret or config reaches no engine object.
- Podman: the quadlet unit and every secret carry the labels. Every `secret create` passes
  them, because `--replace` sets a secret's labels to exactly those passed.
- The value is checked per sink. A Kubernetes label value is at most 63 characters with
  alphanumeric ends; solace-util's `container.name` may be 100 long, which is fine on podman
  and compose. A value is never truncated or hashed.
- `containerart/artifacttest`: a check that fails when a rendered object lacks the labels.
- Both tools adopt the labels in one step, and the parity body gains a row for them.

### Undecided

- The full list of objects the wider listing must discount, fixed against a live namespace.
- How a tool learns this module's version for the `lib` annotation. A build through a
  workspace carries no version for a replaced module.
- Whether `generate` output holds a labelled Namespace document. Applying it would label a
  namespace the tool did not create.

### Until then

solmq-conn-util keeps its namespace removal as it is until it adopts this item (decided
2026-11-13). Its listing covers `all`, PVCs, Secrets and ConfigMaps and nothing else.

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
- `quadlet.Escape` leaves `$` alone, and systemd expands it in the `ExecStart` quadlet
  generates (confirmed from source 2026-11-13). This module's CR-0006 doubles it.
- A field held to a grammar refuses `$`. In solmq-conn-util that covers the image, the
  timezone, host paths, names and `java-options`.

### Decided

solmq-conn-util follows solace-util (decided 2026-11-13): it stops expanding `${VAR}`, expands
only a leading `~`, and passes a `${VAR}` in a value on as written.

A `${` that is passed on earns no warning in either tool, and a field with a grammar refuses
it with its normal error (decided 2026-11-14). In solmq-conn-util that makes the documented
`-Xmx${CONNECTOR_HEAP:1g}` a validate error, so its own decision needs a table of fields:
error or literal.

### Scope

- Write the rule under "Rules for shared code" in [developer.md](developer.md) and add a row
  to the parity body.
- Lift the tilde expansion from solace-util (`expandTilde`, `expandTildeToken`) into the
  module, so both tools share one implementation and one set of tests.
- State, per sink, what happens to a `$` that is passed on: the compose file, the quadlet
  unit, the manifest and the application's own config file.
- [developer.md](developer.md) says leading-`~` expansion stays in each tool. That sentence
  is rewritten with the lift.
- In solmq-conn-util the `~` step runs at the one place every loader passes through. It
  parses env.yaml at several sites today and expands at one.

### Undecided

- Which keys count as path keys in each tool, and whether the list lives in the module.
- The package that receives the lifted code.
- Whether solace-util's `convert` keeps resolving a reference to an earlier assignment in
  the same file. Its "expands nothing" decision would remove that too.

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

The order is stated per command (decided 2026-11-14):

- Deploy: validate and resolve, read-only preflight, compare with what is running, ask, act.
- Remove: load the names, ask, preflight for what was answered, act. Remove resolves no
  credential.
- Every question is answered before the first write or delete. The one exception is the
  namespace offer, which follows the remove it depends on.
- Host checks sit in the read-only preflight. Host changes come after the question.
- The module gains no pipeline. The order is a parity rule each tool follows in its own
  code.

### Scope

- Add the order to the parity body as a rule, and check both tools against it.
- Compare each tool's deploy and remove sequence with that rule and list the differences.
  The findings file of the plan review holds a first list.

### Undecided

- Nothing in this item.

## 4. Builders refuse a name outside its grammar

Raised 2026-11-13 from the solmq-conn-util review of the names its `status`, `logs` and `cli`
verbs put on an engine argv.

### Goal

A builder that places a pod, container, deployment or unit name on an argv refuses a name
outside that name's grammar. A tool that forgets its own check still cannot pass a name that
starts with `-` to an engine.

### Today

- The grammars exist: `kube.ValidLabel` and `containercli.ValidName`. There is none for a
  DNS subdomain, which is what a pod, Secret or StorageClass name is.
- Only `kube.NamespaceManifest` holds its argument to a grammar. Every other builder takes
  the name as given: in `kube`, `Logs`, `Exec`, `Upload`, `Scale`, `RolloutRestart`,
  `RolloutStatus`, `Get`, `GetNamespace` and `DeleteNamespace`; `rbac.CanI` and
  `nsgate.List`; in `containercli`, `Start`, `Stop`, `Restart`, `Logs`, `Exec`, `Upload` and
  `inspect.Containers`; all of `podman` and `systemd`.
- `kube.DeleteNamespace` places the namespace as a bare argument, so a dash-led value would
  be read as a flag of `delete namespace`.
- solace-util passes `pod/<name>` to `kube.Logs`, and `kube.Get` takes type lists and
  dotted names. Neither grammar accepts those.
- solmq-conn-util checked lifecycle names with a private copy of the DNS-1123 rule, and its
  `status`, `logs` and `cli` verbs allowed a leading `-`.

### Decided

solmq-conn-util calls the shared grammars at its own boundary and drops the private copy
(decided 2026-11-13).

Decided 2026-11-14:

- A builder that takes a bare name (scale, rollout, exec pod, container, unit, secret)
  checks a typed grammar and returns an error. A subdomain grammar is added.
- A free operand (the targets of `Get` and `Logs`) refuses only an empty value, a leading
  `-` and control characters.
- The secret emitters (`compose.ServiceSecrets`, `compose.TopSecrets`, `quadlet.Secret`)
  check the name, the target and the env var themselves and return an error, like the line
  builder below. The compose target is quoted.
- Every builder changes in the one breaking tag, and both tools adapt in the same step.

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
- The line builder doubles `$` as well as `%`. Quadlet copies these values into `ExecStart`
  unescaped, and systemd expands both there (confirmed from source 2026-11-13). It refuses a
  `:` in a `Volume=` source.
- Size: 21 call sites in solmq-conn-util, 37 in solace-util, and about 50 direct calls in
  this module's tests. solace-util funnels its builders through one function type, which
  changes with them.
- The package comment of `kube` and [developer.md](developer.md) both say builders trust
  their names. Both are rewritten in this step.

### Undecided

- The grammar each builder holds its name to, fixed in this step's plan. `%` stays legal
  and is escaped, as solace-util already does.

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

Decided 2026-11-14, replacing the word `bridge` as the default:

- The env file takes each platform's own word for the network, and the tool checks it
  against the words that platform accepts.
- With no key, the artifact states the platform's own non-host network. Host networking is
  written only when the env file asks for it.
- The word `bridge` is not that default everywhere. Rootless podman's default is its
  user-mode network, and compose `network_mode: bridge` leaves the project network.

### Scope

- A separate `artifacttest` check requires a stated network mode in compose and quadlet
  output. The existing posture checks do not change, so neither tool's tests break before
  it adopts the new one. A manifest is left out: Kubernetes defaults `hostNetwork` to false.
- Candidate spellings for the default: `Network=private` in a quadlet unit, which podman
  reads as bridge when rootful and as the user-mode network when rootless; in a compose
  file, the project network with its bridge driver stated.
- Host networking: `network_mode: host`, `Network=host`, and `hostNetwork: true` where the
  tool renders the pod itself.
- The parity body gains a row for the network mode.
- A read-only root stays a choice of each tool. The posture check accepts either a stated
  read-only root or a tool's recorded reason for a writable one.

### Undecided

- The accepted words per platform, and whether the list lives in the module.
- Whether quadlet accepts `Network=private` on the oldest supported podman. Read from the
  man pages; not run on a host.
- Whether the read-only root check belongs in `artifacttest` at all, given the two tools
  differ. solmq-conn-util's read-only root also needs its quadlet key in the release table
  of `CheckQuadletKeys`.

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

One new small package holds the helper, and [developer.md](developer.md) names it as the
module's one file-writing helper (decided 2026-11-14). It lifts solace-util's sequence: a
temporary file in the target directory, the mode set, the write, a sync, the rename.

### Scope

- One helper that takes a path, the content and a mode, and writes through a temporary file
  in the same directory.
- Both tools use it for quadlet units in the same step.

### Undecided

- The package's name.
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
- `systemctl stop` exits 5 for a unit that is not loaded, and `is-active` exits 4 for a unit
  that is not found and 3 for one that is inactive (from systemd's source). The module has
  no command that tells the two apart.
- `compose down` exits 0 for a project it cannot find. `COMPOSE_PROJECT_NAME` in the shell
  outranks the `name:` both tools write into the file, and neither passes `-p`. After such a
  down solmq-conn-util prints `ok`, and solace-util went on to empty the data directory.
- solmq-conn-util deletes the podman secrets even when the stop failed, so the unit stays
  and points at secrets that are gone.

### Decided

solmq-conn-util skips the stop when the unit file is absent and the unit is not loaded, then
removes what is left and exits 0 (decided 2026-11-13).

Decided 2026-11-14:

- Every compose call passes `-p <project>`, which outranks the shell variable. The project
  name is an env-file entry in both tools (item 19).
- The module gains a builder and parser for a unit's load state.
- After a remove each tool asks the engine whether the container is gone, and fails if it
  is not.
- Secrets are removed only after the unit is confirmed gone.
- The removal sequence stays in each tool. The module shares the builders and the rule.

### Scope

- Add the rule to the parity body and check both tools against it.
- solace-util's operations guide tells operators to export `COMPOSE_PROJECT_NAME` to change
  the project. That passage goes, and the env-file entry replaces it.
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
- solmq-conn-util makes the same two comparisons and has the same defect (read 2026-11-13):
  a daemon reload or a restart that fails after the writes leaves the next deploy reporting
  the service as unchanged.
- Both tools reload systemd only after a write. After a failed reload systemd still holds
  the old unit, so a restart alone starts the old unit again.
- Quadlet turns a `Label=` line into an argument of `podman run`, which any local user can
  read. solmq-conn-util also writes the unit world-readable.
- Podman reports the secrets a container was created with under `Config.Secrets`, each with
  the secret's ID. A replace gives the stored secret a new ID (from podman's source).

### Decided

solace-util stamps the unit with a deploy-digest label and compares it with the running
container's label, as it does on docker (decided 2026-11-13).

Decided 2026-11-14:

- The label holds a digest of the unit text only. Nothing derived from a secret value is
  written to a label, which keeps the decision of 2026-11-06.
- A secret that was written and never applied is found by comparing the secret IDs the
  running container recorded with the store's current IDs for those names.
- A mismatch of either kind counts as changed, and forces a daemon reload before the
  restart even when the unit and the secrets on disk already match.
- The label key stays one key per tool, passed by the caller:
  `solace-connector/deploy-sha256` and `solace-util.deploy-sha256`.

### Scope

- A digest function beside `ComposeDigest` for a quadlet unit. The unit is rendered without
  the label line, hashed, and rendered again with it.
- `inspect` decodes `Config.Secrets` with each entry's ID (this module's CR-0005 fixes the
  decode), and `podman` gains a builder and parser for a stored secret's ID.
- Whether the container is running is asked of the engine, and its label is read with
  `inspect.Container.Labels`.
- Rewrite the "Redeploy over a running service" decision in the parity body. Both tools
  adopt it in one step.
- Every quadlet golden changes, since the unit gains a `Label=` line.

### Undecided

- Whether the IDs behave on a host as podman's source says: the container keeps the ID it
  was created with, and a replace changes the store's. Not run on a host.

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
- solmq-conn-util never reads, prints or pins a context (read 2026-11-13). Its command may
  carry `--context` or `--kubeconfig`. Its one destructive question names no cluster, so it
  cannot name the wrong one, but nothing ties the deletes to the cluster that was current
  when the question was asked.

### Decided

solace-util pins the context, and the helper lives in this module (decided 2026-11-13).

The context is resolved once per run, stored, and pinned on every kubectl call of that run,
in both tools (decided 2026-11-14).

### Scope

- A helper that finds the effective context: a `--context` in the command tokens wins,
  otherwise the answer of `config current-context`.
- A helper that returns the command tokens with `--context <resolved>` added when they
  carry none.
- The resolved name is kubectl output. It is checked before it is placed on an argv, and a
  name that starts with `-` is refused.
- A parity rule for the context, adopted by both tools in one step.
- The helper resolves with the same command tokens it is given, so a `--kubeconfig` in them
  is used. A run that reaches a second cluster, as solace-util's mate channel does, resolves
  and pins each set of tokens.
- Both `--context X` and `--context=X` are recognised in the tokens. The pin is written
  `--context=<name>`. A context name is an arbitrary string: EKS writes `:` and `/` into it.
- For solmq-conn-util this is new behaviour, and its tests that script kubectl calls in
  order gain a first call.

### Undecided

- What a run does when the kubeconfig has no current context, which is how kubectl runs
  inside a cluster. The review advises pinning nothing and saying so.
- Whether solmq-conn-util prints the resolved context, for example on its "acting on" line.

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

`Delete` and `DeleteNamespace` take a timeout, with a 300s default constant beside the
rollout one. The owner times one real broker remove before the tag, and the default follows
(decided 2026-11-14).

### Scope

- `Delete` and `DeleteNamespace` write `--timeout=<t>`, with a default constant as
  `RolloutStatus` has.
- solace-util builds a few deletes itself (the PVC, pods, Secrets). Each gets the same flag
  in the tool.
- When the bound expires the delete has been accepted and continues in the cluster. The
  message says the deletion was requested and is still in progress, and the namespace gate
  keeps the namespace with that warning.
- Rule 1 of the parity body names the flag.
- Both tools adopt it in one step.

### Undecided

- Whether the timeout arrives as a parameter or in an options struct.
- Whether solace-util goes on to delete its Secrets after the custom resource's delete
  timed out.

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

The loop asks `exitcode.IsUsage` itself. A refresh error that is a usage error ends the loop
and is returned; every other error is retried (decided 2026-11-14). solmq-conn-util's
refresh reports its own errors and returns none, so nothing changes for it.

### Scope

- The loop returns a refresh error that is a usage error, and retries the rest.
- Rule 13 of the parity body says which errors end the loop.
- A test that such an error ends the loop after one pass, and that others are retried.

### Undecided

- Nothing in this item.

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
- solmq-conn-util accepts a count from 0 to 100000 and the word `all`, and documents both.

### Decided

solace-util accepts a count only, and leaving the flag out means every line (decided
2026-11-13).

Both tools accept a count only (decided 2026-11-14). solmq-conn-util drops the word `all`
from its flag, help, docs, completions and tests. Its 100000 cap stays its own.

### Scope

- One function that checks a tail value, called by both tools at their boundary.
- Rule 6 of the parity body states the grammar.

### Undecided

- Whether the `Logs` builders refuse a bad value themselves, now that item 4 gives them an
  error result.
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

Decided 2026-11-14:

- A default run sets `GOWORK=off` itself. A `GOWORK` the caller set, without `--local`, is
  refused.
- `--local` uses the tool's own `go.work`. It stops unless that file exists, replaces this
  module, and carries a `toolchain` line equal to `go.mod`'s. With a workspace in play Go
  ignores `go.mod`'s `toolchain` line, and the GOTOOLCHAIN export that hid this leaves the
  scripts (this module's CR-0017).
- When the pinned version is not published yet, the error says to pass `--local`.
- `dev-all` forwards `--local` to the two tools only when it is given, never to this
  module's own script, and never with `pin` or `resolve`.
- `full --local` is allowed, and says at the start and in the footer that it is not the
  release build. Both tools must pass their complete gate against this checkout before the
  module is tagged.

### Scope

- `dev-all.sh` and `dev-all.ps1` accept `--local` and hand it to each tool's script.
- The notice line follows the flag as well as `GOWORK` and `go.work`.
- "Develop against a local checkout" and steps 2 to 4 and 7 of "Release and merge order" in
  [consuming.md](consuming.md) are rewritten for the flag.
- solmq-conn-util's dev scripts gain the same flag in the same step. Its decision to have
  `full` set `GOWORK=off` (its CR-0037) is replaced by this item.
- The init-devops dev-script contract records the flag and the `go.work` check.
- `dev-all` refuses any word no dev script knows, and its notice line reads `GOWORK` and
  `go.work` only. Both are changed for the flag.
- Each tool's `go.work` gains its `toolchain` line. Neither has one today, the installed Go
  is 1.27.3 and both tools pin go1.27.4.
- solace-util flips its default only after it holds the first module tag it needs.

### Undecided

- Nothing in this item.

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

- The module owns the behaviour and both tools import it (decided 2026-11-13): the label
  keys and values, the helper that writes them for each sink, and the steps that list,
  compare and delete. A tool passes its name, the instance and the names its current config
  renders. The design is worked through in this repository.
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
- `secret create --label` arrived in podman 4.3, before quadlet, so no version floor moves.
  `--replace` (4.7) sets a secret's labels to exactly those passed, so every create passes
  them. Read from podman's release notes and source; not run on a host.
- A listing that fails or comes back partial is a failed listing: the tool warns, deletes
  nothing and continues (decided 2026-11-14). `secret inspect` prints what it found and
  exits non-zero when one of the names is gone.
- The unit gains `Label=` lines too (item 1), and a changed unit asks the redeploy question
  and restarts. So "asks nothing" below holds for the secrets, and the first deploy after
  adoption restarts each service once. That is accepted (migration is out of scope).
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

- What a failed removal does. The review advises a warning on deploy, and a failed remove
  when a secret that holds a private key survives. solace-util's remove treats a surviving
  certificate secret as fatal today.

The label keys are decided in item 1. The shape is decided (2026-11-14): builders, parsers
and one pure compare function, which each tool runs.

## 17. A readiness wait for OnDelete StatefulSets

Raised 2026-11-13 from the solace-util review of `broker start` and `broker restart`
(solace-util CR-0083). Suspected, not confirmed: it waits on a check against a cluster.

Moved to solace-util (decided 2026-11-14). It is the only tool that manages a StatefulSet,
so the wait is built there and the module does not change. The item keeps its number.

The operator's source supports the suspicion: it sets `OnDelete` on its StatefulSets on
every update pass, whatever the custom resource's update strategy says, and kubectl then
answers `rollout status` with "rollout status is only available for RollingUpdate strategy
type" and exit 1. Read on the operator's main branch on 2026-11-13, not on the bundled
1.4.2. The check should be run on `-p`, `-b` and `-m`, under both update strategies.

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

- How the wait treats a pod that does not exist yet. A readiness wait right after a pod
  delete or a scale-up meets NotFound.

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
  Nothing removes the leftover. One exception: solmq-conn-util names its Service among its
  own objects whether or not env.yaml asks for it, so the gate discounts a switched-off
  Service.
- solace-util's remove keeps the admin Secret on purpose, because kept data needs it.
- solmq-conn-util's libs PV is NFS, read-only and `Retain`, and its claim holds no data.
  solace-util's data PVCs are made by the operator and carry no tool label.
- Docker needs no such step. A compose secret is sourced from the environment and copied
  into the container, and nothing is stored outside it. A config change recreates the
  container, and the old container goes with what it held.

### Decided

On Kubernetes the removal goes by ownership labels, not by reading the last deploy. Every
object a tool creates carries the labels of item 1. A deploy or a remove lists the
instance's labelled objects and deletes the ones the current config no longer names. An
unlabelled object is never deleted. Docker needs no step. Both tools are to be aligned on it
(decided 2026-11-13).

Decided 2026-11-14:

- The removal covers Secrets, solmq-conn-util's Service, and its libs PVC and PV. The PV is
  deleted by its derived name after a cluster-wide delete probe. A refused probe warns and
  leaves the PV.
- solace-util's admin Secret is exempt while data is kept. The names to keep are the ones
  the config references, not only the ones the tool builds.
- A listing that is refused, fails or comes back partial: the tool warns, deletes nothing
  and continues.
- The shared code is builders, parsers and one pure compare function, which each tool runs.
- "An unlabelled object is never deleted" is about leftovers. The current set is still
  deleted by name, as today.

### Scope

- The module owns the behaviour and both tools import it, as in item 16 (decided
  2026-11-13): one label vocabulary, one helper that writes the labels into a manifest, and
  the steps that list, compare and delete. The design is worked through in this repository.
- The listing is by label selector in the instance's namespace, for the kinds the tool
  creates. The labels name the tool and the instance, so two instances in one namespace
  never see each other's objects. It asks for names only, never the content of a Secret.
  kubectl prints the kinds it could list and exits 1 when one is refused, which counts as a
  failed listing.
- Item 1's label vocabulary lands first or in the same step. From then on every object a
  tool creates carries the labels.
- The delta is the listed names minus the names the current config renders. The tool names
  each object it removes.
- A parity rule for the removal, which also states that docker has nothing to remove. Both
  tools adopt it in one step.
- `delete` is probed for each kind that has a leftover, in the preflight the command already
  runs before its first write (rule 2). A refused probe warns and deletes nothing. The
  listing is a read and is not probed.
- Questions come first and actions after (decided 2026-11-13). Every question a deploy or a
  remove asks is answered before its first apply or delete. Deleting a leftover asks nothing
  of its own, and the tool names what it deleted (decided 2026-11-13).
- solmq-conn-util's teardown deletes the listed leftovers as well as the current set.
- Known limit: an object of a deployment made before the labels exist gains them at the
  next deploy only while the config still names it. One dropped before that stays unlabelled
  and is never found, so it is removed by hand.

### Undecided

- Whether the tool lists and deletes by name, or hands the delta to kubectl's own
  `apply --prune`. The review advises listing: `--prune` is still alpha, and listing lets
  the tool name what it removes before it does.
- Where the delete sits after a deploy's apply, and what a failed delete does (see item 16).
- The wording of the line that names a deleted leftover. It is settled when the work is
  built.

## 19. One shape for the platform sections of the env file

Raised 2026-11-14 from the decision on item 9.

### Goal

The same thing has the same key under the docker, podman and kubernetes sections of both
tools' env files. A difference that remains is one the two workloads force, and it is
written down.

### Today

- The compose project name is `docker.project-name` in solmq-conn-util, with a default.
  solace-util has no key for it: the name is derived from `docker.container.name`, and its
  code leaves the `COMPOSE_PROJECT_NAME` shell variable working as the override.
- Neither tool has a network key. Item 6 adds one to both.
- The sections were written separately and have not been compared key by key.

### Decided

The project name is an env-file entry in both tools, and both pass it to compose with `-p`
(decided 2026-11-14).

### Scope

- Compare the two tools' platform sections key by key and list every difference.
- Decide each one: align the key, or record it under "Inherent differences" in the parity
  body.
- The parity body gains a row for the shared keys.

### Undecided

- Every key but the project name.
- Whether solace-util's project key defaults to the name derived from the container name,
  as today, or to a fixed one.

## Inputs from the solace-util review to items 1 to 4 and 9

Decided in solace-util on 2026-11-13. These add no item. The IDs are from solace-util's
`docs/review/findings.json`.

- Item 1: solace-util parks its namespace-gate finding (CR-0006) on the wider `nsgate.List`
  and changes nothing in the tool until it exists. The owner also decided that the labels
  and the removal that reads them are one standard behaviour in this module, which both
  tools import (items 16 and 18, 2026-11-13).
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
  podman (CR-0005). solmq-conn-util's docker remove has the false success and not the data
  loss, since it empties nothing (read 2026-11-13). Item 9 now holds the shared answer.
- Parity body, Host limits: solace-util adds a cgroup v2 check for rootless podman
  (CR-0089). The check stays in the tool, and the list under Inherent differences gains it.
