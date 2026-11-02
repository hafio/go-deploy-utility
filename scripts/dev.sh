#!/usr/bin/env bash
# Dev tasks for go-deploy-utility. The only place that knows how to build/vet/
# test/scan this repo. CI calls task names only. Keep dev.ps1 behaviourally
# identical.
#
# go-deploy-utility is a Go library (init-devops library shape): it ships no
# image and no binaries, so build is a compile check that writes nothing, and
# the Docker tasks (image, up, down) do not apply and are omitted.
set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
LOG_DIR="$SCRIPT_DIR/logs"
mkdir -p "$LOG_DIR"
cd "$REPO_ROOT"

export NO_COLOR=1

# Toolchain parity with CI: when go.mod carries a `toolchain` directive it is
# the single pin, and GOTOOLCHAIN makes any go binary honour it exactly. Set
# only when unset, so an exported value wins.
if [ -z "${GOTOOLCHAIN:-}" ] && [ -f "$REPO_ROOT/go.mod" ]; then
  t="$(sed -n 's/^toolchain //p' "$REPO_ROOT/go.mod")"
  [ -n "$t" ] && export GOTOOLCHAIN="$t"
fi

# --- output helpers ---------------------------------------------------------
c() { printf '\033[%sm%s\033[0m\n' "$1" "$2"; }
step() { c '1;36' "==> $*"; }
ok()   { c '1;32' "ok: $*"; }
warn() { c '1;33' "warn: $*"; }
die()  { c '1;31' "error: $*"; exit 1; }

now() { date +%Y-%m-%dT%H:%M:%S%z; }

# Truncate this task's log with a header, then everything tees onto it.
log_begin() {
  printf '=== %s | %s ===\n' "$(now)" "$1" > "$LOG_DIR/$1.log"
}

# finish <task> <exit-code> <elapsed-seconds>
finish() {
  local task=$1 code=$2 secs=$3 status
  if [ "$code" -eq 0 ]; then status=OK; else status="FAILED (exit $code)"; fi
  printf '%s | %s | %ss | %s\n' "$(now)" "$task" "$secs" "$status" \
    | tee -a "$LOG_DIR/$task.log"
}

# Strip ANSI/CSI so logs stay readable plain text.
strip_csi() { sed -E $'s/\x1b\\[[0-9;?]*[a-zA-Z]//g'; }

# run <task> <cmd...> -- tees combined output, returns the command's code.
run() {
  local task=$1; shift
  "$@" 2>&1 | strip_csi | tee -a "$LOG_DIR/$task.log"
  return "${PIPESTATUS[0]}"
}

# --- tasks ------------------------------------------------------------------
# A library has nothing to ship, so build compiles every package and writes
# nothing: no dist/, no version stamp, and TARGET_OS/TARGET_ARCH are ignored.
task_build() { run build go build ./...; }

# Static analysis, plus a formatting CHECK. `gofmt -l` only LISTS unformatted
# files -- it never writes -- so drift is REPORTED here, never silently fixed:
# a gate that repairs what it is measuring proves nothing. vet and the format
# check both always run and both report, so one pass shows every problem
# instead of hiding vet errors behind a formatting failure.
task_vet() {
  local code=0
  run vet go vet ./... || code=$?
  if ! command -v gofmt >/dev/null 2>&1; then
    echo 'gofmt not found on PATH: it ships with the Go toolchain, so this is a broken install' \
      | tee -a "$LOG_DIR/vet.log"
    return 1
  fi
  local unformatted
  unformatted=$(gofmt -l .)
  if [ -n "$unformatted" ]; then
    {
      echo 'gofmt: these files are not formatted:'
      echo "$unformatted"
      echo "fix with: gofmt -w $(echo "$unformatted" | tr '\n' ' ' | sed 's/ *$//')"
    } | tee -a "$LOG_DIR/vet.log"
    if [ "$code" -eq 0 ]; then code=1; fi
  fi
  return "$code"
}

# The race detector is opt-in in both scripts (DEV_RACE=1). It needs cgo, which
# a stock Windows host lacks, and a default that differed between dev.sh and
# dev.ps1 would break the rule that the two behave identically.
task_test() {
  if [ "${DEV_RACE:-}" = "1" ]; then
    run test go test -race -count=1 ./...
  else
    run test go test -count=1 ./...
  fi
}

task_cov() {
  # Coverage profile -> HTML, and PRINT the total so the footer captures it.
  # The previous total in logs/cov.log is the floor (local only -- CI is a
  # fresh checkout with no prior log). -coverpkg credits cross-package
  # execution, so a package exercised through another one's tests still counts.
  run cov go test -coverpkg=./... -coverprofile=coverage.out -count=1 ./... || return $?
  go tool cover -html=coverage.out -o coverage.html
  go tool cover -func=coverage.out | tail -n1 | tee -a "$LOG_DIR/cov.log"
}

# One task, every applicable check. A library ships no image, so this is the Go
# scan alone. govulncheck reports only vulns reachable from called code, and
# any finding fails the run, because it cannot tell fixable from unfixable. That
# is a known gap against the contract's "unfixable warns" rule, accepted while
# the module depends on the standard library alone, where a toolchain bump fixes.
# NOT `go run ...@version`: that form ignores this go.mod and its toolchain
# pin. `go tool` builds the scanner in this module, at the version go.mod pins.
task_scan() {
  run scan go tool govulncheck ./...
}

# Local only: the graph is a developer artifact, not a CI output.
task_graphify() {
  [ -n "${CI:-}" ] && { warn "graphify is local-only; skipping in CI"; return 0; }
  command -v graphify >/dev/null || { warn "graphify not on PATH; skipping"; return 0; }
  run graphify graphify update .
}

# --- dispatch ---------------------------------------------------------------
ALL="build vet test"
FULL="build vet test cov scan graphify"

usage() {
  cat <<EOF
usage: $(basename "$0") <task>...

  build vet test cov scan graphify
  all   = $ALL            (what CI runs, as: all scan)
  full  = $FULL   (pre-tag sweep)

  DEV_RACE=1 adds -race to test (needs cgo).
EOF
}

expand() {
  case "$1" in
    all)  echo "$ALL" ;;
    full) echo "$FULL" ;;
    *)    echo "$1" ;;
  esac
}

[ $# -eq 0 ] && { usage; exit 0; }
case "${1:-}" in -h|--help|help) usage; exit 0 ;; esac

TASKS=""
for a in "$@"; do TASKS="$TASKS $(expand "$a")"; done

FAILED=0
for task in $TASKS; do
  type "task_$task" >/dev/null 2>&1 || die "unknown task: $task"
  step "$task"
  log_begin "$task"
  start=$SECONDS
  code=0
  "task_$task" || code=$?
  finish "$task" "$code" "$((SECONDS - start))"
  if [ "$code" -ne 0 ]; then
    FAILED=1
    warn "$task failed; stopping"
    break   # build/vet/test/scan are all fatal
  fi
  ok "$task"
done
exit "$FAILED"
