# Security Policy

## Supported versions

Only the latest release receives security fixes.

## Reporting a vulnerability

Please report vulnerabilities privately via
[GitHub private vulnerability reporting](https://github.com/marco-montesines/haveibeenpwned/security/advisories/new).
Do not open a public issue for security problems.

You can expect an initial response within 7 days.

## Scope

This project is an unofficial client for the public
[Have I Been Pwned](https://haveibeenpwned.com) API. Password checks use the
k-anonymity range API: only the first five hex characters of the password's
SHA-1 hash ever leave the process, and padded responses are requested.
Vulnerabilities in the HIBP service itself are out of scope — report those to
[haveibeenpwned.com](https://haveibeenpwned.com).

## Automated checks

Every push and pull request runs:

- `govulncheck` — known vulnerabilities in dependencies and the Go standard
  library (only findings reachable from this code fail the build)
- CodeQL — static analysis for common vulnerability patterns
- gitleaks — secret scanning over the full git history
- `go vet` and race-enabled tests

Dependabot keeps Go modules, GitHub Actions, and Docker base images current.
