# go-deploy-utility

Shared Go module `github.com/hafio/go-deploy-utility` for solmq-conn-util and
solace-util (repo paths: docs/developer.md). GPLv3. init-devops library shape:
no image, no binaries; consumers fetch the tag.

## Conventions

- Shared code imports no tool type or config, has no globals, registers no flags
  and never calls `os.Exit`. Builders and parsers return `runner.Cmd` values or
  data and never execute. Full rules: docs/developer.md "Rules for shared code".
- A package needs two consumers. Standard library first; any dependency is
  pinned exactly with its reason in the change.
- Least privilege: every command runs with the lowest permission that does the
  job, and a step that needs more only for convenience warns when refused and
  never blocks the operation (docs/developer.md "Rules for shared code").
- Every build and adoption step is planned and approved on its own
  (docs/developer.md).
- An adoption step lands here first, then one commit per tool on the branch
  `adopt-go-deploy-utility`. Tools link to this checkout through an untracked
  `go.work` until a tag is pinned (docs/consuming.md).
- The parity body is canonical here, in `paritydoc/body.md`. Both tools'
  `docs/parity.md` carry it unchanged below their own notice, checked by
  `paritydoc.Check`.
- README.md and every `docs/*.md` carry the support notice directly under the
  H1. `docs/handoff/` is planning history: mark items superseded, never rewrite.
- Lifted tests keep their names; docs/test.md lists every test and its source.

## Process

- The user runs every gate: `./scripts/dev.sh all scan`, or `full` before a tag.
  Never run build, vet, test, cov, scan or graphify here.
- Commit on `main` only when asked. Never push or tag. A pushed tag is permanent
  in the module proxy: fix forward with a new patch tag and `retract`.
- Plain ASCII everywhere.

| Topic | Where |
| --- | --- |
| Rules, package layout, build and adoption order | docs/developer.md |
| Tests and their sources | docs/test.md |
| Pinning, go.work, release order | docs/consuming.md |
| Planning history and the original stage plan | docs/handoff/ |
