# Secrets assessment: today in A and B against the proposed single rule

Date 2026-11-18. Read-only assessment for the owner. Nothing here is built.

Paths. A = solmq-conn-util (`solace-ibmmq-connector-helper`), paths relative to it, written
`A spec/kubernetes.go:320` for `internal/spec/kubernetes.go:320` (and `A cmd/...` for `cmd/`).
B = solace-util (`solace-cnt-scripts`), written `B config/config.go:114` the same way. Every
"today" claim carries a file:line. "(mine)" marks my own recommendation or inference.

## 1. Purpose and the proposed rule

Purpose: decide whether one rule can govern every Kubernetes Secret in both tools, and what
that costs operators. Today the two tools differ: A says "create" or "existing" in words
(A spec/kubernetes.go:125-129) and refuses a missing block; B infers ownership from which keys
are set (B config/config.go:114, 162, 198) and builds Secrets silently.

The rule (Kubernetes Secrets only):

- Each Secret has a block with `create: true` or `existing: NAME`.
- Block absent: no Secret is built and none is referenced.
- Material present (password, cert, store, registry user/pass, users list) but no block: load
  error "add create: true or existing: NAME".
- `create: true` without material: error. `existing` with material: error.
- A built Secret always has the derived name. No custom names.
- Stated exceptions: B's operator `regcred` stays built implicitly from `image.user` and
  `image.pass` (E10, unlabelled, by name, K10); B with no credentials block and no
  `semp.adminPass` lets the operator generate the admin password.
- Decisions already taken: E8, E9, E10, E11, E18, K8, "Secrets reach" (inside E9), all in
  `docs/review/2026-11-15-surface-review-questions.md`.

Headline finding, explained in section 4 (E1) and section 5: "existing with material = error"
cannot be applied to A's credentials, A's stores or the multi-platform env file as worded. In A
the password fields ARE the key list of an `existing` Secret, and one A env.yaml feeds three
platforms. The rule needs a per-platform, per-secret definition of "material" before it is
written into either tool.

## 2. One table per secret type

Column "Proposed" states the rule as I read it with the owner's decisions applied.

### 2.1 Kubernetes credentials (A connection credentials; B admin, monitor, psk)

A's keys: every credential the config references, one file per name under `/run/secrets`
(A spec/kubernetes.go:121-129; A docs/configuration.md:1000-1008).
B's keys: `username_admin_password`, optional `username_monitor_password`, optional `preshared_auth_key` (B k8s/secrets.go:80-124).

| Situation | A today | B today | Proposed | Change for operators |
|---|---|---|---|---|
| Block absent, material absent | Nothing built, nothing mounted; legal (A validate/validate.go:1252-1260 `credentialsWired`) | No `adminPass`, no `adminSecret`: CR names no Secret, operator generates `<name>-pubsubplus-admin-creds` (B config/config.go:125-146; B config/validate.go:618-622) | Same in both. B is a stated exception (operator generates) | None |
| Block absent, material present | Load error "credentials are referenced but kubernetes.secrets.credentials is omitted ... add create: true or existing" (A validate/validate.go:272-284); only a warning on teardown (:277-283) | Built silently as `<name>-admin` when `semp.adminPass` is set (B config/config.go:114-131; B k8s/prep.go:169, 191) | Error in both: "add create: true or existing: NAME" | A: none (already so). B: every env file with `adminPass` and no block must add `create: true` |
| `create: true` | Builds `<name>-credentials` from literals and resolved `-env` values; applied as `stringData` (A spec/kubernetes.go:320-327; A deploy/deploy.go:375-388; A gen/gen.go:281-288) | No such key. Equivalent is `adminPass` set (B k8s/secrets.go:80-124) | Builds the derived name. Required material: `adminPass` (B), at least one credential (A) | B: new `kubernetes.secrets.credentials.create` (E8) |
| `create` without material | n/a: a spec with no credentials has no positions to build; `create: true` then builds an empty Secret (A validate/validate.go:955-972 checks only create-XOR-existing, not content) (mine: unverified that an empty stringData applies cleanly) | `adminPass` empty: `AdminSecret` fails loud "semp.adminPass must be set" (B k8s/secrets.go:81-83) | Error in both: "create: true needs a credential" | A: new error for a rare file. B: none |
| `existing: NAME` | Mounted by name; never written, never deleted (A spec/kubernetes.go:320-329; A deploy/deploy.go:362, 375). Credential fields in the spec remain and name the keys; their values are never resolved (A gen/gen.go:281-282). This is the documented use (A docs/configuration.md:1007-1022) | `adminSecret: NAME` and no `adminPass`: referenced, never built (B config/config.go:114-131; B docs/configuration.md:285-293). `monitorPass` or `psk` without `adminPass`: refused (B config/validate.go:1399-1415) | `existing` plus material = error (E11 for B). For A, material cannot mean "the password fields" (see E1) | B: `adminPass` + `existing` becomes an error (today it builds over the named Secret). A: needs the rule redefined or its `existing` specs break |
| `create` + `existing` | Error "exactly one of create or existing" (A validate/validate.go:961-963) | n/a (no block). Today a name with `adminPass` builds under that name (B config/config.go:126-127) | Error in both | B: none new |
| Name customised | `create.name` retired: error on validate, ignored by generate and deploy (A spec/kubernetes.go:83-95; A validate/validate.go:1590-1593) | `kubernetes.adminSecret: X` with `adminPass` builds Secret X (B config/config.go:125-132) | Derived name only (E8) | B: drop `adminSecret` for built Secrets; an env file that builds under a custom name now builds `<name>-credentials`, a rename the cluster sees as new object plus orphan |

### 2.2 Kubernetes image pull (broker or connector namespace)

| Situation | A today | B today | Proposed | Change for operators |
|---|---|---|---|---|
| Block absent, material absent | No `imagePullSecrets` entry (A spec/kubernetes.go:344-347) | No `kubernetes.imagePullSecret`, no user/pass: no `pullSecrets` block (B config/config.go:213-220) | Same | None |
| Block absent, material present (`image.user` + `image.pass`) | Silently unused: credentials are "consulted only when the tool is asked to build a pull secret" (A spec/image.go:21-23; A docs/configuration.md:628) | Built silently as `<name>-image-pull` (B config/config.go:198-200, 213-220; B k8s/prep.go:171, 215-221) | Error when the kubernetes section is active: "add create: true or existing: NAME" | A: files that carry `image.user/pass` with no block gain an error (they did nothing). B: add `create: true`. Docker and podman only files: no error (mine, see E2) |
| `create: true` | Builds `<name>-image-pull` (dockerconfigjson) from `image.user/pass`; requires both (A spec/kubernetes.go:344-352; A validate/validate.go:1567-1579; A gen/gen.go:471-499) | No key; equivalent is user and pass set | Builds derived name. Needs `image.user` and `image.pass` | B: new `kubernetes.secrets.imagePull.create` (E8) |
| `create` without material | Error "image-pull.create requires image.user and image.pass" (A validate/validate.go:1576-1578; A gen/gen.go:482-484) | n/a | Same error, in both | B: none |
| `existing: NAME` | Today spelled `image-pull: {name: NAME}`; referenced only, never written (A spec/kubernetes.go:235-252; A gen/gen.go:471-476) | `imagePullSecret: NAME` without credentials: referenced only (B config/config.go:213-216; B docs/configuration.md:217) | `existing: NAME`, with `image.user/pass` an error for the broker namespace. B keeps `regcred` regardless (E10), so B needs the exemption in open point O2 | A: key renamed `name` -> `existing` (E8). B: `imagePullSecret` removed |
| `create` + `existing` | `create: true` plus `name`: error "no longer accepted with create: true" (A validate/validate.go:1603-1606) | `imagePullSecret: X` with user/pass: builds X under that name (B config/config.go:213-216) | Error in both | B: custom name on a built Secret is gone |
| Name customised | Retired with create (A validate/validate.go:1603-1606) | Allowed with material (B config/config.go:213-216) | Derived name only | B: as 2.1 |

### 2.3 B operator `regcred` (A has no operator)

| Situation | A today | B today | Proposed | Change for operators |
|---|---|---|---|---|
| Block absent, material absent | n/a | No user/pass: no `regcred`, no `imagePullSecrets` block in the operator bundle (B k8s/operator.go:38-44, 148-150) | Same | None |
| Block absent, material present | n/a | Built in the operator namespace as `regcred` whenever user and pass are set, applied before the bundle, deleted by name on operator delete (B config/config.go:198-200; B k8s/operator.go:277-284, 535-548; B k8s/secrets.go:177-196) | Unchanged: stated exception (E10). Implicit, fixed name, unlabelled, instance label is the operator deployment name (K10) | None |
| `create` / `existing` / both | n/a | No key. `imagePullSecret: NAME` does not touch `regcred`: it stays built from user/pass (B docs/configuration.md:217) | No block exists for it. E10: `existing` on the broker Secret covers the broker namespace only | None, but see O2: `image.user/pass` are "material" for the broker Secret and must not trip "existing with material" |
| Name customised | n/a | Not possible, fixed `regcred` (B k8s/secrets.go:177-190) | Fixed | None |

### 2.4 TLS (A stores Secret and per-store existing Secrets; B server TLS)

A: one derived Secret `<name>-stores` holds every selected store the tool holds, keyed by store
name; a store kept in an operator's Secret names it on the store (`existing.secret`)
(A spec/kubernetes.go:331-339; A spec/defaults.go:100-125; A deploy/deploy.go:44-90).
B: `kubernetes.io/tls` Secret with `tls.crt` (cert plus CAs) and `tls.key`, digest annotation
(B k8s/secrets.go:140-162; B config/config.go:157-190).

| Situation | A today | B today | Proposed | Change for operators |
|---|---|---|---|---|
| Block absent, material absent | No stores selected or none held: no Secret rendered; remove still deletes the name (A deploy/deploy.go:360-371, 389-402) | No tls files, no name: no Secret, no `tls` block in the CR (B config/config.go:166-188) | Same | None |
| Block absent, material present | Built with no switch: `<name>-stores` is rendered whenever a selected store is held by the tool (`existing.pkcs12` or `create:`) (A spec/kubernetes.go:331-339; A gen/gen.go:1154-1163, 290-296 area; A validate/validate.go:1594-1602). The old `secrets.stores` block is retired and errors (A validate/validate.go:973-985) | Built as `<name>-tls` when `tls.cert` or `tls.certKey` is set (B config/config.go:162-164; B k8s/prep.go:170, 203-209). Cert without key refused (B config/validate.go:685-693) | Error: "add create: true or existing: NAME" in both | A: every working file that holds a store gains a mandatory `kubernetes.secrets.stores.create: true`: the widest break in the table. B: add `secrets.tls.create` (E9) |
| `create: true` | n/a (no switch today) | n/a (no switch today) | A: builds `<name>-stores`. B: builds `<name>-tls`, `tls.cert` and `tls.certKey` required (E9) | Both: new key |
| `create` without material | n/a | n/a | Error | A: new error when no selected store is tool-held. B: none |
| `existing: NAME` | Per store only: `tls.stores.<n>.existing.secret: {name, key}`; the tool holds nothing for that store; the stores volume projects each operator Secret (A spec/defaults.go:100-125; A deploy/deploy.go:44-90; A validate/validate.go:27-40). A store may also keep `create:` or `existing.pkcs12` beside it for docker/podman (A validate/stores.go:29-31) | `tlsServerSecret: NAME` with no files: referenced, not built, not rotated, not deleted (B config/config.go:157-190; B k8s/prep.go:313-318) | `existing: NAME` plus material = error. Open: how a Secret-wide `stores.existing: NAME` coexists with per-store `existing.secret {name,key}` (O3) | A: per-store spelling must stay or be replaced. A's own earlier ruling retired `secrets.stores.existing` for this reason (A validate/validate.go:973-985, 1229-1248). B: `tlsServerSecret` -> `secrets.tls.existing` |
| `create` + `existing` | `create:` beside `existing.secret` is allowed per store: kubernetes reads the Secret, docker/podman build the file (A validate/stores.go:29-31; A validate/validate.go:1243) | Files plus name: builds under that name (B docs/configuration.md:259-262) | Error on one Secret. A's per-store case is a cross-platform feature the rule would remove (O3) | A: a single env.yaml can no longer serve kubernetes (Secret) and docker (file) from one store entry |
| Name customised | `<name>-stores` fixed (A spec/kubernetes.go:337-339) | `tlsServerSecret: X` with files builds X (B config/config.go:180-188) | Derived only | B: as 2.1 |

### 2.5 B additional users

Today the Secret holds `username_<u>_password` and `username_<u>_globalaccesslevel`, and the CR
names it in `spec.extraEnvVarsSecret` (B k8s/secrets.go:126-157; B config/config.go:99-101). It
is projected with `envFrom`, so each key becomes an environment variable in the pod. K8 keeps
this as the one exception, because the operator CRD has no way to mount a Secret as files.

| Situation | A today | B today | Proposed | Change for operators |
|---|---|---|---|---|
| Block absent, material absent | n/a (A's users are credentials, see 2.1) | No users: no Secret, CR omits `extraEnvVarsSecret` (B k8s/secrets.go:134-136; B k8s/prep.go:222-229) | Same | None |
| Block absent, material present | n/a | Built as `<name>-additional-users` whenever `semp.additionalUsers` is non-empty (B k8s/prep.go:172, 224-230) | Error "add create: true or existing: NAME" | B: add `kubernetes.secrets.additionalUsers.create: true` |
| `create: true` | n/a | n/a | Builds the derived name from the users list | B: new key |
| `create` without material | n/a | Empty list: nothing built (B k8s/secrets.go:134-136) | Error | None |
| `existing: NAME` | n/a | Not possible today | `existing` plus a users list = error. The Secret must carry `username_<u>_password` and `username_<u>_globalaccesslevel`; the CR names it | B: new capability. Keys must be env-var-legal because of `envFrom` (B k8s/secrets.go:140-150) |
| `create` + `existing` | n/a | n/a | Error | None |
| Name customised | n/a | Derived only, no key exists (B config/config.go:86-101) | Derived only | None |

### 2.6 Docker compose secrets (both)

Docker has no Secret objects the tools own. Both tools declare compose `secrets:` that read an
environment variable of the compose child process; the value lives in that process only.

| Aspect | A today | B today | Proposed (mine) | Change for operators |
|---|---|---|---|---|
| What is handed over | Every credential the config references, each store the tool holds (encoded), as `VAR=value` in the compose child's environment (A cmd/solmq-conn-util/main.go:731-791; A dockergen/dockergen.go:118-142) | Admin password, each additional user, psk (HA), cert passphrase, the cert bundle (B render/render.go:562-598, 713-753; B container/manager.go:191-208) | Unchanged | None |
| Who creates | Nobody: compose reads the environment at `up`. No object is made | Same (B container/manager.go:513-540) | No block | None |
| Trigger | Every deploy; digest label of file and values recreates the container only on change (A cmd/solmq-conn-util/main.go:778-791) | Same idea, label `solace-util.deploy-sha256` (B render/render.go:694-706) | Unchanged | None |
| Who removes | `compose down`; no object to delete. A still RESOLVES credentials first: an unset or empty `-env` variable fails remove (A cmd/solmq-conn-util/main.go:733-736, 756-762 for stores placeholders) | `compose down` with placeholder values; cert files not read (B container/manager.go:196-198) but `Load` has already resolved `...Env` keys (B config/load.go:46-50) | Remove resolves no credential (gdu 3) | A docker remove: stop resolving. B: needs a load mode that does not resolve (E3) |
| Existing alternative | None | None | None: there is nothing to reference | None |
| Registry | No login; host must be logged in (E18) | Runs `<engine> login --password-stdin` before deploy when `image.user/pass` are set (B container/manager.go:451-478) | E18: no tool logs in | B docker and podman: login stops, so a private-registry pull now needs a prior host login. This is a behavior loss, not a rename (O5) |

### 2.7 Podman secrets (both)

Podman has a real secret store. Both tools create secrets in it and mount them with `Secret=`
lines in a quadlet unit.

| Aspect | A today | B today | Proposed (mine) | Change for operators |
|---|---|---|---|---|
| What is created | Each credential as `<container>-<KEY>`, plus file secrets: `application.yml`, status script, logback, `tls-<store>` (A cmd/solmq-conn-util/main.go:938-948; A gen/gen.go:568-598) | One secret per `ContainerSecret`: `<container>-` plus suffix for admin, users (`-user-<name>-password`), psk, cert passphrase, cert bundle (B render/render.go:562-598, 713-753) | Unchanged | None |
| Who creates | The tool, on every deploy | The tool, on every deploy (B container/manager.go:588-617) | No block | None |
| Trigger | Value read back and compared (podman 4.7+); only a changed secret is written, a running service is restarted only after a prompt (A cmd/solmq-conn-util/main.go:953-1010) | Same compare, `replace` when 4.7+, else remove then create (B container/manager.go:572-617, 791-853) | Unchanged | None |
| Stale ones | Deleted when the spec no longer mounts them; names read back from the old unit's `Secret=` lines (A cmd/solmq-conn-util/main.go:969-976, 1062-1068) | Not tracked: remove walks the CURRENT config only (B container/manager.go:1080-1110) (mine: a secret dropped from the env file between deploys is left in the store) | Unchanged | B gap, not a rule matter (O6) |
| Who removes | `remove`: all credentials, file secrets, stores, after the unit (A cmd/solmq-conn-util/main.go:1027-1060) | `remove`: all, after the unit; failure on the cert secret fails the teardown, others warn (B container/manager.go:1066-1110) | Unchanged | None |
| Remove resolves credentials? | No (podmanRemove does not call `ResolveCredentials`) (A cmd/solmq-conn-util/main.go:1027-1060 versus 922) | Yes, at `Load` (B config/load.go:46-50) | No (gdu 3) | B: as docker |
| Existing alternative | None: names are always derived | None | None | None |

Recommendation for docker and podman (mine): do NOT apply the create/existing block. The block
exists to say who owns an object a user could have made elsewhere. On compose there is no
object. On podman the names are derived from the container name and the tool already deletes by
that name, so the tool already owns them; an `existing` podman secret has no demand in either
tool and would add a path where `remove` must know not to delete. If an owner later wants it,
add it as a separate unit of work. What the rule should still require on all platforms is the
validation half: material set but unusable on that platform is stated, not ignored (E2).

## 3. Summary table

| Secret type | Default today in A | Default today in B | Proposed default | Breaking for whom |
|---|---|---|---|---|
| K8s credentials | Error if credentials referenced and no block (A validate/validate.go:272-284) | Built silently from `adminPass` as `<name>-admin`; no password and no name = operator generates (B config/config.go:114-146) | No block = nothing built or referenced; material without block = error; operator-generated case kept in B | B (every env file with `adminPass` adds `create: true`); A not at all unless "existing with material" is applied to its password fields (E1) |
| K8s image pull (broker) | Nothing built or referenced; `image.user/pass` unused (A spec/image.go:21-23) | Built silently as `<name>-image-pull` when user and pass set (B config/config.go:198-220) | Material without block = error on a kubernetes-enabled file | B (add `create: true`); A files that carry `image.user/pass` with no block gain a new error |
| B operator `regcred` | n/a | Built implicitly from user and pass, fixed name, deleted by name (B k8s/operator.go:277-284, 535-548) | Unchanged (E10 exception) | Nobody, if O2 is settled so `image.user/pass` do not trip "existing with material" |
| TLS | `<name>-stores` built with no switch whenever a selected store is tool-held; per-store `existing.secret` (A spec/kubernetes.go:331-339) | `<name>-tls` built when `tls.cert`/`certKey` set; name only = reference (B config/config.go:157-190) | Block required: `create: true` or `existing: NAME`; derived names only | A (new mandatory key for every file that holds a store; per-store `existing.secret` needs a decision, O3); B (`tlsServerSecret` replaced) |
| B additional users | n/a (A's users are credentials) | Built silently as `<name>-additional-users` when the list is non-empty (B k8s/prep.go:172) | Block required; stays in `extraEnvVarsSecret` (K8) | B |
| Docker compose secrets | Environment-sourced compose secrets, always, no switch (A dockergen/dockergen.go:118-142) | Same (B container/manager.go:191-208) | Unchanged, no block (mine) | B only through E18 (login removed); A docker remove stops resolving credentials |
| Podman secrets | Created on deploy, removed on remove, no switch (A cmd/solmq-conn-util/main.go:938-1060) | Same (B container/manager.go:588-617, 1080-1110) | Unchanged, no block (mine) | Nobody |

## 4. Edge cases the rule must state

E1. "Material" must be defined per secret and per platform, not as "any secret-looking key".
Three cases break a naive reading:

- A credentials: the password fields are the key list of an `existing` Secret. An `-env` field
  is mounted under its variable NAME, and with `existing` the value is never resolved
  (A consolidate/consolidate.go:94-112; A gen/gen.go:281-282; A docs/configuration.md:1007-1022).
  "existing with material = error" would reject every A file that uses `existing`. (Mine) the
  rule for A can only say: with `existing`, no value is read; with `create`, values are read.
  Or A drops value resolution for `existing` explicitly and keeps the fields as key names.
- A stores: `create:` or `existing.pkcs12` beside `existing.secret` is a supported
  cross-platform shape (A validate/stores.go:29-31). The material serves docker and podman
  while kubernetes reads the Secret.
- B `image.user/pass`: material for the broker pull Secret, yet also the input to `regcred`,
  which E10 keeps implicit. With `imagePull: existing: X` the pair must remain legal (O2).
- Scope (mine): the rule applies only when the kubernetes section is enabled. A file for docker
  or podman that sets `image.user/pass` or a Secret block must not fail on a platform it never
  reads. A checks per target already (A validate/validate.go:262-285 sits under
  `ctx.CheckKubernetes`).

E2. A credential set only through an `...Env` variable that is empty at load.

- A: `validate` warns only for an unset variable (A validate/validate.go:421-428). A deploy
  resolves only for `create`, docker and podman, and an unset or empty value is an error naming
  the variable (A gen/gen.go:281-288, 797-810; A cmd/solmq-conn-util/main.go:734, 922). With
  `existing` it is never read.
- B: every `...Env` key is resolved at `Load`, before defaults and validation, for every command
  including validate, remove and generate; an unset or empty variable is a hard error
  (B config/load.go:46-50; B config/secretref.go:73-80, comment at 62-72). After that
  `ManagesAdminSecret` tests the VALUE (B config/config.go:114-116), so the "material present"
  test and the "empty variable" error are one event. `PSKConfigured` already tests the key
  presence instead (B config/config.go:153-155).
- The rule must say (mine): material is present when the literal OR the `...Env` key is set,
  regardless of the value. An empty variable then fails only where the value is needed
  (create, docker/podman deploy) and never on `remove`, `validate` or `existing`.

E3. Remove.

- gdu 3: remove resolves no credential. A kubernetes remove does this: names only
  (A deploy/deploy.go:128-148, 361-371, 480-485; A gen/gen.go:273-280). A docker remove does
  not (A cmd/solmq-conn-util/main.go:733-736). B remove resolves at `Load` (E2).
- What remove deletes (today): A deletes the credentials Secret only if created
  (A deploy/deploy.go:362-365), the image-pull Secret only if created (:368-371), and the
  `<name>-stores` Secret ALWAYS, by derived name, even when this spec rendered none
  (A spec/kubernetes.go:331-339; A deploy/deploy.go:366-367). B deletes `ownedSecretNames`
  with `--ignore-not-found`, never a Secret it only names (B k8s/prep.go:162-174, 273-288), and
  `regcred` with the operator (B k8s/operator.go:535-548).
- Rule (mine): remove deletes exactly the Secrets whose block says `create`, by derived name,
  and never one whose block says `existing`. A's unconditional `<name>-stores` delete becomes
  "only when the stores block says create", else a stale Secret someone else made under that
  name is deleted. That is a deliberate narrowing: a leftover from an older spec would then
  need a manual delete.

E4. `generate` output.

- A: kubernetes generate resolves credentials only for `create` and prints them as
  `stringData` (A deploy/deploy.go:375-388; A gen/gen.go:281-288); an `existing` block prints no
  Secret document (A deploy/deploy.go:375). Image-pull prints a Secret only when the payload
  was built (A deploy/deploy.go:407-415).
- B: `broker generate` prints every built Secret ahead of the CR, base64, so the output is as
  sensitive as the env file (B k8s/prep.go:176-184; B cli/ops_k8s.go:661-683).
- Rule (mine): generate prints a Secret document exactly for `create` blocks, and prints
  nothing for `existing` or an absent block; the CR still references the existing name. State
  that generate output with `create` is secret-bearing in both tools.

E5. A Secret that exists already while the block says `create` (adoption).

- Both tools apply with `kubectl apply`, so a pre-existing Secret of the same name is silently
  overwritten (A deploy/deploy.go:344-346 note on apply; B k8s/prep.go:259-264 `apply`).
  Neither refuses, and neither checks an ownership label. A passes derived names to a
  namespace gate that discounts the tool's own names from "someone else's object"
  (A spec/kubernetes.go:363-373; B k8s/namespace.go:33-41). A name match therefore counts as
  "ours". (Mine: I read the call sites, not the gate library, so I did not verify whether it
  also tests a label.)
- Danger case in B today: `adminSecret: X` plus `adminPass` overwrites an X someone else made
  (B config/config.go:125-132). E11 and E8 close it for the admin Secret.
- The rule must state (mine): with `create`, an existing Secret of the derived name is replaced
  on deploy and deleted on remove; the derived name is reserved for the tool. If the owner
  wants adoption refused, the tool needs a label or annotation to tell its own Secret apart, and
  K10 already ruled out one for `regcred` only. That is a new feature, not part of this rule.

E6. Rotation (material changed between deploys).

- Kubernetes: the Secret is re-applied each deploy with the new value. Neither tool restarts the
  pods for a credential change. A documents that a redeploy does not restart pods and gives the
  command to run (A docs/configuration.md:797-800). B restarts only for a changed TLS
  certificate, detected from the `solace-util.sha256` annotation and offered with consent
  (B k8s/prep.go:330-365; B render/render.go:688-692), and notes the broker reads the admin
  password only on a fresh data volume (B k8s/adminsecret.go:19-30).
- Docker: digest label of file plus values recreates the container (A cmd/solmq-conn-util/
  main.go:778-791; B render/render.go:694-701). Podman: stored value compared, changed ones
  replaced, running service restarted after a prompt (A main.go:953-1010; B manager.go:572-617).
- Rule (mine): rotation is a deploy with new material under `create`; it never applies to
  `existing`, whose rotation is the owner's. State per platform what restarts. Moving between
  `create` and `existing` between deploys: `create` -> `existing` leaves the old derived Secret
  in the cluster (remove no longer deletes it, E3), so the doc must say to delete it by hand.

E7. The derived-name length limit (63).

- A holds every created object name to 63, including each derived Secret, and reports the
  `deployment.name` that derives the offender (A validate/validate.go:906-919). The Secret
  suffixes are `-credentials` (12), `-stores` (7), `-image-pull` (11)
  (A spec/kubernetes.go:311-317). The 63 comes from the Deployment, ConfigMap and PVC names the
  same list covers.
- B bounds `kubernetes.name` at 63 minus the operator's `-pubsubplus-p-0` suffix, so 48
  (B config/validate.go:449-458, 657-666). Its derived Secrets are `-admin` (6), `-tls` (4),
  `-image-pull` (11) and `-additional-users` (17) (B config/config.go:99-101, 125-132, 180-188,
  213-220). `-additional-users` on a 48-character name is 65. B checks typed Secret names
  against the DNS-1123 subdomain rule (253) (B config/validate.go:675-681), and derived ones
  only for collision (B config/validate.go:1424-1450). A Secret name may be up to 253, so 65 is
  legal for the Secret itself. (Mine: I did not check whether a Secret name is reused as a
  label or volume name in B's CR; if it is, 63 applies.)
- Rule (mine): the check lives where each name is derived. A keeps 63. B states the limit it
  needs (253 for a Secret) and adds the derived names to its collision check, which already has
  a slot for `additionalUsers` (B config/validate.go:1433-1435). With custom names gone, a
  too-long derived name is a `kubernetes.name` problem, and the error names `kubernetes.name`.

## 5. Recommendation and open points

Recommendation (mine): adopt the rule for Kubernetes, in this order and with these refinements.

1. Define "material" per secret in a table inside the rule, not by feel. For A credentials the
   fields stay as key names under `existing`; the "existing with material" error applies to B
   (`adminPass`, `monitorPass`, `psk`) and to image pull and TLS files, not to A's password
   positions. Reason: E1. Without it the rule breaks A's documented `existing` use.
2. Keep the block-absent error as scoped to the kubernetes section being enabled. Reason: both
   tools feed docker and podman from the same keys.
3. Remove resolves nothing and deletes only `create` Secrets, in both tools, with A docker
   remove and B `Load` changed to match gdu 3. Reason: E2 and E3 show a remove that fails on an
   empty variable is the one place the two tools still break their own rule.
4. Do not extend the block to docker or podman. Reason: no object, derived names, the tool
   already owns and deletes them. Require only that unusable material is stated.
5. Write E5 (replace on `create`, name reserved) into the docs as the explicit semantics.
   Refuse-adoption is a separate feature if wanted.
6. Treat TLS in A as its own step. It is the widest break (every file with a store) and it
   collides with a retired shape and with the cross-platform `existing.secret` (O3).

Reasons for the rule: it removes B's silent build-over-an-existing-Secret hazard (E5), makes the
two tools say ownership in the same words, and gives one answer to "what does remove delete".
Its cost is a one-time edit of every B env file that builds a Secret, and A's TLS step.

Open points that need the owner:

- O1. "Material" for A (E1): confirm A's password fields stay as key names under `existing`
  (my recommendation), or decide A's `existing` takes no password positions at all (which
  breaks the documented hand-built Secret workflow, A docs/configuration.md:1022-1031).
- O2. B `image.user/pass` with `imagePull: existing: X` or with no `imagePull` block and only an
  operator install: legal (needed for `regcred`) or refused? E10 implies legal. The rule needs
  one sentence saying `image.user/pass` are exempt from both errors in B, and that with no
  block the pair builds only `regcred`.
- O3. A TLS (E9 "Secrets reach"): is `kubernetes.secrets.stores.existing: NAME` a whole-Secret
  reference keyed by store name, living beside per-store `existing.secret {name,key}`? Both?
  A retired this block on purpose (A validate/validate.go:973-985, 1229-1248). And does
  `create` beside `existing.secret` stay legal for the cross-platform store?
- O4. A with a store held only in an operator Secret and no tool-held store: block absent is
  correct (nothing built), but `create: true` then errors. Confirm that is wanted.
- O5. E18 removes B's `docker login` and `podman login` (B container/manager.go:451-478). Do the
  owner and B's operators accept the loss, or is a transitional warning wanted when `image.user`
  is set on docker or podman?
- O6. Not rule matters, found on the way: B podman remove does not clean a secret dropped from
  the env file between deploys (A tracks it from the unit); A docker remove resolves `-env`
  credentials. Fix inside this work or separately?
- O7. Adoption (E5): accept silent replace of a same-name Secret under `create`, or require a
  tool label so a foreign Secret is refused?
- O8. Name migration for B: an env file that built under a custom name now builds the derived
  name. The old Secret is orphaned and the CR switches. Is a one-line load warning that names the
  old Secret enough?
