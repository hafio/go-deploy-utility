# Surface review: questions for the owner, 2026-11-15

This file holds the detail questions that the surface review of 2026-11-15 left open. Each
entry is one decision. The thirteen rulings you made that day are already decided. They are
recorded in docs/design/2026-11-15-shared-deploy-api.md, section 10. Anything a ruling
settles is not asked again: each part ends with the list "Settled by the rulings".

How to answer: write under the "Answer:" line of an entry. Write "ok" to take the
recommendation. Write anything else in your own words to decide differently. An entry you
leave empty takes the recommendation.

Read these first. In each, two reviewers disagreed or a ruling left a gap, and the
recommendation is a pick made while merging, not a finding: S2, S3, S9, E1, E5, C4, Q1,
Q4, Q10, N25, N32, N35. E1 and E5 also rest on a reading of rulings 11 and 9 (whether
camelCase covers A's `solace:` and `mq:` blocks; whether the env file keeps the word
`image` for run-as), so check those readings.

Words used below:

- M is the shared module go-deploy-utility. A is solmq-conn-util. B is solace-util.
- A renderer is the code in M that writes one artifact (the compose file, the quadlet unit
  or the Kubernetes documents) from a description that the tool hands over.
- Extras are the place in that description where a tool may pass settings that M does not
  model itself (ruling 9).
- A golden is a golden test file: the expected text of an artifact.
- Public means a tool can call the name. Private means only M's own code can.
- The reviewers are the 45 readers of the review. The verifier is the second reading that
  checked a reviewer's finding.
- gdu N is item N of docs/todo.md. CR-NNNN is a finding of the earlier critical review
  (docs/review/2026-11-11-critical-review.md). "Not checked" marks a fact that nobody ran on a
  real engine, host or cluster.

| Part | What it covers | IDs | Entries |
| --- | --- | --- | --- |
| Shared by all three renderers | one decision for all three platforms | S1 to S13 | 13 |
| Env-file keys | keys of the env file in A and B | E1 to E25 | 25 |
| The compose file | the compose file M writes | C1 to C12 | 12 |
| The quadlet unit | the podman unit M writes | Q1 to Q11 | 11 |
| The Kubernetes documents | the Kubernetes documents M writes | K1 to K18 | 18 |
| Exported names | the names M exports, in seven package groups | N1 to N75 | 75 |
| Total | | | 154 |

## Shared by all three renderers

Thirteen decisions that the compose, quadlet and Kubernetes parts each asked on their own.
Each is one decision, so each is asked once here. An entry says what your answer means on
each platform.

### S1. What the extras may not carry

- **About:** Ruling 9 gives each renderer a place for "extras". Extras are settings that M
  does not model itself and that a tool may still pass. An extra passes unless M writes that
  setting itself or it widens privilege. That is a deny list. A setting that is not on the
  list passes, so each list must be complete. `artifacttest` forbids the same settings.
  Today only B uses extras: `shm_size`, `ulimits` and `cpuset` on compose, and `ShmSize=`
  and three `Ulimit=` lines on quadlet. Example: a tool passes `cap_add` and gets `compose:
  extra key "cap_add" is refused because it widens privilege`. The key names are from the
  compose, podman and Kubernetes documentation as the reviewers recalled them. Not checked
  against a real engine or cluster, and no compose or podman file was on this machine.
- **Recommended:** Refuse the settings below on each platform. Write each list into the docs
  with the reason for every key.
  - Compose, three groups. (1) Keys M writes itself: image, container_name, hostname, user,
    restart, entrypoint, command, privileged, cap_drop, security_opt, read_only, mem_limit,
    cpuset (see S2), networks, network_mode, ports, environment, secrets, configs, volumes,
    tmpfs, healthcheck, labels. (2) Keys that reach the host or another container: cap_add,
    devices, device_cgroup_rules, sysctls, pid, ipc, uts, userns_mode, cgroup,
    cgroup_parent, group_add, runtime, gpus, volumes_from, use_api_socket, storage_opt,
    oom_kill_disable, links, external_links. (3) Keys that pull in more files or more
    services: build, extends, env_file, label_file, depends_on, post_start, pre_stop, deploy
    (it can reserve devices). One test names the compose version the list was checked
    against. A key that a later compose version adds stays open until someone adds it here.
  - Quadlet, `[Container]` section. (1) Keys that carry any podman flag: `PodmanArgs=` (it
    can hold `--privileged`, `--cap-add`, `--pid=host`), `GlobalArgs=`,
    `ContainersConfModule=`. (2) Keys that add privilege or reach the host: `AddCapability`,
    `AddDevice`, `Sysctl`, `SecurityLabelDisable`, `SecurityLabelType`,
    `SecurityLabelLevel`, `SecurityLabelFileType`, `SecurityLabelNested`, `SeccompProfile`,
    `Unmask`, `UserNS`, `UIDMap`, `GIDMap`, `SubUIDMap`, `SubGIDMap`, `GroupAdd`, `Mount`,
    `EnvironmentHost`, `EnvironmentFile`, `Pod`, `Rootfs`. (3) Keys M writes itself, refused
    even when this unit does not carry them (extras are written last, and for a one-value
    key the last line wins): `Image`, `ContainerName`, `HostName`, `Label`, `User`, `Group`,
    `NoNewPrivileges`, `DropCapability`, `ReadOnly`, `Network`, `PublishPort`,
    `Environment`, `Secret`, `Volume`, `Timezone`, `Memory`, the five `Health*` keys, and
    the memory and CPU lines of S2. (4) `AutoUpdate=`, which needs your choice: it makes
    podman replace the image on its own, outside the tool and its redeploy question.
    Recommended: refuse it. Still passing: `ShmSize`, `Ulimit`, `Tmpfs`, `Exec`,
    `Entrypoint`, `Notify` and the like. They change the workload, not the privilege. The
    check ignores case, and refuses a key that holds `=`, a space or a line break. The
    `[Service]` and `[Unit]` sections are Q1.
  - Kubernetes, pod. `hostNetwork`, `hostPID`, `hostIPC`, `hostUsers`,
    `shareProcessNamespace`, `securityContext`, `volumes` (so no hostPath),
    `ephemeralContainers` (can carry a privileged container), and three judgement keys:
    `nodeName` (skips the scheduler), `runtimeClassName` (picks another runtime) and
    `priorityClassName` (can push other pods out). The three judgement keys, `volumeDevices`
    and `ephemeralContainers` are my additions from Kubernetes knowledge, not from the
    findings.
  - Kubernetes, container. `securityContext`, `envFrom` (turns every key of a whole Secret
    into an environment variable), `volumeDevices` (raw devices), and `env`, `volumeMounts`
    and `ports` because M writes them (so no `hostPort` and no mount propagation). Every key
    M writes itself is refused too.
  - `artifacttest` matches a key and its value together on Kubernetes (the vendor's CRD text
    names `HostNetwork` in a description). Review each list on every compose, podman and
    Kubernetes release.
  Reason: ruling 9 chose a deny list, and these are the shortest lists with no hole the
  reviewers found. The verifier would have preferred an allow list, which refuses what it
  does not know.
- **Other choice:** Shorter lists. For compose: only group 1 and the first half of group 2
  (privilege and namespaces), so hooks, `deploy` and `build` pass. For Kubernetes: the
  reviewers' list (the first six pod keys and `envFrom`), without the three judgement keys,
  `ephemeralContainers` and `volumeDevices`. Less to keep up, more that a tool can reach.
  For quadlet, an allow list for `[Container]` (today `ShmSize` and `Ulimit` only) is safer
  against keys a later podman adds, but it departs from ruling 9, and a third tool that
  needs `Tmpfs=` waits for a module change.
- **Changes for:** M and `artifacttest` (new tokens). No tool today: the settings B passes
  (`shm_size`, `ulimits`, `ShmSize=` and `Ulimit=`) are not on any list. A third tool gets
  an error naming the key. No tool can add a capability back: `AddCapability` and `cap_add`
  are on the lists.
- **Answer:** Full lists (2026-11-17). Owner: a blacklist that denies everything known to be
  wrong or harmful, as far as possible. Added to the lists above from the compose
  specification, the podman 5 quadlet keys and the Kubernetes pod spec as recalled, not yet
  checked against each reference:
  - Compose: `provider` (hands the service to an outside program), `models`, `develop`
    (syncs host files into the container), `profiles` and `scale` (the service would not
    start as written), `annotations` (runtimes such as crun read `run.oci.*`), `logging` (can
    ship logs off the host), `oom_score_adj`, `cpu_rt_runtime`, `cpu_rt_period`,
    `credential_spec`.
  - Quadlet: `Annotation`, `CgroupsMode`, `LogDriver`, `LogOpt` (`path=` writes on the
    host), `HealthLogDestination` (a host directory). `AutoUpdate` is refused.
  - Kubernetes pod: `containers` and `initContainers` (a second container could be
    privileged), `serviceAccountName`, `serviceAccount`, `automountServiceAccountToken`,
    `schedulerName`, `resourceClaims`, pod-level `resources`, `os`, `restartPolicy`,
    `schedulingGates`, `activeDeadlineSeconds`, `priority`, `tolerations` (can reach
    control-plane nodes), `imagePullSecrets`.
  - Kubernetes container: `lifecycle`, `startupProbe`, `livenessProbe`, `readinessProbe`
    (an `httpGet.host` makes the kubelet call any host), `restartPolicy`, and the keys M
    writes (`name`, `image`, `imagePullPolicy`, `command`, `args`, `resources`).
  - Label and annotation keys a tool passes, on every platform, by prefix: M's own
    (`app.kubernetes.io/managed-by`, `services.solace.com/`, `deploy.solace.com/`),
    `com.docker.compose.`, `io.containers.` (the label `io.containers.autoupdate` turns on
    what `AutoUpdate=` would), `io.podman.`, `run.oci.`, `io.kubernetes.cri-o.`,
    `container.apparmor.security.beta.kubernetes.io/`,
    `seccomp.security.alpha.kubernetes.io/`, and on a Namespace `pod-security.kubernetes.io/`.
  A deny list is complete only for the versions it was checked against. The build checks
  each list against the compose, podman and Kubernetes references and names those versions
  in a test; a key a later version adds passes until it is listed.

### S2. Which settings get their own field

- **About:** B writes a host name, a memory limit, a CPU list, a shared-memory size and
  ulimits. A writes an entrypoint, a command, inline configs, a tmpfs and its own labels. On
  compose B's lines are `hostname`, `mem_limit`, `cpuset`, `shm_size` and `ulimits`. On
  quadlet they are `HostName=`, `PodmanArgs=--memory=8g`, `PodmanArgs=--cpuset-cpus=0-3`,
  `ShmSize=` and `Ulimit=`. A named field gives M one place for the spelling and the check.
  A setting in the extras only has its shape checked (S1). `PodmanArgs=` holds free podman
  flags, so a key list cannot judge it: `PodmanArgs=--privileged` passes a key check, which
  is why S1 refuses it. The memory limit cannot use quadlet's own `Memory=`: it exists only
  from podman 5.5, and quadlet refuses a key it does not know. The CPU list has no quadlet
  key at all. Design 8.2 already lists resources among the values the tool supplies.
- **Recommended:** Own fields, written in each platform's spelling by M, for `hostname`, the
  memory limit and the CPU list. On compose also for A's entrypoint, command, configs, tmpfs
  and tool labels (generic compose pieces). The SELinux letter (`:Z`) is a field of the
  mount, because the quadlet `Volume=` line takes the same option. Shared-memory size and
  ulimits go through the extras: only B uses them, and quadlet spells ulimits another way.
  `PodmanArgs` stays refused. On Kubernetes this answer changes nothing: memory and CPU are
  already the Deployment's own resources (E12). My assumption: a CPU list or a host name set
  for a Kubernetes target is refused with an error naming the field, because the Deployment
  has no such field.
- **Other choice:** More own fields, for example a typed `ulimits`. It is checked more, but
  it serves one platform, and a tool needing a new limit waits for a module release. Or
  everything through the extras and `PodmanArgs` allowed, with the value scanned against
  podman's flag list: a flag the list does not name slips through.
- **Changes for:** The module's types only. B's files keep their bytes for these lines,
  apart from the quotes (`shm_size: "2g"`, `mem_limit: "6898m"`). B's rendering code for the
  memory and CPU lines.
- **Answer:** More fields, typed ulimits (2026-11-17), after the owner asked for an assessment.
  Own fields for the host name, the memory limit, the CPU list, the shared-memory size and
  ulimits, and on compose for A's entrypoint, command, configs, tmpfs and tool labels; the
  SELinux letter is a field of the mount. Ulimits take the 16 Linux limit names, soft not
  above hard, a number or -1 for unlimited; `rtprio`, `rttime` and `nice` are refused (they
  let a container out-prioritise the host). Linux has a fixed set of limit names, so the
  typed field never waits for a module release. From the one field M writes compose
  `ulimits:`, quadlet `Ulimit=` in `[Container]` and the matching `Limit*=` in `[Service]`
  (-1 becomes `infinity`), which settles Q1. A host name, CPU list, shared-memory size or
  ulimit for a Kubernetes target is refused with an error naming the field. B passes no
  extras afterwards. `PodmanArgs` stays refused.

### S3. A stated root user against a forgotten one

- **About:** Ruling 9: run-as is a numeric uid and gid, and an empty field means the image's
  own user. The risk is uid 0 with gid 0, because that is also what two unset numbers look
  like. M cannot tell "root, stated" from "forgotten". A accepts `runAs: 0` and `0:0` today
  and only warns. On Kubernetes it writes these lines:

      securityContext:
        runAsUser: 0
        runAsGroup: 0
        fsGroup: 0

  Empty writes only the `seccompProfile` part. A non-zero uid also writes
  `runAsNonRoot: true`. On compose the same choice is `user: "0:0"` against no `user:` line,
  and on quadlet `User=0` and `Group=0` against no line.
- **Recommended:** Make "stated" differ from "empty" in the field (a set flag, or numbers
  that can be absent). A stated `0:0` still writes the root lines, without `runAsNonRoot`.
  Empty still writes nothing and means the image's own user. A's `0:0` keeps working, and M
  never writes root for a field nobody filled. The reviewers' type could not tell the two
  apart. Same on all three platforms.
- **Other choice:** Refuse uid 0 outright. The strictest reading of the container hardening
  rule, but A's `runAs: 0` and `0:0` stop rendering, and an env file that works today fails.
- **Other choice:** Read `0:0` as empty. A silent change: for an image that declares a USER,
  an env file that says root would run as that user.
- **Changes for:** A's env files only under "refuse". Otherwise the golden files only. The
  env-file spelling is E5.
- **Answer:** ok (2026-11-17): stated differs from empty.

### S4. The `solution` and `developed-by` labels

- **About:** Every artifact carries five ownership labels (gdu 1). Two of them hold Solace
  Professional Services values: `services.solace.com/solution: "accelerator"` and
  `services.solace.com/developed-by: "apac-psg"`. Both tools pass them. A tool from outside
  that team should not carry them. The reviewers' model rows disagree: one says both are
  required and non-empty, another says an empty one is skipped. Design note 9.2 says these
  values suit only Professional Services tools.
- **Recommended:** Both are fields of the shared identity that the tool passes. An empty one
  leaves that label out. `product` and `instance` stay required, because removal selects on
  them. A and B pass `accelerator` and `apac-psg`, so nothing changes for them. The answer
  is the same on compose (service labels), quadlet (`Label=` lines) and Kubernetes (object
  labels).
- **Other choice:** M writes the two values always, as gdu 1 reads. The labels never vary,
  but a tool outside the team carries labels that are not true of it.
- **Changes for:** Nobody now. A third tool chooses.
- **Answer:** ok (2026-11-17).

### S5. Quotes on keys that are data

- **About:** Ruling 9 quotes every text value. It says nothing about keys. Some keys are
  data, not keywords: environment names and label keys on compose, and the keys of pod
  labels, annotations, Secret and ConfigMap file names and node selectors on Kubernetes. A
  key such as `123`, `true`, `NO` or `1e3` passes the name checks, but a YAML reader takes
  it for a number or a boolean (from the YAML rules, not run on an engine). Compose today:
  `TZ: "Asia/Singapore"` (bare). Quoted: `"TZ": "Asia/Singapore"`. Kubernetes today:
  `app: web`; quoted: `"app": "web"`. The service, secret and config names on compose are
  quoted already.
- **Recommended:** Quote every key that is data, M's own label keys included, so one map
  reads one way. A key that is a keyword of the file (`image`, `restart`, `apiVersion`,
  `spec`) stays bare. On compose, also refuse `$`, spaces and control characters in data
  keys. On Kubernetes the tool tells M which of its maps hold field names and which hold
  data, and a bare field name must be plain letters and digits and not true, false, yes, no,
  on, off, y, n or null. This closes the hole that the value rule closes. Quadlet writes no
  data keys: `Environment="TZ=Europe/Paris"` already quotes the whole assignment (Q4).

      metadata:
        name: "broker-credentials"
        labels:
          "app.kubernetes.io/managed-by": "solmq-conn-util"
          "deploy.solace.com/instance": "broker"

- **Other choice:** Leave data keys bare under a strict character rule (compose), or quote
  only the keys a tool supplies and leave M's own bare (Kubernetes). A map then mixes both,
  for example `app.kubernetes.io/managed-by: "x"` beside `"app": "web"`: two rules, and a
  name like `NO` or `ON` can be read as a boolean by an old reader. A third way is to quote
  a key only when it would be misread, but that needs a complete list of YAML 1.1 cases
  inside M, the list the reviewers refused to keep for values.
- **Changes for:** The golden files of both tools, on every environment line and label line
  (compose) and wherever a labels, annotations or data map appears (Kubernetes). Nothing for
  env files or for running containers and clusters: the keys parse to the same text.
- **Answer:** ok (2026-11-17).

### S6. Health check times and retries

- **About:** A writes constants (`30s`). B passes a Go duration from its env file as given,
  so `1m30s`, `500ms` or `1.5s` is possible. The reviewers proposed whole seconds and
  refusing the rest, which refuses `500ms` (it loads today). The platforms differ:
  Kubernetes probes take whole seconds only, compose takes `"90s"`, and the quadlet line is
  `HealthInterval=90s`. B also has no check on retries: a negative number reaches the unit
  today (`HealthRetries=-4`).
- **Recommended:** Whole seconds on all three platforms, written `<n>s`. `1m30s` still loads
  and is written `90s`; 60s never turns into `1m0s`. Compose writes it quoted
  (`interval: "90s"`, `start_period: "60s"`). Interval and timeout above zero, start period
  zero or more, retries 1 or more (the retries rule is new, not B's, and fixes B's
  negative-retries defect). B refuses a fraction of a second at load and says so, or its
  render would fail after the questions. Reason: one spelling works everywhere, and a
  half-second probe has no use.
- **Other choice:** Allow milliseconds (`"1500ms"`) or write the Go duration text as given.
  More freedom and B's env file stays as it is, but Kubernetes still needs whole seconds, so
  the shared description would hold two rules.
- **Changes for:** B env files with a fractional time (none known; unlikely) and with a
  negative or zero retries (a bug today). Golden files: compose times gain quotes. A writes
  integers and does not change.
- **Answer:** ok (2026-11-17).

### S7. An image with no tag and no digest

- **About:** The proposed image grammar makes the tag and the digest both optional, so
  `registry.example.com/solace-pubsub-standard` renders and pulls `latest`. Both tools
  refuse an untagged image at load, and A says an untagged image "pins nothing". The same
  image string goes into compose `image:`, quadlet `Image=` and the Kubernetes container.
- **Recommended:** M requires a tag or a digest on every platform. The word `latest` as an
  explicit tag stays allowed, because neither tool refuses it today. The pin rule then
  travels with the grammar to a third tool. This also closes A's `%` hole in the unit's
  `Image=` line. B still cannot name a digest unless its own config gains one.
- **Other choice:** Leave the pin rule to each tool. A third tool can then render an
  unpinned image without noticing.
- **Changes for:** Nobody now. A third tool.
- **Answer:** ok (2026-11-17).

### S8. Error text must not repeat a value

- **About:** Every render error names what failed. Echoing the value helps, but A's
  ConfigMap holds the status password inside `application.yml`, and Java options and other
  environment values are operator text. A sentence such as `env "X" value "..." is refused`
  could print a password into a log. The same risk exists for compose inline configs and for
  quadlet `Environment=`, `Label=` and extras lines, which are free text.
- **Recommended:** Echo a value only for names, keys, words from a fixed list and numbers.
  For every other field the error names the field, the key and the rule, and never the
  value: Secret data, registry user and password, ConfigMap data, inline config text (name
  and an offset only), environment values, command and arguments, label and annotation
  values, the image reference, extras, and every scalar of a custom body. This is ruling 1's
  rule for client errors, applied to rendering. Same on all three platforms.

      error: Deployment "broker": container "app": env "JAVA_TOOL_OPTIONS": the value
      holds a control character; the value is not shown

- **Other choice:** Echo everything except Secret data and the registry login. Kinder
  messages, but A's ConfigMap password and any operator text reach the log.
- **Changes for:** Whoever reads a render error. Tools still check at load with friendlier
  text.
- **Answer:** ok (2026-11-17).

### S9. Mode and owner of a secret file

- **About:** A secret reaches the container as a file. Compose and podman use the engine's
  default: for podman, mode 0444 and owned by root, so any process in the container can read
  it. `mode=0400,uid=1000` is possible on podman, but compose does not apply it on a plain
  docker engine. Kubernetes has `defaultMode`: A writes `0400`, and the model allows any
  mode up to `0777`, so a tool could make secret files readable by every process in the pod.
  A uses `fsGroup` so its uid can read `0400` files. Neither tool sets a mode, uid or gid on
  anything but the default today. Not checked on a real engine.
- **Recommended:** No mode, uid or gid on compose and quadlet. Both tools rely on the
  default, a stricter value would differ from docker, and the container already runs as a
  non-root user with every capability dropped. On Kubernetes M writes `defaultMode: 0400`
  and refuses any mode with a bit for "other" (the last digit not 0). `0440` stays possible.
  Reason: least privilege, the lowest mode that does the job.
- **Other choice:** An optional mode and owner per secret on every platform (tighter on
  podman only), and on Kubernetes allow up to `0777` and leave it to the tool. Less code in
  M, more freedom to make a mistake.
- **Changes for:** Nobody: A and B use the defaults.
- **Answer:** ok (2026-11-17): engine default on compose and quadlet; Kubernetes
  `defaultMode: 0400`, no bit for "other".

### S10. Check at load as well as at render

- **About:** Each renderer refuses a bad value when it renders, and names the field, the
  value (S8 permitting) and the rule. A tool also checks its env file at load, so the
  operator hears about a bad value before any question is asked. Both tools hold their own
  copy of those load checks today (A `validate.go`, B `artifactvalues.go`): the image
  reference, a published port, a health duration, the restart word, the network word.
  Several of M's rules are narrower than the tools' load checks. Example:
  `ValidPort("8080:80")` in the tool's load check, and the same function inside `Render`.
  The word lists of gdu 6 (docker: bridge, host, none; podman: private, bridge, pasta,
  slirp4netns, host, none) are the same kind of check.
- **Recommended:** Both. `Render` refuses on its own. M also exports one check per rule
  beside the renderers (project name, port entry, restart word, network word, image
  reference, variable name), and the renderers call the same ones. For the compose file
  there is also a `File.Validate()` that the tool can call at load with its own field names.
  My assumption: the quadlet and Kubernetes renderers get a matching check; only the compose
  part asked. The tool keeps its field names and its wording. Reason: a value that passes
  load and fails in `Render` gives the operator a worse message and names no env key, two
  copies drift, and it keeps "questions first, actions after" true when M is stricter than
  the tool.
- **Other choice:** `Render` only. Less to export, but each tool keeps hand-written copies
  of the checks, must build the whole description and call `Render` at load just to learn
  that a port is wrong, and an error can surface after the operator answered a question.
- **Changes for:** Tool code at load: both tools delete their own regular expressions and
  call M. Operators see the same refusals, in M's words or the tool's. Nothing for env files
  or running deployments.
- **Answer:** Both, as a core set (2026-11-17). Owner: M has a core set, tools extend it if
  required. M exports one check per rule it renders by, and the renderers call the same
  ones. A tool calls them at load and may add checks of its own on top; it keeps no copy of
  a core check.

### S11. How `generate` writes its output file

- **About:** `generate` output can hold secrets: A's compose file carries the status
  account's password inside `application.yml`, and the Kubernetes output carries Secrets. A
  writes `-o` output with a plain `os.WriteFile` at mode 0600. The mode applies only when
  the file is created, a symbolic link at the path is followed (so the file lands somewhere
  the operator did not choose), and an existing file is overwritten without a question. B
  refuses a link, asks before overwriting, writes a temp file at 0600 and renames it. gdu 7
  gives M one atomic write, but names quadlet units only.
- **Recommended:** Both tools write every `generate` output through M's atomic write: it
  refuses a path that is a symbolic link, creates the temp file at 0600 before the first
  byte, then renames (a rename replaces a link instead of writing through it). The question
  "overwrite an existing file?" stays in each tool. This adds compose and Kubernetes
  `generate` output as users of the helper, and the same helper writes the installed quadlet
  unit (Q5). A gains the refusal and B cannot lose it.
- **Other choice:** Leave both writers as they are. Nothing moves, but A's existing file
  keeps whatever mode it had, A stays open to the link problem, and B must remember to keep
  its own check.
- **Changes for:** A's `generate -o`, on compose and Kubernetes: an operator who sends it to
  a path that is a link now gets an error. Nothing for B.
- **Answer:** Owner (2026-11-17): M's write behaves like the shell for the mode: no chmod,
  the mode comes from the umask (0666 less the umask, 0644 under the usual 022), for every
  output, secrets included. It refuses a path that is a symbolic link and writes a temp
  file then renames it, so a half-written file never appears. The "overwrite?" question
  stays in each tool. Build note: the temp file is opened with mode 0666 so the kernel
  applies the umask (`os.CreateTemp` would give 0600). Accepted cost: under umask 022,
  output that holds a Secret or password is readable by every user of the host, where both
  tools create it 0600 today. The installed unit's mode is Q5.

### S12. What a host name may contain

- **About:** B writes `HostName=<name>` in the unit (needs podman 4.6) and `hostname:` in
  the compose file from its node name. The same string is the broker's router name and part
  of keys like `redundancy_group_node_<name>_connectvia`. B holds it to letters, digits,
  `.`, `_` and `-`, with no length limit. A name like `solace_primary`, or one of 70
  characters, loads today and would fail an RFC 1123 host name check. In standalone mode B
  fills the name from the operating system's host name after its checks have run and never
  checks it, so a bad OS host name reaches `Description=` and `HostName=` raw. Not checked:
  whether podman accepts an underscore host name. I could not tell whether the Solace
  broker's router-name rules, or B's existing env files, use underscores, so the cost to B
  users is unknown.
- **Recommended:** M checks an RFC 1123 host name: letters, digits and `-`, dots allowed,
  each label up to 63 characters, the whole name up to 253, starting with a letter or digit.
  B holds its three node names and the OS host name to the same rule. Reason: it is one
  string used in three places, and Kubernetes names and ruling 11's container name are
  already DNS labels.
- **Other choice:** M checks only what makes the line safe (no whitespace, control character
  or text that is unsafe with `%`). B's env files keep working as they do. A bad name then
  fails at the engine, not at generate.
- **Changes for:** B env files with an underscore or a long node name. A is not affected (it
  writes no host name).
- **Answer:** ok (2026-11-17).

### S13. What a time zone may contain

- **About:** The reviewers proposed letters, digits, `_`, `+`, `-` and `/` (`Europe/Paris`).
  Both tools admit more today: `.`, `,`, `:` and `%` pass, so a POSIX rule such as
  `EST5EDT,M3.2.0,M11.1.0` loads. Ruling 9 makes the time zone a field of the shared
  description. On compose and quadlet it is the `TZ` setting (quadlet:
  `Environment="TZ=..."`, which is escaped), and on Kubernetes it is the environment entry
  `TZ`. The narrow set is not needed for safety.
- **Recommended:** Keep the tools' rule on every platform: no whitespace, control or
  invisible character, and none of the shell characters. A narrower set would refuse values
  that load today, for no gain.
- **Other choice:** The narrow set. A real zone-name check, but a POSIX rule string is
  refused.
- **Changes for:** Nobody with the wider set. With the narrow one, any env file that uses a
  POSIX rule string.
- **Answer:** The narrow set, aligned to the standard (2026-11-17). Owner: allow no more than
  required. The value is an IANA time zone database name: parts separated by `/`, each of
  letters, digits, `_`, `+` and `-`, at most 14 characters, not starting with `-` (the
  database's own naming rule, as recalled, not checked). POSIX rule strings such as
  `EST5EDT,M3.2.0,M11.1.0` are refused. Changes for: env files that use one (none known).

#### Settled by the rulings

- Every text value a tool supplies is written quoted, on all three platforms: ruling 9.
  Booleans are bare, and so are the numbers a platform requires as numbers (Kubernetes
  refuses a quoted integer). `restart: "no"` and `target: "/mnt/secrets/..."` change. Left
  open: quotes on keys (S5), and for quadlet the keys where a quote would stay in the value
  (Q4).
- The root filesystem has no default in the shared description: ruling 9. The tool chooses.
  Read-only writes `read_only: true`, `ReadOnly=true` or `readOnlyRootFilesystem: true`.
  Writable writes nothing, so gdu 6's decision holds. The engine question for compose is C3.
- A mount never becomes writable by a forgotten field: ruling 9. The mount field is
  inverted, so its zero value writes `:ro`. B sets writable on its data bind
  `"/opt/solace/data:/var/lib/solace:Z"`. A's binds keep `:ro`. How the field is shaped (a
  flag inverted so that unset writes `:ro`, or a required word that errors when unset) is a
  detail for the build.
- Ports together with host or none networking: ruling 9. The renderer leaves the ports out
  and its doc says so. No error is added, so a third tool sees the doc only. B keeps
  checking its list in the env file as today.
- Restart: ruling 9, an empty field writes `always`. `no`, `unless-stopped` and
  `on-failure[:N]` stay allowed, and are quoted on compose. Quadlet writes `Restart=always`,
  and A's drop of the `[Service]` section on an empty restart ends.
- Run-as is a numeric uid and gid, and empty means the image's own user: ruling 9. The
  reviewers' wider grammar with names and their `UserFromImage` flag go. What a stated `0:0`
  does is S3. The env-file spelling is E5.
- The time zone is a field of the shared description: ruling 9. Compose and quadlet write
  `TZ` first, as both tools do today, so the bytes stay. On Kubernetes it is the environment
  entry `TZ`, written first (A's order). A tool that also passes `TZ` as an environment
  entry is refused as a duplicate. The character set is S13.
- One set of types for the container description (identity and labels, run-as, mount, port,
  environment entry, health check, timezone): ruling 9. Each renderer checks its own
  platform's rules. The secret type stays per platform, because compose also needs the
  variable it reads from.
- Extras take a deny list, not an allow list: ruling 9, over the verifier's allow list. The
  list must be complete for each platform, and `artifacttest` forbids the same keys. The
  content is S1; the `[Service]` and `[Unit]` sections of a unit are Q1.
- Adding `Exec=` or a capability back for a third tool: ruling 9. `AddCapability` and
  `cap_add` are forbidden, so there is no opt-out. `Exec=` is not a widening key and passes
  the extras.
- The lib annotation and the version: ruling 10. Kubernetes only: a variable that each
  tool's `build` task stamps with the version in its `go.mod`, `devel` when unstamped, so
  goldens stay stable. A compose file and a unit carry no version, so a module upgrade alone
  changes no byte and recreates or restarts nothing.
- The container name is a DNS label of at most 63 characters in both tools: rulings 11 and
  13. B's names of up to 100 characters go. M still checks each platform's own wider grammar
  for the service name and the instance label.
- Kubernetes gets no `hostNetwork` line, and compose states its project network: ruling 13.
  The network's labels are C4.

#### No decision needed

- The hardening lines (`privileged: false`, `cap_drop: ALL`, `no-new-privileges=true` and
  their quadlet and Kubernetes forms) stay a constant with no switch. `artifacttest` keeps
  pinning them.
- An optional field at its zero value writes nothing. A required field left empty is an
  error, and so is an empty list entry. No line ends at `=`. An empty block is left out.
- `Render` returns the bytes and an error that names the field and the rule, and the value
  only where S8 allows it.
- Every free-text value is held with `execguard.IsControl`, `execguard.IsInvisible` and a
  UTF-8 check at the renderer. This does not replace a tool's own load check. `IsControl`
  misses U+0085 today, which is a separate fix in M (N13).
- A duplicate environment name is an error, since quadlet keeps only the last one.

## Env-file keys

The keys of the env file under docker, podman and kubernetes in A and B, and the keys around
them. Key names are written in camelCase, as ruling 11 requires.

### E1. How far the camelCase rename reaches in A

- **About:** Ruling 11 says camelCase in both tools. A has two kinds of keys. Its own
  sections use kebab-case (`docker.project-name`, `kubernetes.deployment.node-selector`).
  Its `solace:` and `mq:` blocks, in the env file and in every workflow file, mix the tool's
  keys with the connector's own properties: `msg-vpn: x` sits beside
  `user-authentication-mqcsp: false`. The ruling does not say whether the second kind is
  renamed too.
- **Recommended:** Rename every key outside the `solace:` and `mq:` blocks (all platform
  sections, `image`, `tls`, `javaOptions`, `leaderElection`, `mqDefaults`,
  `workflows.filePattern`). Leave the keys inside `solace:` and `mq:` blocks as they are.
  The connector's own properties there are not ours to rename, and one block would otherwise
  hold two spellings. A already refuses a second spelling of a key it handles in those
  blocks. The exact key list is fixed when the rename is built.
- **Other choice:** Rename those keys too (`msgVpn`, `queueManager`, `clientUsername`). One
  spelling everywhere in A, but every workflow file changes, and a block then mixes `msgVpn`
  with the connector's kebab-case neighbours.
- **Changes for:** A's operators: every env file. With the other choice also every workflow
  file. B: nobody.
- **Answer:** Owner (2026-11-17): rename only the keys that are A's own; anything meant to be
  passed on to the application stays as written, with no conversion. Inside `solace:` and
  `mq:` the connector's property names stay (`host`, `msg-vpn`, `client-username`,
  `client-password`, `api-properties`, `queue-manager`, `channel`, `conn-name`, `user`,
  `password`, `additional-properties`, and every pass-through key). A's own keys there are
  renamed whole: `conn-ref` to `connRef`, `client-username-env` to `clientUsernameEnv`,
  `client-password-env` to `clientPasswordEnv`, `user-env` to `userEnv`, `password-env` to
  `passwordEnv`, `key-alias` to `keyAlias`. Outside the blocks the same rule applies to
  every key. The exact list is fixed when the rename is built (A's `rawSolace` and `rawMQ`
  in `internal/spec/spec.go`, and every key outside them).

### E2. Detectors for the old key names

- **About:** After the rename an old file still says `image.repo` or `runUser`. Each tool
  could keep a check that says "renamed, use image.repository". Both tools have such checks
  today for keys removed earlier (A refuses `podman.mode`, B refuses `docker.composeFile`).
  The renames ship inside the breaking tag (ruling 8).
- **Recommended:** No new detectors. The breaking tag drops the old ones, and each tool's
  configuration.md gets one table "old key -> new key". Strict decoding (an unknown key is
  an error) reports the old key as unknown. This holds only if A decodes strictly first
  (CR-0046, decided, not built): today A drops unknown keys under docker, podman and
  kubernetes without a word, so an old `image.repo` would vanish and the image would pull
  from Docker Hub. A's strict decode must land in or before the breaking tag.
- **Other choice:** Keep one by-name refusal per old key, so every old file gets a message
  that names the new key. It costs about 75 detectors in B and 30 in A, each with a test, to
  delete again later.
- **Changes for:** Operators with existing files: a plain "unknown key" error and the table
  in the docs. Tests: the old detectors and their tests go.
- **Answer:** Owner (2026-11-17): no detectors, new or old. An old key is reported as an
  unknown key like any other, with no "moved" or "renamed" wording, and the existing by-name
  refusals go. Each tool's configuration.md keeps the "old key -> new key" table. This
  needs strict decoding in both tools: B has it (`KnownFields(true)`, B
  `internal/config/load.go:31`), A does not yet (A decodes without `KnownFields`, see the
  comment at A `internal/spec/kubernetes.go:166`), so A's strict decode (CR-0046) lands in
  or before the breaking tag.

### E3. The image keys

- **About:** `image.repo` means the registry host in A and the repository path in B. Example
  for both tools: `registry: ghcr.io`, `repository: solace/pubsubplus-standard`,
  `tag: 10.26.0`. Today A calls the path `image.name`; B calls the host `image.registry`.
- **Recommended:** `image.registry` (host and optional port only; empty means Docker Hub),
  `image.repository` (path only, never a host), `image.tag`. `repo` goes from both files, so
  an old file fails instead of being read with the other meaning. The verifier added the
  rule that the repository holds no host: both tools accept one there today, and then the
  pull Secret's key is wrong (A writes Docker Hub's key, B writes an empty one). With empty
  registry the key is `https://index.docker.io/v1/`. A repository whose first part has a dot
  or a colon is refused. A digest in `image.tag` stays A only.
- **Other choice:** Keep A's `image.name` for the path. A renames one key instead of two,
  but B renames `repo` to `name`, and `repository` is the word of the OCI spec and of B's
  custom resource.
- **Changes for:** Every env file in both tools. B's registry with a path (`ghcr.io/solace`)
  must split into registry and repository. B's pull-Secret golden changes when the registry
  is empty (empty key becomes Docker Hub's).
- **Answer:** ok (2026-11-17): `image.registry`, `image.repository`, `image.tag` in both.
  Revised by the owner (2026-11-18): the path key is `image.name`, not `image.repository`, to
  prevent confusion; `image.registry` and `image.tag` stay. The same block serves every image
  key (E23). M's Go type keeps the OCI word (`workload.Image.Repository`).

### E4. The network key

- **About:** Todo 6 gives each platform its own word. Today only B has a key, written
  `docker.network.mode: host`, and it fills in `bridge` when the key is absent. A has none.
- **Recommended:** A scalar directly in the section: `docker.network: host`,
  `podman.network: pasta`. Docker takes bridge, host, none. Podman takes private, bridge,
  pasta, slirp4netns, host, none. No default is filled into the config. Absent writes the
  platform's non-host default (compose project network, `Network=private`). B's old
  `network.mode` stops decoding, and its filled-in `bridge` goes. The word `bridge` now
  writes `network_mode: bridge` or `Network=bridge`, which is not the default. Kubernetes
  has no key. Not checked on a real host: `Network=private` on the oldest supported podman
  (todo 6).
- **Other choice:** Keep B's block (`network.mode`, `network.ports`) and move A's ports
  under it. No change for B's layout, but ruling 11 flattens the layout, and an old B file
  with `mode: bridge` would still load and mean something else.
- **Changes for:** B env files (`network.mode: host` becomes `network: host`). Golden files
  of both tools gain the explicit network line.
- **Answer:** Owner (2026-11-18), after the flatten review: keep B's block, `network.mode` and
  `network.ports`, in both tools; A's ports move under it. The words stay todo 6's. What an
  absent `mode` and a written `bridge` mean is C2.

### E5. Run-as in the env file: the key, the word `image`, and a bare uid

- **About:** Ruling 9: the shared description holds a numeric uid and gid, and an empty
  field means the image's own user. The env-file spelling is not decided. A writes
  `runAs: "1000:0"`, keeps the word `image` for the image's own user, and reads a bare
  `1000` as `1000:0`. B writes `container.runUser: "1000001:0"` (rootless podman: `1000:0`)
  and accepts names (`solace:solace`). A bare `1000` in B writes `User=1000` with no
  `Group=` line, and ruling 9 does not let M write a user without a group. B also chowns its
  data directory to that user before deploy (`podman unshare chown`, or `chown`), so it
  needs real numbers, and a changed value changes the directory's group too. I read "empty"
  as the renderer's field, not as the env-file spelling.
- **Recommended:** `<platform>.runAs` in both tools, as `uid:gid`, numbers only. Names stop
  loading. A keeps the word `image` in the env file and hands M an empty field. A keeps
  reading a bare uid as `uid:0`. B refuses `image` with a message that says why (it must
  chown the directory), and refuses a bare uid with "write uid:gid, for example 1000:0". B
  keeps its defaults (`1000:0` rootless, `1000001:0` rootful), which already have both
  numbers. Reason: nothing in B is silently changed, and a name could never have worked with
  the host-side `chown`.
- **Other choice:** A bare uid means gid 0 in both tools (A's rule). The same in both, but a
  B file with `1000` would then state group 0, where today podman picks the group from the
  image's entry for that user (my recollection of podman, not checked against the broker
  image), and the data directory's group changes. Or: the env file also takes an empty
  string (`runAs: ""`) and A drops the word `image`. One spelling, but an empty value then
  cannot be told from "use the default".
- **Changes for:** B env files: `runUser` becomes `runAs`; a name like `solace` and a bare
  uid no longer load. A: only the rename. B's tests that use a name or a bare uid change.
- **Answer:** Owner (2026-11-17), after the clash with A's image (no USER, so blank is root)
  and B's chown was raised: absent means the tool's default, a stated non-zero uid with its
  gid (A `1000:0`; B `1000:0` on rootless podman, `1000001:0` on docker and rootful podman).
  The word `default` asks for the image's own user: M gets an empty field and writes no
  user line. Otherwise `uid:gid`, numbers only; names and a bare uid are refused ("write
  uid:gid, for example 1000:0"). A's word `image` becomes `default`. B with `default` reads
  the image's user (image inspect, ruling 6) for its data-directory chown and refuses when
  that user is not numeric. Samples and examples state a uid:gid. The key's place follows
  the flatten review (E4).
  Follow-up (2026-11-18): the word is `default` in both tools and on every platform,
  Kubernetes included; it tells M to write no identity. Unset in env.yaml means the tool's
  default: A `1000:0`. B on Kubernetes states the operator's own defaults when unset (user
  1000001, group and fsGroup 1000002) and refuses a stated 0, which the CRD reads as "the
  operator's default" (owner, 2026-11-18).

### E6. Restart keys for B, and the two defaults

- **About:** B writes `Restart=always` and `restart: always` with no key. A has
  `docker.restart` (default unless-stopped) and `podman.restart` (default always). Ruling 9:
  an empty field in the shared description writes `always`.
- **Recommended:** B gains `docker.restart` and `podman.restart`, default `always`. The word
  lists (docker: no, always, unless-stopped, on-failure[:N]; podman: systemd's words) move
  into M beside the network words. A's docker default stays `unless-stopped`; the difference
  goes into the parity body and is not aligned, because `always` also restarts a container
  you stopped by hand once the daemon restarts.
- **Other choice:** Align A's docker default to `always`. One default everywhere, but
  running connectors on docker come back after a daemon restart even when stopped on
  purpose.
- **Changes for:** B: two new optional keys, nothing changes until used. A: nobody.
- **Answer:** Owner (2026-11-17): B gains `docker.restart` and `podman.restart`, and A's docker
  default is aligned to `always` as well, so the default is `always` everywhere and the word
  lists move into M. No parity difference is recorded. Accepted cost: a connector stopped by
  hand on docker comes back when the docker daemon restarts.

### E7. Port keys and the port grammar

- **About:** `docker.ports: ["8080:8080"]`. A takes `8080` (meaning 8080:8080) or
  `"host:container"`. B takes `[ip:]host:container[/tcp|/udp]`, with ranges, under
  `network.ports`.
- **Recommended:** `<platform>.ports` directly in the section, one grammar in M: B's. Both
  sides are always written. A's bare integer is refused with a message that says to write
  `8080:8080`. The engines read a bare `-p 8080` as a container port on a random host port,
  so A's meaning contradicts theirs (stated from knowledge, not checked on an engine).
- **Other choice:** Keep A's bare integer meaning N:N. Shorter to write, but M holds two
  grammars and the same text means two things in the tool and in the engine.
- **Changes for:** A env files that use a bare integer (load error). B: the key moves up
  from `network.ports`. A gains bind address, protocol and ranges.
- **Answer:** Owner (2026-11-17): a bare port means the same port on both sides, in both
  tools, handled in M. A bare port or range, with an optional `/tcp` or `/udp`, is read as
  `N:N`; M always writes both sides into the artifact, so nothing reaches the engine bare and
  the engine's own reading of a bare `-p 8080` (a random host port) never applies. Otherwise
  B's grammar: `[ip:]host:container[/tcp|/udp]`, ranges allowed; with a bind address both
  sides are written. A keeps the bare form; B gains it. The key's place follows the flatten
  review (E4).
  Placement (2026-11-18): `<platform>.network.ports` (E4).

### E8. Naming a Secret the tool builds or one someone else made

- **About:** On Kubernetes both tools can build a Secret or point at an existing one. A says
  it in words: `credentials: {create: true}` or `{existing: NAME}`, and
  `image-pull: {name: NAME}`. In B one key, `kubernetes.adminSecret: NAME`, means "build and
  own" when `semp.adminPass` is set and "reference" when it is not. `imagePullSecret` works
  the same way.
- **Recommended:** One block in both tools, `kubernetes.secrets.credentials` and
  `kubernetes.secrets.imagePull`, each with `create: true` or `existing: NAME`. A created
  Secret always has the derived name. An existing one is never written or deleted. A renames
  `imagePull.name` to `existing`. B drops `adminSecret` and `imagePullSecret`, and with them
  a custom name for a Secret the tool builds.
- **Other choice:** Keep B's rule: the name key plus a password key decide. No new block,
  but a missing password silently turns "build" into "reference".
- **Changes for:** B env files that name these Secrets, and B's tests. A: one key renamed.
- **Answer:** ok (2026-11-17).

### E9. The TLS Secret in B

- **About:** B today: `kubernetes.tlsServerSecret: NAME`, built from `tls.cert` and
  `tls.certKey` when they are set, only referenced when they are not. The reviewers proposed
  `kubernetes.secrets.tls.{create|existing}`. The verifier found that A has no such key and
  retired exactly this shape (`secrets.stores`). A derives its own Secret with no switch
  whenever it holds the material, and names a Secret someone else made on the TLS material
  itself (`tls.stores.<name>.existing.secret`).
- **Recommended:** Follow A. B gets an `existing` Secret reference inside its `tls` block
  (for example `tls.existing.secret: my-tls`) and no `create` key. The files in `tls.cert`
  and `tls.certKey` are the instruction to build `<name>-tls`. The exact spelling is fixed
  at build.
- **Other choice:** The reviewers' block, `kubernetes.secrets.tls`, so B's three Secrets
  share one switch. Tidier inside B, but it is a shape A does not have and had retired.
- **Changes for:** B env files that set `tlsServerSecret`. A: nobody.
- **Answer:** Owner (2026-11-17): every Secret follows E8, with `create` and `existing` keys and
  no hidden logic. B: `kubernetes.secrets.tls` with `create: true` (built as `<name>-tls`
  from `tls.cert` and `tls.certKey`, which are then required) or `existing: NAME`. Cert
  files with no block are an error that says to add `create: true`, as in E11. How far the
  rule reaches was asked next: A's TLS stores Secret and B's additional-users Secret get
  `create` and `existing` too (`kubernetes.secrets.stores`,
  `kubernetes.secrets.additionalUsers`; exact spelling at build); B's operator `regcred`
  stays as E10 says.
  Revised 2026-11-27 (also E8, E11): no `create: true` switch for any Secret. Material set
  means built under the derived name, `existing: NAME` means referenced, both is an error,
  neither builds nothing (B's admin: the operator generates one). See
  [2026-11-18-secrets-assessment.md](2026-11-18-secrets-assessment.md) section 6.

### E10. B's operator pull Secret when the broker's is "existing"

- **About:** B builds a second Secret, `regcred`, in the operator's namespace from
  `image.user` and `image.pass` whenever both are set, and removes it by name on operator
  delete. No key covers it. With `imagePull: {existing: NAME}` (E8) the broker's namespace
  is settled but the operator's is not.
- **Recommended:** `existing` covers the broker's namespace only. The operator install keeps
  building `regcred` whenever both `image.user` and `image.pass` are set, and builds nothing
  when they are not. The doc says so. The tool owns `regcred` (a fixed name), so deleting it
  on operator remove does not touch a Secret someone else made.
- **Other choice:** A key for the operator namespace as well (for example
  `kubernetes.operator.imagePullSecret`). Full control for sites that never put credentials
  in the file, at the cost of one more key.
- **Changes for:** B operators of sites that name an existing pull Secret. A: nobody.
- **Answer:** ok (2026-11-17). E9's rule does not reach `regcred` (owner, same day).

### E11. B's admin password together with "existing"

- **About:** `semp.adminPass` is both the source for the Secret B builds and the password
  B's own SEMP calls use. With `credentials: {existing: NAME}` (E8) two sources would exist.
  Today naming the Secret without `adminPass` means "reference it", and no name and no
  password means the operator generates one. `semp.monitorPass` and `redundancy.psk` are
  entries of the same Secret and need `adminPass` today.
- **Recommended:** `existing` and `semp.adminPass` together are refused: the Secret holds
  the password. `create: true` without `adminPass` is refused. `adminPass` with no
  `credentials` block is an error that says to add `create: true`, not a silent fall-back to
  the operator's own Secret. No block and no password keeps today's operator-generated case.
  Not checked: whether every SEMP call of B can read the live Secret (it does for some,
  `ops_k8s.go:754`).
- **Other choice:** Allow `existing` with `adminPass` and use the value for SEMP calls only.
  Handy, but a stale password in the file fails every call with no hint that the Secret
  moved on.
- **Changes for:** B env files that set both a Secret name and `adminPass`.
- **Answer:** ok (2026-11-17).

### E12. Memory and CPU: one key name, two unit grammars

- **About:** Docker memory is `6898m` (m is megabytes). Kubernetes memory is `3410Mi`, and a
  Kubernetes `6898m` means 6898 thousandths of a byte. Both pass today's checks, so a value
  copied from docker to kubernetes asks for about 7 bytes. B's kubernetes memory has no
  check at all today.
- **Recommended:** One name on every platform, `<platform>.resources.memory` (and `cpu` on
  kubernetes, A only): A `kubernetes.resources`; B `kubernetes.resources.memory`,
  `docker.resources.memory`, `podman.resources.memory`, `kubernetes.operator.resources`.
  Each platform keeps its own grammar. The kubernetes check refuses a bare `m` suffix on
  memory, so a copied docker value fails at load. B's CPU stays fixed by the scaling tier.
- **Other choice:** Keep two names (`docker.mem`, `kubernetes.resources.memory`), so a copy
  cannot validate. Or M converts one grammar for every platform. Both cost an extra rule to
  learn.
- **Changes for:** B env files (`container.mem` and `msgNode.mem` are renamed). A: key moves
  from `deployment.resources`.
- **Answer:** ok (2026-11-18): `<platform>.resources.memory` in both tools, A gaining it on docker
  and podman; `kubernetes.resources.cpu` in A only; B's `msgNode.cpu` is dropped, not renamed (the
  scaling tier fixes B's CPU); a Kubernetes memory with a bare `m` suffix is refused.

### E13. Placement keys on Kubernetes

- **About:** A has `kubernetes.deployment.nodeSelector` (a mapping), `tolerations` and
  `affinity` in Kubernetes' own shapes. B has `placement.labelsPrimary` (lists of
  `"key: value"`), `tolerationsPrimary` (strings like `key=value:NoSchedule`) and its own
  reduced affinity words, per role.
- **Recommended:** Both write Kubernetes' shapes under `kubernetes.placement`:
  `nodeSelector` as a mapping, `tolerations` as Kubernetes entries, `affinity` as
  Kubernetes' block. In B, nodeSelector and tolerations are per role
  (`placement.primary.tolerations`); `affinity` is written once and applies to every role,
  as today. The operator's CRD takes all three (checked in the bundled CRD). B's string
  forms and its reduced dialect go. Cost to B: the load checks on operators and topologyKey
  must be rewritten against the new shape.
- **Other choice:** Keep B's string forms, and record the difference. Nothing changes for
  B's files, but `tolerationSeconds` cannot be written and the strings are not checked field
  by field.
- **Changes for:** B env files that set placement. A: the keys move out of `deployment`.
- **Answer:** Owner (2026-11-18): keep today's shapes. A and B deploy different kinds (A its own
  Deployment, B the operator's custom resource), so each handles placement its own way. B
  keeps `labelsPrimary`, `tolerationsPrimary` and its affinity words; A keeps Kubernetes'
  shapes. Recorded as an inherent difference; only the flatten of ruling 11 moves the keys.

### E14. Spread and B's always-on anti-affinity

- **About:** B always writes one "keep brokers on different nodes" term, over the namespaces
  in `antiAffinityNamespaces` (default: its own namespace), weight 100. There is no off
  value. The verifier found this is A's `spread: preferred`. A has
  `spread: none|preferred|required`, default `none`, and refuses it together with a
  `podAntiAffinity` of its own.
- **Recommended:** `kubernetes.placement.spread` in both (none, preferred, required). A
  defaults to `none`, B to `preferred`, so B changes nothing. B keeps
  `antiAffinityNamespaces` and `antiAffinityWeight` beside it, as broker-only keys. B gains
  an off value. A `podAntiAffinity` of the operator's own is then refused only when `spread`
  is not `none`. Without the off value that refusal would remove a feature B has today.
- **Other choice:** No refusal in B: it keeps appending the operator's own terms after the
  built-in one. Nothing breaks, but A and B differ.
- **Changes for:** B operators who write their own `podAntiAffinity`: they must add
  `spread: none`. Others: nobody.
- **Answer:** ok (2026-11-18): `kubernetes.placement.spread` in both, A default `none`, B
  default `preferred`; B gains `none` as an off value; `antiAffinityNamespaces` and
  `antiAffinityWeight` stay B-only.

### E15. The Service block on Kubernetes

- **About:** A: `kubernetes.service.enabled` and `service.port` (`servicePort:targetPort`,
  outside first). B: `kubernetes.ports` entries `"name=containerPort[:servicePort]"` (inside
  first), `loadBalancer.ip`, `loadBalancer.ipPool`, `loadBalancer.annotations`.
- **Recommended:** One block `kubernetes.service` in both. B's ports become entries in the
  custom resource's own words: `{name, containerPort, servicePort, protocol}` (the
  reviewers' `port` and `targetPort` are the Service's words, not the CR's). `annotations`
  becomes a mapping and `loadBalancer.{ip, ipPool}` moves under `service`. A keeps `enabled`
  and `port`. An old B string entry fails to load instead of swapping its two numbers.
  Record the LoadBalancer type as B's choice (the CRD also takes ClusterIP and NodePort) and
  A's `enabled` as A only.
- **Other choice:** Keep B's string entries and fix only the order. Smaller, but a string
  cannot carry a protocol cleanly and the order trap stays in the text.
- **Changes for:** B env files that set ports, annotations or the load balancer.
- **Answer:** Owner (2026-11-18): one `kubernetes.service` block in both, with `service.ports` (so
  no clash with the container ports of E4 and E7). Otherwise as recommended.

### E16. A gains `docker.compose`; command keys take a list

- **About:** B has `docker.compose: docker compose` (or `docker-compose`), so a host with
  only the standalone binary works. A has no key: it adds the word `compose` to
  `docker.command`. B also lets a command key be a list of exact arguments, which is the
  only way to write an argument with a space.
- **Recommended:** A gains `docker.compose`, default `<docker.command> compose`. Both tools
  accept a string or a list for every command key. Not checked on a host: A's compose file
  needs Compose 2.23.1 features (inline configs, environment secrets), and a standalone
  binary may not have them.
- **Other choice:** A stays without the key and keeps the string only. Less to write now,
  but the two tools differ for good, and an A host with only `docker-compose` cannot deploy.
- **Changes for:** A's operators on hosts with a standalone compose binary: they can deploy.
  Others: nobody.
- **Answer:** Superseded by E24's structured keys (2026-11-18): both tools take
  `docker.compose: plugin | standalone` (plugin runs `<cli> compose`, standalone runs
  `docker-compose`). No command key takes a string or a list any more.

### E17. Which programs `docker.command` and `docker.compose` may name

- **About:** B has one list for both keys: docker, docker-compose, nerdctl. The verifier
  found `docker.command: docker-compose` passes the guard, and then every engine call
  (login, inspect) runs `docker-compose`, with `docker-compose compose` for compose. A
  allows only `docker`; `--allow-command` widens it.
- **Recommended:** Per key, in B: `command` takes docker or nerdctl; `compose` takes docker,
  nerdctl or docker-compose. This fixes B's hole. A keeps `docker` for both until someone
  has run A's compose file and `image inspect --format` with nerdctl and a standalone
  compose (not checked on a host); the difference is recorded.
- **Other choice:** Give A the same lists now. One rule, but unverified for A.
- **Changes for:** B operators who set `docker.command: docker-compose` (now refused).
- **Answer:** Superseded by E24's structured keys (2026-11-18): `docker.cli` takes `docker` or
  `nerdctl`, `docker.compose` takes `plugin` or `standalone`, `kubernetes.cli` takes
  `kubectl` or `oc`, in both tools. B's hole closes because no key names a program freely.

### E18. Does A log in to the registry on docker and podman?

- **About:** With `image.user` and `image.pass` set, B runs `docker login --password-stdin`
  before deploy. A uses the same two keys only to build the Kubernetes pull Secret; on
  docker and podman the host must already be logged in.
- **Recommended:** No. Keep it as an inherent difference. The verifier showed that the
  reviewers' reason ("A pulls anyway") holds for one case only: docker, with a created
  store, when the image is missing. Two costs were not counted. A login writes the password
  into the engine's own auth file, against A's rule that no credential is left in a file on
  disk (stated from knowledge, not run). And existing A files that set the keys for
  Kubernetes would start logging in on docker and podman with no key changed.
- **Other choice:** Yes, through one new M method (`containercli.Client.Login`, password on
  stdin). The same keys then mean the same everywhere and the parity row goes. A third way:
  yes, but only behind a new switch such as `image.login: true`, so no old file changes.
- **Changes for:** With "no": nobody. With "yes": A's operators with registry keys.
- **Answer:** Owner (2026-11-18), in part: no separate login check. Deploy pulls the image
  through M first, and M reads the engine's error to say which it is: unauthorized or
  denied (a login problem), manifest or tag unknown (a wrong image), or the registry
  unreachable. Docker Hub answers "denied" both for a missing repository and for a
  missing login, so that message names both causes. On a login failure the tool stops
  with "not logged in to <registry>: run docker login <registry>" and never logs in itself,
  so no password lands in the engine's auth file. `image.user` and `image.pass` serve the
  Kubernetes pull Secret only. B loses its own login, and ruling 6's registry login does not
  move into M.
  Revised 2026-11-27: the pull follows a new `imagePullPolicy` on docker and podman (each
  platform's words), so a local-only image is not pulled under `missing` or `never`.

### E19. B's `podman.quadletDir`

- **About:** B lets the file name the directory the unit is written to. A removed that key.
  The default is the same in both: `/etc/containers/systemd` (rootful), else
  `$XDG_CONFIG_HOME/containers/systemd`, else `~/.config/containers/systemd`.
- **Recommended:** Keep the key in B and not in A; add the missing absolute-path check (B's
  doc says a relative value is refused, the code does not). The reviewers wanted it dropped,
  since systemd only scans fixed directories. The verifier found a real use: podman's
  generator runs under the user's systemd manager, which does not see an `XDG_CONFIG_HOME`
  set only in a shell profile, so the tool writes a unit nothing loads. The key is the way
  round that. Not checked on a host.
- **Other choice:** Drop it in B as A did. One less key, but that operator must fix the
  manager's environment instead.
- **Changes for:** With "keep": B gets one check; nobody else. With "drop": B operators who
  set it.
- **Answer:** Owner (2026-11-18): drop the key in B as well, as A did. The unit directory is
  always the default one; an operator whose user manager lacks `XDG_CONFIG_HOME` fixes the
  manager's environment.

### E20. The legacy `podman.baseDir` key

- **About:** B's `podman.baseDir` and A's `podman.base-dir` name a directory used by an
  earlier build. A already ignores it (a lint warning). B still deletes a legacy key file
  under it on deploy and remove.
- **Recommended:** Drop it in both, in the breaking tag, as migration is out of scope. B
  stops deleting the legacy file.
- **Other choice:** B keeps the clean-up for one more release. Hosts deployed by the old
  build are tidied, at the cost of a key the other tool does not have.
- **Changes for:** B hosts deployed by an earlier build keep a stale key file.
- **Answer:** ok (2026-11-18): dropped in both.

### E21. The read-only root key in A

- **About:** Decided 2026-11-13: A gets a key, read-only by default. Ruling 9: in the shared
  description the field has no default and the tool must choose. B has no key and refuses a
  read-only root by name. The verifier found that is not forced: B's own doc says the broker
  accepts a read-only root from 10.9 and B does not support it yet.
- **Recommended:** `readOnlyRoot: true|false` (default true) in A's docker, podman and
  kubernetes sections. B refuses the key until it supports it, recorded as "B does not
  support it yet", not as an inherent difference.
- **Other choice:** The name `readOnlyRootFilesystem`, Kubernetes' own word. Longer, and
  carries the same meaning on the engines.
- **Changes for:** A's operators: a new optional key.
- **Answer:** Owner (2026-11-18): both tools have the key. M requires a value (ruling 9); A
  defaults to true, B to false. Whether B accepts true depends on B supporting a read-only
  root (B's session).

### E22. The identity a pod runs as, on Kubernetes

- **About:** A: `kubernetes.runAs: "1000:0"`, the gid also as fsGroup. B: four fields of the
  custom resource (`securityContext.runAsUser`, `fsGroup`, `containerSecurity.runAsUser`,
  `runAsGroup`), omitted means the operator's 1000001 and 1000002, 0 means "the operator's
  default or an OpenShift id". The verifier found one uid:gid could fill all four, and the
  word `image` could mean leaving both blocks out.
- **Recommended:** Keep B's four fields and record it as B's choice, not as inherent. It
  loses nothing (a pod id that differs from the container id, and the meaning of 0), and no
  one asked for a change.
- **Other choice:** B also takes `kubernetes.runAs` and fills all four fields from it. One
  key in both tools, at the cost of the two features above.
- **Changes for:** With the recommendation: the parity doc only. With the other: B
  operators.
- **Answer:** Owner (2026-11-18): align with B's convention, which matches the Kubernetes
  shape: pod fields (`securityContext.runAsUser`, `fsGroup`) and container fields
  (`runAsUser`, `runAsGroup`), as in the operator's CRD (`spec.securityContext`,
  `spec.brokerContainerSecurity`) and a native pod. A's kubernetes section takes the same
  pod and container fields instead of one `runAs`. Exact key names at the flatten review.
  Still to settle there: B reads 0 or unset as the operator's default, while M's run-as
  reads a stated 0 as root (S3).
  Follow-up (2026-11-18): an unset field means the tool's default (A: user 1000, group and
  fsGroup 0); the word `default` leaves the identity to the image or operator.

### E23. The rule for a second image, B's operator image

- **About:** B prefixes `image.registry` onto `kubernetes.operator.image`. A takes its
  download image (`libs.download.image`) as a full reference. The verifier found a real file
  in B that repeats the registry inside `operator.image`, so it is prefixed twice. Neither
  rule is forced by a workload.
- **Recommended:** One rule: a second image is always a full reference, no hidden prefix.
  B's default `solace/pubsubplus-eventbroker-operator:1.4.2` stays (Docker Hub).
- **Other choice:** Keep B's prefix and record the difference; add a check that refuses an
  operator image that already starts with the registry. A site with a mirror keeps working
  without a new key.
- **Changes for:** B sites with a mirror: they must write the full operator reference.
- **Answer:** Owner (2026-11-18): every image key in both tools is the full block of E3
  (`registry`, `repository`, `tag`), B's operator image and A's download image included. No
  hidden prefix and no one-string reference.

### E24. `-n` inside `kubernetes.command`

- **About:** A's example env file shows `kubectl --context prod -n solace-connectors`, while
  every object is written for `kubernetes.namespace`. M adds `-n <ns>` to a call, after the
  command tokens, and nothing at all when the namespace is empty. A stray `-n` in the
  command can then win on a call that passes none. Gdu 11 only deals with `--context`.
- **Recommended:** Refuse `-n` and `--namespace` in `kubernetes.command` at load, with a
  message that names `kubernetes.namespace`, in both tools.
- **Other choice:** Leave it. Nothing breaks for files that agree, and a mismatch stays
  silent.
- **Changes for:** Operators who put `-n` in the command: a load error.
- **Answer:** Owner (2026-11-18): replace the free-form command keys on all three platforms
  with structured keys (for example `kubernetes.exec`, `kubernetes.contextFile`), so no
  free-form input reaches a command line and no parsing has to guard it. `-n` then has no
  place to appear. Settled the same day: `kubernetes.cli` (kubectl or oc),
  `kubernetes.contextFile` (`--kubeconfig`, optional), `kubernetes.context` (`--context`,
  optional; empty is the session's current context, pinned for the run, gdu 11);
  `docker.cli` (docker or nerdctl), `docker.compose` (plugin or standalone),
  `docker.context` (optional); podman has no command keys (local only). Programs resolve on
  PATH by their fixed name. Gone: the three `command` keys, A's `--allow-command` and the
  execguard token rules for them. `docker --config` is not offered (it can load plugins).

### E25. The list of inherent differences

- **About:** The parity body lists the differences the two workloads force, and the keys
  that only one tool has. The verifier moved several rows out of "inherent".
- **Recommended:** Accept this list as inherent: a digest in the image tag, JVM options with
  actuator port and syslog, TLS material, application credentials, the HA model, the scaling
  tier, replication access, libs against data directory, cpuset, process limits, the health
  check, replicas, storage, update strategy, operator keys. Record as "B's choice or
  one-sided, not forced": read-only root (E21), pod identity (E22), the second image's rule
  (E23). Record as one-sided, taken by A only on need: `imagePullPolicy`, `podAnnotations`,
  `podLabels`, `serviceAccount` (OpenShift grants its SCC to a service account, so A may
  need it). Spread is common (E14); only B's namespace list and weight are broker-only.
- **Other choice:** Align the one-sided keys now by adding them to A. More keys in A before
  any connector needs them.
- **Changes for:** The parity body only.
- **Answer:** Accepted (2026-11-18), with the list updated for the answers of that day.
  Inherent: a digest in the image tag (A); JVM options, actuator port and syslog (A); TLS
  material and application credentials; the HA model, `redundancy.sameHost`, the scaling
  tier and replication access (B); libs against data directory; cpuset and process limits
  (B); the health check; replicas (A); storage and update strategy (B); operator keys (B);
  placement shapes (E13); spread defaults and B's `antiAffinityNamespaces` and
  `antiAffinityWeight` (E14); the read-only root default (E21); identity defaults (E5, E22).
  One-sided, added to A only on need: `imagePullPolicy`, `podAnnotations`, `podLabels`,
  `serviceAccount`. No longer differences: the restart default, the image keys, the second
  image's rule, the pod identity fields, the command keys, the network block.

#### Settled by the rulings

- Spelling of multi-word keys: ruling 11, camelCase in both tools. The reference suffix
  becomes `Env` (`passEnv`). Kubernetes' own field names are camelCase too, so the
  reviewers' "keep them only inside API-shaped blocks" boundary no longer exists. How far
  A's rename reaches is E1.
- The compose project key: gdu 9 and gdu 19 decided it; ruling 11 spells it
  `docker.projectName` in both tools, default `solace-event-brokers` in B.
- Ports with host or none networking: ruling 9, the renderer leaves the ports out and its
  doc says so. No load error is added. B keeps checking the list in either mode, as today.
- Flatten the layouts: ruling 11. B's `docker.container.*` moves up to `docker.*`; A's
  `kubernetes.deployment.*` dissolves; `<platform>.name` on all three,
  `kubernetes.namespace`. The homes of the remaining keys are in E12, E13 and E14.
- The container name: rulings 11 and 13, a DNS label of at most 63 characters in both tools.
  B narrows (capitals, dots, underscores and 100 characters go). Defaults `connector` and
  `solace` stay.
- Restart when empty: ruling 9, an empty field writes `always`. A key for B is E6.
- Run-as grammar: ruling 9, numeric uid and gid, empty means the image's user. B's problem
  with that is E5.
- Root filesystem has no default in the shared description: ruling 9, with todo 6.
- Network words and defaults: todo 6, 2026-11-14 and 2026-11-15. The key shape is E4.
- Kubernetes gets no `hostNetwork` line: ruling 13.
- Quoting of every text value: ruling 9. A's `restart: no` is then written as `"no"`, which
  closes the YAML 1.1 defect in A's compose file.
- Expansion of `${VAR}` in values: gdu 2, so the keys carry no expansion difference.
- No preview mode: ruling 12, so no env key for one.
- Release order: ruling 8, gdu 19 ships inside the breaking tag, so each tool maps its
  config once and no old-key mapping is kept (see E2).

#### No decision needed

- Every credential pair is `<key>` and `<key>Env`, including `image.userEnv` in B (one more
  entry in its reference list), `semp.monitorPass`, and `replication.sites[].via.semp.pass`.
  The replication one has a third form, `passSecret`: exactly one of the three.
- An unset `...Env` variable fails B at every load, including remove. Gdu 3 decided remove
  resolves no credential, so B changes to match A (fail at deploy only).
- B's `ulimits` and `healthCheck` blocks keep their names under the flat section:
  `docker.ulimits.core`, `docker.healthCheck.enabled`; `cpuset`, `monitorCpuset`, `dataDir`
  move up one level.
- B values written unquoted into the custom resource with no load check (`serviceAccount`,
  `storage.class`, memory, volume claim names, `loadBalancer.ip` and `ipPool`,
  `antiAffinityNamespaces`) are quoted by ruling 9 and held to a grammar at load by gdu 4.
- B's `broker.cliScriptsDir`, `broker.hostDiagnosticDir`, `broker.domainCerts.dirs[].path`
  and `files` join the tilde and rebase path list of gdu 2. `fileExt` is already camelCase.
- B's eight snake_case scaling aliases (`system_scaling_*`, `messagespool_maxspoolusage`)
  stay: they are the broker's own setting names, not tool keys.
- B's `semp.additionalUsers` Secret (`<name>-additional-users`) stays derived with no
  switch, like A's stores Secret.
- A's keyless Kubernetes settings (liveness and readiness probes, container port name,
  Service without a type line, the `app: <name>` label) stay the tool's own. No key is
  added.
- A's `security.users[].name` (it names the derived secret) and `leaderElection.connRef` are
  application-only keys; they follow the rename of E1.
- A container host name (B writes it from `redundancy.<role>.name`) and B's
  `solace-util.sha256` annotation on the TLS Secret are not env keys. M's shared description
  must carry both. The host name is S2 and S12; the annotation is B's own label.
- Doc drift to fix with the reshape: A's `management.port` line that says the published port
  defaults to it (docker and podman ports are never defaulted); A's example that says the
  name is suffixed `-1`, `-2` past 20 workflows (validate refuses it); B's doc that says an
  unset `serviceAccount` uses the namespace default (the CRD creates a new one).
- `WantedBy` of a podman unit has no env key: it is Q11.

## The compose file

The compose file that M will write from the tool's description, for A and B. A decision that
also covers quadlet or Kubernetes is in the shared part (S1 to S13).

### C1. One order of keys in the file

- **About:** A and B write the service keys in different orders today. The module needs one
  order. Neither tool's order survives whole, and the ownership labels (gdu 1) change both
  files anyway. Proposed top level: `name`, `services`, `networks`, `configs`, `secrets`; a
  block with nothing in it is left out. Proposed service: image, container_name, hostname,
  user, restart, entrypoint, command, the hardening lines, read_only, mem_limit, the tool's
  extra keys, network, ports, environment, secrets, configs, volumes, tmpfs, healthcheck,
  labels. The labels come last, so the deploy-digest label ends the service block. B's
  service would read:

  ```
      image: "registry.example.com/solace-pubsub-standard:10.26.0.8755"
      container_name: "solace"
      hostname: "solace-primary"
      user: "1000001:0"
      restart: "always"
      privileged: false
      ...
      mem_limit: "6898m"
      networks:
        - default
      ports:
        - "8080:8080"
      environment:
        ...
      secrets:
        - source: "solace-admin-password"
          target: "/mnt/secrets/username_admin_password"
      volumes:
        - "/opt/solace/data:/var/lib/solace:Z"
      labels:
        "app.kubernetes.io/managed-by": "solace-util"
        ...
        "solace-util.deploy-sha256": "${SOLACE_DEPLOY_SHA256:-}"
  ```

  In A's file the labels move from just under `container_name` to the end, and volumes now
  come before tmpfs and the healthcheck (today the tmpfs and healthcheck come first).
- **Recommended:** Accept this one order, because the golden files change once anyway and
  one order means one writer and one test for every tool, including a third one.
- **Other choice:** Let each tool state its own order to the module. Each tool's golden
  keeps its line order apart from the new lines, but the writer needs an order list per tool
  and a third tool has to invent one.
- **Changes for:** The golden files of both tools (every compose golden changes once). A
  running deployment: the redeploy check hashes the file, so the first deploy after the
  upgrade asks and recreates the container. The env files do not change.
- **Answer:** ok (2026-11-17).

### C2. B's word "bridge" would move every deploy to docker's shared bridge

- **About:** B's env file defaults the docker network word to `bridge`, and today B turns
  `bridge` into no network line, which means the compose project network. Under gdu 6 the
  word `bridge` writes `network_mode: bridge`, which is docker's shared default bridge, not
  the project network. Passed straight through, every B deploy would change from
  `networks: - default` to `network_mode: bridge`. The broker's unpublished ports would then
  sit on the bridge that every other container on the host uses.
- **Recommended:** B drops its `bridge` default so an empty key reaches the module and gives
  the project network, as today. A written `bridge` means docker's shared bridge, as gdu 6
  decided, and B's sample file and configuration doc say so. This keeps what B does now and
  undoes no decision. Not checked on a real engine: the isolation between the two networks
  is standard docker behaviour that nobody ran here.
- **Other choice:** B refuses `bridge` on docker and tells the operator to leave the key
  empty. Nobody lands on the shared bridge by accident, but the words are no longer the
  engine's own, which gdu 6 decided. B's env file changes keys in the same tag (ruling 11),
  so one more line in the upgrade note is cheap.
- **Changes for:** solace-util env files that write `network.mode: bridge` by hand, its
  sample file and its docs. Its golden files gain the network lines.
- **Answer:** ok (2026-11-18), with `network.mode` kept (E4): absent writes the deploy's own
  project network on docker (B stops filling in `bridge`) and `Network=private` on podman; a
  written `bridge` means docker's shared bridge, as docker means it, and B's sample and docs
  say so.

### C3. Read-only root and how compose hands over secrets

- **About:** Ruling 9 settles that the root filesystem is chosen by the tool and read-only
  writes `read_only: true`. gdu 6 makes A's default read-only. The open point is whether
  that works for A's docker file at all. Compose hands environment secrets and inline
  `content:` configs to the engine by copying files into the created container. From memory,
  the engine refuses a copy into a read-only root outside a volume. If so, `up` fails for A,
  whose file carries credentials, `application.yml` and a status script this way. Not
  checked on a real engine.
- **Recommended:** Add a row to docs/smoke.md and run it before A turns the default on: an
  A-shaped file with `read_only: true`, one secret and one inline config, then `up`. The
  module writes what the tool states and refuses nothing, because nobody has shown the
  combination fails.
- **Other choice:** The module refuses `read_only: true` together with any secret or config.
  The failure would be a clear message, but it may refuse a combination that works, and gdu
  6's default for A could then never be met on docker.
- **Changes for:** solmq-conn-util docker only. solace-util writes no `read_only` line, so
  its rule 6 and its test hold.
- **Answer:** ok (2026-11-17).

### C4. Does the project network carry labels

- **About:** Ruling 13 settles that the compose file states its project network:

      networks:
        default:
          driver: bridge

  gdu 1 labelled the service because "neither tool renders a top-level network". Ruling 13
  changes that. The network is now an engine object that `compose down` removes. Still open:
  labels on that block. With gdu 19 two instances of one tool share a project, so they share
  this network, and an instance label would differ between them. A leftover network from a
  failed `down` could be found by its label, though. Not checked on a real engine: whether
  compose compares a network's settings on the first `up` after adoption (the old file left
  the network implicit), and what `down` does for one instance while another still uses the
  network.
- **Recommended:** No labels on the network. A label that differs per instance would make
  the second instance's file disagree with the network the first one made. Add both
  unchecked points to docs/smoke.md.
- **Other choice:** Label it with the tool only (`services.solace.com/product`), which is
  the same for every instance. A leftover network can then be traced to a tool. A second
  tool in the same project would still disagree. A third way is the full ownership labels on
  the network, with `AssertLabels` checking it: gdu 1's aim is that every object M makes
  says who owns it, but the per-instance label collides with sharing.
- **Changes for:** Running deployments: the network and the project are shared by instances,
  so the first deploy after adoption touches it. Golden files of both tools.
- **Answer:** ok (2026-11-17): no labels on the network.

### C5. A secret or config target must not sit under a mount

- **About:** A directory mount hides everything under it. If a secret target is inside a
  bind mount or a tmpfs, the secret is shadowed and nothing reports it. Both tools avoid
  this by construction today: A moved its status script out from under the libs volume for
  this reason, and B has a test that no secret nests with the certificate mount. Plain
  uniqueness of targets does not keep that. Example refused: secret target `/data/pw`, mount
  target `/data`.
- **Recommended:** Refuse a secret or config target that equals or lies under a mount or
  tmpfs target, and the reverse. Allow one nesting: a mount under a tmpfs, because A binds
  existing stores inside its truststores tmpfs. The reason: the rule then lives once, in the
  module, and a third tool gets it for free.
- **Other choice:** Leave it to each tool's tests, as today. A third tool can then ship a
  shadowed secret and only a real run shows it.
- **Changes for:** Nobody now: neither tool nests them. A third tool.
- **Answer:** ok (2026-11-17), after the owner noted that env.yaml sets no mount targets:
  the targets are fixed in each tool's code, so the check guards the description a tool
  builds, not operator input.

### C6. Secret variable names: duplicates and names the child process reads

- **About:** The tool puts `NAME=value` in the environment of the compose process. Two gaps
  in the model. First, B refuses two usernames that fold to one variable (`a-b` and `A_B`
  both become `A_B`), because one user's password would reach both; the module checks only
  names and targets. Second, a variable the child reads for itself (`PATH`, `LD_PRELOAD`,
  `HOME`) would change how it runs. A passes operator-chosen `-env` names straight into this
  list.
- **Recommended:** Refuse duplicates. Refuse `PATH`, `HOME`, `IFS`, `SHELL`, `BASH_ENV` and
  anything starting `LD_`, `DOCKER_` or `COMPOSE_`, for secret variables, the digest
  variable and the interpolated label variables. The reason: the module cannot trust a third
  tool's names. I found no `DOCKER_` or `COMPOSE_` credential name in A's sample files or
  docs.
- **Other choice:** Refuse only the six names B pins today. Fewer operators are touched, but
  a credential variable called `DOCKER_HOST` would point the docker client elsewhere.
- **Changes for:** An A operator whose `-env` variable starts with one of those prefixes
  must rename it. Nothing else.
- **Answer:** ok (2026-11-17).

### C7. Line breaks inside inline config text

- **About:** A inlines `application.yml`, the status script, the logback file and the stores
  script as `content: |` blocks. The reviewers proposed refusing a carriage return, other
  control characters and bad UTF-8. The verifier found three more line breaks that a YAML
  reader honours: U+0085, U+2028 and U+2029. Inside the text, one of them ends the block,
  and the rest is read as compose keys. The module's own helpers let them pass. A has the
  same gap today (it splits on newline only). Example, with `<LS>` for U+2028:
  ```
      content: |
        port: 8080<LS>privileged: true
  ```
  Not checked on a real compose; the YAML library's source says they end a line.
- **Recommended:** Refuse CR, U+0085, U+2028, U+2029, every other control character except
  newline and tab, DEL, the C1 range, U+FEFF and invalid UTF-8. Allow tab: shell and XML
  text may carry one, and a tab after the indent is plain content. The reason: it closes the
  hole in one place for every tool.
- **Other choice:** Refuse tab too, like `execguard.IsControl`. Simpler, but a script with a
  tab fails to render.
- **Changes for:** Nobody now: A's four texts carry none of these. A third tool's text.
- **Answer:** ok (2026-11-17).

### C8. Windows paths as a bind source

- **About:** The module would refuse a relative source and any colon in a source, because a
  colon adds a field to `source:target:options`. But A lets a Windows-authored path through
  on purpose, resolves it on the machine it runs on, and has code for Windows docker hosts
  (a Windows limit on environment size). On such a host the source is `C:\certs\store.jks`.
  B accepts a drive-letter data directory too and writes it with no guard. The line would
  be:
  ```
  - "C:\\certs\\store.jks:/app/external/stores/store.jks:ro"
  ```
- **Recommended:** Accept exactly one drive-letter root (`C:\` or `C:/`) and no other colon,
  and decide "absolute" in the module with its own function, not `filepath.IsAbs`, whose
  answer depends on the machine running. The reason: it keeps A's Windows docker path and
  still refuses the colon that breaks the line. Not checked on a real engine: that compose
  splits a drive-letter source correctly.
- **Other choice:** Refuse every colon and drop Windows docker hosts for A, as gdu 4 already
  does for a quadlet source. Simpler, but A loses a platform it has code for.
- **Changes for:** A on a Windows docker host. B on Windows: its drive-letter data directory
  now passes a check instead of none.
- **Answer:** ok (2026-11-17).

### C9. The `name:` line when `-p` is passed anyway

- **About:** gdu 19 and ruling 11 make the project name an env-file entry in both tools, and
  both pass it with `-p`. So the file's own `name:` is not needed to run. A file written by
  `generate` and run by hand has no `-p`, and a reader sees the project in it.
  `name: "solace-event-brokers"`.
- **Recommended:** The file always carries `name:`, required, in compose's grammar
  (`^[a-z0-9][a-z0-9_-]*$`), never folded and never left out. The tool passes the same value
  to the module's renderer and to `-p`. B's folding of the container name and its empty
  `name: ""` go away. The reason: one value in both places cannot drift.
- **Other choice:** Write `name:` only for `generate`. Fewer lines in deploy output, but the
  two paths then differ. Not checked: that `-p` outranks `COMPOSE_PROJECT_NAME`, which is
  standard compose but nobody ran it. If it holds, B's operators lose the variable as an
  override, a result of gdu 19.
- **Changes for:** B's golden files (A's already carry it when set). B env files only
  through gdu 19.
- **Answer:** ok (2026-11-17).

### C10. Inline text that does not end in exactly one newline

- **About:** For `content: |` A today silently gives the container a final newline, whatever
  the tool passed. A text ending in two newlines also loses one. The module can instead
  write the bytes it was given, with the right marker (`|-` for none, `|+` for several). A's
  four texts end in one newline, so for them the bytes do not change.
- **Recommended:** Write exactly the bytes the tool passed, using the marker the text needs
  (also an indent marker when the first non-empty line starts with a space). The reason: the
  container gets what the tool meant, with no hidden edit.
- **Other choice:** Normalise: always one final newline, as A does. The files stay as today,
  but the module keeps editing text it does not own.
- **Changes for:** Nobody now. A third tool with unusual text.
- **Answer:** ok (2026-11-17).

### C11. One service per file

- **About:** Both tools write one service. The redeploy check, `CheckComposePosture` and
  `CheckSecretTargets` all assume one. A tool with two containers would write two files into
  one project.
- **Recommended:** One service: `File.Service`, not a list. A list would bring `depends_on`
  and per-service secret lists that nothing has decided, and v0.x can widen it later.
- **Other choice:** A list now. A second container is possible in one file, but the redeploy
  digest, the check of the hardening lines and the secret check all need new rules first.
- **Changes for:** Nobody now; a later tool with two containers.
- **Answer:** ok (2026-11-17).

### C12. Named volumes and other networks

- **About:** Should this tag offer named volumes, or extra or external networks? Neither
  tool writes any of them. A named volume or a network is an engine object that needs
  ownership labels and a removal rule (gdu 1, 16, 18). Mode, uid and gid on a secret or
  config are S9.
- **Recommended:** No. A tool that needs one waits for a module change. A new engine object
  with no owner and no removal rule is how leftovers start.
- **Other choice:** Offer them. A tool could keep data in a named volume, but the labels and
  the removal rule have to be decided first.
- **Changes for:** Nobody now.
- **Answer:** ok (2026-11-17).

#### Settled by the rulings

- Does the compose file state its network at all (against gdu 1's "no top-level network")?
  Ruling 13: yes. Its labels are C4.
- B's compose file stays free of `read_only`: writable writes nothing (ruling 9), so its
  rule 6, its test and its audit grep hold.
- Compose file: version and lib annotation. gdu 1 keeps both out of a compose file, and
  ruling 10 feeds only the Kubernetes annotation. A module upgrade alone changes no byte and
  recreates nothing.
- A's `restart: no` is written `restart: "no"`, which closes the YAML 1.1 defect in A's
  compose file (ruling 9, quoting).
- B must refuse an empty run-as, because it chowns its data directory (ruling 9 and E5). B
  env files with a user name must change to numbers. Superseded in part by E5's answer: B's
  `default` reads a numeric image user for the chown.

#### No decision needed

- An error about inline text names the config and an offset and never the text, because A's
  `application.yml` carries the status password (S8).
- An empty environment value is written `"KEY": ""`, never a bare key, and a test pins it. A
  bare key makes compose take the value from its own environment, where the secrets are.
- Where M is stricter than a tool's load check, the tool mirrors it at load: retries at
  least 1 (this fixes B's negative-retries defect), a hostname that starts with a letter or
  digit, a numeric user. For `mem_limit` M takes the wider spelling (units in either case),
  so B loses nothing.
- M does not refuse uid 0: both tools accept it today, and an image's own user can be root
  anyway. Whether root is acceptable stays each tool's rule.
- The project directory (where compose reads a `.env`) stays with `ComposeUp` and
  `ComposeDown`. The renderer writes only absolute paths, so it does not depend on it.
- Values in the extras are text, integers, booleans, lists and maps with no raw text. Text
  gets the UTF-8 and control-character refusal, integers are 64-bit (B's `core`), and map
  keys may contain a hyphen (the example given was `logging`'s `max-size`; S1's answer now
  refuses `logging`).
- A healthcheck is the exec form only (`CMD` added by M, `$` doubled). A tmpfs takes only
  mode and owner. A secret takes no mode, uid or gid, as both tools write today (S9).
- The digest label stays a named field, written last as `"${VAR:-}"` with the `$` single. It
  and `Label.Var` are the only places that write an interpolation.
- An empty block is left out, and a key never holds null or an empty value.
- Every new line (`read_only`, `networks` or `network_mode`, the ownership labels) gets a
  compose version-floor row in the tool docs, as B's rule 8 asks. Inline content and path
  targets already need compose 2.23.1.
- One standard-library line writer serves M, and both tools drop their own for this file.
  Defaults (restart, user, project, healthcheck times) stay in each tool.

## The quadlet unit

The podman unit file (`<name>.container`) that M will write for A and B. A decision that
also covers compose or Kubernetes is in the shared part (S1 to S13).

Podman and systemd facts below come from the reviewers' memory, not from a run. Where an
answer rests on one, the entry says "not checked". This host has no podman source, so nobody
could open it.

### Q1. How B gets its `[Service]` limit lines

- **About:** B needs three `[Service]` lines: `LimitNOFILE=2448`, `LimitMEMLOCK=...` and
  `LimitCORE=...`. Ruling 9 lets extras pass unless M writes the key or the key is on a deny
  list, and the list must be complete. For `[Service]` and `[Unit]` that is hard. systemd
  has hundreds of `[Service]` keys, and some run host commands or change identity:
  `ExecStartPre=/some/cmd` runs as root in a system unit. A complete list would have to hold
  all `Exec*`, `User`, `Group`, `DynamicUser`, `Environment`, `EnvironmentFile`,
  `CapabilityBoundingSet`, `AmbientCapabilities`, `SupplementaryGroups`, `PrivateUsers`,
  `RootDirectory`, `RootImage`, `Type`, `NotifyAccess`, `PIDFile`, `NoNewPrivileges` (the
  systemd key, which binds the podman process) and `Restart`. In `[Unit]` it would also hold
  every key that starts another unit (`Requires`, `Wants`, `BindsTo`, `Upholds`,
  `OnFailure`, `Before`, `After`). M writes `Wants=network-online.target` and
  `After=network-online.target` itself, so ruling 9 also bars a tool from adding
  `After=other.service`. The key names are from systemd as the reviewers recalled them; not
  checked.
- **Recommended:** Extras reach the `[Container]` section only (S1 holds that deny list, as
  ruling 9 says). The unit model gets one small named field for `[Service]` limits, quadlet
  only, that accepts keys starting with `Limit` and checks the value. `[Unit]` gets nothing.
  A dependency on another unit gets a typed field with a unit-name check when a tool needs
  one. Reason: a deny list for systemd cannot be shown complete, and this keeps ruling 9's
  deny list for the extras without leaving a way to run a host command.
- **Other choice:** Follow ruling 9 as written: extras reach `[Service]` and `[Unit]` with
  the deny list above. A tool can then pass more, such as `Documentation=`, but one missed
  key is a way to run a host command. Or: extras reach those two sections with an allow list
  (`[Service]`: `Limit*`, `TimeoutStartSec`, `TimeoutStopSec`, `RestartSec`; `[Unit]`:
  none). That departs from ruling 9 for these two sections.
- **Changes for:** B's three lines move to the new field and keep their text. A third tool
  that wants `After=` or `RestartSec` hears so from the error, not from a silent drop.
- **Answer:** Settled by S2 (2026-11-17). M writes the `[Service]` `Limit*=` lines from the
  typed ulimits field. Extras reach `[Container]` only; `[Unit]` and `[Service]` take none. A
  dependency on another unit gets a typed field when a tool needs one.

### Q2. The compare of the installed unit stays beside the digest label

- **About:** Both tools read the unit file on disk today and compare it with the new text.
  Only a difference rewrites it. The reviewers proposed replacing that with the
  deploy-digest label (todo item 10). The verifier showed a loss: someone edits the file on
  disk (adds `AddCapability=SYS_ADMIN`) and restarts it. The running container's label still
  equals the tool's digest, so the deploy says "unchanged" and never rewrites the file.
- **Recommended:** Keep the on-disk compare and add the label compare. Either one differing
  counts as changed. The module hands back the unit text and the digest. The compare and the
  question stay in the tool, which is what todo item 10 already decided. This also keeps the
  header line "a deploy overwrites this file" true (Q10).
- **Other choice:** Label only. One compare less, but a hand-edited unit goes unnoticed
  until its text changes.
- **Changes for:** Nobody outside the module. Both tools keep what they do now and add the
  label check.
- **Answer:** ok (2026-11-17).

### Q3. Line order inside the unit

- **About:** The two tools order lines differently, and every unit changes once anyway
  (design note 8.4). Proposed: sections `[Unit]`, `[Container]`, `[Service]`, `[Install]`,
  with the module's own keys first in each and the extras' lines last. In `[Unit]` the order
  is `Description`, `Wants`, `After` (A writes `After` first today; B writes `Wants` first).
  In `[Container]`:
  ```
  Image=  ContainerName=  HostName=  Label=...  User=  Group=  NoNewPrivileges=
  DropCapability=  ReadOnly=  Network=  PublishPort=...  PodmanArgs=...  Environment=...
  Secret=...  Volume=...  HealthCmd=  HealthInterval=  HealthTimeout=  HealthRetries=
  HealthStartPeriod=  (then the extras)
  ```
  The groups are identity, privileges, network, resources, content, probe, extras. B's
  `Volume=` moves from before the environment to after the secrets, and its `Network` and
  `PublishPort` move ahead of the health lines.
- **Recommended:** Accept this order. Reason: one rule per section is easy to write down and
  to pin in the golden files.
- **Other choice:** Keep each tool's own order through a per-tool option. The goldens would
  not move, but the module would hold two layouts for one unit.
- **Changes for:** The golden files of both tools, and the first deploy after adoption: a
  running unit with the new text is restarted.
- **Answer:** ok (2026-11-17).

### Q4. Which lines are written quoted, and what is escaped where

- **About:** Ruling 9 says every text value a tool supplies is written quoted. In a unit
  file a quote is removed only for some keys. On other keys it would stay in the value. This
  entry asks where quotes are safe, and so asks you to confirm exceptions to ruling 9.
  Proposed:
  - Quoted whole, with the assignment inside: `Environment="TZ=Europe/Paris"` and
    `Label="deploy.solace.com/instance=prod 1"`. Inside the quotes `\`, `"`, `%` and `$` are
    escaped (the `$` is doubled because systemd expands it where quadlet copies the value
    into the service's command line).
  - Bare, because each value is held to a grammar with no space, quote, backslash or `$`:
    `Image=`, `ContainerName=`, `HostName=`, `User=`, `Group=`, `Network=`, `PublishPort=`,
    `Volume=`, `Secret=`, the `Health*` lines. In every one `%` is doubled, since systemd
    expands it in every line of a unit.
  - `Description=` is bare with `%` doubled. A systemd description is not unquoted, so
    quotes would show up in `systemctl status`.
  - Extras values (`ShmSize=2g`, `Ulimit=nofile=2448:1048576`) are bare, with `%` doubled. M
    does not know which extras keys unquote a value, so it refuses a value with whitespace,
    a quote, a backslash or `$`.

  The reviewers wrote "`%` and `$` doubled in every [Container] value". The verifier showed
  that is wrong for a mount source: as I recall, podman copies an absolute `Volume=` source
  into `RequiresMountsFor=`, where systemd expands `%` but not `$`, so `$$` there names
  another path. See Q7. Not checked on a real engine: that quadlet reads `Label="k=v"` the
  way it reads `Environment="k=v"`, and that it copies the `Volume=` source.
- **Recommended:** The list above, one write form per key by how quadlet reads that key. The
  bare keys are exceptions to ruling 9. If the `Label` quoting does not hold when someone
  reads it at podman's 4.5 tag, write `Label` bare and hold its value to a no-space grammar.
  Reason: quotes where the unit format does not remove them would be written into the value.
- **Other choice:** One rule per section, as the reviewers wrote: double `%` and `$` in
  every `[Container]` value. Simpler to document, wrong for paths.
- **Changes for:** Golden files of both tools. A's labels become quoted. B's `HealthCmd`
  value no longer carries a bare `$` (Q6).
- **Answer:** ok (2026-11-17): one write form per key; the bare keys are exceptions to
  ruling 9.

### Q5. File mode of the installed unit: 0600 or 0644

- **About:** A installs its unit at 0644 (world-readable) and rewrites it in place. B
  installs at 0600, applied only when the file is first created. Both create the directory
  0755. `docs/todo.md` item 10 already records A's world-readable unit as a defect. The unit
  holds `Environment=` values and secret names, and a `Label=` line becomes an argument of
  `podman run`, which any local user can read. The mode belongs to the artifact, so M's
  renderer can state it and hand it back with the text. Both tools would write through the
  atomic write helper (todo item 7, S11) with one mode from M.
- **Recommended:** 0600, with a 0755 directory. The quadlet generator runs as the account
  that owns the file (root for a rootful unit), so it still reads it (reasoned from how
  systemd generators run, not run here). 0600 does not hide values written in `Environment=`
  or `Label=`; see Q9. Reason: least privilege.
- **Other choice:** 0644 for both. Anyone can read the unit, as in A today.
- **Changes for:** A's units: the next deploy replaces a 0644 file by a 0600 one. Anything
  that reads `/etc/containers/systemd/*.container` as another user stops being able to. B's
  existing files, written at 0600, stay as they are.
- **Answer:** ok (2026-11-17): 0600, directory 0755. The S11 umask rule covers `generate`
  output only; the installed unit is written 0600 through the same helper.

### Q6. What a health command means on podman

- **About:** The shared health check holds an argument list: `["curl", "-fs",
  "http://localhost:5550/..."]`. Compose runs that list with no shell. On podman the unit
  line is one string: `HealthCmd=curl -fs http://localhost:5550/...`. As the verifier
  recalls podman's documentation, a `HealthCmd` that is not a JSON array runs through a
  shell in the container. Then `;`, `|`, `&&`, quotes, `$(...)` and globs have meaning, so
  the same description does two things on docker and on podman. B documents `sh -c '...'`
  probes as allowed. Joined with spaces, `sh -c 'curl x || exit 1'` loses its quotes, so
  such a probe works on docker and is broken on podman today. Not checked on a real engine.
- **Recommended:** Keep the argument list. On podman, refuse a token that holds whitespace
  or a shell character (`; | & $ < > ( ) ` ' " \ * ? [ ] { } ~ ! #`) and say so in the
  error. The built-in probes (`curl -fs
  http://localhost:5550/health-check/readiness` and A's `sh /app/external/.status-script
  --health`) pass. The JSON array form, which needs no shell, is a later change after a
  check on the oldest podman. Reason: a probe then means the same on both engines or fails
  at generate, never silently differently.
- **Other choice:** Call the field a command line, join the tokens, and say in the guide
  that podman runs it through a shell. No refusal, but the same description behaves
  differently per engine.
- **Changes for:** B's env files with a `sh -c '...'` probe or a `?` or `&` in a URL: they
  work on docker and fail at generate on podman, with a message, instead of mis-rendering. A
  is not affected (constant command).
- **Answer:** ok (2026-11-17).

### Q7. What a mount source may contain

- **About:** `Volume=/opt/libs:/app/external/libs:ro` is `source:target:options` on one
  line. M will check the source it is handed. That covers A's libs path, which is resolved
  against the env file's directory and today is checked only as written, not after
  resolving. B's data directory is written raw: `Volume=/data:/var/lib/solace:Z`. B's own
  decision to escape `%` and refuse `:` is recorded but not built, though `docs/todo.md`
  words it as done. Proposed grammar: an absolute path, and none of `:`, whitespace, control
  or invisible characters, `$`, quotes, backtick or backslash. `%` is allowed and doubled. A
  relative source is refused, because podman reads it as a named volume (A can produce one
  when its resolver is missing). Consequence: an env file in a directory with a space or a
  `:` in its path fails generate and deploy on podman when `libs` is set, with an error
  naming the path.
- **Recommended:** Accept the grammar. `$` is refused, not doubled, per Q4. Reason: both
  tools already refuse `$` and quotes as written, so nothing is lost, and the resolved path
  is covered.
- **Other choice:** Allow `$` and double it, as the reviewers wrote. A path with `$` then
  means something else wherever systemd expands only `%` (not checked on a real engine).
- **Changes for:** A's operators whose env file sits in a path with a space or `:`. B's
  operators whose `dataDir` holds `%`, `:` or a drive letter such as `C:/x`: refused, where
  today the line is written as is.
- **Answer:** ok (2026-11-17).

### Q8. SELinux relabel of a mounted host directory

- **About:** B writes `Volume=/data:/var/lib/solace:Z`. A writes `...:ro` and no label. On a
  host with SELinux enforcing, an unlabelled bind mount is usually unreadable in the
  container (not run). Adding `z` or `Z` relabels the operator's own directory on the host,
  and stays after the container is gone. The reviewers recommended `ro,z` for A. The
  verifier objected: `z` (shared) lets every container on the host read the directory, and
  A's libs directory often sits under a home directory. SELinux behaviour here is from
  podman's documentation as recalled; not checked.
- **Recommended:** The mount type holds one of: none, `z`, `Z`. The forgotten value is none:
  no relabel. A stays as today (none) and lists `z` as a widening it may ask for; B keeps
  `Z`. Reason: A's unit changes no host state today, and a relabel should be a choice.
- **Other choice:** A relabels with `ro,Z` (private to the container, per instance). The
  libs directory then cannot be shared by two A instances.
- **Changes for:** A: nothing, unless it chooses a label. B: unchanged.
- **Answer:** ok (2026-11-17).

### Q9. Passwords written into environment and label lines

- **About:** The reviewers said the model "has no field that can hold a secret". The
  verifier showed that `Environment=`, `Label=` and the extras values are free text, and the
  module cannot tell a password from a path. Quadlet turns those lines into arguments of
  `podman run` in the generated service, so any local user can read them (`ps`,
  `systemctl show`), whatever the file mode (todo.md already records this for labels).
  Example of what must not happen: `Environment="DB_PASSWORD=hunter2"`. Secrets belong in
  `Secret=name,type=mount,target=/run/secrets/db`. B's own keys such as
  `username_admin_passwordfilepath` carry a file path and are fine.
- **Recommended:** Words, not name matching. The guide and type docs say secrets go only
  through the secret list. A render error names an entry by its name and never prints its
  value. A check in `artifacttest` lets a tool run the same test on its own units. Reason: a
  name filter refuses B's `...passwordfilepath` keys and misses a password under another
  name.
- **Other choice:** Refuse an environment name that holds `password`, `secret` or `token`
  unless it ends in `filepath`. Catches the obvious slip, but it is wrong in both
  directions.
- **Changes for:** Tool authors. No env file changes.
- **Answer:** ok (2026-11-17).

### Q10. First line of the unit

- **About:** A writes `# Generated by solmq-conn-util -- podman quadlet unit for the Solace
  PubSub+ Connector for IBM MQ.` above `[Unit]`. B writes no comment. Proposed for both: `#
  Generated by solace-util. Do not edit: a deploy overwrites this file.` The tool name comes
  from the identity. No tool version and no module version, so an upgrade alone changes no
  unit and restarts nothing (todo decision of 2026-11-14). `artifacttest`'s own fixture
  models A's shape.
- **Recommended:** Yes, the one line above for both tools. Someone who edits the unit is
  told what will happen. It stays true only if Q2 is accepted.
- **Other choice:** An optional header text on the unit, which A fills with its current line
  and B leaves empty. A's golden stays as it is. Or no comment, as B today: then nothing in
  the file says who wrote it (the labels say so only to someone who reads them).
- **Changes for:** A's goldens and A's unit text (the old product line goes). B's unit gains
  one line. With the optional header: nothing for A.
- **Answer:** ok (2026-11-17).

### Q11. `WantedBy` follows rootful or rootless

- **About:** No env key controls it. B writes `WantedBy=default.target` rootless and
  `WantedBy=multi-user.target` rootful. A writes `default.target` in both modes, though it
  has the key `podman.rootful`. One renderer must pick one rule. Proposed: M writes B's rule
  from a rootless flag. The tool also passes that flag when it asks for the unit directory,
  and the `systemd` client's user mode comes from the same flag.
- **Recommended:** B's rule: `default.target` when rootless, `multi-user.target` when
  rootful, derived from `podman.rootful` in both tools. A system unit belongs to the
  multi-user boot target, and one flag then chooses the directory and the target. It is also
  systemd's own convention. Not checked on a real host: whether A's rootful units start at
  boot today.
- **Other choice:** `default.target` always, A's rule. Nothing changes for A; B's rootful
  unit loses its stated target. Not checked on a real host: that `default.target` pulls the
  unit in at boot just as well for a system unit.
- **Changes for:** A's rootful podman hosts: one line of the unit and the golden change, and
  the next deploy restarts the unit anyway. Rootless units and B are unchanged.
- **Answer:** ok (2026-11-17).

#### Settled by the rulings

- Writable paths on a read-only root, for A: they wait for todo item 6's plan and use the
  extras (`Tmpfs=` passes). The root filesystem rule itself is in the shared part.
- A label is written quoted whole: ruling 9. That quadlet reads `Label=` like `Environment=`
  is from memory and unchecked. Q4 has the fallback.
- The module version: ruling 10 stamps it only for Kubernetes's lib annotation. A unit
  carries no version (todo decision of 2026-11-14), so the header (Q10) has none and `devel`
  does not touch units.

#### No decision needed

- The lists of forbidden keys live in `quadlet` or a small shared package, and
  `artifacttest` reads them. `artifacttest` imports `quadlet`, `compose` and `testing`, so a
  renderer cannot import it.
- A key that M writes itself is refused in the extras even when the unit does not carry it:
  extras lines come last, and for a one-value key the last line wins.
- M checks the container name with `containercli.ValidName` and that `<name>.container` fits
  255 bytes. The tools' DNS-label rule comes in on top.
- `Render` stamps the deploy-digest label itself (two passes, digest last among the labels)
  and returns the digest: todo item 10 decides the label and design note 8.2 puts labels in
  the renderer. The quadlet digest is a method of the unit renderer, not an exported twin of
  `ComposeDigest`.
- The zero value of the network field writes `Network=private`, from the six podman words;
  `container:`, `ns:` and network names are refused (todo item 6). Whether the oldest podman
  accepts `Network=private` is still the owner check in the decisions file.
- `[Service]` is always written, `Restart=` first, then B's three `Limit*` lines (Q1).
- The unit directory is outside the model: `Render` gives the file name and the tool joins
  the directory. B's `podman.quadletDir` key is E19.
- M offers no unit parser. A's read of the installed `Secret=` lines goes away when the
  labels land.
- M keeps no podman floor: its own lines stay within podman 4.5, `HostName` (4.6) and the
  extras' keys raise it only for the tool that sets them, and `CheckQuadletKeys` stays the
  tool's test. `ReadOnly` goes into its release table with the release read from podman's
  source, not yet read.
- Errors name an `Environment`, `Label` or extras entry by its key and never print its value
  (golden rule 3, secrets masked; S8).
- The image string follows the reference grammar (S7). B still cannot name a digest unless
  its own config gains one.
- The port grammar is B's wider one (`[ip:]host:container[/tcp|/udp]`, ranges). A's two
  integers fit inside it.
- Environment entries keep the tool's order. A duplicate name is an error, since quadlet
  keeps only the last. `TZ` beside the timezone field is an error.
- Secret lines are `Secret=<name>,type=mount,target=<absolute path>`. An empty name, a
  relative target, a comma or a duplicate target is an error, and A's silent skip of an
  empty name ends.
- `Description` is one string the tool supplies. `%` is doubled, and a control character, a
  line break or a trailing backslash is refused.
- M writes the five ownership labels, then the tool's own, then the digest. A repeated key
  or one that M owns is refused.
- The installed unit and generate's output file both use the atomic write helper of todo
  item 7 (S11), with the one mode from Q5. The symlink refusal and the overwrite question
  stay in the tools.
- B's `/etc/sysctl.d/99-solace.conf` is not a quadlet artifact and stays in B, moving to the
  atomic helper with the unit.

## The Kubernetes documents

The Kubernetes documents that M will write for A and B: Namespace, Secret, ConfigMap,
Service, volumes, Deployment, a custom resource and raw vendor text. A decision that also
covers compose or quadlet is in the shared part (S1 to S13).

### K1. Image pull policy

- **About:** The reviewers wanted M to always write `imagePullPolicy: IfNotPresent` when the
  tool sets nothing. A writes no such line today, so Kubernetes picks its own default:
  Always for the tag `latest` or no tag, IfNotPresent otherwise. A allows `tag: latest`, so
  it can run today. A forced IfNotPresent would use whatever image the node already holds
  under that name.
- **Recommended:** Write the line only when the tool sets a policy. An empty field writes
  nothing, so Kubernetes keeps its own default and A's behaviour does not change. The
  reviewers' default was weaker than Kubernetes' own for `latest`, and B's IfNotPresent (in
  its custom resource) is no precedent for A's Deployment. B keeps its own default inside
  its body.
- **Other choice:** Always write the line, computed by M: Always for `latest`, no tag or no
  digest, otherwise IfNotPresent. Explicit in the golden, but M then encodes an image-tag
  rule that Kubernetes already has.
- **Changes for:** Nobody with the recommended answer. With the reviewers' default, A's
  users who deploy `latest` would stop pulling new images on a restart.
- **Answer:** ok (2026-11-17).

### K2. Secrets written as data only

- **About:** A writes its credentials Secret with `stringData:` and plain text. Its stores
  and image-pull Secrets, and all of B's, use `data:` with base64. The reviewers recommend
  `data` everywhere. One reason is that a key you remove from `stringData` can stay behind
  in the live Secret's `data` after the next apply (not checked on a real engine).
- **Recommended:** Data only, from raw bytes that M encodes itself. A value then cannot be
  written twice-encoded or unencoded, and A's `generate` output stops showing passwords as
  text. Base64 is not protection, it only stops accidental reading.

      type: Opaque
      data:
        "password": "c2VjcmV0"

- **Other choice:** Keep `stringData` where it is today. Output a person can read and check,
  but two forms, and the stale-key problem stays.
- **Changes for:** A's credentials Secret golden and `generate` output. A running deployment
  sees the same Secret content, so no pod restarts because of this.
- **Answer:** ok (2026-11-17).

### K3. Name-only references for a remove

- **About:** A remove must not need a certificate or a password. A's remove stream today
  names every Secret by name only, and always names the stores Secret, even when this env
  file renders none, so one left by an earlier deploy is removed. As the reviewers modelled
  it, a Secret cannot exist in M without its data, so a remove would have to read
  credentials, or invent them, just to learn names.
- **Recommended:** A small reference document in M: kind, name, namespace and nothing else.
  The tool builds its remove list from names, in reverse order, and M writes it. M can also
  cut a full document down to its reference. B's ownership list (admin, TLS, pull,
  additional-users) works the same way. A's remove stream stops carrying `application.yml`
  and its status password, and B's stops carrying the whole custom resource.

      apiVersion: v1
      kind: Secret
      metadata:
        name: "broker-stores"
        namespace: "prod"

  Standard kubectl behaviour: a delete needs only apiVersion, kind and name.
- **Other choice:** Derive the remove list from the deploy documents only. Then a remove
  needs material it must not read, and A's leftover stores Secret is no longer named.
- **Changes for:** A's and B's remove streams. B deletes its Secrets one by one today with
  `--ignore-not-found`; whether it moves to the stream is B's adoption step.
- **Answer:** ok (2026-11-17).

### K4. A CRD in a remove stream

- **About:** B's operator bundle holds the CustomResourceDefinition. Deleting it removes
  every custom resource of that kind in the whole cluster. B deletes it today only with
  `--delete-crd`, after a gate that checks for dependants. If the bundle text passes through
  the reference function as raw documents, that CRD would pass too.
- **Recommended:** The reference function refuses a CRD document, raw or not. B keeps its
  own `kubectl delete crd` command, run through the generic run (ruling 2), behind its own
  gate, as today. M cannot judge whether other custom resources exist, so it should not
  offer the delete.
- **Other choice:** M offers a separately named call for CRDs. One more guarded path in M,
  and M still cannot see the cascade B's gate checks.
- **Changes for:** B only, and only in how it spells the command. Nobody else writes CRDs.
- **Answer:** Owner (2026-11-17): a CRD is never deleted. The reference function refuses a
  CRD as recommended, and B drops its own CRD delete (`operator remove --delete-crd`, B
  `internal/cli/ops_k8s.go:597`) as well. The CRD is applied when the operator is deployed,
  as today. A change for B's session.

### K5. Custom resource body as a tree

- **About:** B's custom resource is written by about 160 lines of B's code. Either B hands M
  a tree of keys and values and M writes the text, or B hands over finished text and M only
  adds the metadata. A tree means M's quoting also covers the values B writes bare and
  unchecked today: service account, storage class, node memory, claim names, load-balancer
  address and pool, anti-affinity namespaces. B still decides every key and value.
- **Recommended:** A tree. B's golden changes once either way. The cost is in B: it must
  pass the numbers as numbers, because a quoted number is refused by the operator's schema.
  Those are `runAsUser`, `fsGroup`, `runAsGroup`, `containerPort` and `servicePort`, which B
  holds as strings today. Not checked against the real operator.

      image:
        repository: "solace/solace-pubsub-standard"
        tag: "10.10"
        pullPolicy: "IfNotPresent"
      securityContext:
        runAsUser: 1000001

- **Other choice:** Text the tool hands over. No change in B's code, but the unchecked bare
  values stay and M's quoting rule does not reach the largest document either tool writes.
- **Changes for:** B's golden and its code that builds the body. The operator receives the
  same values after parsing, apart from the quoting of text.
- **Answer:** ok (2026-11-17).

### K6. Custom resources and the typed kinds

- **About:** The reviewers want the custom-resource document to refuse any kind M has a
  typed document for (a Deployment, a Service), so nobody writes a Deployment around the
  hardening lines. The verifier found that this is no guarantee: raw text accepts any kind
  (B's bundle needs a raw Deployment), and the custom document accepts a Pod, StatefulSet,
  DaemonSet or Job with a hand-written template.
- **Recommended:** Keep the refusal, but do not count on it. The guarantee comes from
  `artifacttest`: its Kubernetes check must require the three container lines
  (`allowPrivilegeEscalation: false`, `privileged: false`, drop `ALL`) and the pod's
  `seccompProfile`, not only forbid bad tokens, on every workload a tool ships except raw
  vendor text. Raw text is the owner's explicit exception: the vendor Deployment states only
  two of the three lines.
- **Other choice:** Also refuse Pod, StatefulSet, DaemonSet, Job and CronJob in the custom
  document until M models them. Strongest, but a third tool that needs a StatefulSet is
  blocked.
- **Changes for:** Tools that ship tests. A third tool with a StatefulSet passes through the
  custom document and is checked by the test.
- **Answer:** Owner (2026-11-17), after YAML examples: M types every workload kind, so the
  hardening lines are built in everywhere: Deployment, StatefulSet, ReplicaSet, DaemonSet,
  Job, CronJob and Pod. `Custom` refuses every kind M types, and `artifacttest` still
  requires the hardening lines in every workload a tool ships, raw vendor text excepted.
  This overrides K18 for these kinds. Asked next, the owner also typed now: Ingress,
  NetworkPolicy, PodDisruptionBudget, HorizontalPodAutoscaler, ServiceAccount, Role,
  RoleBinding, ResourceQuota, LimitRange, PriorityClass and StorageClass. ClusterRole and
  ClusterRoleBinding are refused (cluster-wide privilege). Gateway API kinds wait.

### K7. Labels on the PersistentVolume and A's warning

- **About:** Every object carries the ownership labels and the lib and version annotations,
  the cluster-wide PersistentVolume included (decided 2026-11-14 and 15). A already prints a
  warning whenever its apply output says `persistentvolume/<name> configured`: that a
  volume's size and nfs path cannot change by applying again. Adding labels makes the output
  say `configured` on the first deploy after adoption. The lib and version annotations
  change on every module or tool upgrade, so it would then say so each time, and the warning
  would then be false.
- **Recommended:** Keep the rule for every object. Change A's warning (in A's adoption step)
  to read the volume's current size and nfs path with one `kubectl get` and warn only when
  they differ from what it renders. If the identity may not read volumes, the read is
  skipped with a note and nothing is blocked, the least-privilege rule for a convenience
  step.
- **Other choice:** Leave the lib and version annotations off the PersistentVolume only. A
  special case in M, against the decided "every object", and the first deploy after adoption
  still warns once because of the labels.
- **Changes for:** A's users, on warnings only. Pods are not touched: labels and annotations
  on an object are not part of the pod template.
- **Answer:** ok (2026-11-17).

### K8. Secrets in environment variables

- **About:** The Deployment model has no way to put a Secret key into an environment
  variable (`secretKeyRef`) and no `envFrom`. Secrets reach the container as files. A's
  rule: credentials are never environment variables, where a child process or crash dump
  sees them. B's only use of `envFrom` sits inside its custom resource (the operator's
  `extraEnvVarsSecret`), not in a Deployment.
- **Recommended:** No, for both. The container model stays files only, and `envFrom` is on
  the forbidden list of S1. A third tool whose program reads only environment variables must
  write the value to a file the program reads or pass it some other way, or wait until it
  brings its case.
- **Other choice:** Allow `valueFrom.secretKeyRef` (one named key per variable), still no
  `envFrom`. Easier for third tools, but it puts the secret where A's rule says it must not
  go.
- **Changes for:** Third tools only.
- **Answer:** Owner (2026-11-17): no for M (files only, no `envFrom`). B's additional users
  on Kubernetes keep `spec.extraEnvVarsSecret` as the one stated exception: the operator's
  CRD (1.4.2, embedded in B at `internal/k8s/assets/operator-1.4.2.yaml.tmpl`) has no field
  that mounts a Secret as files, only fixed-key Secrets, env-var sources and a data-volume
  PVC. The exception ends when the operator can mount files.

### K9. The operator bundle stays raw text

- **About:** B embeds the Solace operator's vendor YAML (about 1,970 lines, eight kinds, six
  substitution points). The reviewers recommend M carries it as raw documents and never
  models it. M then adds no ownership labels or annotations to it, and its own Namespace,
  with the `control-plane: controller-manager` label, stays as the vendor wrote it. Remove
  never deletes that namespace and the leftover removal of gdu 18 does not cover it.
- **Recommended:** Yes, raw, unlabelled. Re-modelling it would turn each operator upgrade
  into new Go types. The "every object carries the lib annotation" rule has this one stated
  exception. M refuses a Namespace document in a remove stream by its kind, raw or not (see
  the list at the end). Only the `regcred` Secret is rendered by M (K10).
- **Other choice:** M stamps labels into the vendor text. M would then rewrite upstream YAML
  line by line, and a change in vendor layout breaks it.
- **Changes for:** B's `operator generate` and apply output: unchanged text, except the
  stream ends with a newline, which the bundle itself lacks today.
- **Answer:** ok (2026-11-17).

### K10. Instance label of the operator's pull Secret

- **About:** M writes the label `deploy.solace.com/instance` on every object it renders, and
  a destructive step selects on it. B's fixed `regcred` Secret belongs to the operator
  install, shared by every env file in the cluster, not to one broker. So no broker name
  fits as its instance.
- **Recommended:** The operator's own deployment name, `pubsubplus-eventbroker-operator`. It
  is the one name the install already has, and it never matches a broker's instance.

      metadata:
        name: "regcred"
        namespace: "solace-operator"
        labels:
          "deploy.solace.com/instance": "pubsubplus-eventbroker-operator"

  (`solace-operator` is an example namespace.)
- **Other choice:** A value B makes up, such as `operator`. Also fine, but a second name to
  keep in step with the deployment.
- **Changes for:** B's operator Secret only.
- **Answer:** Owner (2026-11-17): `regcred` is treated like the raw operator bundle: the fixed
  name `regcred`, no ownership labels, created and deleted by name. A second stated exception
  to "every object carries the labels", beside the bundle (K9).

### K11. Registry key when the image names none

- **About:** An image-pull Secret keys its login by the registry host. A writes
  `https://index.docker.io/v1/` when the image has no host. B writes the key as typed, so an
  env file with no registry writes an empty key (`"auths":{"":{...}}`), which a pull would
  not match (not checked on a real engine). M will refuse an empty key.
- **Recommended:** M exports the Docker Hub key as a constant. B uses it when
  `image.registry` is unset, and the host as typed otherwise, as A does. B's Docker Hub pull
  Secret then starts to match.

      "auths": {"https://index.docker.io/v1/": {"username": "u", "password": "p", "auth": "dTpw"}}

- **Other choice:** B refuses an env file with credentials and no registry. Stricter, but an
  env file that loads today stops loading.
- **Changes for:** B's env files with credentials and no registry host.
- **Answer:** ok (2026-11-17).

### K12. A's own affinity block

- **About:** A lets the operator write a free `kubernetes.deployment.affinity` block, which
  A copies into the Deployment as it is. A reads it with a YAML library M must not import,
  so A would convert it to a tree and give it to M. Three things change in that block's
  bytes: a list item that is a mapping now starts on the dash line (today a lone `-` and the
  mapping two deeper), a number stays a number only if A converts by what the YAML says it
  is, and an empty `{}` is kept. Today A writes `labelSelector: {}` as a bare
  `labelSelector:`, which is null. `{}` selects every pod and null selects none, so that is
  a bug A has today.
- **Recommended:** Keep the block. A converts it: whole numbers bare (a `weight: 50` must
  not become the text "50"), booleans bare, `{}` as an empty mapping, everything else
  quoted. The empty-mapping bug is fixed, with a test.

      podAntiAffinity:
        preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 100
            podAffinityTerm:
              topologyKey: "kubernetes.io/hostname"

- **Other choice:** Narrow the block to M's typed terms. Simpler, but a field an operator
  already uses in the block would stop being accepted.
- **Changes for:** A's operators who use the block: the text moves, the meaning does not.
  The `{}` fix changes the meaning of a block that has one, on purpose.
- **Answer:** ok (2026-11-17).

### K13. A dollar sign in exec probes

- **About:** The kubelet expands `$(NAME)` in a container's command and args, and reads `$$`
  as one `$`. M already doubles every `$` in env values, command and args. The verifier says
  it expands the command of an exec probe the same way (from kubelet source as remembered,
  not checked on a real engine). If so, a `$` in a probe command needs doubling too. A's one
  probe is three constants, so nothing changes for it.
- **Recommended:** Double the `$` in probe commands, as everywhere else. Add a hand check on
  a real cluster before the tag. If the kubelet did not expand, a probe such as
  `sh -c 'echo $$'` would show the shell's id and the probe fails loudly.
- **Other choice:** Refuse `$` in a probe command. Safe whatever the kubelet does, but a
  third tool must put its probe in a script file.
- **Other choice:** Write the command as given. The least code, and a `$(VAR)` in it would
  be expanded without anyone meaning that.
- **Changes for:** Third tools only.
- **Answer:** ok (2026-11-17).

### K14. How the Secret file mode is written

- **About:** A Secret mounted as files has `defaultMode`. A writes `0400`. kubectl reads
  that as octal 256. A YAML 1.2 reader reads it as decimal 400, which is a different mode
  (not checked on a real engine). The decimal `256` cannot be misread but is harder for
  people to read.

      secret:
        secretName: "broker-credentials"
        defaultMode: 0400

- **Recommended:** `0400`, the text A writes today. kubectl is the only reader of these
  files and people read the generate output. The golden lines do not change.
- **Other choice:** `256`. Right for any reader, including a strict YAML 1.2 tool a team
  might put in front of the output. The golden lines change and the cluster does not notice.
- **Changes for:** Nobody with `0400`.
- **Answer:** ok (2026-11-17): `0400`.

### K15. One list of what a manifest writes

- **About:** A and B each keep a list of the objects they write, for the permission checks,
  the namespace gate and the leftover compare. They can drift from what the renderer writes.
  A also leaves its cluster-wide PersistentVolume out of its list today, so its permission
  probes ask only about namespaced objects.
- **Recommended:** M gives one list, built from the documents, with each object flagged as
  cluster-wide or not. The tool decides what to probe: A keeps asking only about namespaced
  objects and treats the volume as it does today. Raw vendor text cannot be listed (M knows
  only its kind), so B's operator probe list stays hand-written.
- **Other choice:** Leave each tool's list as it is. Nothing to build, and the drift stays.
- **Changes for:** Nobody outside the tools: the same probes, from one source.
- **Answer:** ok (2026-11-17).

### K16. The size limit of one applied object

- **About:** `kubectl apply` stores the whole object in an annotation, which allows about
  256 KiB. A checks this for its stores Secret and its ConfigMap (limit less 4096 bytes),
  and skips the check on a remove.
- **Recommended:** M offers the constant and a plain check of one rendered document. Each
  tool decides when to call it, as A already does.
- **Other choice:** Render itself refuses an oversized document. It would also block a
  `generate` of something that may be applied by other means.
- **Changes for:** Nobody. The tools' own checks call M's.
- **Answer:** Owner (2026-11-18): M enforces it. A deploy refuses an object over the limit
  (the apply method checks each rendered document); remove does not check; `generate`
  prints a warning naming the object and its size.

### K17. Volume sources other than NFS

- **About:** A writes an NFS PersistentVolume. The model also had a catch-all `Source` field
  for other volume kinds, refusing `hostPath`. The verifier notes `local` is a host
  directory like `hostPath` and needs node affinity the model does not have.
- **Recommended:** No catch-all field. NFS only, until a second tool needs another kind,
  which then gets a typed source. This keeps a host directory out of the module and fits "a
  package needs two consumers".
- **Other choice:** Keep the catch-all and refuse `hostPath` and `local`. A third tool can
  use a cluster storage driver today, and M must keep that refusal list complete.
- **Changes for:** Third tools that need CSI or similar volumes.
- **Answer:** ok (2026-11-17).

### K18. Kinds and fields nobody uses yet

- **About:** The model has kinds only B's vendor bundle contains (ServiceAccount, Role,
  ClusterRole, RoleBinding, ClusterRoleBinding) and fields neither tool writes: an HTTP
  probe, a startup probe, container `args`, service types NodePort and LoadBalancer, pod
  annotations and `serviceAccountName`. Today the bundle stays raw text (K9).
- **Recommended:** Build the fields, since a third tool will most likely need an HTTP probe,
  `args` and a service type, and each is a few lines. Do not build the RBAC kinds until the
  example tool or a real tool authors one; keep their design in the note. A wildcard `*`
  verb or resource is refused unless the tool sets a field that allows it. The CRD is never
  modelled.
- **Other choice:** Build everything the review modelled. One round of work and one set of
  tests, and the RBAC kinds have no user to prove them.
- **Changes for:** Nobody until a tool uses them. The coverage floor needs a test for every
  line.
- **Answer:** Owner (2026-11-17): build nothing ahead of need. No HTTP probe, `args` or
  Service type and no RBAC kinds until a tool needs one. The CRD is never modelled.
  Revised 2026-11-18: the Service types are built now. M checks ClusterIP, NodePort and
  LoadBalancer and requires one; the tools set the default (A ClusterIP, B LoadBalancer).

#### Settled by the rulings

- A bare word list for custom bodies (the verifier's correction of the word helper) follows
  ruling 9: B's `pullPolicy` and `updateStrategy` become quoted text, since the API receives
  the same string. M keeps bare words only inside its own documents.
- A stated writable root filesystem on Kubernetes writes no line: ruling 9. A states
  writable, so its output stays as it is. Read-only writes `readOnlyRootFilesystem: true`.
- The pod and the container refuse `envFrom` in the extras: ruling 9's deny list, in S1.
- Should a pod state `hostNetwork: false`? Ruling 13: Kubernetes gets no network key and the
  model has no host fields.
- Name grammar: ruling 13 says names are DNS labels, not 100 characters. Ruling 11: the
  instance is `kubernetes.name` and the namespace is `kubernetes.namespace`.
- Does `generate` stay, and is there a dry-run mode? Ruling 12: no preview mode, `generate`
  stays, and M's render is what both call.

#### No decision needed

- Metadata is written in the order name, namespace, labels, annotations for every kind. B's
  custom resource, which writes namespace first today, changes order in its golden.
- A list under a key is indented two further than the key, as compose already does. B's
  custom resource, which puts the dash at the key's indent, moves in its golden.
- The model takes Secret bytes and M encodes base64. Data keys are sorted by byte value, so
  A's credentials Secret keys change order.
- Documents are joined by one `---` line, blank ones dropped, one newline at the end. B's
  operator stream without credentials gains the final newline the vendor text lacks.
- An empty map or list is dropped in M's own documents. In a tool-built tree it is an error
  and the tool chooses. B's `spec` per node, which the operator's schema requires, is
  written as an explicit empty mapping by B.
- A mapping with the same key twice is refused. B's load-balancer annotations can repeat a
  key.
- Tolerations are one type: operator always stated (Equal when unset), effect left out when
  unset (B writes an empty one), an empty toleration refused. B's string form stays its
  config.
- The node selector is a map, sorted by key, quoted. B's config order and duplicates are
  lost.
- Quantities (cpu, memory, storage) are checked once against one grammar and written as
  text.
- M checks each kind's own name rule: Namespace a label, Service a letter-first label, the
  rest a subdomain. A digit-led name that loads in A today fails at its Service on a real
  apply, so A tightens the check at load for that name.
- A tool that holds numbers as text (B: `runAsUser`, `fsGroup`, `runAsGroup`,
  `containerPort`, `servicePort`) must pass them as numbers, or the operator's schema
  refuses them (K5).
- M states these defaults, as they are hardening and not product: Service type ClusterIP
  (revised 2026-11-18: M requires a type and the tool sets the default),
  `automountServiceAccountToken: false`, read-only on every mount, `defaultMode: 0400`, the
  volume reclaim policy Retain, the toleration operator. A already writes most of them.
  Replicas, ids, sizes and images are never defaulted by M.
- A remove stream refuses a Namespace document by its kind, raw text included, and refuses a
  raw document whose kind cannot be read (the existing rule: a namespace goes only through
  the namespace gate and its own delete).
- A `|` text block never uses the "keep" ending (`|+`): content with several trailing
  newlines is written as a quoted string, and the end of a block follows its content
  exactly, so a file with no final newline is no longer given one. The indentation mark is
  chosen from the first non-empty line.
- The lib annotation reads `go-deploy-v0.2.0` for a stamped build, as an example, and
  `go-deploy-devel` otherwise, with the leading `v` of the stamped version not doubled. The
  exact text and the strip rule are fixed at build; no file states them yet.
- M exports its reserved-key check so B applies it to `kubernetes.podLabels`: M cannot see
  inside a custom body.
- Tool-owned label and annotation keys (A's `solace-connector/role`, B's
  `solace-util.sha256`) pass through unchanged. When an image-pull Secret is built, and
  which defaults live where, stay with each tool.
- The namespace document is labelled in `generate` and in a deploy (decided 2026-11-15). The
  vendor's own Namespace stays raw (K9).

## Exported names

The names that M exports: the functions, types and constants that a tool can call. The
questions ask whether a name stays public (a tool can call it), goes private (only M's own
code can) or moves. Seven groups by package, widest effect first. A few entries use Go
terms: `errors.Is` and `errors.As` are Go's two ways to test which error came back, and a
"zero value" is what a field holds when nobody set it.

### Group: runner, exit codes, character rules and command guards

### N1. Who marks a bad input as exit code 2

- **About:** A tool exits 2 when the operator typed something wrong and 1 when something ran
  and failed. Today M never marks an error as 2. Both tools do it themselves (A
  `runner.go:108`, B `cli/platform.go:188`). Example: `docker.command: sudo docker` is
  refused by `execguard.Check`, and the tool turns that refusal into exit 2. The new name
  checks of gdu 4 raise the same question for every builder.
- **Recommended:** M marks exit 2 only on errors that can come from nothing but what the
  operator typed: the `execguard` refusals and the `envref` errors. Builder name errors stay
  unmarked, because only the tool knows whether a name came from the env file or was read
  from the cluster. Rewrite the two sentences that say "the caller marks" (exitcode doc,
  developer.md "A tool marks usage errors") to say this.
- **Other choice:** M marks nothing, as today. Every new tool repeats the wrapping, and A
  keeps matching the guard's message text in a test.
- **Other choice:** M marks every refusal, builder names included. No wrapping anywhere, but
  a name read from the cluster that fails the check would exit 2 where 1 is right.
- **Changes for:** nobody's exit codes today. Both tools drop their wrappers around guard
  and env-reference errors. A's `errors.As` on `UnsetError` still works, because the mark
  unwraps.
- **Answer:** ok (2026-11-18).

### N2. The three capture helpers `Combined`, `Split`, `Output`

- **About:** These run a command straight on the `Runner`, not through a client, and capture
  its output as text. A has 20 call sites for `Combined` and 5 for `Split`. B's
  `engine.Output` sits on `Output`. Rulings 1 and 2 give every client `Run` and `Output`,
  with the tool's writers, so a tool-built command now goes through the client and gets the
  base command, the pinned context and the typed errors. The helpers over a bare `Runner`
  would skip all of that.
- **Recommended:** Make all three private once the clients exist, and move the tools'
  remaining uses to `client.Run` and `client.Output` (a buffer as the call's writer replaces
  `Combined` and `Split`). Reason: ruling 2 says the client is the way to run a command the
  module does not offer. M's own tests keep using them from inside the package.
- **Other choice:** Keep them exported. Tools can still bypass the client, which is the gap
  ruling 2 closes, and two ways to capture output stay to explain in the guide.
- **Changes for:** A (about 25 sites) and B (one wrapper) move to the client; no change for
  operators. Not decided by the reviewers: they kept all three before rulings 1 and 2.
- **Answer:** ok (2026-11-18).

### N3. What exit code means when a child is killed by a signal

- **About:** A child killed by a signal has no exit number; Go gives -1. Today `Exec.Attach`
  turns that into 1. The design's `Result.ExitCode` uses -1 for "never started". With both
  meanings, `exitcode.Exited(name, -1)` would word a start failure as "exit status -1".
- **Recommended:** Once a command has run, `ExitCode` is never negative: a signalled child
  reports 1, as `Attach` does now. -1 means only "never started". Reason: `exitcode` assumes
  a non-negative status, and ruling 5 puts the code in the result for all three methods.
- **Other choice:** -1 for both, and the tool folds it. Every tool repeats the rule.
- **Changes for:** nobody today (same outcome as now). The doc of `Result` states it.
- **Answer:** ok (2026-11-18).

### N4. `exitcode.ChildStatus` stays or goes private

- **About:** `ChildStatus` reads whether an error carries a child's own exit code. Only
  `exitcode.Code` calls it in production; B's tests call it twice. B prints
  `Error: kubectl: exit status 7` and then exits 7. A prints nothing and returns the code.
- **Recommended:** Keep it exported. Reason (verifier): it is the only way to read the exit
  code that an error carries, so hiding it means a tool cannot tell a passed-through exit 1
  from its own failure. Ruling 5 makes a non-zero `Exec` an error, and `ChildStatus` reads
  that error's exit code unchanged. Its doc states that an error with an `ExitCode()` method
  is part of the contract.
- **Other choice:** Make it private (the reviewer). B's two test assertions go through
  `Code`, which cannot tell exit 1 from "no status".
- **Changes for:** nobody. B can close its print-then-exit difference with A later.
- **Answer:** ok (2026-11-18).

### N5. What `Result` holds when one operation runs two commands

- **About:** `Result` has one `Argv` and one `ExitCode`. A podman secret create below
  version 4.7 runs the version read, `secret rm --ignore`, then `secret create`. Also,
  `Argv[0]` is the name as built, not the path found on PATH, because the program lookup
  happens inside the runner, after the command is built.
- **Recommended:** `Result` describes the last command, the one whose output the method
  returns. The method's doc lists the commands it may run. `Exec.Announce` stays as the
  per-command trail, and `Argv[0]` is documented as the name as built. Reason: the simplest
  shape, and the trail exists already.
- **Other choice:** A list of `Result`, one per command: every caller indexes into it. Or a
  callback per command, which brings back the hook ruling 1 removed.
- **Changes for:** tools reading `Result` see the last command only. B's `--verbose` trail
  is unchanged.
- **Answer:** ok (2026-11-18).

### N6. `runner.DefaultGrace` and `Exec.Grace`

- **About:** After a command exits, `Exec` waits two seconds for its output to drain before
  giving up. `Exec.Grace` changes that wait; zero means `DefaultGrace`. No tool sets it.
  Only M's tests do. With ruling 1, a call given a buffer as its writer is piped, so the
  wait applies there. A call given the real stdout and stderr stays direct.
- **Recommended:** Keep both. Reason: the constant lets a third tool write "the default
  times 3", and M's other knobs follow the same pattern (`kube.RolloutTimeout`,
  `watch.DefaultInterval`). Piped progress output for compose and pull is not measured, so
  add a smoke row.
- **Other choice:** Make the constant private and state two seconds in the field doc. Or
  hide both. Smaller surface, and a third tool cannot name the default.
- **Changes for:** nobody.
- **Answer:** ok (2026-11-18).

### N7. One text form for a whole command (`Quote`, `MaskEnv`, `Call.String`, `Announce`)

- **About:** A command is shown as text in four places: the `Echo` line, `Result.Argv`, a
  test failure message (`Call.String`, which `fmt` uses without anyone calling it) and B's
  `--verbose` trail (`Exec.Announce`, which is given only the path and arguments, so it
  cannot show the stdin size or env keys). Example form: `+ kubectl apply -f - (stdin 1204
  bytes, env TOKEN=***)`. Also `runner.Quote` (display quoting) now shares its name with
  `compose.Quote` (YAML quoting).
- **Recommended:** One `Cmd` method gives the masked display form, and `Echo`, `Call.String`
  and `Announce` all use it (`Announce` takes the `Cmd`). Keep `MaskEnv` exported: B uses it
  on a bare env list, and hiding it pushes tools to hand-roll secret masking. Keep
  `Call.String` (verifier: removing it would dump raw stdin and env values into failure
  messages). Rename `runner.Quote` to `DisplayQuote`. The reviewers left both to you, but
  the verifier argued for keeping both.
- **Other choice:** Leave the four forms and both `Quote` names. Nothing changes, and a
  third tool meets two functions with the same name and different rules.
- **Changes for:** B's `--verbose` trail can show env keys and stdin size. B's 5 uses of
  `Quote` and `runnertest` rename. Test failure messages change text.
- **Answer:** ok (2026-11-18).

### N8. Should M offer a wrapper that refuses a command carrying environment variables

- **About:** B wraps its runner (`NoEnv`) so only compose, which needs credentials in its
  environment, may pass `Cmd.Env`. A has no such rule and adds session variables to every
  systemd call. The reviewers proposed moving `NoEnv` into M. The parity body lists the env
  policy as each tool's own.
- **Recommended:** Leave it in B. The systemd client sets the session variables itself (A's
  pattern), so a general "refuse env" wrapper in M would clash with it. B's table row
  changes from "compose only" to "compose and the user session". `IsEcho` is not needed
  (ruling 12).
- **Other choice:** `runner.NoEnv(r)` in M, about 30 lines, wrapping the runner given to
  `kube.Client` only. B deletes its copy; the parity body text changes.
- **Changes for:** B only. Operators see nothing.
- **Answer:** ok (2026-11-18).

### N9. Which program names each command family may run

- **About:** Each tool keeps a list of programs allowed as the first word of a command: both
  allow `kubectl`, `oc` and `podman`. For docker, A allows `docker` and B allows `docker`,
  `docker-compose` and `nerdctl`. A comment in B says `docker-compose` is for a host with
  only the standalone binary. But M's compose secrets need Docker Compose 2.23.1 or later.
- **Recommended:** M states, beside each client, the programs whose commands that client
  issues (kube: `kubectl`, `oc`; docker: `docker`, `nerdctl`, and `docker-compose` for
  compose; podman: `podman`). A tool picks its own subset for `Rules.Allowed`, so A stays at
  `docker`. The doc names the 2.23.1 floor. Reason: only M knows what its builders and
  renderers need. Not checked on a real engine.
- **Other choice:** Lists stay tool customisation, as today. They can silently drift from
  what M runs; this one already has.
- **Changes for:** nobody's env file. B must decide whether `docker-compose` still belongs.
- **Answer:** Reshaped by E24 (2026-11-18): each client states the programs it issues, and the
  env file's `cli` words pick from them (`kubectl` or `oc`; `docker` or `nerdctl`; compose as
  plugin or standalone `docker-compose`; `podman`). No per-tool `Rules.Allowed` list.

### N10. The two flag names M fixes: `--allow-command` and `--no-prompt`

- **About:** `execguard.ApproveFlag` is `"--allow-command"` with the dashes. Both tools
  register the bare name `allow-command` with their flag library, so each keeps its own copy
  of the text. `confirm.DefaultNoPromptFlag` has the same problem, plus an
  `Options.NoPromptFlag` override that no tool sets. CR-0001 already decided both tools take
  the approval flag name from M.
- **Recommended:** Export both spellings for both flags: `ApproveName = "allow-command"` and
  `ApproveFlag = "--" + ApproveName`, and the same pair for no-prompt. Drop the `Default`
  prefix and `Options.NoPromptFlag` (if N18 lands, the per-call words replace it). The guide
  says the two flag names are fixed by M. Reason: both tools need both spellings (dashed in
  help text, bare at registration).
- **Other choice:** Keep `DefaultNoPromptFlag` and the override, so a third tool may rename
  the flag. Two answers to who owns the name, and the verifier found no tool that wants it.
- **Changes for:** nobody's command line. Tools stop holding the copies (A in three places,
  B in three).
- **Answer:** Reshaped by E24 (2026-11-18): `--allow-command` is gone with the free-form
  command keys. M exports only the no-prompt pair, `NoPromptName` (`no-prompt`) and
  `NoPromptFlag` (`--no-prompt`).

### N11. One exported check for a plain value (`execguard` charset check)

- **About:** A value that ends up bare in a command or artifact (a timezone, a java option,
  a name) must be valid UTF-8, with no control, space or invisible character, and none of
  `UnsafeChars`. M has this only as the private `checkToken`. A wrote `SafeToken`, B wrote
  `checkTimezone`, and B copied the private whitespace test. The three disagree on the
  tilde: M refuses a leading one only, A refuses one anywhere, B has no rule. The renderers
  of ruling 9 also need this check for every value they write.
- **Recommended:** Export `CheckValue(field, value) error`. Keep `IsControl` and
  `IsInvisible` exported and export the whitespace test too, because B's probe check and A's
  java-option check deliberately allow spaces and use the pieces. Tilde rule: refuse a
  leading tilde only, as M does. Ruling 9 quotes every value, so a bare `~` is harmless.
  Names on a command line get typed grammars from gdu 4 and do not use this.
- **Other choice:** Refuse a tilde anywhere (A's rule). Stricter, and B's timezone check
  tightens.
- **Changes for:** A and B drop their copies. A accepts an embedded `~` it refused before.
  No change to B's operators beyond that.
- **Answer:** ok (2026-11-18).

### N12. Where the character rules live

- **About:** `IsControl`, `IsInvisible`, `UnsafeChars` and the private whitespace test are a
  character rule set, but sit in `execguard`, whose doc says it decides whether a command
  from config may run. After the decided work, `runner`, `containercli/inspect`
  (`Printable`), the tilde package, the free-operand check of gdu 4 and the renderers all
  import it. M also holds two other private definitions: `runner`'s own printable test, and
  `quadlet.Escape`'s test of `c < 0x20`. `inspect.Printable` uses `unicode.IsControl`, which
  differs from `execguard.IsControl` on U+0085.
- **Recommended:** A small neutral package holds the character tests (`IsControl` and the
  like), `UnsafeChars` and `CheckValue`. `execguard`, `runner`, `inspect` and the renderers
  import it, and the two private copies go. Reason: five importers of a command guard just
  for character tests is the wrong shape, and one definition gives `Printable` the same set.
- **Other choice:** Leave it all in `execguard`. No move, but renderers import the guard for
  characters. If N11 lands, `UnsafeChars` has no tool user and could become internal.
- **Changes for:** both tools change an import path for `IsControl` and `IsInvisible` (A 2
  uses, B 3, for each of the two). Nothing for operators.
- **Answer:** ok (2026-11-18).

### N13. `IsControl` and U+0085 (NEL)

- **About:** `IsControl` says it covers C0, DEL and C1 controls, but leaves out U+0085.
  Inside `checkToken` that is harmless, because the whitespace test catches it. In A's java
  options and B's probe tokens, spaces are allowed and no whitespace test follows, so a NEL
  passes into the artifact and the logs. The review lists this as a defect.
- **Recommended:** `IsControl` includes U+0085. `checkToken` tests whitespace first for that
  one character so its message stays clear. Reason: a public test function should not carry
  a hole that only one caller covers.
- **Other choice:** Leave the documented hole. `inspect.Printable` then cannot share the
  definition (N12) without letting NEL through.
- **Changes for:** operators: an env file with a NEL in a java option or probe token is now
  refused. The M test `TestExportedCharset` and B's `hostpath_test.go:100` change.
- **Answer:** ok (2026-11-18).

### N14. The host-path check for a mount source

- **About:** A path that lands in a quadlet `Volume=` line or a compose volume (`/srv/data`
  in `dataDir`) is checked by both tools with the same character set. They differ on two
  points. B refuses a leading `~` because it expands tildes first; A permits it and has no
  expansion yet. A also permits `:`, while gdu 4's `Volume=` builder refuses a colon in the
  source (a colon splits the fields there).
- **Recommended:** One exported `CheckHostPath(field, path) error`, callable at load before
  any rendering, kept beside the new `tilde` package of gdu 2, using the shared character
  check (N11). Refuse the leading `~` there. Colon and `%` are decided per platform by each
  renderer (ruling 9): quadlet refuses a colon and escapes `%`; compose may accept a colon
  only if it writes the long volume syntax. Not checked on a real engine. `execguard` does
  not take it, as its doc scopes it to commands.
- **Other choice:** Leave the check in each tool. The two rules keep drifting, and a third
  tool writes a third copy.
- **Changes for:** A, which gets tilde expansion in the same tag (ruling 8) and then refuses
  an unexpanded `~`. A's validate keeps refusing at load, not at deploy.
- **Answer:** ok (2026-11-18).

### N15. `envref.ValidName`

- **About:** The grammar of an environment variable name (`TOKEN_1`, not `1TOKEN`). Only M
  uses it today (inside `Ref.Check`). B keeps a regex with the same rule (`validate.go:398`,
  I compared them: the same). gdu 4 needs it for compose secrets and the Deployment
  renderer. It shares its bare name with `containercli.ValidName`, which is a different
  grammar.
- **Recommended:** Keep it exported and keep its name. The package name already says "env".
  B drops its regex and calls it. If renderers importing `envref` just for this bothers you,
  it moves with N12's package.
- **Other choice:** Rename to `ValidEnvName`. Clearer on its own, but part of a naming
  scheme for all the grammars that gdu 4 adds, which is not this part's to settle.
- **Changes for:** B only (one regex gone).
- **Answer:** ok (2026-11-18).

### N16. One home for the terminal facts, and the Windows console switch for attached sessions

- **About:** M holds three terminal facts in two packages: "is this a terminal"
  (`confirm.IsCharDevice`), the same test copied inside `watch.CanRedraw`, and the Windows
  switch that makes a console understand colour and cursor codes (`enableVT`, private in
  `watch`). A has its own copy of the switch for its `shell` command. B has none, so its
  attached session on a plain Windows console is not covered.
- **Recommended:** One small package holding the terminal test, the console switch and
  `CanRedraw` built on both. `Exec.Attach` (ruling 5) turns the switch on. Reason: the
  attached session is exactly where B's gap is, and `watch` is the wrong place for the
  interactive clients to import. Not checked on a real Windows console.
- **Other choice:** Leave `IsCharDevice` in `confirm`, accept the copy in `watch`, and move
  only the switch to an internal package. One package fewer, and the copy stays.
- **Changes for:** A deletes its copy of the switch. B gains working colour in an attached
  session on Windows. Names `CanRedraw` and `IsCharDevice` may change package.
- **Answer:** ok (2026-11-18).

### N17. Keeping the tool alive on Ctrl-C while a shell is attached

- **About:** When an operator presses Ctrl-C in `kubectl exec -it` or `docker exec -it`, the
  terminal sends it to the tool and to the shell. A installs a catcher for the life of the
  session so the tool survives (`shell.go:268`). B has none (its only signal handling is the
  watch loop). A's own comment says the catcher mostly matters before the engine puts the
  terminal in raw mode, and to let the engine restore the console.
- **Recommended:** `Exec.Attach` installs the catcher when the session starts and removes it
  when it ends. It is not a global variable, only a registration for the length of one call.
  Reason: B lacks it and any third tool would too. Not checked on a real engine or terminal.
- **Other choice:** Leave it a tool step, as the developer.md rule against process-wide
  state can be read. B's gap stays or B copies A's function.
- **Changes for:** A deletes its function. B stops being killed by Ctrl-C before the engine
  has restored the console. The runner doc ("the terminal owns the session") gains one line.
- **Answer:** ok (2026-11-18).

### Group: confirm, abbrev, watch and platform words

### N18. Asking a yes/no question that a flag may answer (`confirm.Gate`)

- **About:** `Gate` takes a flag value (`noPrompt`). If it is true, `Gate` says yes without
  asking. If there is no terminal, its refusal says "pass --no-prompt". The namespace
  question of gdu 1 needs different words: `--no-prompt` alone must never delete a
  namespace, only `--no-prompt --delete-namespace` does, and the refusal must name both.
  `Gate` cannot say that, so three places already rebuild it by hand (A `status.go:1007`, A
  `main.go:1151`, B `confirm.go:148`), and the two namespace sites cannot use it as it
  stands.
- **Recommended:** Change `Gate` so the tool computes "approved" itself and passes the words
  for its own alternative, for example `"--no-prompt --delete-namespace"`. `Ask` and `Gate`
  take their arguments in the same order, and `Options.NoPromptFlag` goes. The step that
  turns an unanswered prompt into an error moves inside, so `Unanswered` becomes private.
  Reason: the verifier showed `Gate` cannot word the decided question as it stands, so a
  reorder is not enough.
- **Other choice:** Leave `Gate` and add a second method beside it (the reviewers'
  `Require`). Less churn, but two near-identical methods, and `Gate` keeps a parameter name
  that invites passing the flag unchecked.
- **Changes for:** both tools at about a dozen call sites. Operators see the same prompts.
  The namespace refusal now names both flags. Optional in the same change: `Ask` and `Gate`
  append the `[y/N]` hint themselves, so it cannot disagree with the form (every call writes
  it by hand today).
- **Answer:** ok (2026-11-18): the owner took every remaining N recommendation.

### N19. How `confirm` reports "no terminal" and "no answer"

- **About:** `NotTerminalError` carries no data, so `errors.As` needs an awkward target and
  comparing values is unspecified in Go. "Unanswered" is a function that makes an error, and
  its one test checks `io.EOF`. The design's example branches on a plain value
  (`kube.ErrForbidden`). Also "unanswered" now means two things in M: end of input at a
  prompt, and (ruling 4) a read that returned nothing.
- **Recommended:** Two plain error values with the same text, for example `ErrNotTerminal`
  and `ErrNoReply`, replacing `NotTerminalError` and `Unanswered()`. Pick the second name so
  it is not "unanswered". Reason (verifier): same call-site cost, and it matches how clients
  report. B's tests that pin the text still hold.
- **Other choice:** Keep the type and the function. No change, two styles in one package.
- **Changes for:** neither tool names the type. A and B's hand-written folds disappear if
  N18 lands.
- **Answer:** ok (2026-11-18): the owner took every remaining N recommendation.

### N20. A numbered menu question and `Prompter.ReadLine`

- **About:** Both tools write the same menu by hand: print `1) docker 2) podman`, read a
  line, convert to a number, range-check (A `main.go:1144`, B `platform.go:268`). B prints
  the list straight to the process stderr, so its tests cannot capture it. Once the menu and
  the typed name question (decided in gdu 1, `AskName`) are in M, neither tool calls
  `ReadLine` any more.
- **Recommended:** Add `Choose(heading, choices) (index, err)` on `Prompter`, writing
  through its `Out`, with its own error for an out-of-range reply. Keep `ReadLine` on
  purpose as the free-text primitive for a third tool, and say so in the guide.
- **Other choice:** Leave the menus in the tools and unexport `ReadLine` with `Line`.
- **Changes for:** both tools drop about 20 lines. Operators see the same menu.
- **Answer:** ok (2026-11-18): the owner took every remaining N recommendation.

### N21. Platform words (`kubernetes`, `docker`, `podman`) and their short forms

- **About:** Both tools declare the same three words and the same short forms (`kube`, `dk`,
  `pm`). The reviewers asked whether a small M package should export them. M already exports
  `containercli.Engine` with `docker` and `podman`.
- **Recommended:** No new package. Each tool keeps its few declaration lines. Reason: the
  `abbrev` doc says it owns the rules, not the vocabulary, and the word doubles as the
  env-file section key (`kubernetes.namespace`, ruling 11), which M must not hard-code. The
  new-tool guide shows the declaration.
- **Other choice:** A `platform` package beside `abbrev` that reuses `containercli.Engine`.
  A third tool declares nothing, but the spelling becomes a contract in M.
- **Changes for:** nobody if no. If yes, the short forms are fixed for all tools.
- **Answer:** ok (2026-11-18): the owner took every remaining N recommendation.

### N22. `abbrev`: error text, reference table, and `Set.Kind`

- **About:** Each tool words "unknown value" for a set in its own way, three shapes between
  them. B alone generates a reference table of the short forms in a test. `Set.Kind` is read
  only by `New`'s own panic messages. The reviewers listed removing `Kind` unless M adds an
  error helper.
- **Recommended:** Add `Resolve(word) (string, error)` returning `invalid <kind> "<word>":
  expected <list>`, which gives `Kind` a reader. Skip the Markdown table until a second tool
  generates one. Reason: wording the same fact three ways is a parity gap between the tools.
- **Other choice:** Remove `Kind` and leave all wording to the tools. No change, and A keeps
  building its list by hand beside the set.
- **Changes for:** operators of A see the new wording of the platform error. B's wording
  changes only if it adopts `Resolve`.
- **Answer:** ok (2026-11-18): the owner took every remaining N recommendation.

### N23. `watch.CheckInterval`

- **About:** `--watch 5` is checked against `MinInterval` and `MaxInterval`. Both tools
  check it themselves with their own flag wording, because `CheckInterval`'s error names no
  flag and is not marked exit 2. Neither tool calls it. The loop calls it as a second line
  of defence.
- **Recommended:** Make it private. Tools keep their own wording. The loop's own refusal of
  a bad `Every` becomes a usage error (needed anyway by gdu 13, and ruling 13 says A is
  affected).
- **Other choice:** Keep it exported, return an exit-2 error that carries the range, and
  have each tool wrap it with its flag name.
- **Changes for:** nobody; tool code and operator output stay the same.
- **Answer:** ok (2026-11-18): the owner took every remaining N recommendation.

### N24. `watch.Sleep`

- **About:** A pause that stops when the context is cancelled. Only `watch` uses it. B has
  its own copy for its poll loop (`broker.go:94`).
- **Recommended:** Make it private. Reason: it is a ten-line helper, and moving it would
  need a package for one function. `Options.Sleep` stops sharing a name with it.
- **Other choice:** Export it in a general package so B's copy goes. A new package for ten
  lines.
- **Changes for:** nobody. Three mentions to fix: the watch row of the adoption table in
  docs/developer.md, the `Options.Sleep` doc, A's comment at `status.go:1061`.
- **Answer:** ok (2026-11-18): the owner took every remaining N recommendation.

### Group: containercli, inspect, the compose and quadlet helpers and artifacttest

### N25. Do the compose and quadlet helpers and the hardening constants stay public?

- **About:** M hands tools small text helpers for the compose file and the quadlet unit.
  `compose.Quote` writes a value as a quoted compose scalar and doubles `$`, so compose does
  not expand it (A calls it 18 times, B 9). `compose.Escape` is the `$` doubling alone.
  `compose.FlowList`, `compose.ServiceSecrets` and `compose.TopSecrets` write a list line
  and the two secrets blocks. `quadlet.Escape`, `quadlet.EscapePercent`,
  `quadlet.Environment` and `quadlet.Secret` write the quoting and the `Environment=` and
  `Secret=` lines. Also `compose.Posture` and `quadlet.Posture`: the hardening lines
  (privileged off, capabilities dropped, no new privileges, non-root) as text. Each tool
  writes them once today. The only other reader is `artifacttest`; neither tool names them
  in a test, both use `AssertComposePosture` and `AssertQuadletPosture`. Once M writes the
  whole file (design 8.2), every call sits on a line M writes itself.
- **Recommended:** Make all the line helpers private. The shared writer package of design
  8.2 exports two quotes instead, a plain one and one that doubles `$`. A tool needs them
  for text it still writes itself, such as B's broker custom resource. M needs both too: A's
  Kubernetes Deployment doubles `$` in environment values, so `$` doubling is not only
  compose's. Move both `Posture` constants to an `internal/` package that the renderers and
  `artifacttest` read, so no tool can hand-write a service with them. Go needs them exported
  only for visibility across packages, and `internal/` solves that. The move also removes
  the `artifacttest` import of both renderers. `compose.Secret` becomes the shared secret
  entry (ruling 9); `EnvVar` stays compose-only.
- **Other choice:** Keep `Quote`, `Escape`, `Posture` and `Secret` public, and make only
  `FlowList`, `ServiceSecrets` and `TopSecrets` private. Then a tool can still hand-write a
  line or a service, which a typed renderer is meant to stop. Or the other direction: make
  everything but `File` private, a smaller surface, but each tool's tests need their own
  quoting copy.
- **Changes for:** both tools' source (the 27 `Quote` calls move into the renderers). Not
  env files and not running deployments. Goldens change through the renderer, not through
  this answer. `TestPostureBytes` moves with the constants.
- **Answer:** ok (2026-11-18).

### N26. Where the inspect types and `Decode` live

- **About:** `inspect.Container` is one container as the engine reports it (name, image,
  labels, state, mounts). `State`, `Health` and `Mount` are its parts. `Decode` turns the
  engine's JSON into containers and never reads the container's environment. `Containers`
  builds the `inspect` command. A names `inspect.Container` twice, B ten times. Once
  `Client.Inspect` runs the command, `Containers` goes private. The open point is where the
  rest sits.
- **Recommended:** Keep package `containercli/inspect` as the plain decode package:
  `Container`, `State`, `Health`, `Mount` and an exported `Decode`. Only the command builder
  moves into `containercli` and goes private. The verifier's reason: A's status tests and
  B's `decodeOne` parse engine JSON with no engine in reach, and the "inspect: " messages
  stay true. The reviewer's fold into `containercli` would also force renames, since a bare
  `State` or `Health` says nothing.
- **Other choice:** Fold it all into `containercli` and make `Decode` private. `State`,
  `Health` and `Mount` need new names, both tools build `Container` values by hand in their
  tests instead of using engine JSON, and the error prefix changes in message tests.
- **Changes for:** with the recommendation, no line in either tool changes for these names.
  With the other choice, A's `statusreport` and B's `internal/container/inspect.go` change.
- **Answer:** ok (2026-11-18): the owner took every remaining N recommendation.

### N27. Which package holds `Printable`

- **About:** `inspect.Printable` strips control and invisible characters from text the
  engine sent, so a terminal escape inside a container name cannot act on the operator's
  terminal. A calls it 12 times, on whole command output. B calls it 11 times, on listing
  rows. It sits in `inspect` only by history. A second form is missing: both tools clean
  multi-line output and want the line breaks kept. (Which characters it drops is N12 and
  N13; this entry is only the home.)
- **Recommended:** Move it to `runner` as `runner.Printable`, with a second form that keeps
  newline and tab. `runner` is the one package that `containercli`, `kube`, `podman` and
  `systemd` all import, and kubectl text has the same exposure. `runner` already has an
  unexported helper with that name and other rules; rename that one.
- **Other choice:** Put it in `containercli`. `kube` would then import `containercli` for a
  text helper, or kubectl output (A copies kubelet reasons unfiltered today) stays
  unfiltered.
- **Changes for:** A and B change an import path on 23 calls. Nothing else.
- **Answer:** ok (2026-11-18): the owner took every remaining N recommendation.

### N28. Is the stderr in an error cleaned or exact

- **About:** Ruling 4 says an error carries the command's stderr as a field. Stderr is text
  from outside, like a container name. A cleans every engine error text today (`safeOut`).
  Example: an engine message holds an escape sequence that rewrites the terminal title, and
  the tool prints the error.
- **Recommended:** The field holds the exact bytes, so a tool can match on them and a test
  can compare them. The text from `Error()` uses a cleaned copy (`runner.Printable`, N27). A
  tool that prints the field itself cleans it. Reason: M cleans what M prints and never
  changes what the tool chooses to look at.
- **Other choice:** Clean the field too. A tool can never print it raw by mistake, but the
  field no longer shows what the engine said, which matters because docker and podman word
  "not found" differently.
- **Changes for:** A's `safeOut` goes. B's `printableRow` stays for listing rows.
- **Answer:** ok (2026-11-18): the owner took every remaining N recommendation.

### N29. What the client says when a container does not exist

- **About:** Asking for a missing container fails the run in both tools today. A's `status`
  asks about several containers in one `inspect` call, so one missing name fails the whole
  status. The reviewer expected exit 0 with `[]`. The verifier read both tools and showed
  the run error comes first, so that case never arrives. Ruling 4 gives the typed "not
  found" and "partial listing" errors; this asks when each is used. The exit code and stdout
  of a missing name on docker and podman are not checked on a real engine.
- **Recommended:** `Client.Inspect` returns "not found" when the only name asked is missing.
  When several were asked, it returns the containers found and the "partial listing" error
  that names the missing ones, and the tool decides (A's status can print the rest).
  `Start`, `Stop` and `Restart` return the plain failure with the engine's stderr and do not
  guess "not found", because the engines word it differently. Add a docs/smoke.md row for
  the missing-name exit code before this is trusted.
- **Other choice:** Inspect returns an error and nothing else when any name is missing (A's
  status keeps failing as a whole). Or classify `Start`, `Stop` and `Restart` too, which
  needs engine-text matching per engine.
- **Changes for:** A's `status` with several containers. B's deployed-digest read, which
  takes a missing container as "changed" today.
- **Answer:** ok (2026-11-18): the owner took every remaining N recommendation.

### N30. One container-name grammar or two

- **About:** `containercli.ValidName` accepts what docker and podman accept,
  `[A-Za-z0-9][A-Za-z0-9_.-]*`. Tools run it on names read back from the engine (`--all`,
  `status`) before putting them on a command line. B also keeps a regexp copy for its env
  file. Ruling 11 says a name a tool deploys is a DNS label of at most 63 characters
  (`solace-1`, not `Solace_1`). So there are two questions: which names we create, and which
  names we find. Example: a container `my_app` made by hand carries the tool's image.
- **Recommended:** Keep two checks. `ValidName` stays the engine's grammar, for names read
  back and for the name argument of every client method, so a tool can still stop `my_app`.
  A new `CheckName(name) error` words the DNS-label rule for names a tool will create, and
  the renderers call it. B's regexp copy goes. Ruling 11 changes only the second.
- **Other choice:** One grammar (the DNS label) everywhere. Simpler, but `--all` and
  `status` then skip, with a note, every container whose name has a capital, `_` or `.`,
  including B's containers made under the old rule.
- **Changes for:** B's env file (the name rule itself is ruling 11). Discovery of old
  containers stays as it is with the recommendation.
- **Answer:** ok (2026-11-18): the owner took every remaining N recommendation.

### N31. Finding the tool's own containers by label

- **About:** gdu 1 puts ownership labels on every compose service and quadlet unit. Both
  tools still find their containers by image text (`ByImage` with their own marker). `List`
  shows no labels. Nothing decides whether discovery moves to labels. Example: `docker ps
  --filter label=services.solace.com/tool=solace-util`. Containers made before the labels
  exist do not carry them until the next deploy recreates them. Not checked on a real
  engine.
- **Recommended:** Add a label filter to the options of `Client.List` (the engine filters,
  so no parsing of the label column). `Row` does not gain labels. `ByImage` stays for
  containers older than the labels. Reason: label discovery is the point of gdu 1, and a
  filter avoids parsing a comma-separated label column whose values may hold commas.
- **Other choice:** `Row` gains a `Labels` map and tools filter in Go. More parsing and a
  quoting risk. Or leave discovery on image text, which keeps matching a foreign container
  that runs the same image.
- **Changes for:** `--all` and `status` in both tools, once they adopt it. After adoption a
  container made by an older tool version is found only through the image-text fallback.
- **Answer:** ok (2026-11-18): the owner took every remaining N recommendation.

### N32. Log options: `Tail`, `Since`, and who checks them

- **About:** `containercli.LogOptions` and `kube.LogOptions` pass `Tail` and `Since` as
  text, as given. gdu 14 decided `--tail` is a count only, and left open whether `Logs`
  refuses a bad value itself and whether `Since` gets the same check. `Since` is a time such
  as `5m`. The tools differ on `--since`: A refuses 0 and below, B refuses only below 0, so
  B lets `--since 0s` through.
- **Recommended:** `Tail` becomes a count and `Since` a `time.Duration`; M writes the engine
  or kubectl spelling. `Logs` refuses a bad value itself, in both clients, with one rule. A
  zero or negative `Since` is refused (`0s` reads nothing useful), which is A's rule. A's
  100000 cap on `--tail` stays A's own. The `--tail` unset value needs a name, since 0 is a
  real request. Reason: a tool that forgets its check still cannot pass a bad value, as gdu
  4 wants for names.
- **Other choice:** `Since` stays text. `Logs` refuses a `Tail` that is not a count and a
  `Since` with a leading `-` or a control character, as for any free operand (gdu 4), and
  does not parse the time: kubectl rejects a bad one with its own error. Or keep text and
  one shared check function that the tools call at their boundary (the first scope line of
  gdu 14); every tool must then remember to call it.
- **Changes for:** operators of B: `--since 0s` is refused. A's flags are unchanged.
- **Answer:** ok (2026-11-18).

### N33. The refusal for `--previous` on docker and podman

- **About:** `NoPrevious(engine)` is the one shared refusal: "--previous is a kubernetes
  option, docker keeps no previous container's log". A calls it twice, B once, and each tool
  also keeps its own second check before calling it. The text hard-codes `--previous`, which
  both tools happen to spell the same. `kube.LogOptions` has a `Previous` field that its
  `Logs` turns into `-p`.
- **Recommended:** Give `containercli.LogOptions` a `Previous` field like kube's.
  `Client.Logs` refuses it with this message, `NoPrevious` goes private, and the second
  check in each tool goes away. The message names the option by meaning, not by one tool's
  flag spelling.
- **Other choice:** Keep `NoPrevious` exported and take the flag name as an argument, as
  `execguard.ApproveFlag` does. Both tools keep their own second check.
- **Changes for:** both tools drop a guard. Their message tests may change text.
- **Answer:** ok (2026-11-18): the owner took every remaining N recommendation.

### N34. Small helpers both tools write twice in containercli

- **About:** Several small things are hand-written in both tools. (1) Splitting a listing
  into names safe to hand back to the engine and rows skipped (A `instances.go`, B
  `manager.go`). (2) The engine's display name (B keeps `platformTitle` because
  `Engine.title` is private). (3) Picking one container by exact name from a decoded answer
  (B `Status`, A `DockerDeployDigest`; B's `deployedDigest` takes the first result without
  matching the name). (4) "Is it running": list, parse, match. (5) The compose digest and
  then appending `VAR=digest` to the child's environment.
- **Recommended:** Add (1) `SafeNames(rows)`, (2) `Engine.Title()`, (3) a find-by-name over
  rows and over decoded containers, and (5) one helper that returns the digest and the
  environment entry together, so the variable name cannot drift. Do not add (4): `List` and
  `HasName` already do it, and a third read method adds a name for two lines.
- **Other choice:** Add nothing. Each tool keeps its loop, and B's missing name match stays
  its own to fix.
- **Changes for:** small source deletions in both tools. Nothing for operators.
- **Answer:** ok (2026-11-18): the owner took every remaining N recommendation.

### N35. The `Check*` twins in `artifacttest` and `noticetest`

- **About:** `artifacttest` has `AssertAbsent`, `AssertComposePosture`,
  `AssertQuadletPosture`, `AssertQuadletKeys` and `AssertSecretTargets` for test code, and a
  `Check*` twin of each that returns an error. `noticetest` has the same pair twice: `Check`
  and `CheckDocs` behind `Assert` and `AssertDocs`. That makes seven twins, one decision.
  Both tools call only the `Assert*` ones. The twins exist so M's own tests need no fake
  `testing.T`, and those tests sit in the same package. CR-0001 kept `watch.Loop` and
  `runnertest.FailOn` although unused, so there is precedent in both directions.
  `paritydoc.Check` is different: both tools call it, so it stays.
- **Recommended:** Make the seven `Check*` twins private. Keep every `Assert*`. No tool
  calls a twin, nothing decided asks for running the checks outside a test, and the idea
  that a tool wants the error outside a test is weak: both packages import `testing` anyway.
- **Other choice:** Keep all seven public, for symmetry with `paritydoc.Check`, so the rule
  can be tested without a fake `testing.T`, and so a tool can run a check outside a test,
  for example in a `generate --verify`. If `CheckDocs` stays, export its "nothing was
  checked" error as a type so a direct caller can tell that case with `errors.As`.
- **Changes for:** nobody today. No tool calls a twin.
- **Answer:** ok (2026-11-18).

### N36. `artifacttest.HostNetwork`

- **About:** `HostNetwork()` returns six spellings of host networking, and the first two are
  "the compose one, then the quadlet one", so the order is API. B picks them by index in
  three places, A uses the list once to forbid host networking everywhere. gdu 6 decided
  that every artifact states its network and host is written only when the env file asks.
  That makes A's "forbid everywhere" wrong and B's index picks the same check.
- **Recommended:** Remove it. The four podman flag spellings join `Widening()`. A new
  `AssertNetwork(t, name, artifact, want)` checks that the artifact states the word the tool
  expects. Per-platform accessors would add names in the same tag that removes their
  callers. The verifier overturned the reviewer's accessors.
- **Other choice:** Replace the list with one accessor per platform. Works, but a tool that
  never forbids host networking without stating a word (gdu 6 rules that out) is the only
  reason.
- **Changes for:** A's and B's tests (4 uses) change to `AssertNetwork`. Not env files.
- **Answer:** ok (2026-11-18): the owner took every remaining N recommendation.

### N37. Where `artifacttest` lives, and the Kubernetes checks

- **About:** `artifacttest` sits under `containerart`. The module will write the
  Deployment's security context, and gdu 1's label check covers manifests too. Today the
  Kubernetes forbidden words live as strings inside A's test
  (`allowPrivilegeEscalation: true`, `hostNetwork`, `hostPID`, `hostIPC`). `Widening()`'s
  doc says it leaves Kubernetes out. Ruling 13 says Kubernetes gets no `hostNetwork` line,
  so the check forbids a key the renderer never writes.
- **Recommended:** Move the package to the module's top level (`artifacttest`) and add the
  manifest posture check and the Kubernetes widening words there, so one `AssertLabels` and
  one deny list serve all three platforms. Reason: a sibling package would hold a second
  list of the same forbidden words.
- **Other choice:** Leave it under `containerart` and add a sibling under `kube`. No import
  change for tests that exist, but two deny lists and two label checks.
- **Changes for:** import path in A's and B's tests (2 files). Nothing for operators.
- **Answer:** ok (2026-11-18): the owner took every remaining N recommendation.

### Group: podman, systemd and rootless

### N38. Does the session repair reach podman commands too

- **About:** A rootless run under `sudo`, `su` or cron loses `XDG_RUNTIME_DIR` and the bus
  address. `rootless.UserSession` rebuilds both. A attaches them to `systemctl` and
  `loginctl` commands only. B sets them for the whole process, so every podman and compose
  child gets them too. B chose that on purpose, to keep `Cmd.Env` free for compose secrets.
  Ruling 2 gives every client a session environment. Not checked on a real engine.
- **Recommended:** Both tools pass the same session to the `systemd` client and the podman
  clients. B moves from process-wide `os.Setenv` to the client, and accepts `Cmd.Env` for
  those two non-secret variables. Reason: one rule, and A's podman commands then behave as
  B's do under `sudo`.
- **Other choice:** The session goes to `systemd` only, as in the design note's catalogue.
  Rootless podman under `sudo` can differ between the tools.
- **Changes for:** A: its podman commands gain two variables when the session was repaired.
  B: how the variables reach children, not what they hold.
- **Answer:** ok (2026-11-18): the owner took every remaining N recommendation.

### N39. Which hosts get the user-session repair

- **About:** A repairs the session only on Linux. B repairs it on any host with a POSIX
  effective uid, so on a Mac too. The runtime directory, `loginctl` and the subid files
  exist only on Linux.
- **Recommended:** One test in `rootless`, stated beside the `euid -1` rule: Linux and a
  real uid. Reason: the repair means nothing elsewhere, and B would fail its probe on a Mac
  that has no `/run/user`.
- **Other choice:** Keep B's wider test. B on a non-Linux POSIX host runs a probe that
  cannot succeed.
- **Changes for:** B on a non-Linux POSIX host; nobody on Linux.
- **Answer:** ok (2026-11-18): the owner took every remaining N recommendation.

### N40. The account name in the fix text, now that M reads it in Go

- **About:** The linger and sub-id fix lines name the account itself, because an
  administrator will run them: `sudo loginctl enable-linger alice`. A reads it with
  `os/user`. B runs `id -un`, because both tools build with `CGO_ENABLED=0`, where `os/user`
  reads only `/etc/passwd` and misses directory (LDAP or AD) accounts. Ruling 7 says one Go
  implementation, no `id` child. B's comment says the number works as a fallback; whether
  `usermod` and `loginctl` take a number is not checked on a real host.
- **Recommended:** `os/user`, then the numeric uid, as A does. Reason: it follows ruling 7.
  For a directory account the fix line then names the number. Check on a real host that the
  printed lines work with a number, and add it to docs/smoke.md.
- **Other choice:** Also try `$USER` or `$LOGNAME` before the number. It names the account,
  but it is operator-controlled text printed in a line meant to be pasted, so it must pass a
  name check first.
- **Changes for:** B's operators with directory accounts: the fix line may name a number.
- **Answer:** ok (2026-11-18): the owner took every remaining N recommendation.

### N41. The sub-id refusal: one function, and the range size

- **About:** When the container's user has no sub-id range, each tool joins the same pieces:
  the `CheckIDMap` error, a period and newline, then `SubIDFix` fed by `NextFreeRange`. Both
  are exported today. The tools size the proposed range differently: A uses max(uid, gid)+1,
  B uses the failing id+1. Ruling 7 has M read `/etc/subuid` and `/etc/subgid` itself.
  Example refusal tail: `usermod --add-subuids 165536-231071 alice`.
- **Recommended:** One exported function assembles the whole refusal (cause, account, ids);
  `NextFreeRange` and `SubIDFix` go private. The tool appends only its own alternative. The
  verifier moved this into the breaking tag, because the next tag ships everything. Size the
  range from the larger of uid and gid plus one, so one command fixes both maps.
- **Other choice:** Keep both exported and the join in the tools. The size stays a per-tool
  choice, and each tool writes the same `%w.\n%s`.
- **Changes for:** B's operators: the proposed range can grow when gid exceeds uid. A: none.
- **Answer:** ok (2026-11-18): the owner took every remaining N recommendation.

### N42. `CheckIDMap`: typed error, and both maps in one call

- **About:** `podman.IDMap` reads the uid map or the gid map, chosen by a bool. Both tools
  loop over the pair, one `podman info` call each. `rootless.CheckIDMap` picks the same
  choice with the strings "subuid" and "subgid". B re-reads the raw output to word its
  report row after the check fails. Ruling 9 makes run-as numeric uid and gid, so both ids
  are known.
- **Recommended:** `podman.Client.IDMap` returns both maps from one call as a typed value.
  `CheckIDMap(maps, uid, gid)` checks both, skips only when both ids are 0, and returns a
  typed `*IDMapError` that carries the file, the id and the ranges. The string kind and the
  bool disappear from every signature. One `podman info` template with both ranges is not
  checked on a real engine.
- **Other choice:** Keep one call per kind and add only the typed error. A typed kind
  replaces the strings; both tools keep their loop.
- **Changes for:** both tools' sub-id loops go. Report rows keep their words.
- **Answer:** ok (2026-11-18): the owner took every remaining N recommendation.

### N43. The core-limit checks: common or solace-util's own

- **About:** `ProcLimits` and `HardLimit` read the session's hard core limit. `CheckCore`
  refuses a session whose limit is below what the container's core ulimit needs, because
  `podman exec` would then fail with `setrlimit RLIMIT_CORE`. `Unbounded`, `ParseLimit` and
  `LimitText` are the limit helpers beside them (B uses `LimitText` 10 times). Only B's unit
  sets a core ulimit. A has no such limit. The parity body lists the core check under
  differences that stay in solace-util.
- **Recommended:** Keep all six in M. The failure belongs to rootless podman, not to Solace,
  and any tool that sets a core ulimit needs it. Reword the parity body line. Reason: the
  consumer rule is gone (design 7, point 8), and the rule now asks whether it is common, not
  how many use it today.
- **Other choice:** Move all six into solace-util. M shrinks by six names and A never sees
  them. A third tool with a core ulimit writes it again.
- **Changes for:** nobody's behaviour. Docs: the parity body and developer.md.
- **Answer:** ok (2026-11-18): the owner took every remaining N recommendation.

### N44. Reading `/proc/self/limits` through a command or in Go

- **About:** `ProcLimits` returns the command `cat /proc/self/limits`, and the tool runs it.
  Its doc says this keeps the read on "the host the runner reaches". Today no runner reaches
  another host. Ruling 7 names three host facts to read in Go and gives this one no ruling.
- **Recommended:** Read it in Go inside the rootless check, with one implementation, like
  the other host facts. `ProcLimits` goes. Reason: ruling 7's reason applies equally, and
  every runner today runs on the local host.
- **Other choice:** Keep the command builder. Four host probes then remain as bare commands
  outside the typed errors and writers, and each tool still wraps run, capture and parse.
- **Changes for:** B's `limits.go` drops its probe call. Not checked on a non-Linux host.
- **Answer:** ok (2026-11-18): the owner took every remaining N recommendation.

### N45. `systemctl show`: one `Show` or named methods, and which manager

- **About:** Ruling 6 moves `systemctl show` into M. B uses it three ways: restart count
  (`NRestarts`), unit limits read from the system manager about `user@<uid>.service` while
  the run is rootless, and a probe that the manager answers (`show --property=Version`). The
  new load-state read of gdu 9 is a fourth. Today every systemctl builder takes `user`
  first; the client will hold user mode as a field.
- **Recommended:** One `Client.Show(unit, props...)` returns the properties as a map, and
  `NRestarts`, `LoadState` and a version probe are small callers of it. The manager (user or
  system) is selectable for one call, since B asks the system manager from a rootless run.
  Use one spelling for every show-style read (`--property=`).
- **Other choice:** Named methods only (`ManagerVersion`, `Limits`). No generic map, but a
  new property needs a new method, and B's system-manager call needs a second client.
- **Changes for:** B's `limits.go` and `rootless.go` probes. A gains the manager probe.
- **Answer:** ok (2026-11-18): the owner took every remaining N recommendation.

### N46. Typed unit state for `Active` and `Stopped`

- **About:** `systemd.State` returns the `is-active` word as a string, and `Active` and
  `Stopped` judge that string. `Stopped` accepts the word `unknown` as stopped, and an empty
  answer can read as `unknown`. Ruling 13 fixes that: an empty answer is not read as
  unknown. Ruling 4 gives the typed "unanswered" error.
- **Recommended:** `Client.State` returns a typed `UnitState` and the error "unanswered"
  when nothing was read, never a state word. `Active` and `Stopped` become methods of the
  type. A teardown then cannot proceed on a read that answered nothing.
  `kube.NamespaceState` is the precedent (zero value is unknown).
- **Other choice:** Keep `Active` and `Stopped` as functions on a plain string. Fewer
  renames, but any untrimmed or unrelated string can be passed.
- **Changes for:** both tools change 2 call sites each. `Stopped` stops taking `unknown`.
- **Answer:** ok (2026-11-18): the owner took every remaining N recommendation.

### N47. A single call that resolves the whole podman scope

- **About:** A runs `Decide`, then `QuadletDir`, then `UserSession` in one function. B does
  the same three at three different moments. A third tool would write the three calls
  itself. Example: `rootless.Resolve(euid, rootful, ...)` returning rootless, directory and
  session.
- **Recommended:** Do not add it. B lets the operator set `podman.quadletDir` and resolves
  the session lazily, so B would not use the composite, and the three calls are short. The
  guide shows the three calls once.
- **Other choice:** Add it for a third tool and the example tool. The guide gets shorter.
  One more name that two tools do not call.
- **Changes for:** nobody today.
- **Answer:** ok (2026-11-18): the owner took every remaining N recommendation.

### N48. Does `QuadletDir` stay in `podman`

- **About:** `QuadletDir` returns where a unit is written (the rootless config directory, or
  the system one), and `SystemQuadletDir` is the rootful default. A calls `QuadletDir` once,
  B once, and B reads `SystemQuadletDir` at config load. No client covers them. Once
  `containerart/quadlet` owns the unit they could sit beside it.
- **Recommended:** Leave both in `podman`, with the other podman host facts. Reason: the
  directory is a path rule, not part of the unit text, and B needs the rootful constant at
  config load, where it cannot take the error path.
- **Other choice:** Move both to `containerart/quadlet`. Import paths change in both tools,
  and the unit and its directory sit together.
- **Changes for:** an import path in both tools, only with the other choice.
- **Answer:** ok (2026-11-18): the owner took every remaining N recommendation.

### Group: kube, nsgate and rbac

### N49. How the namespace check lists what is in a namespace

- **About:** Before a Kubernetes remove offers to delete the namespace, both tools list what
  is left in it. Today that is `kubectl get all,persistentvolumeclaims,secrets,configmaps -n
  <ns> -o json`. `-o json` prints every Secret and ConfigMap value, and the tool reads all
  of it into memory. The gate only needs kind, name, uid, owners, labels, whether the object
  is being deleted, and for a Secret its type. Decision gdu 1 widens the list to every kind
  that `kubectl api-resources --verbs=list --namespaced -o name` names, so more objects,
  including custom resources, would come through the same way.
- **Recommended:** Ask kubectl for a fixed short form of each object that holds only those
  fields (a go-template or jsonpath), for every kind alike. Secret and ConfigMap data then
  never reaches the tool. The gate reads this form instead of raw JSON. Reason: gdu 18 and
  the label listing already decided "names or metadata only, never `-o json` on Secrets";
  this is the same call and should follow the same rule. Ruling 1 already keeps the answer
  out of the writers and out of error text, so what is left to remove is the values in the
  tool's memory and in any partial listing. Not checked on a real cluster: the exact
  template, and what a fresh namespace shows under the wider list (the owner check already
  on record for gdu 1).
- **Other choice:** Keep `-o json` and rely on ruling 1 alone. Simplest, and `Occupants`
  does not change. The cost is that every Secret value still enters the tool, and a rule
  that says "never `-o json` on Secrets" would have one exception.
- **Other choice:** Names only (`-o name`) plus the label selector of gdu 1. No values, but
  no owner references and no Secret type either. Pods and ReplicaSets of a Deployment would
  then count as leftovers, so the namespace would be kept far more often.
- **Changes for:** both tools' Kubernetes remove (the gate), and the gate's test fixtures.
  No env file, and no running deployment.
- **Answer:** ok (2026-11-18): the owner took every remaining N recommendation.

### N50. What `NamespaceState` returns when the read fails

- **About:** Both tools read whether the namespace exists with `kubectl get namespace <ns>
  --ignore-not-found -o name`. An empty answer means absent. `namespace/<ns>` means it
  exists. A failed read, or any other answer, is `NamespaceUnknown`, and then the tool lets
  the "may I create namespaces" probe decide. With ruling 4 a failed read now comes back as
  a typed error (refused, unanswered, no context) as well. Ruling 13 says an empty answer is
  not read as "unknown"; here it stays "absent", because that is what `--ignore-not-found`
  prints.
- **Recommended:** Return the state and the error together. A failed or odd read returns
  `NamespaceUnknown` with the typed error; the zero value stays Unknown, so a state nobody
  checked is never taken for an answer. Both tools keep their branch: any error means
  "unknown". Reason: it is what both tools do today, and the safety of the zero value is
  kept.
- **Other choice:** Drop `NamespaceUnknown` and return the error only. One name less, but
  the zero value is then Absent or Exists, and a tool that forgets to check the error acts
  on a guess.
- **Other choice:** Read without `--ignore-not-found` and make "absent" the typed not-found
  error. No empty-answer rule at all, but "absent" then rests on kubectl's NotFound wording,
  and both tools' absent branch becomes an `errors.Is` test. Not checked on a real cluster.
- **Changes for:** both tools' create-namespace step on deploy. Nobody's env file.
- **Answer:** ok (2026-11-18): the owner took every remaining N recommendation.

### N51. The delete check asks about one namespace, not all of them

- **About:** `nsgate.DeleteProbe()` asks "may I delete namespaces" with no name, and `CanI`
  turns every question about namespaces into `--all-namespaces`. So to delete one namespace
  the tool checks the right to delete every namespace. An identity that may delete only its
  own namespace is refused by the check, although the delete itself would pass. The refusal
  then tells the operator to ask for a cluster role binding.
- **Recommended:** `DeleteProbe(ns)` names the namespace, and the probe asks about that one
  (`delete namespaces/<ns>`) without `--all-namespaces`. Reason: the project rule is to ask
  for the lowest permission the write needs, and the wrong answer here is the over-asking
  one. Not checked on a real cluster: whether `can-i` for a named namespace with no `-n`
  gives the same answer as the real delete; it needs one hand check.
- **Other choice:** Keep the cluster-wide ask. The error is in the safe direction (nobody is
  deleted by mistake), but an operator confined to one namespace can never use the gate.
- **Changes for:** operators whose identity may delete only its own namespace: today refused
  by the check, afterwards offered the question. Both tools' remove step passes the name.
- **Answer:** ok (2026-11-18): the owner took every remaining N recommendation.

### N52. The module refuses to delete a protected namespace itself

- **About:** `nsgate.Protected(ns)` says `default`, `kube-system`, `kube-public`,
  `kube-node-lease` and an empty name must never be deleted. Both tools call it as the first
  gate step. `kube.DeleteNamespace` says "run it only through the gate", and nothing
  enforces that. Once `DeleteNamespace` is a client method that a third tool can call, a
  tool that forgets the gate could delete `kube-system`.
- **Recommended:** Move the check into `kube` and make `Client.DeleteNamespace` refuse a
  protected name with an error. The tools keep their first gate step for the wording. `kube`
  cannot import `nsgate` (`nsgate` will import `kube`), so the name has to live in `kube`.
  Reason: it is the same kind of backstop as the name check of gdu 4.
- **Other choice:** Leave `Protected` in `nsgate` and the delete unguarded. Nothing moves,
  but the only guard is a convention that the guide has to repeat.
- **Changes for:** both tools change one call (`nsgate.Protected` becomes `kube.Protected`).
  Nobody else.
- **Answer:** ok (2026-11-18): the owner took every remaining N recommendation.

### N53. How the gate learns that the tool created the namespace

- **About:** Decision gdu 1: a namespace the tool created carries the tool's label. The
  remove gate asks one yes for a labelled namespace, and a yes plus the typed name for an
  unlabelled one. No code reads that label today. The namespace read prints names only
  (`-o name`) and the listing does not return the Namespace object.
- **Recommended:** A separate client method that reads the namespace's labels, used only by
  the remove gate. `NamespaceState` and the deploy path stay as they are. A refused or
  failed label read counts as unlabelled, so the operator gets two questions and never fewer
  (my assumption; nothing decided covers it). Reason: the deploy path should not change for
  a remove-only need.
- **Other choice:** `NamespaceState` returns the labels too. One call, one method, but every
  deploy reads more than it needs and the signature of the most-used read changes.
- **Changes for:** both tools' remove. A namespace created before this change has no label,
  so it takes the two-question path (already accepted in gdu 1).
- **Answer:** ok (2026-11-18): the owner took every remaining N recommendation.

### N54. The gate's question table as a function in the module

- **About:** Decision gdu 1 fixes the table: labelled namespace, one yes; unlabelled, a yes
  and then the typed name; `--no-prompt` deletes only together with `--delete-namespace`; no
  terminal keeps the namespace. Both tools code this step by hand (A `namespace.go:129`, B
  `ops_k8s.go:812`). The package doc of `nsgate` still says "`--no-prompt` approves".
- **Recommended:** Add one pure function to `nsgate`. It takes: labelled or not, terminal or
  not, `--no-prompt`, `--delete-namespace`. It returns one of: keep, delete without asking,
  ask once, ask and require the name. It decides and never asks, so it fits the module.
  Rewrite the package doc to the new table. Reason: two copies of a destructive rule will
  drift.
- **Other choice:** Each tool keeps its own table. No new name, and the parity body keeps
  carrying the rule in words.
- **Changes for:** both tools' remove step, which calls the function and then asks through
  `confirm`. The wording of the questions stays in the tools.
- **Answer:** ok (2026-11-18): the owner took every remaining N recommendation.

### N55. `Object` moves to `kube`, and `Occupants` takes and returns `Object`

- **About:** `nsgate.Object{Kind, Name}` names one object. A uses it 27 times, well outside
  the gate: what a deploy wrote, its permission list, its `generate` result. The new
  leftover removal (gdu 18) lists objects and compares them with what was rendered, and
  needs the same type. `Occupants(out, owned)` today takes the listing as text and returns
  lower-case strings such as `secret/db`; A prints them, B cuts them apart again. Gdu 1 adds
  one more discount, "labelled for this tool and this instance", which the `owned` argument
  cannot say.
- **Recommended:** `Object` lives in `kube` (a listing method returns it, and `kube` should
  not import the documents package for its own return type). `Occupants` takes the listing
  and an `Owned` value (objects, product, instance) and returns `[]Object`; `Object` gets a
  `String` that prints `kind/name` in lower case. An empty product or instance means "count
  everything", as a nil `owned` does now. Reason: one type from the listing to the compare
  to the warning.
- **Other choice:** `Object` stays in `nsgate` and `kube` returns its own type. Fewer
  renames, and two types for one idea.
- **Changes for:** A renames 27 uses, B 3. Warnings that name occupants keep their text.
- **Answer:** ok (2026-11-18): the owner took every remaining N recommendation.

### N56. A read that returns only the key names of a Secret

- **About:** A checks that a Secret named in the env file has the keys the pod reads. It
  does so with `kubectl get secret <name> -o json`, reads the whole Secret, and looks only
  at the key names (`secretprobe.go:36`, `statusreport.SecretDataKeys`). `status --details`
  does the same. Every value comes into the tool. B reads one annotation by jsonpath to keep
  the private key out (`prep.go:350`). Ruling 6 moves "a Secret value read" into the module,
  but not this read.
- **Recommended:** Add a client method that returns key names only (a template that prints
  the keys, never the values), next to the value read. Reason: it is the secret-safe form of
  what A does, and the module should not offer only the dangerous one. Not checked on a real
  cluster: the template. The values still travel to `kubectl`, but never into the tool.
- **Other choice:** A keeps `get secret -o json` over the generic `Output`. No new method,
  and A's deploy and status keep reading values they do not need.
- **Changes for:** A's `deploy` Secret check and `status --details`. B nothing.
- **Answer:** ok (2026-11-18): the owner took every remaining N recommendation.

### N57. What `Attach` carries: terminal and input as two switches

- **About:** Ruling 5 gives `Attach` "hands over the terminal". But A's one-shot
  `cli -- <command>` attaches with no terminal, and passes its own input through only when
  something is piped in. With a terminal it uses both (`-i -t`). B always uses both. Today
  `ExecOptions` has the two switches (`Stdin`, `TTY`). For `Exec` and `ExecOutput` the `-i`
  follows from stdin bytes being given, so they need no switch.
- **Recommended:** `Attach` takes two separate switches, input and terminal, as
  `ExecOptions` has them now, and refuses a terminal without input. `Exec` and `ExecOutput`
  take neither. Reason: A's one-shot would otherwise not fit `Attach` and would stay on the
  generic run.
- **Other choice:** `Attach` is only the full terminal session. A's one-shot goes over
  `Run`, and A keeps the logic of "input only when piped" itself.
- **Changes for:** A's `cli -- <command>` code path; its behaviour for the operator stays.
- **Answer:** ok (2026-11-18): the owner took every remaining N recommendation.

### N58. `NamespaceManifest` becomes a Namespace document, and where it lives

- **About:** `kube.NamespaceManifest("prod")` returns the five-line Namespace text both
  tools apply when they create a namespace. It must now carry the tool's labels (gdu 1) and
  the `lib` and `version` annotations (ruling 10), so a bare name cannot be its input. The
  compose, quadlet and Kubernetes renderers of design section 8 do not have a package yet.
- **Recommended:** Replace it with a `Namespace` document value that carries the shared
  metadata and renders itself. Put the Kubernetes documents in a package under `kube` (for
  example `kube/manifest`) that imports `kube`. `kube` itself imports no documents package,
  because `rbac` and `nsgate` import `kube`. The package name is yours. Reason: `kube` stays
  commands and their answers, and the import order stays free of cycles.
- **Other choice:** Put the documents in `kube` itself. One package less, and `kube` grows
  into commands plus text rendering.
- **Changes for:** A's 8 and B's 5 call sites; the Namespace text in every golden changes
  once (section 8.4 already accepts that).
- **Answer:** ok (2026-11-18): the owner took every remaining N recommendation.

### N59. The name of `kube.ValidLabel`, and the checks beside it

- **About:** `kube.ValidLabel("my-ns")` checks the Kubernetes name rule: 1 to 63 lower-case
  letters, digits and `-`, no `-` at the ends. A calls it 7 times, also for container and
  compose project names. B keeps its own copy and never calls it. A Kubernetes label value
  is a different rule (upper case, `_` and `.` allowed), and gdu 1 brings labels into the
  module, so "label" will mislead. Gdu 4 already adds a subdomain rule (pod, Secret and
  StorageClass names). The old defect that `123` and `true` pass is closed by ruling 9:
  every text value is quoted.
- **Recommended:** Rename it `ValidDNSLabel`, add `ValidDNSSubdomain` (gdu 4), and give the
  label value rule its own name. B adopts all of them and drops its copies. Reason: the
  rename is cheapest in the breaking tag that changes every builder call anyway.
- **Other choice:** Keep `ValidLabel` and name the new rules `ValidSubdomain` and
  `ValidLabelValue`. No rename in A, and the old name stays ambiguous.
- **Other choice:** Put all name rules, with `containercli.ValidName`, in one names package.
  One place to look, and a new package that the review's package inventory has to place.
- **Changes for:** A's 7 call sites; B replaces its two private rules.
- **Answer:** ok (2026-11-18): the owner took every remaining N recommendation.

### N60. Where `MissingNamespace` goes

- **About:** `kube.MissingNamespace(base, ns, err)` words the error "namespace X does not
  exist and this identity may not create it ... or ask an admin to run
  `kubectl create namespace X`". Both tools use it for the same branch (A `namespace.go:75`,
  B `prep.go:72`). It takes the command tokens only to print them. Under ruling 3 the client
  holds the tokens and they cannot be read back.
- **Recommended:** Make it a method on `Client`. The tokens stay inside the client, and the
  one wording stays in the module. Reason: it is the only free function in `kube` that still
  needs the raw tokens, and two tools would otherwise copy the sentence.
- **Other choice:** Free function that takes the command text as an argument; the client
  then has to show its tokens, which ruling 3 tries to avoid.
- **Other choice:** Each tool words it. Nothing in the module, two copies that can drift.
- **Changes for:** the two call sites. The error text for the operator stays the same.
- **Answer:** ok (2026-11-18): the owner took every remaining N recommendation.

### N61. One way to ask "was this refused"

- **About:** A refusal shows in two ways. `*rbac.RefusedError` means the permission check
  (`can-i`) said no. kubectl's "forbidden" text means a command ran and was refused. Ruling
  4 gives every client one "refused" error in `runner`, so the second becomes `errors.Is`. A
  tool also needs to tell "the check said no" from "the server refused".
- **Recommended:** `*RefusedError` also answers yes to
  `errors.Is(err, runner's refused error)`, and keeps its own fields for `errors.As`. A tool
  writes one test for "refused" and, when it cares, a second for "the check said no".
  Reason: both tools turn either refusal into the same warning today (parity rule 16).
- **Other choice:** Keep the two apart. Tools that handle both write two tests.
- **Changes for:** how both tools test refusals. No output changes.
- **Answer:** ok (2026-11-18): the owner took every remaining N recommendation.

### N62. A readiness wait for a pod (`kubectl wait`, gdu 17)

- **About:** Gdu 17 was left to solace-util because it is the only tool with a StatefulSet.
  That reason (a second consumer) was dropped on 2026-11-15. It is still waiting for an
  owner check: `kubectl get sts <name>-pubsubplus-p -n <ns> -o
  jsonpath='{.spec.updateStrategy.type}'`. If the answer is `OnDelete`, `rollout status`
  fails with "only available for RollingUpdate", and a readiness wait on the pod is needed.
- **Recommended:** Add nothing now. If the check says `OnDelete`, add one client method for
  `wait --for=condition=Ready pod/<pod>` in the module, since it is a common operation.
  Reason: nobody has shown it is needed. Not checked on a real cluster.
- **Other choice:** Add the method now as a common operation. No second step later, and a
  method that may never be used.
- **Changes for:** solace-util's start and restart, only if the check says `OnDelete`.
- **Answer:** ok (2026-11-18): the owner took every remaining N recommendation.

### Group: notice, noticetest and paritydoc

### N63. The version command's output block

- **About:** Both tools print the same first line (name, version, Go version, os/arch). A
  prints the short support notice under it. B prints nothing more, so B's binary carries no
  support statement. Example, A: `solmq-conn-util v0.4.0 go1.23 linux/amd64`, then the
  notice. The parity body has no rule about the notice or version output at all.
- **Recommended:** `notice` gains one function that returns the first line plus the short
  notice from name and version. The first line stays unchanged (a script reads it and A pins
  it). B adopts it. The parity body gains one row: which docs carry the notice, that it sits
  directly under the H1, and that `version` prints the short notice under its first line.
  Reason: it is a common artifact, and the rule is to put those in the module.
- **Other choice:** B only calls `notice.Short` under its own line. No new name, and two
  copies of the first line.
- **Other choice:** As the recommendation, and also print the module version on its own
  line. Support gets the version, but one more line in every report.
- **Changes for:** B's `version` output gains the notice lines. A's output stays byte for
  byte.
- **Answer:** ok (2026-11-18): the owner took every remaining N recommendation.

### N64. `notice.Must`

- **About:** `notice.New(name, kind, covers)` returns an error. All six callers pass
  constants and treat any error as fatal, so each wraps it in its own `mustNotice` (A once,
  B twice, M three times). The reviewer called a panic "a new shape for M". It is not:
  `abbrev.New` already panics on a bad declaration, for the same reason as
  `regexp.MustCompile`.
- **Recommended:** Add `notice.Must(name, kind, covers)` that panics with the `New` error.
  Reason: it follows the `abbrev.New` precedent and removes six copies.
- **Other choice:** Leave the wrappers. No new name, and six copies stay.
- **Changes for:** nobody outside the code; both tools delete their wrapper.
- **Answer:** ok (2026-11-18): the owner took every remaining N recommendation.

### N65. The notice wording for tools outside Solace Professional Services

- **About:** The notice sentence is fixed: "created by Solace Professional Services and
  supported only by Solace Professional Services". Only the name, the kind word and the
  covers clause vary. A tool from another team cannot use it. Design 9.2 lists this as a
  gap, beside the `services.solace.com` label values.
- **Recommended:** Leave it. The guide says the notice fits Solace PS tools only. Decide
  before v1, which design step 9.3.4 puts before other teams adopt the module. Reason: no
  consumer needs it now, and parameterising it would change every committed copy.
- **Other choice:** Make the organisation a fourth argument now. One template still, but
  every `docs/*.md` copy in both tools is rewritten once.
- **Changes for:** third tools only.
- **Answer:** ok (2026-11-18): the owner took every remaining N recommendation.

### N66. Should the zero `noticetest.Options` be strict

- **About:** `Options.Strict` demands the notice directly under the H1. Its zero value is
  the weak check. A's `support_test.go:47` runs eight hand-written docs without `Strict`, so
  the weak check is in force by omission. A needs the weak mode only for its two generated
  docs, which carry a "GENERATED" comment between the H1 and the notice. The project rule is
  "directly under the H1", and a third tool that follows the guide gets the weak check from
  `Options{}`.
- **Recommended:** Invert it: the zero value is strict, and an explicit field allows a
  comment between title and notice. Keep `Exclude`: A's allow list left `docs/test.md`
  unchecked (it carries no notice), and a default glob with a deny list would have caught
  it. Reason: the safe default should be the one a new tool gets for free.
- **Other choice:** Keep `Strict` and move A's generated comment below the notice in A's
  generators. No change in M, and the zero value stays weak.
- **Changes for:** call sites in A (2), B (1) and M (1). The compiler does not catch a
  flipped meaning, so each is edited with the change.
- **Answer:** ok (2026-11-18): the owner took every remaining N recommendation.

### N67. `CheckDocs` on a repository with a README and no `docs/*.md`

- **About:** With default options, `CheckDocs` globs `docs/*.md`. If nothing matches it
  fails with "nothing was checked", before it adds `README.md`. The example tool of the
  guide is likely README-only.
- **Recommended:** The default always includes `README.md` when it exists. The
  nothing-checked error stays for when neither a doc nor a README is found. Reason: a
  README-only tool is the smallest real case, and it should not need `Options.Files`.
- **Other choice:** Leave the default and document that a README-only tool passes `Files`.
  Nothing changes in M.
- **Changes for:** third tools and the example tool. Today's two tools and M are unaffected.
- **Answer:** ok (2026-11-18): the owner took every remaining N recommendation.

### N68. Argument order of `AssertDocs`

- **About:** The four checks disagree: `Check(doc, n, o)`, `CheckDocs(root, n, o)`,
  `Assert(t, name, doc, n, o)`, but `AssertDocs(t, n, root, o)`.
- **Recommended:** Make it `AssertDocs(t, root, n, o)`. The types differ, so the compiler
  catches a missed call. Reason: the breaking tag is the only cheap time to do it.
- **Other choice:** Leave it. No edit in A (`support_test.go:47`), B
  (`commanddoc_test.go:98`) or M (`docs_test.go:18`), and the guide shows an odd order.
- **Changes for:** three test call sites.
- **Answer:** ok (2026-11-18): the owner took every remaining N recommendation.

### N69. Two names nobody outside the module uses: `notice.Width` and `paritydoc.Title`

- **About:** `Width` is the 80 column limit of a notice line; no tool reads the number.
  `Title` is the H1 text of `docs/parity.md`; after `Compose` goes (CR-0001), tools only
  call `Check`. Both are reached only through other functions.
- **Recommended:** Make both private. Reword four comments that say "wrapped at Width" to
  "80 columns". For `Title`, spell the page head out in the messages that now tell a reader
  to regenerate the page: the title line, a blank line, the notice, a blank line, then run
  `pin`. `pin` cannot fix line 1 itself, because it keeps each tool's own head lines.
- **Other choice:** Keep both exported. Nothing changes, and a name no tool needs stays in
  the public surface.
- **Changes for:** nobody outside the module.
- **Answer:** ok (2026-11-18): the owner took every remaining N recommendation.

### N70. `paritydoc.CheckFile`

- **About:** Both tools have the same eight-line test: read `docs/parity.md`, call
  `paritydoc.Check`, fail on an error. `Check` names `docs/parity.md` in every message, so
  B, which prefixes the path again, prints the file twice.
- **Recommended:** Do not add it. Keep `paritydoc` free of file access. Instead `Check`
  leaves the path out of its messages and the caller adds it. Reason: eight lines in two
  places is cheap, and a module package that reads files is a new kind of dependency.
- **Other choice:** Add `CheckFile(path)` that reads the file and names the real path. One
  test line per tool, and `paritydoc` starts to touch the file system.
- **Changes for:** the two tools' parity tests.
- **Answer:** ok (2026-11-18): the owner took every remaining N recommendation.

### N71. The parity page for a third tool

- **About:** `paritydoc` serves exactly two tools: its title and the body name them. A new
  tool built on the module (design 9.3) shares the Docker, Podman and Kubernetes contract
  but not those names.
- **Recommended:** The guide says a third tool does not run `paritydoc.Check`, and the
  example tool has no parity page. Revisit with the v1 decision. Reason: a split of the body
  into a general contract and "differences of the pair" is a rewrite of a text both tools
  carry, and no third tool exists.
- **Other choice:** Split the body now into a general contract and the pair's differences. A
  third tool can check the general part, and both tools take a new body in the same adoption
  round.
- **Changes for:** third tools only; today's two tools and their pin step stay.
- **Answer:** ok (2026-11-18): the owner took every remaining N recommendation.

### Group: test support

### N72. Test helpers beyond `Fake`

- **About:** Both tools write their own fake runners (A 11, B 8). They skip
  `runner.CheckRun`, so a builder bug passes. Each writes the same things: answer by what
  the argv contains, answer the permission probe with "yes", and a pause that records and
  stops after n. Ruling 1 means tests now read writers, not a hook.
- **Recommended:** Add to `runnertest` a matcher script (match on whole tokens, then fall
  through) and a recording writer. Put the permission-probe default in a kube test package
  that imports `runnertest`, and a `watchtest` package for the fake sleep. Keep `Call.Has`
  exported (the hand-written Script helper). Reason: the helpers are additive, and today's
  fakes skip the guard.
- **Other choice:** Ship only the matcher and the recorder now. The other two follow when
  the tools ask.
- **Changes for:** tools may delete most of their own fakes. Nothing for operators.
- **Answer:** ok (2026-11-18): the owner took every remaining N recommendation.

### N73. `AssertASCII` moves to the test-support package, with a bytes form

- **About:** `AssertASCII(t, name, lines)` checks that text is plain ASCII. It knows nothing
  about notices. A uses it on both notice forms and twice hand-rolls the same loop on whole
  outputs (`completion_test.go:427`, `statusscript_test.go:67`). `notice` cannot import
  `noticetest` (a cycle), so `notice_test.go:157` and `paritydoc_test.go:47` hand-roll it
  too. The renderer tests of section 8 want it for every artifact. The reviewer left it in
  `noticetest` "until the test-support package exists"; todo items 5 and 8 create that
  package in this same tag.
- **Recommended:** Move it into the test-support package in the breaking tag, with a second
  form that takes `[]byte` and allows a newline. Reason: a second break after the tag that
  carries every break is worse than doing it now. The package name is yours.
- **Other choice:** Leave it in `noticetest` and import that from every renderer test. No
  move, a wrong-way dependency, and the hand-rolled loops stay.
- **Changes for:** A's `support_test.go:45`, the `docs/test.md` row at line 144, and M's two
  loops.
- **Answer:** ok (2026-11-18): the owner took every remaining N recommendation.

### N74. The one diff function for golden and doc mismatches

- **About:** A has three copies of `lineDiff`, B two of `firstDiff`, M a third wording in
  `paritydoc.firstDifference` and a fourth in `noticetest.describe`. All of them print the
  first differing line, both sides quoted. Todo item 5 left open: the output shape, and
  whether `paritydoc` uses it. The copies disagree on argument order and type;
  `artifacttest` takes strings, `noticetest` and `paritydoc` take bytes.
- **Recommended:** One pure function `(got, want []byte)` that returns the first differing
  line, with its number and both sides quoted and clipped, as all copies do today. It lives
  in a package that does not import `testing`, so `paritydoc` can use it without pulling
  `testing` into every importer. `noticetest.describe` and `paritydoc` call it. Bytes
  everywhere new, because a renderer returns bytes. Reason: one wording, and no new
  behaviour to learn.
- **Other choice:** A full line diff. Easier to read for a long golden, and more code with
  no library to lean on (the module takes no dependency without a reason).
- **Changes for:** the golden tests of both tools; every failure message reads the same.
- **Answer:** ok (2026-11-18): the owner took every remaining N recommendation.

### N75. A drift-check helper for generated files

- **About:** Both tools write a test that renders a doc, compares it with the committed file
  and rewrites it under an `-update` flag. B has six package-level update flags, A three.
  Two of the doc gates print no diff at all. With rendering in M, the artifact goldens move
  into M's own tests, so M becomes a third user. `docs/developer.md` forbids a package flag
  in M.
- **Recommended:** The test-support package gets one helper that takes the path, the
  rendered bytes and the update switch as a parameter. It normalises CRLF and reports
  through the diff function of N74. Each tool keeps its own flag. Reason: section 8 forces
  it, so it is not optional.
- **Other choice:** Only the diff function (N74), and every test keeps its own compare and
  update code. Less to build, and 15 copies of the loop.
- **Changes for:** the golden and doc tests of A, B and M.
- **Answer:** ok (2026-11-18): the owner took every remaining N recommendation.

#### Settled by the rulings

- Ruling 1, output. The output hook, `Stream`, chunk marking and the helper that feeds a
  hook go: a client holds two writers the tool supplies. A method that returns an answer
  returns its stdout and does not also write it, and no error text embeds stdout. Follows: a
  recording writer in `runnertest` (N72); `Call.Stdout` and `Call.Stderr` keep telling a
  direct call from a captured one; the hook-versus-`Options.Warn` note in `watch` is moot; a
  real `*os.File` goes direct and a buffer is piped, so piped compose and pull output is not
  measured (N6). The methods that return an answer (can-i, the namespace read, parsed gets,
  the namespace listing) never write it. That closes A's defect of printing a failed
  listing's output. `SecretMatches` never writes the value. `Get` needs no capture-only form
  and is not kept as a fallback. B's `Describe` prints raw `container inspect` as it does
  today, `Config.Env` included; whether that environment can hold secret values in either
  tool was not checked. A's merged stdout and stderr order holds only if the runner gives
  one writer to both streams and runs the child on one pipe; not checked in the runner
  source.
- Ruling 2, generic run. `Run` and `Output` on each client, so there is an exported
  primitive and `rbac.Run` and `nsgate.Read` can use `kube` from other packages. Follows:
  N2. `kube.Forbidden` goes private; it must read stderr alone, because stdout can be
  workload text. Describe, set env, image pull, `systemctl status` and `cat` stay with the
  tool over `Run`.
- Ruling 3, set-up. A constructor runs the guard once and fixes the tokens (`Rules.Guard`
  and "check once per client"). Follows: A must declare `compose` through `Rules.Subword` as
  B does; the pinned context is a constructor input; B builds one client per replication
  site; the engine value is checked at set-up, so an engine other than docker or podman
  fails there; `Docker` and `Podman` become the values the tools use in place of their own
  casts.
- Ruling 4, errors. One set in `runner`: refused, not found, unanswered, timed out,
  unsupported, no context, partial listing. The command's stderr is a field, and the result
  comes back on failure too. This also gives the timeout error the verifier found missing,
  the "no context" login hint that `rbac.Check` words today, and the engine-unreachable
  hint. `InfoError`'s `out` argument goes. `confirm`'s own errors are not among them (N19).
- Ruling 5, exec. `Exec`, `ExecOutput` (optional stdin bytes) and `Attach`; the `Exec`
  builders go private, a pod target and a container target share one small interface, and
  `ExecOptions` shrinks to what the three methods need. An attached session that ran is a
  success with its code. Follows: B calls `exitcode.Exited` on that code (N3, N4). The one
  case ruling 5 leaves open is N57.
- Ruling 6, one-tool operations. Image inspect, stats, top, registry login, a Secret value
  read, compose version, server version, `podman unshare`, copy in and out and
  `systemctl show` move into M. They become `containercli.Client` methods (`ImageInspect`,
  `Stats`, `Login`, `CopyTo`, `CopyFrom`, `ComposeVersion`), `podman.Client.Unshare`,
  `systemd.Client.Show` (N45) and the matching `kube` methods (`Top`, `Copy`, server
  version, which is also the cluster reachability check). `systemd.Systemctl` goes private;
  the `--no-pager` gap of `cat` is fixed by the new methods.
- Ruling 7, host facts. M reads the subid files, the runtime-directory test and the account
  name in Go. No `SubIDFiles`, `DirProbe` or `id -un` builder is added, and the `isDir`
  argument of `UserSession` stops being the tool's. The account name is N40.
- Ruling 8, release order. gdu 2 and gdu 19 ship inside the breaking tag, so adoption
  happens once, and the order of the host-path check against A's tilde lift does not matter
  (N14).
- Ruling 9, one container description. Compose and quadlet each hold the list of keys they
  own and `artifacttest` reads them, so `Widening()` must be complete for each platform and
  is no longer hand-kept in the test package. A read-only root writes its line, and the
  quadlet release table gains `ReadOnly` and `Tmpfs`. The mount type settles `:ro` against
  `:Z`. `compose.Secret` becomes the secret entry (`EnvVar` stays compose-only) and
  `quadlet.Secret` stops being a function. Run-as is numeric for `CheckIDMap`: M always
  receives both numbers, so names are gone and a bare uid is never passed on. Whether the
  env file still accepts a bare uid is E5. Quoting every text value closes `kube.ValidLabel`
  accepting `123` and `true` as bare YAML scalars. A Service name that starts with a digit
  passes A's load and fails at the Service, so each renderer checks its own platform's
  rules; whether a tool also gets an exported checker is S10.
- Ruling 10, the lib annotation. A variable stamped at link time, `devel` under test. The
  unit's `Label=` line and the Namespace document read it.
- Ruling 11 and 13, names. A container name is a DNS label, which is already a valid compose
  project name, so no `ValidProject` fold is needed and B's fold helper goes. The limit of
  100 characters goes (the check is N30). Kubernetes gets no `hostNetwork` line, and compose
  states its project network (the label question is C4); the Kubernetes forbidden words stay
  in the check (N37).
- Ruling 12, no preview. `Echo` stays for tests and debugging, so `runner.IsEcho` is not
  added and B's production `isEcho` branches are B's own clean-up. The Echo branches of
  `Check`, `Occupants` and `NamespaceState` go, "a read answers nothing under Echo" is moot,
  and `Echo.W` becomes required instead of defaulting to the process stdout.
- Ruling 13, corrections. `podman.Client.Version` and `SupportsReplace` go private and no
  tool gets a version reader. `systemd.Stopped` never reads an empty answer as stopped
  (N46). `ParseNamespace` keeps reading an empty answer as absent, because
  `--ignore-not-found` prints nothing for a missing namespace; what the method returns on
  failure is N50. gdu 13 changes A too, so the usage-error rule of `watch` applies to A.
- CR-0001, an earlier decision and not a ruling of section 10: `watch.Loop` and
  `runnertest.FailOn` are kept, `confirm.IsNotTerminal` is removed, and `paritydoc.Body` and
  `Compose` are removed.

#### No decision needed

- `Cmd`, `Runner`, `Runner.Run`, `Runner.Attach`, `Exec` and its methods, `Resolve`,
  `CheckRun`, `CheckAttach` and `Cmd.Check` stay as they are. Making `Cmd.Check` and the two
  free functions one form is the implementer's.
- `Run` takes a context and `Attach` takes none: ruling 5 gives the new methods a context
  where the child allows it, and the implementer states what `Attach` does with it.
- `Echo`, `Echo.Run` and `Echo.Attach` stay.
- `Cmd` gets a copy-safe way to add environment entries, for the systemd client. It is new
  code, not a choice.
- `runnertest`: `Call`, `Reply`, `Fake`, `Fake.Run`, `Fake.Attach`, `Fake.Calls`, `Seq` and
  `FailOn` stay. The implementer names the `Script` type, lets a script see the context,
  lets `Reply.Code` apply to `Run`, adds a strict mode for an unscripted `Attach`, and makes
  `FailOn` match whole tokens (`rm` matches `--format` today).
- `execguard`: `Rules`, `Approvals`, `Check`, `Approve`, `IsInvisible` and `PathSeparators`
  stay. `Check` stays a free function (house pattern; `Rules.Check` would churn about 17
  sites for nothing). Messages that say "this tool" and "the env file" are reworded for a
  third tool. The package doc and `developer.md` say "call `Check` in every executor"; with
  ruling 3 they say "once, at set-up".
- `exitcode`: `OK`, `Failure`, `Usage`, `Usagef`, `AsUsage`, `IsUsage`, `ChildExit`,
  `Exited` and `Code` stay. The constants stay untyped.
- `envref`: `Ref`, `UnsetError`, `Ref.Check` and `Ref.Resolve` stay. The nil-lookup message
  stops saying "run a command that deploys", and the guide shows the one intended `Resolve`
  call.
- `confirm`: `Line`, `Yes` and `YesNo` go private under other names (lower-casing clashes
  with locals). `StdinIsTerminal` goes private. `Form`, `Lenient`, `Strict`, `Options`,
  `Prompter`, `New`, `Terminal` and `Ask` stay. `AskName` is added because gdu 1 decided the
  typed-name question. An unknown `Form` is treated as strict, and a nil `Prompter` is
  documented as losing a buffered answer.
- `abbrev`: `Entry`, `Set`, `New` (it panics, documented), `Expand`, `Short`, `Entries`,
  `Names`, `Shorts` and `List` stay. Renaming `Set.Short` is optional.
- `watch`: `ClearScreen`, `DefaultInterval`, `MinInterval`, `MaxInterval`, `Options`,
  `Loop`, `LoopFrames` and `CanRedraw` stay. `Options.Now` goes private (M's tests set it
  in-package). `LoopFrames` hands the refresh a writer that serialises writes (B's
  `lockedWriter` goes). The stale `watch.Loop` mentions and the package docs naming two
  tools are fixed with the guide.
- `Engine`, `Docker`, `Podman`, `Row`, `HasName`, `ByImage` and `ComposeDigest` stay.
  `ComposeDigest` stays in `containercli` because its second input is the compose child's
  environment, a value known only at command time.
- `Info`, `InfoError`, `EngineHint`, `ComposeUp`, `ComposeDown`, `Start`, `Stop`, `Restart`,
  `List`, `ParseList`, `Upload`, `Logs` and `Exec` go private behind the matching client
  method. `ComposeUp` and `ComposeDown` take an options struct (project for `-p`, directory,
  environment, recreate). Every method returns a result and an error. `Remove` is removed
  (CR-0001).
- `inspect.Containers` returns an error for an empty name list, in place of the `ok` bool,
  and checks each name. `Container.Secrets` becomes name and ID (gdu 10). `State.Restarting`
  and `State.Paused` stay, since they are decoded at no cost.
- `CheckComposePosture` counts the block once per service, not once per file.
- The quadlet key release table moves into `quadlet`, read by `artifacttest`. One podman
  release parser in `podman` serves the key floor.
- The new checks `AssertLabels` and `AssertNetwork`, and a shared golden-diff helper
  (todo.md item 5), are built as decided.
- `Unit.Filename` and the `.service` name come from the unit renderer, so a tool passes the
  container name once. The podman secret store name (container name, dash, suffix) is
  derived once in M; both tools already build names of that shape.
- The podman builders `Version`, `SupportsReplace`, `CreateSecret`, `SecretData`,
  `SecretMatches`, `RemoveSecrets` and `IDMap` go private. Labels become mandatory typed
  input (gdu 16), names are held to a grammar (gdu 4), and `RemoveSecrets`' `ok` bool
  becomes an error or a no-op for zero names. Below podman 4.7 `SecretMatches` reports "does
  not match", so every secret counts as changed, as today. An unanswered version read is
  treated as below 4.7 for the run, as in both tools. The pure parts of gdu 10 and 16
  (labelled names minus rendered names, compare of stored IDs, a partial `secret inspect` is
  a failed listing) are written in `podman`.
- `systemd`: `DaemonReload`, `Start`, `Stop`, `Restart`, `IsActive`, `State`, `NRestarts`,
  `ParseNRestarts`, `ShowLinger`, `EnableLinger` and `ParseLinger` go private behind client
  methods. The client takes the session as `[]string`, so `systemd` imports no `rootless`.
  `Linger` refuses a negative uid. `EnableLinger` stays unclassified and names no user, by
  design. An unknown unit gives "not found", and a linger read that cannot run carries the
  "needs logind" hint.
- `rootless`: `Decide`, `Session`, `UserSession`, `Session.Env`, `Session.String`,
  `CheckLinger`, `Unbounded`, `ParseLimit` and `LimitText` stay. `LingerFix(user)` is added
  beside `SubIDFix`, since both tools print the same two lines. `Session.Env` keeps
  returning only the derived entries.
- `podman.Client` and `containercli.Client` are filled from one base, runner and writers,
  and the guide says so.
- The result of `CreateSecret` below podman 4.7 (version read, `rm`, `create`) and of
  `SecretLabels` (`ls`, then `inspect`) is the last command run (N5).
- `kube`: `Apply`, `Delete`, `Scale`, `RolloutRestart`, `RolloutStatus`, `Logs`, `Upload`,
  `GetNamespace`, `ParseNamespace`, `DeleteNamespace`, `Get` and `Exec` go private behind
  client methods.
- `NamespaceState`, `NamespaceAbsent`, `NamespaceExists`, `GetOptions`, `ExecOptions`,
  `LogOptions`, `RolloutTimeout`, `Probe`, `Probe.String`, `ClusterScoped` and
  `RefusedError` stay as they are.
- `rbac.CanI`, `Check`, `Answer`, `Refused` and `Verdict` go private behind `rbac.Run`.
  `Verdict` is then unused and goes with `TestVerdict`. `nsgate.List` becomes the private
  builder behind `Read`. `CanI` and `List` build their argv with `internal/build` when they
  go private, and the `nsgate` package doc is rewritten to name `Read`, `rbac.Run` and
  `Client.DeleteNamespace`.
- The table from an object kind to its plural resource and scope lives in `kube`, and the
  probes built from a set of objects live in `rbac`: gdu 18 needs it in both tools, and
  `kube` cannot import `rbac`. `ClusterScoped` keeps its name and reads the same table.
- The listing parser and the "listed minus rendered" compare of gdu 18 are decided there.
  Two traps for the build: `-o name` prints `deployment.apps/x` for non-core kinds, which
  `Object`'s kind fold does not equate with `Deployment`; and a partial listing (kubectl
  exits 1 after printing some names) is the typed partial-listing error with no names
  (ruling 4).
- `ValidDNSSubdomain`, `CurrentContext`, `WhoAmI` and the helpers that find and pin
  `--context` (gdu 4, gdu 11), and a client method that deletes one object by kind and name
  with a bound (gdu 12, gdu 18; `DeletePod` becomes a call of it) are added as decided.
- The label keys, the label value grammar and the product-and-instance selector are in the
  label design note (review section 6, item 4). They are not names of this part.
- `Probe.Namespace` empty still means the context's namespace, and B still fills it in
  first. No change is proposed.
- The type of a timeout argument (string against `time.Duration`), where the namespace sits
  in a signature, and whether captured answers are bytes or text are Go details for the
  catalogue step. None touches an env file or an artifact.
- `notice.Notice`, `New`, `Lines`, `Markdown`, `Plain` and `Short`, `noticetest.Options`,
  `Assert` and `AssertDocs`, and `paritydoc.Check` stay. `notice.New` says "pass tool or
  module" but accepts any lower-case word: the message is reworded.
- `paritydoc.Body` and `Compose` are removed (CR-0001). Four messages that tell a reader to
  run `paritydoc.Compose(notice.Markdown())` and the two scripts' remedies are reworded in
  the same change. `TestBodyIsACopy` and `TestComposeRoundTrip` and their `docs/test.md`
  rows go.
- B declares the notice's covers clause twice, in two test packages, and its notice exists
  only in test files. A's `docs/test.md` has no notice and no gate looks at it. Both are
  fixed in the tools during the adoption round.
- The runner package doc's table of retired calls moves to the docs. `developer.md` and the
  parity body are updated with the change. Docs and tests that name a changed name change in
  the same step: `developer.md` (pieces, quoting policy, `Posture`, secret emitters,
  `CheckQuadletKeys`), `consuming.md`, the parity body's rules 4 and 5, the package docs
  saying "nothing executes", the rows of `docs/test.md`, and the compose floor line (Docker
  Compose 2.23.1 for environment secrets) in the `Render` doc.

