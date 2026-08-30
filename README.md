<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="frontend/public/assets/servasec-mark.svg">
    <img src="frontend/public/assets/servasec-mark.svg" alt="servasec open source ASPM" width="100" height="100">
  </picture>
</p>

<h1 align="center">servasec — Open Source ASPM Platform</h1>

<p align="center">
<picture><source media="(prefers-color-scheme: dark)" srcset="https://www.shieldcn.dev/github/stars/servasec/servasec.svg?variant=secondary&amp;size=sm&amp;mode=dark"><img alt="GitHub Stars" src="https://www.shieldcn.dev/github/stars/servasec/servasec.svg?variant=secondary&amp;size=sm&amp;mode=light"></picture>
<a style="text-decoration:none" href="https://github.com/servasec/servasec/releases" target="_blank"><picture><source media="(prefers-color-scheme: dark)" srcset="https://www.shieldcn.dev/github/release/servasec/servasec.svg?size=sm&amp;mode=dark"><img alt="Release" src="https://www.shieldcn.dev/github/release/servasec/servasec.svg?size=sm&amp;mode=light"></picture></a>
<a href="https://discord.com/invite/wJZ2v5YNtD" target="_blank"><picture><source media="(prefers-color-scheme: dark)" srcset="https://shieldcn.dev/discord/members/wJZ2v5YNtD.svg?variant=secondary&amp;mode=dark"><img alt="badge" src="https://shieldcn.dev/discord/members/wJZ2v5YNtD.svg?variant=secondary&amp;mode=light"></picture></a>
<a href="https://docs.servasec.com" target="_blank"><picture><source media="(prefers-color-scheme: dark)" srcset="https://shieldcn.dev/badge/docs.servasec.com.svg?variant=ghost&color=6f4599"><img alt="badge" src="https://shieldcn.dev/badge/docs.servasec.com.svg?variant=ghost&color=6f4599"></picture></a>
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
| Issue Tracker (GitHub, GitLab) | ✓ | ✓ |
| Resource-level permissions | ✓ | ✓ |
| Scanner ingest (Semgrep, Trivy, etc.) | ✓ | ✓ |
| Version comparison | ✓ | ✓ |
| SSO / OIDC | | ✓ |
| Audit log | | ✓ |
| MCP Server | | ✓ |
| SLA management (planned) | | ✓ |
| Advanced reporting (planned) | | ✓ |

## Quick Start

```bash
curl -fsSL https://servasec.com/install.sh | sh
```

Default admin: `admin` / password displayed at the end of the install.

**Interactive mode** (prompts for admin password, domain, registration):

```bash
curl -fsSL https://servasec.com/install.sh | sh -s -- -i
```

**Build from source** (local build instead of pulling published images):

```bash
curl -fsSL https://servasec.com/install.sh | sh -s -- --local-build
```

**Install with pro features**:

```bash
# Pull mode: uses the private backend-pro image (authenticate first)
docker login registry.gitlab.com
curl -fsSL https://servasec.com/install.sh | sh -s -- --pro

# Local build mode: requires the servasec-pro repo
curl -fsSL https://servasec.com/install.sh | sh -s -- --local-build --pro --pro-repo ../servasec-pro
```

The installer downloads only the deploy files, then pulls the container images
from `registry.gitlab.com` (no full git clone). Prefix the image tag with
`--version <tag>` to install a specific release.

**Manual setup** (without the install script):

```bash
git clone https://github.com/servasec/servasec.git
cd servasec
cp .env.example .env
# Edit secrets (JWT_SECRET, REFRESH_SECRET, CSRF_SECRET, SSC_ADMIN_PASSWORD)

# Pull the published images from registry.gitlab.com
make community

# Or build from source instead
make community-build
```

## Scanner Support

Look at https://docs.servasec.com/scanners/overview/ for the complete list of scanners supported.

## Frequently Asked Questions

### Is servasec free and open source?

Yes, servasec is licensed under AGPLv3. Self-host it for free, forever. A separate commercial license is required only for Pro features (audit log, MCP Server, SSO, etc).

### Which scanners does servasec support?

servasec ingests Semgrep, Trivy, Gitleaks, OSV-Scanner, tfsec, SARIF exports and a lot more ! See [supported scanners](https://servasec.com/scanners) for details.

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
