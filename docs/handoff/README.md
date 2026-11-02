# go-deploy-utility: handoff for the solmq-conn-util / solace-util alignment

> Superseded 2026-11-02: these files moved to `docs/handoff/` as planning history. The
> user chose to build this module first and adopt it per component; the live plan is
> `docs/developer.md`. Statements below about this directory describe the export only.

Exported 2026-11-02 from a Claude Code session (Fable 5.1). The next session runs on
Opus 5.5 and starts from these files. This directory is not a git repository and holds no
code yet.

| File | What it is |
|------|------------|
| `HANDOFF.md` | Read first. Context, state of both repos at export, house rules, decisions already taken, what was scanned (with file and line anchors), the next stages, and the prompting policy for both the process and the product. |
| `PLAN.md` | Verbatim copy of the approved plan file. Stage 0 (done) plus Appendix A (licensing and shared-module research), Appendix B (solace-util tests pinned to the on-disk compose file), Appendix C (deferred Stages 1 to 7 and open decisions). |
| `parity-body.md` | Verbatim shared body of `docs/parity.md` as committed in both repos. The exact similarities (Contract, 14 rules), exact differences (Divergences, 14 rows), Decisions, Inherent differences and Runner seams. |

Suggested opening line for the next session:

    Read /opt/claude-workspace/projects/go-deploy-utility/HANDOFF.md, then PLAN.md Appendix C.
    Plan Stage <N> in detail before touching code. The user runs all gates and commits.

About the directory name: the plan proposes a shared Go module (PLAN.md Appendix A.8, name
open, `solace-shared` was the placeholder) only after Stages 1 to 6 converge the two tools.
If this directory is meant to become that module, nothing here presumes its layout yet.
