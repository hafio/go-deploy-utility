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
`paritydoc/body.md` and copied into both tools by `scripts/dev-all.sh pin`
before the tag, so both tools are tested against it first.

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

In workspace mode Go reads the `go` and `toolchain` lines of `go.work`, not the
tool's `go.mod`, so a `go.work` without a `toolchain` line runs the gates on
whatever Go is installed and a scan can report a standard-library fix the pinned
toolchain already has. Copy the tool's `toolchain` line into its `go.work`, or run
the gates with `GOWORK=off` once the required tag is published.

Both `go.work` and `go.work.sum` are gitignored in every repo. Keep the file in
the tool's root, never in a shared parent directory, because Go applies the
nearest `go.work` above the current directory to every module below it.

The workspace hides a wrong `require` line, since the local checkout satisfies
every import. `go mod tidy` ignores the workspace and resolves through the
module proxy, so it fails until the required tag exists. `scripts/dev-all.sh`
prints one line per tool before its tasks, and warns when this checkout holds
code the tool's pinned version does not contain.

## What each version adds

- v0.1.0: the first release.
- v0.1.1: tests that hold on Windows; no API change.
- v0.1.2: `watch.LoopFrames`, which runs each refresh into a buffer and draws
  the clear, header and frame in one write, so a `status --watch` screen keeps
  its last report while a slow refresh is collected, and Ctrl-C mid-refresh
  leaves it in place. Use it instead of `watch.Loop` when a refresh takes
  noticeable time; `Loop` keeps its signature and output. Moving to it, the
  refresh writes to the frame writer instead of stdout and returns its error
  rather than printing it, and the tool re-syncs `docs/parity.md`, whose rule 13
  now says a refresh is collected before the screen is cleared. Both loops trim a
  failed refresh's error before ` -- trying again in <d>`, so an error that ends
  in a command's newline no longer pushes the retry onto a line of its own.

The module stays on v0.x while both tools ship as v1. A tool's version is a
promise to its users; this module's is a promise to its two consumers, and
staying below v1 keeps its API free to change in step with them.

## Release and merge order

Both tools adopt a change through their `go.work` before this repo is tagged,
so a tag is cut only once all three pass together. `scripts/dev-all.sh` runs
the steps across the three repos, and `scripts/dev-all.ps1` is its Windows
mirror with the same words.

1. Make the change in this repo, committed or not, and leave it untagged.
2. Run `./scripts/dev-all.sh pin v0.x.y`. It sets each tool's `require` to the
   coming tag, which works before the tag exists because each tool's `go.work`
   replaces every version with this checkout, and copies `paritydoc/body.md`
   into each tool's `docs/parity.md`, printing any change to the shared rules.
3. Adopt the change in each tool on its adoption branch. A change to the parity
   body must hold in both tools before step 5; an additive API change can be
   adopted later, since the tools still build without using it.
4. Run `./scripts/dev-all.sh regen all scan cov` until all three pass, and
   `.\scripts\dev-all.ps1 regen all scan cov` on Windows.
5. In this repo, run `GOWORK=off ./scripts/dev.sh full` on a clean checkout and
   `.\scripts\dev.ps1 full` on a Windows clean checkout, since the tag
   pipeline runs the gates on both. Commit, then tag `v0.x.y` and push the tag.
   The tag pipeline runs the gates and publishes a notes-only release.
6. Run `./scripts/dev-all.sh resolve`, which runs `GOWORK=off go mod tidy` in
   each tool to write `go.sum` from the module proxy.
7. Run `GOWORK=off ./scripts/dev-all.sh all scan`, and the same on Windows, so
   each tool is checked against the published tag instead of this checkout.
8. Commit each tool, merge its adoption branch to `main`, then tag the tool as
   usual.

A pushed tag is public at once, and the Go module proxy keeps it permanently.
Never move or reuse a tag. A bad release is fixed by a new patch tag, and a
`retract` line in this module's `go.mod` marks the bad one.
