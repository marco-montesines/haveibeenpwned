# Releases and Versioning

## Versioning scheme

Semantic versioning via git tags (`v*`, e.g. `v1.2.3`). Within `v1.x`, the public Go API is stable — breaking changes would mean a new major version.

## What a release produces

Pushing a `v*` tag triggers [`docker.yml`](https://github.com/marco-montesines/haveibeenpwned/blob/master/.github/workflows/docker.yml), which publishes multi-arch (linux/amd64 + linux/arm64) images to GHCR:

| Tag | Example | Use for |
| --- | ------- | ------- |
| `:X.Y.Z` | `:1.2.3` | Production — fully pinned |
| `:X.Y` | `:1.0` | Auto-pick patch releases |
| `:latest` | | Experiments only |
| `:sha-<commit>` | | Exact provenance |

The same git tag **is** the Go module release: `go get github.com/marco-montesines/haveibeenpwned@v1.2.3`. There is no separate registry publish step — the Go module proxy fetches the tag from GitHub and caches it permanently, and [pkg.go.dev](https://pkg.go.dev/github.com/marco-montesines/haveibeenpwned) regenerates the API docs from the source.

## Immutability

Published module versions are **immutable**: once the Go module proxy has cached a version, its contents can never change (re-tagging would break checksums for consumers). Any fix — even a one-liner — ships as a new patch tag.

## When images are built (deliberately not on every push)

Container images are published **only** on:

1. a `v*` tag (the normal release path), or
2. a manual workflow run (Actions → "Docker" → "Run workflow").

Regular pushes and pull requests run tests ([`ci.yml`](https://github.com/marco-montesines/haveibeenpwned/blob/master/.github/workflows/ci.yml)) but never build or publish images. This keeps `:latest` meaning "latest release", not "latest commit".

## Verifying a release

```bash
docker run --rm -p 8080:8080 ghcr.io/marco-montesines/haveibeenpwned:latest &
curl -s localhost:8080/healthz
curl -s -X POST localhost:8080/v1/pwnedpassword -d '{"password":"P@ssw0rd"}'
# expect {"count":6421042,"pwned":true} — the count grows over time
```

## Where to find artifacts

- **Images**: repo page → Packages sidebar, or directly at [ghcr.io/marco-montesines/haveibeenpwned](https://github.com/marco-montesines/haveibeenpwned/pkgs/container/haveibeenpwned).
- **Go module**: [pkg.go.dev](https://pkg.go.dev/github.com/marco-montesines/haveibeenpwned).
- **Tags**: [github.com/marco-montesines/haveibeenpwned/tags](https://github.com/marco-montesines/haveibeenpwned/tags).
