# FrankenPHP Extension

The [`frankenphp/`](https://github.com/marco-montesines/haveibeenpwned/tree/master/frankenphp) directory contains a [FrankenPHP extension written in Go](https://frankenphp.dev/docs/extensions/) that compiles this library **into** PHP — HIBP checks become native PHP functions with no HTTP round-trip to a sidecar service.

## Functions

```php
function hibp_pwned_password_count(string $password): int
function hibp_breaches(string $domain = ""): string   // JSON, decode with json_decode()
```

```php
<?php
if (hibp_pwned_password_count($_POST['password']) > 0) {
    throw new RuntimeException('This password appears in known data breaches.');
}

$breaches = json_decode(hibp_breaches('adobe.com'), true);
```

The extension deliberately exposes only the **unauthenticated** endpoints (password range check, breach catalogue). For account-level lookups from PHP, call the [[HTTP API|HTTP-API-Reference]] instead. More functions are [[under consideration|Roadmap]].

## Try it

From a clone of the repository:

```bash
docker compose up frankenphp
# open http://localhost:8081/?password=P@ssw0rd&domain=adobe.com
```

Or build the image directly (compiles PHP + the extension, takes several minutes):

```bash
docker build -f frankenphp/Dockerfile -t hibp-frankenphp .
docker run --rm -p 8081:80 -e SERVER_NAME=":80" hibp-frankenphp
```

## How it's put together

| Piece | Role |
| ----- | ---- |
| `frankenphp/extension/extension.go` | Go implementation of the PHP functions |
| `extension.c` / `extension.h` / `extension_arginfo.h` | C bridge (Zend arginfo, kept in sync with `extension.stub.php`) |
| `frankenphp/Dockerfile` | xcaddy build that compiles FrankenPHP with the extension |
| `frankenphp/Caddyfile` | Overrides the stock Caddyfile (the stock one enables brotli, which this custom build doesn't include) |
| `frankenphp/example/` | The demo PHP page |

The extension is a **separate Go module** that needs cgo and PHP development headers, which only exist inside the Docker build. That's why:

- Your IDE will report phantom errors (`php.h` not found) if it type-checks `frankenphp/extension/` locally — the repo's `go.work` exists precisely to keep that module out of the local workspace. The Docker build is the source of truth. See [[FAQ and Troubleshooting|FAQ-and-Troubleshooting]].
- Both the extension module **and** the root library module are mapped into the xcaddy build with `--with mod=path` — replace directives in a dependency's `go.mod` are ignored by Go, so both mappings are required.

## Attribution

If your PHP application displays breach data from `hibp_breaches()`, you must visibly credit [Have I Been Pwned](https://haveibeenpwned.com) — the data is CC BY 4.0. Password counts from `hibp_pwned_password_count()` carry no attribution requirement. Details in [`frankenphp/README.md`](https://github.com/marco-montesines/haveibeenpwned/blob/master/frankenphp/README.md).
