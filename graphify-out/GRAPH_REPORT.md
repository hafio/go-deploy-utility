# Graph Report - go-deploy-utility  (2026-11-02)

## Corpus Check
- 13 files · ~17,983 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 2 file(s) not represented in the graph (top: (none) 2)

## Summary
- 124 nodes · 150 edges · 12 communities (9 shown, 3 thin omitted)
- Extraction: 100% EXTRACTED · 0% INFERRED · 0% AMBIGUOUS
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `d6423f31`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- dev.sh
- Developer guide
- Handoff: aligning solmq-conn-util and solace-util on Docker, Podman and Kubernetes
- Appendix C: deferred code stages (not in this plan, separate approval)
- dev.ps1
- Appendix A: research notes (licensing and shared module)
- parity-body.md
- Tests
- go-deploy-utility
- handoff/README.md
- github.com/hafio/go-deploy-utility

## God Nodes (most connected - your core abstractions)
1. `Appendix C: deferred code stages (not in this plan, separate approval)` - 10 edges
2. `dev.sh script` - 9 edges
3. `Handoff: aligning solmq-conn-util and solace-util on Docker, Podman and Kubernetes` - 9 edges
4. `Appendix A: research notes (licensing and shared module)` - 9 edges
5. `Invoke-Logged()` - 8 edges
6. `run()` - 8 edges
7. `Get-Log()` - 6 edges
8. `go-deploy-utility` - 6 edges
9. `Developer guide` - 6 edges
10. `c()` - 5 edges

## Surprising Connections (you probably didn't know these)
- None detected - all connections are within the same source files.

## Import Cycles
- None detected.

## Communities (12 total, 3 thin omitted)

### Community 0 - "dev.sh"
Cohesion: 0.19
Nodes (19): c(), expand(), finish(), log_begin(), NO_COLOR, run(), dev.sh script, die() (+11 more)

### Community 1 - "Developer guide"
Cohesion: 0.11
Nodes (16): Consuming go-deploy-utility, Develop against a local checkout, Pin an exact version, Release and merge order, Adding a package, Adoption order, Dev tasks, Developer guide (+8 more)

### Community 2 - "Handoff: aligning solmq-conn-util and solace-util on Docker, Podman and Kubernetes"
Cohesion: 0.11
Nodes (18): 1. What this is, 2. State at export, 3. House rules that bind the next session, 4. Decisions already taken, 5.1 Same in both (the contract), 5.2 Different by policy, 5.3 Runner seams (why nothing is shared yet), 5. Exact similarities and differences (+10 more)

### Community 3 - "Appendix C: deferred code stages (not in this plan, separate approval)"
Cohesion: 0.11
Nodes (17): Align docker, podman and kubernetes behaviour across solmq-conn-util and solace-util, Appendix B: B tests and doc lines pinned to the on-disk compose file (Stage 2 input), Appendix C: deferred code stages (not in this plan, separate approval), Context, Files, Open decisions (recorded in parity.md, not settled here), parity.md content, Stage 0: capture (both repos, docs only) (+9 more)

### Community 4 - "dev.ps1"
Cohesion: 0.22
Nodes (12): Get-Log(), Get-Now(), Invoke-Logged(), Task-build(), Task-cov(), Task-graphify(), Task-scan(), Task-test() (+4 more)

### Community 5 - "Appendix A: research notes (licensing and shared module)"
Cohesion: 0.20
Nodes (10): A.1 Facts that shape the answer, A.2 Enforcement options, weakest to strongest, A.3 Key format and verification (options 1 and 2), A.4 Where the key comes from, A.5 Where the gate plugs in, A.6 Shared module, A.7.1 Runtime layer: docker, podman, kubectl, A.7 Code the two repos could share (synergy) (+2 more)

### Community 6 - "parity-body.md"
Cohesion: 0.33
Nodes (5): Contract, Decisions, Divergences, Inherent differences, Runner seams

### Community 7 - "Tests"
Cohesion: 0.40
Nodes (5): Coverage, Running, Summary, Tests, Tests by package

### Community 8 - "go-deploy-utility"
Cohesion: 0.50
Nodes (3): Conventions, go-deploy-utility, Process

## Knowledge Gaps
- **2 isolated node(s):** `github.com/hafio/go-deploy-utility`, `NO_COLOR`
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 72 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **3 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Appendix A: research notes (licensing and shared module)` connect `Appendix A: research notes (licensing and shared module)` to `Appendix C: deferred code stages (not in this plan, separate approval)`?**
  _High betweenness centrality (0.026) - this node is a cross-community bridge._
- **What connects `github.com/hafio/go-deploy-utility`, `NO_COLOR` to the rest of the system?**
  _2 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Developer guide` be split into smaller, more focused modules?**
  _Cohesion score 0.11052631578947368 - nodes in this community are weakly interconnected._
- **Should `Handoff: aligning solmq-conn-util and solace-util on Docker, Podman and Kubernetes` be split into smaller, more focused modules?**
  _Cohesion score 0.10526315789473684 - nodes in this community are weakly interconnected._
- **Should `Appendix C: deferred code stages (not in this plan, separate approval)` be split into smaller, more focused modules?**
  _Cohesion score 0.1111111111111111 - nodes in this community are weakly interconnected._