# servasec

**Application Security Posture Management (ASPM)** platform that aggregates SAST/DAST/SCA scanner findings into a unified dashboard for triaging, tracking, and remediating security vulnerabilities across your application portfolio.

## Features

| Feature | Free | Pro |
|---------|:----:|:---:|
| Findings management & lifecycle | ✓ | ✓ |
| Dashboard with enriched stats | ✓ | ✓ |
| Team-based collaboration | ✓ | ✓ |
| Webhook notifications | ✓ | ✓ |
| Role-based access control (Casbin) | ✓ | ✓ |
| Resource-level permissions | ✓ | ✓ |
| Scanner ingest (Semgrep, Trivy, etc.) | ✓ | ✓ |
| Version comparison | ✓ | ✓ |
| SSO / OIDC (planned) | ? | ✓ |
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

Default admin: `admin` / password from `SSC_ADMIN_PASSWORD` (random hex if unset, check logs of backend container ;) ).

## Scanner Support

| Scanner | Type | Parser |
|---------|------|--------|
| Semgrep | SAST | ✅ Implemented |
| Trivy | Vulnerability | ✅ Implemented |
| Gitleaks | Secrets | ✅ Implemented |
| Grype | Vulnerability | ✅ Implemented |
| Snyk | SCA | ⏳ In progress |
| Checkov | IaC | ⏳ In progress |

## Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `DATABASE_URL` | ✓ | - | PostgreSQL connection string |
| `JWT_SECRET` | ✓ | - | Access token signing key |
| `REFRESH_SECRET` | ✓ | - | Refresh token signing key |
| `CSRF_SECRET` | ✓ | (auto) | CSRF protection key |
| `SSC_ADMIN_PASSWORD` | | random | Initial admin password (8 chars min.) |
| `SSC_SITE_NAME` | | `servasec` | Site name |
| `SSC_PUBLIC_DOMAIN` | ✓ | - | Public hostname (for CORS & Caddy) |
| `SSC_REGISTRATION_ENABLED` | | `true` | Allow user registration |
| `SSC_DEBUG_ENABLED` | | `false` | Enable debug logging |
| `SSC_SEED_DATABASE` | | `false` | Seed default admin + scanner types |
| `SSC_SEED_CASBIN_CSV` | | `false` | Seed Casbin policies from CSV |
| `SSC_LICENSE_KEY` | | - | License JWT for pro features |
| `TRUSTED_PROXIES_CIDR` | | `172.70.1.0/24` | Caddy Docker network CIDR |


## License

**servasec** is dual-licensed:

- **AGPLv3** - Free, open-source. All standard features included.
- **Commercial License** - Required for pro features (audit log, SSO, SLA management, advanced reporting).

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
