# syntax=docker/dockerfile:1
# Builds FrankenPHP with the hibp PHP extension (written in Go) compiled in,
# then serves the demo app from frankenphp/example.
#
# Build from the repository root:
#   docker build -f frankenphp/Dockerfile -t hibp-frankenphp .

FROM dunglas/frankenphp:builder AS builder

COPY --from=caddy:builder /usr/bin/xcaddy /usr/bin/xcaddy

WORKDIR /src
COPY . .

ENV CGO_ENABLED=1 \
    XCADDY_SETCAP=1 \
    XCADDY_GO_BUILD_FLAGS="-ldflags='-w -s' -tags=nobadger,nomysql,nopgx"

RUN CGO_CFLAGS="$(php-config --includes)" \
    CGO_LDFLAGS="$(php-config --ldflags) $(php-config --libs)" \
    xcaddy build \
      --output /usr/local/bin/frankenphp \
      --with github.com/dunglas/frankenphp/caddy \
      --with github.com/marco-montesines/haveibeenpwned/frankenphp/extension=./frankenphp/extension \
      --with github.com/marco-montesines/haveibeenpwned=.

FROM dunglas/frankenphp

COPY --from=builder /usr/local/bin/frankenphp /usr/local/bin/frankenphp
COPY frankenphp/Caddyfile /etc/frankenphp/Caddyfile
COPY frankenphp/example /app/public
