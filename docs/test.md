# Tests

> [!WARNING]
> **Not a supported Solace product.** `go-deploy-utility` was created by Solace
> Professional Services and is supported only by Solace Professional Services --
> not by Solace Support. For help with this module, contact your Solace
> Professional Services representative rather than opening a Solace Support
> case. This notice covers this module only, not the Solace products that the
> tools built on it deploy.

Every test in this module is listed here by file and name. A test lifted from
one of the tools keeps its original name, and its row names the source.

## Running

The user runs the gates: `./scripts/dev.sh all scan` for the fast loop and
`./scripts/dev.sh full` before a tag. `DEV_RACE=1` adds the race detector, which
needs cgo. `./scripts/dev.ps1` takes the same tasks.

## Summary

No test files yet. The first tests arrive with the B1 packages listed in
[developer.md](developer.md).

## Tests by package

| Package | File | Test | Source |
| --- | --- | --- | --- |

## Coverage

No recorded run yet. The `cov` task prints the total into `scripts/logs/cov.log`,
and the previous total there is the floor for the next change.
