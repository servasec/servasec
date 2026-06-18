FROM caddy:2.10-alpine

ARG CADDY_ENV=default
ARG SSC_PUBLIC_DOMAIN
ENV SSC_PUBLIC_DOMAIN=${SSC_PUBLIC_DOMAIN}

COPY ${CADDY_ENV}.Caddyfile /etc/caddy/Caddyfile

RUN sed -i "s/%SSC_PUBLIC_DOMAIN%/$SSC_PUBLIC_DOMAIN/" /etc/caddy/Caddyfile
