# Reliant workflow demo: fix, review, approve

A small Go repository for trying one [Reliant](https://reliantlabs.io) workflow
on a real test suite. You break a test, and the workflow:

1. **runs the tests** in your checkout;
2. has an **agent fix the failure in its own git workspace**, re-running the
   tests itself until they pass (at most 3 rounds);
3. **checks that no test file was touched**, then has a **second agent review
   the diff**;
4. **pauses for your approval** before anything is published. By default
   nothing is published even after you approve: the fix stays on the
   workspace's local branch.

The workflow is two YAML files in [`.reliant/workflows/`](.reliant/workflows/),
with 13 scenario tests beside them.

## What's in the repo

| Path | What it is |
|---|---|
| `pricing/` | A cart pricer: subtotal, discount codes, free shipping from $50.00, 8.25% tax, all in integer cents |
| `pricing/pricing_test.go` | The test suite: a table test plus three edge-case tests |
| `cmd/quote/` | A CLI around it: `go run ./cmd/quote -code WELCOME10 kettle:5000:1` |
| `scripts/break-it.sh` | Commits a realistic off-by-one bug on a new branch |
| `.reliant/workflows/fix-review-approve.yaml` | The workflow you run |
| `.reliant/workflows/fix-until-green.yaml` | Its fix-and-retest loop, called with `ref: project://fix-until-green` (hidden from pickers) |
| `.reliant/workflows/*/scenarios/` | One scenario per file, one for every ending |

The only dependency is Go 1.22 or later. There are no modules to download.

## Try it

### 1. Check it without an account

`validate`, `lint` and `scenario run` read the files on disk. They need no
sign-in, no machine and no network, and they call no model
([Workflows from the terminal](https://docs.reliantlabs.io/cli/workflows)):

```bash
go test ./...                     # green
reliant workflow validate         # structure, CEL, types, project:// refs
reliant workflow lint             # unattended-safety rules
reliant workflow scenario run     # 13 scenarios on the real engine, nodes mocked
```

What we saw with reliant 1.29.0:

```text
  ✓ fix-review-approve.yaml (fix-review-approve)
  ✓ fix-until-green.yaml (fix-until-green) (2 warnings)
2 workflow(s): 0 lint error(s), 0 warning(s)
Running 13 scenario(s) across 2 workflow(s)
...
  Passed:  13
  Failed:  0
```

The two `fix-until-green` warnings say the types of two outputs are only known
when the run happens (`dyn`). They don't block anything.
`unattended_run_publishes_nothing` also prints `input "unattended" is not
defined in workflow schema` and still passes.

### 2. Break a test

```bash
./scripts/break-it.sh
```

This creates branch `try/break-shipping` and commits one change in
`pricing/pricing.go`: `discounted < FreeShippingThreshold` becomes `<=`. The
result is that an order of exactly $50.00 now gets charged shipping. One test
fails:

```text
--- FAIL: TestPrice (0.00s)
    --- FAIL: TestPrice/order_exactly_at_the_threshold_ships_free (0.00s)
        pricing_test.go:59: Price() = {Subtotal:5000 Discount:0 Shipping:599 Tax:413 Total:6012}, want {Subtotal:5000 Discount:0 Shipping:0 Tax:413 Total:5413}
FAIL
FAIL	example.com/checkout/pricing	0.004s
```

The workspace starts from your **committed** code, so the bug has to be
committed. The script does that for you. To break something yourself instead,
change the logic in `pricing/pricing.go` (a rate, a rounding rule, a code's
percentage) and commit it. You can also break `pricing_test.go`, but then the
right fix *is* a test change, and the guard in step 4 refuses that by design. To undo:
`git checkout main && git branch -D try/break-shipping`.

### 3. Run it by hand

**Start the run yourself, from a chat or the CLI. Don't wire it to a trigger.**
A run an automation starts (a schedule, webhook, GitHub event or workflow event)
is unattended. Its approval doesn't wait: it resolves at once with status
`unattended`, and this workflow then ends **failed without publishing**
([Unattended by design](https://docs.reliantlabs.io/automations/overview#unattended-by-design),
[Require human approval](https://docs.reliantlabs.io/guides/workflows/require-human-approval)).
That outcome is safe, but no person ever decides anything. The human gate only
works in a run someone started by hand. The workflow declares no `triggers:`.

To run it, you need a signed-in CLI and a connected machine that holds this
checkout ([Sign in](https://docs.reliantlabs.io/cli/sign-in),
[Run a machine](https://docs.reliantlabs.io/cli/run-a-machine)). Then either:

- **In the app:** open a chat in this project, pick **fix-review-approve** from
  the workflow selector, and send a message such as "Fix the failing test."
- **From the terminal:**

  ```bash
  reliant workflow run fix-review-approve --project-path . \
    -m "Fix the failing pricing test." --exit-on-gate
  ```

  `--exit-on-gate` exits with code `3` once the run is waiting on you.
  `reliant workflow watch <id>` shows each node as it runs.

The run **uses model credits or your connected AI account**. The two agents are
capped at 40 and 30 turns, and the fix loop at 3 rounds (`max_attempts`, at
most 5).

### 4. Watch what happens

| Step | Node | What it does |
|---|---|---|
| Test | `tests` | Runs `go test ./...` in your checkout with `credentials: none`. If it passes, the run ends at `already_green`. If it fails, the output is saved to the thread for the agents to read. |
| Isolate | `workspace` | A `create_worktree` workspace: its own directory and branch, started from your current commit. Your checkout is never edited. |
| Fix | `attempts` → `fix-until-green` | `builtin://agent` works in the workspace. Its shell is bound to `credentials: none`, and it can't spawn sub-agents. After each round the workflow re-runs the tests itself. The agent doesn't decide when it's done. |
| Guard | `guard` | `git status --porcelain -- '*_test.go'` in the workspace. Any change to a test file stops the run (`tests_touched`) before review. |
| Review | `review` | `builtin://structured-agent` in a forked thread. It has only `view` and shell tools, re-runs the tests, and must answer `approve` or `request_changes`. A `request_changes` verdict ends the run (`review_blocked`). |
| Brief | `brief` | Posts the test result, the reviewer's verdict, the workspace path and branch, what approval will do, and the full diff. It sits right above the gate. |
| **Approve** | `approve` | **The run pauses here.** |

### 5. Approve or deny

The approval appears in the run and in your **Inbox**, titled *"Publish the
agent's fix? Tests pass and the reviewer approved."* Only a person signed in to
the Reliant app (web or desktop) can approve or deny it. The server refuses a
decision made with the CLI login or with a daemon's credential
([Who can decide an approval](https://docs.reliantlabs.io/automations/overview#who-can-decide-an-approval)).

| Decision | Ending | Run outcome |
|---|---|---|
| Approve (no `publish_repo`, the default) | `kept_local`: nothing is pushed. The fix sits uncommitted in the workspace. | success |
| Approve, with `publish_repo` set | `publish` (a `git_publish` node) pushes branch `reliant-demo/fix-<run id>`, then `published` | success |
| Deny | `denied`: nothing is pushed | success |
| No decision within `approval_timeout` (4h) | `undecided` | failure |
| Unattended (an automation started it) | `undecided`, with a note to start it by hand | failure |

Every path is gated on `status == 'approved'`, never on `!= 'denied'`. That way,
a timeout or an unattended run can never publish.

To look at the fix after approving, run `git -C <workspace path> diff`. The
`brief` and `kept_local` messages print the path. Workspaces show up in the
app's Workspaces panel.

## Inputs

| Input | Default | Meaning |
|---|---|---|
| `test_command` | `go test ./...` | The command that must exit 0 |
| `protected_pathspec` | `*_test.go` | Git pathspec the fixer must not change |
| `max_attempts` | `3` (1–5) | Most fix-and-test rounds |
| `approval_timeout` | `4h` | How long the gate waits |
| `publish_repo` | empty | `owner/name` to push the approved fix to. Leave it empty to evaluate. |
| `unattended` | `false` | Marks the run as having nobody available, so the gate can't wait. An automation's run is unattended whatever this says. |

## Limits

- **Nothing here has been run against a live account yet.** The workflow
  validates, lints clean and passes 13 scenarios on reliant 1.29.0. In those
  scenarios every agent, command, worktree and approval is mocked. We haven't
  measured how well an agent fixes this bug, or what a run costs.
- **The guard covers test files only.** The agent is told to fix code, not
  tests, and `guard` enforces that for `*_test.go`. Any other change is caught
  only by the reviewer, which is a model's judgment.
- **`credentials: none` is a full sandbox only on Linux.** On macOS and Windows
  it only cleans the environment. In both cases the command keeps network
  access and can write its checkout
  ([Daemon security model](https://docs.reliantlabs.io/deep-dives/daemon-security-model)).
- **The reviewer's tools aren't a sandbox.** It gets no edit tool, but its shell
  could still change files. Its shell is not bound to `credentials: none`.
- **Publishing is optional, and we haven't tested it here.** `git_publish`
  refuses `main`, `master`, protected branches and anything outside
  `reliant-demo/*`, and never force-pushes
  ([Publish Branch](https://docs.reliantlabs.io/reference/nodes#publish-branch)).
  Which GitHub identity a hand-started run pushes as isn't documented on the
  pages we cite, so we make no claim about it. For an evaluation, leave
  `publish_repo` empty.
- **Untracked files the agent adds** are listed in the diff the reviewer gets
  (`git status --short`), but their contents are not.

## Features used, and where they're documented

| Feature | Docs |
|---|---|
| Workflow files in `.reliant/workflows/`, scenarios in `<name>/scenarios/` | [Creating custom workflows](https://docs.reliantlabs.io/workflows/custom-workflows) |
| `run` node: `exit_code`, `timeout`, `credentials: none` | [Node reference: Run Command](https://docs.reliantlabs.io/reference/nodes#run-command) |
| `save_message` on a condition, `outcome: success / failure` | [Node reference: common fields](https://docs.reliantlabs.io/reference/nodes) |
| `create_worktree` workspace, `project.path` to run an agent in it | [Workspaces](https://docs.reliantlabs.io/features/worktrees), [Parallel agents in git worktrees](https://docs.reliantlabs.io/guides/workflows/parallel-agents-git-worktrees) |
| Do-while `loop` bounded by `iter.iteration`, routed on the test's exit code | [Loop until tests pass](https://docs.reliantlabs.io/guides/workflows/run-tests-until-they-pass) |
| Sub-workflow via `ref: project://…` with `args` | [Node reference: Agent](https://docs.reliantlabs.io/reference/nodes#agent) |
| `builtin://agent` with `max_turns`, `spawn_presets: []`, `tool_params: {shell: {credentials: none}}` | [Creating custom workflows](https://docs.reliantlabs.io/workflows/custom-workflows) |
| `builtin://structured-agent` reviewer with a response schema | [Write your first multi-agent workflow](https://docs.reliantlabs.io/guides/workflows/first-multi-agent-workflow) |
| `approval` node: statuses `approved` / `denied` / `timeout` / `unattended` | [Require human approval](https://docs.reliantlabs.io/guides/workflows/require-human-approval) |
| Automations are unattended, and only a signed-in person decides approvals | [Automations](https://docs.reliantlabs.io/automations/overview#unattended-by-design) |
| `git_publish` | [Node reference: Publish Branch](https://docs.reliantlabs.io/reference/nodes#publish-branch) |
| `on_failure:` handler | [Workflow schema](https://docs.reliantlabs.io/reference/workflow-schema) |
| `validate`, `lint`, `scenario run`, `run --exit-on-gate`, `watch` | [Workflows from the terminal](https://docs.reliantlabs.io/cli/workflows), [Lint rules](https://docs.reliantlabs.io/workflows/lint), [Testing workflows](https://docs.reliantlabs.io/workflows/testing) |

## License

This demo has no license file yet; add one before publishing it. Reliant itself is source-available under the
Business Source License 1.1. It isn't open source: each version converts to
Apache 2.0 four years after release
([License FAQ](https://docs.reliantlabs.io/license-faq)).
