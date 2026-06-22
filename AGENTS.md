# AGENTS.md

Guidance for AI coding agents working in this repository. Humans may find it useful too.

## What this project is

`pull-request-code-coverage` is a CI plugin (a Go CLI shipped as a Docker image) that reports code
coverage for **only the lines changed in a pull request** — not whole files or the whole repo. It
reads a unified diff plus a coverage report, works out which changed lines your tests executed, and
writes the result to the CI console and (optionally) as a GitHub PR comment.

It is **format-driven, not language-driven**: support for a language is really support for that
language's coverage report format.

| `coverage_type` | Language(s) | Format | Loader package |
|---|---|---|---|
| `jacoco` | JVM (Java/Kotlin/Scala) | JaCoCo XML | `internal/plugin/coverage/jacoco` |
| `cobertura` | Go | Cobertura XML | `internal/plugin/coverage/cobertura` |
| `python` | Python | coverage.py XML | `internal/plugin/coverage/pythoncov` |
| `lcov` (aliases `javascript`, `typescript`) | JS/TS | LCOV `lcov.info` | `internal/plugin/coverage/lcov` |

## Tech stack

- **Go 1.26.3**, module `github.com/target/pull-request-code-coverage`.
- Deps: `pkg/errors`, `sirupsen/logrus`, `stretchr/testify` (see `go.mod`). Keep deps minimal.
- Ships as a Docker image (`Dockerfile`) published to `ghcr.io/target/pull-request-code-coverage`.

## Commands (run from repo root)

```bash
go build ./...        # compile everything
go test ./...         # run all tests (do this before declaring done)
go test ./internal/plugin/coverage/lcov/...   # run one package
go vet ./...          # vet

make format           # CHECK gofmt only — does NOT modify files; fails if anything is unformatted
gofmt -w .            # actually auto-format (use this to fix what `make format` flags)

make lint             # golangci-lint (downloads pinned v2.12.2 into ./bin on first run)
```

Before finishing any change, all of these must pass: `go test ./...`, `make format`, `make lint`.
The CI (`.github/workflows/test.yml`) runs build, test, `make format`, and `make lint` on every PR.

## Architecture & data flow

Entry point `main.go` → `plugin.NewRunner().Run(os.LookupEnv, os.Stdin, os.Stdout)`.
The runner (`internal/plugin/runner.go`) reads **config from env vars** (`PARAMETER_*`,
`BUILD_PULL_REQUEST_NUMBER`, `REPOSITORY_ORG`, `REPOSITORY_NAME`), the **diff** (from stdin by
default, or fetched from the GitHub API — see `PARAMETER_DIFF_SOURCE` below), and the
**coverage report from the file** at `PARAMETER_COVERAGE_FILE`.

```
diff (unified) ──► sourcelines/unifieddiff ──► []domain.SourceLine
                                                          │  {Module,SrcDir,Pkg,FileName,LineNumber,LineValue}
coverage file ──► coverage.Loader.Load() ──► coverage.Report
                                                          │
        calculator.DetermineCoverage(lines, report) ─────┘
            └ for each line: report.GetCoverageData(...) ──► []domain.SourceLineCoverage
                                                          │
                          reporter.Forking{ Simple, GithubPullRequest, StepSummary }.Write(...)
                              ├─ Simple          → plain-text report to stdout (always)
                              ├─ GithubPullRequest → Markdown PR comment (only if creds present)
                              └─ StepSummary     → Markdown to $GITHUB_STEP_SUMMARY (only if set)
```

Key packages:
- `internal/plugin/sourcelines/unifieddiff/changed_source_loader.go` — parses the unified diff into
  changed `SourceLine`s. `PARAMETER_SOURCE_DIRS` controls how a path prefix is split into `SrcDir`/`Pkg`.
  Handles both `--unified=0` diffs (the stdin/Vela path, no context lines) and diffs that carry context
  lines (e.g. from the GitHub API) — context lines advance the new-file line counter but aren't recorded.
- `internal/plugin/githubdiff/diff.go` — alternative diff source. When `PARAMETER_DIFF_SOURCE=github`,
  the runner fetches the PR diff from `GET /repos/{owner}/{repo}/pulls/{n}` with the
  `application/vnd.github.v3.diff` media type instead of reading stdin. Default is `stdin`
  (unchanged behavior). The `github` mode requires `PARAMETER_GH_API_KEY` + the three build-context vars.
- `internal/plugin/coverage/` — `report.go` defines the two interfaces every format implements:
  `Loader.Load(file) (Report, error)` and `Report.GetCoverageData(module, sourceDir, pkg, fileName, lineNumber) (*CoverageData, bool)`.
- `internal/plugin/calculator/calculator.go` — joins changed lines to coverage data.
- `internal/plugin/reporter/` — `simple.go` (console), `github_pr.go` (PR comment markdown),
  `step_summary.go` (writes the same markdown to `$GITHUB_STEP_SUMMARY`), `forking.go` (runs all
  reporters), `utils.go` (`filePath`, `lineDescription`). Per-file aggregation
  (`collectFileCoverage`), `coverageStatusEmoji`, and the shared markdown builder
  (`buildMarkdownReport`, used by both `github_pr.go` and `step_summary.go`) live in `github_pr.go`.
  The PR comment is **sticky**: `github_pr.go` first GETs the PR's comments, and if it finds the one
  carrying the hidden `commentMarker` it PATCHes that comment instead of POSTing a new one.
- `internal/plugin/domain/domain.go` — core types. Coverage is counted in **instructions**
  (`CoveredInstructionCount`/`MissedInstructionCount`), not lines (see below).

## Concepts you must not "simplify" away

- **Lines vs. instructions are deliberately different.** For JaCoCo, one source line maps to several
  JVM bytecode *instructions*, so a line can be partly covered. For Go/Python/LCOV the loaders emit
  exactly 1 instruction per line. The reports surface both units on purpose — do not "fix" this as if
  it were a bug. The user has explicitly asked for this distinction to be clear.
- **Two output formats, one dataset.** `Simple` (plain text, stdout) renders one way;
  `GithubPullRequest` and `StepSummary` both render the shared `buildMarkdownReport` output. Change
  `Simple` and `buildMarkdownReport` if you change what's reported.
- **The PR comment is posted only** when `gh_api_key` AND `BUILD_PULL_REQUEST_NUMBER` AND
  `REPOSITORY_ORG` AND `REPOSITORY_NAME` are all present; otherwise console-only. `GithubPullRequest`
  also returns early when there are zero changed lines with coverage data.
- **Path matching differs by loader.** `jacoco`/`cobertura` match using the report's source root;
  `python`/`lcov` match on the **repo-relative path** (and `lcov` also suffix-matches absolute
  `SF:` paths). When adding/altering a loader, preserve its matching contract.

## How to add a new coverage format

1. Create `internal/plugin/coverage/<name>/report.go` implementing `coverage.Loader` and
   `coverage.Report`. Mirror an existing loader; per-loader helpers are duplicated by convention
   (each loader has its own `silentlyCall`, path helpers, etc. — that's the established style here,
   not shared utilities).
2. Add a `case` in `getCoverageReportLoader` in `internal/plugin/runner.go` (and the import).
3. Add a fixture under `internal/test/` (e.g. `example_<name>.<ext>`) and a matching diff fixture.
4. Add a loader unit test `report_test.go` and a full end-to-end test in
   `internal/plugin/runner_test.go` (see golden-test note below).
5. Update `README.md`: the supported-formats table, a usage section, and the `coverage_type`
   parameter values.

## Testing conventions (read before editing reporter output)

`internal/plugin/runner_test.go` contains **golden-string assertions** for the exact console output
(`buf.String()`) and the exact PR comment body. If you change anything the reporters print, these
goldens will fail and must be regenerated **exactly** — do not hand-edit them (emoji, box-drawing
chars, tabs, and trailing spaces all matter).

To regenerate, write a temporary `internal/plugin/dump_test.go` that runs `NewRunner().Run(...)` for
the affected scenarios against an `httptest` server, and prints `strconv.Quote(output)` for both the
console buffer and the captured request body. Copy those quoted strings into the goldens, then
**delete the dump test**. (This pattern has been used repeatedly here; it is the reliable way to keep
goldens byte-exact.)

Other test notes:
- Mocks live in `internal/test/mocks` (`MockPropertyGetter`, `WithMockGithubAPI`). `propGetter`
  uses testify mock with `AssertExpectations`, so only stub the env vars the runner actually reads
  for that scenario.
- `_test.go` files are exempt from `funlen`, `goconst`, `gosec` (see `.golangci.yml`).

## Conventions & gotchas

- **Lint is strict** (`.golangci.yml`: `gocyclo`, `gosec`, `unused`, `staticcheck`, `errcheck`,
  `unparam`, `goconst`, etc.). Notably: prefer `fmt.Fprintf(&b, ...)` over
  `b.WriteString(fmt.Sprintf(...))`; remove dead code (`unused`); file opens with user-supplied
  paths need a `// nolint: gosec` with a reason (see existing loaders).
- `make format` only *checks*; run `gofmt -w .` to actually format.
- This plugin **dogfoods itself**: `.github/workflows/pr-coverage.yml` runs it on this repo's own
  PRs, so reporter changes will visibly change the coverage comment on your PR. That's expected.
- `release.yml` builds and pushes the Docker image to GHCR on GitHub release (uses the built-in
  `GITHUB_TOKEN`, no extra secrets).

## Commits & PRs

- **Do not add AI/agent co-author trailers or "Generated with…" lines to commits.** Commits are
  attributed solely to the repository author. Write a normal, descriptive commit message.
- Only commit/push when explicitly asked. If on the default branch, create a feature branch first.
- A PR is ready only when `go test ./...`, `make format`, and `make lint` all pass.
