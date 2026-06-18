# servasec — Architecture & Tests

## What was done

### Security audit (SECURITY-ANALYSIS-001.md)
- 22 Go files scanned with Opengrep (15 custom rules, 0 community rule matches)
- 3 critical (CWE-330, CWE-798, CWE-613) + 6 P1–P3 issues, all fixed

### Phase 1 — New data models

| File | Model | Notes |
|------|-------|-------|
| `backend/models/group.go` | `Group` | Flat (no nesting), `Path` is unique |
| `backend/models/application.go` | `Application` | Belongs to `Group`, auto-generated `ApiToken` (64 hex chars) |
| `backend/models/scan.go` | `Scan` | Ingest-only, no `os/exec` — CI/CD pushes results via API |
| `backend/models/finding.go` | `Finding` | Linked to `Scan` + `Application`, severity/status indexed |
| `backend/models/team.go` | `Team` | Named group of users |
| `backend/models/team_member.go` | `TeamMember` | Join table with `Role` (admin/member), unique per team+user |

### Phase 2 — Casbin resource-level RBAC

**`backend/config/casbin_model.conf`**
- Added `[role_definition]` with `g = _, _`
- Changed matcher from `r.sub == p.sub` to `g(r.sub, p.sub)`

**`backend/config/casbin_policies.csv`** — expanded from 8 to 18:

| Policy | Subject |
|--------|---------|
| `/login`, `/register`, `/refresh`, `/csrf-token` | `anonymous` |
| `/logout`, `/me`, `/dashboard` | `member` |
| `/groups`, `/groups/*`, `/applications`, `/applications/*`, `/teams`, `/teams/*`, `/scans`, `/scans/*`, `/findings`, `/findings/*` | `member` |
| `/*` | `admin` |

**`backend/middleware/resource_access.go`** — new middleware
- `RequireResourceAccess(path, action)` — checks `user:{id}` then each `team:{id}` via Casbin `Enforce`
- `RequireResourceAccessByParam(type, param, action)` — extracts `:id` from URL, builds path
- Admin bypasses resource-level check

### Phase 3 — Controllers & routes

| File | Endpoints | Auth |
|------|-----------|------|
| `controllers/group.go` | CRUD `/groups` | JWT + CSRF + resource-level on `:id` |
| `controllers/application.go` | CRUD `/applications` + regenerate-token | JWT + CSRF + resource-level on `:id` |
| `controllers/team.go` | CRUD `/teams` + add/remove members | JWT + CSRF |
| `controllers/scan.go` | `POST /scans/ingest`, `GET /scans`, `GET /scans/:id` | ApiToken header for ingest, JWT for list |
| `controllers/finding.go` | `GET /findings` (filtered), `GET /findings/:id`, `PATCH .../status` | JWT |

### Auto‑grant on creation
When a user creates a Group or Application, the controller injects a Casbin policy:

```
p, user:{creatorID}, /groups/{id}, *
p, user:{creatorID}, /applications/{id}, *
```

This ensures the `RequireResourceAccess` middleware allows access.

---

## Manual test plan (curl)

All tests assume `make dev` is running (backend on `localhost:8080`, Caddy on `:8080`).

### 1. Setup: register + login

```bash
# Register
curl -s -c /tmp/cookies.txt -b /tmp/cookies.txt \
  -X POST http://localhost:8080/register \
  -H 'Content-Type: application/json' \
  -d '{"username":"testuser","email":"test@example.com","password":"testpass123"}'

# Register a second user
curl -s -c /tmp/cookies2.txt -b /tmp/cookies2.txt \
  -X POST http://localhost:8080/register \
  -H 'Content-Type: application/json' \
  -d '{"username":"user2","email":"user2@example.com","password":"testpass123"}'

# Login (saves cookies)
curl -s -c /tmp/cookies.txt -b /tmp/cookies.txt \
  -X POST http://localhost:8080/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"testuser","password":"testpass123"}'

curl -s -c /tmp/cookies2.txt -b /tmp/cookies2.txt \
  -X POST http://localhost:8080/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"user2","password":"testpass123"}'
```

### 2. CSRF token (needed for mutating endpoints)

```bash
CSRF=$(curl -s -c /tmp/cookies.txt -b /tmp/cookies.txt \
  http://localhost:8080/csrf-token | jq -r '.csrfToken')
echo $CSRF
```

### 3. Groups

```bash
# Create group (user1)
curl -s -c /tmp/cookies.txt -b /tmp/cookies.txt \
  -X POST http://localhost:8080/groups \
  -H 'Content-Type: application/json' \
  -H "X-CSRF-Token: $CSRF" \
  -d '{"name":"My App Group","description":"Test","path":"my-app-group"}'

# List groups
curl -s -c /tmp/cookies.txt -b /tmp/cookies.txt \
  http://localhost:8080/groups

# Get group by ID
curl -s -c /tmp/cookies.txt -b /tmp/cookies.txt \
  http://localhost:8080/groups/1

# User2 tries to access user1's group (should 403)
curl -s -c /tmp/cookies2.txt -b /tmp/cookies2.txt \
  http://localhost:8080/groups/1
```

### 4. Applications

```bash
# Create application (user1)
APP=$(curl -s -c /tmp/cookies.txt -b /tmp/cookies.txt \
  -X POST http://localhost:8080/applications \
  -H 'Content-Type: application/json' \
  -H "X-CSRF-Token: $CSRF" \
  -d '{"name":"Web API","description":"Main API","slug":"web-api","groupId":1,"repositoryUrl":"https://github.com/org/repo"}')
echo $APP
API_TOKEN=$(echo $APP | jq -r '.apiToken')

# Regenerate API token
curl -s -c /tmp/cookies.txt -b /tmp/cookies.txt \
  -X POST http://localhost:8080/applications/1/regenerate-token \
  -H "X-CSRF-Token: $CSRF"

# User2 tries to get app 1 (should 403)
curl -s -c /tmp/cookies2.txt -b /tmp/cookies2.txt \
  http://localhost:8080/applications/1
```

### 5. Scan ingestion (via API token, no JWT)

```bash
# Ingest scan results
curl -s -X POST http://localhost:8080/scans/ingest \
  -H "X-Api-Token: $API_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{
    "scannerType": "gitleaks",
    "status": "completed",
    "findings": [
      {
        "ruleId": "Gitleaks-AWS-Key",
        "title": "AWS Access Key",
        "severity": "high",
        "description": "Hardcoded AWS key in config file",
        "filePath": "src/config/aws.go",
        "lineStart": 42,
        "lineEnd": 42,
        "cweId": "CWE-798",
        "remediation": "Use environment variables or secrets manager"
      }
    ]
  }'

# List scans (authenticated)
curl -s -c /tmp/cookies.txt -b /tmp/cookies.txt \
  http://localhost:8080/scans

# List scans filtered by application
curl -s -c /tmp/cookies.txt -b /tmp/cookies.txt \
  "http://localhost:8080/scans?applicationId=1"
```

### 6. Findings

```bash
# List findings
curl -s -c /tmp/cookies.txt -b /tmp/cookies.txt \
  http://localhost:8080/findings

# With filters
curl -s -c /tmp/cookies.txt -b /tmp/cookies.txt \
  "http://localhost:8080/findings?severity=high&status=open"

# Triage: confirm finding
curl -s -c /tmp/cookies.txt -b /tmp/cookies.txt \
  -X PATCH http://localhost:8080/findings/1/status \
  -H 'Content-Type: application/json' \
  -d '{"status":"confirmed"}'

# Triage: mark as fixed
curl -s -c /tmp/cookies.txt -b /tmp/cookies.txt \
  -X PATCH http://localhost:8080/findings/1/status \
  -H 'Content-Type: application/json' \
  -d '{"status":"fixed"}'
```

### 7. Teams

```bash
# Create team (user1)
curl -s -c /tmp/cookies.txt -b /tmp/cookies.txt \
  -X POST http://localhost:8080/teams \
  -H 'Content-Type: application/json' \
  -H "X-CSRF-Token: $CSRF" \
  -d '{"name":"Security Team","description":"Security engineers"}'

# List teams
curl -s -c /tmp/cookies.txt -b /tmp/cookies.txt \
  http://localhost:8080/teams

# Add user2 to team
curl -s -c /tmp/cookies.txt -b /tmp/cookies.txt \
  -X POST http://localhost:8080/teams/1/members \
  -H 'Content-Type: application/json' \
  -H "X-CSRF-Token: $CSRF" \
  -d '{"userId":2,"role":"member"}'

# Remove user2 from team
curl -s -c /tmp/cookies.txt -b /tmp/cookies.txt \
  -X DELETE http://localhost:8080/teams/1/members/2 \
  -H "X-CSRF-Token: $CSRF"
```

### 8. Admin: user management

```bash
# Get CSRF token as admin
# (login as admin first — SSC_ADMIN_PASSWORD env var)
curl -s -c /tmp/admin-cookies.txt -b /tmp/admin-cookies.txt \
  -X POST http://localhost:8080/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"your_admin_password"}'

# List users
curl -s -c /tmp/admin-cookies.txt -b /tmp/admin-cookies.txt \
  http://localhost:8080/users

# Ban user
curl -s -c /tmp/admin-cookies.txt -b /tmp/admin-cookies.txt \
  -X PUT http://localhost:8080/users/2 \
  -H 'Content-Type: application/json' \
  -d '{"banned":true}'
```

### 9. Error cases to test

```bash
# No auth → 401
curl -s http://localhost:8080/groups

# Invalid API token → 401
curl -s -X POST http://localhost:8080/scans/ingest \
  -H "X-Api-Token: invalid" \
  -H 'Content-Type: application/json' \
  -d '{"scannerType":"test"}'

# Missing CSRF on mutating endpoint → 403
curl -s -c /tmp/cookies.txt -b /tmp/cookies.txt \
  -X POST http://localhost:8080/groups \
  -H 'Content-Type: application/json' \
  -d '{"name":"No CSRF"}'  

# Invalid finding status → 400
curl -s -c /tmp/cookies.txt -b /tmp/cookies.txt \
  -X PATCH http://localhost:8080/findings/1/status \
  -H 'Content-Type: application/json' \
  -d '{"status":"invalid"}'

# Duplicate team member → 409
curl -s -c /tmp/cookies.txt -b /tmp/cookies.txt \
  -X POST http://localhost:8080/teams/1/members \
  -H 'Content-Type: application/json' \
  -H "X-CSRF-Token: $CSRF" \
  -d '{"userId":1,"role":"member"}'
```

---

## What's next (Phase 4–7)

| Phase | Scope |
|-------|-------|
| **4** | Permissions API — endpoints to manage Casbin policies via UI (grant/revoke) |
| **5** | Scan parser — convert Opengrep/Gitleaks/Nuclei JSON into structured Findings |
| **6** | Enriched dashboard — stats by severity, scanner, trend over time |
| **7** | Webhook notifications — notify on new high/critical findings |

## Key architecture decisions

| Decision | Rationale |
|----------|-----------|
| Groups are flat (no `ParentID`) | Simplicity; nesting can be added later as a separate hierarchy table if needed |
| Scans are ingest-only | Backend never executes scanners; CI/CD pushes results. Keeps backend stateless |
| Casbin for all permissions (no `ResourcePermission` table) | Casbin already handles RBAC + grouping (`g`). Avoids dual-authorization logic |
| `user:{id}` / `team:{id}` subjects | Allows fine-grained resource policies alongside existing role-based policies |
| Auto-grant on create | Creator gets implicit `*` policy on the new resource; avoids manual policy setup |
| Scan ingestion uses `X-Api-Token` | CI/CD pipelines don't have user JWT tokens; each app has its own API token |
