# smokegauge

[![CI](https://github.com/Pablo997/smokegauge/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/Pablo997/smokegauge/actions/workflows/ci.yml)
[![Go](https://img.shields.io/github/go-mod/go-version/Pablo997/smokegauge?branch=main)](https://go.dev/)
[![License: MIT](https://img.shields.io/github/license/Pablo997/smokegauge?branch=main)](LICENSE)

**smokegauge** is a small CLI that runs HTTP smoke checks from a YAML file. It loads the config, runs probes in parallel (with a concurrency limit), compares status codes, and exits with codes suitable for shell scripts and CI.

---

## Features

- YAML config with schema version, defaults (`timeout`, `concurrency`, `max_body_bytes`), and a list of checks
- Concurrent execution bounded by `defaults.concurrency`
- Per-request timeout from `defaults.timeout` (Go duration string, e.g. `5s`)
- Optional response body substring checks with `want_body_contains`
- Human-readable failures on **stderr** (`-format text`, default)
- Machine-readable report on **stdout** (`-format json`)
- Stable exit codes: `0` success, `1` check failure, `2` config/usage/encode errors

---

## Requirements

- Go **1.25+** (see `go.mod`)

---

## Installation

### Build from source

```bash
git clone https://github.com/Pablo997/smokegauge.git
cd smokegauge
go build -o smokegauge ./cmd/smokegauge
```

On Windows, `smokegauge.exe` is fine.

### Install with Go

```bash
go install github.com/Pablo997/smokegauge/cmd/smokegauge@latest
```

Ensure `$(go env GOPATH)/bin` is on your `PATH`.

---

## Quick start

A sample config lives in [`testdata/checks.yaml`](testdata/checks.yaml):

```yaml
version: 1

defaults:
  timeout: 5s
  concurrency: 4
  max_body_bytes: 1000

checks:
  - name: example health
    method: GET
    url: https://example.com/
    want_status: 200
    want_body_contains: "ok"
```

Run checks:

```bash
go run ./cmd/smokegauge -file testdata/checks.yaml
```

JSON output (stdout; errors still go to stderr):

```bash
go run ./cmd/smokegauge -file testdata/checks.yaml -format json
```

Example JSON shape:

```json
{
  "ok": true,
  "failures": []
}
```

On failure, `ok` is `false` and `failures` lists each failed check with `name`, `error` (transport/body message, or empty for status-only failures), `status_code`, and `want_status`.

---

## CLI

```text
smokegauge -file <path> [-format text|json]
smokegauge -version
```

| Flag | Default | Description |
|------|---------|-------------|
| `-file` | *(required)* | Path to the checks YAML file |
| `-format` | `text` | `text`: human messages on stderr; `json`: report on stdout |
| `-version` | `false` | Print the smokegauge version and exit |

The standard library `flag` package is used; `-file` and `--file` both work. `-version` does not require `-file`.

---

## Configuration

### Top level

| Field | Type | Description |
|--------|------|-------------|
| `version` | int | Config schema version. Only **`1`** is supported. |
| `defaults` | object | Default timeout, concurrency, and body read limit for all checks |
| `checks` | array | HTTP checks to run (at least one) |

### `defaults`

| Field | Type | Description |
|--------|------|-------------|
| `timeout` | string | Go duration for each request (e.g. `5s`, `500ms`) |
| `concurrency` | int | Maximum concurrent requests (must be **> 0**) |
| `max_body_bytes` | int | Maximum response body bytes read when checking `want_body_contains` (defaults to **1 MiB**) |

### Each check

| Field | Type | Description |
|--------|------|-------------|
| `name` | string | Label in logs / JSON (recommended) |
| `method` | string | HTTP method (e.g. `GET`) |
| `url` | string | Full URL |
| `want_status` | int | Expected HTTP status code |
| `want_body_contains` | string | Optional substring expected in the response body |

---

## Exit codes

| Code | Meaning |
|------|---------|
| `0` | All checks passed |
| `1` | One or more checks failed (network error, unexpected status, or body mismatch) |
| `2` | Invalid flags, missing file, invalid YAML, config validation failed, or JSON encoding failed |

In **text** mode, transport errors are printed before status mismatch messages, and status mismatches take precedence over body mismatch messages.

---

## Project layout

```text
smokegauge/
  cmd/smokegauge/     # main, flags
  internal/
    config/           # YAML model and validation
    runner/           # HTTP execution and concurrency
    logger/           # stderr / stdout formatting
  testdata/           # sample checks.yaml
  .github/workflows/  # CI (test, vet, build)
```

---

## Development

```bash
go test ./...
go vet ./...
gofmt -l .   # should print nothing
```

CI runs on push/PR to `main` or `master` (see [`.github/workflows/ci.yml`](.github/workflows/ci.yml)).

---

## License

MIT — see [LICENSE](LICENSE).

---

## Author

[Pablo997](https://github.com/Pablo997)
