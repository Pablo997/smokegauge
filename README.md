# smokegauge

**smokegauge** is a small command-line tool that runs fast HTTP smoke checks from a declarative YAML file. Point it at a config, it performs the requests, validates status codes (and optional response rules), aggregates results, and exits with a status suitable for CI pipelines.

Use it after deploys, in GitHub Actions, or locally to prove that critical endpoints still behave as expected—without maintaining a full integration suite for every trivial probe.

---

## Why this exists

- **Fast feedback:** parallel checks with sensible timeouts and cancellation.
- **Contract-ish checks in one file:** URLs, methods, expected status, optional headers/body rules—versioned next to your infra or app repo.
- **Script-friendly:** stable exit codes and optional machine-readable output.

---

## Features

- Load checks from a YAML configuration file.
- Execute HTTP requests with per-check and global timeouts (`context`-aware).
- Run checks concurrently with a configurable concurrency limit.
- Match expected HTTP status codes; optional assertions on headers or response body (substring or regex, depending on implementation).
- Human-readable summary table and optional JSON output for logs and dashboards.
- Non-zero exit code when any check fails or the config is invalid—ideal for `&&` in shell scripts and CI gates.

---

## Installation

### From source (requires Go 1.22+)

```bash
go install github.com/Pablo997/smokegauge/cmd/smokegauge@latest
```

The binary is installed to `$GOPATH/bin` or `$(go env GOPATH)/bin`. Ensure that directory is on your `PATH`.

### Build locally

```bash
git clone https://github.com/Pablo997/smokegauge.git
cd smokegauge
go build -o smokegauge ./cmd/smokegauge
```

On Windows, use `smokegauge.exe` as the output name if you prefer.

---

## Quick start

1. Create a config file (e.g. `checks.yaml`):

```yaml
version: 1

defaults:
  timeout: 5s
  concurrency: 8

checks:
  - name: api health
    method: GET
    url: https://api.example.com/health
    want_status: 200

  - name: docs redirect
    method: GET
    url: https://example.com/docs
    want_status: 301
    want_header:
      Location: "https://example.com/docs/"
```

2. Run smokegauge:

```bash
smokegauge --file checks.yaml
```

3. In CI, fail the job on failures:

```bash
smokegauge --file checks.yaml --format text
```

---

## Configuration reference

### Top level

| Field | Type | Description |
|--------|------|-------------|
| `version` | int | Config schema version. Currently `1`. |
| `defaults` | object | Optional defaults applied to all checks unless overridden. |
| `checks` | array | List of checks to run. |

### `defaults` (optional)

| Field | Type | Description |
|--------|------|-------------|
| `timeout` | duration | Default per-request timeout (e.g. `10s`, `500ms`). |
| `concurrency` | int | Max in-flight requests across all checks. |

### Each check

| Field | Type | Required | Description |
|--------|------|----------|-------------|
| `name` | string | recommended | Short label for logs and output. |
| `method` | string | yes | HTTP method (`GET`, `POST`, …). |
| `url` | string | yes | Full URL to request. |
| `want_status` | int or list | yes | Expected HTTP status code(s). |
| `headers` | map | no | Extra request headers. |
| `body` | string | no | Raw request body (for `POST`/`PUT`). |
| `want_header` | map | no | Expected response headers (exact or prefix rules as implemented). |
| `want_body_contains` | string | no | Response body must contain this substring. |
| `want_body_regex` | string | no | Response body must match this regex. |
| `timeout` | duration | no | Override default timeout for this check only. |

Exact matching rules for headers and body may be tightened over time; treat the YAML as the source of truth documented in releases.

---

## CLI

```text
smokegauge --file <path> [flags]
```

| Flag | Description |
|------|-------------|
| `--file` | Path to the YAML configuration file (required). |
| `--format` | Output format: `text` (default) or `json`. |
| `--verbose` | Log each check start/finish and errors to stderr. |
| `--version` | Print version and exit. |

Environment variables are not required; keep secrets out of committed YAML (use CI-injected files or private overlays).

---

## Exit codes

| Code | Meaning |
|------|---------|
| `0` | All checks passed and config was valid. |
| `1` | One or more checks failed (wrong status, assertion, timeout). |
| `2` | Usage error, missing file, or invalid YAML / schema. |

---

## Project layout

```text
smokegauge/
  cmd/
    smokegauge/          # CLI entrypoint: flags, wiring, os.Exit
  internal/
    config/              # Load and validate YAML into structs
    runner/              # HTTP client, concurrency, per-check execution
  go.mod
  README.md
```

- **`cmd/`** — thin `main` packages; one folder per binary.
- **`internal/`** — libraries used only by this module; not importable by other modules at stable paths (Go enforcement).

---

## Development

```bash
go test ./...
go vet ./...
```

HTTP behavior should be covered with `net/http/httptest` so tests do not depend on the public internet.

---

## License

Specify your license in a `LICENSE` file (e.g. MIT) when you publish the repository.

---

## Author

[Pablo997](https://github.com/Pablo997) — *smokegauge* is a personal learning and tooling project.
