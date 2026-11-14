#requires -Version 5.1
# dev-all.ps1 -- run the same dev.ps1 tasks in all three repos of this workspace:
# go-deploy-utility first (it lands first), then solmq-conn-util, then solace-util;
# and move both tools' pin on this module. Committed, but outside the dev-script
# contract: CI never calls it, since a runner checks out only this repo.
# The Windows mirror of dev-all.sh, with the same words and behaviour.
#
#   .\scripts\dev-all.ps1 all scan
#   .\scripts\dev-all.ps1 regen all scan cov
#   $env:GOWORK='off'; .\scripts\dev-all.ps1 all scan   # the tools against their published pin
#   .\scripts\dev-all.ps1 pin v0.1.3                    # before the tag: require it, re-sync parity.md
#   .\scripts\dev-all.ps1 resolve                       # after the tag: GOWORK=off go mod tidy
#
# Task words: each repo receives only the words its own dev.ps1 knows (its Task-*
# functions and the aggregates in its argument switch, such as all, full or
# binaries), in the order given; a repo that knows none of them is skipped. A word
# no repo knows is refused before anything runs. Every repo runs even when an
# earlier one fails, each in its own PowerShell process so no repo's environment
# reaches the next, and the exit status is 1 if any failed. Environment variables
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
[CmdletBinding()]
param([Parameter(ValueFromRemainingArguments = $true)][string[]]$Words)

$ErrorActionPreference = 'Continue'
$Here   = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
$Module = 'github.com/hafio/go-deploy-utility'
$Semver = '^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$'
$SolmqDir      = if ($env:DEV_ALL_SOLMQ_DIR) { $env:DEV_ALL_SOLMQ_DIR } else { Join-Path $Here '..\solace-ibmmq-connector-helper' }
$SolaceUtilDir = if ($env:DEV_ALL_SOLACE_UTIL_DIR) { $env:DEV_ALL_SOLACE_UTIL_DIR } else { Join-Path $Here '..\solace-cnt-scripts' }
$Repos = @($Here, $SolmqDir, $SolaceUtilDir)
$Tools = @($SolmqDir, $SolaceUtilDir)
$Utf8NoBom = New-Object System.Text.UTF8Encoding($false)

function Show-Usage {
  foreach ($line in (Get-Content -LiteralPath $PSCommandPath | Select-Object -Skip 1)) {
    if ($line -notmatch '^#') { break }
    Write-Host ($line -replace '^# ?', '')
  }
}

function Die { param([string]$Msg) [Console]::Error.WriteLine("dev-all: $Msg"); exit 2 }

# Get-Name <repo>: the repo's directory name, as the summary shows it.
function Get-Name { param([string]$Repo) Split-Path -Leaf (Resolve-Path -LiteralPath $Repo).Path }

# Get-Known <script>: every word the script accepts as a task or an aggregate.
function Get-Known {
  param([string]$Script)
  $text = Get-Content -Raw -LiteralPath $Script
  $words = @()
  $words += [regex]::Matches($text, '(?m)^function Task-([a-z_]+)') | ForEach-Object { $_.Groups[1].Value }
  $words += [regex]::Matches($text, '''([a-z_]+)''\s*\{\s*\$queue') | ForEach-Object { $_.Groups[1].Value }
  return ($words | Sort-Object -Unique)
}

# Get-Required <tool>: the version of this module the tool's go.mod requires, in
# either require form; empty when it requires none.
function Get-Required {
  param([string]$Tool)
  $m = Select-String -LiteralPath (Join-Path $Tool 'go.mod') -Pattern '^(require\s+)?\s*github\.com/hafio/go-deploy-utility\s+(v\S+)' |
    Select-Object -First 1
  if ($m) { return $m.Matches[0].Groups[2].Value }
  return ''
}

# Test-Tag <version>: the version is a tag in this checkout.
function Test-Tag {
  param([string]$Version)
  git -C $Here rev-parse -q --verify "refs/tags/$Version" 2>$null | Out-Null
  return ($LASTEXITCODE -eq 0)
}

# Get-NoticeEnd <lines>: the index of the blank line that closes the support
# notice, after which the shared body starts; -1 when there is no such line.
function Get-NoticeEnd {
  param([string[]]$Lines)
  $inq = $false
  for ($i = 0; $i -lt $Lines.Count; $i++) {
    $l = $Lines[$i] -replace "`r$", ''
    if ($l.StartsWith('>')) { $inq = $true; continue }
    if ($inq) { if ($l -eq '') { return $i } else { return -1 } }
  }
  return -1
}

# Sync-Parity <tool>: replace everything below the tool's own title and notice in
# docs/parity.md with paritydoc/body.md and show the change. Returns $true when the
# body changed, $false when it was already current.
function Sync-Parity {
  param([string]$Tool)
  $doc  = Join-Path $Tool 'docs\parity.md'
  $old  = [IO.File]::ReadAllText($doc)
  $body = [IO.File]::ReadAllText((Join-Path $Here 'paritydoc\body.md'))
  $lines = $old -split "`n"
  $end = Get-NoticeEnd $lines
  $new = (($lines[0..$end]) -join "`n") + "`n" + $body
  if ($new -ceq $old) { return $false }
  $before = [IO.Path]::GetTempFileName()
  [IO.File]::WriteAllText($before, ((($lines | Select-Object -Skip ($end + 1))) -join "`n"), $Utf8NoBom)
  git --no-pager diff --no-index --no-color -- $before (Join-Path $Here 'paritydoc\body.md') | Write-Host
  Remove-Item -LiteralPath $before
  [IO.File]::WriteAllText($doc, $new, $Utf8NoBom)
  return $true
}

function Invoke-Pin {
  param([string]$Version)
  if ($Version -cnotmatch $Semver) { Die "pin: '$Version' is not a release version such as v0.1.3" }
  if (-not (Test-Tag $Version)) {
    foreach ($tool in $Tools) {
      if (-not (Test-Path -LiteralPath (Join-Path $tool 'go.work'))) {
        Die ("pin: $Version is not tagged yet and $(Get-Name $tool) has no go.work, so it would look $Version up " +
          "in the module proxy and fail; add the go.work from docs/consuming.md first")
      }
    }
  }
  foreach ($tool in $Tools) {
    $lines = [IO.File]::ReadAllText((Join-Path $tool 'docs\parity.md')) -split "`n"
    if ((Get-NoticeEnd $lines) -lt 0) {
      Die ("pin: $tool\docs\parity.md has no support notice closed by a blank line, so its shared body cannot " +
        "be found; regenerate the page with paritydoc.Compose(notice.Markdown())")
    }
  }
  $changed = $false
  $summary = @()
  foreach ($tool in $Tools) {
    $name = Get-Name $tool
    $old = Get-Required $tool
    if (-not $old) { $old = 'none' }
    Write-Host ''
    Write-Host "=================== $name ==================="
    Push-Location -LiteralPath $tool
    go mod edit "-require=$Module@$Version"
    $code = $LASTEXITCODE
    Pop-Location
    if ($code -ne 0) {
      [Console]::Error.WriteLine("dev-all: pin: go mod edit failed in $name; nothing more was changed there")
      exit 1
    }
    if (Sync-Parity $tool) {
      $changed = $true
      $summary += "${name}: $old -> $Version; parity body changed"
    } else {
      Write-Host 'dev-all: parity body unchanged'
      $summary += "${name}: $old -> $Version; parity body unchanged"
    }
  }
  Write-Host ''
  Write-Host '=================== summary ==================='
  $summary | ForEach-Object { Write-Host $_ }
  if ($changed) {
    Write-Host 'The parity rules shown above changed: both tools must honour them before go-deploy-utility is tagged.'
  }
  Write-Host "go.sum is not touched: run dev-all.ps1 resolve once $Version is tagged and pushed."
  exit 0
}

function Invoke-Resolve {
  foreach ($tool in $Tools) {
    $v = Get-Required $tool
    if (-not $v) { Die "resolve: $(Get-Name $tool) does not require $Module" }
    if (-not (Test-Tag $v)) {
      Die "resolve: $(Get-Name $tool) requires $v, which is not tagged in go-deploy-utility; tag and push $v first (docs/consuming.md)"
    }
  }
  $failed = 0
  $summary = @()
  $saved = $env:GOWORK
  foreach ($tool in $Tools) {
    $name = Get-Name $tool
    $v = Get-Required $tool
    Write-Host ''
    Write-Host "=================== $name ==================="
    Write-Host "dev-all: GOWORK=off go mod tidy ($Module $v)"
    Push-Location -LiteralPath $tool
    $env:GOWORK = 'off'
    go mod tidy
    $code = $LASTEXITCODE
    $env:GOWORK = $saved
    Pop-Location
    if ($code -eq 0) {
      $summary += "OK    ${name}: $v resolved"
    } else {
      [Console]::Error.WriteLine("dev-all: resolve: go mod tidy failed in $name; a tag pushed moments ago can take " +
        "a minute to reach the module proxy, so try again before looking further")
      $summary += "FAIL  ${name}: $v"
      $failed = 1
    }
  }
  Write-Host ''
  Write-Host '=================== summary ==================='
  $summary | ForEach-Object { Write-Host $_ }
  exit $failed
}

# Show-Notice: one line per tool on which go-deploy-utility code its tasks will
# build against. Read-only, and never fails the run.
function Show-Notice {
  $describe = git -C $Here describe --tags --always --dirty 2>$null
  if ($LASTEXITCODE -ne 0 -or -not $describe) { $describe = 'unknown' }
  Write-Host "dev-all: go-deploy-utility checkout is at $describe"
  foreach ($tool in $Tools) {
    $name = Get-Name $tool
    $v = Get-Required $tool
    if (-not $v) {
      Write-Host "dev-all: $name does not require $Module"
    } elseif ($env:GOWORK -eq 'off' -or -not (Test-Path -LiteralPath (Join-Path $tool 'go.work'))) {
      Write-Host "dev-all: $name builds $v from the module proxy; local go-deploy-utility changes are not tested"
    } elseif ($v -ceq $describe) {
      Write-Host "dev-all: $name requires $v, which matches the checkout"
    } elseif (-not (Test-Tag $v)) {
      Write-Host "dev-all: $name requires $v (not tagged yet), built against the checkout at $describe"
    } else {
      Write-Host ("dev-all: warning: $name requires $v but the checkout is at ${describe}: it is tested with " +
        "go-deploy-utility code $v does not contain; pin the next version with dev-all.ps1 pin before tagging")
    }
  }
}

if (-not $Words -or $Words.Count -eq 0) { Show-Usage; exit 2 }
if ($Words[0] -in @('-h', '--help', 'help')) { Show-Usage; exit 0 }

if (-not (Test-Path -LiteralPath (Join-Path $Here 'scripts\dev.ps1'))) { Die "$Here\scripts\dev.ps1 is missing" }
if (-not (Test-Path -LiteralPath (Join-Path $SolmqDir 'scripts\dev.ps1'))) {
  Die "$SolmqDir has no scripts\dev.ps1; clone solmq-conn-util there or set DEV_ALL_SOLMQ_DIR to its checkout"
}
if (-not (Test-Path -LiteralPath (Join-Path $SolaceUtilDir 'scripts\dev.ps1'))) {
  Die "$SolaceUtilDir has no scripts\dev.ps1; clone solace-util there or set DEV_ALL_SOLACE_UTIL_DIR to its checkout"
}

switch -CaseSensitive ($Words[0]) {
  'pin' {
    if ($Words.Count -ne 2) { Die 'pin takes exactly one version, as in: dev-all.ps1 pin v0.1.3' }
    Invoke-Pin $Words[1]
  }
  'resolve' {
    if ($Words.Count -ne 1) { Die 'resolve stands alone; run the tasks in a separate call' }
    Invoke-Resolve
  }
}

$Known = @{}
foreach ($repo in $Repos) { $Known[$repo] = @(Get-Known (Join-Path $repo 'scripts\dev.ps1')) }

foreach ($word in $Words) {
  if ($word -ceq 'pin') { Die 'pin stands alone: dev-all.ps1 pin <version>' }
  if ($word -ceq 'resolve') { Die 'resolve stands alone: dev-all.ps1 resolve' }
  $hit = $false
  foreach ($repo in $Repos) { if ($Known[$repo] -ccontains $word) { $hit = $true } }
  if (-not $hit) { Die "no repo's dev.ps1 knows '$word'" }
}

Show-Notice

# Each repo runs in a fresh process of this same PowerShell, so a variable one
# dev.ps1 sets (GOTOOLCHAIN from its own go.mod, say) cannot reach the next repo.
$Shell = (Get-Process -Id $PID).Path
$summary = @()
$failed = 0
foreach ($repo in $Repos) {
  $name = Get-Name $repo
  $taskArgs = @($Words | Where-Object { $Known[$repo] -ccontains $_ })
  $skipped  = @($Words | Where-Object { $Known[$repo] -cnotcontains $_ })
  Write-Host ''
  Write-Host "=================== $name ==================="
  if ($skipped.Count -gt 0) { Write-Host "dev-all: not a task here, skipped: $($skipped -join ' ')" }
  if ($taskArgs.Count -eq 0) { $summary += "SKIP  $name"; continue }
  Write-Host "dev-all: .\scripts\dev.ps1 $($taskArgs -join ' ')"
  & $Shell -NoProfile -File (Join-Path $repo 'scripts\dev.ps1') @taskArgs
  if ($LASTEXITCODE -eq 0) {
    $summary += "OK    ${name}: $($taskArgs -join ' ')"
  } else {
    $summary += "FAIL  ${name}: $($taskArgs -join ' ') (see $name\scripts\logs\)"
    $failed = 1
  }
}

Write-Host ''
Write-Host '=================== summary ==================='
$summary | ForEach-Object { Write-Host $_ }
exit $failed
