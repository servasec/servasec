<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="frontend/public/assets/servasec-mark.svg">
    <img src="frontend/public/assets/servasec-mark.svg" alt="servasec" width="100" height="100">
  </picture>
</p>

<h1 align="center">servasec</h1>

<p align="center">
  <a href="./LICENSE"><img src="https://img.shields.io/badge/license-AGPLv3%20%7C%20Commercial-blue" alt="License"></a>
  <a href="https://github.com/servasec/servasec/releases"><img src="https://img.shields.io/github/v/release/servasec/servasec" alt="Release"></a>
  <a href="https://github.com/servasec/servasec/stargazers"><img src="https://img.shields.io/github/stars/servasec/servasec" alt="GitHub Stars"></a>
  <a href="https://discord.gg/jqFmBEPQz"><img src="https://img.shields.io/badge/chat-Discord-5865F2" alt="Discord"></a>
</p>

**Application Security Posture Management (ASPM)** platform that aggregates SAST/DAST/SCA scanner findings into a unified dashboard for triaging, tracking, and remediating security vulnerabilities across your application portfolio.

## Features

| Feature | Free | Pro |
|---------|:----:|:---:|
| Findings management & lifecycle | ✓ | ✓ |
| Dashboard with enriched stats | ✓ | ✓ |
| Team-based collaboration | ✓ | ✓ |
| Webhook notifications | ✓ | ✓ |
| RBAC | ✓ | ✓ |
| Resource-level permissions | ✓ | ✓ |
| Scanner ingest (Semgrep, Trivy, etc.) | ✓ | ✓ |
| Version comparison | ✓ | ✓ |
| SSO / OIDC | ? | ✓ |
| Audit log |  | ✓ |
| MCP Server |  | ✓ |
| SLA management (planned) |  | ✓ |
| Advanced reporting (planned) |  | ✓ |

## Quick Start

```bash
cp .env.example .env
# Edit secrets (JWT_SECRET, REFRESH_SECRET, CSRF_SECRET, SSC_ADMIN_PASSWORD)
make prod
# You're done ! Visit https://servasec.local
```

Default admin: `admin` / password from `SSC_ADMIN_PASSWORD` env variable.


## Scanner Support

Look at https://docs.servasec.com/scanners/overview/ for the complete list of scanners supported.

## Environment Variables

See https://docs.servasec.com/getting-started/configuration/

## License

**servasec** is dual-licensed:

- **AGPLv3** - Free, open-source. All standard features included.
- **Commercial License** - Required for pro features (audit log, MCP, SLA management, advanced reporting).

See [`LICENSE`](./LICENSE) and [`COMMERCIAL_LICENSE.md`](./COMMERCIAL_LICENSE.md).

## Development

```bash
# Start dev stack with hot-reload
make dev

# View logs
make logs

# Stop
make down
```

Requires: Docker, Docker Compose, `make`.
