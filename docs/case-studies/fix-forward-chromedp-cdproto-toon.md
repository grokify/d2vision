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
| **Real cause** | Three Dependabot updates, each fine alone, were merged one after another and left `go.mod` inconsistent |
| **Dependencies** | `chromedp` v0.16.0 to v0.21.0, `cdproto` to v0.159.0, `toon-go` to v1.0.1 |
| **Size of the fix** | One file ported (15 lines added, 13 removed), one test added, two documentation samples and one test expectation updated |
| **Outcome** | All three taken forward. `main` builds and all tests pass |

## What happened

Three Dependabot pull requests were open at once: chromedp to v0.21.0 (#28),
toon-go to v1.0.1 (#29), and cdproto to v0.158.6 (#37). Each was written against
the `go.mod` that `main` had when it was opened. They were merged one after
another, and the merges of `go.mod` and `go.sum` did not agree.

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

## What we would change in the process

| Gap | Change |
|-----|--------|
| Three updates merged together without being rebuilt | Group updates that must move together (chromedp with cdproto), and require a pull request to be up to date with `main` before it merges |
| `main` was broken by merges that each passed alone | Run the full build and tests after every merge to `main`, and treat a failure there as the top priority |
| A code path with no test | Cover every path that reaches outside the process, such as a browser, with a test that skips when the tool is missing |
| A format change that fails silently | A consumer that cannot read its input should say so and not produce nothing |

## Why this supports fix-forward

The case is a fair test of the rule, because the changes were not small. Three
libraries changed their APIs in one step, and one changed what its output looks
like. Taking them cost one port of about thirty changed lines, one test, and
some documentation. Holding old versions would have meant carrying a build that
no longer matches its dependencies, and fixes for security findings would
have to be applied to it by hand as well as to the current one.

What made fix-forward cheap was that each upstream change was documented, so
the port was mechanical. What made it safe was checking the real output. What
nearly made it expensive was the process gap: that three updates could break
`main` together without anyone being told which one to read.
