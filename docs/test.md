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

11 test files, 111 test functions. One is not a test: `TestHelperExitProcess` in
`exitcode` is an os/exec helper-process shim, a no-op unless
`GO_DEPLOY_UTILITY_TEST_CHILD_EXIT_CODE` is set. It exists because a real
`*exec.ExitError` cannot be built by hand.

| Package | Files | Tests |
| --- | --- | --- |
| root (`deployutility`) | 1 | 1 |
| exitcode | 1 | 11 |
| abbrev | 1 | 14 |
| notice | 1 | 8 |
| notice/noticetest | 1 | 14 |
| paritydoc | 1 | 9 |
| confirm | 3 | 20 |
| execguard | 1 | 24 |
| envref | 1 | 10 |
| **Total** | **11** | **111** |

The confirm tests always inject `Options.IsTerminal`. Under `go test` the real
stdin is `/dev/null`, which counts as a character device, so the default
detector would give a host-shaped answer.

## Tests by package

Source "new" means the test was written here. A lifted test names its origin
as repo and file: A is solmq-conn-util, B is solace-util.

### root, docs_test.go

| Test | What it covers | Source |
| --- | --- | --- |
| `TestEveryDocCarriesTheSupportNotice` | README.md and every `docs/*.md` open with this module's notice directly under the H1; `docs/handoff/` is not checked | B `internal/cli/commanddoc_test.go`, re-targeted |

### exitcode, exitcode_test.go

| Test | What it covers | Source |
| --- | --- | --- |
| `TestChildExitFallsBackWhenThereIsNoStatus` | A marked error with no exit status is a failure, and a usage marker beats a child status | B `internal/cli/exit_test.go` |
| `TestChildExitKeepsItsMessage` | The child-exit marker keeps the message and the cause | B `internal/cli/exit_test.go` |
| `TestExitCodeIsNeverNegative` | A killed child never yields a negative status | B `internal/cli/exit_test.go` |
| `TestUsageErrorKeepsItsMessage` | The usage marker keeps the message and the cause, marking twice is marking once, nil stays nil | B `internal/cli/exit_test.go` |
| `TestHelperExitProcess` | Not a test: the child process the two tests above start | B `internal/cli/exit_test.go` |
| `TestCodeClassifiesEveryBranch` | Every outcome of `Code`, including status passthrough and an unmarked status staying a failure | new |
| `TestChildStatusReportsOnlyMarkedRealStatus` | Each early return of `ChildStatus` | new |
| `TestUsagefWrapsWithPercentW` | `Usagef` with `%w` keeps the cause and reads like `fmt.Errorf` | new |
| `TestIsUsageAcrossWrapping` | The marker is found through `%w` and `errors.Join` | new |
| `TestExitValuesAreTheContract` | The values are 0, 1 and 2 | new |
| `TestAsUsageNeverDoubleWraps` | Marking a marked error returns the same value | new |

### abbrev, abbrev_test.go

| Test | What it covers | Source |
| --- | --- | --- |
| `TestExpandResolvesBothSpellings` | Canonical and short words resolve; anything else is unknown | B `internal/abbrev/abbrev_test.go` |
| `TestOrderIsDeclarationOrder` | Names, shorts and entries keep declaration order | B `internal/abbrev/abbrev_test.go` |
| `TestAccessorsReturnCopies` | Returned slices cannot change the set | B `internal/abbrev/abbrev_test.go` |
| `TestShortOfUnknownWord` | An undeclared word has no short forms | B `internal/abbrev/abbrev_test.go` |
| `TestList` | The error-sentence rendering | B `internal/abbrev/abbrev_test.go` |
| `TestKind` | The set's kind | B `internal/abbrev/abbrev_test.go` |
| `TestNewPanicsOnAWiringBug` | Every construction rule panics with a message naming it | B `internal/abbrev/abbrev_test.go` |
| `TestDigitsAreAllowed` | Digits are legal in a short form | B `internal/abbrev/abbrev_test.go` |
| `TestNewPanicsOnShadowedCanonicalDeclaredFirst` | The shadow rule holds in the other declaration order | new |
| `TestNewPanicsOnShortFormRepeatedWithinOneEntry` | A short form listed twice by one word is a duplicate claim | new |
| `TestNewPanicMessagesSayWhatToDo` | Every panic ends with the fix | new |
| `TestEntriesCloneKeepsNilShort` | A word with no short form keeps a nil slice | new |
| `TestSetIsIndependentOfInputSlice` | Changing New's input afterwards does not change the set | new |
| `TestToolAliasDeclarationsConstruct` | A's command, platform and target aliases build and resolve | new |

### notice, notice_test.go

| Test | What it covers | Source |
| --- | --- | --- |
| `TestNoticeMatchesCommittedBlocks` | The renderer reproduces the A, B and M notices byte for byte | new |
| `TestNoticeLineWidth` | Lines fit 80 columns, are quoted, and an over-long word gets its own line | new |
| `TestNewRejectsBadInput` | Every invalid name, kind and covers value is refused with the field and value named | new |
| `TestRenderedFormsAreASCII` | Every rendered form is printable ASCII | new, from A `support_test.go` |
| `TestPlainStripsMarkup` | The words-only form A compares its generator page against | new |
| `TestShortMatchesVersionForm` | The two-line form A's `version` prints | new |
| `TestLinesReturnsFreshSlice` | Returned slices are not shared | new |
| `TestZeroNoticeRendersNothing` | A Notice built without New renders nothing | new |

### notice/noticetest, noticetest_test.go

| Test | What it covers | Source |
| --- | --- | --- |
| `TestCheckAcceptsBothToolShapes` | B's strict placement and A's generated-doc placement | new |
| `TestCheckCRLF` | CRLF and mixed line endings pass | new, from A `normLF` |
| `TestCheckRejectsMissingNotice` | The error carries the block to paste | new |
| `TestCheckRejectsAlteredNotice` | A changed line is named with got and want; a truncated notice is reported | new |
| `TestCheckRequiresBlankLines` | A blank line is required before and after the notice | new, from A `assertNoticeAtTop` |
| `TestCheckNoticeBelowFirstSection` | The notice must come before the first section heading | new, from A `assertNoticeAtTop` |
| `TestCheckRequiresH1` | Line 1 must be the title | new, from B |
| `TestCheckZeroNotice` | An unbuilt notice is refused | new |
| `TestCheckDocsGlobExcludesHandoff` | The default glob does not descend into `docs/` subfolders | new |
| `TestCheckDocsNoDocs` | An empty glob is an error naming the pattern | new |
| `TestCheckDocsExplicitFilesAndExclude` | An explicit list, an exclusion, and a read failure that keeps its cause | new |
| `TestCheckDocsReportsEveryFailure` | All failing docs are reported in one run | new |
| `TestAssertHelpers` | The test helpers report each failure once and an empty tree fatally | new |
| `TestAssertASCIIRejectsControlAndNonASCII` | Tab, DEL and multi-byte text are reported with offsets | new, from A `support_test.go` |

### paritydoc, paritydoc_test.go

| Test | What it covers | Source |
| --- | --- | --- |
| `TestBodyShape` | The embedded body is plain ASCII, 157 lines, LF endings, one final newline | new |
| `TestBodyIsACopy` | Callers get a copy | new |
| `TestCheckAcceptsEitherToolsNotice` | Both tools' notices pass; a longer notice is the notice gate's concern | new |
| `TestCheckCRLF` | A CRLF page passes | new |
| `TestCheckTitle` | A changed or padded title fails with both shown | new |
| `TestCheckBoundaryStructure` | Each break in the frame around the body is named | new |
| `TestCheckReportsFirstDifference` | Body mismatches give doc and body line numbers, both texts and the fix | new |
| `TestCheckEmptyAndNil` | Empty input is reported without a panic | new |
| `TestComposeRoundTrip` | Composed pages pass; a bad notice block is refused | new |

### confirm, terminal_test.go, line_test.go, prompter_test.go

| Test | What it covers | Source |
| --- | --- | --- |
| `TestIsCharDeviceReportsAPipe` | A pipe and a nil file are not terminals | new |
| `TestStdinIsTerminalFalseForPipeFileAndClosed` | Pipe, regular file and closed file stdins are not terminals | new |
| `TestIsCharDeviceDevNull` | `/dev/null` counts as a character device | new |
| `TestPromptYesNo` | The lenient answer form and answered versus unanswered | B `internal/cli/cli_test.go` |
| `TestPromptYes` | The strict answer form | B `internal/cli/cli_test.go` |
| `TestLineReusesBufferedReader` | A buffered reader keeps later answers; end of input ends the prompt line | new |
| `TestPromptKeepsBufferedAnswersForLaterQuestions` | Two answers in one write serve two questions | B `internal/cli/cli_test.go`, over a Prompter |
| `TestPromptSeamsBeatStdin` | Seams are used and the real stdin is never read | B `internal/cli/cli_test.go`, over a Prompter |
| `TestPrompterDefaultsFromZeroOptions` | Empty options and a nil Prompter fall back to the defaults | new |
| `TestGateNoPromptSkipsEverything` | `--no-prompt` confirms without probing, asking or reading | new |
| `TestGateNonTerminalFailsNamingTheFlag` | A non-terminal fails with A's exact wording and nothing is read | new |
| `TestGateCustomNoPromptFlagAndDoing` | The tool's flag and action text appear in the error | new |
| `TestGateTerminalMatrix` | Each reply on a terminal, in both forms | new |
| `TestGateUnansweredOnTerminalErrors` | End of input at the question is an error naming the flag | new |
| `TestAskNonTerminalNeverWrites` | Ask and ReadLine refuse a non-terminal before writing or reading | new |
| `TestAskUnansweredIsNotAnError` | Ask reports end of input as unanswered | new |
| `TestReadLineReturnsTrimmedText` | A free-text reply comes back trimmed | new |
| `TestNotTerminalErrorMessage` | The cause text tools embed | new |
| `TestIsNotTerminalThroughWrapping` | The check sees through wrapping | new |
| `TestUnansweredWrapsEOF` | The end-of-input error wraps `io.EOF` | new |

### execguard, execguard_test.go

B's tests drove `*Config` accessors; here they drive `Check` and `Approve` with
rules fixtures that mirror B's four command fields. B's Config-bound guard tests
(validator and executor agreement, compose derivation, tilde expansion, schema
keys) stay in B.

| Test | What it covers | Source |
| --- | --- | --- |
| `TestCheckCommandAccepts` | Bare binaries, flag profiles, the compose subword, 8.3 names, `.exe`, approved chained runners | B `internal/config/execguard_test.go` |
| `TestCheckCommandRejects` | About 70 rejects, each with its message; the escalator rows now say "never permitted" | B `internal/config/execguard_test.go` |
| `TestEscalationFloorFoldsCaseBothWays` | The deny list folds case at both doors; the allowlist stays exact | B `internal/config/execguard_test.go` |
| `TestFlagValuePositionIsNotGuaranteed` | The accepted flag-value limit, and the charset still applying there | B `internal/config/execguard_test.go` |
| `TestCharsetAgreesAcrossBothYAMLForms` | The string form's split and the list form's refusal agree on Unicode spaces | B `internal/config/execguard_test.go`, without YAML |
| `TestGuardErrorsAreActionable` | Every message is one line, names the field and a remedy | B `internal/config/execguard_test.go` |
| `TestAllowCommandsAccepts` | Approvals record names with `.exe` folded | B `internal/config/execguard_test.go`, over `Approve` |
| `TestAllowCommandsRejects` | Bad approval values are refused naming the flag | B `internal/config/execguard_test.go`, over `Approve` |
| `TestAllowCommandsRejectsEscalation` | Escalators are never approved, and the message names `<escalator> <tool> ...` | B `internal/config/execguard_test.go`, over `Approve` |
| `TestEscalationCannotBeAllowedByAnyRoute` | Forced approvals and `Rules.Allowed` cannot carry an escalator | B `internal/config/execguard_test.go` |
| `TestAllowedBinaryIsNotGloballyAllowed` | An approval belongs to its own Approvals value | B `internal/config/execguard_test.go` |
| `TestCheckDeployCommandAcceptReject` | A's matrix over `strings.Fields`; `sudo podman` with approval is now refused | A `internal/validate/validate_deploycommand_test.go` |
| `TestCheckDeployCommandEndOfFlagsMarkerMidCommand` | `--` after an allowed binary is refused | A `internal/validate/validate_deploycommand_test.go` |
| `TestTokenCharsetRows` | A's safe-token rows in a flag-value position; `a~b` now passes | new, rows from A `TestSafeToken` |
| `TestCheckRejectsC1Controls` | C1 controls are refused with the code point; NEL stays whitespace | new |
| `TestCheckRejectsInvalidUTF8` | A bad byte is refused; a literal U+FFFD is not | new |
| `TestEscalatorRefusedInEveryBarePosition` | argv[0] and subcommand position say "never permitted"; a flag value named sudo passes | new |
| `TestEscalatorListCoversTheCategory` | All 22 names, upper case and `.exe`, through both doors | new |
| `TestExeSuffixFoldsAnyCase` | `.exe` folds in any case; the name does not | new |
| `TestSubwordOnlyLast` | The subword passes only in last place | new |
| `TestRulesWiringErrors` | An empty `Field` or `Allowed`, or a `Subword` that names an escalator or is not a plain word, is a wiring error | new |
| `TestApproveWiring` | Empty tool, no values, partial failure, repeats | new |
| `TestScopeOmitted` | The allowlist message with and without a scope | new |
| `TestExportedCharset` | `UnsafeChars`, `PathSeparators`, `IsControl`, `IsInvisible`, `ApproveFlag` | new |

### envref, envref_test.go

Neither tool tests its resolver on its own: B through `Load`, A through
`checkCred` callbacks. These tests are new, with rows from A `TestCheckCredRules`
and B `TestLoadSecretRefErrors`.

| Test | What it covers | Source |
| --- | --- | --- |
| `TestValidName` | The portable name shape | new |
| `TestCheckBothSet` | Both keys set is an error naming both | new |
| `TestCheckReferenceForm` | `${FOO}`, `$FOO` and an embedded reference get the bare-name message | new |
| `TestCheckInvalidName` | A bad name names which key takes the name and which the value | new |
| `TestRefWiringErrors` | A Ref without its key names is a wiring error | new |
| `TestResolveLiteralSkipsLookup` | A literal never consults the environment | new |
| `TestResolveFromEnv` | A set variable resolves; a literal `${X}` stays verbatim | new |
| `TestResolveUnsetAndEmpty` | Both are an `*UnsetError`, told apart | new |
| `TestResolveNoEnvironmentAccess` | A nil lookup is its own error, not an unset variable | new |
| `TestMessagesNeverCarryTheValue` | No error path prints the literal, a resolved value, or a name that fails the name rules | new |

## Coverage

The last recorded total is 99.1%, from B1 (2026-11-02; the scaffold was 0.0%). The
`cov` task prints the total into `scripts/logs/cov.log`, and the previous total
there is the floor for the next change.
