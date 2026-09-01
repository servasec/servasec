#!/bin/sh
set -e

if [ -n "$SSC_PUBLIC_DOMAIN" ]; then
    if [ "${SSC_EDGE_MODE:-false}" = "true" ]; then
        # Derrière le reverse proxy principal (qui termine le TLS) : servir le
        # site en HTTP sur :80 seulement. Le proxy préserve le Host header, donc
        # le routage par domaine ci-dessous reste valide.
        sed -i "s/%SSC_PUBLIC_DOMAIN%/http:\/\/$SSC_PUBLIC_DOMAIN:80/" /etc/caddy/Caddyfile
    else
        sed -i "s/%SSC_PUBLIC_DOMAIN%/$SSC_PUBLIC_DOMAIN/" /etc/caddy/Caddyfile
    fi
fi

exec caddy run --config /etc/caddy/Caddyfile
