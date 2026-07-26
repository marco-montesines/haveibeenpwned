# HTTP API Reference

`hibp serve` exposes the library as a JSON-over-HTTP API — the same thing the [GHCR container image](https://github.com/marco-montesines/haveibeenpwned/pkgs/container/haveibeenpwned) runs. Intended as an **internal service** for your own stack (PHP, Python, Node, shell) — see the deployment note at the bottom.

Base URL in the examples: `http://localhost:8080`.

## Endpoints

### `GET /healthz`

Health/liveness check. No upstream call.

```bash
curl -s localhost:8080/healthz
# {"status":"ok"}
```

### `POST /v1/pwnedpassword`

Pwned Passwords check via k-anonymity (only the first 5 chars of the SHA-1 hash reach HIBP). A `POST` with a JSON body, deliberately — the password never appears in URLs, query strings, or access logs. No API key required.

Request body: `{"password": "..."}`

```bash
curl -s -X POST localhost:8080/v1/pwnedpassword -d '{"password":"P@ssw0rd"}'
# {"count":6421042,"pwned":true}
```

`400` if the body is not JSON or `password` is empty.

### `GET /v1/breaches`

All breaches in the HIBP catalogue. No API key required.

| Query param | Meaning |
| ----------- | ------- |
| `domain` | Optional — only breaches for this domain |

```bash
curl -s 'localhost:8080/v1/breaches?domain=adobe.com'
# [{"Name":"Adobe","Title":"Adobe","Domain":"adobe.com","BreachDate":"2013-10-04","PwnCount":152445165,...}]
```

### `GET /v1/breaches/{name}`

A single breach by its name (e.g. `Adobe`). No API key required.

```bash
curl -s localhost:8080/v1/breaches/Adobe
```

### `GET /v1/dataclasses`

All data classes HIBP tracks. No API key required.

```bash
curl -s localhost:8080/v1/dataclasses
# ["Account balances","Email addresses","Passwords",...]
```

### `GET /v1/breachedaccount/{email}`

Breaches containing an account. **Requires the server to be started with an API key.**

| Query param | Default | Meaning |
| ----------- | ------- | ------- |
| `domain` | — | Filter results by breach domain |
| `truncate` | `true` | `true` → breach names only; `false` → full breach objects |
| `unverified` | `true` | Include unverified breaches |

```bash
curl -s 'localhost:8080/v1/breachedaccount/info@example.com?truncate=false'
```

An account not found in any breach returns `200` with `[]` — **not** a 404.

### `GET /v1/pasteaccount/{email}`

Pastes containing an account. **Requires an API key.** Not-found likewise returns `200` with `[]`.

```bash
curl -s localhost:8080/v1/pasteaccount/info@example.com
```

## Error responses

Errors are JSON: `{"error": "..."}`. Upstream HIBP errors pass through their status code:

| Status | Meaning |
| ------ | ------- |
| `400` | Bad request (malformed body, invalid input) |
| `401` | Missing/invalid HIBP API key (account endpoints) |
| `429` | HIBP rate limit — the `Retry-After` header is forwarded; wait that many seconds |
| `502` | Upstream/network failure that wasn't an HIBP API error |

The server never retries toward HIBP; honor `Retry-After` in your caller.

## Field names

Breach/paste objects use HIBP's PascalCase JSON field names (`Name`, `Title`, `BreachDate`, `PwnCount`, `DataClasses`, `IsVerified`, …) — see the [HIBP data model docs](https://haveibeenpwned.com/API/v3#BreachModel) and [[Library Guide|Library-Guide]].

## Deployment note

The server has **no authentication, TLS, or rate limiting of its own** — it is designed to sit on an internal network (Compose network, cluster-internal Service) behind your own perimeter. Don't expose it directly to the public internet; if external exposure is unavoidable, put it behind an authenticating reverse proxy. Remember the shared HIBP API key's rate limit is consumed by everyone calling the service. See [[Deployment and IaC|Deployment-and-IaC]].
