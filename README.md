# pull-request-code-coverage

A CI plugin that reports code coverage for **only the lines changed in a pull request** — not the whole file, not the whole repo.

When you're working to raise a repo's coverage, a whole-repo percentage doesn't tell you whether *your* change is tested. This plugin looks at just the lines your PR adds or edits and reports coverage for those lines, so a reviewer can immediately see whether the new code is covered.

It supports **JVM, Go, and Python** projects, and works out of the box for [Vela](https://github.com/go-vela) (Target's open-source CI/CD) as well as any CI that can run a Docker container (e.g. GitHub Actions).

---

## Contents

- [How it works](#how-it-works)
- [Lines vs. instructions](#lines-vs-instructions)
- [What you'll see](#what-youll-see)
  - [In the CI/CD console](#in-the-cicd-console)
  - [As a pull-request comment](#as-a-pull-request-comment)
- [Supported coverage formats](#supported-coverage-formats)
- [Usage](#usage)
  - [Docker image](#docker-image)
  - [JVM projects (jacoco)](#jvm-projects-jacoco)
  - [Go projects (cobertura)](#go-projects-cobertura)
  - [Python projects (python)](#python-projects-python)
  - [JavaScript / TypeScript projects (lcov)](#javascript--typescript-projects-lcov)
  - [Running outside Vela](#running-outside-vela)
- [Parameters](#parameters)
- [Development](#development)
- [License](#license)

---

## How it works

1. It reads the PR's unified diff to find the lines you changed.
2. It reads your coverage report (JaCoCo / Cobertura / coverage.py).
3. For each changed line, it checks whether your tests executed it.
4. It reports the result in two places: the **CI/CD console** (always) and a **pull-request comment** (when GitHub credentials are provided).

---

## Lines vs. instructions

The report uses two different units, and they are **not** the same thing:

- **Lines** — the source lines your PR changed.
- **Instructions** — the smaller executable units the coverage tool counts *inside* those lines.

For **JaCoCo (JVM)**, a single source line compiles to several JVM bytecode instructions, so one line can be partly covered — e.g. `8 covered / 3 missed` instructions spread across only `2` measurable lines. For **Go (cobertura)** and **Python (coverage.py)**, the plugin counts one instruction per line, so there the two numbers line up.

Every changed line falls into one of these buckets:

| Bucket | Meaning |
|---|---|
| 🟢 Covered instructions | changed code your tests executed |
| 🔴 Missed instructions | changed code your tests never ran |
| 📈 Tracked changed lines | changed lines the coverage tool could measure |
| ⚪ Untracked changed lines | changed lines with no coverage data: comments, blanks, declarations |

**Diff coverage** is the headline number: `covered ÷ (covered + missed)` instructions.

---

## What you'll see

The same data is rendered for two audiences: plain text for the build log, and Markdown for the PR.

### In the CI/CD console

Every run prints a report to the step's console (stdout):

```
──────────────────────────────────────────────────────────────
 📊 Patch Coverage Report — 73% 🟡 — changed lines only
──────────────────────────────────────────────────────────────
 Modules: category-search

 Diff coverage: 73% 🟡  —  8 of 11 changed instructions covered

 Summary
   Covered instructions       73%  (8)
   Missed instructions        27%  (3)
   Tracked changed lines      22%  (2)
   Untracked changed lines    78%  (7)

 Note: "lines" are the source lines you changed; "instructions" are the
 executable units the coverage tool counts inside them (one line can hold
 several, e.g. JaCoCo bytecode), so the two counts differ.

 Coverage by file  (lowest coverage first)
    73%     8 cov /   3 miss   category-search/src/main/java/com/tgt/CategorySearchApplication.java
   (3 file(s) with no measurable lines omitted)

 Uncovered lines (1)
   - category-search/src/main/java/com/tgt/CategorySearchApplication.java:52
             System.out.print("Something");

──────────────────────────────────────────────────────────────
```

What each part shows:

- **Diff coverage** — the headline: how much of your changed, *measurable* code ran.
- **Summary** — the four buckets from [Lines vs. instructions](#lines-vs-instructions).
- **Coverage by file** — per-file diff coverage, **lowest first** so the riskiest files surface at the top. Files whose only changed lines aren't measurable (config, docs, tests) are collapsed into a count.
- **Uncovered lines** — each changed line your tests never ran, with the source line.

### As a pull-request comment

When GitHub credentials are set, the same report is posted as a PR comment. It renders like this:

> ### 🛡️ Patch Coverage Report — `73%` 🟡
>
> Scope: **changed lines only** — the code this PR adds or edits, not whole files or the repo. It answers one thing — *did your tests run the code you just touched?*
>
> *Modules:* category-search
>
> **Diff coverage:** `73%` 🟡 — `8` of `11` changed instructions covered
>
> | Metric | Value | |
> | :-- | --: | :-- |
> | 🟢 Covered instructions | `8` (73%) | changed code your tests executed |
> | 🔴 Missed instructions | `3` (27%) | changed code your tests never ran |
> | 📈 Tracked changed lines | `2` (22%) | lines the coverage tool could measure |
> | ⚪ Untracked changed lines | `7` (78%) | comments, blanks, declarations |
>
> **Coverage by file**
>
> | File | Diff coverage | Covered / Missed |
> | :-- | :-: | :-: |
> | `category-search/src/main/java/com/tgt/CategorySearchApplication.java` | 🟡 73% | 8 / 3 |
>
> <details><summary>🔍 Uncovered lines (1)</summary>
>
> `category-search/.../CategorySearchApplication.java:52` → `System.out.print("Something");`
> </details>

It carries the same sections as the console — diff-coverage headline, summary table, per-file breakdown, and a collapsible list of uncovered lines.

> **Note:** the PR comment is posted only when `gh_api_key`, the PR number, the org, and the repo name are all available. Without them the plugin still prints the console report.

---

## Supported coverage formats

| `coverage_type` | Language(s) | Report format |
|---|---|---|
| `jacoco` | Java, Kotlin, Scala (JVM) | JaCoCo XML |
| `cobertura` | Go | Cobertura XML via [gocov-xml](https://github.com/AlekSi/gocov-xml) |
| `python` | Python | coverage.py XML (`coverage xml` / pytest-cov `--cov-report=xml`) |
| `lcov` | JavaScript, TypeScript | LCOV `lcov.info` from Jest / nyc / Vitest / c8 (aliases: `javascript`, `typescript`) |

---

## Usage

### Docker image

The plugin runs as a Docker container, published to the GitHub Container Registry (GHCR) on every release:

```
ghcr.io/target/pull-request-code-coverage:latest
ghcr.io/target/pull-request-code-coverage:<release-tag>
```

```
docker pull ghcr.io/target/pull-request-code-coverage:latest
```

The examples below use [Vela](https://github.com/go-vela) step syntax. See [Running outside Vela](#running-outside-vela) for other CIs.

### JVM projects (jacoco)

You need a JaCoCo XML report as input. Generating it (via Gradle/Maven) is outside the scope of this project. Once you have it, pass its path to `coverage_file`:

```yaml
- name: check-pr-code-coverage
   image: ghcr.io/target/pull-request-code-coverage:latest
   pull: true
   ruleset:
     event: [pull_request]
   parameters:
     coverage_type: jacoco
     coverage_file: some-sub-module/build/reports/jacoco/test/jacocoTestReport.xml
     source_dirs:
       - src/main/java
       - src/main/kotlin
     # omit for public github.com (defaults to https://api.github.com)
     # for GitHub Enterprise, use the full API root including /api/v3
     gh_api_base_url: https://git.target.com/api/v3
     module: some-sub-module
   secrets:
     - source: pull_request_api_key
       target: plugin_gh_api_key
```

### Go projects (cobertura)

Use [gocov-xml](https://github.com/AlekSi/gocov-xml) to convert Go's coverage profile to Cobertura XML:

```
 - go get github.com/axw/gocov/gocov
 - go get github.com/AlekSi/gocov-xml
 - go test -v -coverpkg=./... -coverprofile=coverage.txt ./...
 - go tool cover -func=coverage.txt
 - gocov convert coverage.txt | gocov-xml > ./coverage.xml
```

Then pass `coverage.xml` to the plugin:

```yaml
- name: check-pr-code-coverage
   image: ghcr.io/target/pull-request-code-coverage:latest
   pull: true
   ruleset:
     event: [pull_request]
   parameters:
     coverage_type: cobertura
     # coverage.xml generated in the step above
     coverage_file: coverage.xml
     source_dirs:
       - /vela/src/github.com/targetOSS/pull-request-code-coverage
     gh_api_base_url: https://git.target.com/api/v3
   secrets:
     - source: pull_request_api_key
       target: plugin_gh_api_key
```

> For `cobertura`, `source_dirs` must match the `<source>` path in the generated XML (the directory the tests ran in), and only a single source dir is supported.

### Python projects (python)

Generate the XML report with [coverage.py](https://coverage.readthedocs.io) or pytest-cov:

```
 # with coverage.py
 - coverage run -m pytest
 - coverage xml          # writes coverage.xml

 # or directly with pytest-cov
 - pytest --cov=myapp --cov-report=xml
```

Then pass `coverage.xml` with `coverage_type: python`:

```yaml
- name: check-pr-code-coverage
   image: ghcr.io/target/pull-request-code-coverage:latest
   pull: true
   ruleset:
     event: [pull_request]
   parameters:
     coverage_type: python
     # coverage.xml generated in the step above
     coverage_file: coverage.xml
     source_dirs:
       # repo root; use e.g. "src" if your package lives under src/
       - .
     gh_api_base_url: https://git.target.com/api/v3
   secrets:
     - source: pull_request_api_key
       target: plugin_gh_api_key
```

> Unlike `cobertura`, the `python` type matches files by their **repo-relative path**, so `source_dirs` does **not** need to be an absolute build path. Run from the repo root with `source_dirs: ["."]`, or set it to your source folder (e.g. `src`) if your code lives under one.

### JavaScript / TypeScript projects (lcov)

Most JS/TS coverage tools (Jest, nyc, Vitest, c8 — all built on Istanbul) emit an `lcov.info` file. Generate it with the `lcov` reporter:

```
 # Jest
 - jest --coverage --coverageReporters=lcov
 # nyc
 - nyc --reporter=lcov npm test
 # Vitest
 - vitest run --coverage --coverage.reporter=lcov
```

Then pass the report (commonly `coverage/lcov.info`) with `coverage_type: lcov`:

```yaml
- name: check-pr-code-coverage
   image: ghcr.io/target/pull-request-code-coverage:latest
   pull: true
   ruleset:
     event: [pull_request]
   parameters:
     coverage_type: lcov            # aliases: javascript, typescript
     # lcov.info generated in the step above
     coverage_file: coverage/lcov.info
     source_dirs:
       # repo root; use e.g. "src" if your code lives under src/
       - .
     gh_api_base_url: https://git.target.com/api/v3
   secrets:
     - source: pull_request_api_key
       target: plugin_gh_api_key
```

> Like `python`, the `lcov` type matches files by their **repo-relative path**, and it also handles the absolute `SF:` paths Istanbul commonly writes (e.g. `/home/runner/work/app/app/src/x.ts`) by suffix-matching. Set `source_dirs` to `.` (repo root) or to your source folder.

### Running outside Vela

On other CIs (e.g. GitHub Actions), run the same image and pass the inputs as environment variables instead of Vela `parameters:`. Each parameter maps to a `PARAMETER_<NAME>` env var, and the build context maps to `BUILD_PULL_REQUEST_NUMBER`, `REPOSITORY_ORG`, and `REPOSITORY_NAME` (see the table below). Pipe the PR's unified diff to the container on stdin:

```
git --no-pager diff --unified=0 "origin/$BASE_REF" -- '*.go' | docker run --rm -i \
  -e PARAMETER_COVERAGE_TYPE -e PARAMETER_COVERAGE_FILE -e PARAMETER_SOURCE_DIRS \
  -e PARAMETER_GH_API_KEY -e BUILD_PULL_REQUEST_NUMBER -e REPOSITORY_ORG -e REPOSITORY_NAME \
  ghcr.io/target/pull-request-code-coverage:latest
```

A working GitHub Actions example lives in [`.github/workflows/pr-coverage.yml`](.github/workflows/pr-coverage.yml).

---

## Parameters

**Plugin inputs** — set via Vela `parameters:` / `secrets:`, or as `PARAMETER_*` env vars on other CIs.

| Parameter | Env var | Required | Default | Description |
|---|---|---|---|---|
| `coverage_type` | `PARAMETER_COVERAGE_TYPE` | yes | | coverage format: `jacoco`, `cobertura`, `python`, or `lcov` (aliases `javascript`/`typescript`) |
| `coverage_file` | `PARAMETER_COVERAGE_FILE` | yes | | path to the coverage report, relative to the working dir |
| `source_dirs` | `PARAMETER_SOURCE_DIRS` | yes | | array of source dirs, relative to the working dir (see per-language notes above) |
| `module` | `PARAMETER_MODULE` | no | _(empty)_ | sub-module path prefix to strip, for multi-module projects (e.g. a Gradle multi-project build) |
| `gh_api_key` | `PARAMETER_GH_API_KEY` (or `PLUGIN_GH_API_KEY`) | no | | token used to post the PR comment. If unset, no comment is posted (console only) |
| `gh_api_base_url` | `PARAMETER_GH_API_BASE_URL` | no | `https://api.github.com` | GitHub API root. For GitHub Enterprise, use the full root including `/api/v3` |
| `enabled` | `PARAMETER_ENABLED` | no | `true` | master on/off switch. Set to `false` to make the plugin do nothing and exit 0, so it can stay wired into a pipeline and be toggled off without removing the step. Any other value (or unset) means enabled |
| `min_coverage` | `PARAMETER_MIN_COVERAGE` | no | _(empty)_ | diff-coverage gate: a percentage (e.g. `80`). When set, the step **fails** (exit 1) if diff coverage falls below it. Leave unset/empty to disable the gate (report only). A PR that changed no measurable lines counts as 100% and never fails |
| `debug` | `PARAMETER_DEBUG` | no | `false` | enable debug logging |

**Build context** — provided automatically by Vela; set these yourself on other CIs to enable the PR comment.

| Env var | Description |
|---|---|
| `BUILD_PULL_REQUEST_NUMBER` | the PR number to comment on |
| `REPOSITORY_ORG` | repository owner / org |
| `REPOSITORY_NAME` | repository name |

> The PR comment is posted only when `gh_api_key` **and** all three build-context values are present. Otherwise the plugin prints to the console and exits successfully.

---

## Development

This project needs go (>= 1.26.3) installed. Before submitting a PR, run:

* `make format`
* `make lint`

---

## License

This project is licensed under the Apache License, Version 2.0.
