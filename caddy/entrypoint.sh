#!/bin/sh
set -e

if [ -n "$SSC_PUBLIC_DOMAIN" ]; then
    sed -i "s/%SSC_PUBLIC_DOMAIN%/$SSC_PUBLIC_DOMAIN/" /etc/caddy/Caddyfile
fi

exec caddy run --config /etc/caddy/Caddyfile
