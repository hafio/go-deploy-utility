# Surface review, 2026-11-15

Reviewed: everything go-deploy-utility (M) exports, against the shape decided in
[../design/2026-11-15-shared-deploy-api.md](../design/2026-11-15-shared-deploy-api.md):
clients that run one operation each, rendering in the module, and the rule that common
artifacts and operations belong to the module. It replaces section 6 of
[2026-11-13-plan-review.md](2026-11-13-plan-review.md), which was written before that
design. Nothing was built, run or edited.

How: 45 read-only reviewers in four streams, each with its own adversarial check, and one
critic over all of them.

- **Surface**: every exported name, in 8 package groups. One reviewer per group, then one
  verifier that re-counted every use in solmq-conn-util (A) and solace-util (B), and one
  that challenged every recommendation.
- **Operations**: every external command each tool runs, mapped onto the planned clients,
  then verified.
- **Renderers**: every field each tool writes into its compose file, quadlet unit and
  Kubernetes documents, a typed model per artifact, then verified.
- **Env keys** (gdu 19): both tools' platform keys, aligned key by key, then verified.

Their full output is in
[2026-11-15-surface-review-findings.md](2026-11-15-surface-review-findings.md). The lines
behind sections 2 and 4 below were read again for this report; the rest are leads with
their evidence cited.

Counts:

- 245 exported names: 155 stay, 48 go private behind a client method, 27 go private because
  only the module uses them, 5 are removed, 6 move package, 4 change signature. 47 of those
  calls are the owner's to make. The verifiers disputed 31 and corrected 9 usage counts.
- A runs 48 distinct external command shapes and B runs 87. They map onto 58 client
  methods: 36 used by both tools, 4 by A only, 13 by B only, 5 by neither today (decided
  items need them). 13 common operations have no method in the catalogue yet.
- Rendering: A writes 267 fields or lines over its three artifacts and B 311. The three
  models hold 208 fields, 41 workload extras and 73 places where the tools write the same
  thing differently.
- Env keys: 110 in A, 129 in B, 61 concepts. 7 are aligned, 24 align to one tool, 7 need a
  new common key, 20 are forced by the workloads (the verifier re-argued 6 of those) and
  3 rows came back without a label.
- 90 questions for the owner from the streams, and from the critic 29 contradictions, 19
  gaps, 14 unverified points.

## 1. Verdict

1. **The direction holds.** 48 of today's builders and parsers are exactly what a client
   method would run, and each tool wraps them in its own capture, classify and error code.
   Both tools hand-write the same Secret, compose and unit shapes. Nothing found argues
   against executors or against rendering in the module.
2. **The design is not ready to plan.** Three streams each designed their part alone and
   disagree with one another, and five points disagree with what is recorded as decided.
   Section 2 lists the rulings that unblock it.
3. **The models need one pass together.** The compose, quadlet and Kubernetes models give
   the same container three different shapes (identity, run-as, mounts, defaults, the slot
   for extras). About 14 new packages are proposed with no import graph.
4. **Nothing here ran.** Several model choices rest on engine behaviour stated from memory
   (section 5), and smoke S7 to S9 are still open.
5. **The review found defects in today's code** (section 4), some of them about secrets.

## 2. Rulings needed, in order

Each line gives the reviewers' recommendation. The numbers in brackets are the critic's
contradictions in the findings file, section 5.

1. **Where a command's output goes [1].** Decided: the tool's hook sees every byte. The
   operations map and three verifiers say that cannot stand: a method that parses or
   returns its stdout must keep it, because at least four reads carry secrets (a podman
   secret read-back, both Secret reads, the namespace listing, a raw container inspect).
   Recommended: stdout that a method returns never reaches the hook; stderr does; a method
   that returns nothing sends both; a tool may pass a hook for one call.
2. **A generic run on every client [7].** Decided: a command the module does not offer, the
   tool builds itself. Without a way to run it through the client, it skips the base
   tokens, the pinned context, the session environment, the hook and the result.
   Recommended: `Run` and `Output` on each client over one exported primitive.
3. **How a client is set up [4].** Decided: a struct the tool fills in. Recommended: a
   constructor that checks the command tokens with `execguard` once and fails at set-up,
   since an executor must not assume its caller checked.
4. **Typed errors [13].** One set for every client (refused, not found, unanswered, timed
   out, unsupported, no context, partial listing), defined once in `runner`, carrying the
   command's stderr as a field.
5. **The exec family [2, 3].** One `Exec` per client cannot serve a streamed run, a
   captured run, a run with stdin and an attached session. Recommended: `Exec`,
   `ExecOutput` and `Attach`, and one shape for the child's exit status.
6. **Which one-tool operations move into the module [15].** Under the new rule: image
   inspect, stats and top (A); registry login, a Secret value read, compose and server
   version, `podman unshare`, copy in and out, `systemctl show` (B). The design note names
   image inspect as tool-built, so this is a change to a recorded point.
7. **Host probes [12].** Reading the subid files, the runtime directory and the account
   name: recommended to stay in the tools, as executor rule 4 says.
8. **Release order [24].** gdu 2 and gdu 19 are recorded as following the breaking tag,
   while "the next tag ships everything". The models take the tools' config, and the env
   table renames 75 of B's keys: if gdu 19 follows, each tool maps its config twice and
   B's operators meet two breaking env-file changes. Recommended: both inside the tag.
9. **One container description [19, 20, 21].** The three models must share their leaf
   types. Open inside it: zero values that are the hardened ones (a forgotten field must
   not mean read-write, root or host), an allow list for the extras slot, whether the root
   filesystem line is always written (gdu 6 leaves it to each tool), ports together with
   host networking (refuse or drop), restart with or without a default, numeric run-as.
10. **The lib annotation against a pure renderer [25].** A renderer that reads build info is
    not pure, and every golden would change with the module version. Recommended: one
    small function reads the version; the renderer takes it as a value; tests pass a fixed
    one.
11. **The env file (gdu 19).** Kebab-case everywhere (B renames its camelCase keys), one
    section layout, one name grammar, no detectors for renamed keys, and the other
    alignment rows as a set.
12. **The preview.** Design rule 6 keeps a preview through `runner.Echo`. No tool ships a
    preview today: B's `--dry-run` is retired and only its tests install `Echo`. Is it a
    requirement?
13. **Decided items to correct.**
    - gdu 13 says nothing changes for A. A's refresh does return its error to the loop
      [28].
    - gdu 1 says neither tool renders a top-level network; gdu 6 makes the compose file
      state one [29].
    - gdu 6's scope still lists `hostNetwork: true`; the confirmed reading is that
      Kubernetes gets no network key [29].
    - gdu 1 assumes a container name of up to 100 characters; the env table narrows it to
      a DNS label of 63 [29].
    - The catalogue lists `podman.Client.Version`; no tool needs it once the client picks
      the fallback itself [5].
    - Design rule 6 says a parser reports "unknown" for an empty answer. `ParseNamespace`
      reads empty as absent, and `systemd.Stopped` accepts the word unknown as stopped
      [26].

## 3. What each stream found

### 3.1 Surface

- Private behind a client method (48): 12 in `containercli`, 12 in `kube`, 11 in `systemd`,
  7 in `podman`, 3 in `kube/rbac`, 2 in `containercli/inspect`, 1 in `kube/nsgate`.
- Private because only the module uses them (27): the compose and quadlet line helpers once
  the renderers exist, the `Check*` functions behind the `Assert*` ones, `confirm.Line`,
  `Yes` and `YesNo`, `watch.CheckInterval` and `Sleep`, `rbac.Answer` and `Refused`, and
  others.
- Removed (5): the four of CR-0001 and `abbrev.Set.Kind`.
- Disputed by the verifiers, among others: `inspect.Decode` stays exported as a pure parser;
  `exitcode.ChildStatus` stays; the two `Posture` constants move to an internal package;
  `systemd.Systemctl` goes private if the client gains `Show`, `Status` and `Cat`;
  `nsgate.DeleteProbe` should ask about one namespace, not all of them;
  `confirm.Prompter.Gate` cannot word the decided namespace question as it stands.
- The reviewers also listed 82 pieces of behaviour the module should own, 68 placement
  notes and 177 signature notes. They are inputs to the plan, not questions.

### 3.2 Operations

- 16 methods on `containercli.Client`, 8 on `podman.Client`, 9 on `systemd.Client`, 22 on
  `kube.Client`, plus `rbac.Run`, `nsgate.Read` and one primitive under them all.
- Not in the catalogue yet (13): delete by kind and name with a bound, exec that returns
  stdout, a generic run, Secret reads that keep values out of the output, `systemctl show`,
  compose version, registry login, copy in and out, `podman unshare`, server version, image
  inspect, resource usage, and a timed-out error class.
- Both tools build the same things by hand outside any builder (10): the capture wrapper
  around every call, the platform switch for exec, logs and upload, the session
  environment for systemctl, the quadlet unit's write, compare and remove, and the
  re-check of the command tokens.
- Of 30 operations only one tool runs, 15 are judged common and 15 stay with the tool.
  That judgement is one stream's and was not verified.

### 3.3 Renderers

- Compose: a 59-field model, 10 extras, 22 differences. Quadlet: 38 fields, 11 extras, 27
  differences. Kubernetes: 111 fields, 20 extras, 24 differences.
- One writer cannot serve all three as modelled: the Kubernetes model has no flow
  sequences and never doubles `$`, and compose needs both [17].
- The word lists the renderers need (`Widening`, `HostNetwork`) sit in `artifacttest`,
  which imports the renderer packages and `testing`, so a renderer cannot import them [18].
- The verifiers' main theme: a slot for extras opens a key set that is closed today, so it
  needs an allow list, and every boolean must default to the safe side.

### 3.4 Env keys

- The tools differ most in spelling and layout: B nests under `container` and writes
  camelCase; A nests the Kubernetes instance under `deployment`.
- Proposed common keys: the network word, restart, resources, run-as, the image keys, the
  Kubernetes Secret block, placement, the Service ports.
- The verifier showed that six "forced" differences are choices (read-only root, spread,
  pod identity, service account, the second image's registry, the TLS Secret) and that B
  cannot take the run-as word `image`, because it must chown its data directory.

## 4. Defects found on the way

Reported by the verifiers with file and line in the findings file; the critic confirmed the
ones marked so. Each belongs to its own repository's session.

| Where | Defect |
| --- | --- |
| A and B | The namespace gate's listing asks for `secrets -o json`, so every Secret value in the namespace is read into the tool. In A a failed listing prints the captured output in its error (confirmed: `internal/runner/runner.go:870-876`, `cmd/solmq-conn-util/namespace.go:103-107`). |
| A | `status --details` reads each referenced Secret in full only to learn that it exists (confirmed: `cmd/solmq-conn-util/status.go:582-585`). |
| A | `remove` resolves the real credentials and stops when one is unset, against "remove resolves no credential" (decided, not built). |
| A | `restart: no` is accepted and written bare into the compose file (confirmed: `internal/validate/validate.go:1686`, `internal/dockergen/dockergen.go:184`); a YAML 1.1 reader such as docker-compose v1 takes a bare `no` as false. |
| A | A Kubernetes name that starts with a digit passes the load and fails the Service. |
| A | Its doc test runs `noticetest` without `Strict`, so it passes when a notice is missing. |
| B | The quadlet `Volume=` line writes the data directory raw (confirmed: `internal/render/render.go:985`); the verifier reports that `%` is not escaped and `:` is not refused. |
| B | The run-as value admits names and a leading `-`, and `User=` and `Group=` are written raw. |
| B | `healthCheck.retries` has no check: a negative count reaches the artifact. |
| B | A standalone router name taken from the host name is never checked. |
| M | `kube.ValidLabel` accepts `123` and `true`, which a bare YAML scalar turns into a number or a boolean (confirmed). |
| M | `ParseNamespace` reads an empty answer as "absent", so a read that printed nothing looks like a missing namespace (confirmed). |
| M | `systemd.Stopped` accepts the word `unknown` as stopped (confirmed). |
| M | `execguard.IsControl` documents C1 controls and leaves U+0085 out. |
| M | `nsgate.DeleteProbe` asks for delete on every namespace to delete one. |
| M, compose | Inline config content is not checked for the line breaks NEL, LS and PS, which end a block scalar and let the rest be read as compose keys. |

## 5. Not verified

- Nothing ran on an engine or a cluster. Smoke S7 to S9 and the v0.1.3 checks are open, and
  A's Kubernetes path has never run on a real cluster.
- Stated from memory, each carrying a model decision: that quadlet reads `Label=` like
  `Environment=`; that an absolute `Volume=` source is copied into `RequiresMountsFor`;
  that a non-JSON `HealthCmd` runs through a shell; that compose copies environment
  secrets into the container and fails under `read_only: true`; that `-p` outranks
  `COMPOSE_PROJECT_NAME`; the kubelet's `$(VAR)` expansion in an exec probe; systemd's
  exit codes 5, 3 and 4.
- What a dependency's version reads as under `go test`, under `go.work`, and when the
  module is the main module. The lib annotation rests on it.
- What a hook writer costs: with a non-file writer every run is piped, so compose and pull
  progress change. Nobody measured it.
- The roughly 150 additions the second verifiers made were checked by nobody; the critic
  sampled four and they held.
- The owner checks already on record (decisions file, "Checks the owner runs").

## 6. Next work

1. The rulings of section 2.
2. One catalogue from the surface review and the operations map: per method its signature,
   the grammar of each argument (the table gdu 4 leaves to its plan), its typed errors, and
   the builder or parser that goes private.
3. One container description and one writer for the three renderer models.
4. The label design note the plan review asked for: keys, per-sink value checks, the
   identity type, the version reader, and who owns the `--delete-namespace` flag name.
5. A package inventory and import graph for everything proposed, test support included.
6. The namespace gate as a whole for gdu 1: a listing that keeps Secret values out yet
   yields type, annotations, owner references and labels.
7. The scope of the guide and the example tool: where the nested module sits and how it is
   gated, since every dev-script task runs on `./...` from the root.
8. Three edit lists nobody owns yet: the parity body, `docs/developer.md` with the package
   docs, and `docs/test.md` under the rule that a lifted test keeps its name; plus the
   smoke rows for the new executors and artifacts.
9. The design of gdu 7 (atomic write), checked on Windows, where CI gates.
10. One sweep of both tools for parallel code outside these four lenses, starting with B's
    `internal/output`, the two `examples` packages and the generate output writer.
11. A check of the client and env-key design against the recorded later step "root
    deploying on behalf of a rootless user", so it does not need a second break.

Then the plan for the breaking tag.
