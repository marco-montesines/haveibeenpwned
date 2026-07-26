# CLI Guide

`hibp` is a single binary exposing the whole library from the command line, plus the HTTP API server. All output is JSON (pretty-printed), so it composes with `jq`.

## Install

```bash
go install github.com/marco-montesines/haveibeenpwned/cmd/hibp@latest
```

Or run it from the container image without a Go toolchain:

```bash
docker run --rm ghcr.io/marco-montesines/haveibeenpwned:latest breaches -domain adobe.com
```

## Usage

```
hibp [global flags] <command> [command flags] [arguments]
```

### Global flags

| Flag | Default | Meaning |
| ---- | ------- | ------- |
| `-api-key` | `$HIBP_API_KEY` | HIBP API key (needed for `account` and `pastes` only) |
| `-base-url` | HIBP production | Override the API base URL (testing, proxies) |
| `-timeout` | `30s` | Request timeout |

Global flags go **before** the command: `hibp -timeout 10s breaches`.

### Commands

| Command | API key | Description |
| ------- | :-----: | ----------- |
| `breaches [-domain example.com]` | no | List all breaches, optionally filtered by domain |
| `breach <name>` | no | Single breach by name, e.g. `hibp breach Adobe` |
| `dataclasses` | no | List all data classes |
| `password [<password>]` | no | Pwned Passwords check — reads stdin if the argument is omitted |
| `account <email> [flags]` | **yes** | Breaches for an account |
| `pastes <email>` | **yes** | Pastes for an account |
| `serve [-addr :8080]` | optional | Run the HTTP JSON API ([[HTTP API Reference|HTTP-API-Reference]]) |

`account` flags (after the email): `-domain` (filter by breach domain), `-truncate` (default `true`, names only), `-unverified` (default `true`, include unverified breaches).

## Examples

```bash
hibp breaches -domain adobe.com                 # all breaches for a domain
hibp breach Adobe | jq .PwnCount                # one field of one breach
hibp dataclasses

# Password checks — prefer stdin, it keeps the password out of shell history:
echo -n 'P@ssw0rd' | hibp password
hibp password                                    # interactive prompt on a TTY
# {"count": 6421042, "pwned": true}

export HIBP_API_KEY=...
hibp account info@example.com
hibp account info@example.com -truncate=false    # full breach objects
hibp pastes info@example.com
```

## Behavior notes

- **Not breached is not an error.** For `account`/`pastes`, an account absent from HIBP prints `good news: account not found in any breach` on **stderr** and an empty JSON array `[]` on **stdout**, exiting 0 — scripts consuming stdout always get valid JSON.
- **Exit codes**: `0` success, `1` on errors (message on stderr, prefixed `hibp:`), `2` for usage errors.
- **Rate limits**: a `429` surfaces as an error including the status; the CLI does not retry. Wait the advertised `Retry-After` before rerunning.
