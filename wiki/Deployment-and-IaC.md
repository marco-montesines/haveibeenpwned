# Deployment and IaC

How to run the HTTP API server (`hibp serve`) as a service — from a bare `docker run` up to infrastructure-as-code. The repository ships Docker + Compose; the Kubernetes/systemd/Terraform recipes below are equivalent replacements you can copy into your own IaC.

## The contract (all deployment methods)

| Aspect | Value |
| ------ | ----- |
| Image | `ghcr.io/marco-montesines/haveibeenpwned` (multi-arch: linux/amd64, linux/arm64) |
| Tags | `:latest`, `:1`-style semver (`:1.0.3`, `:1.0`), `:sha-<commit>` — see [[Releases and Versioning|Releases-and-Versioning]] |
| Listen port | `8080` |
| Health check | `GET /healthz` → `{"status":"ok"}` |
| Configuration | `HIBP_API_KEY` env var (optional — required only for the account endpoints) |
| State | None — the server is stateless; scale horizontally at will |
| Shutdown | Graceful on SIGTERM/SIGINT (10s drain) |

Treat `HIBP_API_KEY` as a secret everywhere: secret store, not plain env in committed files.

## Docker (plain)

```bash
docker run -d --name hibp \
  -p 8080:8080 \
  -e HIBP_API_KEY=... \
  --restart unless-stopped \
  ghcr.io/marco-montesines/haveibeenpwned:1.0.3
```

Pin a version tag in anything long-lived; use `:latest` only for experiments.

## Docker Compose

The repo's [`docker-compose.yml`](https://github.com/marco-montesines/haveibeenpwned/blob/master/docker-compose.yml) starts the API plus the FrankenPHP demo. A minimal production-style service for your own stack:

```yaml
services:
  hibp-api:
    image: ghcr.io/marco-montesines/haveibeenpwned:1.0.3
    restart: unless-stopped
    environment:
      HIBP_API_KEY: ${HIBP_API_KEY}
    # No `ports:` — reach it via the Compose network as http://hibp-api:8080
    healthcheck:
      test: ["CMD", "wget", "-qO-", "http://localhost:8080/healthz"]
      interval: 30s
      timeout: 3s
      retries: 3
```

Other services on the same Compose network call `http://hibp-api:8080/v1/...` — no host port exposure needed.

## Kubernetes

Equivalent manifests (Deployment + Service + Secret). The service is stateless, so replicas are safe — but remember all replicas share one HIBP API key and therefore one upstream rate limit.

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: hibp-api-key
stringData:
  HIBP_API_KEY: "<your key>"   # in real IaC: sealed-secrets / external-secrets / SOPS
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: hibp-api
spec:
  replicas: 2
  selector:
    matchLabels: { app: hibp-api }
  template:
    metadata:
      labels: { app: hibp-api }
    spec:
      containers:
        - name: hibp-api
          image: ghcr.io/marco-montesines/haveibeenpwned:1.0.3
          ports:
            - containerPort: 8080
          envFrom:
            - secretRef: { name: hibp-api-key }
          readinessProbe:
            httpGet: { path: /healthz, port: 8080 }
          livenessProbe:
            httpGet: { path: /healthz, port: 8080 }
          resources:
            requests: { cpu: 10m, memory: 32Mi }
            limits: { memory: 128Mi }
---
apiVersion: v1
kind: Service
metadata:
  name: hibp-api
spec:
  selector: { app: hibp-api }
  ports:
    - port: 80
      targetPort: 8080
```

Keep the Service `ClusterIP` (internal). If you must expose it, do so behind an authenticating ingress/gateway — the server itself has no auth or TLS (see the deployment note in [[HTTP API Reference|HTTP-API-Reference]]). A Helm chart / Kustomize base is [[under consideration|Roadmap]] — until then, these manifests are the template.

## Terraform (example)

Any Terraform provider that runs containers works — the image is a plain OCI image on GHCR (public, no pull auth). Docker provider example:

```hcl
resource "docker_image" "hibp" {
  name = "ghcr.io/marco-montesines/haveibeenpwned:1.0.3"
}

resource "docker_container" "hibp_api" {
  name    = "hibp-api"
  image   = docker_image.hibp.image_id
  restart = "unless-stopped"
  env     = ["HIBP_API_KEY=${var.hibp_api_key}"]   # var marked sensitive = true
  ports {
    internal = 8080
    external = 8080
  }
}
```

The same contract (image/port/env/healthcheck) maps directly onto ECS task definitions, Cloud Run services, Nomad jobs, etc.

## systemd (no containers)

```ini
# /etc/systemd/system/hibp-api.service
[Unit]
Description=HIBP HTTP JSON API
After=network-online.target
Wants=network-online.target

[Service]
ExecStart=/usr/local/bin/hibp serve -addr 127.0.0.1:8080
EnvironmentFile=/etc/hibp-api.env        # HIBP_API_KEY=... (chmod 600)
DynamicUser=yes
NoNewPrivileges=yes
ProtectSystem=strict
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

Install the binary with `go install github.com/marco-montesines/haveibeenpwned/cmd/hibp@latest` (or build with `go build ./cmd/hibp`), then `systemctl enable --now hibp-api`.

## Building images yourself

```bash
docker build -t hibp .                                   # the API server image (what GHCR hosts)
docker build -f frankenphp/Dockerfile -t hibp-frankenphp .   # FrankenPHP demo (slow: compiles PHP + extension)
```

## Upgrades

Bump the image tag (new versions appear on [[releases|Releases-and-Versioning]]), roll the deployment. The server is stateless with graceful shutdown, so rolling updates are zero-downtime as long as you run ≥2 replicas behind a load balancer or Service.
