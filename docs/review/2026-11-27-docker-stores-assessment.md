# Docker delivery of tool-created TLS stores: is the script still the right way

> [!WARNING]
> **Not a supported Solace product.** `go-deploy-utility` was created by Solace
> Professional Services and is supported only by Solace Professional Services --
> not by Solace Support. For help with this module, contact your Solace
> Professional Services representative rather than opening a Solace Support
> case. This notice covers this module only, not the Solace products that the
> tools built on it deploy.

Date 2026-11-27. Read-only assessment for the owner. Nothing here is built, and
no engine, build, test or dev script was run.

Paths. A = solmq-conn-util, written `A internal/dockergen/dockergen.go:150`
(relative to its repo). M = this module, written `M docs/design/...:520`. "(mine)"
marks my own inference. "(unverified)" marks anything I could not confirm from
code or a page I fetched; those are the items a smoke run on a real engine must
settle before any of this is built.

## 1. Purpose

Decide whether A's way of delivering a tool-created PKCS12 store on docker
(compose) -- base64 compose secret, startup script, `entrypoint:`, `command:`,
`tmpfs:` -- is still the right one, given the constraints now on the table: no
key material in a file on the host's disk, a stated posture (read-only root,
all capabilities dropped, no-new-privileges, non-root), least privilege, remote
engines refused for bind mounts, no registry login by the tool, and M building
nothing ahead of need. The owner is also deciding whether M supports
entrypoint, command and tmpfs on quadlet and Kubernetes at all; today only this
workaround uses them (`M docs/design/2026-11-17-container-description.md:528-536,
648-657`). The answer changes what M's container description must carry.

## 2. Today

### 2.1 Where the store comes from

- A store under `tls.stores.<name>` has `create: {certs | keys | jks}` or an
  `existing:` source (`A internal/spec/defaults.go:23-47, 101-106, 137-138`).
  Every store is PKCS12 (`A internal/spec/defaults.go:17-21`), AES-256, built
  deterministically so an unchanged store has the same bytes (`A docs/DESIGN.md:486-497`).
- A created store has no host file: deploy builds it in memory and hands the
  bytes straight to the platform (`A internal/spec/defaults.go:140-143`). Only an
  `existing.pkcs12` store is a host file, and docker bind-mounts it read-only
  (`A internal/dockergen/dockergen.go:259-264`).
- Inside the container every selected store is
  `/app/external/classpath/truststores/<name>` on every platform
  (`A internal/spec/targets.go:78`; `A docs/configuration.md:420-421`).

### 2.2 The docker mechanism

1. Deploy base64-encodes each created store and puts it in the environment of
   the `docker compose` child process only, under the secret's own name; a size
   check refuses a store whose encoding exceeds one environment variable
   (131072 bytes on Linux, 32767 on Windows)
   (`A internal/gen/stores.go:346, 374-393`; `A cmd/solmq-conn-util/main.go:758`).
2. The compose file declares each as a top-level secret from the `environment:`
   provider, so the value is never in the file (`A internal/dockergen/dockergen.go:118-124`),
   and mounts it long-syntax at `/app/external/var/stores/<name>`, deliberately
   beside, not inside, the credentials directory the configtree import reads
   (`A internal/dockergen/dockergen.go:135-146, 229`; `A internal/spec/targets.go:99-104`).
3. The service starts through a script delivered as a compose `configs:` entry
   (`A internal/dockergen/dockergen.go:115, 246-247`). It runs `set -eu`, `umask 077`,
   decodes each file with `base64 -d` to `<name>.tmp`, `mv`s it into place, then
   `exec "$@"` (`A internal/storescript/storescript.go:60-77`).
4. `entrypoint: [/bin/sh, /app/external/.stores-init]` and `command: [<image
   entrypoint> <image cmd>]` are written together, because setting
   `entrypoint:` drops the image's CMD (`A internal/dockergen/dockergen.go:155-161`).
   The image's own launch line is a hard-coded constant, Jib's `java -cp
   @/app/jib-classpath-file ...JmsConnector` plus the Spring config location
   (`A internal/storescript/storescript.go:42-52`).
5. The decode target is `tmpfs: /app/external/classpath/truststores:mode=0700[,uid,gid]`
   (`A internal/dockergen/dockergen.go:248-257`). Existing stores bind-mount
   inside it (`A internal/dockergen/dockergen.go:259-264`; M allows one nesting under a
   tmpfs, `M docs/design/2026-11-17-container-description.md:798-799`).
6. Before any write, a deploy carrying created stores runs
   `checkImageCommand`: `image inspect`, a `pull` if the engine does not have
   the image, a second inspect, then a refusal unless the image's entrypoint and
   cmd equal the two constants (`A cmd/solmq-conn-util/main.go:765-772, 825-846`).
7. Rotation: a stores digest label and a deploy digest label make compose
   recreate the container when a store changes (`A internal/dockergen/dockergen.go:185-198`;
   `A docs/DESIGN.md:301-305`).

### 2.3 Who reads the stores in this connector

- MQ side: `spring.ssl.bundle.jks.<name>` is rendered with `truststore.location`,
  `keystore.location`, passwords, `type` and `key.alias`
  (`A internal/render/render.go:94-122`), and the MQ binder references it with
  `ibm.mq.ssl-bundle` (`A internal/render/render.go:185-187`). The binder context
  loads `SslAutoConfiguration` so the bundle resolves
  (`A internal/render/render.go:91, 160-176`). So the MQ consumer is Spring Boot's
  SSL bundle machinery handing an `SSLContext` to IBM's connection factory.
- Solace side: the tool writes `SSL_TRUST_STORE`, `SSL_TRUST_STORE_PASSWORD`,
  `SSL_TRUST_STORE_FORMAT`, `SSL_KEY_STORE`, `SSL_KEY_STORE_PASSWORD`,
  `SSL_KEY_STORE_FORMAT` as binder `api-properties` pointing at the mount path
  (`A internal/tls/tls.go:57-66`). That is the JCSMP property map, not a Spring
  bundle. Whether the Solace binder can take a Spring bundle instead is
  (unverified); A does not use one.
- Passwords: store passwords are separate credentials, mounted as files at
  `/app/external/var/secrets/<name>` and read through configtree
  (`A internal/spec/targets.go:81-97`; `A internal/dockergen/dockergen.go:135-139`).

### 2.4 The other platforms

- Podman: one podman secret per store, mounted as a file by a `Secret=` line
  (`A internal/podmangen/podmangen.go:103-105`); podman's secret store holds binary,
  capped at 511999 bytes (`A internal/gen/gen.go:626-634`). No script.
- Kubernetes: the stores Secret is mounted as a projected volume
  (`A internal/deploy/deploy.go:721-727`). No script.

### 2.5 Facts that matter for the comparison

- A emits no `read_only:` line today (grep of `A internal/` finds none); the
  read-only root is a planned key, `readOnlyRoot` default true for A
  (`M docs/design/2026-11-17-container-description.md:520-524`). Every
  alternative below must therefore also work with a read-only root; a tmpfs and
  a mounted secret both stay writable or readable under it, so the baseline
  already fits. Whether the JVM then needs a writable `/tmp` is (unverified) and
  independent of this question.
- The script is restart-safe by construction: it runs on every container start
  inside the container, so docker's restart policy, a crash loop and a host
  reboot all re-decode with no help from compose or the tool.
- The weak spot is the contract with the image: the constants in
  `A internal/storescript/storescript.go:42-43` must equal what the image does,
  and the check costs an inspect and possibly a pull before every deploy that
  carries a created store.

## 3. Alternatives

### 3.1 Keep today's script (baseline)

How it works: section 2.

Security.
- At rest on the host: nothing. The value lives in deploy's process memory and
  the compose child's environment (a host process environment is readable by the
  same uid and root via /proc, briefly). It is not in the compose file.
- In the engine: compose copies the secret content into the container's
  secret mount (a tmpfs the engine manages) (unverified in detail); the base64
  copy is therefore in RAM at `/app/external/var/stores/<name>`, mode and owner
  per the long syntax defaults, and decoded copies sit in the 0700 tmpfs owned by
  the run-as uid.
- Two RAM copies per store per container (encoded and decoded). Both are the
  PKCS12 as built, encrypted under the store password (AES-256), so the exposure
  is the encrypted store, not a bare key. The store password is a separate
  secret file.
- Who can read: the container's uid, the engine's root, and anyone with
  `docker exec` or inspect rights (inspect shows no secret content; exec does).
- The script itself is public text in the compose file's configs; it holds no
  secret.

Compatibility.
- Compose environment-provider secrets need Compose 2.23.1 or newer already
  (`A internal/dockergen/dockergen.go:118-122`). Configs with `content` also
  came in 2.23.1 (fetched: configs reference).
- Remote engine: works. Nothing bind-mounts; the secret, config and tmpfs all
  travel through the API. This is the only baseline property the alternatives
  must match.
- Rootless docker: tmpfs with `uid=,gid=` options works as a mount option;
  rootless remaps uids (unverified for the exact `uid=` mapping).
- nerdctl compose: environment-source secrets and `tmpfs:` support is
  (unverified); the file is Compose-spec and nerdctl follows the spec unevenly.
- Needs a shell and `base64` in the image: busybox in this image (A
  `storescript.go:11-12`); a distroless future image breaks it.

Restart and reboot: re-decodes on every start. No ordering dependency. Best of
all alternatives on this axis.

What it adds to A and M.
- A: `internal/storescript` (77 lines), the entrypoint/command/tmpfs/configs
  block in renderService, `checkImageCommand` with an inspect-and-pull, the
  image constants, the env-size limit.
- M: `Spec.Entrypoint`, `Spec.Command`, `Spec.Tmpfs` with `Tmpfs.Check`, the
  "mount under a tmpfs" rule, and refusals of those fields on quadlet and
  Kubernetes (`M docs/design/2026-11-17-container-description.md:163-164,
  296, 528-536, 648-657, 798-799, 811-812`).
- No extra image.

Parity: podman and Kubernetes differ (secret file, projected volume) but all
three present the same path inside the container, which is the parity that
matters (`A docs/configuration.md:420-421`).

Weaknesses: image-launch constants and the pre-deploy pull-and-inspect;
131072-byte size cap; two RAM copies; an exec form that every future image
change must keep matching.

### 3.2 One-shot init service decoding into a tmpfs-backed named volume

How it works (as the owner described): service `stores-init` (busybox)
mounts the base64 secrets and a named volume declared with
`driver: local`, `driver_opts: {type: tmpfs, device: tmpfs, o: "size=..,uid=..,mode=0700"}`;
it decodes and exits 0; the app service has
`depends_on: {stores-init: {condition: service_completed_successfully}}` and
mounts the volume read-only. The app keeps its own entrypoint and command.

Docker's own reference confirms the local driver accepts tmpfs volumes:
`docker volume create --driver local --opt type=tmpfs --opt device=tmpfs
--opt o=size=100m,uid=1000` is the documented example, with data in RAM and a
caution that permissions on tmpfs may reset after a container restart
(search result page for the CLI reference, below; the compose volumes page I
fetched shows only nfs and bind examples and does not mention tmpfs).

The problem (mine, unverified, and the reason I do not recommend this):
- Lifecycle. The local driver mounts a tmpfs volume when the first container
  using it starts and unmounts it when the last one stops. The init container
  has exited by the time the app starts, so the volume may be empty again at
  that moment. I believe this makes the one-shot pattern unreliable, but I did
  not find a page saying so and it is the first thing to smoke before anything
  else. A long-running init (sleep after decoding, held by a healthcheck and
  `service_healthy`) would keep the mount alive, but it then holds the secrets
  and a shell for the life of the stack, which is worse than the baseline.
- Reboot and restart ordering. `depends_on` is enforced by `docker compose up`,
  not by the daemon. After a host reboot, `restart: unless-stopped` has the
  daemon start each container by itself, in no guaranteed order, so the app can
  start before the init has run or against an empty volume. The baseline has no
  such window. The tool would have to document that a reboot needs
  `compose up`, or the app would fail at start and loop.
- `uid=`/`gid=` and `mode=` as mount options: a restarted app container gets a
  tmpfs that docker may remount with the volume's defaults (see the docker
  caution above).

Security. Key material (encrypted PKCS12) is in RAM only, in the volume, which
every container that mounts it can read, and which the init container can write.
Two additional containers' worth of attack surface (the init runs as whatever
uid it is given; it needs the secrets mounted and write access, but no
capabilities). The init service must also carry the full posture block, a
user, and a read-only root, doubling the stated-posture surface in the file.

Compatibility. Compose 2.23.1 or newer already; `service_completed_successfully`
needs Compose v2.1+ (unverified exact version). Remote engine: works (named
volume and secrets through the API). Rootless: tmpfs mount options inside a
user namespace (unverified). nerdctl: `driver_opts` for the local driver is
(unverified), likely unsupported.

What it removes and adds.
- Removes from A and M: `Entrypoint`, `Command`, `checkImageCommand`'s inspect
  and pull, the image constants. The app service becomes ordinary.
- Adds in A: a second service in the file (posture, user, image, healthcheck or
  completion handling), a named volume with driver options, depends_on, the
  stores-init config. Adds in M: a multi-service concept the description does
  not have (today one container description renders one service); named
  volumes with `driver_opts`; `depends_on`; or an "init container" notion that
  Kubernetes would map to an initContainer and quadlet to a second unit. That
  is building ahead of need, because podman and Kubernetes do not need it.
- Extra image: yes. busybox must be pulled, and the tool does not log in to a
  registry (per the rules), so a private mirror or air-gapped host needs the
  operator to supply it. Variant 3.2b: reuse the app image as the init image
  with its own `entrypoint: [/bin/sh, ...]`. That avoids a second image but
  reintroduces an `entrypoint:` field (without the command constants), and it
  starts a JVM image for a decode step.

Restart and reboot: see above. App-only restart after the init has exited: the
volume may be gone (unverified), so it is not equivalent to the baseline.

Parity: Kubernetes would use an initContainer plus an emptyDir with
`medium: Memory`, but Kubernetes has no need (the Secret is already binary and
mounted); so the extra machinery exists only to emulate what the other two
platforms get for free.

Verdict: removes one fragile thing (the image-launch constants) and replaces it
with a less certain one (volume lifecycle and boot ordering). Not recommended.

### 3.3 PEM instead of PKCS12

How it works: a store becomes a PEM bundle (trust certificates, or a private
key plus chain), which is text, so a compose environment secret carries it
and compose mounts it directly at the final path. No script, no tmpfs, no
entrypoint.

Who reads the stores here:
- MQ (Spring SSL bundle): supported. Spring Boot's SSL bundle docs (fetched)
  define a `pem` bundle type with `truststore.certificate`, `keystore.certificate`
  and `keystore.private-key`, each a file location or inline PEM, and
  `ibm.mq.ssl-bundle` consumes whichever bundle type it names. An encrypted
  private key needs `private-key-password` (the page lists it; the key must be
  PKCS8-encrypted PBES2 for Java, mine). This part works, subject to the Spring
  Boot version the connector ships (`A docs/DESIGN.md:591-592` says connector 2.x
  is Spring Boot 3, 3.x is Boot 4; PEM bundles exist since 3.1, per the Spring
  blog result).
- Solace (JCSMP via `api-properties`): not supported. The JCSMP secure-sessions
  page I fetched lists only JKS and PKCS12 for `SSL_TRUST_STORE_FORMAT`, takes a
  file path, and offers `SSL_IN_MEMORY_TRUST_STORE` as a `java.security.KeyStore`
  object; it says nothing of PEM. The Solace Spring Cloud Stream binder passes
  `api-properties` through, so a PEM file at `SSL_TRUST_STORE` would fail as
  not a keystore. I found no key store counterpart on the JCSMP page, and the
  JMS API page names only JKS and PKCS12 (search result, not fetched).
- So PEM cannot be the format for the Solace connections, which are most of
  what this tool does. PEM for MQ only and PKCS12 for Solace would split the
  store model by system and break "every store is PKCS12" and "store paths are
  the same everywhere" (`A docs/configuration.md:417-421`), plus the validate
  and error messages built on them.

Security: key material is plain PEM text (a private key in PEM can be encrypted
with a password, but then it is the same shape of secret again). If the key is
unencrypted, the key is unprotected in RAM at the mount and in the env of the
compose child; with PKCS12 it is wrapped in AES-256 until Java opens it. That is
a weaker at-rest-in-memory posture than the baseline (mine).

Compatibility: works on any Compose that supports environment secrets; remote
engine fine; rootless fine; nerdctl as for any env secret (unverified). Size
limit (131072 bytes) the same, in practice never hit by PEM.

Restart and reboot: the secret mount is rebuilt by the engine at each start, so
no ordering dependency.

What it removes and adds: removes everything the baseline adds, in A and M.
Adds: a second store format and builder (PEM writer, encrypted PKCS8), new
`create:` options, a PEM-to-JKS conversion nowhere to run, and for `existing:`
stores a different operator contract. Podman and Kubernetes parity: they would
have to carry PEM too, or the tool would keep two formats; neither platform
needs it.

Verdict: cannot cover the Solace consumer, so it does not replace the script
(it only could for an MQ-only deployment).

### 3.4 `secrets:` with `file:` from an in-memory host location

How it works: deploy writes each PKCS12 to a file under `$XDG_RUNTIME_DIR`
(a per-user tmpfs on systemd hosts), and the compose secret uses `file:`. The
compose secrets reference (fetched) lists exactly two sources, `file` and
`environment`, and neither is stated to be binary-safe or text-only; a `file`
secret with binary content is a plain file read, so binary should work (mine).

Why not:
- It puts the key material in a file on the host, even though in RAM. The owner's
  rule says no credential or private key in a file on the host's disk; a tmpfs is
  not disk, but it is a host file with the host's permissions that the tool
  must create, protect and delete, and it is a file any process of that user can
  read. That is the thing the rule is trying to avoid.
- The file must outlive the `up`: compose bind-mounts a file secret from the
  host (unverified for the exact mechanism on a non-swarm engine), so a
  container restart reuses it. After a host reboot the tmpfs is empty and the
  `restart: unless-stopped` container fails to start or finds an empty mount.
  `$XDG_RUNTIME_DIR` is also removed at the last logout on many systems
  (unverified for this host).
- A remote engine cannot see the host path; the tool already refuses bind
  sources on a remote engine, and this is the same refusal in disguise.
- Not portable: no `$XDG_RUNTIME_DIR` on Windows or macOS, where Docker Desktop
  also runs the engine in a VM, so a host path is not the engine's.
- Rootless: works on Linux; ownership of the file is the host user's, not the
  run-as uid (mount ownership caveats; unverified).

What it removes and adds: removes the script, entrypoint, command, tmpfs and
the image check; adds a host-file writer, a cleanup path, and platform
conditionals. Parity: podman secrets and Kubernetes Secrets take the bytes with
no host file, so this one is the odd one out.

Verdict: violates the host-file rule in spirit and breaks on reboot and remote
engines. Rejected.

### 3.5 Newer Docker Compose features that carry binary

What the docs say (fetched):
- Configs: sources are `file`, `environment`, `content` (inlined text) and
  `external`; `environment` and `content` arrived in Compose 2.23.1; mounted
  mode 0444 by default. The page says nothing about binary or base64.
- Secrets: sources are `file` and `environment`; no `content` on the page I
  read. The page says nothing about binary or base64, and `environment` secrets
  are not supported by `docker stack deploy`.
- Neither page, nor the compose volumes page, mentions base64 decoding or a
  binary-safe inline source. I found no compose feature that decodes base64 for
  the container. (Not finding one on these pages is not proof none exists in a
  release newer than my sources; the owner could check the compose-spec repo's
  changelog -- unverified.)
- An environment variable cannot hold NUL, so `environment:` cannot carry the
  raw PKCS12 on any compose version (A already says this,
  `A internal/storescript/storescript.go:5-8`). A `content:` config is YAML text,
  so the same limit applies in a different guise: raw bytes would not survive
  as text.
- A binary-safe route exists only with `file:` (alternative 3.4) or with
  swarm secrets/`docker secret create`, which need swarm mode and a stdin
  upload (out of scope; unverified fit).

Related idea I considered and set aside: after `docker compose create`, have
the tool stream a tar of the stores into the container with `docker cp -`. That
is binary-safe, works on a remote engine and leaves no host file, but a tmpfs is
emptied when the container stops, so every restart and every host reboot would
need the tool to re-inject; the tool is not running then. Rejected (mine,
unverified).

Verdict: nothing newer removes the need as of the pages I read.

### 3.6 Letting Spring decode base64 itself

Evidence (fetched): the Spring Boot SSL reference, under "Configuring SSL With
Java KeyStore Files", says "Rather than the location to a file, its Base64
encoded content can be provided ... the value of the property should start with
`base64:`", and the same for PEM. The 3.3.0 M3 release notes (search result)
introduce the `base64:` resource prefix; I read that sentence in the 3.3.x docs,
so the minimum Spring Boot is 3.3 (mine). Exact example:

```
spring.ssl.bundle.jks.<bundle>.keystore.location=base64:<encoded PKCS12>
```

How it would work here: the connector already imports every file in
`/app/external/var/secrets/` as a property through configtree, using the file
name as the property key (`A internal/spec/targets.go:81-97`). A compose secret
named `spring.ssl.bundle.jks.<bundle>.truststore.location`, whose value is
`base64:<encoded PKCS12>`, would set that property with no script, no decoded
file and no tmpfs: the JVM decodes it in memory. The compose secret stays an
`environment:` secret (text), mounted in the credentials directory like any
other. (Whether configtree accepts that file name is mine, from Spring's rule
that a file's relative path becomes the property name; unverified for dotted
names with hyphens in this connector.)

Consumers it covers:
- MQ: yes. The MQ binder reads the bundle (section 2.3), so this works, on a
  connector built on Spring Boot 3.3 or newer. Which connector releases are
  that is (unverified); A ships support for 2.13.0 to 3.2.1
  (`A internal/storescript/storescript.go:36-37`), and Boot 3.3 is not
  guaranteed in the oldest.
- Solace: not today. A hands JCSMP a path (`A internal/tls/tls.go:57-66`), and
  JCSMP takes a path or an in-memory KeyStore, not base64
  (JCSMP page, fetched). Spring's decoding only helps if the Solace binder builds
  that KeyStore from a Spring bundle, which I could not verify (searches found
  nothing on the starter). If it does not, a Solace-side truststore still needs a
  file at `SSL_TRUST_STORE`, so the script stays for exactly the half this tool
  was built for.

Security: best of the set for the MQ half. No decoded file, no tmpfs, one RAM
copy as a Java property value (visible to any Spring property inspection such as
the actuator env endpoint if exposed; Boot hides values by default, unverified
for this connector's `management.*` settings, which A writes). The PKCS12 is
still AES-256 encrypted. A property whose name carries the word `location` is
not masked by default sanitising keys, but its value is only the encrypted
PKCS12.

Compatibility: no compose dependency beyond what A already requires; remote
engine, rootless and nerdctl are as for any env secret. No engine feature
needed.

Restart and reboot: identical to the baseline in spirit; the app reads the
secret mount at every start, no ordering.

What it removes and adds: for a pure-MQ deployment removes all baseline
machinery in A and M. In a mixed deployment it removes nothing, because the
Solace half still needs the script. It adds a second delivery path to the
renderer (secret as a property), a Spring-version floor, and different error
modes (a bad base64 string fails the app context at start with a Spring
message, rather than the script's `set -e`).

Verdict: promising for MQ, blocked by the Solace consumer. Worth a smoke test
of the Solace side before dismissing, because if the Solace binder (or a small
`SSL_IN_MEMORY_TRUST_STORE` bridge) can use a bundle, this would retire the
whole workaround on docker.

### 3.7 Variants of the baseline (mine)

Cheaper than replacing it, and each removes a weakness listed in 3.1.

- 3.7a Learn the command from the image. Instead of comparing the inspected
  entrypoint and cmd to constants, take them from the inspect and render the
  `command:` from that. The compose file then never goes stale, but it still
  needs the inspect-and-pull before the file is written, and an image that
  starts through a shell wrapper or with no CMD makes `exec "$@"` meaningless.
  Keep the comparison as a guard too if the owner prefers a loud failure.
- 3.7b Set a size on the tmpfs (`size=`). With no size, a tmpfs defaults to 50%
  of host RAM (docker tmpfs page, search result); a store of at most a few
  hundred KB needs far less, and a cap states the posture.
- 3.7c Decode only what is needed. Skip the `.tmp`/`mv` dance? No: keep it, the
  partial-read protection is real.
- 3.7d Raise the size cap by splitting a store across two secrets. Not
  worth it; a 131072-byte base64 cap is a store of about 98 KB, which is
  far larger than a trimmed truststore, and the error already says what to do.

## 4. Side by side

| | 3.1 Script | 3.2 Init + tmpfs volume | 3.3 PEM | 3.4 file: in RAM | 3.5 Compose feature | 3.6 Spring base64 |
|---|---|---|---|---|---|---|
| Host file | none | none | none | yes, tmpfs file | n/a | none |
| Key in container RAM | 2 copies, encrypted | volume copy | PEM, plain unless encrypted | 1 copy | n/a | 1 copy, encrypted |
| Covers MQ | yes | yes | yes | yes | n/a | yes (Boot 3.3+) |
| Covers Solace | yes | yes | no (JKS/PKCS12 only) | yes | n/a | no (unverified) |
| Remote engine | yes | yes | yes | no | n/a | yes |
| Survives reboot | yes | doubtful (ordering) | yes | no | n/a | yes |
| App-only restart | yes | doubtful (tmpfs lifetime) | yes | host file needed | n/a | yes |
| Extra image | no | busybox (or app image) | no | no | n/a | no |
| Entrypoint/command in M | yes | no (3.2b: entrypoint) | no | no | n/a | no |
| tmpfs field in M | yes | no (volume options) | no | no | n/a | no |
| Image inspect and pull | yes | no | no | no | n/a | no |
| New M concepts | none beyond today | second service, driver_opts | second format | host-file writer | n/a | secret-as-property |
| Parity with podman/K8s | path parity | worse | needs PEM everywhere | worse | n/a | partial |
| Verdict | keep | reject | reject | reject | none exists | follow up |

Headline: only the baseline covers both consumers, survives restart and reboot
with no ordering, works against a remote engine and leaves nothing on the host.
The one alternative with a real upside (3.6) cannot yet cover the Solace half.

## 5. Recommendation

1. Keep the script (3.1) on docker. It is the only option that satisfies every
   constraint for both the Solace and MQ consumers, and its cost is contained in
   one package and one compose block. The cost it imposes on M (entrypoint,
   command, tmpfs) is also what the owner asked about; on that question the
   answer is: **support them on compose only**, as `M docs/design/2026-11-17-container-description.md`
   already specifies (refused on quadlet and Kubernetes, extras pass through).
   Neither podman nor Kubernetes uses them, and they would be built ahead of need.
2. Do not adopt the init service (3.2): it trades one fragile contract for a
   volume-lifecycle and boot-ordering risk, adds an image to pull and a second
   posture block, and introduces a multi-service notion into M that no other
   platform wants.
3. Reject 3.3, 3.4 and 3.5. PEM cannot feed JCSMP. A host file violates the
   rule and fails on reboot and remote engines. No Compose feature carries
   binary.
4. Take 3.7b (a `size=` on the tmpfs) as a small hardening, and consider 3.7a
   (render the `command:` from the inspected image) only if image launch lines
   start changing between releases; today they have not (four releases checked,
   `A internal/storescript/storescript.go:36-37`).
5. Ask one cheap question before closing the door on 3.6 (section 6, Q1). If the
   Solace binder can build its truststore from a Spring SSL bundle, the
   workaround can retire on docker for connectors on Spring Boot 3.3 or newer,
   and M would not need entrypoint, command or tmpfs at all. That is the only
   path I found that would let M drop them.

Costs of the recommendation: the 131072-byte cap and the image-launch check
stay; a distroless future image would break the script and force 3.6.

## 6. Open questions for the owner

1. Smoke on a real engine: can the Solace binder take its trust and key store
   from a Spring SSL bundle or an in-memory KeyStore, so `SSL_TRUST_STORE` need
   not be a path? If yes, 3.6 is viable for both consumers; if no, it is MQ-only
   and the script stays. (I could not verify this from the pages I could read.)
2. Which Spring Boot version does each supported connector release run? The
   `base64:` resource needs 3.3 or newer; A supports 2.13.0 to 3.2.1 and the
   oldest may be below that.
3. Is the volume-lifecycle concern in 3.2 real on the engines you support? A
   one-line smoke (create a tmpfs-driver volume, write from a container, let it
   exit, read from a second) settles it, though I recommend not pursuing 3.2
   either way.
4. M scope: confirm that `Entrypoint`, `Command` and `Tmpfs` are compose-only in
   M, refused on quadlet and Kubernetes with the extras escape hatch as
   designed. The alternative that would change this is Q1.
5. Does any supported image or planned future image lack a shell or `base64`
   (distroless, hardened)? That would force 3.6 or a different delivery.
6. Is a deploy against nerdctl compose or rootless docker in scope? I could not
   verify environment-source secrets or `tmpfs` `uid=` handling on either.
7. Should `checkImageCommand` keep failing the deploy on an unknown image
   command (today), or warn and use the inspected values (3.7a)? The first is
   safer, the second drops the constants.
8. A emits no `read_only:` today; the planned `readOnlyRoot` default true is a
   separate change. Does the connector need a writable `/tmp` under it
   (unverified), and should that be decided before the docker block is
   finalised?

## Sources read

- https://docs.docker.com/reference/compose-file/secrets/ (sources: file, environment)
- https://docs.docker.com/reference/compose-file/configs/ (file, environment, content; 2.23.1; mode 0444)
- https://docs.docker.com/reference/compose-file/volumes/ (driver_opts; no tmpfs example)
- https://docs.spring.io/spring-boot/reference/features/ssl.html (PEM inline, JKS and PEM `base64:`)
- https://docs.spring.io/spring-boot/3.3.0/reference/features/ssl.html (served as 3.3.13; `base64:` for JKS and PEM)
- https://docs.solace.com/API/API-Developer-Guide-JCSMP/JCSMP-API-Creating-Secure-Sessions.htm (JKS and PKCS12; in-memory KeyStore)
- Search results only, not fetched: Spring Boot 3.3.0 M3 release notes (`base64:` prefix);
  Docker `volume create` reference and tmpfs mounts page (tmpfs volume example; 50% RAM default);
  Solace JMS SSL property page; IBM MQ Spring Boot starter listings.
- Not found: any page confirming Solace binder support for a Spring SSL bundle, or the starter version that added `ibm.mq.ssl-bundle`.
