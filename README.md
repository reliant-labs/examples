# Reliant examples

Small, runnable examples of [Forge](https://github.com/reliant-labs/forge) and
[Reliant](https://reliantlabs.io). Each one lives in its own directory with its
own README, and each keeps the commit history it was built with, so you can
read it step by step with `git log -- <directory>`.

| Example | What it shows | Needs |
| --- | --- | --- |
| [`forge-notes/`](forge-notes/) | A full-stack Forge app (Go Connect RPC API, Postgres, Next.js frontend) built one commit at a time, and two Forge checks catching real mistakes: an unsafe migration, and a hand edit to a generated file | Go, Node 22.12+, Docker. No account. |
| [`reliant-fix-review-approve/`](reliant-fix-review-approve/) | A Reliant workflow: an agent fixes a failing test in its own git workspace, a second agent reviews the diff, and the run waits for a person to approve before anything is published | Go 1.22+ and the `reliant` CLI. A Reliant account only for a live run. |

## forge-notes

The history is the point. One commit runs `forge project new`, one declares the
`Note` entity (6 hand-written lines), one runs `forge scaffold`, and the rest
show what Forge's checks refuse:

| Ref | What happens |
| --- | --- |
| tag `forge-notes/unsafe-migration` | `ALTER TABLE notes ADD COLUMN pinned BOOLEAN NOT NULL;`, which `forge lint --migration-safety` rejects (exit 1) |
| tag `forge-notes/safe-migration` | the fix (`... NOT NULL DEFAULT false`) and the regenerated ORM; the lint passes |
| branch `forge-notes/hand-edit-generated` | a hand edit to `internal/db/note_orm_gen.go`, which `forge lint --generated-drift` rejects (exit 1) |

```bash
git clone https://github.com/reliant-labs/examples && cd examples
git checkout forge-notes/unsafe-migration
cd forge-notes            # Forge commands run here, where forge.yaml is
npx -y @reliantlabs/forge@1.11.0 lint --migration-safety     # exit 1
```

To run the app locally: `npx -y @reliantlabs/forge@1.11.0 env up dev` from
`forge-notes/` on `main`. The full walkthrough, with the real command output,
is in [`forge-notes/README.md`](forge-notes/README.md).

**The notes app has no sign-in.** Anyone who can reach it can create and edit
notes. Don't deploy it anywhere public without adding auth first
([Forge auth](https://docs.reliantlabs.io/forge/auth)).

## reliant-fix-review-approve

A small Go cart pricer with a test suite, a script that commits an off-by-one
bug, and the workflow `fix-review-approve` in `.reliant/workflows/`:

1. run the tests;
2. an agent fixes the failure in its own git workspace, with the workflow
   re-running the tests after each round (at most 3);
3. a guard stops the run if any test file changed, then a second agent reviews
   the diff;
4. the run pauses for your approval. By default nothing is pushed even after
   you approve.

The workflow expects to be at the root of a git repository, so copy the
directory out before you try it:

```bash
cp -r reliant-fix-review-approve ~/fix-review-approve && cd ~/fix-review-approve
git init -q -b main && git add -A && git commit -qm "Start from reliant-labs/examples"
reliant workflow validate && reliant workflow lint && reliant workflow scenario run
```

Those three commands need no account and call no model. The walkthrough,
including a live run and its limits, is in
[`reliant-fix-review-approve/README.md`](reliant-fix-review-approve/README.md).

## See a deployed Forge app's console

[demo.reliantlabs.io](https://demo.reliantlabs.io) is the console of a sample
Forge app (a shop, with made-up, read-only data): releases and the health check
on each one, errors grouped into issues, traces, dashboards, product analytics
and session replays, plus its data, storage, auth and security pages. An app
you deploy to Reliant hosting gets the same console. The demo records usage
and session replays; its banner links the privacy policy.

Reliant and Forge's own control plane are themselves Forge apps.

## Links

- Docs: https://docs.reliantlabs.io
- Forge quick start: https://docs.reliantlabs.io/forge/quick-start
- Forge source: https://github.com/reliant-labs/forge
- Writing Reliant workflows: https://docs.reliantlabs.io/workflows/custom-workflows
- Console demo: https://demo.reliantlabs.io

## License

The examples in this repository are licensed under the
[Apache License 2.0](LICENSE). Forge is also Apache-2.0. Reliant itself is
source-available under the Business Source License 1.1, not open source; each
version converts to Apache 2.0 four years after release
([License FAQ](https://docs.reliantlabs.io/license-faq)).
