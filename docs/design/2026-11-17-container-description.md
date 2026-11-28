# One container description and one writer (consolidation draft)

Status 2026-11-27: this draft matches section 11 of the 2026-11-15 design note as of that
date; the design note wins on any conflict. Settled since the first draft: the Secrets rule (no
`create: true` switch, derived names, a key check before deploy), Service types (M requires a
word and writes no default), `imagePullPolicy` on all three sinks, the read-only root key in
both tools, Kubernetes identity in pod and container fields, the image block, C2 and E4, K4,
K6, K10, K16, K18 and every N entry, and P9 and P11 (owner, 2026-11-27, section 8.1). R1
shipped `containerart/workload` (checks only): names in section 3 that exist in that package
are marked R1. R2 shipped `containerart/writer`, `internal/posture` and the follow-ups to
the checks (`ValidMemory` and `ValidRestart` for Kubernetes, `ValidKubeRestart`,
`ValidPullPolicy`, `ValidServiceType`), marked R2; the rest is the plan for R3 onward.
Plan: 2026-11-27-plan-r2-writer.md.

Draft for item 3 of section 6 of docs/review/2026-11-15-surface-review.md. Read-only inputs:
docs/design/2026-11-15-shared-deploy-api.md sections 8-11 (10 wins over 3-8; 11 wins over
3-8; 10 wins over 11), docs/review/2026-11-15-surface-review-questions.md (every S, E, C, Q,
K and N entry answered), docs/review/2026-11-17-env-layout.md section 4,
docs/review/2026-11-18-secrets-assessment.md section 6, findings section 3, and the current
code of M, A and B.

Words: M = go-deploy-utility. A = solmq-conn-util (repo solace-ibmmq-connector-helper).
B = solace-util (repo solace-cnt-scripts). Citations "A:file:line" and "B:file:line" are
read from the checkouts of 2026-11-17: A:dockergen = internal/dockergen/dockergen.go,
A:podmangen = internal/podmangen/podmangen.go, A:deploy = internal/deploy/deploy.go,
B:render = internal/render/render.go, B:secrets = internal/k8s/secrets.go.

Env-file key names are not the source of a field; the flattened layout is
docs/review/2026-11-17-env-layout.md (answered 2026-11-18). Each "A source" and "B source"
names the Go-side setting that feeds the field today.

## 1. Purpose

M renders the compose file, the quadlet unit and the Kubernetes documents from one container
description that each tool hands over, so the text, the order, the quoting and the hardening
lines are written once and checked once. The description is a set of value types with
parsers and checks in a new package, containerart/workload, and the three renderers take it
as their input; the tool keeps the content (what to run) and M owns the platform artifact
(how it is written). Every value rule is an exported check that the tool calls at load and
the renderer calls again, so a bad value is refused before the first question and never
reaches a file.

## 2. Package layout and imports

```
containerart/writer      one YAML node writer, one unit-line writer, the two quotes
containerart/workload    the description, value types, parsers, checks, FieldError
containerart/compose     File (one service) -> bytes; imports workload, writer
containerart/quadlet     Unit -> File{Name, Content, Digest}; imports workload, writer
containerart/kubernetes  Manifest of documents; Deployment takes a workload.Spec;
                         imports workload, writer
containerart/outfile     S11/Q5 write helper (stdlib only)
internal/posture         Posture constants and the S1 deny lists (module-level internal)
artifacttest             module top level (N37): checks on rendered text; imports renderers,
                         workload and internal/posture, never the reverse
```

Import graph (an arrow reads "imports"; no cycle is possible because every arrow points
down this list):

```
artifacttest -> compose, quadlet, kubernetes, workload, posture, testing
compose      -> workload, writer, posture
quadlet      -> workload, writer, posture
kubernetes   -> workload, writer, posture
workload     -> execguard, envref (R1: nothing else)
writer       -> standard library only
posture      -> standard library only
outfile      -> standard library only
```

Rules that keep it acyclic:
- workload and writer import no renderer and no artifacttest. A renderer never imports
  artifacttest (it imports `testing`).
- workload imports neither `kube` nor `containercli` (R1): the DNS-label rule is written in
  workload (`ValidDNSLabel`) and `kube.ValidLabel` delegates to it in the kube step (N59), so
  no cycle arises. `kube.NamespaceManifest` still leaves `kube` and becomes a Namespace
  document value in a package under `kube` (N58); see P1.
- The Posture constants and the deny lists (S1) are read by both renderers and
  artifacttest. N25 puts them in an `internal/` package; it sits at the module root, because
  the top-level artifacttest of N37 cannot import a `containerart/internal` package. The
  existing `compose.Posture` and `quadlet.Posture` go away as exported names.
- `compose.Quote`, `compose.Escape`, `compose.FlowList`, `compose.ServiceSecrets`,
  `compose.TopSecrets` and the `quadlet` helpers become private; `writer` exports the two
  quotes (`writer.Quote`, `writer.QuoteDollar`), N25.
- The renderers write no file and run nothing. Writing is `outfile`'s, called by the tool.

What each package holds:

- writer: `Node` tree (`Str`, `Int`, `Bool`, `Word`, `Block`, `Map`, `Seq`, `EmptyMap`),
  `Join` for documents, a line writer for unit files, `Quote`, `QuoteDollar`, `UnitQuoted`,
  `Percent`. No YAML library, stdlib only, deterministic bytes for goldens.
- workload: section 3.
- compose: `File{Project, Service workload.Spec, Extra []Setting}` and `Render`. Compose only
  facts: project name grammar, docker network words, `Setting` extras and their deny list,
  top-level `networks`/`configs`/`secrets` derived from the one service.
- quadlet: `Unit{Spec workload.Spec, Description, Rootful, Extra []Setting}` and
  `Render() (File, error)`. Quadlet only facts: podman network words, unit layout, `Ulimit=`
  and `Limit*=` lines, digest label two-pass, the `[Container]` extras deny list.
- kubernetes: `Manifest{Owner, Docs}`, `Meta`, `Namespace`, `Secret`, `ConfigMap`, `Service`
  (requires a type word), `PersistentVolume`, `PersistentVolumeClaim`, `Deployment` (takes
  `workload.Spec` plus the Kubernetes-only pod fields) and the other workload kinds of K6,
  `Custom`, `Raw`, `Refs`, `Objects`, `Join`. Pod and container extras and their deny lists.
- outfile: `Write` (umask mode, for `generate` output) and `WriteMode` (explicit mode, for
  the installed unit).
- artifacttest: today's checks, plus the stated-network check (`AssertNetwork`, N36, which
  replaces `HostNetwork`), the label check, the Kubernetes posture check (K6) and the deny
  lists read from internal/posture.

## 3. The description (Go sketch)

Rules for the whole description:
- The zero value of every optional field writes nothing. A required field left at its zero
  value is an error. No field defaults to writable, root, host or widened (ruling 9).
- Text from a tool is checked with `CheckText` (valid UTF-8, no control character via
  `execguard.IsControl`, no invisible character via `execguard.IsInvisible`) unless its own
  grammar is stricter. S8: no error echoes a free-text value.
- A field that a sink cannot express is refused for that sink with an error naming the
  field; it is never dropped silently. The one exception is ports under host or none
  networking, which are left out (ruling 9, doc says so).

```go
package workload

// Sink names the renderer a check runs for. The zero value is invalid.
type Sink int

const (
	SinkCompose Sink = iota + 1
	SinkQuadlet
	SinkKubernetes
)

// Spec is the container description. One Spec is one container. A tool fills it once and
// passes it to compose.File, quadlet.Unit or kubernetes.Deployment.
type Spec struct {
	Identity   Identity   // required. Tool and Instance non-empty; see Identity.
	Name       string     // required. Container name; compose service key and container_name,
	                      // the unit stem and ContainerName=, the Kubernetes container name.
	Image      Image      // required. Tag or digest required (S7).
	PullPolicy string     // "" writes nothing (the engine's or cluster's own default). Words per
	                      // sink, see ValidPullPolicy. A tool sets its own default (E25).
	HostName   string     // "" writes nothing. RFC 1123 (S12). Refused on Kubernetes.
	RunAs      RunAs      // zero value = not stated = the image's own user; writes nothing. The
	                      // tool maps its word `default` to the zero value and states its own
	                      // default when the env key is unset (E5); M fills in nothing.
	Restart    string     // "" writes "always" (E6, ruling 9). Words per sink, see ValidRestart.
	Root       RootFS     // required (RootUnset is an error). Read-only writes the line;
	                      // writable writes nothing (ruling 9, gdu 6). A true, B false (E21).
	Timezone   string     // "" writes nothing. IANA name (S13). Written as TZ, first.
	Env        []EnvVar   // nil writes nothing. Given order kept. TZ here beside Timezone
	                      // is an error (duplicate).
	Ports      []Port     // nil writes nothing. Dropped under host or none (ruling 9).
	Network    string     // "" = the deploy's own network (compose project network,
	                      // Network=private); a written `bridge` is docker's shared bridge (C2).
	                      // Refused on Kubernetes (ruling 13).
	Memory     string     // "" writes nothing. Grammar per sink (ValidMemory); on Kubernetes
	                      // `limits.memory`, a bare `m` suffix refused (E12).
	CPUSet     string     // "" writes nothing. Refused on Kubernetes.
	ShmSize    string     // "" writes nothing. Refused on Kubernetes.
	Ulimits    []Ulimit   // nil writes nothing. Refused on Kubernetes (S2).
	Secrets    []Secret   // nil writes nothing. Files at absolute targets.
	Configs    []Config   // nil writes nothing. Inline content (see Config).
	Mounts     []Mount    // nil writes nothing. Host bind mounts; refused on Kubernetes.
	Health     *Health    // nil writes nothing (the image's own check applies).
	Labels     []Label    // the tool's own labels, after M's own. nil writes nothing.
	Digest     DigestLabel // zero writes nothing.
}

// Identity is the gdu 1 identity. M writes the five ownership labels from it.
type Identity struct {
	Tool        string // required. managed-by and product labels. Label-value grammar.
	Instance    string // required. deploy.solace.com/instance. Per-sink length rule.
	Solution    string // "" leaves the services.solace.com/solution label out (S4)
	DevelopedBy string // "" leaves the services.solace.com/developed-by label out (S4)
	Version     string // Kubernetes only: deploy.solace.com/version. "" omits it. Never on
	                   // compose or quadlet (no byte changes on a tool upgrade alone).
}

// RootFS has no default: the tool states one (ruling 9).
type RootFS int

const (
	RootUnset    RootFS = iota // error
	RootReadOnly               // read_only: true | ReadOnly=true | readOnlyRootFilesystem: true
	RootWritable               // writes nothing (C3: B's file stays free of read_only)
)

// Image is registry/repository:tag@digest (E3). Registry "" means Docker Hub and is not
// written. Repository is the path only, never a host (the env key is `image.name`, E3
// revised; the Go name keeps the OCI word). Every image a tool names, A's libs download
// image and B's operator image included, is a full block (E23): no hidden prefix.
type Image struct {
	Registry   string // host[:port]
	Repository string // lower-case path, components [a-z0-9]+([._-][a-z0-9]+)*
	Tag        string // [A-Za-z0-9_][A-Za-z0-9_.-]{0,127}; "latest" allowed (S7)
	Digest     string // sha256:<64 hex>
}

// RunAs: stated is distinguishable from empty (S3). Numeric uid and gid, 0..2147483647.
type RunAs struct{ uid, gid uint32; stated bool } // built by RunAsIDs or ParseRunAs

// Port is [ip:]host:container[/tcp|/udp]; ranges lo-hi; IPv6 in brackets (E7: a bare N
// means N:N; String always writes both sides).
type Port struct {
	IP        string // "" or an IP literal; IPv6 without brackets here
	Host      string // port or lo-hi
	Container string // port or lo-hi
	Protocol  string // "", "tcp" or "udp"
	Name      string // Kubernetes only: Service/container port name, DNS label <= 15
}

// Mount is a host bind mount (compose short syntax, quadlet Volume=). The zero value is
// read-only (ruling 9): the flag is inverted so a forgotten field writes ":ro".
type Mount struct {
	Source   string  // absolute host path (CheckMountSource)
	Target   string  // absolute container path, no ':'
	Writable bool    // false writes the ro option
	Relabel  Relabel // none (zero), z, Z (Q8)
}

type Relabel int

const (
	RelabelNone Relabel = iota
	RelabelShared  // z
	RelabelPrivate // Z
)

// Secret is a secret file. The value never enters the description.
type Secret struct {
	Name   string // compose secret name | podman secret name | Kubernetes Secret name
	Target string // required, absolute file path in the container
	EnvVar string // compose only: the variable the environment provider reads (C6 refusals)
	Key    string // Kubernetes only: the key inside the Secret
	Volume string // Kubernetes only: volume name; same Volume = one projected volume
}

// Config is inline text. Compose: a configs entry with content. Kubernetes: a key of
// the Deployment's ConfigMap, mounted by subPath. Quadlet: a Secret= line named Name;
// the tool creates the podman secret from Content (open P5).
type Config struct {
	Name    string // compose config name | podman secret name
	Target  string // required, absolute
	Key     string // Kubernetes only: ConfigMap key; "" = basename of Target
	Content string // CheckInlineText (C7); bytes kept as given (C10); never echoed (S8)
}

type EnvVar struct{ Name, Value string } // Name via envref.ValidName

// Label is one of the tool's own labels. Kubernetes: a pod template label.
type Label struct {
	Key   string
	Value string
	Var   string // compose only: write "${Var:-}" instead of Value (C-digest rule)
}

// DigestLabel is the redeploy label. Compose: "${EnvVar:-}", last. Quadlet: the digest of
// the unit text, last; EnvVar is ignored. Kubernetes: refused.
type DigestLabel struct{ Key, EnvVar string }

// Health is the exec-form probe. Whole seconds (S6).
type Health struct {
	Test        []string      // argv; at least one token, none empty
	Interval    time.Duration // > 0, whole seconds
	Timeout     time.Duration // > 0, whole seconds
	StartPeriod time.Duration // >= 0, whole seconds
	Retries     int           // >= 1
}

// Ulimit is one typed limit (S2). Name is one of 13 accepted Linux names (the 16 names of
// getrlimit less rtprio, rttime and nice, which are refused). Soft <= Hard; -1 is unlimited.
type Ulimit struct {
	Name       string
	Soft, Hard int64 // Hard == Soft writes the single-value form where the sink has one
}
```

Exported names and the check each value type carries (the S10 core set). Each check takes a
field path for its error and never echoes free text (S8):

| Name | Kind | What it checks |
| --- | --- | --- |
| `Spec.Check(Sink)` | method | every field below for that sink, then the cross checks of 6 |
| `Image.Check`, `ParseImage`, `Image.String` | parser | grammar above; tag or digest required (R1: `ParseImage`, `String`) |
| `ParseRunAs`, `RunAsIDs`, `RunAs.Stated/UID/GID/String` | parser | numeric uid:gid, both halves |
| `ParsePort`, `Port.Check`, `Port.String` | parser | port grammar; range lo <= hi; 1..65535 |
| `Mount.Check(Sink)`, `CheckMountSource(Sink, s)` | check | absolute; one drive-letter root; no ':' else (R1: the second) |
| `Secret.Check(Sink)` | check | target absolute; per-sink required fields; names |
| `Config.Check(Sink)` | check | target absolute; `CheckInlineText` on Content |
| `CheckInlineText` | check | C7: no CR, NEL, LS, PS, control (tab, LF ok), DEL, C1, BOM; UTF-8 (R1) |
| `Health.Check(Sink)` | check | S6 rules; on quadlet Q6 shell-character refusal (R1) |
| `ParseSeconds` | parser | a Go duration string that is a whole number of seconds (R1) |
| `Ulimit.Check(Sink)`, `UlimitNames`, `ParseLimit` | check | 13 names, soft <= hard, -1, infinity (R1) |
| `ValidHostName` | check | RFC 1123 (S12) (R1) |
| `ValidTimezone` | check | IANA name grammar (S13) (R1) |
| `CheckComposeVar` | check | C6: `envref.ValidName`, then `PATH`, `HOME`, `IFS`, `SHELL`, `BASH_ENV`, `LD_`, `DOCKER_`, `COMPOSE_` refused (R1; replaces the planned `CheckEnvName`) |
| `CheckLabelKey`, `CheckLabelValue` | check | qualified name; reserved prefixes (S1 answer) |
| `ReservedLabelPrefixes` | list | M's own prefixes and the engine prefixes of S1 |
| `ValidRestart(Sink, word)` | check | compose/quadlet words; Kubernetes `Always`, `OnFailure`, `Never` as a union, `always` refused (P9; R2 changes R1) |
| `ValidKubeRestart(kind, word)` | check | R2. The per-kind rule of P9: Deployment, StatefulSet, ReplicaSet, DaemonSet `Always` only; Job, CronJob `OnFailure`, `Never`; Pod all three; empty allowed; an unknown kind refused naming the seven kinds. The kind is a plain word the renderer passes; `Spec.Check` runs the kind-free `ValidRestart` |
| `ValidNetwork(Sink, word)` | check | docker and podman word lists of gdu 6; Kubernetes: empty (R1) |
| `ValidPullPolicy(Sink, word)` | check | R2. Compose `pull_policy`: `always`, `missing`, `never`, `daily`, `weekly`, `every_<duration>`, `if_not_present`; `build` refused. Quadlet `Pull=`: `always`, `missing`, `never`, `newer`. Kubernetes: `Always`, `IfNotPresent`, `Never`. "" allowed. Exact case; `build` refused with "M never builds". Lists checked against each engine's docs at build |
| `ValidServiceType(word)` | check | R2. `ClusterIP`, `NodePort`, `LoadBalancer`, exact case; `ExternalName` refused; empty is an error (M writes no default) |
| `ValidMemory(Sink, s)`, `ValidCPUSet(s)`, `ValidShmSize(Sink, s)` | check | per-sink grammar (R1); R2 (E12) gives `ValidMemory` on Kubernetes the quantity grammar (digits, one fraction, `k M G T P E`, `Ki Mi Gi Ti Pi Ei` or `e<digits>`), bare `m`, `u` and `n` refused with "write Mi or M, not m"; `ValidShmSize` still refuses Kubernetes |
| `CheckName` | check | container name: DNS label <= 63 (rulings 11, 13; N30) (R1) |
| `ValidDNSLabel` | check | the DNS-1123 label rule `kube.ValidLabel` delegates to (R1; N59) |
| `CheckText` | check | UTF-8, controls, invisibles (R1) |
| `CheckTargets` | check | duplicates and nesting (C5); section 6 |
| `FieldError` | type | R1: `Field`, `Rule`, `Value`; `At(field)` renames; `Error()` never holds free text. Sink and object context: P11 |

`FieldError` carries a `Rule` sentence and a `Field` path such as
`env["JAVA_TOOL_OPTIONS"]` (R1 has no sink or object member; a renderer adds them, P11). A value is echoed only for names, keys, fixed-list words and
numbers (S8). A render returns `errors.Join` of every problem found, nothing partial.

Exported names in package workload, flat:
`Sink`, `SinkCompose`, `SinkQuadlet`, `SinkKubernetes`, `Spec`, `Identity`, `RootFS`,
`RootUnset`, `RootReadOnly`, `RootWritable`, `Image`, `ParseImage`, `RunAs`, `RunAsIDs`,
`ParseRunAs`, `Port`, `ParsePort`, `Mount`, `Relabel`, `RelabelNone`, `RelabelShared`,
`RelabelPrivate`, `CheckMountSource`, `Secret`, `Config`, `CheckInlineText`,
`EnvVar`, `Label`, `DigestLabel`, `Health`, `ParseSeconds`, `Ulimit`, `UlimitNames`,
`ParseLimit`, `ValidHostName`, `ValidTimezone`, `CheckComposeVar`, `ValidDNSLabel`,
`ValidPullPolicy`, `ValidServiceType`, `CheckLabelKey`,
`CheckLabelValue`, `ReservedLabelPrefixes`, `ValidRestart`, `ValidNetwork`, `ValidMemory`,
`ValidCPUSet`, `ValidShmSize`, `CheckName`, `CheckText`, `CheckTargets`, `FieldError`.

### 3.1 Sink-owned inputs (not in workload)

```go
package compose
type File struct {
	Project string    // required; ^[a-z0-9][a-z0-9_-]*$; never folded, never omitted (C9)
	Service workload.Spec // one service (C11)
	Extra   []Setting // deny list S1; [Container]-like keys of the service only
}
type Setting struct { Key string; Value writer.Node }
func (f File) Render() ([]byte, error)
func ValidProject(s string) bool

package quadlet
type Unit struct {
	Spec        workload.Spec
	Description string    // required, one line, % doubled
	Rootful     bool      // false: WantedBy=default.target; true: multi-user.target (Q11)
	Extra       []Setting // [Container] only (Q1); S1 deny list
}
type File struct { Name, Service string; Content []byte; Digest string }
func (u Unit) Render() (File, error)
const FileMode os.FileMode = 0o600 // Q5

package kubernetes
// Deployment takes the shared description plus what only Kubernetes has.
type Deployment struct {
	Meta       Meta
	Replicas   int
	Selector   map[string]string
	Container  workload.Spec
	Init       []workload.Spec  // init containers: same hardening, same description
	Pod        Pod              // placement, identity (E22), volumes, pull secrets, probes
	ConfigMap  string           // name of the ConfigMap that Container.Configs go into
	ConfigVolume string         // volume name for that ConfigMap
	Resources  *Resources       // requests and limits, cpu and memory (K-only)
	Liveness   *Probe           // tcp, http or exec (K-only)
	Readiness  *Probe           // when set, Container.Health must be nil
	Mounts     []VolumeMount    // volume name, path, subPath, writable (zero = readOnly)
	Extra      writer.Map       // pod extras; deny list S1 (pod)
	ContainerExtra writer.Map   // container extras; deny list S1 (container)
}
```

The other Kubernetes documents (`Namespace`, `Secret`, `ConfigMap`, `Service`,
`PersistentVolume`, `PersistentVolumeClaim`, `Custom`, `Raw`, `Manifest`) keep the model of
findings 3.3, corrected by rulings 9-13 and S/K answers: data-only Secrets from raw bytes
(K2), name-only references for a remove (K3; a CRD is never deleted, K4), data keys quoted
(S5), `defaultMode: 0400` (S9, K14), the Docker Hub key constant (K11), one `---` joiner, one
`Objects` list (K15). They take `Meta` (name, namespace, labels, annotations) and write the
ownership labels and the lib and version annotations themselves (ruling 10). The operator
bundle and `regcred` carry no ownership labels and are handled by name (K9, K10). They are not
part of the container description: a Secret or ConfigMap holds data, not a container.

Typed kinds (K6, K18, 2026-11-17 and 2026-11-18). Every workload kind is typed with the
hardening lines built in: Deployment, StatefulSet, ReplicaSet, DaemonSet, Job, CronJob and
Pod; `Custom` refuses every kind M types, and `artifacttest` still requires the hardening
lines in every workload a tool ships, raw vendor text excepted. Also typed: Ingress,
NetworkPolicy, PodDisruptionBudget, HorizontalPodAutoscaler, ServiceAccount, Role,
RoleBinding, ResourceQuota, LimitRange, PriorityClass and StorageClass. ClusterRole and
ClusterRoleBinding are refused; Gateway API kinds wait. Nothing else is built ahead of need.
The CRD is never modelled. The build order of these kinds is the plan's, not this draft's.

Service (2026-11-18). `Service.Type` is required and checked by `ValidServiceType`: M writes
no default. A defaults to `ClusterIP`, B to `LoadBalancer` (written into its custom resource,
not by M). No fixed node port is offered: the cluster picks one.

Secrets (owner, 2026-11-27, secrets-assessment section 6). No Secret has a `create: true`
switch, TLS included. The tool decides per Secret from its env file:
- Material set (a literal or its `...Env` key, whatever the value): the tool builds the
  `kubernetes.Secret` under the derived name.
- `existing: NAME`: the tool references that Secret and builds and deletes nothing.
- Both: a load error in the tool, with two exceptions. A's credential fields under `existing`
  name the keys of that Secret (the host variable is never read). B's `image.user` and
  `image.pass` also build `regcred`, so under `imagePull.existing` they build only `regcred`.
- Neither: nothing is built or referenced (B's admin Secret: the operator generates one).
- Derived names belong to the tool: a Secret already under one is replaced on deploy, and
  the docs say so. Remove deletes every derived Secret name with `--ignore-not-found`, needing
  no material, and `existing:` may not name a derived name.
- Before deploy M checks that an `existing` Secret holds the keys the tool reads (key names
  only, the tool passes them; see `kube.Client.SecretKeys` in the catalogue). A refused read
  warns and does not block (least privilege).
- `create:` blocks that hold material stay: A's `libs.pvc.create` and A's per-store `create`.
  `generate` prints a Secret for each built one and none for `existing`.

## 4. One table per field

Legend: "C" compose, "Q" quadlet, "K" Kubernetes. "Check" names an exported check of 3.
"Answer" is the entry that decides it. "open X" means the Recommended choice was used.

### Spec.Identity

| Column | Value |
| --- | --- |
| A source | tool name const; docker.name / podman.name / kubernetes.deployment.name |
| B source | tool name const; container name (cb.Name), k8s name (B:render:81-82) |
| C | five labels first in `labels:`, quoted keys; no labels on the network (C4) |
| Q | five `Label="k=v"` lines after ContainerName/HostName; header line (Q10) |
| K | object labels (Meta), never in selector or pod template; version annotation (K only) |
| Check | `CheckLabelValue`, `CheckName`, per-sink instance length (K: <= 63) |
| Answer | gdu 1, S4 (solution, developed-by optional), ruling 10 (lib annotation), C4; the instance is `<platform>.name`, and B's `redundancy.sameHost` names stay the tool's |

### Spec.Name

| Column | Value |
| --- | --- |
| A source | `inst.Name` (A:dockergen:152,154; A:podmangen:149; A:deploy:537,608) |
| B source | `cb.Name` (B:render:1071,1073; :902) |
| C | service key and `container_name`, both quoted |
| Q | `ContainerName=`; `<Name>.container` must fit 255 bytes |
| K | container `name`; Deployment name comes from `Meta` |
| Check | `CheckName` (DNS label <= 63); `containercli.ValidName` stays for names read back |
| Answer | rulings 11, 13; N30 |

### Spec.Image

| Column | Value |
| --- | --- |
| A source | `inst.Image` string built from image.* (A:dockergen:153; :podmangen:148; :deploy:609) |
| B source | `c.Image.Ref()` (B:render:1072; :901) |
| C | `image: "<ref>"` quoted, first service key |
| Q | `Image=<ref>`, first [Container] key, `%` doubled (grammar leaves nothing else) |
| K | container `image`, quoted |
| Check | `ParseImage`, `Image.Check`: tag or digest required |
| Answer | E3 (`image.name` feeds Repository), E23 (full block for every image), S7 |

### Spec.PullPolicy

| Column | Value |
| --- | --- |
| A source | new key `imagePullPolicy` on docker, podman and kubernetes (env-layout section 4) |
| B source | the same key; B's Kubernetes value goes into its custom resource, not through M |
| C | `pull_policy: "<word>"`, after `image`; zero writes nothing |
| Q | `Pull=<word>`, after `Image=`; zero writes nothing |
| K | `imagePullPolicy: <word>` after `image` (5.4); zero writes nothing (K1: the cluster's own default) |
| Check | `ValidPullPolicy` (words in section 3); compose `build` refused (no build section is written) |
| Answer | 2026-11-27: each platform's own words, M checks them, the tools set the defaults (docker and podman `missing`, B Kubernetes `IfNotPresent`, A Kubernetes nothing written); the pre-deploy pull (E18) follows the value, see the catalogue |

### Spec.HostName

| Column | Value |
| --- | --- |
| A source | none |
| B source | `id.Hostname` (B:render:1074; :903) |
| C | `hostname: "<h>"` after container_name |
| Q | `HostName=<h>` (podman 4.6) |
| K | refused, error names the field |
| Check | `ValidHostName` |
| Answer | S12, S2 |

### Spec.RunAs

| Column | Value |
| --- | --- |
| A source | `inst.RunAs` (UID, GID, Image flag) (A:dockergen:191-193,253-255) |
| B source | `cb.RunUser` string "uid:gid" (B:render:1075; :904-907) |
| C | `user: "uid:gid"`; zero writes no line |
| Q | `User=<uid>` then `Group=<gid>`; zero writes neither |
| K | pod and container fields (E22), see P11; `runAsNonRoot` when the uid is not 0 |
| Check | `ParseRunAs`, `RunAs` constructors; uid 0 allowed when stated |
| Answer | S3, E5 (the word `default` is the zero value), E22 (pod and container fields on Kubernetes in both tools) |

The word `default`, in both tools and on every platform, Kubernetes included, maps to the zero
`RunAs` and tells M to write no identity. An unset env key means the tool's stated default
(A `1000:0`; B `1000:0` on rootless podman, `1000001:0` on docker and rootful podman; B on
Kubernetes user 1000001, group and fsGroup 1000002, and a stated 0 refused, because the CRD
reads 0 as the operator's default). A stated `0:0` writes the root lines without
`runAsNonRoot`. With `default` B reads the image's user (image inspect, ruling 6) for its
data-directory chown and refuses a user that is not numeric (E5 answer).

### Spec.Restart

| Column | Value |
| --- | --- |
| A source | `d.Restart` / `p.Restart` (A:dockergen:183-185; :podmangen:223-227) |
| B source | constant always (B:render:1076; :1010) |
| C | `restart: "<word>"` quoted; `"no"` quoted |
| Q | `[Service]` `Restart=<word>`, always written |
| K | no line; Deployment always restarts. Any word but "" and always is refused |
| Check | `ValidRestart` |
| Answer | ruling 9, E6 (default always, word lists move into M) |

### Spec.Root

| Column | Value |
| --- | --- |
| A source | new key `readOnlyRoot`, default true (E21) |
| B source | new key `readOnlyRoot`, default false (E21) |
| C | `read_only: true` after the posture block; writable writes nothing |
| Q | `ReadOnly=true` after the posture lines; writable writes nothing |
| K | `readOnlyRootFilesystem: true` in the container securityContext |
| Check | `RootUnset` is an error |
| Answer | ruling 9, C3 (smoke row before A turns it on), E21: both tools have the key and M requires a value; whether B accepts true depends on B supporting a read-only root (B's session) |

### Entrypoint and command (not fields, P4)

Not fields (P4, owner 2026-11-27). Only A's docker store script uses an entrypoint and a
command (A:dockergen:158-161), so they are A's, not common: A passes `entrypoint:` and
`command:` through the compose extras, held to the extras deny list (S1) and `CheckText`.
Quadlet and Kubernetes have neither.

### Spec.Timezone

| Column | Value |
| --- | --- |
| A source | `inst.Timezone` (A:dockergen:207-209; :podmangen:168-170; :deploy:622-625) |
| B source | TZ pair in EnvPairs (B:render:784-786) |
| C | first `environment` entry `"TZ": "<zone>"` |
| Q | first `Environment="TZ=<zone>"` |
| K | first env entry TZ |
| Check | `ValidTimezone`; Env named TZ beside it is a duplicate error |
| Answer | ruling 9, S13 (IANA grammar, POSIX rules refused) |

### Spec.Env

| Column | Value |
| --- | --- |
| A source | JAVA_TOOL_OPTIONS, JDK_JAVA_OPTIONS, LOGGING_SYSLOG_* (A:dockergen:204-222) |
| B source | `EnvPairs(c, id)` broker settings (B:render:779-864, :1154-1160, :986-988) |
| C | mapping `"NAME": "value"`, key quoted (S5), value quoted, `$` doubled; `""` for empty |
| Q | `Environment="NAME=value"`, whole assignment quoted, `\ " % $` escaped |
| K | `env:` list, name and value quoted, `$` doubled; no valueFrom, no envFrom |
| Check | `envref.ValidName`, `CheckText`; duplicates refused |
| Answer | S5, S8, C6, C7 not applicable (inline text only), ruling 9 quoting |

### Spec.Ports

| Column | Value |
| --- | --- |
| A source | `d.Ports` / `p.Ports` (A:dockergen:195-200; :podmangen:165-167) |
| B source | `net.Ports` (B:render:1141-1148; :978-984) |
| C | `ports:` list, each `"<entry>"`; left out under host or none |
| Q | one `PublishPort=<entry>` per entry; left out under host or none |
| K | container ports (`containerPort` from Container side); Service ports from Host/Container |
| Check | `ParsePort`, `Port.Check`; K: no IP, no range, Host == Container for container ports |
| Answer | E7 (bare N = N:N, M writes both sides), ruling 9 (no error under host/none) |

### Spec.Network

| Column | Value |
| --- | --- |
| A source | none today (A:dockergen writes none) |
| B source | `net.Mode` bridge or host (B:render:1141-1142; :978-979) |
| C | empty: `networks: - default` + top-level `default: driver: bridge`; word: `network_mode` |
| Q | "" -> `Network=private`; word -> `Network=<word>` |
| K | refused when set; no network line |
| Check | `ValidNetwork` |
| Answer | gdu 6, ruling 13, C2 (absent `network.mode` is the deploy's own network: B stops filling in `bridge`; a written `bridge` is docker's shared bridge, and B's sample and docs say so), E4 (`network.mode` and `network.ports` in both tools) |

### Spec.Memory, Spec.CPUSet, Spec.ShmSize

| Column | Value |
| --- | --- |
| A source | new `<platform>.resources.memory` on docker and podman (E12); K resources (A:deploy:758-772) |
| B source | `cb.Mem`/`MonitorMem`, `cb.CPUSet`, `ContainerShmSize` (B:render:1108-1120; :945-957) |
| C | `mem_limit: "<v>"`, `cpuset: "<v>"`, `shm_size: "<v>"`, in that order |
| Q | `PodmanArgs=--memory=<v>`, `PodmanArgs=--cpuset-cpus=<v>`, `ShmSize=<v>` |
| K | Memory -> `limits.memory`; CPUSet, ShmSize refused. Requests/CPU: `kubernetes.Resources` |
| Check | `ValidMemory`, `ValidCPUSet`, `ValidShmSize` |
| Answer | S2, E12 (memory under `resources` on every platform, CPU on Kubernetes in A only, B's `msgNode.cpu` dropped, a bare `m` on Kubernetes refused); P7: request = limit |

### Spec.Ulimits

| Column | Value |
| --- | --- |
| A source | none |
| B source | nofile, memlock, core (B:render:1121-1126; :958-960,1007-1009) |
| C | `ulimits:` map; `soft`/`hard` map when they differ, single integer when equal |
| Q | `Ulimit=<name>=<soft>[:<hard>]` in [Container]; `Limit<NAME>=` in [Service], -1 -> infinity |
| K | refused |
| Check | `Ulimit.Check`, `UlimitNames` |
| Answer | S2 answer, Q1 (settled by S2) |

### Spec.Secrets

| Column | Value |
| --- | --- |
| A source | `secretList(in)` credentials and stores (A:dockergen:135-144; :podmangen:183-198) |
| B source | `ContainerSecrets`, `Target()`, `EnvVar()` (B:render:713-752,1165-1186; :996-998) |
| C | service `secrets:` long syntax, target quoted; top-level `environment: "<EnvVar>"` |
| Q | `Secret=<name>,type=mount,target=<path>`, `%` doubled; no mode, no owner |
| K | secret volume per `Volume`, items key -> basename(Target), `defaultMode: 0400` |
| Check | `Secret.Check`, `CheckComposeVar` on EnvVar, C5 and duplicate checks |
| Answer | S9, C5, C6, the Secrets rule of section 3.1; open P2 (Kubernetes mapping) |

### Spec.Configs

| Column | Value |
| --- | --- |
| A source | app yml, status script, logback, stores-init (A:dockergen:108-116,236-247,299-315) |
| B source | none |
| C | top-level `configs:` with a `content` block and service `configs:` list, target quoted |
| Q | `Secret=<Name>,type=mount,target=<path>`; the tool creates the podman secret (open P5) |
| K | one ConfigMap (data keys sorted, `Block`), mounted by `subPath`, `readOnly: true` |
| Check | `Config.Check`, `CheckInlineText` |
| Answer | S2 (configs own field), C7, C10, S8 |

### Spec.Mounts

| Column | Value |
| --- | --- |
| A source | `in.Stores`, `in.Libs` (A:dockergen:261-269; :podmangen:199-201) |
| B source | `cb.DataDir` bind with `:Z` (B:render:1149-1150; :985) |
| C | `volumes:` `"<src>:<dst>[:ro][,Z]"`, one string quoted |
| Q | `Volume=<src>:<dst>[:opts]`, `%` doubled, `$` refused |
| K | refused (host directory); Kubernetes volumes are `Pod` fields |
| Check | `Mount.Check`, `CheckMountSource` |
| Answer | ruling 9 (zero value `:ro`), C8, Q7, Q8; on a remote docker engine a bind source is refused (flatten answer; where the endpoint test runs is in the catalogue) |

### Tmpfs (not a field, P4)

Not a field (P4, owner 2026-11-27). A's decoded-stores directory (A:dockergen:252-257) is
A's own: A passes `tmpfs:` through the compose extras and states a size (the deep dive:
unsized, docker lets it grow to half the host's memory). Quadlet and Kubernetes have none.
Whether A's image needs a writable `/tmp` under its read-only root (E21) is checked at A's
adoption; if it does, A adds it the same way.

### Spec.Health

| Column | Value |
| --- | --- |
| A source | status script constants (A:dockergen:287-293; :podmangen:214-220; :deploy:679-707) |
| B source | `cb.HealthCheck`, six keys (B:render:1127-1136; :961-975) |
| C | `test: ["CMD", ..]`, `interval: "<n>s"`, `timeout`, `retries: <n>`, `start_period: "<n>s"` |
| Q | `HealthCmd=` argv joined by space, five `Health*` lines, `%` doubled |
| K | readinessProbe exec unless `Deployment.Readiness` is set (then Health is nil) |
| Check | `Health.Check`, `ParseSeconds`; Q: tokens with whitespace or shell characters refused |
| Answer | S6, Q6; open P3 (Kubernetes numbers) |

### Spec.Labels and Spec.Digest

| Column | Value |
| --- | --- |
| A source | le-mode, role, stores digest, deploy digest (A:dockergen:165-182; :podmangen:115-129) |
| B source | deploy-sha256 only (B:render:1174-1175) |
| C | `labels:` last block: five M labels, the tool's, then digest `"${VAR:-}"`; keys quoted |
| Q | five M labels, the tool's `Label="k=v"`, digest `Label="<key>=<sha>"` last |
| K | tool labels are pod template labels; Digest refused; M adds nothing to the template |
| Check | `CheckLabelKey`, `CheckLabelValue`, `CheckComposeVar` on Var and EnvVar |
| Answer | C1, C4, S1 (reserved prefixes), todo 10, Q2 (the on-disk compare stays beside the digest label) |

### compose.File.Project

| Column | Value |
| --- | --- |
| A source | `in.Docker.ProjectName`, line omitted when empty (A:dockergen:102-104) |
| B source | `ComposeProject(cb.Name)` folding (B:render:1030-1041,1063) |
| C | `name: "<project>"` always, first line; also passed as `-p` |
| Q, K | not applicable |
| Check | `compose.ValidProject` |
| Answer | C9 |

### quadlet.Unit.Description, Rootful

| Column | Value |
| --- | --- |
| A source | text (A:podmangen:142), `WantedBy=default.target` (:230), comment (:139) |
| B source | hostname and node type (B:render:896), `c.Podman.WantedBy` (:1013), no comment |
| Q | `Description=`, `WantedBy`, header comment |
| Check | description: no control, no line break, no trailing backslash |
| Answer | Q3 layout; Q10 (every unit starts with one comment line naming the tool, from `Identity.Tool`), Q11 |

## 5. The writer

One writer package, no YAML library. Output is deterministic and LF only; each artifact ends
in exactly one newline.

### 5.1 YAML sinks (compose, Kubernetes)

- Layout: two spaces per level. A list under a key is indented two further than the key.
  A map inside a list item nests under its first key. Block sequences only; flow lists only
  for compose `entrypoint`, `command` and `healthcheck.test` (as both tools write them).
- Scalars: every text value a tool supplies is a double-quoted scalar (ruling 9). Booleans,
  integers the sink requires as numbers, and M's own keyword constants are bare. A bare
  word that a tool chooses goes through `Word` against a closed list.
- Keys (S5): every key that is data is quoted, M's own label keys included. A keyword key
  (`image`, `restart`, `apiVersion`, `spec`) stays bare. Compose additionally refuses `$`,
  spaces and control characters in data keys. On Kubernetes a bare field name must be plain
  letters and digits and not a YAML 1.1 boolean or null word.
- `$`: compose doubles `$` in every value (`QuoteDollar`) except in the two interpolations
  M writes itself (`Label.Var` and `DigestLabel`), written `"${VAR:-}"` with the `$` single.
  Kubernetes doubles `$` in env values, `command`, `args` and probe commands (K13, with a hand
  check on a cluster before the tag); it does not double elsewhere.
- Invalid UTF-8 is refused (it cannot be quoted safely). Control characters come out as
  `\xNN` or `\uNNNN` escapes of `strconv.Quote` for valid input.
- Block text: compose `content: |` writes the bytes as given (C10) with the chomping marker
  the text needs (`|-`, `|`, `|+`) and an indent marker when the first non-empty line starts
  with a space; empty lines carry no indent. Kubernetes `|` never uses `|+`: text with
  several trailing newlines is written as a quoted string (K "no decision needed"). The two
  differ on purpose; see P8. Compose block text has every `$`
  doubled, as any compose value does (R2's reading of C10: the bytes are the tool's, the sink's
  escaping is the writer's); Kubernetes block text is left alone.
- Empty values: an empty map or list is dropped in M's own documents; a key never holds null.
  An empty environment value is `"KEY": ""`, never a bare key. A tool-built tree with an empty
  node is an error.
- Metadata order on every Kubernetes object: name, namespace, labels, annotations. Data keys
  of a Secret or ConfigMap sorted bytewise.

### 5.2 Quadlet lines (Q3, Q4)

Per-key write forms (Q4, answered: the bare keys are exceptions to ruling 9):

| Form | Keys | Escapes |
| --- | --- | --- |
| Quoted whole | `Environment="K=V"`, `Label="k=v"` | `\`, `"`, `%`, `$` and control chars escaped |
| Bare, `%` doubled | Image, ContainerName, HostName, User, Group, Network | value grammar |
| Bare, `%` doubled | PublishPort, Volume, Secret, Health*, Description, PodmanArgs | same |
| Bare, `%` doubled | ShmSize, Ulimit (no space, quote, backslash or `$`) | same |
| Refused | an extra whose value holds whitespace, quote, backslash or `$` | error names the key |

If `Label="k=v"` is not read like `Environment="k=v"` at podman 4.5 (Q4 fallback), `Label`
goes bare under a no-space grammar; one switch in the writer.

Line order (Q3): sections `[Unit]`, `[Container]`, `[Service]`, `[Install]`, one blank line
between them, none after the last; `[Service]` always written. Header comment (one line, Q10) first.
`[Unit]`: Description, Wants, After. `[Container]`: Image, Pull, ContainerName,
HostName, the five Label lines, the tool's labels, the digest label is held back to the end
of the label group, User, Group, NoNewPrivileges, DropCapability, ReadOnly, Network,
PublishPort..., Memory/CPU `PodmanArgs`, ShmSize, Ulimit..., Environment (TZ first), Secret
(secrets, then configs), Volume, Health* (Cmd, Interval, Timeout, Retries, StartPeriod), then
extras. `[Service]`: Restart, then `Limit*` lines. `[Install]`: WantedBy. Where Q3's sketch and
S2 differ (Ulimit and ShmSize are now own fields), the own-field lines sit in the resources
group after the ports, ahead of Environment.

### 5.3 Compose order (C1)

Top level: `name`, `services`, `networks` (only when the service uses the project network),
`configs`, `secrets`; a block with nothing in it is left out. Service: image, pull_policy, container_name,
hostname, user, restart, posture block, read_only, mem_limit, cpuset,
shm_size, ulimits, extras, networks or network_mode, ports, environment, secrets, configs,
volumes, healthcheck, labels. A's `entrypoint:`, `command:` and `tmpfs:` come in the extras (P4). The labels end the block, so the digest label closes the
service. The network block has no labels (C4). `name:` is always written (C9).

### 5.4 Kubernetes order

Document: apiVersion, kind, metadata, body. Pod spec: serviceAccount (not written: the key is
refused, `automountServiceAccountToken: false` is a constant), imagePullSecrets, nodeSelector,
tolerations, affinity, securityContext, initContainers, containers, volumes, extras. Container:
name, image, imagePullPolicy (when set), command, args, securityContext (constant three
lines, plus `readOnlyRootFilesystem` when read-only), ports, env, volumeMounts, probes,
resources, extras. Documents are joined by one `---` line; the stream ends in one newline.
No `hostNetwork` line anywhere (ruling 13).

### 5.5 Errors (S8)

`FieldError` names the sink, the object, the field path, the rule. It echoes a value only for
a name, a key, a fixed-list word or a number. Never echoed: Secret data, registry login,
ConfigMap data, inline text (name and offset only), environment values, command and
arguments, label and annotation values, the image reference, extras, custom-body scalars.
Example: `compose: service "broker": env "JAVA_TOOL_OPTIONS": the value holds a control
character; the value is not shown`. Every problem is collected; nothing partial is returned.

## 6. Cross-field checks

All in `workload.CheckTargets` and `Spec.Check(Sink)`, run at load and again at render.

1. Mount nesting (C5). A secret or config target must not equal or lie under a mount target,
   and the reverse. Compose, quadlet and Kubernetes (volume mounts) all run it. A's stores
   tmpfs is an extra (P4), so A keeps its own check for the existing stores it binds inside.
2. Duplicates. Targets across secrets, configs and mounts; environment names (and TZ
   against `Timezone`); label keys (including M's five and the digest key); secret names;
   compose `Secret.EnvVar` values and the digest variable and every `Label.Var` (one
   namespace, C6); Kubernetes volume names and container names; ulimit names.
3. Variable names (C6). `Secret.EnvVar`, `DigestLabel.EnvVar` and `Label.Var` refuse `PATH`,
   `HOME`, `IFS`, `SHELL`, `BASH_ENV` and any `LD_`, `DOCKER_` or `COMPOSE_` prefix.
4. Ports under host or none (ruling 9): the renderer leaves them out; its doc says so; no
   error. A tool keeps its own list check at load (B does).
5. Network and ports on Kubernetes: `Network` set is refused; a published `Host` that is not
   the container port is refused (no `hostPort`).
6. Sink refusals: `HostName`, `CPUSet`, `ShmSize`, `Ulimits`, `Mounts`, `Network`,
   `Label.Var`, `Digest` on Kubernetes; `Label.Var` on quadlet; `Secret.EnvVar` is required on compose and `Secret.Key`, `Secret.Volume` on
   Kubernetes (each error names the field).
7. Run-as against root: a stated `0:0` is written, never refused (both tools accept it).
8. Health: Kubernetes `Readiness` set together with `Health` is an error (one source).
9. Extras deny lists (S1, with its 2026-11-17 additions), each naming the version it was
   checked against in a test:
   - compose service: the keys M writes, keys that reach the host or another container
     (cap_add, devices, sysctls, pid, ipc, uts, userns_mode, cgroup, runtime, volumes_from,
     group_add, links ...), keys that pull in files or services (build, extends, env_file,
     depends_on, deploy, post_start, pre_stop, provider, develop, models), plus profiles,
     scale, annotations, logging, oom_score_adj, cpu_rt_*, credential_spec. `pull_policy` is
     now a key M writes (so is quadlet `Pull`), and `imagePullPolicy` was one already.
   - quadlet `[Container]` only (Q1): PodmanArgs, GlobalArgs, ContainersConfModule, the
     privilege keys (AddCapability, AddDevice, Sysctl, SecurityLabel*, SeccompProfile, Unmask,
     UserNS, UIDMap, GIDMap, SubUIDMap, SubGIDMap, GroupAdd, Mount, EnvironmentHost,
     EnvironmentFile, Pod, Rootfs), every key M writes (including the memory, cpu, shm,
     ulimit and health keys now), AutoUpdate, Annotation, CgroupsMode, LogDriver, LogOpt,
     HealthLogDestination. Case-insensitive; a key holding `=`, space or line break refused.
   - Kubernetes pod: hostNetwork, hostPID, hostIPC, hostUsers, shareProcessNamespace,
     securityContext, volumes, ephemeralContainers, nodeName, runtimeClassName,
     priorityClassName, containers, initContainers, serviceAccountName, serviceAccount,
     automountServiceAccountToken, schedulerName, resourceClaims, resources, os, restartPolicy,
     schedulingGates, activeDeadlineSeconds, priority, tolerations, imagePullSecrets.
   - Kubernetes container: securityContext, envFrom, volumeDevices, env, volumeMounts, ports,
     lifecycle, startupProbe, livenessProbe, readinessProbe, restartPolicy, and every key M
     writes (name, image, imagePullPolicy, command, args, resources).
   - Label and annotation keys a tool passes, every sink, by prefix: app.kubernetes.io/
     managed-by, services.solace.com/, deploy.solace.com/, com.docker.compose.,
     io.containers., io.podman., run.oci., io.kubernetes.cri-o.,
     container.apparmor.security.beta.kubernetes.io/, seccomp.security.alpha.kubernetes.io/,
     and on a Namespace pod-security.kubernetes.io/.
   - artifacttest reads the same lists (from internal/posture) and matches a key with its
     value on Kubernetes.
10. Pull policy and Service type: the word is checked by `ValidPullPolicy(Sink, word)` and
    `ValidServiceType(word)`; the Service type has no default and no `nodePort` field.
11. Object size (K16): the apply method refuses a rendered document over the size limit;
    remove does not check; `generate` warns, naming the object and its size.
12. Extras values: text, integers, booleans, lists and maps; no raw text; text gets the UTF-8
    and control refusal; integers 64-bit; no `$` interpolation can arise (compose `Str` always
    doubles).

## 7. Coverage of what A and B write today

"Field" is a Spec field or a sink-owned input of 3.1. "Stays" means the line is not a platform
artifact line, or it is content, and the reason is stated.

### 7.1 A, compose (internal/dockergen/dockergen.go)

| Line | What | Home |
| --- | --- | --- |
| :103 | `name:` | `compose.File.Project` (C9: now always) |
| :105 | `services:` | writer |
| :108-116 | top-level `configs:` and entries | `Spec.Configs` (derived block) |
| :124 | top-level `secrets:` | `Spec.Secrets` (derived block) |
| :152 | service key | `Spec.Name` |
| :153 | `image:` | `Spec.Image` |
| :154 | `container_name:` | `Spec.Name` |
| :159-160 | `entrypoint:` `command:` | A's compose extras (P4) |
| :169-179 | labels le-mode, role, stores digest | `Spec.Labels` (digest via `Var`) |
| :182 | deploy digest label | `Spec.Digest` |
| :184 | `restart:` | `Spec.Restart` (empty now writes always) |
| :192 | `user:` | `Spec.RunAs` |
| :194 | posture block | M constant |
| :196-199 | `ports:` | `Spec.Ports` |
| :205-222 | `environment:` TZ, JAVA_*, SYSLOG_* | `Spec.Timezone`, `Spec.Env` |
| :229 | service `secrets:` | `Spec.Secrets` |
| :236-247 | service `configs:` | `Spec.Configs` |
| :252-257 | `tmpfs:` | A's compose extras (P4), with a size |
| :288-293 | `healthcheck:` | `Spec.Health` |
| :262-267 | `volumes:` stores then libs | `Spec.Mounts` |
| :299-315 | inline config block text | writer frames `Spec.Configs[].Content`; payload stays in A |
| (none) | `read_only:`, network lines | new: `Spec.Root`, `Spec.Network` |

Stays in A: the payloads (application.yml, status script, logback XML, stores-init script),
the le-mode/role rule, the Java options merge, and the store selection.

### 7.2 A, quadlet (internal/podmangen/podmangen.go)

| Line | What | Home |
| --- | --- | --- |
| :139 | header comment | written by M on every unit (Q10; text becomes M's) |
| :141-144 | `[Unit]`, Description, After, Wants | `Unit.Description`; M constants |
| :147-149 | `[Container]`, Image, ContainerName | `Spec.Image`, `Spec.Name` |
| :150-152 | `Label=` le-mode, role | `Spec.Labels` |
| :160-163 | `User=` `Group=` | `Spec.RunAs` |
| :164 | posture | M constant |
| :165-167 | `PublishPort=` | `Spec.Ports` |
| :168-182 | `Environment=` TZ, JAVA_*, syslog | `Spec.Timezone`, `Spec.Env` |
| :183-189 | `Secret=` credentials | `Spec.Secrets` |
| :193-198 | `Secret=` app yml, stores, status, logback | `Spec.Configs`, `Spec.Secrets` (P5) |
| :199-201 | `Volume=` libs | `Spec.Mounts` |
| :214-220 | `Health*` | `Spec.Health` (written only when set) |
| :223-227 | `[Service]` `Restart=` | `Spec.Restart` ([Service] now always written) |
| :229-230 | `[Install]` WantedBy | `Unit.Rootful` (Q11) |
| :232 | file name | `File.Name` |

### 7.3 A, Kubernetes (internal/deploy/deploy.go)

| Line | What | Home |
| --- | --- | --- |
| :348 | Namespace | `kubernetes.Namespace` (replaces `kube.NamespaceManifest`) |
| :377-386 | credentials Secret (stringData) | `kubernetes.Secret` (data only, K2; built only when the credential fields are set) |
| :392-401 | stores Secret | `kubernetes.Secret` |
| :409-416 | image-pull Secret | `kubernetes.Secret` via `RegistrySecret` (built when `image.user` and `image.pass` are set and no `existing`) |
| :426-458 | PV and PVC | `kubernetes.PersistentVolume`, `PersistentVolumeClaim` |
| :477 | `---` join | `writer.Join` |
| :483-489 | name-only Secret for remove | `Manifest.Refs` (K3); remove deletes every derived Secret name |
| :496-527 | ConfigMap | `kubernetes.ConfigMap` from `Spec.Configs` |
| :534-546 | Deployment, replicas, selector | `kubernetes.Deployment` |
| :549-560 | pod labels | `Spec.Labels` |
| :565 | `automountServiceAccountToken: false` | M constant |
| :569-572 | imagePullSecrets | `Pod.ImagePullSecrets` |
| :577-587 | nodeSelector | `Pod.NodeSelector` |
| :169-203 | tolerations | `Pod.Tolerations` |
| :205-238 | affinity (verbatim block + spread) | `Pod.Affinity` (K12) |
| :249-261 | pod securityContext | pod identity fields (E22, P11); seccomp constant |
| :267-274 | container securityContext | M constant |
| :591-606 | libs-download init container | `Deployment.Init` (a `workload.Spec`) |
| :608-613 | container name, image, ports | `Spec.Name`, `Spec.Image`, `Spec.Ports` |
| :617-641 | env TZ, JAVA_*, syslog | `Spec.Timezone`, `Spec.Env` |
| :643-678 | volumeMounts | `Deployment.Mounts` and the secret/config mapping |
| :682-707 | liveness, readiness | `Deployment.Liveness`, `Readiness` (open P3) |
| :709-773 | resources | `kubernetes.Resources` |
| :711-755 | volumes (configMap, secret, projected, claim, emptyDir) | `Pod.Volumes` |
| :775-788 | Service | `kubernetes.Service` (type word required; A passes `ClusterIP` unless set) |
| :806-823 | `Written` list | `Manifest.Objects` (K15) |

Stays in A: the libs download shell text, which claim source applies, the spread choice, and
the content of every payload.

### 7.4 B, compose (internal/render/render.go)

| Line | What | Home |
| --- | --- | --- |
| :1063 | `name:` | `compose.File.Project` (B's folding retires, C9) |
| :1070-1073 | services, key, image, container_name | `Spec.Name`, `Spec.Image` |
| :1074 | `hostname:` | `Spec.HostName` |
| :1075 | `user:` | `Spec.RunAs` (B refuses empty itself) |
| :1076 | `restart: always` | `Spec.Restart` (empty) |
| :1090 | posture | M constant |
| :1108-1117 | cpuset, mem_limit (monitor swap) | `Spec.CPUSet`, `Spec.Memory`; swap stays in B |
| :1120-1126 | shm_size, ulimits | `Spec.ShmSize`, `Spec.Ulimits` (B passes no extras) |
| :1127-1136 | healthcheck | `Spec.Health` |
| :1141-1148 | network_mode host or ports | `Spec.Network`, `Spec.Ports` |
| :1149-1150 | data bind `:Z` | `Spec.Mounts` (Writable, RelabelPrivate) |
| :1154-1160 | environment | `Spec.Env` (EnvPairs stays in B) |
| :1174-1175 | deploy digest label | `Spec.Digest` |
| :1181-1186 | service and top-level secrets | `Spec.Secrets` |

### 7.5 B, quadlet (internal/render/render.go)

| Line | What | Home |
| --- | --- | --- |
| :895-898 | `[Unit]`, Description, Wants, After | `Unit.Description`, M constants |
| :900-903 | Image, ContainerName, HostName | `Spec.Image`, `Spec.Name`, `Spec.HostName` |
| :904-907 | User, Group (names allowed today) | `Spec.RunAs` (numeric only, E5) |
| :924 | posture | M constant |
| :945-954 | `PodmanArgs` cpuset, memory | `Spec.CPUSet`, `Spec.Memory` |
| :957-960 | ShmSize, Ulimit lines | `Spec.ShmSize`, `Spec.Ulimits` |
| :961-975 | Health* | `Spec.Health` |
| :978-984 | Network=host or PublishPort | `Spec.Network`, `Spec.Ports` |
| :985 | `Volume=...:Z` | `Spec.Mounts` |
| :986-988 | Environment | `Spec.Env` |
| :996-998 | Secret lines | `Spec.Secrets` |
| :1000-1009 | `[Service]` Limit lines | `Spec.Ulimits` (-1 -> infinity) |
| :1010 | `Restart=always` | `Spec.Restart` |
| :1012-1013 | `[Install]` | `Unit.Rootful` |
| (none) | ownership labels, digest label | new: identity labels, `Spec.Digest` |

### 7.6 B, Kubernetes (internal/k8s, internal/render)

| Line | What | Home |
| --- | --- | --- |
| B:secrets:35-66 | `secretManifest.render` | `kubernetes.Secret` (annotations via `Meta`) |
| B:secrets:93,139 | admin, additional-users Secrets | `kubernetes.Secret` (content stays in B) |
| B:secrets:168-229 | TLS Secret, digest annotation | `kubernetes.TLSSecret`; file reads stay |
| B:secrets:249-292 | registry Secret and `regcred` | `kubernetes.RegistrySecret` (K10: `regcred` unlabelled, by name; K11) |
| B:prep.go:418-426 | `joinManifests` | `writer.Join` |
| B:operator.go:411-424 | `joinYAMLDocs` | `writer.Join` |
| B:render:78-226 | PubSubPlusEventBroker body | `kubernetes.Custom` body tree (K5); content stays |
| operator bundle template | vendor YAML | `kubernetes.Raw` (K9) |
| `kube.NamespaceManifest` use | Namespace | `kubernetes.Namespace` |

B has no Deployment of its own: the operator makes it. So the container description does not
reach B's Kubernetes path; B's Kubernetes artifacts use the document types only.

Stays in B: `EnvPairs` content, `/etc/sysctl.d/99-solace.conf` (not a container artifact; it
moves to `outfile` with the unit), the operator bundle text, and the `kubectl set env` command.

### 7.7 Lines no field covers

None found. Lines the tools write today with no home in an earlier model are now fields:
A's `restart` omission (now always), A's dropped `[Service]` (now always written), B's empty
`user:` and `name: ""` (now refused), B's `Network=host` under ports (ports left out).

## 8. Open points

Settled points are written into the body above. This section lists them in one line each,
then the points the sources leave undecided, marked "open".

### 8.1 Settled since the first draft

- C2, E4: absent `network.mode` is the deploy's own network (docker project network, podman
  `private`); `bridge` written is docker's shared bridge. `Spec.Network` is the field.
- E12: memory under `resources` on every platform; Kubernetes CPU in A only; bare `m` refused.
- E21: both tools have `readOnlyRoot` (A true, B false); `Spec.Root` has no default.
- E22: Kubernetes identity in pod and container fields, in both tools; the mapping is P11.
- Q7: mount source is an absolute path with no `:` (one drive-letter root on compose, C8),
  whitespace, control, invisible, `$`, quotes, backtick or backslash; `%` allowed and doubled
  on quadlet.
- Q8: `Mount.Relabel` is none, `z` or `Z`; forgotten means none; A none, B `Z`.
- Q9: no value matching by name; secrets go through `Spec.Secrets`; `artifacttest` checks units.
- Q10, Q11: one comment line naming the tool on every unit, no version; `WantedBy` is
  `default.target` rootless and `multi-user.target` rootful.
- K1: `imagePullPolicy` is `Spec.PullPolicy` (2026-11-27), checked by M, written only when set.
- K2, K3, K5: Secrets are `data` only from raw bytes; `Manifest.Refs` is a name-only document
  and refuses a Namespace; the custom resource body is a tree with numbers as numbers.
- K4: a CRD is never deleted or modelled; B drops `--delete-crd`.
- K6, K18: every workload kind is typed with hardening built in; nothing else ahead of need,
  but the Service types are built (section 3.1).
- K7: ownership labels and annotations on every object, PV included, apart from the bundle
  and `regcred` (K9, K10); A's PV warning reads size and path with one `kubectl get`.
- K8: no Secret in environment variables and no `envFrom`; B's additional users keep
  `spec.extraEnvVarsSecret` as the one exception the operator's CRD forces.
- K9, K10: the operator bundle stays `Raw` and unlabelled; `regcred` is unlabelled and handled
  by name; a Namespace in a remove stream is refused.
- K11, K12, K13, K14, K15, K17: `kubernetes.DockerHubKey` when the registry is empty; A's
  affinity block becomes a tree; `$` doubled in exec probes; `defaultMode: 0400` as an octal
  literal; `Manifest.Objects` with a cluster-wide flag; NFS only for volumes.
- K16: the apply method refuses an object over the size limit; `generate` warns (6, item 11).
- N11 to N15: the character checks are `CheckText` over `execguard.IsControl` and
  `IsInvisible`; `IsControl` takes U+0085; `CheckMountSource` stands for N14's host-path
  check; `envref.ValidName` keeps its name.
- N25, N35, N36, N37: line helpers private; posture in `internal/posture`; the `Check*` twins
  private; `HostNetwork` removed for `AssertNetwork`; `artifacttest` at the top level.
- N30: `containercli.ValidName` for names read back, `workload.CheckName` for names a tool
  creates.
- P1: settled by R1: workload imports neither `kube` nor `containercli` (section 2); the move
  of `kube.NamespaceManifest` is N58's.
- P6: settled by the Q5 answer: one helper, `outfile`, writes `generate` output (umask mode)
  and the installed unit (0600), and refuses a symbolic link in both.
- P2 (2026-11-27): falls away. Every Secret mount lists its keys (the Secrets decision and
  the key check before deploy), so no tool mounts a whole Secret.
- P3 (owner, 2026-11-27): one health check on every platform, the same command and the
  same timings. `Spec.Health` maps to the compose healthcheck, the quadlet Health* keys and
  the Kubernetes readiness probe (period = interval, timeoutSeconds = timeout, stated since
  Kubernetes defaults it to 1s, failureThreshold = retries). A's timings become 10s interval,
  5s timeout, 3 retries, 60s start period everywhere (its docker and podman checks were
  30s/10s). Liveness stays the tool's own: A passes a TCP check on the management port, so a
  flapping downstream never restarts the connector. B's broker probes are the operator's.
- P4 (owner, 2026-11-27): entrypoint, command and tmpfs are not common fields. Only A's
  docker store script needs them, and the deep dive keeps that script
  (docs/review/2026-11-27-docker-stores-assessment.md), so A passes them through the compose
  extras, the tmpfs with a size. Quadlet and Kubernetes have none. Carried to A's adoption:
  whether its image needs a writable `/tmp` under the read-only root (unverified).
- P5 (2026-11-27): `Spec.Configs` on quadlet become `Secret=` lines, the podman secret
  created from `Content`, as A does today.
- P7 (owner, 2026-11-27): one value per resource, written as both request and limit
  (guaranteed QoS, as A today). No request keys. `Spec.Memory` and `kubernetes.Resources`
  are one value, not two.
- P8: settled by C10 (compose `|-`, `|`, `|+`) and the settled Kubernetes rule (no `|+`; a
  quoted string instead).
- P9 (owner, 2026-11-27): Kubernetes restart words in Kubernetes spelling, checked per kind:
  Deployment, StatefulSet, ReplicaSet and DaemonSet `Always` only; Job and CronJob
  `OnFailure` and `Never`; Pod all three. Empty writes nothing. R2 replaced R1's lowercase
  `always` (built: `ValidRestart` takes the three words as a union, `ValidKubeRestart` the kind).
- P11 (owner, 2026-11-27, per the R2 plan Part 4; built with `Spec` and the `kubernetes` package,
  nothing in R2): (a) Kubernetes identity reuses `workload.RunAs`, no new type. `Spec.RunAs`
  (uid:gid) feeds the container block, `runAsUser` = UID and `runAsGroup` = GID.
  `kubernetes.Pod` carries a second `workload.RunAs` built with `RunAsIDs(uid, fsGroup)`: its
  UID is the pod `runAsUser` and its GID the pod `fsGroup`, and the doc names the mismatch. The
  word `default` is the zero `RunAs` in all four fields and writes nothing; unset is the tool's
  default, filled in before the call (A `1000:0`; B 1000001 and 1000002 on Kubernetes).
  `runAsNonRoot: true` follows a stated container uid other than 0, and a stated 0 writes none;
  B's refusal of a stated 0 on Kubernetes stays in B. (b) `FieldError` gains no sink or object
  member: each renderer prefixes its messages with `fmt.Errorf("compose: service %q: %w", name,
  err)`, so `errors.As` still reaches the `*FieldError` and the example of 5.5 reads as written.
  The writer's `*Error` follows the same rule (path and rule only).
- P10 stays a working-name note: package names `containerart/writer`, `containerart/workload`,
  `containerart/kubernetes` and `containerart/outfile` are working names (R1 shipped
  `containerart/workload`).

### 8.2 Still open

- Build notes, not decisions: the word lists of `ValidPullPolicy` and the deny lists are
  checked against the compose, podman and Kubernetes references at build and name the
  versions in a test (S1, env-layout section 4). The compose position of `pull_policy` and
  the quadlet position of `Pull=` are fixed by the goldens.
  R2 names `Docker Compose 2.40` and `Podman 5.6` in `posture.ComposeCheckedAgainst` and
  `QuadletCheckedAgainst`; the lists are recalled from the references and not yet run on an
  engine, so the hand check (docs/smoke.md) confirms or corrects both names.
