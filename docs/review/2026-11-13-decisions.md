# Decisions on the 2026-11-11 critical review

Owner decisions on every finding in
[2026-11-11-critical-review.md](2026-11-11-critical-review.md), taken 2026-11-13, and on the
points [2026-11-13-plan-review.md](2026-11-13-plan-review.md) left open, taken 2026-11-14.
Nothing here is built yet. Evidence and suggested diffs stay in the two reports.

Home says where the work lands:

- `here`: this repository.
- `gdu N`: item N of [docs/todo.md](../todo.md), for the part both tools share.
- `init-devops`: the skill's templates, skeletons or contract first, then a re-deploy here.
- `-`: no work.

## Rules the owner stated

- Questions come first and actions after. The one named exception is the namespace offer,
  which follows the remove it depends on (2026-11-14).
- Migration is out of scope (2026-11-14). A deployment made before a change is not carried
  over: no back-fill, no fallback kept for old deployments, no transition warning. A restart
  on the first deploy after a change is accepted.
- Behaviour both tools share is built once in this module and imported. The shared code is
  builders, parsers and pure functions. Each tool runs them in its own order (2026-11-14).
- Who executes: the code steps are done in the main session, and any agent it uses for them
  runs on Opus. The owner runs every gate (2026-11-14).

## Releases

Decided 2026-11-14.

1. A patch tag first: CR-0005, CR-0006 and the small fixes marked `patch` below. No parity
   change and no API break.
2. One breaking tag for everything else: the builders, the removals, the labels, the digest
   and the network mode. Both tools adopt it in one round. Their adoption work stays on a
   branch until this tag, so each tool can still ship its own fixes from main.
3. The rules that only need the two tools aligned follow: gdu 2, 3, 15 and 19.

Every step is still planned, approved and gated on its own.

### Shipped

- v0.1.3, tagged 2026-11-14 at `b9dfcf8`: CR-0002, CR-0005, CR-0006, CR-0007, CR-0008,
  CR-0010 to CR-0014, CR-0022, CR-0024 and statement 1 of CR-0025. No API change.
  Coverage held at 99.7%. Both tools repinned on main the same day (solmq-conn-util
  `ef112aa`, solace-util `12e68a7`), with their gates green against the checkout and
  against the published tag; their coverage held at 96.3% and 97.7%. The new
  `internal/build` package holds the helpers CR-0007 moved. The host checks for CR-0005
  and CR-0006 are the v0.1.3 section of [smoke.md](../smoke.md), not yet run.

## Decisions on the findings

| ID | Severity | Decision | Home |
|----|----------|----------|------|
| CR-0001 | minor | Both tools take their flag name from `execguard.ApproveFlag`. `watch.Loop` and `runnertest.FailOn` stay. `containercli.Remove`, `paritydoc.Body`, `paritydoc.Compose` and `confirm.IsNotTerminal` are removed in the breaking tag. | here |
| CR-0002 | minor | The package doc names the two consequences (`docker --config`, `kubectl --kubeconfig`) beside the boolean-flag case, and a test pins them. No behaviour change. Patch. | here |
| CR-0003 | major | The listing is widened to every namespaced type that can be listed. Object labels are one more way the gate recognises the instance's own objects. A namespace without the tool's label needs a yes and then the typed name. `--no-prompt` deletes a namespace only together with `--delete-namespace`. | gdu 1 |
| CR-0004 | minor | The secret emitters check the name, the target and the env var themselves and return an error. The compose target is quoted. | gdu 4 |
| CR-0005 | minor | Confirmed from podman's source on 2026-11-13. The decoder reads `Config.Secrets` and keeps the top-level key as a fallback. The fixture takes podman's real shape. Patch. | here |
| CR-0006 | minor | Confirmed from quadlet's source on 2026-11-13. `quadlet.Escape` writes `$` as `$$`. The line builder of gdu 4 does the same. Patch. | here, gdu 4 |
| CR-0007 | nit | The duplicated helpers move into an `internal/` package. Patch. | here |
| CR-0008 | nit | The two `dev-all` scrapers use one pattern, and developer.md states the shape a dev script must keep. Patch. | here |
| CR-0009 | major | Deferred by the owner. Smoke S7 to S9 are run later, with the rows listed under "Checks the owner runs". | here |
| CR-0010 | minor | The podman flag spellings join both forbidden lists, and the match folds case. Patch. | here |
| CR-0011 | minor | CRLF is normalised, and a `secrets:` line in a shape the check does not recognise is an error. Patch. | here |
| CR-0012 | minor | A Windows-only test opens `CONOUT$` and skips without a console. docs/test.md says the branch is otherwise smoke-only and the coverage total is Linux-only. Patch. | here |
| CR-0013 | nit | `CheckDocs` errors when it checked no doc. Patch. | here |
| CR-0014 | nit | The line count and the fixed heading index leave `TestBodyShape`. Patch. | here |
| CR-0015 | minor | Accepted: the gates run locally and the tag follows. The contradictory header comment in `ci.yml` is corrected with the template work. | - |
| CR-0016 | minor | The scripts add `-race` when gcc is on PATH and log which mode ran. `cov` races too, in atomic mode, so CI's `all cov scan` keeps race detection. | init-devops |
| CR-0017 | minor | CI runs `all cov scan`. `test` skips the suite when `cov` is queued. gofmt comes from the selected toolchain. The GOTOOLCHAIN export leaves the scripts. `--local` refuses a `go.work` whose `toolchain` line is missing or differs from `go.mod`'s, and the contract records that check. | init-devops, gdu 15 |
| CR-0018 | minor | `cov` prints the previous total before it truncates the log and warns on a drop. | init-devops |
| CR-0019 | minor | All four differences between `dev.sh` and `dev.ps1` are fixed. The PowerShell skeleton's timestamp is fixed first, so a rebuild does not undo it. | init-devops |
| CR-0020 | nit | `pin` has no branch check. It refuses a tool whose `go.mod` or `docs/parity.md` holds changes beyond what `pin` itself writes, and it checks both tools before the first write. The adoption-branch rule is retired: a tool adopts on main or on a branch of the owner's choosing. CLAUDE.md, developer.md and consuming.md follow. | here |
| CR-0021 | nit | Accepted: `dev-all.sh` runs on Linux only. | - |
| CR-0022 | nit | `graphify-out/go.mod` takes the graph out of the module zip. The same is worth doing for `docs/review/`, which now holds a 306 KB findings file. Patch. | here |
| CR-0023 | nit | Closed by gdu 4: every builder that takes a name checks it and returns an error. | gdu 4 |
| CR-0024 | nit | `inspect.Printable` strips format characters (category Cf) too. Patch. | here |
| CR-0025 | nit | Statement 1 (the Kubernetes spellings and B7) is reworded now. Statement 2 (who checks names) is rewritten by gdu 4. Patch for statement 1. | here, gdu 4 |

## Decisions on the plan review

Taken 2026-11-14. Each is written into its item of [docs/todo.md](../todo.md).

### Labels (gdu 1, 16, 18)

- Labels, on every sink: `app.kubernetes.io/managed-by=<tool>`,
  `services.solace.com/product=<tool>`, `services.solace.com/solution=accelerator`,
  `services.solace.com/developed-by=apac-psg` and `deploy.solace.com/instance=<instance>`.
- Annotations, on Kubernetes only and for information only:
  `deploy.solace.com/version=<the tool's version>` and
  `deploy.solace.com/lib=go-deploy-v<this module's version>`.
- A destructive step selects on `services.solace.com/product` and
  `deploy.solace.com/instance` only. `managed-by` is written and never read.
- A namespace the tool creates is labelled. A namespace without the label is deleted only
  after a yes and then the typed namespace name. `--no-prompt` alone never deletes a
  namespace; `--no-prompt --delete-namespace` does, labelled or not.
- A label listing that is refused, fails or comes back partial: warn, delete nothing,
  continue.
- The leftover removal on Kubernetes covers Secrets, solmq-conn-util's Service and its libs
  PVC and PV. The PV is deleted by its derived name after a cluster-wide delete probe; a
  refused probe warns and leaves it. solace-util's admin Secret is exempt while data is kept.
- solmq-conn-util's old-unit read goes in the same step as the labels (migration is out of
  scope).

### Other items

- gdu 2: a `${` that is passed on earns no warning. A field with a grammar refuses it with
  its normal error.
- gdu 3: the order is stated per command. Deploy: validate and resolve, read-only preflight,
  compare, ask, act. Remove: load the names, ask, preflight for what was answered, act.
  Remove resolves no credential. The module gets no pipeline.
- gdu 4: bare-name builders check a typed grammar and return an error. Free operands refuse
  only an empty value, a leading `-` and control characters. A subdomain grammar is added.
  Every builder changes in the one breaking tag.
- gdu 6: the env file takes each platform's own word for the network, and the tool checks it
  against the words that platform accepts. The default is the platform's non-host network,
  stated in the artifact. Host networking is written only when the env file asks for it.
- gdu 7: one new small package holds the atomic write. The developer guide names it as the
  module's one file-writing helper.
- gdu 9: every compose call passes `-p <project>`, and the project name is an env-file entry
  in both tools (gdu 19). The module gains a unit load-state builder. After a remove each
  tool asks the engine whether the container is gone. Secrets go only after the unit is
  confirmed gone. The sequence stays in each tool.
- gdu 10: the label holds a digest of the unit text only. An unapplied secret is found by
  comparing the secret IDs the running container recorded with the store's current IDs.
- gdu 11: the context is resolved once per run, stored, and pinned on every kubectl call.
- gdu 12: the two delete builders take a timeout with a 300s default. The owner times one
  real broker remove before the tag.
- gdu 13: the loop ends on a refresh error that `exitcode.IsUsage` reports.
- gdu 14: `--tail` takes a count only in both tools. solmq-conn-util drops the word `all`.
- gdu 15: a default run sets `GOWORK=off`. `--local` uses the tool's `go.work`. A `GOWORK`
  set by the caller without `--local` is refused. `dev-all` forwards `--local` to the tools
  only when it is given. `full --local` is allowed and says loudly that it is not the release
  build: both tools must pass their complete gate against this checkout before the module is
  tagged.
- gdu 17: the readiness wait is built in solace-util, since it is the only consumer.
- gdu 19, new: the platform sections of the two tools' env files are aligned key by key,
  starting with the compose project name.

## My readings, to correct if wrong

- `--delete-namespace` has an effect only together with `--no-prompt`. At a terminal the
  questions are asked as usual: one yes for a labelled namespace, a yes and the typed name
  for an unlabelled one.
- The `services.solace.com` labels go on every sink, as the other labels do.
- When the network key is absent the artifact states the platform's own non-host default:
  `Network=private` in a quadlet unit, the project network with its bridge driver in a
  compose file. Kubernetes gets no network key.
- gdu 11 resolves once per set of command tokens, so a `--kubeconfig` in them is honoured
  and a run that reaches a second cluster pins each. With no current context nothing is
  pinned and the run says so.

## Review points carried into the items without a separate decision

The plan review raised these and nothing decided since contradicts them.

- The gate keeps its by-name roots and its ownerReference chain. A label is one more
  discount, for this tool and this instance only (P5).
- Labels never go into a selector, a pod template or the custom resource's `podLabels` (P6).
- Every secret create passes the labels, because `--replace` sets them to exactly those
  passed (P10).
- A label listing asks for names or metadata only, never `-o json` on Secrets (P10).
- A digest mismatch forces `daemon-reload` before the restart (P12).
- The deploy-digest key stays one key per tool, passed by the caller (P13).

## Still open

- gdu 1: the list of objects the wider listing must discount, fixed against a live
  namespace. How a tool learns this module's version for the `lib` annotation, since a
  workspace build carries none. Whether `generate` output holds a labelled Namespace
  document.
- gdu 2: the package that receives the lifted `~` code, and each tool's list of path keys.
- gdu 6: the accepted words per platform.
- gdu 16 and 18: what a failed removal does. The review advises a warning on deploy, and a
  failed remove when a secret that holds a key survives.
- gdu 19: every key but the project name.
- The two tools' own files: the plan review lists what each still lacks (P33 to P44).

## Checks the owner runs

- CR-0009: smoke S7 to S9, with rows added for a foreign Ingress or RoleBinding (CR-0003),
  the secret rows of `status` on podman (CR-0005) and a `$` in a podman environment value
  (CR-0006).
- gdu 10: `podman container inspect <name>` before and after `podman secret create
  --replace`, to read `Config.Secrets[].ID` against the store's ID.
- gdu 6: that quadlet accepts `Network=private` on the oldest supported podman.
- gdu 12: the time one real broker remove takes.
- gdu 17, in solace-util: `kubectl get sts <name>-pubsubplus-{p,b,m} -n <ns> -o
  jsonpath='{.spec.updateStrategy.type}'` under both update strategies.
- Every commit, and the gates after each change.
