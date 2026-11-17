# Env-file layout: today and flattened (platform sections)

This file shows, for each tool and each platform section (docker, podman, kubernetes) plus the
top-level image block, the env.yaml key tree as it is today (read from the yaml tags in the
structs) and as it would be after the decided rules (ruling 11 flatten, E1 camelCase, E3, E5,
E6, E7 grammar, E8, E9, E11). Placements still open in the questions file are shown with the
proposal and marked "# open: E<n>". The owner reopened the flatten itself (2026-11-17): every
placement below is up for review, decided or not. A = solmq-conn-util, B = solace-util. Keys that exist only to refuse a removed
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
  repository: solace/connector  # E3: path only, never a host
  tag: 1.2.3               # unchanged (a digest stays A only)
  user: u                  # unchanged
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
  command: docker          # unchanged
  compose: docker compose  # open: E16 (new; default "<command> compose"; string or list)
  name: connector          # unchanged (ruling 11: <platform>.name; DNS label <= 63)
  projectName: solace-connectors  # ruling 11 + E1: compose project
  restart: always          # E6: default now always (was unless-stopped)
  network: bridge          # open: E4 (new; bridge|host|none; absent = project network)
  ports: [8090, "9000:9000"]  # E7: bare N = N:N, handled in M (unchanged for A)
  resources:               # open: E12 (new on docker for A)
    memory: 2g             # open: E12
  readOnlyRoot: true       # open: E21 (new; default true)
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
  command: podman          # unchanged
  # base-dir               # open: E20 (dropped)
  name: connector          # unchanged
  network: private         # open: E4 (new; private|bridge|pasta|slirp4netns|host|none)
  ports: [8090]            # E7: bare N = N:N, handled in M (unchanged for A)
  restart: always          # unchanged
  resources:               # open: E12 (new on podman for A)
    memory: 2g             # open: E12
  readOnlyRoot: true       # open: E21 (new; default true)
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
  command: kubectl         # unchanged
  name: connector          # ruling 11 (was deployment.name)
  namespace: solace        # ruling 11 (was deployment.namespace)
  replicas: 1              # ruling 11 (was deployment.replicas)
  resources:               # ruling 11 dissolve; open: E12 for the name
    cpu: 500m              # open: E12
    memory: 512Mi          # open: E12
  placement:               # open: E13 / E14
    nodeSelector: {zone: a}    # open: E13 (was deployment.node-selector)
    tolerations: []        # open: E13 (was deployment.tolerations; entries unchanged)
    affinity: {}           # open: E13 (was deployment.affinity)
    spread: none           # open: E14 (was deployment.spread; default none)
  service:                 # open: E15 (A keeps both keys)
    enabled: false         # unchanged
    port: "80:8090"        # unchanged
  readOnlyRoot: true       # open: E21 (new; default true)
  libs:                    # unchanged: 9 keys (pvc.create.storage, nfs.server, nfs.path, ...)
    pvc: {...}
    download: {...}
  secrets:
    credentials: {create: true}    # unchanged ({existing: NAME} alternative)
    imagePull: {create: true}      # E8 + E1 (was image-pull; {existing: NAME} replaces name:)
  runAs: "1000:0"          # E1 + E5
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
  repository: solace-pubsub-standard  # E3 (was repo)
  tag: "10.26.0.8755"      # unchanged
  user: repo-user          # unchanged
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
  command: docker          # unchanged
  compose: docker compose  # unchanged
  projectName: solace-event-brokers  # new (ruling 11 / gdu 9); default solace-event-brokers
  name: solace             # ruling 11 (was container.name); DNS label <= 63
  restart: always          # E6 (new; default always)
  runAs: "1000001:0"       # E5 (was container.runUser); absent = default; word default allowed
  network: bridge          # open: E4 (was network.mode; scalar; no filled-in default)
  ports: ["8080:8080"]     # was network.ports; E7: bare 8080 now also accepted
  resources:
    memory: 3410m          # open: E12 (was container.mem)
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
  command: podman          # unchanged
  rootful: false           # unchanged
  quadletDir: /etc/containers/systemd  # open: E19 (B keeps; A has no such key)
  # baseDir                # open: E20 (dropped; B stops deleting the legacy file)
  name: solace             # ruling 11 (was container.name)
  restart: always          # E6 (new; default always)
  runAs: "1000:0"          # E5 (was container.runUser); default 1000:0 / 1000001:0 rootful
  network: private         # open: E4 (was network.mode; podman words)
  ports: ["8080:8080"]     # was network.ports; E7: bare 8080 now also accepted
  resources:
    memory: 3410m          # open: E12 (was container.mem)
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
  command: kubectl         # unchanged
  name: solace             # unchanged
  namespace: solace        # unchanged
  secrets:                 # E8
    credentials: {create: true}     # was adminSecret; or {existing: NAME}; E11: no mixes
    imagePull: {existing: regcred2} # was imagePullSecret; or {create: true}
    tls: {create: true}             # E9 (was tlsServerSecret); or {existing: NAME}
  imagePullPolicy: Always  # unchanged
  updateStrategy: automatedRolling  # unchanged
  serviceAccount: sa       # unchanged
  storage: {...}           # unchanged: 4 keys
  resources:               # open: E12 (was msgNode)
    cpu: "2"               # open: E12 (was msgNode.cpu)
    memory: 3410Mi         # open: E12 (was msgNode.mem; bare "m" suffix refused)
  operator:                # 6 keys; open: E12 for cpu/mem, E23 for image rule
    image: solace/pubsubplus-eventbroker-operator:1.4.2  # E23 open: full reference
    namespace: solace-operator
    watchNamespaces: ""
    watchBrokerNs: true
    resources: {cpu: 200m, memory: 256Mi}   # open: E12 (was operator.cpu, operator.mem)
  securityContext: {...}   # unchanged (E22 open: keep B's four fields)
  containerSecurity: {...} # unchanged (E22 open)
  podAnnotations: {a: b}   # unchanged
  podLabels: {a: b}        # unchanged
  placement:               # open: E13 / E14
    primary:               # open: E13 (per role; also backup, monitor)
      nodeSelector: {zone: a}       # was labelsPrimary (list of "k: v" -> mapping)
      tolerations: []      # was tolerationsPrimary (strings -> Kubernetes entries)
    affinity: {...}        # open: E13 (was nodeAffinity/podAffinity/podAntiAffinity; one block)
    spread: preferred      # open: E14 (new; none|preferred|required; B default preferred)
    antiAffinityNamespaces: [solace]  # open: E14 (stays, broker-only)
    antiAffinityWeight: 100           # open: E14 (stays, broker-only)
  service:                 # open: E15 (one block)
    ports: [{name: semp, containerPort: 8080, servicePort: 80, protocol: TCP}]  # open: E15
    loadBalancer: {ip: 10.0.0.5, ipPool: pool1}  # open: E15 (was kubernetes.loadBalancer)
    annotations: {k: v}    # open: E15 (list of strings -> mapping)
```

## 3. Moves

"docker|podman" in one row counts as two moved keys. "new" rows are listed separately below the
moves table and not counted as moves.

| tool | today | flattened | decided by |
| --- | --- | --- | --- |
| A | image.repo | image.registry | E3 |
| A | image.name | image.repository | E3 |
| A | image.user-env | image.userEnv | E1 |
| A | image.pass-env | image.passEnv | E1 |
| A | docker.project-name | docker.projectName | ruling 11, E1 |
| A | docker.run-as | docker.runAs (word image -> default) | E1, E5 |
| A | podman.run-as | podman.runAs (word image -> default) | E1, E5 |
| A | kubernetes.run-as | kubernetes.runAs | E1, E5 |
| A | docker.restart default unless-stopped | default always | E6 |
| A | podman.base-dir | dropped | open E20 |
| A | kubernetes.deployment.name | kubernetes.name | ruling 11 |
| A | kubernetes.deployment.namespace | kubernetes.namespace | ruling 11 |
| A | kubernetes.deployment.replicas | kubernetes.replicas | ruling 11 |
| A | kubernetes.deployment.resources.cpu | kubernetes.resources.cpu | ruling 11 (open E12) |
| A | kubernetes.deployment.resources.memory | kubernetes.resources.memory | ruling 11 (open E12) |
| A | kubernetes.deployment.node-selector | kubernetes.placement.nodeSelector | open E13 |
| A | kubernetes.deployment.tolerations | kubernetes.placement.tolerations | open E13 |
| A | kubernetes.deployment.affinity | kubernetes.placement.affinity | open E13 |
| A | kubernetes.deployment.spread | kubernetes.placement.spread | open E14 |
| A | kubernetes.secrets.image-pull | kubernetes.secrets.imagePull | E8, E1 |
| A | kubernetes.secrets.image-pull.name | kubernetes.secrets.imagePull.existing | E8 |
| B | image.repo | image.repository | E3 |
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
| B | docker.container.mem | docker.resources.memory | open E12 |
| B | podman.container.mem | podman.resources.memory | open E12 |
| B | docker.network.mode | docker.network | open E4 |
| B | podman.network.mode | podman.network | open E4 |
| B | docker.network.ports | docker.ports | open E7 |
| B | podman.network.ports | podman.ports | open E7 |
| B | podman.baseDir | dropped | open E20 |
| B | kubernetes.adminSecret | kubernetes.secrets.credentials | E8 |
| B | kubernetes.imagePullSecret | kubernetes.secrets.imagePull | E8 |
| B | kubernetes.tlsServerSecret | kubernetes.secrets.tls | E9 |
| B | kubernetes.msgNode.cpu | kubernetes.resources.cpu | open E12 |
| B | kubernetes.msgNode.mem | kubernetes.resources.memory | open E12 |
| B | kubernetes.operator.cpu | kubernetes.operator.resources.cpu | open E12 |
| B | kubernetes.operator.mem | kubernetes.operator.resources.memory | open E12 |
| B | placement.labels{Primary,Backup,Monitor} | placement.<role>.nodeSelector | open E13 |
| B | placement.tolerations{Primary,Backup,Monitor} | placement.<role>.tolerations | open E13 |
| B | placement.nodeAffinity / podAffinity / podAntiAffinity | placement.affinity | open E13 |
| B | kubernetes.ports | kubernetes.service.ports | open E15 |
| B | kubernetes.loadBalancer.ip | kubernetes.service.loadBalancer.ip | open E15 |
| B | kubernetes.loadBalancer.ipPool | kubernetes.service.loadBalancer.ipPool | open E15 |
| B | kubernetes.loadBalancer.annotations | kubernetes.service.annotations | open E15 |
| B | kubernetes.operator.image prefix rule | full reference, no prefix | open E23 |

New keys (not moves): A: docker.network and podman.network (open E4), docker.compose (open
E16), readOnlyRoot on three platforms (open E21), docker|podman.resources.memory (open E12).
B: docker.projectName (ruling 11), docker.restart and podman.restart (E6), image.userEnv (no
decision needed), kubernetes.placement.spread (open E14).

Other A keys renamed by E1 outside the platform sections (listed for completeness, not
counted): java-options, leader-election (and its conn-ref), solace-defaults, mq-defaults,
health-show-details, transform-headers, fail-over, conn-ref, password-env, and (inside solace:
and mq:) client-username-env, client-password-env, key-alias, user-env, password-env.
workflows.file_pattern -> filePattern. This list is from the yaml tags; E1 says the exact list
is fixed at build.

## 4. Questions the flatten raises

- Name clash: B docker.name and podman.name are the container name, but HA runs three
  containers (primary, backup, monitor) and redundancy.<role>.name is the host name. Does
  <platform>.name stay one stem the roles derive from, as docker.container.name is today?
- E5 says <platform>.runAs in both tools, but B has no single run-as on kubernetes (four
  fields: securityContext.runAsUser/fsGroup, containerSecurity.runAsUser/runAsGroup, kept per
  E22 open). So kubernetes.runAs exists in A only; confirm that is intended, or E22 changes it.
- E7 puts <platform>.ports in the section. On kubernetes B's kubernetes.ports (Service ports)
  would land on the same name with a different meaning (container publication vs Service
  port). E15 proposes kubernetes.service.ports; the two open answers must agree.
- E4/E7 together: B's network.mode moves to a scalar network while its ports move out of the
  network block. B's check "ports with host or none networking" (kept, ruling 9) then reads two
  sibling keys instead of one block.
- B's per-role keys (placement.primary/backup/monitor, E13) have no A counterpart (A has one
  Deployment, no roles). Layout would be placement.nodeSelector in A and
  placement.primary.nodeSelector in B. Same key names, different depth.
- B's spread default is preferred, A's none (E14); the same key name on one path carries two
  defaults. Also antiAffinityNamespaces/Weight sit under placement but only apply to B.
- E12: B's kubernetes.resources.cpu would be fixed by the scaling tier (no cpu count key), so
  the key exists in the section but is not freely settable in B; A's cpu is free. E12 says
  "memory (and cpu on kubernetes, A only)". Confirm B's msgNode.cpu is dropped, not renamed.
- E12 and A: A has no docker/podman memory key today, so docker.resources.memory is new in A,
  not a move. Confirm A should gain it, or whether resources stays kubernetes-only for A.
- E8 and A: kubernetes.secrets.imagePull for A becomes {create: true} or {existing: NAME}; A's
  old image-pull.name is a Secret the operator manages (maps to existing). The old
  create: false default is now the absence of the block. Confirm absent means no pull Secret.
- E8 and B: custom names for a Secret the tool builds go (derived name only), including the
  name tlsServerSecret gave a built TLS Secret.
- B's docker.compose (open E16 for A) stays flat in B, but docker.command and docker.compose
  both accept a string or a list; E17 (open) restricts programs per key, B only.
- Retired by-name keys (A: base-dir, timezone, secrets, stores, mount-path, quadlet, mode;
  B: composeFile, runtime, rootless, shmSize, ulimits.nofile/memlock) are decode-only today.
  E2 drops all refusals, so under strict decoding they become plain unknown keys; A needs strict
  decode (CR-0046) first, otherwise an old key vanishes silently.
