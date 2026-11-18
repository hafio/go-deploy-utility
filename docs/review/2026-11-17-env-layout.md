# Env-file layout: today and flattened (platform sections)

This file shows, for each tool and each platform section (docker, podman, kubernetes) plus the
top-level image block, the env.yaml key tree as it is today (read from the yaml tags in the
structs) and as it would be after the decided rules (ruling 11 flatten, E1 camelCase, E3, E5,
E6, E7 grammar, E8, E9, E11). Placements still open in the questions file are shown with the
proposal and marked "# open: E<n>". The owner reopened the flatten itself (2026-11-17): every
placement below is up for review, decided or not. Updated 2026-11-18 with E16 to E24 and the
Secrets reach: the free-form `command` keys give way to the structured `cli` keys (E24),
`quadletDir` and `baseDir` go (E19, E20), both tools get `readOnlyRoot` (E21), Kubernetes
identity takes B's pod and container fields in both tools (E22), and every image key is a
full block (E23). A = solmq-conn-util, B = solace-util. Keys that exist only to refuse a removed
key by name (A: timezone, secrets, stores, mount-path, quadlet, mode, image under a platform;
B: composeFile, runtime, rootless, shmSize, ulimits.nofile/memlock) are left out of the trees;
E2 drops all by-name refusals. Rule applied to the trees: a key in the Flattened tree without
"open" is decided; "new" marks a key the tool does not have today.

## 1. Tool A: solmq-conn-util

### A: image

Today (A:internal/spec/image.go:30-37)

```yaml
image:
  repo: ghcr.io            # registry host[:port]; empty = Docker Hub
  name: solace/connector   # path inside the registry
  tag: 1.2.3               # tag (a digest is allowed)
  user: u                  # registry user (literal)
  user-env: REG_USER       # ...or env var name
  pass: p                  # registry password (literal)
  pass-env: REG_PASS       # ...or env var name
```

Flattened

```yaml
image:
  registry: ghcr.io        # E3: host[:port]; empty = Docker Hub
  name: solace/connector   # E3, revised 2026-11-18: path only, never a host
  tag: 1.2.3               # unchanged (a digest stays A only)
  user: u                  # unchanged; the Kubernetes pull Secret only (E18: no login)
  userEnv: REG_USER        # E1: Env suffix
  pass: p                  # unchanged
  passEnv: REG_PASS        # E1: Env suffix
```

### A: docker

Today (A:internal/spec/targets.go:201-217)

```yaml
docker:
  command: docker          # default docker; may carry flags, e.g. "docker --context foo"
  name: connector          # container name
  project-name: solace-connectors  # compose project (top-level name:)
  restart: unless-stopped  # default unless-stopped
  ports: [8090, "9000:9000"]  # bare int = N:N, or "host:container"
  libs:
    dir: ./libs            # host dir of the MQ jars (relative to env.yaml or absolute)
  run-as: "1000:0"         # uid[:gid]; the word image = image's own user
```

Flattened

```yaml
docker:
  cli: docker              # E24 (was command): docker | nerdctl, resolved on PATH
  compose: plugin          # E24 (new): plugin (<cli> compose) | standalone (docker-compose)
  context: remote1         # E24 (new, optional): --context; absent = the CLI's current
  name: connector          # unchanged (ruling 11: <platform>.name; DNS label <= 63)
  projectName: solace-connectors  # ruling 11 + E1: compose project
  restart: always          # E6: default now always (was unless-stopped)
  network:                 # E4 (new for A): B's block, kept
    mode: bridge           # bridge|host|none; absent: open C2
    ports: [8090, "9000:9000"]  # E7 (was docker.ports): bare N = N:N, handled in M
  resources:               # E12 (new on docker for A)
    memory: 2g             # E12
  readOnlyRoot: true       # E21 (new): A default true; M requires a value
  libs:
    dir: ./libs            # unchanged
  runAs: "1000:0"          # E1 + E5: absent = 1000:0; word default = image's own user
```

### A: podman

Today (A:internal/spec/targets.go:239-268)

```yaml
podman:
  command: podman          # default podman
  base-dir: /opt/x         # ignored legacy key (lint warning only)
  name: connector          # container name
  ports: [8090]            # bare int = N:N, or "host:container"
  restart: always          # default always
  libs:
    dir: ./libs            # host dir of the MQ jars
  run-as: "1000:0"         # uid[:gid]; the word image = image's own user
  rootful: false           # true = rootful engine; default false (rootless)
```

Flattened

```yaml
podman:
  # command                # E24: gone; podman is local only, the program is podman
  # base-dir               # E20: dropped
  name: connector          # unchanged
  network:                 # E4 (new for A): B's block, kept
    mode: private          # private|bridge|pasta|slirp4netns|host|none; absent: open C2
    ports: [8090]          # E7 (was podman.ports): bare N = N:N, handled in M
  restart: always          # unchanged
  resources:               # E12 (new on podman for A)
    memory: 2g             # E12
  readOnlyRoot: true       # E21 (new): A default true; M requires a value
  libs:
    dir: ./libs            # unchanged
  runAs: "1000:0"          # E1 + E5: absent = 1000:0; word default = image's own user
  rootful: false           # unchanged
```

### A: kubernetes

Today (A:internal/spec/kubernetes.go:18-48, 126-232, 262-292, 405-416)

```yaml
kubernetes:
  command: kubectl         # deploy CLI, e.g. "oc" or "kubectl --context prod"
  deployment:
    name: connector        # Deployment name
    namespace: solace      # namespace of every object
    replicas: 1            # replica count
    resources:
      cpu: 500m            # cpu request/limit
      memory: 512Mi        # memory request/limit
    node-selector: {zone: a}   # node label map
    tolerations:           # Kubernetes toleration entries
      - {key: k, operator: Equal, value: v, effect: NoSchedule, tolerationSeconds: 60}
    spread: none           # none|preferred|required; default none
    affinity: {}           # Kubernetes affinity block, written as is
  service:
    enabled: false         # create a Service
    port: "80:8090"        # servicePort:targetPort
  libs:                    # exactly one mode
    pvc:
      create: {storage: 1Gi, nfs: {server: nfs1, path: /exp}}   # 3 keys + retired name
      existing: my-claim   # ...or an existing PVC
    download: {urls: [], image: busybox:1.37, pvc: ""}          # 3 keys
  secrets:
    credentials: {create: true}    # or {existing: NAME}
    image-pull: {name: NAME, create: false}   # name = existing Secret; create = tool builds it
  run-as: "1000:0"         # uid:gid; gid also fsGroup
```

Flattened

```yaml
kubernetes:
  cli: kubectl             # E24 (was command): kubectl | oc, resolved on PATH
  contextFile: ~/.kube/prod  # E24 (new, optional): --kubeconfig; leading ~ expands (gdu 2)
  context: prod            # E24 (new, optional): --context; absent = the session's
                           # current context, pinned for the run (gdu 11)
  name: connector          # ruling 11 (was deployment.name)
  namespace: solace        # ruling 11 (was deployment.namespace)
  replicas: 1              # ruling 11 (was deployment.replicas)
  resources:               # ruling 11 dissolve; E12 for the name
    cpu: 500m              # E12
    memory: 512Mi          # E12
  placement:               # E13: each tool its own shape; E14: spread
    nodeSelector: {zone: a}    # E13: A's Kubernetes shape kept (was deployment.node-selector)
    tolerations: []        # E13: Kubernetes entries kept (was deployment.tolerations)
    affinity: {}           # E13: Kubernetes block kept (was deployment.affinity)
    spread: none           # E14 (was deployment.spread; default none)
  service:                 # E15: A keeps enabled and port
    enabled: false         # unchanged
    type: ClusterIP        # Service types (new): ClusterIP|NodePort|LoadBalancer;
                           # A default ClusterIP (recommended, to confirm)
    port: "80:8090"        # unchanged
  readOnlyRoot: true       # E21 (new): A default true; M requires a value
  libs:                    # unchanged but for download.image (pvc.create.storage, nfs, ...)
    pvc: {...}
    download:
      urls: []
      pvc: ""
      image:               # E23 (was the string busybox:1.37): the full image block
        registry: ""       # empty = Docker Hub
        name: busybox
        tag: "1.37"
  secrets:
    credentials: {create: true}    # unchanged ({existing: NAME} alternative)
    imagePull: {create: true}      # E8 + E1 (was image-pull; {existing: NAME} replaces name:)
    stores: {create: true}         # Secrets reach (new): or {existing: NAME}; was implicit
  securityContext:         # E22 (was run-as "1000:0"): pod fields, as in B and Kubernetes
    runAsUser: 1000
    fsGroup: 0
  containerSecurity:       # E22: container fields
    runAsUser: 1000
    runAsGroup: 0
```

## 2. Tool B: solace-util

### B: image

Today (B:internal/config/config.go:301-308)

```yaml
image:
  repo: solace-pubsub-standard   # image path (no host)
  tag: "10.26.0.8755"      # tag
  registry: registry.example.com # optional registry host prefix; omit = docker.io
  user: repo-user          # registry user (literal only)
  pass: CHANGE-ME          # registry password (literal)
  passEnv: SOLACE_REG_PASS # ...or env var name
```

Flattened

```yaml
image:
  registry: registry.example.com # unchanged (host[:port])
  name: solace-pubsub-standard  # E3 (was repo; image.name, revised 2026-11-18)
  tag: "10.26.0.8755"      # unchanged
  user: repo-user          # unchanged; pull Secrets only (E18: no login on docker/podman)
  userEnv: SOLACE_REG_USER # new: every credential pair is <key> and <key>Env
  pass: CHANGE-ME          # unchanged
  passEnv: SOLACE_REG_PASS # unchanged
```

### B: docker

Today (B:internal/config/config.go:965-974, 1047-1120, 1190-1192)

```yaml
docker:
  command: docker          # default docker
  compose: docker compose  # compose invocation; default "<command> compose"
  network:
    mode: bridge           # bridge|host; default bridge
    ports: ["8080:8080"]   # [ip:]host:container[/proto]; none when empty
  container:
    name: solace           # container name; default solace
    runUser: "1000001:0"   # uid:gid; default 1000001:0
    cpuset: "0-1"          # host cpus; default from the scaling tier
    monitorCpuset: "0"     # same, for the HA monitor
    mem: 3410m             # memory; default from the scaling tier
    dataDir: /opt/solace/data  # host bind mount for broker data
    ulimits:
      core: "-1"           # core file limit (1 live key)
    healthCheck:           # 6 keys: enabled, cmd, interval, timeout, retries, startPeriod
      enabled: true
```

Flattened

```yaml
docker:
  cli: docker              # E24 (was command): docker | nerdctl, resolved on PATH
  compose: plugin          # E24 (was a command string): plugin | standalone
  context: remote1         # E24 (new, optional): --context; absent = the CLI's current
  projectName: solace-event-brokers  # new (ruling 11 / gdu 9); default solace-event-brokers
  name: solace             # ruling 11 (was container.name); DNS label <= 63; one name for
                           # every role (separate hosts); redundancy.sameHost: true (new)
                           # makes <name>-primary, -backup, -monitor on one host
  restart: always          # E6 (new; default always)
  runAs: "1000001:0"       # E5 (was container.runUser); absent = default; word default allowed
  readOnlyRoot: false      # E21 (new): B default false; M requires a value
  network:                 # E4: unchanged
    mode: bridge           # absent: open C2 (B fills in bridge today)
    ports: ["8080:8080"]   # E7: a bare 8080 is now accepted too
  resources:
    memory: 3410m          # E12 (was container.mem)
  cpuset: "0-1"            # ruling 11 (was container.cpuset)
  monitorCpuset: "0"       # ruling 11 (was container.monitorCpuset)
  dataDir: /opt/solace/data  # ruling 11 (was container.dataDir)
  ulimits:
    core: "-1"             # ruling 11 (was container.ulimits.core)
  healthCheck:             # ruling 11 (was container.healthCheck): 6 keys unchanged inside
    enabled: true
```

### B: podman

Today (B:internal/config/config.go:990-1010, 1020-1028, 1047-1120)

```yaml
podman:
  command: podman          # default podman
  rootful: false           # true = rootful engine; default false (rootless)
  quadletDir: /etc/containers/systemd  # unit directory override
  baseDir: /opt/x          # legacy key; B still deletes a legacy key file under it
  network:
    mode: bridge           # bridge|host
    ports: ["8080:8080"]   # host:container list
  container:
    name: solace           # container name
    runUser: "1000:0"      # uid:gid; default 1000:0 rootless, 1000001:0 rootful
    cpuset: "0-1"          # host cpus
    monitorCpuset: "0"     # HA monitor cpus
    mem: 3410m             # memory
    dataDir: /opt/solace/data  # host bind mount
    ulimits: {core: "-1"}  # 1 key
    healthCheck: {...}     # 6 keys
```

Flattened

```yaml
podman:
  # command                # E24: gone; podman is local only, the program is podman
  rootful: false           # unchanged
  # quadletDir             # E19: dropped; the unit directory is always the default one
  # baseDir                # E20: dropped; B stops deleting the legacy file
  name: solace             # ruling 11 (was container.name)
  restart: always          # E6 (new; default always)
  runAs: "1000:0"          # E5 (was container.runUser); default 1000:0 / 1000001:0 rootful
  readOnlyRoot: false      # E21 (new): B default false; M requires a value
  network:                 # E4: unchanged; podman words (todo 6)
    mode: private          # absent: open C2
    ports: ["8080:8080"]   # E7: a bare 8080 is now accepted too
  resources:
    memory: 3410m          # E12 (was container.mem)
  cpuset: "0-1"            # ruling 11
  monitorCpuset: "0"       # ruling 11
  dataDir: /opt/solace/data  # ruling 11
  ulimits: {core: "-1"}    # ruling 11 (unchanged inside: 1 key)
  healthCheck: {...}       # ruling 11 (unchanged inside: 6 keys)
```

### B: kubernetes

Today (B:internal/config/config.go:670-700, 715-731, 755-770, 783-807, 815-876)

```yaml
kubernetes:
  command: kubectl         # default kubectl
  name: solace             # PubSubPlusEventBroker name
  namespace: solace        # broker namespace
  adminSecret: my-admin    # name: referenced, or built when semp.adminPass is set
  tlsServerSecret: my-tls  # name: referenced, or built from tls.cert/certKey
  imagePullSecret: regcred2  # dockerconfigjson Secret name
  imagePullPolicy: Always  # empty = the CR's IfNotPresent
  updateStrategy: automatedRolling   # automatedRolling|manualPodRestart
  serviceAccount: sa       # optional
  storage:                 # 4 keys: class, msgNodeSize, monNodeSize, customVolumeMount
    class: standard
  msgNode:
    cpu: "2"               # broker cpu
    mem: 3410Mi            # broker memory; default from the scaling tier
  operator:                # 6 keys
    image: solace/pubsubplus-eventbroker-operator:1.4.2
    namespace: solace-operator
    watchNamespaces: ""
    watchBrokerNs: true
    cpu: 200m
    mem: 256Mi
  securityContext: {runAsUser: "1000001", fsGroup: "1000002"}      # pod, 2 keys
  containerSecurity: {runAsUser: "1000001", runAsGroup: "1000002"} # container, 2 live keys
  podAnnotations: {a: b}   # map
  podLabels: {a: b}        # map
  placement:               # 8 keys + nodeAffinity/podAffinity/podAntiAffinity (B's own words)
    labelsPrimary: ["zone: a"]        # also labelsBackup, labelsMonitor
    tolerationsPrimary: ["k=v:NoSchedule"]  # also tolerationsBackup, tolerationsMonitor
    antiAffinityNamespaces: [solace]  # default: own namespace
    antiAffinityWeight: 100           # always-on anti-affinity weight
    nodeAffinity: {...}    # {preferred: [...], required: [...]}
    podAffinity: [...]     # entries {weight, topologyKey, matchLabels, namespaces}
    podAntiAffinity: [...] # same entry shape
  loadBalancer:
    ip: 10.0.0.5           # fixed LB ip
    ipPool: pool1          # address pool
    annotations: ["k: v"]  # strings "key: value"
  ports: ["semp=8080:80"]  # strings name=containerPort[:servicePort][/proto]
```

(B has no runAs key on kubernetes; no network key; no resources key.)

Flattened

```yaml
kubernetes:
  cli: kubectl             # E24 (was command): kubectl | oc, resolved on PATH
  contextFile: ~/.kube/prod  # E24 (new, optional): --kubeconfig
  context: prod            # E24 (new, optional): --context; absent = pinned current (gdu 11)
  name: solace             # unchanged
  namespace: solace        # unchanged
  secrets:                 # E8
    credentials: {create: true}     # was adminSecret; or {existing: NAME}; E11: no mixes
    imagePull: {existing: regcred2} # was imagePullSecret; or {create: true}
    tls: {create: true}             # E9 (was tlsServerSecret); or {existing: NAME}
    additionalUsers: {create: true} # Secrets reach (new): or {existing: NAME}; was implicit
                                    # K8: still named in the CR's extraEnvVarsSecret
  imagePullPolicy: Always  # unchanged
  updateStrategy: automatedRolling  # unchanged
  serviceAccount: sa       # unchanged
  storage: {...}           # unchanged: 4 keys
  resources:               # E12 (was msgNode)
    cpu: "2"               # E12 (was msgNode.cpu)
    memory: 3410Mi         # E12 (was msgNode.mem; bare "m" suffix refused)
  operator:                # 6 keys; E12 for cpu/mem
    image:                 # E23 (was a string with image.registry prefixed): full block
      registry: ""         # empty = Docker Hub; no hidden prefix
      name: solace/pubsubplus-eventbroker-operator
      tag: "1.4.2"
    namespace: solace-operator
    watchNamespaces: ""
    watchBrokerNs: true
    resources: {cpu: 200m, memory: 256Mi}   # E12 (was operator.cpu, operator.mem)
  securityContext: {...}   # E22: unchanged; A aligns to these pod fields
  containerSecurity: {...} # E22: unchanged; A aligns to these container fields
  podAnnotations: {a: b}   # unchanged
  podLabels: {a: b}        # unchanged
  placement:               # E13: each tool its own shape; E14: spread
    labelsPrimary: ["zone: a"]        # E13: unchanged; also labelsBackup, labelsMonitor
    tolerationsPrimary: ["k=v:NoSchedule"]  # E13: unchanged; also Backup, Monitor
    nodeAffinity: {...}    # E13: unchanged, B's own words
    podAffinity: [...]     # E13: unchanged
    podAntiAffinity: [...] # E13: unchanged; refused unless spread is none (E14)
    spread: preferred      # E14 (new; none|preferred|required; B default preferred)
    antiAffinityNamespaces: [solace]  # E14 (stays, broker-only)
    antiAffinityWeight: 100           # E14 (stays, broker-only)
  service:                 # E15 (one block)
    type: LoadBalancer     # Service types (new): LoadBalancer|ClusterIP|NodePort;
                           # B default LoadBalancer, written into the CR as today
    ports: [{name: semp, containerPort: 8080, servicePort: 80, protocol: TCP}]  # E15
    loadBalancer: {ip: 10.0.0.5, ipPool: pool1}  # E15 (was kubernetes.loadBalancer);
                           # refused unless type is LoadBalancer
    annotations: {k: v}    # E15 (list of strings -> mapping)
```

## 3. Moves

"docker|podman" in one row counts as two moved keys. "new" rows are listed separately below the
moves table and not counted as moves.

| tool | today | flattened | decided by |
| --- | --- | --- | --- |
| A | image.repo | image.registry | E3 |
| A | image.user-env | image.userEnv | E1 |
| A | image.pass-env | image.passEnv | E1 |
| A | docker.project-name | docker.projectName | ruling 11, E1 |
| A | docker.run-as | docker.runAs (word image -> default) | E1, E5 |
| A | podman.run-as | podman.runAs (word image -> default) | E1, E5 |
| A | kubernetes.run-as | kubernetes.securityContext.{runAsUser,fsGroup}, containerSecurity.{runAsUser,runAsGroup} | E22 |
| A | docker.restart default unless-stopped | default always | E6 |
| A | podman.base-dir | dropped | E20 |
| A | docker.ports, podman.ports | docker.network.ports, podman.network.ports | E4, E7 |
| A | docker.command | docker.cli, with new docker.compose and docker.context | E24 |
| A | podman.command | dropped (podman is local only) | E24 |
| A | kubernetes.command | kubernetes.cli, with new kubernetes.contextFile and kubernetes.context | E24 |
| A | kubernetes.libs.download.image (string) | kubernetes.libs.download.image.{registry,name,tag} | E23 |
| A | kubernetes.deployment.name | kubernetes.name | ruling 11 |
| A | kubernetes.deployment.namespace | kubernetes.namespace | ruling 11 |
| A | kubernetes.deployment.replicas | kubernetes.replicas | ruling 11 |
| A | kubernetes.deployment.resources.cpu | kubernetes.resources.cpu | ruling 11 (E12) |
| A | kubernetes.deployment.resources.memory | kubernetes.resources.memory | ruling 11 (E12) |
| A | kubernetes.deployment.node-selector | kubernetes.placement.nodeSelector | ruling 11 (E13: shape kept) |
| A | kubernetes.deployment.tolerations | kubernetes.placement.tolerations | ruling 11 (E13: shape kept) |
| A | kubernetes.deployment.affinity | kubernetes.placement.affinity | ruling 11 (E13: shape kept) |
| A | kubernetes.deployment.spread | kubernetes.placement.spread | ruling 11, E14 |
| A | kubernetes.secrets.image-pull | kubernetes.secrets.imagePull | E8, E1 |
| A | kubernetes.secrets.image-pull.name | kubernetes.secrets.imagePull.existing | E8 |
| B | image.repo | image.name | E3 (revised 2026-11-18) |
| B | docker.container.name | docker.name | ruling 11 |
| B | podman.container.name | podman.name | ruling 11 |
| B | docker.container.runUser | docker.runAs | E5, ruling 11 |
| B | podman.container.runUser | podman.runAs | E5, ruling 11 |
| B | docker.container.cpuset | docker.cpuset | ruling 11 |
| B | podman.container.cpuset | podman.cpuset | ruling 11 |
| B | docker.container.monitorCpuset | docker.monitorCpuset | ruling 11 |
| B | podman.container.monitorCpuset | podman.monitorCpuset | ruling 11 |
| B | docker.container.dataDir | docker.dataDir | ruling 11 |
| B | podman.container.dataDir | podman.dataDir | ruling 11 |
| B | docker.container.ulimits.core | docker.ulimits.core | ruling 11 |
| B | podman.container.ulimits.core | podman.ulimits.core | ruling 11 |
| B | docker.container.healthCheck.* (6) | docker.healthCheck.* | ruling 11 |
| B | podman.container.healthCheck.* (6) | podman.healthCheck.* | ruling 11 |
| B | docker.container.mem | docker.resources.memory | E12 |
| B | podman.container.mem | podman.resources.memory | E12 |
| B | podman.baseDir | dropped | E20 |
| B | podman.quadletDir | dropped | E19 |
| B | docker.command | docker.cli, with new docker.context | E24 |
| B | docker.compose (a command string) | docker.compose: plugin or standalone | E24 |
| B | podman.command | dropped (podman is local only) | E24 |
| B | kubernetes.command | kubernetes.cli, with new kubernetes.contextFile and kubernetes.context | E24 |
| B | kubernetes.adminSecret | kubernetes.secrets.credentials | E8 |
| B | kubernetes.imagePullSecret | kubernetes.secrets.imagePull | E8 |
| B | kubernetes.tlsServerSecret | kubernetes.secrets.tls | E9 |
| B | kubernetes.msgNode.cpu | kubernetes.resources.cpu | E12 |
| B | kubernetes.msgNode.mem | kubernetes.resources.memory | E12 |
| B | kubernetes.operator.cpu | kubernetes.operator.resources.cpu | E12 |
| B | kubernetes.operator.mem | kubernetes.operator.resources.memory | E12 |
| B | kubernetes.ports | kubernetes.service.ports | E15 |
| B | kubernetes.loadBalancer.ip | kubernetes.service.loadBalancer.ip | E15 |
| B | kubernetes.loadBalancer.ipPool | kubernetes.service.loadBalancer.ipPool | E15 |
| B | kubernetes.loadBalancer.annotations | kubernetes.service.annotations | E15 |
| B | kubernetes.operator.image (string, registry prefixed) | kubernetes.operator.image.{registry,name,tag} | E23 |

New keys (not moves): A: docker.network.mode and podman.network.mode (E4), docker.compose and
docker.context (E24), kubernetes.contextFile and kubernetes.context (E24), readOnlyRoot on
three platforms (E21), kubernetes.secrets.stores (Secrets reach), kubernetes.service.type
(Service types),
docker|podman.resources.memory (E12). B: docker.projectName (ruling 11), docker.restart
and podman.restart (E6), docker.context, kubernetes.contextFile and kubernetes.context (E24),
readOnlyRoot on docker and podman (E21), kubernetes.secrets.additionalUsers (Secrets reach),
image.userEnv (no decision needed), kubernetes.placement.spread (E14),
kubernetes.service.type (Service types).

Other A keys renamed by E1 outside the platform sections (listed for completeness, not
counted): java-options, leader-election (and its conn-ref), solace-defaults, mq-defaults,
health-show-details, transform-headers, fail-over, conn-ref, password-env, and (inside solace:
and mq:) client-username-env, client-password-env, key-alias, user-env, password-env.
workflows.file_pattern -> filePattern. This list is from the yaml tags; E1 says the exact list
is fixed at build.

## 4. The owner's answers (2026-11-18)

- Name: `<platform>.name` stays one container name, since HA brokers run on separate hosts.
  New key `redundancy.sameHost: true | false` (B): true makes the names `<name>-primary`,
  `<name>-backup` and `<name>-monitor` for three brokers on one host.
- Identity: the word `default` in both tools, on every platform, tells M to write none.
  Unset means the tool's default: A `1000:0` (Kubernetes: user 1000, group 0, fsGroup 0).
  B's Kubernetes default: open.
- Remote engine: detectable on docker, since the selected context's endpoint is not
  `unix://` or `npipe://`. A bind-mount source (A's `libs.dir`, B's `dataDir`) is then
  refused, and the docs say so. Podman is local only (E24); Kubernetes writes no host path.
- Kubernetes Service ports: `kubernetes.service.ports` (E15).
- Network: keep B's block, `network.mode` and `network.ports`, in both tools (E4, E7).
- Per-role placement: B's roles and A's single Deployment are an inherent difference.
- E12: memory under `resources` on every platform, A gaining it on docker and podman; CPU on
  Kubernetes in A only; B's `msgNode.cpu` dropped.
- E8 and B: no custom name for a Secret a tool builds. Yes.
- E24: ok.
- Time zone: the top-level `timezone` key, in both tools today, written into the container
  as `TZ`. The platform-level `timezone` keys A retired stay gone.
- Unknown keys: each tool accepts only the keys of its own env.yaml shape (strict decoding).
  This is a tool convention, not module code (M has no YAML library). Exceptions: blocks
  passed on verbatim (A's connector properties in `solace:` and `mq:`, B's custom-resource
  pass-through). The guide states it for a third tool; A needs CR-0046.

- C2: an absent `network.mode` is the deploy's own network (docker project network, podman
  `private`); a written `bridge` is docker's shared bridge.
- E14: one `kubernetes.placement.spread` key, A default `none`, B `preferred`, B gaining `none`.
- E13: each tool keeps today's placement shapes; A and B deploy different kinds.
- B on Kubernetes states user 1000001, group and fsGroup 1000002 when unset, and refuses 0.
- Service types: a new `kubernetes.service.type` in both tools. M owns the words (ClusterIP,
  NodePort, LoadBalancer, exact case) and checks them; the tools own the default. B: default
  LoadBalancer, accepts all three. A: see "Still open".

## 5. Still open

- One rule for the `create` defaults of every Secret.
- A's default Service type. Recommended: ClusterIP.
