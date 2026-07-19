# FrankenPHP extension

This directory contains a [FrankenPHP PHP extension written in Go](https://frankenphp.dev/docs/extensions/)
that exposes the `haveibeenpwned` library as **native PHP functions** — no HTTP
round-trip, no FFI, the Go code runs inside the PHP process.

## PHP API

```php
/**
 * Number of times the password appears in the Pwned Passwords corpus.
 * k-anonymity: only the first 5 chars of the SHA-1 hash leave the process.
 * Returns -1 on lookup failure. No API key required.
 */
function hibp_pwned_password_count(string $password): int {}

/**
 * JSON-encoded array of breaches, optionally filtered by domain
 * (empty string = all breaches). Returns {"error": "..."} on failure.
 * No API key required.
 */
function hibp_breaches(string $domain = ""): string {}
```

Example:

```php
<?php
if (hibp_pwned_password_count($_POST['password']) > 0) {
    throw new RuntimeException('This password appears in known data breaches.');
}

$adobeBreaches = json_decode(hibp_breaches('adobe.com'), true);
```

## Layout

- `extension/` — the extension itself: Go implementation (`extension.go`),
  the C bridge to the Zend engine (`extension.c`, `extension.h`,
  `extension_arginfo.h`), and IDE stubs (`extension.stub.php`).
- `example/` — a small demo page served by the compose service.
- `Dockerfile` — builds FrankenPHP with the extension compiled in.

## Build & run (Docker, recommended)

PHP extensions must be compiled into the FrankenPHP binary with
[xcaddy](https://github.com/caddyserver/xcaddy). The Dockerfile does this for
you. From the **repository root**:

```bash
docker build -f frankenphp/Dockerfile -t hibp-frankenphp .
docker run --rm -p 8081:80 -e SERVER_NAME=":80" hibp-frankenphp
# open http://localhost:8081/?password=P@ssw0rd&domain=adobe.com
```

Or via docker compose from the repository root:

```bash
docker compose up frankenphp
```

## Build locally (without Docker)

Requires PHP ≥ 8.2 built with ZTS, the PHP development headers, and xcaddy:

```bash
CGO_ENABLED=1 \
CGO_CFLAGS="$(php-config --includes)" \
CGO_LDFLAGS="$(php-config --ldflags) $(php-config --libs)" \
XCADDY_GO_BUILD_FLAGS="-ldflags='-w -s' -tags=nobadger,nomysql,nopgx" \
xcaddy build \
  --output frankenphp \
  --with github.com/dunglas/frankenphp/caddy \
  --with github.com/marco-montesines/haveibeenpwned/frankenphp/extension=./frankenphp/extension \
  --with github.com/marco-montesines/haveibeenpwned=.

./frankenphp php-server -r frankenphp/example/
```

## Configuration

| Environment variable | Purpose                                                        |
| -------------------- | -------------------------------------------------------------- |
| `HIBP_API_KEY`       | Optional; only needed if you extend the extension with account lookups. |

## Data source & attribution

Breach data returned by `hibp_breaches()` comes from
[Have I Been Pwned](https://haveibeenpwned.com) and is licensed
[CC BY 4.0](https://creativecommons.org/licenses/by/4.0/): if your PHP
application displays it, you must clearly attribute Have I Been Pwned as the
source with a visible link. Password checks via
`hibp_pwned_password_count()` use the Pwned Passwords range API, which has
no attribution requirement. This project is not affiliated with or endorsed
by Have I Been Pwned or Troy Hunt.
