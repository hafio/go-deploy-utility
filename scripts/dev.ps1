#requires -Version 5.1
# Dev tasks for go-deploy-utility. Behaviourally identical to dev.sh -- same
# task names, same gating, same footer format.
#
# go-deploy-utility is a Go library (init-devops library shape): it ships no
# image and no binaries, so build is a compile check that writes nothing, and
# the Docker tasks (image, up, down) do not apply and are omitted.
[CmdletBinding()]
param([Parameter(ValueFromRemainingArguments = $true)][string[]]$Tasks)

$ErrorActionPreference = 'Continue'
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$RepoRoot  = Split-Path -Parent $ScriptDir
$LogDir    = Join-Path $ScriptDir 'logs'
New-Item -ItemType Directory -Force -Path $LogDir | Out-Null
Set-Location $RepoRoot
$env:NO_COLOR = '1'

# Toolchain parity with CI: when go.mod carries a `toolchain` directive it is
# the single pin, and GOTOOLCHAIN makes any go binary honour it exactly. Set
# only when unset, so an exported value wins.
$goMod = Join-Path $RepoRoot 'go.mod'
if (-not $env:GOTOOLCHAIN -and (Test-Path $goMod)) {
  $m = Select-String -Path $goMod -Pattern '^toolchain (\S+)' | Select-Object -First 1
  if ($m) { $env:GOTOOLCHAIN = $m.Matches[0].Groups[1].Value }
}

# --- output helpers ---------------------------------------------------------
function Step { param($m) Write-Host "==> $m" -ForegroundColor Cyan }
function Ok   { param($m) Write-Host "ok: $m"    -ForegroundColor Green }
function Warn { param($m) Write-Host "warn: $m"  -ForegroundColor Yellow }
function Die  { param($m) Write-Host "error: $m" -ForegroundColor Red; exit 1 }

# Offset with no colon, matching dev.sh's `date +%z` (+0800, not +08:00) so both
# scripts write byte-identical footer shapes.
function Get-Now {
  $d = Get-Date
  '{0}{1}' -f $d.ToString('yyyy-MM-ddTHH:mm:ss'), ($d.ToString('zzz') -replace ':', '')
}
function Get-Log { param($Task) Join-Path $LogDir "$Task.log" }

function Start-TaskLog {
  param($Task)
  Set-Content -Path (Get-Log $Task) -Encoding utf8 `
    -Value ("=== {0} | {1} ===" -f (Get-Now), $Task)
}

function Write-Finish {
  param([string]$Task, [int]$Code, [int]$Seconds)
  $status = if ($Code -eq 0) { 'OK' } else { "FAILED (exit $Code)" }
  $line = '{0} | {1} | {2}s | {3}' -f (Get-Now), $Task, $Seconds, $status
  # Add-Content, never Tee-Object: Tee doubles lines and writes UTF-16.
  Add-Content -Path (Get-Log $Task) -Value $line -Encoding utf8
  Write-Host $line
}

# Capture once, write once. "$_" flattens stderr ErrorRecords; -Width stops
# column wrap; the CSI strip keeps the file readable plain text.
function Invoke-Logged {
  # NB: $CmdArgs, not $Args -- $Args is a PowerShell automatic variable, so a
  # param named $Args binds empty and `& $Exe @Args` runs $Exe with no args.
  param([string]$Task, [string]$Exe, [string[]]$CmdArgs)
  $out = (& $Exe @CmdArgs 2>&1 | ForEach-Object { "$_" } | Out-String -Width 4096)
  $code = $LASTEXITCODE
  $out = $out -replace "\x1b\[[0-9;?]*[a-zA-Z]", ""
  Add-Content -Path (Get-Log $Task) -Value $out -Encoding utf8
  Write-Host $out
  return $code
}

# --- tasks ------------------------------------------------------------------
# PowerShell returns EVERY uncaptured pipeline value, not just `return`: a bare
# external command inside a task turns $code into an array. Route commands
# through Invoke-Logged, or pipe anything you don't return to Out-Null.

# A library has nothing to ship, so build compiles every package and writes
# nothing: no dist, no version stamp, and TARGET_OS/TARGET_ARCH are ignored.
function Task-build { return (Invoke-Logged 'build' 'go' @('build','./...')) }

# Static analysis, plus a formatting CHECK. `gofmt -l` only LISTS unformatted
# files -- it never writes -- so drift is REPORTED here, never silently fixed:
# a gate that repairs what it is measuring proves nothing. vet and the format
# check both always run and both report, so one pass shows every problem
# instead of hiding vet errors behind a formatting failure.
function Task-vet {
  $code = Invoke-Logged 'vet' 'go' @('vet','./...')
  if (-not (Get-Command gofmt -ErrorAction SilentlyContinue)) {
    $msg = 'gofmt not found on PATH: it ships with the Go toolchain, so this is a broken install'
    Add-Content -Path (Get-Log 'vet') -Value $msg -Encoding utf8
    Write-Host $msg
    return 1
  }
  $unformatted = @(& gofmt -l . | Where-Object { "$_" -ne '' })
  if ($unformatted.Count -gt 0) {
    $msg = "gofmt: these files are not formatted:`n" + ($unformatted -join "`n") +
           "`nfix with: gofmt -w " + ($unformatted -join ' ')
    Add-Content -Path (Get-Log 'vet') -Value $msg -Encoding utf8
    Write-Host $msg
    if ($code -eq 0) { $code = 1 }
  }
  return $code
}

# The race detector is opt-in in both scripts (DEV_RACE=1). It needs cgo, which
# a stock Windows host lacks, and a default that differed between dev.sh and
# dev.ps1 would break the rule that the two behave identically.
function Task-test {
  $testArgs = @('test')
  if ($env:DEV_RACE -eq '1') { $testArgs += '-race' }
  $testArgs += @('-count=1','./...')
  return (Invoke-Logged 'test' 'go' $testArgs)
}

function Task-cov {
  # Coverage profile -> HTML, and PRINT the total so the footer captures it.
  # The previous total in logs/cov.log is the floor (local only -- CI is a
  # fresh checkout with no prior log). -coverpkg credits cross-package
  # execution, so a package exercised through another one's tests still counts.
  $code = Invoke-Logged 'cov' 'go' @('test','-coverpkg=./...','-coverprofile=coverage.out','-count=1','./...')
  if ($code -ne 0) { return $code }
  go tool cover "-html=coverage.out" -o coverage.html | Out-Null
  $total = (go tool cover "-func=coverage.out" | Select-Object -Last 1)
  Add-Content -Path (Get-Log 'cov') -Value $total -Encoding utf8
  Write-Host $total
  return 0
}

# One task, every applicable check. A library ships no image, so this is the Go
# scan alone. govulncheck reports only vulns reachable from called code, and
# any finding fails the run, because it cannot tell fixable from unfixable. That
# is a known gap against the contract's "unfixable warns" rule, accepted while
# the module depends on the standard library alone, where a toolchain bump fixes.
# NOT `go run ...@version`: that form ignores this go.mod and its toolchain
# pin. `go tool` builds the scanner in this module, at the version go.mod pins.
function Task-scan {
  return (Invoke-Logged 'scan' 'go' @('tool','govulncheck','./...'))
}

# Local only: the graph is a developer artifact, not a CI output.
function Task-graphify {
  if ($env:CI) { Warn 'graphify is local-only; skipping in CI'; return 0 }
  if (-not (Get-Command graphify -ErrorAction SilentlyContinue)) {
    Warn 'graphify not on PATH; skipping'; return 0
  }
  return (Invoke-Logged 'graphify' 'graphify' @('update','.'))
}

# --- dispatch ---------------------------------------------------------------
$All  = @('build','vet','test')
$Full = @('build','vet','test','cov','scan','graphify')

function Show-Usage {
  @"
usage: dev.ps1 <task>...

  build vet test cov scan graphify
  all   = $($All -join ' ')            (what CI runs, as: all scan)
  full  = $($Full -join ' ')   (pre-tag sweep)

  DEV_RACE=1 adds -race to test (needs cgo).
"@ | Write-Host
}

if (-not $Tasks -or $Tasks[0] -in @('-h','--help','help')) { Show-Usage; exit 0 }

$queue = @()
foreach ($t in $Tasks) {
  switch ($t) { 'all' { $queue += $All } 'full' { $queue += $Full } default { $queue += $t } }
}

$failed = 0
foreach ($task in $queue) {
  if (-not (Get-Command "Task-$task" -ErrorAction SilentlyContinue)) { Die "unknown task: $task" }
  Step $task
  Start-TaskLog $task
  $sw = [Diagnostics.Stopwatch]::StartNew()
  $code = 0
  try { $code = & "Task-$task" } catch { $code = 1 }
  if ($null -eq $code) { $code = 0 }
  $sw.Stop()
  Write-Finish -Task $task -Code $code -Seconds ([int]$sw.Elapsed.TotalSeconds)
  if ($code -ne 0) { $failed = 1; Warn "$task failed; stopping"; break }
  Ok $task
}
exit $failed
