#!/usr/bin/env bash
# dev-all.sh -- run the same dev.sh tasks in all three repos of this workspace:
# go-deploy-utility first (it lands first), then solmq-conn-util, then solace-util;
# and move both tools' pin on this module. Committed, but outside the dev-script
# contract: CI never calls it, since a runner checks out only this repo.
# dev-all.ps1 is its Windows mirror.
#
#   ./scripts/dev-all.sh all scan
#   ./scripts/dev-all.sh regen all scan cov
#   DEV_RACE=1 ./scripts/dev-all.sh test
#   GOWORK=off ./scripts/dev-all.sh all scan   # the tools against their published pin
#   ./scripts/dev-all.sh pin v0.1.3            # before the tag: require it, re-sync parity.md
#   ./scripts/dev-all.sh resolve               # after the tag: GOWORK=off go mod tidy
#
# Task words: each repo receives only the words its own dev.sh knows (its task_*
# functions and the aggregates in its argument switch, such as all, full or
# binaries), in the order given; a repo that knows none of them is skipped. A word
# no repo knows is refused before anything runs. Every repo runs even when an
# earlier one fails, and the exit status is 1 if any failed. Environment variables
# (DEV_RACE, GOWORK, ...) pass through unchanged. Before the tasks, one line per
# tool says which go-deploy-utility code that tool is tested with.
#
# pin and resolve each stand alone and change files in the two tools only:
#   pin <version>  sets each tool's require to <version>, which need not be tagged
#                  yet (each tool's go.work points at this checkout), and copies
#                  paritydoc/body.md under each tool's own docs/parity.md notice,
#                  printing what changed in the shared rules.
#   resolve        once the required tag is pushed, runs GOWORK=off go mod tidy in
#                  each tool, so go.sum comes from the module proxy.
# The release order that uses them is in docs/consuming.md.
#
# The tools are expected next to this repo; DEV_ALL_SOLMQ_DIR and
# DEV_ALL_SOLACE_UTIL_DIR point at checkouts elsewhere.
set -uo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
MODULE=github.com/hafio/go-deploy-utility
SEMVER='^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$'
SOLMQ_DIR="${DEV_ALL_SOLMQ_DIR:-$HERE/../solace-ibmmq-connector-helper}"
SOLACE_UTIL_DIR="${DEV_ALL_SOLACE_UTIL_DIR:-$HERE/../solace-cnt-scripts}"
REPOS=("$HERE" "$SOLMQ_DIR" "$SOLACE_UTIL_DIR")
TOOLS=("$SOLMQ_DIR" "$SOLACE_UTIL_DIR")

usage() {
  awk 'NR > 1 && /^#/ { sub(/^# ?/, ""); print; next } NR > 1 { exit }' "${BASH_SOURCE[0]}"
}

die() {
  echo "dev-all: $*" >&2
  exit 2
}

# name <repo>: the repo's directory name, as the summary shows it.
name() { basename "$(cd "$1" && pwd)"; }

# known <script>: every word the script accepts as a task or an aggregate.
known() {
  local script=$1
  grep -oE '^task_[a-z_]+\(\)' "$script" | sed -E 's/^task_([a-z_]+)\(\)$/\1/'
  grep -oE '^[[:space:]]*[a-z_]+\)[[:space:]].*(ALL|FULL|tasks)' "$script" |
    sed -E 's/^[[:space:]]*([a-z_]+)\).*/\1/'
}

# required <tool>: the version of this module the tool's go.mod requires, in either
# require form; empty when it requires none.
required() {
  sed -nE 's#^(require[[:space:]]+)?[[:space:]]*github\.com/hafio/go-deploy-utility[[:space:]]+(v[^[:space:]]+).*#\2#p' \
    "$1/go.mod" | head -n 1
}

# is_tag <version>: the version is a tag in this checkout.
is_tag() { git -C "$HERE" rev-parse -q --verify "refs/tags/$1" >/dev/null 2>&1; }

# notice_end <parity.md>: the number of the blank line that closes the support
# notice, after which the shared body starts; fails when there is no such line.
notice_end() {
  awk '/^>/ { inq = 1; next }
       inq && /^\r?$/ { print NR; found = 1; exit }
       inq { exit }
       END { exit !found }' "$1"
}

# sync_parity <tool>: replace everything below the tool's own title and notice in
# docs/parity.md with paritydoc/body.md and show the change. Returns 0 when the
# body changed, 1 when it was already current.
sync_parity() {
  local tool=$1 doc="$1/docs/parity.md" end tmp
  end=$(notice_end "$doc")
  tmp=$(mktemp)
  { head -n "$end" "$doc"; cat "$HERE/paritydoc/body.md"; } >"$tmp"
  if cmp -s "$tmp" "$doc"; then
    rm -f "$tmp"
    return 1
  fi
  diff -u --label "$(name "$tool") docs/parity.md (before)" --label "paritydoc/body.md" \
    <(tail -n +"$((end + 1))" "$doc") "$HERE/paritydoc/body.md"
  cat "$tmp" >"$doc"
  rm -f "$tmp"
  return 0
}

pin() {
  local version=$1 tool old changed=0
  [[ $version =~ $SEMVER ]] || die "pin: '$version' is not a release version such as v0.1.3"
  if ! is_tag "$version"; then
    for tool in "${TOOLS[@]}"; do
      [ -f "$tool/go.work" ] || die "pin: $version is not tagged yet and $(name "$tool") has no go.work," \
        "so it would look $version up in the module proxy and fail; add the go.work from docs/consuming.md first"
    done
  fi
  for tool in "${TOOLS[@]}"; do
    notice_end "$tool/docs/parity.md" >/dev/null ||
      die "pin: $tool/docs/parity.md has no support notice closed by a blank line, so its shared body cannot be found;" \
        "regenerate the page with paritydoc.Compose(notice.Markdown())"
  done
  local summary=()
  for tool in "${TOOLS[@]}"; do
    old=$(required "$tool")
    echo
    echo "=================== $(name "$tool") ==================="
    if ! (cd "$tool" && go mod edit -require="$MODULE@$version"); then
      echo "dev-all: pin: go mod edit failed in $(name "$tool"); nothing more was changed there" >&2
      exit 1
    fi
    if sync_parity "$tool"; then
      changed=1
      summary+=("$(name "$tool"): ${old:-none} -> $version; parity body changed")
    else
      echo "dev-all: parity body unchanged"
      summary+=("$(name "$tool"): ${old:-none} -> $version; parity body unchanged")
    fi
  done
  echo
  echo "=================== summary ==================="
  printf '%s\n' "${summary[@]}"
  if [ "$changed" -eq 1 ]; then
    echo "The parity rules shown above changed: both tools must honour them before go-deploy-utility is tagged."
  fi
  echo "go.sum is not touched: run dev-all.sh resolve once $version is tagged and pushed."
}

resolve() {
  local tool v failed=0 summary=()
  for tool in "${TOOLS[@]}"; do
    v=$(required "$tool")
    [ -n "$v" ] || die "resolve: $(name "$tool") does not require $MODULE"
    is_tag "$v" || die "resolve: $(name "$tool") requires $v, which is not tagged in go-deploy-utility;" \
      "tag and push $v first (docs/consuming.md)"
  done
  for tool in "${TOOLS[@]}"; do
    v=$(required "$tool")
    echo
    echo "=================== $(name "$tool") ==================="
    echo "dev-all: GOWORK=off go mod tidy ($MODULE $v)"
    if (cd "$tool" && GOWORK=off go mod tidy); then
      summary+=("OK    $(name "$tool"): $v resolved")
    else
      echo "dev-all: resolve: go mod tidy failed in $(name "$tool"); a tag pushed moments ago can take" \
        "a minute to reach the module proxy, so try again before looking further" >&2
      summary+=("FAIL  $(name "$tool"): $v")
      failed=1
    fi
  done
  echo
  echo "=================== summary ==================="
  printf '%s\n' "${summary[@]}"
  exit "$failed"
}

# notice: one line per tool on which go-deploy-utility code its tasks will build
# against. Read-only, and never fails the run.
notice() {
  local describe tool v
  describe=$(git -C "$HERE" describe --tags --always --dirty 2>/dev/null || echo unknown)
  echo "dev-all: go-deploy-utility checkout is at $describe"
  for tool in "${TOOLS[@]}"; do
    v=$(required "$tool")
    if [ -z "$v" ]; then
      echo "dev-all: $(name "$tool") does not require $MODULE"
    elif [ "${GOWORK:-}" = off ] || [ ! -f "$tool/go.work" ]; then
      echo "dev-all: $(name "$tool") builds $v from the module proxy; local go-deploy-utility changes are not tested"
    elif [ "$v" = "$describe" ]; then
      echo "dev-all: $(name "$tool") requires $v, which matches the checkout"
    elif ! is_tag "$v"; then
      echo "dev-all: $(name "$tool") requires $v (not tagged yet), built against the checkout at $describe"
    else
      echo "dev-all: warning: $(name "$tool") requires $v but the checkout is at $describe: it is tested with" \
        "go-deploy-utility code $v does not contain; pin the next version with dev-all.sh pin before tagging"
    fi
  done
}

[ $# -eq 0 ] && { usage; exit 2; }
case "$1" in -h|--help|help) usage; exit 0 ;; esac

[ -x "$HERE/scripts/dev.sh" ] || die "$HERE/scripts/dev.sh is missing or not executable"
[ -x "$SOLMQ_DIR/scripts/dev.sh" ] ||
  die "$SOLMQ_DIR has no executable scripts/dev.sh; clone solmq-conn-util there or set DEV_ALL_SOLMQ_DIR to its checkout"
[ -x "$SOLACE_UTIL_DIR/scripts/dev.sh" ] ||
  die "$SOLACE_UTIL_DIR has no executable scripts/dev.sh; clone solace-util there or set DEV_ALL_SOLACE_UTIL_DIR to its checkout"

case "$1" in
  pin)
    [ $# -eq 2 ] || die "pin takes exactly one version, as in: dev-all.sh pin v0.1.3"
    pin "$2"
    exit 0
    ;;
  resolve)
    [ $# -eq 1 ] || die "resolve stands alone; run the tasks in a separate call"
    resolve
    ;;
esac

declare -A KNOWN
for repo in "${REPOS[@]}"; do
  KNOWN[$repo]=" $(known "$repo/scripts/dev.sh" | sort -u | tr '\n' ' ') "
done

for word in "$@"; do
  case "$word" in
    pin) die "pin stands alone: dev-all.sh pin <version>" ;;
    resolve) die "resolve stands alone: dev-all.sh resolve" ;;
  esac
  hit=0
  for repo in "${REPOS[@]}"; do
    case "${KNOWN[$repo]}" in *" $word "*) hit=1 ;; esac
  done
  [ "$hit" -eq 1 ] || die "no repo's dev.sh knows '$word'"
done

notice

SUMMARY=()
FAILED=0
for repo in "${REPOS[@]}"; do
  args=()
  skipped=()
  for word in "$@"; do
    case "${KNOWN[$repo]}" in
      *" $word "*) args+=("$word") ;;
      *) skipped+=("$word") ;;
    esac
  done
  echo
  echo "=================== $(name "$repo") ==================="
  [ ${#skipped[@]} -gt 0 ] && echo "dev-all: not a task here, skipped: ${skipped[*]}"
  if [ ${#args[@]} -eq 0 ]; then
    SUMMARY+=("SKIP  $(name "$repo")")
    continue
  fi
  echo "dev-all: ./scripts/dev.sh ${args[*]}"
  if (cd "$repo" && ./scripts/dev.sh "${args[@]}"); then
    SUMMARY+=("OK    $(name "$repo"): ${args[*]}")
  else
    SUMMARY+=("FAIL  $(name "$repo"): ${args[*]} (see $(name "$repo")/scripts/logs/)")
    FAILED=1
  fi
done

echo
echo "=================== summary ==================="
printf '%s\n' "${SUMMARY[@]}"
exit "$FAILED"
