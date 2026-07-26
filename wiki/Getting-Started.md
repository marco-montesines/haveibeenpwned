# Getting Started

Pick the form that matches how you want to consume the client, then follow that section. Configuration shared by all forms is at the bottom.

## Which form do I want?

| You are… | Use | Guide |
| -------- | --- | ----- |
| Writing a Go application | Go library | [[Library Guide|Library-Guide]] |
| In a terminal, a shell script, or CI | `hibp` CLI | [[CLI Guide|CLI-Guide]] |
| Using PHP/Python/Node/anything over HTTP | `hibp serve` / Docker image | [[HTTP API Reference|HTTP-API-Reference]], [[Deployment and IaC|Deployment-and-IaC]] |
| Running PHP on FrankenPHP | Native PHP extension | [[FrankenPHP Extension|FrankenPHP-Extension]] |

## 1. Go library

Requires Go 1.25+.

```bash
go get github.com/marco-montesines/haveibeenpwned
```

```go
client := hibp.New(os.Getenv("HIBP_API_KEY")) // key may be empty for unauthenticated endpoints
count, err := client.PwnedPasswordCount(ctx, "P@ssw0rd")
```

A complete runnable example lives in [`examples/quickstart`](https://github.com/marco-montesines/haveibeenpwned/blob/master/examples/quickstart/main.go). Full API docs: [[Library Guide|Library-Guide]] and [pkg.go.dev](https://pkg.go.dev/github.com/marco-montesines/haveibeenpwned).

## 2. CLI

```bash
go install github.com/marco-montesines/haveibeenpwned/cmd/hibp@latest

hibp breaches -domain adobe.com
echo -n 'P@ssw0rd' | hibp password
```

No Go toolchain on the target machine? The container image includes the same binary:

```bash
docker run --rm ghcr.io/marco-montesines/haveibeenpwned:latest --help
```

Full command reference: [[CLI Guide|CLI-Guide]].

## 3. HTTP API server

The fastest path is the published image (multi-arch, amd64 + arm64):

```bash
docker run --rm -p 8080:8080 -e HIBP_API_KEY=... ghcr.io/marco-montesines/haveibeenpwned:latest

curl -s localhost:8080/healthz
curl -s -X POST localhost:8080/v1/pwnedpassword -d '{"password":"P@ssw0rd"}'
```

Without Docker:

```bash
go install github.com/marco-montesines/haveibeenpwned/cmd/hibp@latest
hibp serve -addr :8080
```

Endpoint reference: [[HTTP API Reference|HTTP-API-Reference]]. Production setups (Compose, Kubernetes, systemd): [[Deployment and IaC|Deployment-and-IaC]].

## 4. FrankenPHP demo (API + PHP page)

From a clone of the repository:

```bash
HIBP_API_KEY=... docker compose up --build
# API:             http://localhost:8080
# FrankenPHP demo: http://localhost:8081/?password=P@ssw0rd&domain=adobe.com
```

Details: [[FrankenPHP Extension|FrankenPHP-Extension]].

## Configuration (all forms)

| Setting | Library | CLI | Server / Docker |
| ------- | ------- | --- | --------------- |
| HIBP API key | `New(apiKey)` | `-api-key` flag or `$HIBP_API_KEY` | `HIBP_API_KEY` env var |
| Listen address | — | — | `serve -addr :8080` (default `:8080`) |
| Request timeout | your `context.Context` | `-timeout` (default 30s) | inherited per request |
| API base URL override | `WithBaseURL` | `-base-url` | — |
| User-Agent | `WithUserAgent` | default UA | default UA |

**Do you need an API key?** Only for the per-account endpoints (`account`/`pastes`, `GetBreachedAccount`/`GetPastedAccount`, `/v1/breachedaccount`, `/v1/pasteaccount`). Password checks, the breach catalogue, and data classes work without one. Get a key at [haveibeenpwned.com/API/Key](https://haveibeenpwned.com/API/Key).

**Set a descriptive User-Agent** identifying your project (`WithUserAgent("my-app/1.0")`) — the HIBP acceptable use policy asks for it.
