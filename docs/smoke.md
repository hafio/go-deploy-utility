# Smoke checks

> [!WARNING]
> **Not a supported Solace product.** `go-deploy-utility` was created by Solace
> Professional Services and is supported only by Solace Professional Services --
> not by Solace Support. For help with this module, contact your Solace
> Professional Services representative rather than opening a Solace Support
> case. This notice covers this module only, not the Solace products that the
> tools built on it deploy.

Hand checks for each adoption step, run against real engines once the gates pass. Unit tests mock
the engines, so only a real host proves this behaviour. Tick a box when it holds, and record the
host and versions under "Runs" below.

## Environment

- A Linux host with docker and compose v2 (`docker compose version`).
- Rootless podman 4.7 or later (`podman version`): 4.7 brings `secret inspect --showsecret`,
  which the secret comparison relies on.
- Optional: `docker-compose` v1 and nerdctl, for the compose-from-stdin check.
- Both tools built from `adopt-go-deploy-utility`, linked to this checkout through `go.work`
  (docs/consuming.md).
- A working env file for each tool, with at least one credential you can rotate.

## S7: redeploy asks before any write (not yet run)

Run once per tool and per platform: `solmq-conn-util deploy --platform docker|podman` and
`solace-util broker deploy`. Docker plus rootless podman for one tool covers most of the shared
code.

Baseline: deploy once and let the container come up. Note its start time with
`docker inspect -f '{{.State.StartedAt}}' <name>` (or `podman inspect`). On podman, also note the
UPDATED column of `podman secret ls` and the mtime of the quadlet file.

| Check | solmq-conn-util docker | solmq-conn-util podman | solace-util docker | solace-util podman |
| --- | --- | --- | --- | --- |
| 1. Unchanged redeploy: no question; start time, secrets and unit file unchanged | [ ] | [ ] | [ ] | [ ] |
| 2. Rotated credential: asks before restarting or recreating anything | [ ] | [ ] | [ ] | [ ] |
| 3. Check 2 with `< /dev/null`: exits 1, names `--no-prompt`, nothing changed | [ ] | [ ] | [ ] | [ ] |
| 4. Check 2, answered "n": unit, secrets and container untouched | [ ] | [ ] | [ ] | [ ] |
| 5. Check 2 with `--no-prompt`: applies and restarts without asking | [ ] | [ ] | [ ] | [ ] |

Compose from stdin (docker only; the compose file now goes in through `-f -`):

- [ ] `docker compose` v2 (covered by the checks above)
- [ ] `docker-compose` v1, if installed: one deploy succeeds
- [ ] nerdctl, if installed: one deploy succeeds

## S8: the shared watch loop (not yet run)

- [ ] `solmq-conn-util status all -w` on a terminal: redraws in place under
      `Every 5s: solmq-conn-util status all   <date time>   Ctrl-C to stop`.
- [ ] `solmq-conn-util status all -w | cat`: a `=== <time> ===` section line before each refresh,
      and no escape bytes (`| cat -v` shows no `^[`).
- [ ] Ctrl-C on either: exits 0 (`echo $?`), with no trailing blank line.
- [ ] `solace-util broker status --watch` on a terminal and piped: the same two shapes, with the
      title `solace-util broker status`.
- [ ] `solace-util broker status --watch --interval 2h`: exits 2 with "is above 1h".

## S9: kube, rbac and the namespace gate (not yet run)

Needs a cluster (kind, or OpenShift with `oc`) and two identities: a cluster admin, and one
bound only to `edit` (plus the broker resource for solace-util) in an existing namespace.

| Check | solmq-conn-util | solace-util |
| --- | --- | --- |
| 1. Namespace-scoped identity deploys into an existing namespace: no namespace probe, no create | [ ] | [ ] |
| 2. Admin deploys into an absent namespace: it is created on its own, then the deploy | [ ] | [ ] |
| 3. Namespace-scoped identity, absent namespace: refused, names `create namespace`, nothing applied | [ ] | [ ] |
| 4. Identity without `get` on the rollout target: start or restart warns about the wait and exits 0 | [ ] | [ ] |
| 5. Remove with a foreign object in the namespace: kept, the object named | [ ] | [ ] |
| 6. Remove at a terminal, Ctrl-D at the namespace question: kept, exit 0 | [ ] | [ ] |
| 7. Remove with `--no-prompt` on an emptied namespace: deleted | [ ] | [ ] |
| 8. Remove as the namespace-scoped identity: `delete namespaces` refused, kept with a warning, exit 0 | [ ] | [ ] |
| 9. `status`, `logs` and the shell or cli run with a read-only identity (no probes) | [ ] | [ ] |

solace-util only:

- [ ] `broker deploy` by an identity that may create but not get the broker resource: warns that
      the read-back was refused, exits 0.
- [ ] `broker remove --delete-data` with retained claims of another broker: kept.

## Runs

| Date | Host | docker / compose | podman | Steps | Result |
| --- | --- | --- | --- | --- | --- |
