# go-deploy-utility

> [!WARNING]
> **Not a supported Solace product.** `go-deploy-utility` was created by Solace
> Professional Services and is supported only by Solace Professional Services --
> not by Solace Support. For help with this module, contact your Solace
> Professional Services representative rather than opening a Solace Support
> case. This notice covers this module only, not the Solace products that the
> tools built on it deploy.

`go-deploy-utility` is the shared Go module behind two Solace Professional
Services tools. `solmq-conn-util` deploys the Solace PubSub+ Connector for IBM
MQ, and `solace-util` deploys the Solace PubSub+ Event Broker. Both drive Docker,
Podman and Kubernetes with the same external commands. This module holds the
code they share.

Import path: `github.com/hafio/go-deploy-utility`.

## Status

The module is at v0. The first packages are available, and the rest land one at
a time in the order [docs/developer.md](docs/developer.md) lists. Any v0 release
may change the API.

## Packages

| Package | Purpose | Status |
| --- | --- | --- |
| `exitcode` | Exit codes 0, 1 and 2, usage and child-exit markers | available |
| `notice`, `notice/noticetest` | Support notice text and the doc gate that checks it | available |
| `paritydoc` | Canonical `docs/parity.md` body and the check both tools run against it | available |
| `abbrev` | Short forms for commands, platforms and roles | available |
| `confirm` | Confirmation prompts for destructive commands | available |
| `execguard`, `envref` | Command allowlist and token checks; environment-name references | available |
| `runner`, `runner/runnertest` | The command runner seam and its test fake | available |
| `containerart/compose`, `containerart/quadlet`, `containerart/artifacttest` | Compose and quadlet escaping and quoting, hardening lines, secret wiring, posture checks | available |
| `podman`, `systemd`, `rootless` | Podman and systemd commands, rootless rules, host checks | available |
| `containercli`, `containercli/inspect`, `watch` | Docker and compose commands, inspect decoding, the watch loop | planned |
| `kube`, `kube/rbac`, `kube/nsgate` | kubectl commands, permission probes, the namespace gate | planned |

## Using it

Pin an exact version in `go.mod`. [docs/consuming.md](docs/consuming.md)
covers local development against a checkout and the release order.

## Development

[docs/developer.md](docs/developer.md) describes the rules for shared code and
how to add a package. [docs/test.md](docs/test.md) lists every test and where it
came from. Gates run through `./scripts/dev.sh` or `./scripts/dev.ps1`.

## Licence

GNU General Public License v3.0. See [LICENSE](LICENSE).
