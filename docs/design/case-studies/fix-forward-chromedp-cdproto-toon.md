# Fixing Forward Through Three Breaking Upgrades

d2vision takes breaking changes from its dependencies and fixes forward. It
does not hold old versions or branch for them. This is a record of one time that
rule was tested: three dependencies shipped intentional breaking changes, they
were merged together, and `main` stopped building.

## Summary

| | |
|---|---|
| **Symptom** | CI and `go test ./...` failed with `missing go.sum entry for module providing package github.com/gobwas/ws` |
| **Where it appeared** | On a pull request for an unrelated feature, which looked like the cause and was not |
| **Real cause** | Three Dependabot updates were merged one after another and left `go.mod` inconsistent |
| **Dependencies** | `chromedp` v0.16.0 to v0.21.0, `cdproto` to v0.159.0, `toon-go` to v1.0.1 |
| **Size of the fix** | One file ported (15 lines added, 13 removed), one test added, two documentation samples and one test expectation updated |
| **Outcome** | All three taken forward. `main` builds and all tests pass |

## What happened

Three Dependabot pull requests were open at once: chromedp to v0.21.0 (#28),
toon-go to v1.0.1 (#29), and cdproto to v0.158.6 (#37). Each was written against
the `go.mod` that `main` had when it was opened. They were merged one after
another, and the merges of `go.mod` and `go.sum` did not agree.

The CI results for those three pull requests are no longer available, so we
cannot say whether any of them passed on its own. What we can show is that
`main` built before the first of them and did not build after the third.

The result listed chromedp v0.16.0, the old version, beside the new cdproto, and
did not list the modules chromedp v0.16.0 needs. Nothing that imported chromedp
could build.

A pull request for a separate feature merged next, and CI failed on `main`, so
the failure showed up on that feature's merge. Building the commit just before
it showed that `main` was already broken one merge earlier. **The first question
after a failure on `main` is where it first broke, and that is quick to answer
by building the previous merge.**

## What each library changed

Each change was deliberate and documented by its maintainers.

| Dependency | Change | Evidence |
|------------|--------|----------|
| `chromedp` v0.21 | `Run` is generic and actions are typed (`Action[T]`). `ActionFunc` and `Tasks` are gone, and protocol commands are called with `cdp.Call` | It ships `docs/MIGRATION.md`, which lists every renamed and removed name from v0.16.0 to v0.21.0 in the order to apply them |
| `cdproto` v0.157.3 and later | Commands are values, not functions, and the types are the typed API. It does not work with chromedp v0.16.0 | The same migration document |
| `toon-go` v1.0 | A struct field without a `toon` tag falls back to its `json` tag, where it was the Go field name | Its README states the rule and ties it to the TOON specification |

Reading these first matters. They showed that none of the three would be
reverted, so waiting or pinning could only postpone the work.

## The two options

**Pin back.** Return the three modules to the versions the code was written
for. It worked: one commit, and the tests passed. It needed three pins,
including a cdproto pseudo-version, because the versions in `go.mod` did not
match each other. And it left a gap to close later, with a migration guide that
would only get longer.

**Fix forward.** Take all three.

| Part | Change |
|------|--------|
| `export/chrome.go` | `chromedp.Run` with several actions became `chromedp.Do`. `ActionFunc` became `chromedp.Func`. `page.GetFrameTree().Do(ctx)` became `cdp.Call(ctx, t, page.GetFrameTree, cdp.Empty{})`. `Evaluate` and `CaptureScreenshot` return values and no longer fill a pointer |
| `export/chrome_test.go` | A new test that renders an SVG in Chrome and checks the PNG |
| TOON | One test expectation, and the sample output in two documents, changed from `Version` to `version` and so on |

The forward fix was a small port and one new test. We chose it. A pin would
have been the cheaper commit and the more expensive decision.

## What the fix-forward path needed

- **Run the code, not only the tests.** The Chrome path had no test, so the
  tests passed and said nothing about it. The port was checked by capturing a
  GIF through the real browser and looking at the frame. A test now covers it
  and skips when Chrome is not installed.
- **Look at the output, not only the compile.** A port can compile and still
  draw nothing. The frame it captured, with both nodes and the arrow, is what
  showed it worked.
- **Say what changed for users.** TOON keys are now lowercase. A `.toon` file
  written with the old capitalized names generates nothing and reports no
  error. The changelog says so. The deeper fix is to fail loudly when a spec has
  no recognized keys, which is still open.

## How we take dependency updates

Dependency updates are merged to `main` as they arrive. `main` may break. The
fix is made locally and goes forward: take the new version and port the code.
We do not pin back to postpone the work, unless we know of a reason to, such as
a transitive dependency that needs the old version.

A broken `main` costs a diagnosis, and the first step is to find where it first
broke. Three things make that step and the fix cheaper:

| Need | What we do |
|------|------------|
| Find which change broke the build | Build the commit before each recent merge, as in this case, and read the upstream migration notes first |
| A code path with no test | Cover every path that reaches outside the process, such as a browser, with a test that skips when the tool is missing |
| A format change that fails silently | A consumer that cannot read its input should say so and not produce nothing |

## How the upstreams ship

None of the three keeps maintenance branches or backports fixes to old versions.
We read this from their public repositories (branches, tags, and release dates).
We did not find a written policy, so this describes what they do and not what
they have promised.

| Module | Branches | Tags | Releases |
|--------|----------|------|----------|
| `chromedp` | `main`, and a stale `v0.1` branch from 2019 | 84, latest v0.21.0, none at v1 | v0.15.1 in the spring, then v0.19.0, v0.19.1, v0.20.0, v0.20.1 and v0.21.0 within about a week |
| `cdproto` | `main` | 20, latest v0.159.0, none at v1 | Tagging starts at v0.157.0. Breaking changes ride minor and patch tags, such as v0.157.3 for the typed API and v0.157.4 for optional numbers as pointers |
| `toon-go` | `main` | 4, latest v1.0.1 | v0.1.0, v0.2.0, v1.0.0 and v1.0.1 within about a week. The key-name change is in v1.0.1 and was not in the pre-release d2vision used before |
| `kin-openapi` (for comparison, used in other projects of ours) | `master` | 156, latest v0.152.0, none at v1 | First release in 2018, still v0, and about ten minor releases a year since 2024 |

Four things stand out:

- **One line of development.** Every fix lands on `main`. A user of an old
  version is expected to move forward. Nothing here backports.
- **Versions that stay at v0.** cdproto is at v0.159.0 and kin-openapi at
  v0.152.0, with no v1 tag. Semantic versioning reserves `0.y.z` for
  development, where "anything may change at any time", so a long-lived v0 with
  many minor releases says that each minor release is where breaking changes go
  and that no stable v1 contract is promised. We do not know why these authors
  stay at v0. Go's import path rule for v2 and above may be one reason, but that
  is a guess.
- **Releases come in bursts.** chromedp shipped five in a week and toon-go four.
  A version you pin stops matching its neighbors quickly, and nobody will patch
  it.
- **Help to move.** chromedp ships `docs/MIGRATION.md`, which lists every renamed
  and removed name in the order to apply them, and a `Legacy` adapter so old
  code can move step by step. A project that breaks its users often can make the
  move cheap, and this one did.

So taking the change forward is not an outlier choice. It is how the ecosystem
d2vision depends on already works. A consumer that expects stability from a v0
module is relying on something its maintainers have not offered.

## Why this supports fix-forward

The case is a fair test of the rule, because the changes were not small. Three
libraries changed their APIs in one step, and one changed what its output looks
like. Taking them cost one port of about thirty changed lines, one test, and
some documentation. Holding old versions would have meant carrying a build that
no longer matches its dependencies, and fixes for security findings would
have to be applied to it by hand as well as to the current one.

What made fix-forward cheap was that each upstream change was documented, so
the port was mechanical. What made it safe was checking the real output. What
cost the most was the diagnosis: three updates had broken `main` together, and
the failure appeared on an unrelated merge.
