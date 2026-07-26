# Roadmap

Where the project is headed. Items below are **intentions, not commitments** — this is a spare-time open-source project. If you need one of these sooner, [open an issue](https://github.com/marco-montesines/haveibeenpwned/issues) or send a PR.

## Shipped (v1.0.x, July 2026)

- ✅ Library modernization: functional options (`WithHTTPClient`, `WithUserAgent`, `WithBaseURL`, `WithPwnedPasswordsBaseURL`), context support, URL-escaped path parameters, typed `*HIBPErrorResponse` errors with `Retry-After` on 429.
- ✅ Pwned Passwords k-anonymity support (`PwnedPasswordCount`) with response padding.
- ✅ `404` on account lookups treated as "not breached" (`nil, nil`), not an error.
- ✅ Newer breach model flags: `IsMalware`, `IsStealerLog`, `IsSubscriptionFree`.
- ✅ CLI (`cmd/hibp`) with JSON output and stdin password input.
- ✅ HTTP JSON API (`hibp serve`) + multi-arch Docker image on GHCR.
- ✅ FrankenPHP extension exposing native PHP functions.
- ✅ Security pipeline: race-enabled tests, `govulncheck`, CodeQL, gitleaks, Dependabot.
- ✅ Documentation: in-repo `docs/` published as a [website](https://marco-montesines.github.io/haveibeenpwned/) via GitHub Pages.

## Planned — next

API surface (tracking HIBP API v3 growth):

- [ ] `GetLatestBreach` — the `latestbreach` endpoint (most recently added breach).
- [ ] `GetSubscriptionStatus` — the `subscription/status` endpoint, so key holders can inspect their plan and rate limit programmatically.
- [ ] NTLM hash mode for Pwned Passwords range queries (`?mode=ntlm`).

Platform:

- [ ] Expose the remaining library endpoints through `hibp serve` and the CLI as they are added.
- [ ] GitHub Releases with generated changelogs for existing and future tags (tags currently exist without release notes).

## Under consideration

- **Domain search** — `breacheddomain/{domain}` and `subscribeddomains` (requires a subscribed, domain-verified API key; needs test fixtures designed around that).
- **Stealer-log endpoints** — email/website/email-domain stealer-log lookups (subscription-gated).
- **Optional response caching helper** — e.g. for the breach catalogue, which changes rarely. Would be opt-in and documented with the CC BY 4.0 attribution obligations in mind.
- **More FrankenPHP functions** — e.g. `hibp_breached_account()`; the current extension deliberately exposes only the unauthenticated endpoints.
- **Additional deployment recipes** — Helm chart or Kustomize base if there is demand (see [Deployment and IaC](deployment.md) for what exists today).

## Non-goals

Things this project intentionally will **not** do:

- **Automatic retries on 429.** The HIBP acceptable use policy expects clients to back off; the client surfaces `Retry-After` and leaves the decision to you.
- **Bundled breach-data storage or redistribution.** This is a live API client, not a dataset mirror.
- **Claiming official status.** This is and will remain an unofficial client.
- **Feature-flag sprawl.** The library stays a thin, predictable HTTP client; anything opinionated (queueing, caching, persistence) belongs in your application or a separate package.

## Feedback

The roadmap is shaped by actual usage. If your use case ([Use Cases](use-cases.md)) needs something that isn't here, open an issue — small, well-scoped feature requests with a concrete use case get picked up first.
