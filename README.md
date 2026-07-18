# haveibeenpwned

[![CI](https://github.com/marco-montesines/haveibeenpwned/actions/workflows/ci.yml/badge.svg)](https://github.com/marco-montesines/haveibeenpwned/actions/workflows/ci.yml)
[![Docker](https://github.com/marco-montesines/haveibeenpwned/actions/workflows/docker.yml/badge.svg)](https://github.com/marco-montesines/haveibeenpwned/actions/workflows/docker.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/marco-montesines/haveibeenpwned.svg)](https://pkg.go.dev/github.com/marco-montesines/haveibeenpwned)
[![Go Report Card](https://goreportcard.com/badge/github.com/marco-montesines/haveibeenpwned)](https://goreportcard.com/report/github.com/marco-montesines/haveibeenpwned)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A Go client for [Troy Hunt's Have I Been Pwned API v3](https://haveibeenpwned.com/API/v3),
including the [Pwned Passwords](https://haveibeenpwned.com/Passwords) k-anonymity range API.

Ships in four forms:

- **Go library** — `github.com/marco-montesines/haveibeenpwned`
- **CLI** — `hibp` (`cmd/hibp`)
- **HTTP JSON API** — `hibp serve`, published as a container image at
  [`ghcr.io/marco-montesines/haveibeenpwned`](https://ghcr.io/marco-montesines/haveibeenpwned)
- **FrankenPHP extension** — native PHP functions implemented in Go
  ([`frankenphp/`](frankenphp/README.md))

## Features

| Method                          | HIBP endpoint                     | API key required |
| ------------------------------- | --------------------------------- | :--------------: |
| `PwnedPasswordCount`            | `range/{first5HashChars}`         |        no        |
| `GetBreaches`                   | `breaches`                        |        no        |
| `GetBreachedSite`               | `breach/{name}`                   |        no        |
| `GetDataClasses`                | `dataclasses`                     |        no        |
| `GetBreachedAccount`            | `breachedaccount/{account}`       |       yes        |
| `GetPastedAccount`              | `pasteaccount/{account}`          |       yes        |

An API key for the account endpoints is available at
[haveibeenpwned.com/API/Key](https://haveibeenpwned.com/API/Key).
Password checks use k-anonymity: only the first five characters of the
password's SHA-1 hash ever leave your process, and responses are padded.

## Library

### Install

```bash
go get github.com/marco-montesines/haveibeenpwned
```

### Quickstart

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	hibp "github.com/marco-montesines/haveibeenpwned"
)

func main() {
	ctx := context.Background()
	client := hibp.New(os.Getenv("HIBP_API_KEY")) // key may be empty for unauthenticated endpoints

	// Has this password appeared in a breach? (no API key needed)
	count, err := client.PwnedPasswordCount(ctx, "P@ssw0rd")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("password seen %d times in breaches\n", count)

	// All breaches for a domain (no API key needed).
	breaches, err := client.GetBreaches(ctx, "adobe.com")
	if err != nil {
		log.Fatal(err)
	}
	for _, b := range breaches {
		fmt.Printf("%s (%s): %d accounts\n", b.Title, b.BreachDate, b.PwnCount)
	}

	// Breaches for an account (requires API key). nil result = not breached.
	accountBreaches, err := client.GetBreachedAccount(ctx, "info@example.com", "", true, true)
	if err != nil {
		log.Fatal(err)
	}
	if accountBreaches == nil {
		fmt.Println("account not found in any breach")
	}
}
```

A runnable version lives in [`examples/quickstart`](examples/quickstart/main.go).

### Options

```go
client := hibp.New(apiKey,
	hibp.WithHTTPClient(customHTTPClient),
	hibp.WithUserAgent("my-app/1.0"),
	hibp.WithBaseURL(proxyURL),                 // e.g. tests, corporate proxy
	hibp.WithPwnedPasswordsBaseURL(mirrorURL),
)
```

### Error handling

Non-2xx API responses are returned as `*hibp.HIBPErrorResponse`, which
preserves the HTTP status code and the `Retry-After` value on rate limits:

```go
var apiErr *hibp.HIBPErrorResponse
if errors.As(err, &apiErr) && apiErr.Code == http.StatusTooManyRequests {
	time.Sleep(time.Duration(apiErr.RetryAfter) * time.Second)
}
```

A `404` on account lookups is **not** an error — it means "not breached" and
yields a `nil` result with a `nil` error.

## CLI

```bash
go install github.com/marco-montesines/haveibeenpwned/cmd/hibp@latest

hibp breaches -domain adobe.com          # all breaches for a domain
hibp breach Adobe                        # a single breach by name
hibp dataclasses                         # all data classes
echo -n 'P@ssw0rd' | hibp password       # check a password (stdin keeps it out of shell history)

export HIBP_API_KEY=...                  # required for account endpoints
hibp account info@example.com
hibp pastes info@example.com
```

All output is JSON, so it composes with `jq`.

## HTTP API server & Docker

`hibp serve` exposes the same functionality as a JSON API — useful from PHP,
scripts, or any non-Go service:

| Route                              | Description                                        |
| ---------------------------------- | -------------------------------------------------- |
| `GET /healthz`                     | Health check                                       |
| `GET /v1/breaches?domain=`         | All breaches, optional domain filter               |
| `GET /v1/breaches/{name}`          | Single breach by name                              |
| `GET /v1/dataclasses`              | All data classes                                   |
| `GET /v1/breachedaccount/{email}`  | Breaches for an account (`?domain=&truncate=&unverified=`) |
| `GET /v1/pasteaccount/{email}`     | Pastes for an account                              |
| `POST /v1/pwnedpassword`           | Body `{"password":"..."}` → `{"pwned":true,"count":N}` |

Run it from the published image:

```bash
docker run --rm -p 8080:8080 -e HIBP_API_KEY=... ghcr.io/marco-montesines/haveibeenpwned:latest

curl -s -X POST localhost:8080/v1/pwnedpassword -d '{"password":"P@ssw0rd"}'
# {"count":6421042,"pwned":true}
```

Or with docker compose (starts the API **and** the FrankenPHP demo):

```bash
HIBP_API_KEY=... docker compose up --build
# API:       http://localhost:8080
# FrankenPHP demo: http://localhost:8081/?password=P@ssw0rd
```

Images are built for `linux/amd64` and `linux/arm64` and published to GHCR by
[`.github/workflows/docker.yml`](.github/workflows/docker.yml) on every push
to `master` (`:latest`, `:sha-…`) and on version tags (`:1.2.3`, `:1.2`).

## FrankenPHP: call the library natively from PHP

The [`frankenphp/`](frankenphp/README.md) directory contains a
[FrankenPHP extension written in Go](https://frankenphp.dev/docs/extensions/)
that compiles this library **into** PHP:

```php
<?php
if (hibp_pwned_password_count($_POST['password']) > 0) {
    throw new RuntimeException('This password appears in known data breaches.');
}

$breaches = json_decode(hibp_breaches('adobe.com'), true);
```

Build and try it:

```bash
docker compose up frankenphp
# open http://localhost:8081/?password=P@ssw0rd&domain=adobe.com
```

See [frankenphp/README.md](frankenphp/README.md) for details.

## Development

```bash
go build ./...        # build library, CLI, examples
go test ./... -race   # run the test suite (uses httptest mocks, no network)
go vet ./...
```

CI runs formatting, vet, build, and race-enabled tests on every push and pull
request.

## Rate limits & acceptable use

The authenticated endpoints are rate limited per API key; on `429` the client
surfaces the `Retry-After` value via `HIBPErrorResponse.RetryAfter`. Please
respect the [HIBP acceptable use policy](https://haveibeenpwned.com/API/v3#AcceptableUse)
and set a descriptive User-Agent (`WithUserAgent`) identifying your project.

## License

[MIT](LICENSE). Not affiliated with or endorsed by Have I Been Pwned or Troy
Hunt. Data is provided by the [Have I Been Pwned](https://haveibeenpwned.com)
service.
