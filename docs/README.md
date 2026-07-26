# haveibeenpwned

A Go client for [Troy Hunt's Have I Been Pwned API v3](https://haveibeenpwned.com/API/v3), including the [Pwned Passwords](https://haveibeenpwned.com/Passwords) k-anonymity range API.

<div class="hibp-hero">
  <span class="hibp-hero__label">k-anonymity range query</span>
  <div class="hibp-hero__hash">SHA1("password")&nbsp;=&nbsp;<span class="keep">5BAA6</span><span class="drop">1E4C9B93F3F0682250B6CF8331B7EE68FD8</span></div>
  <p class="hibp-hero__caption">Only the five highlighted characters ever leave your process — the password itself never does, and responses are padded.</p>
</div>

One codebase, four ways to consume it:

<div class="hibp-cards">
  <a class="hibp-card" href="library/"><strong>Go library</strong><span>Typed client with functional options, context support, and honest errors.</span></a>
  <a class="hibp-card" href="cli/"><strong>hibp CLI</strong><span>JSON output for shells, scripts, and CI. Reads passwords from stdin.</span></a>
  <a class="hibp-card" href="http-api/"><strong>HTTP JSON API</strong><span>hibp serve — one shared breach-check service for any language.</span></a>
  <a class="hibp-card" href="frankenphp/"><strong>PHP extension</strong><span>Native PHP functions compiled from this library via FrankenPHP.</span></a>
</div>

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

> **Unofficial client.** This project is not affiliated with, endorsed by, or sponsored by Have I Been Pwned or Troy Hunt. Breach and paste data is provided by the [Have I Been Pwned](https://haveibeenpwned.com) service under [CC BY 4.0](https://creativecommons.org/licenses/by/4.0/); applications displaying it must visibly credit haveibeenpwned.com. The code is [MIT licensed](https://github.com/marco-montesines/haveibeenpwned/blob/master/LICENSE).
