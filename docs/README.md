# Documentation

A Go client for [Troy Hunt's Have I Been Pwned API v3](https://haveibeenpwned.com/API/v3), including the [Pwned Passwords](https://haveibeenpwned.com/Passwords) k-anonymity range API.

One codebase, four ways to consume it:

| Form | For whom | Where |
| ---- | -------- | ----- |
| **Go library** | Go applications | `go get github.com/marco-montesines/haveibeenpwned` |
| **CLI** (`hibp`) | Shell scripts, ad-hoc checks, CI | `go install .../cmd/hibp@latest` |
| **HTTP JSON API** (`hibp serve`) | Any language over HTTP | `ghcr.io/marco-montesines/haveibeenpwned` |
| **FrankenPHP extension** | PHP, as native functions | `frankenphp/` directory |

## Where to start

- **New here?** → [Getting Started](getting-started.md) — install, set up, and configure each form.
- **What can I build with it?** → [Use Cases](use-cases.md) — signup password policies, breach monitoring, and more.
- **Integrating the Go library** → [Library Guide](library.md) — every method, option, and error type.
- **Calling it over HTTP** → [HTTP API Reference](http-api.md) — every endpoint with request/response examples.
- **Using the CLI** → [CLI Guide](cli.md).
- **PHP integration** → [FrankenPHP Extension](frankenphp.md).
- **Running the platform** → [Deployment and IaC](deployment.md) — Docker, Compose, Kubernetes, systemd, configuration reference.

## Project

- [Project Status](status.md) — what works today, current version, HIBP endpoint coverage.
- [Roadmap](roadmap.md) — where the project is headed.
- [Releases and Versioning](releases.md) — how releases and image tags work.
- [FAQ and Troubleshooting](faq.md).

## Key properties

- **Privacy by design** — password checks use k-anonymity: only the first five characters of the password's SHA-1 hash ever leave your process, and responses are padded.
- **No API key needed** for password checks, breach catalogue, and data classes. An [API key](https://haveibeenpwned.com/API/Key) is required only for per-account lookups (breached account, pastes).
- **Well-behaved client** — rate-limit aware (`Retry-After` surfaced, never auto-retried), honest User-Agent, compliant with the [HIBP acceptable use policy](https://haveibeenpwned.com/API/v3#AcceptableUse).
- **Tested without the network** — the full suite runs against `httptest` mocks; CI adds race detection, `govulncheck`, CodeQL, and secret scanning.

> **Unofficial client.** This project is not affiliated with, endorsed by, or sponsored by Have I Been Pwned or Troy Hunt. Breach and paste data is provided by the [Have I Been Pwned](https://haveibeenpwned.com) service under [CC BY 4.0](https://creativecommons.org/licenses/by/4.0/); applications displaying it must visibly credit haveibeenpwned.com. The code is [MIT licensed](../LICENSE).
