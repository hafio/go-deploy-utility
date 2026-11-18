# One container description and one writer (consolidation draft)

Status 2026-11-18: Q7 to Q11, K1 to K18 and every N entry are now answered; see section 11
of the 2026-11-15 design note, which wins over this draft. Notable changes: K6 types every
workload kind; K10 leaves `regcred` unlabelled; K18 builds nothing ahead of need; E22 moves
Kubernetes identity to pod and container fields; E23 makes every image key a full block.

Draft for item 3 of section 6 of docs/review/2026-11-15-surface-review.md. Read-only inputs:
docs/design/2026-11-15-shared-deploy-api.md sections 8-11 (10 wins over 3-8; 11 wins over
3-8; 10 wins over 11), docs/review/2026-11-15-surface-review-questions.md (S, C, Q1-Q6
answered; the rest "open <ID>" with the Recommended choice), findings section 3, and the
current code of M, A and B.

Words: M = go-deploy-utility. A = solmq-conn-util (repo solace-ibmmq-connector-helper).
B = solace-util (repo solace-cnt-scripts). Citations "A:file:line" and "B:file:line" are
read from the checkouts of 2026-11-17: A:dockergen = internal/dockergen/dockergen.go,
A:podmangen = internal/podmangen/podmangen.go, A:deploy = internal/deploy/deploy.go,
B:render = internal/render/render.go, B:secrets = internal/k8s/secrets.go.

Env-file key names are NOT used as the source of a field (E4 and the flatten are reopened).
Each "A source" and "B source" names the Go-side setting that feeds the field today.

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
containerart/artifacttest  checks on rendered text; imports renderers and workload,
                         never the reverse
```

Import graph (an arrow reads "imports"; no cycle is possible because every arrow points
down this list):

```
artifacttest -> compose, quadlet, kubernetes, workload, writer(no), testing
compose      -> workload, writer
quadlet      -> workload, writer
kubernetes   -> workload, writer
workload     -> execguard, envref, containercli (ValidName), kube (ValidLabel)
writer       -> standard library only
outfile      -> standard library only
```

Rules that keep it acyclic:
- workload and writer import no renderer and no artifacttest. A renderer never imports
  artifacttest (it imports `testing`).
- workload imports `kube` for `ValidLabel` and `containercli` for `ValidName`, so neither
  may import workload or any renderer. `kube.NamespaceManifest` therefore leaves `kube`
  (it becomes `kubernetes.Namespace`); see open point P1.
- The Posture constants and the deny lists (S1) are read by both renderers and
  artifacttest. With N25 as recommended they live in `containerart/internal/posture`, which
  the renderers and artifacttest import; the existing `compose.Posture` and
  `quadlet.Posture` go away as exported names (open N25).
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
- quadlet: `Unit{Spec workload.Spec, Description, Rootful, Header, Extra []Setting}` and
  `Render() (File, error)`. Quadlet only facts: podman network words, unit layout, `Ulimit=`
  and `Limit*=` lines, digest label two-pass, the `[Container]` extras deny list.
- kubernetes: `Manifest{Owner, Docs}`, `Meta`, `Namespace`, `Secret`, `ConfigMap`, `Service`,
  `PersistentVolume`, `PersistentVolumeClaim`, `Deployment` (takes `workload.Spec` plus the
  Kubernetes-only pod fields), `Custom`, `Raw`, `Refs`, `Objects`, `Join`. Pod and container
  extras and their deny lists.
- outfile: `Write` (umask mode, for `generate` output) and `WriteMode` (explicit mode, for
  the installed unit).
- artifacttest: today's checks, plus the stated-network check (N36), the label check, the
  Kubernetes posture check (K6) and the deny lists read from internal/posture.

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
	HostName   string     // "" writes nothing. RFC 1123 (S12). Refused on Kubernetes.
	RunAs      RunAs      // zero value = not stated = the image's own user; writes nothing.
	Restart    string     // "" writes "always" (E6, ruling 9). Words per sink, see ValidRestart.
	Root       RootFS     // required (RootUnset is an error). Read-only writes the line;
	                      // writable writes nothing (ruling 9, gdu 6).
	Entrypoint []string   // nil writes nothing. Compose entrypoint, Kubernetes command.
	Command    []string   // nil writes nothing. Compose command, Kubernetes args.
	Timezone   string     // "" writes nothing. IANA name (S13). Written as TZ, first.
	Env        []EnvVar   // nil writes nothing. Given order kept. TZ here beside Timezone
	                      // is an error (duplicate).
	Ports      []Port     // nil writes nothing. Dropped under host or none (ruling 9).
	Network    string     // "" = the sink's non-host default (compose project network,
	                      // Network=private); refused on Kubernetes (ruling 13).
	Memory     string     // "" writes nothing. Grammar per sink (ValidMemory).
	CPUSet     string     // "" writes nothing. Refused on Kubernetes.
	ShmSize    string     // "" writes nothing. Refused on Kubernetes.
	Ulimits    []Ulimit   // nil writes nothing. Refused on Kubernetes (S2).
	Secrets    []Secret   // nil writes nothing. Files at absolute targets.
	Configs    []Config   // nil writes nothing. Inline content (see Config).
	Mounts     []Mount    // nil writes nothing. Host bind mounts; refused on Kubernetes.
	Tmpfs      []Tmpfs    // nil writes nothing. Compose only; refused on quadlet (extras
	                      // Tmpfs= passes there) and on Kubernetes.
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
// written. Repository is the path only, never a host.
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
	Relabel  Relabel // none (zero), z, Z (open Q8)
}

type Relabel int

const (
	RelabelNone Relabel = iota
	RelabelShared  // z
	RelabelPrivate // Z
)

// Tmpfs is a compose tmpfs entry.
type Tmpfs struct {
	Target string      // absolute, no ':'
	Mode   fs.FileMode // permission bits only, written 0700 style
	Owner  string      // "" or "uid:gid" digits, both halves
}

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

type EnvVar struct{ Name, Value string } // Name via envref.ValidName + CheckEnvName

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
| `Image.Check`, `ParseImage`, `Image.String` | parser | grammar above; tag or digest required |
| `ParseRunAs`, `RunAsIDs`, `RunAs.Stated/UID/GID/String` | parser | numeric uid:gid, both halves |
| `ParsePort`, `Port.Check`, `Port.String` | parser | port grammar; range lo <= hi; 1..65535 |
| `Mount.Check(Sink)`, `CheckMountSource` | check | absolute; one drive-letter root; no ':' else |
| `Tmpfs.Check` | check | absolute, no ':', permission bits, owner digits |
| `Secret.Check(Sink)` | check | target absolute; per-sink required fields; names |
| `Config.Check(Sink)` | check | target absolute; `CheckInlineText` on Content |
| `CheckInlineText` | check | C7: no CR, NEL, LS, PS, control (tab, LF ok), DEL, C1, BOM; UTF-8 |
| `Health.Check(Sink)` | check | S6 rules; on quadlet Q6 shell-character refusal |
| `ParseSeconds` | parser | a Go duration string that is a whole number of seconds |
| `Ulimit.Check`, `UlimitNames`, `ParseLimit` | check | 13 names, soft <= hard, -1, infinity |
| `ValidHostName` | check | RFC 1123 (S12) |
| `ValidTimezone` | check | IANA name grammar (S13) |
| `CheckEnvName` | check | C6 name refusals on top of `envref.ValidName` |
| `CheckLabelKey`, `CheckLabelValue` | check | qualified name; reserved prefixes (S1 answer) |
| `ReservedLabelPrefixes` | list | M's own prefixes and the engine prefixes of S1 |
| `ValidRestart(Sink, word)` | check | compose/quadlet words; Kubernetes: "" or always only |
| `ValidNetwork(Sink, word)` | check | docker and podman word lists of gdu 6; Kubernetes: empty |
| `ValidMemory(Sink, s)`, `ValidCPUSet`, `ValidShmSize` | check | per-sink grammar (E12 open) |
| `CheckName` | check | container name: DNS label <= 63 (rulings 11, 13; N30) |
| `CheckText` | check | UTF-8, controls, invisibles |
| `CheckTargets` | check | duplicates and nesting (C5); section 6 |
| `FieldError` | type | Sink, Object, Field, Rule, Cause; `Error()` never holds free text |

`FieldError` carries a `Rule` word and a `Field` path such as
`env["JAVA_TOOL_OPTIONS"]`. A value is echoed only for names, keys, fixed-list words and
numbers (S8). A render returns `errors.Join` of every problem found, nothing partial.

Exported names in package workload, flat:
`Sink`, `SinkCompose`, `SinkQuadlet`, `SinkKubernetes`, `Spec`, `Identity`, `RootFS`,
`RootUnset`, `RootReadOnly`, `RootWritable`, `Image`, `ParseImage`, `RunAs`, `RunAsIDs`,
`ParseRunAs`, `Port`, `ParsePort`, `Mount`, `Relabel`, `RelabelNone`, `RelabelShared`,
`RelabelPrivate`, `CheckMountSource`, `Tmpfs`, `Secret`, `Config`, `CheckInlineText`,
`EnvVar`, `Label`, `DigestLabel`, `Health`, `ParseSeconds`, `Ulimit`, `UlimitNames`,
`ParseLimit`, `ValidHostName`, `ValidTimezone`, `CheckEnvName`, `CheckLabelKey`,
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
	Rootful     bool      // false: WantedBy=default.target; true: multi-user.target (open Q11)
	Header      bool      // writes the one comment line (open Q10)
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
	Pod        Pod              // placement, volumes, pull secrets, probes (see below)
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
(K2), name-only references for a remove (K3), data keys quoted (S5), `defaultMode: 0400`
(S9, K14), the Docker Hub key constant (K11), one `---` joiner, one `Objects` list (K15).
They take `Meta` (name, namespace, labels, annotations) and write the ownership labels and
the lib and version annotations themselves (ruling 10). They are not part of the container
description: a Secret or ConfigMap holds data, not a container.

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
| Answer | gdu 1, S4 (solution, developed-by optional), ruling 10 (lib annotation), C4 |

### Spec.Name

| Column | Value |
| --- | --- |
| A source | `inst.Name` (A:dockergen:152,154; A:podmangen:149; A:deploy:537,608) |
| B source | `cb.Name` (B:render:1071,1073; :902) |
| C | service key and `container_name`, both quoted |
| Q | `ContainerName=`; `<Name>.container` must fit 255 bytes |
| K | container `name`; Deployment name comes from `Meta` |
| Check | `CheckName` (DNS label <= 63); `containercli.ValidName` stays for names read back |
| Answer | rulings 11, 13; open N30 |

### Spec.Image

| Column | Value |
| --- | --- |
| A source | `inst.Image` string built from image.* (A:dockergen:153; :podmangen:148; :deploy:609) |
| B source | `c.Image.Ref()` (B:render:1072; :901) |
| C | `image: "<ref>"` quoted, first service key |
| Q | `Image=<ref>`, first [Container] key, `%` doubled (grammar leaves nothing else) |
| K | container `image`, quoted; `imagePullPolicy` only if the tool sets one (K1 open) |
| Check | `ParseImage`, `Image.Check`: tag or digest required |
| Answer | E3, S7; open K1 |

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
| K | pod securityContext: runAsUser, runAsGroup, fsGroup = gid; runAsNonRoot when uid != 0 |
| Check | `ParseRunAs`, `RunAs` constructors; uid 0 allowed when stated |
| Answer | S3, E5 (tools map "default" to the zero value); open E22 |

A's `run-as: image` and B's `default` word map to the zero `RunAs`. A stated `0:0` writes
the root lines without `runAsNonRoot`. B refuses an empty value itself because it chowns its
data directory (E5 answer).

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
| A source | none today; todo 6 default read-only (open E21) |
| B source | none; refuses read-only (writable) |
| C | `read_only: true` after the posture block; writable writes nothing |
| Q | `ReadOnly=true` after the posture lines; writable writes nothing |
| K | `readOnlyRootFilesystem: true` in the container securityContext |
| Check | `RootUnset` is an error |
| Answer | ruling 9, C3 (smoke row before A turns it on); open E21 |

### Spec.Entrypoint and Spec.Command

| Column | Value |
| --- | --- |
| A source | storescript shell/path and its command (A:dockergen:158-161) |
| B source | none |
| C | `entrypoint: [..]` then `command: [..]`, flow lists, tokens quoted, `$` doubled |
| Q | refused; the [Container] extras `Entrypoint=` and `Exec=` pass (open P4) |
| K | `command` (Entrypoint) and `args` (Command), block lists, `$` doubled |
| Check | `CheckText` per token |
| Answer | S2 (compose fields), K13 open |

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
| Check | `envref.ValidName`, `CheckEnvName`, `CheckText`; duplicates refused |
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
| Answer | gdu 6, ruling 13; open C2, open E4 (tools hand over "" for the project/private default) |

### Spec.Memory, Spec.CPUSet, Spec.ShmSize

| Column | Value |
| --- | --- |
| A source | none on compose/podman; K resources (A:deploy:758-772) |
| B source | `cb.Mem`/`MonitorMem`, `cb.CPUSet`, `ContainerShmSize` (B:render:1108-1120; :945-957) |
| C | `mem_limit: "<v>"`, `cpuset: "<v>"`, `shm_size: "<v>"`, in that order |
| Q | `PodmanArgs=--memory=<v>`, `PodmanArgs=--cpuset-cpus=<v>`, `ShmSize=<v>` |
| K | Memory -> `limits.memory`; CPUSet, ShmSize refused. Requests/CPU: `kubernetes.Resources` |
| Check | `ValidMemory`, `ValidCPUSet`, `ValidShmSize` |
| Answer | S2; open E12 (one name, per-sink grammar), open P7 |

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
| Check | `Secret.Check`, `CheckEnvName` on EnvVar, C5 and duplicate checks |
| Answer | S9, C5, C6; open P2 (Kubernetes mapping, whole-Secret mounts) |

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
| Answer | ruling 9 (zero value `:ro`), C8; open Q7, open Q8 |

### Spec.Tmpfs

| Column | Value |
| --- | --- |
| A source | decoded stores dir, mode 0700 and the run-as ids (A:dockergen:252-257) |
| B source | none |
| C | `tmpfs:` `"<target>:mode=0700[,uid=U,gid=G]"` |
| Q | refused (extras `Tmpfs=` passes) |
| K | refused (open P4) |
| Check | `Tmpfs.Check`; a mount under a tmpfs is allowed (C5) |
| Answer | S2, C5 |

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
| Check | `CheckLabelKey`, `CheckLabelValue`, `CheckEnvName` on Var and EnvVar |
| Answer | C1, C4, S1 (reserved prefixes), todo 10; open Q2 compare stays in the tool |

### compose.File.Project

| Column | Value |
| --- | --- |
| A source | `in.Docker.ProjectName`, line omitted when empty (A:dockergen:102-104) |
| B source | `ComposeProject(cb.Name)` folding (B:render:1030-1041,1063) |
| C | `name: "<project>"` always, first line; also passed as `-p` |
| Q, K | not applicable |
| Check | `compose.ValidProject` |
| Answer | C9 |

### quadlet.Unit.Description, Rootful, Header

| Column | Value |
| --- | --- |
| A source | text (A:podmangen:142), `WantedBy=default.target` (:230), comment (:139) |
| B source | hostname and node type (B:render:896), `c.Podman.WantedBy` (:1013), no comment |
| Q | `Description=`, `WantedBy`, header comment |
| Check | description: no control, no line break, no trailing backslash |
| Answer | Q3 layout; open Q10, open Q11 |

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
  Kubernetes doubles `$` in env values, `command`, `args` and probe commands (K13 open,
  recommended doubled); it does not double elsewhere.
- Invalid UTF-8 is refused (it cannot be quoted safely). Control characters come out as
  `\xNN` or `\uNNNN` escapes of `strconv.Quote` for valid input.
- Block text: compose `content: |` writes the bytes as given (C10) with the chomping marker
  the text needs (`|-`, `|`, `|+`) and an indent marker when the first non-empty line starts
  with a space; empty lines carry no indent. Kubernetes `|` never uses `|+`: text with
  several trailing newlines is written as a quoted string (K "no decision needed"). The two
  differ on purpose; see open P8.
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
between them, none after the last; `[Service]` always written. Header comment first when
`Header` is set. `[Unit]`: Description, Wants, After. `[Container]`: Image, ContainerName,
HostName, the five Label lines, the tool's labels, the digest label is held back to the end
of the label group, User, Group, NoNewPrivileges, DropCapability, ReadOnly, Network,
PublishPort..., Memory/CPU `PodmanArgs`, ShmSize, Ulimit..., Environment (TZ first), Secret
(secrets, then configs), Volume, Health* (Cmd, Interval, Timeout, Retries, StartPeriod), then
extras. `[Service]`: Restart, then `Limit*` lines. `[Install]`: WantedBy. Where Q3's sketch and
S2 differ (Ulimit and ShmSize are now own fields), the own-field lines sit in the resources
group after the ports, ahead of Environment.

### 5.3 Compose order (C1)

Top level: `name`, `services`, `networks` (only when the service uses the project network),
`configs`, `secrets`; a block with nothing in it is left out. Service: image, container_name,
hostname, user, restart, entrypoint, command, posture block, read_only, mem_limit, cpuset,
shm_size, ulimits, extras, networks or network_mode, ports, environment, secrets, configs,
volumes, tmpfs, healthcheck, labels. The labels end the block, so the digest label closes the
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

1. Mount nesting (C5). A secret or config target must not equal or lie under a mount target
   or a tmpfs target, and the reverse. One nesting is allowed: a mount under a tmpfs
   (A binds existing stores inside its truststores tmpfs). Compose, quadlet and Kubernetes
   (volume mounts) all run it.
2. Duplicates. Targets across secrets, configs, mounts and tmpfs; environment names (and TZ
   against `Timezone`); label keys (including M's five and the digest key); secret names;
   compose `Secret.EnvVar` values and the digest variable and every `Label.Var` (one
   namespace, C6); Kubernetes volume names and container names; ulimit names.
3. Variable names (C6). `Secret.EnvVar`, `DigestLabel.EnvVar` and `Label.Var` refuse `PATH`,
   `HOME`, `IFS`, `SHELL`, `BASH_ENV` and any `LD_`, `DOCKER_` or `COMPOSE_` prefix.
4. Ports under host or none (ruling 9): the renderer leaves them out; its doc says so; no
   error. A tool keeps its own list check at load (B does).
5. Network and ports on Kubernetes: `Network` set is refused; a published `Host` that is not
   the container port is refused (no `hostPort`).
6. Sink refusals: `HostName`, `CPUSet`, `ShmSize`, `Ulimits`, `Mounts`, `Tmpfs`, `Network`,
   `Label.Var`, `Digest` on Kubernetes; `Entrypoint`, `Command`, `Tmpfs`, `Label.Var` on
   quadlet; `Secret.EnvVar` is required on compose and `Secret.Key`, `Secret.Volume` on
   Kubernetes (each error names the field).
7. Run-as against root: a stated `0:0` is written, never refused (both tools accept it).
8. Health: Kubernetes `Readiness` set together with `Health` is an error (one source).
9. Extras deny lists (S1, with its 2026-11-17 additions), each naming the version it was
   checked against in a test:
   - compose service: the keys M writes, keys that reach the host or another container
     (cap_add, devices, sysctls, pid, ipc, uts, userns_mode, cgroup, runtime, volumes_from,
     group_add, links ...), keys that pull in files or services (build, extends, env_file,
     depends_on, deploy, post_start, pre_stop, provider, develop, models), plus profiles,
     scale, annotations, logging, oom_score_adj, cpu_rt_*, credential_spec.
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
10. Extras values: text, integers, booleans, lists and maps; no raw text; text gets the UTF-8
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
| :159-160 | `entrypoint:` `command:` | `Spec.Entrypoint`, `Spec.Command` |
| :169-179 | labels le-mode, role, stores digest | `Spec.Labels` (digest via `Var`) |
| :182 | deploy digest label | `Spec.Digest` |
| :184 | `restart:` | `Spec.Restart` (empty now writes always) |
| :192 | `user:` | `Spec.RunAs` |
| :194 | posture block | M constant |
| :196-199 | `ports:` | `Spec.Ports` |
| :205-222 | `environment:` TZ, JAVA_*, SYSLOG_* | `Spec.Timezone`, `Spec.Env` |
| :229 | service `secrets:` | `Spec.Secrets` |
| :236-247 | service `configs:` | `Spec.Configs` |
| :252-257 | `tmpfs:` | `Spec.Tmpfs` |
| :288-293 | `healthcheck:` | `Spec.Health` |
| :262-267 | `volumes:` stores then libs | `Spec.Mounts` |
| :299-315 | inline config block text | writer frames `Spec.Configs[].Content`; payload stays in A |
| (none) | `read_only:`, network lines | new: `Spec.Root`, `Spec.Network` |

Stays in A: the payloads (application.yml, status script, logback XML, stores-init script),
the le-mode/role rule, the Java options merge, and the store selection.

### 7.2 A, quadlet (internal/podmangen/podmangen.go)

| Line | What | Home |
| --- | --- | --- |
| :139 | header comment | `quadlet.Unit.Header` (open Q10; text becomes M's) |
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
| :229-230 | `[Install]` WantedBy | `Unit.Rootful` (open Q11) |
| :232 | file name | `File.Name` |

### 7.3 A, Kubernetes (internal/deploy/deploy.go)

| Line | What | Home |
| --- | --- | --- |
| :348 | Namespace | `kubernetes.Namespace` (replaces `kube.NamespaceManifest`) |
| :377-386 | credentials Secret (stringData) | `kubernetes.Secret` (data only, K2 open) |
| :392-401 | stores Secret | `kubernetes.Secret` |
| :409-416 | image-pull Secret | `kubernetes.Secret` via `RegistrySecret` |
| :426-458 | PV and PVC | `kubernetes.PersistentVolume`, `PersistentVolumeClaim` |
| :477 | `---` join | `writer.Join` |
| :483-489 | name-only Secret for remove | `Manifest.Refs` (K3 open) |
| :496-527 | ConfigMap | `kubernetes.ConfigMap` from `Spec.Configs` |
| :534-546 | Deployment, replicas, selector | `kubernetes.Deployment` |
| :549-560 | pod labels | `Spec.Labels` |
| :565 | `automountServiceAccountToken: false` | M constant |
| :569-572 | imagePullSecrets | `Pod.ImagePullSecrets` |
| :577-587 | nodeSelector | `Pod.NodeSelector` |
| :169-203 | tolerations | `Pod.Tolerations` |
| :205-238 | affinity (verbatim block + spread) | `Pod.Affinity` (K12 open) |
| :249-261 | pod securityContext | `Spec.RunAs`; seccomp constant |
| :267-274 | container securityContext | M constant |
| :591-606 | libs-download init container | `Deployment.Init` (a `workload.Spec`) |
| :608-613 | container name, image, ports | `Spec.Name`, `Spec.Image`, `Spec.Ports` |
| :617-641 | env TZ, JAVA_*, syslog | `Spec.Timezone`, `Spec.Env` |
| :643-678 | volumeMounts | `Deployment.Mounts` and the secret/config mapping |
| :682-707 | liveness, readiness | `Deployment.Liveness`, `Readiness` (open P3) |
| :709-773 | resources | `kubernetes.Resources` |
| :711-755 | volumes (configMap, secret, projected, claim, emptyDir) | `Pod.Volumes` |
| :775-788 | Service | `kubernetes.Service` |
| :806-823 | `Written` list | `Manifest.Objects` (K15 open) |

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
| B:secrets:249-292 | registry Secret and `regcred` | `kubernetes.RegistrySecret` (K10, K11 open) |
| B:prep.go:418-426 | `joinManifests` | `writer.Join` |
| B:operator.go:411-424 | `joinYAMLDocs` | `writer.Join` |
| B:render:78-226 | PubSubPlusEventBroker body | `kubernetes.Custom` body tree (K5); content stays |
| operator bundle template | vendor YAML | `kubernetes.Raw` (K9 open) |
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

Each point states what this draft assumed. Entries are the questions file's IDs, "open <ID>",
and P-points that the file does not ask.

- open C2 (with E4): B hands over an empty `Network` for the project network (compose) or
  `private` (podman); `bridge` written means `network_mode: bridge` / `Network=bridge`.
  Recommended: B drops its `bridge` default. The field is `Spec.Network` whichever key feeds
  it.
- open E4: network key placement. Recommended: a scalar in each platform section; the key
  name is not used here, only the field `Spec.Network`.
- open E12: one memory name, each sink its own grammar. Recommended: `ValidMemory(Sink, s)`;
  a Kubernetes `m` suffix on memory is refused; Kubernetes CPU lives in `kubernetes.Resources`.
- open E21: A's read-only root, default true in A. Recommended: the field `Spec.Root` has no
  default; A's default stays in A; B refuses read-only until it supports it.
- open E22: the pod identity for B stays B's four custom fields; A's `runAs` feeds `Spec.RunAs`.
- open Q7: mount source grammar. Recommended: absolute path, none of `:` (one drive-letter
  root on compose, C8), whitespace, control, invisible, `$`, quotes, backtick, backslash;
  `%` allowed and doubled on quadlet; a relative source refused.
- open Q8: SELinux relabel letter as `Mount.Relabel`, none forgotten, A stays none, B keeps Z.
- open Q9: no value matching by name; the guide says secrets go through `Spec.Secrets`;
  `artifacttest` offers a check on units.
- open Q10: header comment on every unit: `# Generated by <tool>. Do not edit: a deploy
  overwrites this file.`; no version. `Unit.Header` exists as a switch only to keep this open.
- open Q11: `WantedBy` from `Unit.Rootful`: `default.target` rootless, `multi-user.target`
  rootful, derived from the tool's rootful flag.
- open K1: `imagePullPolicy` written only when the tool sets one (`kubernetes.Deployment`
  field); empty writes nothing.
- open K2: Secrets as `data` only, from raw bytes M encodes.
- open K3: `Manifest.Refs`, a name-only document, refuses a Namespace and a CRD.
- open K4: the reference function refuses a CRD, raw or not; B keeps its own CRD delete.
- open K5: custom resource body is a tree (`writer.Map`), numbers passed as numbers.
- open K6: Custom refuses kinds M types, but the guarantee is `artifacttest` requiring the
  three container lines and the pod `seccompProfile` on every workload except raw text.
- open K7: ownership labels and annotations on every object, PV included; A's PV warning
  reads size and path with one `kubectl get` (A's adoption step).
- open K8: no Secret in env vars, no `envFrom`; secrets are files only.
- open K9: the operator bundle stays `Raw`, unlabelled; a Namespace in a remove stream refused.
- open K10: the operator pull Secret's instance label is `pubsubplus-eventbroker-operator`.
- open K11: `kubernetes.DockerHubKey` constant, used when the registry is empty.
- open K12: A's verbatim affinity block is converted to a tree: whole numbers and booleans
  bare, `{}` as an empty mapping, everything else quoted.
- open K13: `$` doubled in exec probe commands; a hand check on a cluster before the tag.
- open K14: `defaultMode: 0400` written as the octal literal.
- open K15: `Manifest.Objects` with a cluster-wide flag; each tool decides what to probe.
- open K16: M exports the size constant and a plain check; the tool decides when to call.
- open K17: NFS only for volumes; no catch-all source.
- open K18: build the HTTP probe, `args` and the service types; no RBAC kinds until a tool
  authors one. Owner: nothing ahead of need, but the Service types are built now
  (2026-11-18): M checks the word and requires one, the tools set the default (design note
  section 11).
- open N11-N15: the character checks are `CheckText` over `execguard.IsControl` and
  `IsInvisible`; `IsControl` takes U+0085 (N13); `CheckMountSource` stands for N14's host-path
  check; `envref.ValidName` keeps its name (N15).
- open N25: line helpers and Posture constants private; Posture and the deny lists in
  `containerart/internal/posture`; `writer` exports `Quote` and `QuoteDollar`.
- open N30: two checks: `containercli.ValidName` for names read back, `workload.CheckName`
  (DNS label <= 63) for names a tool creates.
- open N35, N36, N37: the `Check*` twins private, `HostNetwork` removed with an
  `AssertNetwork`, and `artifacttest` kept under `containerart` in this draft (a move to the
  top level is N37's recommendation; the import rule in 2 holds either way).
- P1: `workload` imports `kube` and `containercli`, so `kube.NamespaceManifest` moves to
  `kubernetes.Namespace`. Recommended: do it; the alternative is a copy of the two grammars in
  `workload`. Belongs to the package inventory (review section 6, item 5).
- P2: Kubernetes secret mapping. Recommended: `Secret.Volume` names the volume, secrets with
  one `Volume` share a projected volume and one directory, items are key -> basename(Target);
  a whole-Secret mount (an existing Secret with unknown keys) needs `Key` empty on Kubernetes
  only, with `Target` read as the directory. Needs the owner's call because it changes the
  meaning of `Target` for one case.
- P3: Kubernetes probe numbers. Recommended: A passes `Deployment.Readiness` and `Liveness`
  with its own numbers (15/10/5/3 and 30/15) and leaves `Spec.Health` nil on Kubernetes, so
  the probes do not move; `Spec.Health` on Kubernetes maps to a readiness exec probe with
  `periodSeconds` = Interval, `timeoutSeconds` = Timeout, `failureThreshold` = Retries,
  `initialDelaySeconds` = StartPeriod.
- P4: refused on quadlet and Kubernetes: `Entrypoint`, `Command`, `Tmpfs` (quadlet reaches
  `Exec=`, `Entrypoint=`, `Tmpfs=` through extras). Recommended: refuse, since neither tool
  writes them there.
- P5: `Spec.Configs` on quadlet become `Secret=` lines named `Config.Name`; the tool creates
  the podman secret from `Content` (A does this today with its app yml and status script).
- P6: S11's answer (M's write refuses a symbolic link) against the quadlet "no decision
  needed" bullet that left the link refusal in the tools. Recommended: S11 wins; one helper,
  `outfile`, refuses a link for `generate` output and the installed unit.
- P7: `Spec.Memory` on Kubernetes writes `limits.memory` only; requests and CPU come from
  `kubernetes.Resources`; setting both for memory is a duplicate error.
- P8: block text: compose follows C10 (`|-`, `|`, `|+`); Kubernetes follows the K note (no
  `|+`, quoted string instead). Recommended: keep both as decided.
- P9: Kubernetes `Restart`: only empty or `always`; any other word is refused because a
  Deployment always restarts.
- P10: package names `containerart/writer`, `containerart/workload`, `containerart/kubernetes`
  and `containerart/outfile` are working names.
