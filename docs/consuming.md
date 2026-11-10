# Consuming go-deploy-utility

> [!WARNING]
> **Not a supported Solace product.** `go-deploy-utility` was created by Solace
> Professional Services and is supported only by Solace Professional Services --
> not by Solace Support. For help with this module, contact your Solace
> Professional Services representative rather than opening a Solace Support
> case. This notice covers this module only, not the Solace products that the
> tools built on it deploy.

This page covers how `solmq-conn-util` and `solace-util` depend on this module.

## Pin an exact version

Each tool requires one exact tag in its `go.mod`, such as
`require github.com/hafio/go-deploy-utility v0.1.0`. Dependabot manages GitHub
Actions only, so a module bump is always a deliberate commit with gates re-run.
A tool built against a new module version must be rebuilt and redeployed to
pick it up.

The module's `go` line is `go 1.27` with no patch level, because it is a floor
for every consumer. Its `toolchain` line applies to this repo only.

Each tool's tests check its own `docs/parity.md` with `paritydoc.Check` against
the body pinned with the module. A change to the shared parity text is made in
`paritydoc/body.md`, tagged, and then picked up by both tools in the same step.

## Develop against a local checkout

While a component is being adopted, a tool builds against the working copy of
this module through an untracked `go.work` at the tool's repo root:

```
go 1.27

use .

replace github.com/hafio/go-deploy-utility => ../go-deploy-utility
```

The `replace` points every version of this module at the checkout, including the
exact tag the tool's `go.mod` requires before that tag exists. Listing the
checkout under `use` instead is not enough while the tag is missing: Go still
reads the required version's `go.mod` to build the module graph, and an
unpublished version fails with `unknown revision`.

Both `go.work` and `go.work.sum` are gitignored in every repo. Keep the file in
the tool's root, never in a shared parent directory, because Go applies the
nearest `go.work` above the current directory to every module below it.

The workspace hides a wrong `require` line, since the local checkout satisfies
every import. `go mod tidy` ignores the workspace and resolves through the
module proxy, so it fails until the required tag exists.

## What each version adds

- v0.1.0: the first release.
- v0.1.1: tests that hold on Windows; no API change.
- v0.1.2: `watch.LoopFrames`, which runs each refresh into a buffer and draws
  the clear, header and frame in one write, so a `status --watch` screen keeps
  its last report while a slow refresh is collected, and Ctrl-C mid-refresh
  leaves it in place. Use it instead of `watch.Loop` when a refresh takes
  noticeable time; `Loop` is unchanged. Moving to it, the refresh writes to the
  frame writer instead of stdout and returns its error rather than printing it,
  and the tool re-syncs `docs/parity.md`, whose rule 13 now says a refresh is
  collected before the screen is cleared.

## Release and merge order

1. In this repo, run `GOWORK=off ./scripts/dev.sh full` on a clean checkout,
   and `.\scripts\dev.ps1 full` on a Windows clean checkout: the tag pipeline
   runs the gates on both, and a test that holds only on Linux fails there.
2. Tag this repo `v0.x.y` and push the tag. The tag pipeline runs the gates and
   publishes a notes-only release.
3. In each tool, set the exact `require` to the new tag.
4. In each tool, run `GOWORK=off ./scripts/dev.sh all scan`.
5. Merge each tool's `adopt-go-deploy-utility` branch to `main`, then tag the
   tool as usual.

A pushed tag is public at once, and the Go module proxy keeps it permanently.
Never move or reuse a tag. A bad release is fixed by a new patch tag, and a
`retract` line in this module's `go.mod` marks the bad one.
