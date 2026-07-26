# FAQ and Troubleshooting

## FAQ

### Do I need an API key?

Only for the per-account endpoints (breached account, pastes). Password checks, the breach catalogue, and data classes are free and keyless. Keys are sold by HIBP at [haveibeenpwned.com/API/Key](https://haveibeenpwned.com/API/Key) — this project has no key of its own to share.

### Is it safe to send passwords to this?

The plaintext password never leaves your process (library/extension) or your server (`hibp serve`). Checks use the k-anonymity range API: SHA-1 locally, send the first five hex characters of the hash, receive a padded candidate list, compare locally. The HTTP API takes the password as a `POST` body specifically so it stays out of URLs and access logs.

### Why does an account lookup return an empty result instead of a 404?

By design, "not found in any breach" is the *good* outcome, not an error. The library returns `(nil, nil)`, the CLI prints `[]` (plus a stderr note), and the HTTP API returns `200` with `[]`. Real errors (bad key, rate limit, network) are still errors.

### I got a 429. What now?

You hit the HIBP rate limit for your API key. The error carries the `Retry-After` value (`HIBPErrorResponse.RetryAfter` in Go, the `Retry-After` header from `hibp serve`). Wait that many seconds and retry. The client will never retry for you — that's [deliberate](roadmap.md#non-goals).

### Is this an official HIBP client?

No. It is not affiliated with, endorsed by, or sponsored by Have I Been Pwned or Troy Hunt.

### Can I show breach data in my app?

Yes, with attribution: breach and paste data is licensed CC BY 4.0, so anything *displaying* it must visibly credit [haveibeenpwned.com](https://haveibeenpwned.com) with a link. Pwned Passwords range checks carry no attribution requirement.

### How do I suggest a change to this documentation?

Open a pull request against the [`docs/`](https://github.com/marco-montesines/haveibeenpwned/tree/master/docs) directory — the docs are versioned and reviewed together with the code.

### Which Go version do I need?

Go 1.25+ for the library and CLI. The Docker images have no Go requirement at all.

## Troubleshooting

### My IDE shows errors in `frankenphp/extension/` (`php.h` not found)

Expected. The FrankenPHP extension is a separate cgo module whose C headers only exist inside `frankenphp/Dockerfile`; the repo's `go.work` keeps it out of the local workspace so `gopls` ignores it. Don't delete `go.work`, and don't try to "fix" those errors locally — `docker build -f frankenphp/Dockerfile .` is the source of truth for that module.

### JetBrains flags Go snippets in README.md as broken

False positive: injected markdown code fences have no Go module context, so imports never resolve — even though the quickstart snippet is a complete program mirrored in `examples/quickstart` and compiled in CI. IDE fix: Settings → Languages & Frameworks → Markdown → "Hide errors in code fences".

### `401 Unauthorized` on account/pastes endpoints

The API key is missing or invalid. Library: pass it to `hibp.New(...)`. CLI: `-api-key` or `$HIBP_API_KEY`. Server: start the container with `-e HIBP_API_KEY=...` — the key is read at startup, so restart after changing it.

### The FrankenPHP container exits with `module not registered: http.encoders.br`

The stock FrankenPHP Caddyfile enables brotli, which this custom build doesn't include. The repo's `frankenphp/Caddyfile` overrides it — make sure your image build includes that file (the shipped `frankenphp/Dockerfile` does).

### `hibp serve` works locally but times out from another container

Bind to all interfaces inside containers. The image already does; if you run the binary yourself, use `hibp serve -addr :8080` (not `127.0.0.1:8080`) and reach it via the service name on the container network (e.g. `http://hibp-api:8080`), not `localhost`.

### Password check counts differ from last month

Normal — HIBP ingests new breaches continuously, so counts only grow. Don't pin exact counts in tests; assert `pwned == true` or `count > 0` instead, or mock the API (see the testing section of the [Library Guide](library.md)).

### Something else?

[Open an issue](https://github.com/marco-montesines/haveibeenpwned/issues) — please include the form you're using (library / CLI / HTTP API / FrankenPHP), the version or image tag, and the exact error output. For security reports, see [SECURITY.md](https://github.com/marco-montesines/haveibeenpwned/blob/master/SECURITY.md) — please use private vulnerability reporting, not a public issue.
