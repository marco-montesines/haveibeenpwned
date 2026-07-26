# Project Status

_Last updated: 2026-07-26._

**Current release: [v1.0.4](https://github.com/marco-montesines/haveibeenpwned/releases)** — stable, actively maintained, low-churn by design.

## Health

| Signal | State |
| ------ | ----- |
| CI (build, vet, gofmt, race-enabled tests) | [![CI](https://github.com/marco-montesines/haveibeenpwned/actions/workflows/ci.yml/badge.svg)](https://github.com/marco-montesines/haveibeenpwned/actions/workflows/ci.yml) |
| Container image publish | [![Docker](https://github.com/marco-montesines/haveibeenpwned/actions/workflows/docker.yml/badge.svg)](https://github.com/marco-montesines/haveibeenpwned/actions/workflows/docker.yml) |
| CodeQL static analysis | [![CodeQL](https://github.com/marco-montesines/haveibeenpwned/actions/workflows/codeql.yml/badge.svg)](https://github.com/marco-montesines/haveibeenpwned/actions/workflows/codeql.yml) |
| API docs | [pkg.go.dev](https://pkg.go.dev/github.com/marco-montesines/haveibeenpwned) (auto-generated from source) |
| Vulnerability scanning | `govulncheck` on every push/PR; Dependabot monthly, grouped per ecosystem |
| Secret scanning | gitleaks over full git history on every push/PR |

## What works today

All four delivery forms are functional and released:

- **Go library** — full client with functional options, typed errors, context support.
- **CLI** — `breaches`, `breach`, `account`, `pastes`, `dataclasses`, `password`, `serve`.
- **HTTP JSON API** — 7 routes, published as a multi-arch (amd64/arm64) image on GHCR.
- **FrankenPHP extension** — `hibp_pwned_password_count()` and `hibp_breaches()` as native PHP functions.

## HIBP API v3 endpoint coverage

| HIBP endpoint | Library method | Status |
| ------------- | -------------- | :----: |
| `range/{first5HashChars}` (Pwned Passwords) | `PwnedPasswordCount` | ✅ |
| `breaches` | `GetBreaches` | ✅ |
| `breach/{name}` | `GetBreachedSite` | ✅ |
| `dataclasses` | `GetDataClasses` | ✅ |
| `breachedaccount/{account}` | `GetBreachedAccount` | ✅ |
| `pasteaccount/{account}` | `GetPastedAccount` | ✅ |
| `latestbreach` | — | not yet ([Roadmap](roadmap.md)) |
| `subscription/status` | — | not yet ([Roadmap](roadmap.md)) |
| `breacheddomain/{domain}` (domain search) | — | not yet ([Roadmap](roadmap.md)) |
| Stealer-log endpoints | — | not yet ([Roadmap](roadmap.md)) |
| NTLM mode for Pwned Passwords | — | not yet ([Roadmap](roadmap.md)) |

The `Breach` model includes the newer flags `IsMalware`, `IsStealerLog`, and `IsSubscriptionFree`.

## Maintenance model

This project is deliberately **low-maintenance**:

- Scans and builds run on push/PR only, so badges cannot go red while the repo is idle.
- Dependabot runs monthly with grouped PRs (one per ecosystem).
- Container images are published **only** on `v*` tags or a manual workflow run — never on regular pushes.
- Published Go module versions are immutable (Go module proxy); any fix ships as a new patch tag.

See [Releases and Versioning](releases.md) for the release procedure and [Roadmap](roadmap.md) for what's next.
