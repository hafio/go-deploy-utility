# One catalogue (consolidation draft, 2026-11-17)

Status 2026-11-18: every N entry marked "open N<n>" below is now decided as recommended.
E24 (2026-11-18) replaces the free-form command tokens with structured keys (`cli`,
`context`, `contextFile`, `compose`), so each constructor takes those fields, not tokens,
and `--allow-command` is gone. Section 11 of the 2026-11-15 design note holds the answers.

Draft for the owner. Nothing is built. Section 6 item 2 of docs/review/2026-11-15-surface-review.md.
M = go-deploy-utility, A = solmq-conn-util, B = solace-util. Paths under A or B are relative to
that repository. Sources: design note sections 3, 10, 11; review findings 1 and 2; questions part N.

## 1. Purpose

This is the one list of what each M client offers: per method its signature, the grammar of each
argument, its typed errors, the answer it returns and the builder or parser that goes private. It
applies the rulings of design section 10 (1 to 7, 11, 13) exactly, and takes the Recommended
choice of every part N entry the owner has not answered yet, marked "open N<n>". Where a ruling
and the code as it stands disagree, section 7 says so instead of choosing silently.

## 2. Per client (conventions, then each client)

Record shape (cells are split into lines because of the 100-column limit):

    sig      Go signature sketch
    cmd      command per engine: docker | podman | kubectl or oc | systemctl
    args     grammar per argument (code from the list below)
    errors   typed errors the method can return (codes below)
    answer   what it hands back, and which streams reach the writers
    private  builder or parser that becomes private (or is removed)
    A, B     caller today (file:line); "-" means no caller

Result and errors (ruling 4, defined once in `runner`):

    runner.Result  {Argv []string; ExitCode int; Duration time.Duration}
                   Argv = command as run, stdin as a byte count, env values masked (N7).
                   Describes the LAST command a method ran (open N5). Returned on failure too.
                   ExitCode never negative once a command ran; -1 = never started (open N3).
    runner.Error   {Op string; Kind error; Stderr string; ExitCode int; Err error}
                   Stderr = exact bytes (open N28); Error() text uses a cleaned copy and does not
                   repeat stderr (the writer already got it); Unwrap gives Err, so
                   errors.Is(err, context.Canceled) holds; ExitCode() lets exitcode.ChildStatus
                   read it (open N4). No error text embeds captured stdout.

Kind codes (the seven of ruling 4, plus two the method tables need):

    REF   refused (kubectl Forbidden on stderr, or can-i "no")    NF    not found
    UNA   unanswered (no verdict, or nothing on stdout)           TO    timed out
    UNS   unsupported                                             NOC   no context
    PART  partial listing                                         FAIL  ran, non-zero, no kind
    START never started (guard, resolve, argument check)          ("FAIL" and "START" are not in
                                                                   ruling 4: see open X3)

Stream class in `answer`:

    ANS  returns stdout (decoded or raw), stdout never written; stderr to the stderr writer
    ACT  returns nothing; stdout and stderr both to the writers
    PRB  probe: stdout dropped (today both tools drop it); stderr to the writer (open X2)
    STR  streamed live, unbuffered, no copy kept (logs, attach)
    SEC  ANS where stdout is secret: never to a writer, the Result or an error

Grammar codes (the gdu 4 table is still to be written; "gdu 4 grammar" marks a gap):

    G-ctr   containercli.ValidName (engine grammar, [A-Za-z0-9][A-Za-z0-9_.-]*)  (open N30)
    G-name  containercli.CheckName: DNS label, <= 63, for names the tool creates (open N30)
    G-lbl   kube.ValidDNSLabel (renamed from ValidLabel, open N59)
    G-sub   kube.ValidDNSSubdomain (new, gdu 4; pod, Secret, StorageClass names)
    G-free  free operand: non-empty, no leading '-', no control character (open N11, N12)
    G-env   envref.ValidName
    G-dur   time.Duration > 0, written as whole seconds; zero = the default constant
    G-host  CheckHostPath (open N14)
    G-ctx   gdu 11 grammar: any string except empty, leading '-', control characters (':' and '/'
            allowed); written as --context=<name>

Every client method takes `ctx` first (Attach: see its entry), is safe for concurrent use, and
returns `(answer, runner.Result, error)` or `(runner.Result, error)`.

### 2.1 Set-up, writers, Run and Output (rulings 1 to 3)

    package runner:
      type Setup struct {
          Runner    runner.Runner
          Approvals execguard.Approvals
          Stdout    io.Writer   // nil = discard; a real *os.File goes direct, a buffer is piped
          Stderr    io.Writer
      }

    each client:
      New(s runner.Setup, ...client inputs) (*Client, error)   // runs execguard.Check ONCE
      (c *Client) With(stdout, stderr io.Writer) *Client       // copy for one call (ruling 1)
      (c *Client) Run(ctx, args []string, o RunOptions) (runner.Result, error)     // to writers
      (c *Client) Output(ctx, args []string, o RunOptions) ([]byte, runner.Result, error)

    RunOptions {Stdin []byte}        // no Env: only ComposeUp, ComposeDown and the session env
                                     // set Cmd.Env (open N8)

- The guard error from `New` is returned unwrapped and unmarked, so A can keep its exit-2 mark
  (open N1). Tokens cannot be changed after `New`; a client value is a copy.
- Run and Output put `args` after the client's base tokens, pinned context and session
  environment. args: only `runner.Cmd.Check` plus no control character; the module decides
  nothing about the command (ruling 2; open X13).
- `With` shares the client's cached state (podman version, pinned context) through a pointer.

### 2.1.1 Method counts

Method counts (operations, not counting `New` and `With`): containercli 21, podman 9, systemd 12,
kube 26, rbac 1, nsgate 1. Total 70.

### 2.2 containercli.Client (docker, podman, nerdctl engines and compose)

    New(s runner.Setup, c containercli.Config) (*Client, error)
      Config {Engine Engine; Rules execguard.Rules; Command []string;
              ComposeRules execguard.Rules; Compose []string; Rootless bool; Env []string}
    Checks at set-up: execguard.Check on Command (and on Compose, whose Rules.Subword is "compose"
    for `docker compose`: A must declare it as B does); Engine is Docker or Podman, else error;
    Env entries are KEY=VALUE (the session env of open N38, set on every podman command).
    Rootless feeds EngineHint. Programs the client issues: docker, nerdctl, docker-compose for
    compose, podman (open N9; nerdctl and Engine: open X14).

#### Info
    sig      Info(ctx) (runner.Result, error)
    cmd      <engine> info
    args     none
    errors   UNA (message carries EngineHint(engine, rootless)); START
    answer   PRB
    private  Info, InfoError (its `out` argument goes), EngineHint
    A        internal/runner/runner.go:391
    B        internal/container/preflight.go:33

#### List
    sig      List(ctx, o ListOptions{RunningOnly bool; Label string}) ([]Row, Result, error)
    cmd      <engine> ps --all --no-trunc [--filter status=running] [--filter label=<l>]
             --format '{{.Names}}\t{{.Image}}\t{{.Status}}'
    args     Label: key=value, each part G-free (open N31); no name filter (names matched in Go)
    errors   UNA (listing could not run); REF; START
    answer   ANS decoded as []Row; HasName, ByImage, ValidName stay pure; SafeNames added (N34)
    private  List, ParseList
    A        internal/runner/runner.go:974 (HasName at :210)
    B        internal/container/manager.go:988, 1222

#### Inspect
    sig      Inspect(ctx, names ...string) ([]inspect.Container, Result, error)
    cmd      <engine> container inspect <names...>
    args     names: each G-ctr; none given = START error, nothing runs
    errors   NF (one name asked and missing); PART (several asked; found ones returned, error
             names the missing ones) (open N29); UNA (undecodable); REF
    answer   ANS decoded; the raw answer holds Config.Env, so B's raw print uses Output
    private  inspect.Containers; inspect.Decode stays (open N26); Container.Secrets gains the ID
    A        internal/runner/runner.go:916 (Decode at :217); internal/statusreport/parse.go:553
    B        internal/container/inspect.go:66, 71; manager.go:931, 1380

#### InspectImage (new, ruling 6)
    sig      InspectImage(ctx, ref string) (inspect.Image, Result, error)
    cmd      <engine> image inspect <ref>
    args     ref: gdu 4 grammar (image reference); G-free until the table is written
    errors   NF (engine does not have the image: A then pulls and asks again); UNA (undecodable)
    answer   ANS decoded: RepoDigests, Config.Entrypoint, Config.Cmd. One method replaces A's
             --format template (runner.go:335) and its plain inspect (runner.go:934)
    private  new builder and decoder beside inspect.Decode
    A        internal/runner/runner.go:335, 934
    B        -

#### Stats (new, ruling 6)
    sig      Stats(ctx, names ...string) ([]StatRow, Result, error)
    cmd      <engine> stats --no-stream --format <name, CPU %, memory use, memory %; template>
    args     names: each G-ctr
    errors   FAIL (A turns any failure into a note)
    answer   ANS decoded: name, CPU percent, memory use, memory percent
    private  new builder and parser, lifted from A's statusreport
    A        internal/runner/runner.go:942-963
    B        -

#### ComposeUp
    sig      ComposeUp(ctx, o ComposeOptions{File []byte; Project, Dir string; Env []string;
             Recreate bool}) (Result, error)
    cmd      <compose> -f - -p <project> --project-directory <dir> up -d [--force-recreate]
    args     Project: G-name (a DNS label is a valid project name); Dir: G-host; File: non-empty,
             stdin; Env: KEY=VALUE, name G-env, value secret (masked in Result)
    errors   FAIL; START
    answer   ACT (pull progress; piped when a writer is a buffer: smoke row, open N6)
    private  ComposeUp; ComposeDigest stays (pure; its second input is the child env)
    A        internal/runner/runner.go:183 (digest cmd/solmq-conn-util/main.go:782)
    B        internal/container/manager.go:917 (digest :891)

#### ComposeDown
    sig      ComposeDown(ctx, o ComposeOptions) (Result, error)   // Recreate ignored
    cmd      <compose> -f - -p <project> --project-directory <dir> down
    args     as ComposeUp. Env carries the real credentials in both tools, so masked the same
    errors   FAIL; START. Exits 0 for an unknown project: "gone" is the tool's own List check
    answer   ACT
    private  ComposeDown
    A        internal/runner/runner.go:185
    B        internal/container/manager.go:1135

#### ComposeVersion (new, ruling 6)
    sig      ComposeVersion(ctx) (Result, error)
    cmd      <compose> version
    args     none
    errors   UNS (compose plugin missing; message carries B's hint) (open X8); START
    answer   PRB
    private  new builder
    A        - (A meets a missing plugin only at `up`)
    B        internal/container/manager.go:333-340

#### Start, Stop, Restart
    sig      Start(ctx, name string) (Result, error)   // Stop, Restart alike
    cmd      <engine> start|stop|restart <name>     (docker only; on podman both tools use systemd)
    args     name: G-ctr
    errors   FAIL with stderr (engines word "not found" differently: no NF, open N29); START
    answer   ACT
    private  Start, Stop, Restart; Remove is removed (CR-0001)
    A        internal/runner/runner.go:292, 294, 296
    B        internal/container/manager.go:1178, 1190, 1202

#### Logs
    sig      Logs(ctx, name string, o LogOptions{Follow, Timestamps, Previous bool; Tail int;
             Since time.Duration}) (Result, error)
    cmd      <engine> logs [-f] [--timestamps] [--tail <n>] [--since <t>] <name>
    args     name: G-ctr; Tail a count (an unset value needs a name; A's 100000 cap stays A's);
             Since > 0, written as seconds (open N32)
    errors   UNS (Previous on an engine: replaces NoPrevious, open N33, before any run);
             FAIL; context.Canceled stays reachable with errors.Is
    answer   STR, stdout and stderr kept apart, no copy
    private  Logs, NoPrevious
    A        internal/runner/runner.go:1139 (cmd/solmq-conn-util/logs.go:211, :304)
    B        internal/container/manager.go:1401

#### Exec, ExecOutput, Attach (ruling 5)
    sig      Exec(ctx, name string, argv ...string) (Result, error)
             ExecOutput(ctx, name string, stdin []byte, argv ...string) ([]byte, Result, error)
             Attach(name string, o AttachOptions{Stdin, TTY bool; In, Out, Err *os.File},
                    argv ...string) (Result, error)
    cmd      <engine> exec [-i] [-t] <name> <argv...>      (no `--` on an engine)
    args     name: G-ctr; argv: non-empty list, tokens are the tool's; Attach: TTY needs Stdin
             (open N57); ExecOutput stdin non-nil adds -i
    errors   Exec, ExecOutput: non-zero child exit = FAIL carrying ExitCode; START.
             Attach: START only; a session that ran is nil error and the code is in Result
    answer   Exec ACT; ExecOutput ANS; Attach STR, hands over the terminal, installs the
             interrupt catcher and the Windows console mode for the call (open N16, N17)
    private  Exec (builder), containercli.ExecOptions shrinks to AttachOptions
    A        internal/runner/runner.go:1074 (ExecArgv); cmd/solmq-conn-util/shell.go:271
    B        internal/container/transport.go:67; manager.go:1473
    note     A's status script needs stdout and stderr merged in order: use Exec with
             With(buf, buf), not ExecOutput (open X16). A runs 4 at once: use per-call With.

#### Upload
    sig      Upload(ctx, name, dest string, body []byte, mkdir bool) (Result, error)
    cmd      <engine> exec -i <name> sh -c "[mkdir -p '<dir>' && ]cat > '<dest>'"
    args     name: G-ctr; dest: absolute path, G-free (gdu 4 grammar, path operand); body stdin
    errors   FAIL; START
    answer   ACT (stdin shown as a byte count only)
    private  Upload
    A        internal/runner/runner.go:1170 (mkdir true)
    B        internal/container/transport.go:111 (mkdir false)

#### CopyTo, CopyFrom (new, ruling 6)
    sig      CopyTo(ctx, name, local, dest string) (Result, error)
             CopyFrom(ctx, name, remote, local string) (Result, error)
    cmd      <engine> cp <local> <name>:<dest>   |   <engine> cp <name>:<remote> <local>
    args     name: G-ctr; local: G-free; dest, remote: G-free
    errors   FAIL (B counts failures per file); START
    answer   ACT
    private  new builders
    A        -
    B        internal/container/transport.go:119-136

#### Login (new, ruling 6)
    sig      Login(ctx, registry, user string, password []byte) (Result, error)
    cmd      <engine> login --username <user> --password-stdin [<registry>]
    args     user: G-free; registry: gdu 4 grammar (host[:port]); password on stdin, never argv
    errors   FAIL; START
    answer   ACT (stdin a byte count)
    private  new builder
    A        - (A relies on stored credentials)
    B        internal/container/manager.go:455-479

#### Run, Output
    sig      see 2.1
    cmd      <engine> <args the tool built>
    args     see 2.1
    errors   FAIL carrying ExitCode; START
    answer   Run ACT; Output ANS
    private  gdurunner.Combined, Split, Output (open N2)
    A        internal/runner/runner.go:356 (image pull), runParsed :86
    B        internal/container/manager.go:147-153 (Manager.run), :329 (`<engine> version`: drop,
             Info answers it)

### 2.3 podman.Client

    New(s runner.Setup, c podman.Config) (*Client, error)
      Config {Rules execguard.Rules; Command []string; Env []string}
    Checks: execguard.Check on Command; Env entries KEY=VALUE. Filled from the same base, runner
    and writers as the containercli client (guide says so). The session env reaches every podman
    command (open N38). The version is read once, on first use of CreateSecret or SecretMatches,
    and kept behind a pointer; an unanswered read counts as below 4.7. No Version method
    (ruling 13).

#### CreateSecret
    sig      CreateSecret(ctx, name string, value []byte, labels Labels) (Result, error)
    cmd      podman >= 4.7: podman secret create --replace [--label k=v]... <name> -
             below 4.7 or version unread: podman secret rm --ignore <name>, then
             podman secret create [--label k=v]... <name> -
    args     name: G-ctr (store name = container name, dash, suffix, derived in M); value stdin;
             labels: mandatory, keys and values checked per the label design note (gdu 1, 16)
    errors   FAIL naming the step; the message says a failed create after the rm leaves the
             secret gone until the next deploy; START
    answer   ACT; Result = last command run (open N5). Value never in Argv
    private  Version, SupportsReplace, CreateSecret
    A        internal/runner/runner.go:502, 508, 513 (version :480, main.go:957)
    B        internal/container/manager.go:612 (version :627-631)

#### RemoveSecrets
    sig      RemoveSecrets(ctx, names ...string) (Result, error)
    cmd      podman secret rm --ignore <names...>
    args     names: each G-ctr; none given = no-op, zero Result, nil error
    errors   FAIL (--ignore absorbs "already gone", so any failure is real)
    answer   ACT
    private  RemoveSecrets (the `ok` bool goes)
    A        internal/runner/runner.go:531 (one call, many names)
    B        internal/container/manager.go:605, 1101 (one name per call)

#### SecretMatches
    sig      SecretMatches(ctx, name string, value []byte) (bool, Result, error)
    cmd      podman secret inspect --showsecret --format '{{.SecretData}}' <name>   (>= 4.7 only)
    args     name: G-ctr; value compared inside with the pure podman compare
    errors   none for absent, unreadable, different, below 4.7: all (false, Result, nil), as both
             tools read them as "changed". START only for a bad name
    answer   SEC: the stored value reaches no writer, Result or error
    private  SecretData, SecretMatches
    A        internal/runner/runner.go:550-555
    B        internal/container/manager.go:554-562

#### SecretLabels (decided only: gdu 16)
    sig      SecretLabels(ctx) (map[string]Labels, Result, error)
    cmd      podman secret ls --format <names>; then podman secret inspect --format <name,labels>
             <names...>      (ls cannot filter by label, so two commands)
    args     none; the tool filters on product and instance with the pure compare
    errors   PART (inspect exits non-zero when a name is gone); UNA; REF
    answer   ANS decoded; Result = last command (open N5)
    private  new builders and parser
    A        -
    B        -

#### SecretIDs (decided only: gdu 10)
    sig      SecretIDs(ctx, names ...string) (map[string]string, Result, error)
    cmd      podman secret inspect --format <name,ID> <names...>
    args     names: each G-ctr
    errors   NF, PART (a named secret is gone: counts as changed)
    answer   ANS decoded; compared by a pure function with the IDs Inspect decodes
    private  new builder
    A        -
    B        -

#### IDMap
    sig      IDMap(ctx) (IDMaps{UID, GID string}, Result, error)      (open N42)
    cmd      podman info --format '{{range .Host.IDMappings.UIDMap}}...{{end}}...GIDMap...'
    args     none; rootless with a numeric run-as only (tool's condition)
    errors   UNA (both tools stop with "cannot read the range"); an empty answer IS an answer
    answer   ANS decoded; rootless.CheckIDMap(maps, uid, gid) is pure and returns *IDMapError
    private  IDMap (the gid bool goes); one-call template is unchecked on an engine (smoke row)
    A        internal/runner/runner.go:566
    B        internal/container/rootless.go:216

#### Unshare (new, ruling 6)
    sig      Unshare(ctx, argv ...string) (Result, error)
    cmd      podman unshare <argv...>
    args     argv: non-empty; tokens are the tool's (B: chmod, chown, find -delete, test -w)
    errors   FAIL carrying ExitCode (B reads any error as "not writable"); START
    answer   ACT; exit code in Result
    private  new builder
    A        -
    B        internal/container/manager.go:432, 438, 1153; rootless.go:439

#### Run, Output
    sig      see 2.1;  cmd podman <args>;  errors FAIL, START;  answer Run ACT, Output ANS
    A        -
    B        -

### 2.4 systemd.Client

    New(s runner.Setup, c systemd.Config) (*Client, error)
      Config {User bool; Env []string}
    Checks: Setup (runner set), Env entries KEY=VALUE. No tokens: the programs are the module's
    (systemctl, loginctl), so the guard has nothing to check (open X6). Env = rootless.Session.Env()
    passed as []string, so systemd imports no rootless. B's second client (system manager) is a
    second value, or the per-call option of Show.

#### DaemonReload
    sig      DaemonReload(ctx) (Result, error)
    cmd      systemctl [--user] daemon-reload
    args     none
    errors   FAIL; START
    answer   ACT
    private  DaemonReload
    A        internal/runner/runner.go:758, 804
    B        internal/container/manager.go:839, 1061

#### Start, Stop, Restart
    sig      Start(ctx, unit string) (Result, error)   // Stop, Restart alike
    cmd      systemctl [--user] start|stop|restart <unit>
    args     unit: gdu 4 grammar (unit name; leading '-' refused)
    errors   NF (unit not loaded; Stop exits 5 from systemd source: not run on a host, smoke row
             of todo item 9); FAIL; START
    answer   ACT
    private  Start, Stop, Restart
    A        internal/runner/runner.go:312, 314, 316; Stop :793
    B        internal/container/manager.go:844, 847, 851 (Start/Restart), 1028 (Stop), 1176,
             1188, 1200

#### State
    sig      State(ctx, unit string) (UnitState, Result, error)         (open N46)
    cmd      systemctl [--user] is-active <unit>
    args     unit: gdu 4 grammar (unit name)
    errors   UNA when nothing was read, never a state word (ruling 13); exit code ignored on
             purpose (non-zero for every state but active)
    answer   ANS typed UnitState; Active() and Stopped() become its methods; "unknown" is not
             Stopped
    private  IsActive, State (parser)
    A        internal/runner/runner.go:738-739
    B        internal/container/manager.go:968-969

#### LoadState (decided only: gdu 9)
    sig      LoadState(ctx, unit string) (string, Result, error)
    cmd      systemctl [--user] show <unit> --property=LoadState --value    (a wrapper over Show)
    args     unit: gdu 4 grammar (unit name)
    errors   UNA
    answer   ANS (loaded, not-found, masked ...)
    private  none new
    A        -
    B        -

#### NRestarts
    sig      NRestarts(ctx, unit string) (int, Result, error)
    cmd      systemctl [--user] show <unit> --property=NRestarts --value    (a wrapper over Show)
    args     unit: gdu 4 grammar (unit name)
    errors   UNA (empty or non-numeric: both tools fall back; replaces the `known` bool)
    answer   ANS
    private  NRestarts, ParseNRestarts
    A        internal/runner/runner.go:993-997
    B        internal/container/inspect.go:99-106

#### Show (new, ruling 6)
    sig      Show(ctx, unit string, o ShowOptions{System bool}, props ...string)
             (map[string]string, Result, error)                         (open N45)
    cmd      systemctl [--user] show [<unit>] --property=<p> [--property=<p>...]
             System=true drops --user; unit "" asks the manager itself (--property=Version)
    args     unit: gdu 4 grammar (unit name) or empty; props: [A-Za-z0-9]+, at least one
    errors   UNA when exit 0 and nothing printed (B: SKIP row); FAIL on a run error (B: FAIL row)
    answer   ANS decoded as property -> value
    private  Systemctl (generic builder)
    A        -
    B        internal/container/rootless.go:384; limits.go:182-190 (user@<euid>.service, system)

#### Linger, EnableLinger
    sig      Linger(ctx, uid int) (string, Result, error)
             EnableLinger(ctx) (Result, error)
    cmd      loginctl show-user <uid> --property=Linger   |   loginctl enable-linger
    args     uid >= 0 (negative refused); EnableLinger names no user, by design
    errors   Linger: UNA (message carries the "needs logind" hint); EnableLinger: FAIL
    answer   Linger ANS ("yes", "no" or what logind said); EnableLinger ACT. Both need the session
             env. The fix text is rootless.LingerFix(user) (new)
    private  ShowLinger, EnableLinger, ParseLinger
    A        internal/runner/runner.go:701-708 (Linger), :687 (Enable)
    B        internal/container/rootless.go:340-349 (Linger), :364 (Enable)

#### Run, Output
    sig      see 2.1;  cmd systemctl [--user] <args>
    args     see 2.1 (B: `status <unit> --no-pager`, `cat <unit>`)
    errors   FAIL carrying ExitCode (a warning in B); START
    answer   Run ACT; Output ANS
    A        -
    B        internal/container/manager.go:1329, 1375 (and 1049, which only prints the argv)

### 2.5 kube.Client (kubectl or oc)

    New(s runner.Setup, c kube.Config) (*Client, error)
      Config {Rules execguard.Rules; Command []string}
    Checks: execguard.Check on Command (Rules.Allowed kubectl, oc). Runs no command.
    (c *Client) Pin(ctx) (*Client, error)         // ruling 11, todo item 11
      With `--context` or `--context=` in Command: that one wins, nothing runs. Else
      `<kubectl> config current-context`; returns a NEW client with `--context=<name>` in its
      tokens. No context (kubectl in a pod): returns the unpinned receiver and a NOC error, which
      the tool treats as "continue" after WhoAmI (open X5). B's mate channel calls New and Pin
      once per site.

#### Pin (CurrentContext)
    sig      Pin(ctx) (*Client, Result, error)
    cmd      <kubectl> config current-context
    args     none; the answer is checked as G-ctx before it goes on an argv
    errors   NOC (nothing set); UNA; START (name fails G-ctx)
    answer   ANS (the resolved name via Context() string)
    private  none new; the find-and-pin helpers are pure and private
    A        - (gdu 11 adds A)
    B        internal/cli/context.go:254-264 (display only today)

#### WhoAmI (decided only: gdu 11)
    sig      WhoAmI(ctx) (string, Result, error)
    cmd      <kubectl> auth whoami
    args     none
    errors   REF; UNS (cluster older than 1.28); UNA (not decided what the tool does)
    answer   ANS (identity)
    private  new builder
    A        -
    B        -

#### Apply, Delete
    sig      Apply(ctx, manifest []byte) (Result, error)
             Delete(ctx, manifest []byte, timeout time.Duration) (Result, error)
    cmd      <kubectl> apply -f -   |   <kubectl> delete -f - --ignore-not-found --timeout=<t>
    args     manifest: non-empty, stdin only, never argv; timeout G-dur (default constant beside
             RolloutTimeout; the owner times a real broker remove first, gdu 12)
    errors   REF; FAIL; Delete also TO (accepted and continuing in the cluster)
    answer   ACT (A scans an apply's own output with With(buf, w))
    private  Apply, Delete
    A        internal/runner/runner.go:144 (Apply), :146 (Delete)
    B        internal/k8s/cluster.go:190 (Apply), :196 (Delete)

#### NamespaceState, NamespaceLabels
    sig      NamespaceState(ctx, ns string) (NamespaceState, Result, error)        (open N50)
             NamespaceLabels(ctx, ns string) (map[string]string, Result, error)    (open N53)
    cmd      <kubectl> get namespace <ns> --ignore-not-found -o name
             <kubectl> get namespace <ns> -o <template printing the labels only>
    args     ns: G-lbl
    errors   NamespaceState: a failed or odd read returns NamespaceUnknown WITH the typed error
             (REF, UNA, NOC); an empty answer stays Absent (--ignore-not-found prints nothing).
             NamespaceLabels: a refused or failed read is the tool's "unlabelled"
    answer   ANS decoded. NamespaceLabels is for the remove gate only
    private  GetNamespace, ParseNamespace
    A        internal/runner/runner.go:269-273 (State)
    B        internal/k8s/prep.go:52-56 (State)

#### DeleteNamespace
    sig      DeleteNamespace(ctx, ns string, timeout time.Duration) (Result, error)
    cmd      <kubectl> delete namespace <ns> --ignore-not-found --timeout=<t>
    args     ns: G-lbl, and not a protected name (kube.Protected, open N52); timeout G-dur
    errors   TO; REF; FAIL; START (protected or bad ns: not a cluster refusal, open X4)
    answer   ACT
    private  DeleteNamespace; nsgate.Protected moves to kube
    A        internal/runner/runner.go:279
    B        internal/k8s/prep.go:129

#### DeleteNamed (replaces DeletePod)
    sig      DeleteNamed(ctx, kind string, names []string, ns string, timeout time.Duration)
             (Result, error)
    cmd      <kubectl> delete <kind> <names...> [-n <ns>] --ignore-not-found --timeout=<t>
    args     kind: gdu 4 grammar (kind; "namespace" refused, it goes through DeleteNamespace);
             names: each G-sub; ns: G-lbl, empty for a cluster-scoped kind (the PV); timeout G-dur
    errors   TO; REF; FAIL; START. The loop and the policy stay in the tool
    answer   ACT
    private  none (new)
    A        - (gdu 18 adds it)
    B        internal/k8s/ops.go:138 (pod), prep.go:283 (secret), deploy.go:215 (pvc)

#### Scale, RolloutRestart, RolloutStatus
    sig      Scale(ctx, kind, name, ns string, replicas int) (Result, error)
             RolloutRestart(ctx, kind, name, ns string) (Result, error)
             RolloutStatus(ctx, kind, name, ns string, timeout time.Duration) (Result, error)
    cmd      <kubectl> scale <kind>/<name> [-n <ns>] --replicas=<n>
             <kubectl> rollout restart <kind>/<name> [-n <ns>]
             <kubectl> rollout status <kind>/<name> [-n <ns>] --timeout=<t>
    args     kind: gdu 4 grammar (kind); name: G-sub; ns: G-lbl or empty; replicas >= 0;
             timeout G-dur (zero = RolloutTimeout, 300s)
    errors   REF (a warning and a success in both tools, parity rule 16); TO and FAIL for the
             wait; FAIL
    answer   ACT (rollout progress)
    private  Scale, RolloutRestart, RolloutStatus; kube.Forbidden (reads stderr alone)
    A        internal/runner/runner.go:241, 243 (Scale), :245 (Restart), :262 (Status);
             cmd/solmq-conn-util/lifecycle.go:136 (Forbidden)
    B        internal/k8s/ops.go:182, 200 (Scale), :158 (Status), :159 (Forbidden);
             operator.go:590 (Restart), :616 (Scale)

#### Get
    sig      Get(ctx, o GetOptions, what ...string) ([]byte, Result, error)
    cmd      <kubectl> get <what...> [--all-namespaces | -n <ns>] [-l <sel>] [--ignore-not-found]
             [-o <output>]
    args     what: each G-free (type lists and dotted names allowed); Namespace G-lbl;
             Selector gdu 4 grammar (label selector); Output one of name|json|wide|jsonpath=...
             (the custom-columns read of B moves to jsonpath); IgnoreNotFound never on a list
    errors   REF; NF (when IgnoreNotFound is off; matched on kubectl's NotFound line only);
             UNA (exit 0, nothing on stdout where a document was asked for)
    answer   ANS raw bytes for the tool's own decoder. A Secret read through Get is the tool's
             risk: SecretKeys and SecretValue are the safe forms (open X9)
    private  Get, kube.GetOptions stays
    A        internal/runner/runner.go:841, 859
    B        internal/k8s/deploy.go:74; adminsecret.go:78; statusreport.go:378, 432, 452;
             inspect.go:66; ops.go:20, 23, 26; operator.go:633, 638; prep.go:353, 378

#### ListLabelled (decided only: gdu 18)
    sig      ListLabelled(ctx, ns string, kinds []string, sel Selector{Product, Instance string})
             ([]Object, Result, error)
    cmd      <kubectl> get <kind,kind,...> -n <ns> -l <product>=<tool>,<instance>=<inst> -o name
    args     ns: G-lbl; kinds: gdu 4 grammar (kind); sel: label value grammar (label design note)
    errors   PART (kubectl prints some kinds and exits 1: no names returned); REF; UNA
    answer   ANS decoded as []Object (kind folded: `deployment.apps/x` equals Deployment)
    private  new builder and parser
    A        -
    B        -

#### Logs
    sig      Logs(ctx, target string, o LogOptions{Namespace, Container string; Previous, Follow,
             Timestamps bool; Tail int; Since time.Duration}) (Result, error)
    cmd      <kubectl> logs <target> [-n <ns>] [-c <c>] [-p] [-f] [--timestamps] [--tail <n>]
             [--since <t>]
    args     target: G-free (pod/<x>, deployment/<x>); Namespace G-lbl; Container G-lbl;
             Tail and Since as the engine Logs (open N32)
    errors   FAIL; context.Canceled reachable
    answer   STR
    private  Logs
    A        internal/runner/runner.go:1127
    B        internal/k8s/ops.go:43; operator.go:646

#### Exec, ExecOutput, Attach
    sig      Exec(ctx, t Pod, argv ...string) (Result, error)
             ExecOutput(ctx, t Pod, stdin []byte, argv ...string) ([]byte, Result, error)
             Attach(t Pod, o AttachOptions, argv ...string) (Result, error)
             Pod {Namespace, Name, Container string}
    cmd      <kubectl> exec [-i] [-t] <pod> [-n <ns>] [-c <c>] -- <argv...>
    args     Name: G-sub; Namespace G-lbl; Container G-lbl; argv as the engine Exec
    errors   as the engine Exec; kubectl gives "pod not reachable" and "command exited 1" the same
             status, so no NF is promised
    answer   as the engine Exec
    private  kube.Exec (builder), ExecOptions shrinks
    A        internal/runner/runner.go:1067; cmd/solmq-conn-util/shell.go:256-271
    B        internal/k8s/transport.go:55; ops.go:53; matechannel.go:88, 98
    note     A names the container on every call (-c connector), B never does. One small
             interface serves a pod and a container target (open X1)

#### Upload, CopyTo, CopyFrom
    sig      Upload(ctx, t Pod, dest string, body []byte, mkdir bool) (Result, error)
             CopyTo(ctx, t Pod, local, dest string) (Result, error)
             CopyFrom(ctx, t Pod, remote, local string) (Result, error)
    cmd      <kubectl> exec -i <pod> [-n <ns>] [-c <c>] -- sh -c "[mkdir -p '<dir>' && ]cat > '<d>'"
             <kubectl> cp -n <ns> <local> <pod>:<dest>
             <kubectl> cp -n <ns> <pod>:<remote> <local>
    args     as the engine Upload and Copy; Pod as Exec
    errors   FAIL; START
    answer   ACT
    private  Upload (builder); new Copy builders
    A        internal/runner/runner.go:1175 (Upload, mkdir true, -c connector)
    B        internal/k8s/transport.go:92 (Upload), :100-117 (Copy); matechannel.go:69

#### SecretKeys (new, open N56)
    sig      SecretKeys(ctx, ns, name string) ([]string, Result, error)
    cmd      <kubectl> get secret <name> [-n <ns>] -o <template printing the data key names only>
    args     ns: G-lbl; name: G-sub
    errors   NF (A stops the deploy); REF (A warns); FAIL
    answer   SEC-safe: values never reach the tool (template unchecked on a cluster)
    private  new builder
    A        cmd/solmq-conn-util/secretprobe.go:36 (via internal/runner/runner.go:859);
             status.go:582 (existence only: use Get with -o name, not this)
    B        -

#### SecretValue (new, ruling 6)
    sig      SecretValue(ctx, ns, name, key string) ([]byte, Result, error)
    cmd      <kubectl> get secret <name> -n <ns> -o jsonpath={.data.<key>}   (decoded in M)
    args     ns: G-lbl; name: G-sub; key: [-._A-Za-z0-9]+, '.' escaped for jsonpath
    errors   NF (secret, or key with an empty answer); REF; FAIL
    answer   SEC: never to a writer, the Result or an error
    private  new builder
    A        -
    B        internal/k8s/matechannel.go:139-160

#### ServerVersion (new, ruling 6)
    sig      ServerVersion(ctx) (Versions{Client, Server string}, Result, error)
    cmd      <kubectl> version -o json
    args     none
    errors   UNA (API server cannot be reached: B stops validate)
    answer   ANS decoded (B discards it today; also settles "older than 1.28")
    private  new builder
    A        -
    B        internal/k8s/check.go:39-44

#### Top (new, ruling 6)
    sig      Top(ctx, o TopOptions{Namespace string; AllNamespaces bool; Selector string},
             names ...string) ([]TopRow, Result, error)
    cmd      <kubectl> top pod [<names...>] [--all-namespaces | -n <ns>] [-l <sel>] --containers
             --no-headers
    args     names: each G-sub; Namespace G-lbl; Selector as Get
    errors   FAIL (A turns it into a note: no metrics API)
    answer   ANS decoded: per pod and container
    private  new builder and parser, lifted from A's statusreport
    A        internal/runner/runner.go:886-904
    B        -

#### MissingNamespace (not an operation: a method, open N60)
    sig      (c *Client) MissingNamespace(ns string, err error) error
    note     words the "namespace does not exist and this identity may not create it" error; the
             tokens stay inside the client
    A        cmd/solmq-conn-util/namespace.go:75
    B        internal/k8s/prep.go:72

#### Run, Output
    sig      see 2.1;  cmd <kubectl> <args>
    args     see 2.1 (B: describe, set env WATCH_NAMESPACE; the pinned context applies to each)
    errors   FAIL carrying ExitCode; START
    answer   Run ACT; Output ANS
    A        -
    B        internal/k8s/ops.go:31, 36; operator.go:652; watch.go:456

### 2.6 kube/rbac and kube/nsgate

#### rbac.Run
    sig      rbac.Run(ctx, kc *kube.Client, p Probe) (runner.Result, error)
    cmd      <kubectl> auth can-i <verb> <resource>[/<name>] [--subresource <s>]
             (-n <ns> | --all-namespaces | nothing)
    args     Verb: gdu 4 grammar (verb); Resource: gdu 4 grammar (plural resource, [a-z.]+);
             Name: G-sub; Namespace: G-lbl; Subresource: G-free (no caller in either tool)
    errors   *RefusedError, which also answers errors.Is(err, runner REF) (open N61); UNA (run
             error with no verdict, or exit 0 without "yes"; carries the login hint); NOC when
             kubectl reports no configuration
    answer   PRB: the verdict is the stdout; nothing is written
    private  CanI, Check, Answer, Refused; Verdict is removed (unused)
    A        internal/runner/runner.go:408-409
    B        internal/k8s/preflight.go:74, 95, 120
    note     the set of probes and the loop stay in each tool

#### nsgate.Read
    sig      nsgate.Read(ctx, kc *kube.Client, ns string) (Listing, runner.Result, error)
    cmd      today: <kubectl> get all,persistentvolumeclaims,secrets,configmaps -n <ns> -o json
             recommended (open N49): metadata-only template for every kind, and with gdu 1 the
             kinds from `<kubectl> api-resources --verbs=list --namespaced -o name` (two commands,
             Result = last)
    args     ns: G-lbl (Protected is not checked here: Read is read-only)
    errors   REF; PART; UNA (nothing on stdout: B skips quietly)
    answer   SEC: Secret and ConfigMap values never reach the tool, a writer or an error (A prints
             a failed listing's output today)
    private  nsgate.List
    A        internal/runner/runner.go:871
    B        internal/k8s/namespace.go:57
    pure     Occupants(Listing, Owned) []Object (open N55); Gate(...) decision function (open
             N54); DeleteProbe(ns) asks about one namespace (open N51)

## 3. The 13 gap operations (findings 2.2)

    1  delete by kind and name, bounded    kube.Client.DeleteNamed (2.5). Replaces DeletePod.
    2  exec returning stdout, stdin bytes   ExecOutput on containercli.Client and kube.Client,
                                            beside Exec and Attach (ruling 5).
    3  run a tool-built command             Run and Output on every client (2.1); host commands
                                            (mkdir, chown, find, sysctl) stay tool-built: B
                                            runs them over the runner it holds, no client.
    4  Secret read without values           kube.Client.SecretKeys, SecretValue; A's status
                                            existence read is Get -o name (open X9).
    5  systemctl show, also system manager  systemd.Client.Show with ShowOptions{System}.
    6  compose version                      containercli.Client.ComposeVersion.
    7  registry login, password on stdin    containercli.Client.Login.
    8  copy in and out                      CopyTo and CopyFrom on containercli.Client and
                                            kube.Client.
    9  podman unshare                       podman.Client.Unshare (what runs inside stays B's).
    10 API server answers                   kube.Client.ServerVersion.
    11 image inspect, decoded               containercli.Client.InspectImage with NF.
    12 resource usage                       containercli.Client.Stats and kube.Client.Top.
    13 timed-out vs failed                  Kind TO, returned by Delete, DeleteNamespace,
                                            DeleteNamed and RolloutStatus.

Stays tool-built via Run (findings 2.5): image pull (A, plain argv, nothing to parse), describe
(B), `set env WATCH_NAMESPACE` (B, the operator's own variable), `systemctl status` and `cat`
(B, display only), `<engine> version` row (B: drop, Info answers it), the in-broker commands
(over Exec, ExecOutput, Upload), the StorageClass reads (over Get). `wait --for=condition=Ready`
(gdu 17) is added only if the owner check says OnDelete (open N62).

## 4. Methods neither tool needs today (findings 2.3)

    podman Version                 Drop as a method; the read is private (ruling 13). Both tools
                                   read it only to choose replace or the fallback.
    kube DeletePod                 Drop; DeleteNamed covers it (B ops.go:138 and two more).
    systemd.LoadState (gdu 9)      Keep: a decided item needs it. Callers in both tools, same tag.
    podman.SecretIDs (gdu 10)      Keep, same reason.
    podman.SecretLabels (gdu 16)   Keep, same reason.
    kube.WhoAmI (gdu 11)           Keep, same reason.
    kube.ListLabelled (gdu 18)     Keep, same reason.
    kube.NamespaceLabels (gdu 1)   Keep (open N53): the remove gate needs it.
    kube.Client.Pin               Both tools call it (A is new by gdu 11).
    Stats, Top, InspectImage       A only, kept (ruling 6).
    Login, SecretValue, Unshare,   B only, kept (ruling 6).
    CopyTo, CopyFrom, ServerVersion, ComposeVersion, systemd.Show
    rbac.Probe.Subresource and     No non-test caller in either tool (findings 2.8); the fields
    AllNamespaces                  stay as listed under "No decision needed"; drop is the owner's.
    Run and Output on podman and   No caller today; ruling 2 puts them on every client.
    systemd

## 5. Host facts (ruling 7)

M reads each in Go. No `sh`, `test`, `cat` or `id` child, no SubIDFiles, DirProbe or id builder.
Reads go through an injected `rootless.Host`, because M has no globals and its tests must not
touch the machine:

    package rootless
      type Host struct {
          ReadFile func(path string) ([]byte, error)
          Stat     func(path string) (fs.FileInfo, error)
          Getenv   func(key string) string
          LookupID func(uid string) (*user.User, error)
      }
      func OSHost() Host                       // the real functions; the tool passes it in

      func (h Host) SubIDs() (subuid, subgid string, readable bool)
          reads /etc/subuid and /etc/subgid; readable=false (either file unreadable) makes
          SubIDFix print its "check this does not overlap" caveat; feeds NextFreeRange
      func (h Host) RuntimeDir(euid int) (dir string, exists bool, err error)
          $XDG_RUNTIME_DIR, else /run/user/<euid>; os.Stat: a missing dir is (dir,false,nil),
          any other error is returned; replaces the isDir argument of UserSession
      func UserSession(euid int, h Host) (Session, error)       // signature change
      func (h Host) AccountName(uid int) string
          os/user LookupID, then the numeric uid (open N40); never `id -un`, never $USER
      func (h Host) HardLimit(name string) (int, error)
          reads /proc/self/limits (open N44); ProcLimits goes

Files read: /etc/subuid, /etc/subgid, /proc/self/limits, the runtime directory (stat only).
Account name: os/user, which under CGO_ENABLED=0 reads /etc/passwd and misses directory accounts.
One test for "does the repair apply": Linux and a real uid (open N39). The pure parts stay:
Decide, Session.Env and String, CheckLinger, ParseLimit, LimitText, CheckCore (open N43).
New or changed pure functions: `SubIDRefusal(cause, user, uid, gid, maps)` assembling the whole
refusal (NextFreeRange and SubIDFix private, range sized from the larger of uid and gid plus one,
open N41); `CheckIDMap(maps, uid, gid) *IDMapError` (open N42); `LingerFix(user)` (added).

## 6. Exported names that go private, move or change

Counts are mine, from "No decision needed" in the questions file and the N entries, for the
packages this draft covers (runner, execguard, containercli, inspect, podman, systemd, rootless,
kube, rbac, nsgate). The review's 48, 27, 5, 6 and 4 cover all packages and are not redone here.

Go private behind a client method or because only M uses them (58):

    containercli (14): Info, InfoError, EngineHint, ComposeUp, ComposeDown, Start, Stop, Restart,
                       List, ParseList, Upload, Logs, Exec, NoPrevious (N33)
    kube (13):         Apply, Delete, Scale, RolloutRestart, RolloutStatus, Logs, Upload,
                       GetNamespace, ParseNamespace, DeleteNamespace, Get, Exec, Forbidden
    systemd (12):      DaemonReload, Start, Stop, Restart, IsActive, State, NRestarts,
                       ParseNRestarts, ShowLinger, EnableLinger, ParseLinger, Systemctl
    podman (7):        Version, SupportsReplace, CreateSecret, SecretData, SecretMatches,
                       RemoveSecrets, IDMap
    rbac (4):          CanI, Check, Answer, Refused
    inspect (1):       Containers
    nsgate (1):        List
    runner (3):        Combined, Split, Output (N2)
    rootless (3):      ProcLimits (N44), NextFreeRange, SubIDFix (N41); HardLimit becomes a
                       Host method that reads the file (N44); UserSession loses its isDir argument

Removed (2): containercli.Remove (CR-0001, already decided), rbac.Verdict (unused after Run).

Move or rename (7):

    inspect.Printable            -> runner.Printable, plus a form keeping newline and tab (N27)
    nsgate.Protected             -> kube.Protected (N52)
    nsgate.Object                -> kube.Object (N55)
    kube.ValidLabel              -> kube.ValidDNSLabel; ValidDNSSubdomain added (N59)
    execguard.IsControl and IsInvisible -> a neutral character package (N12)
    runner.Quote                 -> runner.DisplayQuote (N7)

Change signature or shape (18):

    nsgate.Occupants(out, owned) -> Occupants(Listing, Owned) []Object (N55)
    nsgate.DeleteProbe()         -> DeleteProbe(ns) (N51)
    kube.MissingNamespace        -> method on Client (N60)
    kube.NamespaceManifest       -> a Namespace document value (N58)
    kube.GetOptions, kube.ExecOptions, kube.LogOptions: LogOptions Tail int and Since Duration
                                 (N32); ExecOptions shrinks to Pod and AttachOptions (ruling 5)
    containercli.LogOptions      -> Tail, Since, Previous (N32, N33)
    containercli.ExecOptions     -> AttachOptions (ruling 5, N57)
    kube.RolloutTimeout          -> time.Duration constant (open X7)
    runner.Exec.Announce         -> takes the Cmd (N7)
    runner.Echo.W                -> required, no stdout default (ruling 12)
    inspect.Container.Secrets    -> name and ID (gdu 10)
    systemd.Active, Stopped      -> methods of UnitState; "unknown" no longer Stopped (N46)
    rootless.UserSession         -> (euid, Host) (ruling 7)
    rootless.CheckIDMap          -> (maps, uid, gid) *IDMapError (N42)
    execguard.IsControl          -> includes U+0085 (N13; behaviour)
    podman.RemoveSecrets         -> no `ok` bool; zero names is a no-op
    podman.CreateSecret          -> mandatory Labels (gdu 16)
    inspect.Containers           -> error, not `ok` bool (settled)

Added (names): runner.Result, Setup, Error and seven kinds; execguard.CheckValue (N11),
ApproveName (N10); containercli.CheckName (N30), Engine.Title (N34), SafeNames; kube.Pin,
WhoAmI, DeleteNamed, ListLabelled, NamespaceLabels, SecretKeys, SecretValue, ServerVersion, Top,
Copy; the clients and their `New` for containercli, podman, systemd, kube; rbac.Run, nsgate.Read
and Gate; rootless.Host and its methods, LingerFix, SubIDRefusal.

Stay (no change): Cmd, Runner, Exec and its methods, Resolve, CheckRun, CheckAttach, Echo, MaskEnv,
Engine, Docker, Podman, Row, HasName, ByImage, ValidName, ComposeDigest, inspect.Decode, State,
Health, Mount, NamespaceState and its constants, Probe, ClusterScoped, RefusedError,
podman.QuadletDir and SystemQuadletDir (N48), Decide, Session, Env, String, CheckLinger,
Unbounded, ParseLimit, LimitText, CheckCore, execguard.Rules, Approvals, Check, Approve.

## 7. Open points

Open N entries this draft depends on (recommendation used):

- open N1: M marks exit 2 only for execguard and envref errors; New returns the guard error
  unwrapped.
- open N2: Combined, Split and Output go private; tools move to client.Run and client.Output.
- open N3: after a command ran Result.ExitCode is never negative (a signalled child is 1); -1 is
  "never started".
- open N4: exitcode.ChildStatus stays exported; runner.Error carries ExitCode().
- open N5: Result describes the last command run; the method doc lists the commands.
- open N6: Exec.Grace and DefaultGrace stay; add a smoke row for piped compose and pull output.
- open N7: one masked display form on Cmd, Announce takes the Cmd, Quote becomes DisplayQuote,
  MaskEnv and Call.String stay.
- open N8: NoEnv stays in B; only ComposeUp, ComposeDown and the session env set Cmd.Env.
- open N9: each client states the programs it issues; a tool picks its subset for Rules.Allowed.
- open N11: execguard exports CheckValue; tilde refused leading only.
- open N12: a neutral character package holds IsControl, IsInvisible, UnsafeChars, CheckValue.
- open N13: IsControl includes U+0085.
- open N14: one exported CheckHostPath, leading tilde refused.
- open N16: one terminal package; Attach turns on the Windows console mode.
- open N17: Attach installs and removes the interrupt catcher for the call.
- open N26: containercli/inspect stays as the plain decode package; only the builder goes private.
- open N27: Printable moves to runner with a form that keeps newline and tab.
- open N28: the stderr field holds exact bytes; Error() uses a cleaned copy.
- open N29: Inspect returns NF for one missing name, PART for several; Start, Stop, Restart do
  not guess NF; add a smoke row for the missing-name exit code.
- open N30: ValidName stays the engine grammar; CheckName words the DNS-label rule for creation.
- open N31: List gets a label filter; Row gains no Labels; ByImage stays.
- open N32: Tail becomes a count, Since a time.Duration; zero or negative Since refused.
- open N33: containercli.LogOptions gains Previous; Logs refuses it; NoPrevious goes private.
- open N34: add SafeNames, Engine.Title, find-by-name and the digest-plus-env helper; no running
  helper.
- open N38: the session env goes to systemd and the podman clients; B stops using os.Setenv.
- open N39: the user-session repair applies on Linux with a real uid only.
- open N40: account name by os/user, then the numeric uid; smoke row for a number in the fix text.
- open N41: one function assembles the sub-id refusal; range sized from the larger of uid and gid
  plus one.
- open N42: IDMap returns both maps from one call; CheckIDMap returns *IDMapError.
- open N43: the six core-limit helpers stay in M; the parity body line is reworded.
- open N44: /proc/self/limits is read in Go; ProcLimits goes.
- open N45: one Show(unit, props...) returning a map; manager selectable per call.
- open N46: State returns a typed UnitState and UNA; Active and Stopped become its methods.
- open N47: no composite that resolves the whole podman scope.
- open N48: QuadletDir and SystemQuadletDir stay in podman.
- open N49: the namespace listing asks for a metadata-only template, no -o json.
- open N50: NamespaceState returns the state and the typed error together; zero value Unknown.
- open N51: DeleteProbe(ns) asks about the one namespace, without --all-namespaces.
- open N52: Protected moves to kube; Client.DeleteNamespace refuses a protected name.
- open N53: a separate NamespaceLabels method for the remove gate; a failed read is "unlabelled".
- open N54: nsgate gains one pure gate function for the question table.
- open N55: Object lives in kube; Occupants takes an Owned value and returns []Object.
- open N56: add SecretKeys (names only) beside SecretValue.
- open N57: Attach keeps two switches, input and terminal; a terminal without input is refused.
- open N58: a Namespace document value in a package under kube.
- open N59: rename ValidLabel to ValidDNSLabel, add ValidDNSSubdomain and a label-value rule.
- open N60: MissingNamespace becomes a Client method.
- open N61: *RefusedError also answers errors.Is for the runner refused error.
- open N62: add no pod-wait method now.

New open points from this draft (no N entry exists; the recommendation is mine):

- open X1: the pod and container target interface (Exec, ExecOutput, Attach, Upload, CopyTo,
  CopyFrom, Logs) lives in runner, with thin bound types in containercli and kube.
- open X2: a fourth stream class "probe" (Info, ComposeVersion, ServerVersion, can-i) drops stdout;
  ruling 1 as worded would print `docker info` on every deploy (findings 2.8).
- open X3: ruling 4's seven kinds lack "failed with an exit code" and "never started"; add both as
  a runner.Error with no Kind, plus START for argument and guard failures.
- open X4: a refusal made by the module itself (protected namespace, bad argument) is not a REF;
  it is a START-class validation error.
- open X5: Pin returns the unpinned client with NOC; the tool then calls WhoAmI (ruling 11 needs a
  command at start-up, ruling 3 says set-up only guards).
- open X6: systemd.New has no tokens, so ruling 3's guard check has nothing to run there.
- open X7: kube.RolloutTimeout changes from the string "300s" to a time.Duration.
- open X8: a missing compose plugin is UNS with B's hint.
- open X9: Get on a Secret with -o json is not refused in this draft; the owner may want the
  client to refuse it so SecretKeys and SecretValue are the only reads.
- open X10: the podman version cache and the pinned context live behind a pointer and need
  sync.Once; "no global state" holds because the state is per client.
- open X11: the gdu 4 grammar table is still to be written for: unit name, kind, verb, resource,
  image reference, registry, label selector, path operand.
- open X12: Error() text does not repeat stderr; a tool whose writer discards stderr loses today's
  "<cmd>: <output>" wording.
- open X13: Run and Output check their args only for control characters and an empty list.
- open X14: B allows nerdctl and docker-compose; Config.Engine allows Docker or Podman only
  (ruling 3), so nerdctl runs as Docker with its own tokens.
- open X15: smoke rows to add: missing-name inspect exit code (N29), piped compose output (N6),
  can-i for a named namespace (N51), SecretKeys and namespace templates (N49, N56), the one-call
  IDMap template (N42), systemd exit codes 5, 3 and 4 (todo item 9), a number in the linger fix
  line (N40).
- open X16: A's status script needs both streams merged in order, so it uses Exec with
  With(buf, buf); ExecOutput returns stdout only (findings 2.8).

Where the rulings collide with the code as it is:

- Ruling 1 against probes: see X2. Ruling 1 against A: `runParsed` folds stdout into error text
  (internal/runner/runner.go:86-91) and a failed namespace listing prints Secret documents; ruling
  4 removes that.
- Ruling 3 against ruling 11: see X5. Ruling 3 against systemd: see X6.
- Ruling 4 "result returned on failure too": `runner.Exec.Run` returns only an error and `Attach`
  returns (int, error), so each client builds the Result; Duration and masked Argv are new.
- Ruling 5 "a non-zero exit is an error": `Exec.Run` today wraps *exec.ExitError as
  "<name>: <err>"; exitcode.ChildStatus finds it through ExitCode(), which runner.Error keeps.
- Ruling 7 against design 3.3 rule 4 ("never reads the environment or a file") and the doc of
  `rootless` ("reads nothing"): both are reworded; the injected Host keeps tests off the machine.
- Ruling 13 `systemd.Stopped`: `Stopped("unknown")` is true in the code; the draft removes it.
  `ParseNamespace` keeps an empty answer as Absent (the questions file, ruling 13 settled).
- Ruling 6 `podman unshare` and ruling 5's single exit-code rule: B reads any error, not the exit
  code, as "not writable" (findings 2.8).
- The questions file lists `ExecOptions`, `LogOptions` and `RolloutTimeout` under "stay as they
  are" and elsewhere shrinks or retypes them (ruling 5, N32); this draft takes the change.
- Design section 3.4 lists `kube.Client.CurrentContext`; this draft makes it `Pin` (X5).
- gdu 4 builders return errors; the lib annotation, labels and label design note (review 6 item 4)
  are not designed here, so `labels Labels` and `Selector` are placeholders.
