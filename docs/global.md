# servasec — Documentation Générale

Application Security Posture Management (ASPM) platform.

**Stack :** Go 1.24+ / Gin / GORM / Casbin / PostgreSQL / Next.js Pages Router / Tailwind CSS / shadcn-ui / Caddy

---

## Table des matières

1. [Démarrage rapide](#1-démarrage-rapide)
2. [Architecture](#2-architecture)
3. [Authentification & Sécurité](#3-authentification--sécurité)
4. [Modèles de données](#4-modèles-de-données)
5. [RBAC & Permissions (Casbin)](#5-rbac--permissions-casbin)
6. [Endpoints API](#6-endpoints-api)
7. [Scan Ingestion](#7-scan-ingestion)
8. [Cycle de Vie des Findings](#8-cycle-de-vie-des-findings)
9. [Webhooks](#9-webhooks)
10. [Dev Setup](#10-dev-setup)
11. [CI/CD](#11-cicd)
12. [Audit de Sécurité](#12-audit-de-sécurité)

---

## 1. Démarrage rapide

```bash
cp .env.example .env   # éditer les secrets
make dev               # lance tout via docker-compose.dev.yml
```

Services exposés :

| Service | URL |
|---------|-----|
| Frontend | `https://servasec.local` |
| Backend API | `https://servasec.local/api/*` (Caddy reverse proxy) |
| PostgreSQL | `localhost:5432` |

Caddy sert le frontend Next.js sur `/` et forwarde `/api/*` vers le backend Go (en stripant le préfixe `/api`). Les sessions transfrontières (frontend → API même domaine) évitent les CORS.

---

## 2. Architecture

```
                    ┌─────────────┐
                    │   Browser   │
                    └──────┬──────┘
                           │ https://servasec.local
                    ┌──────▼──────┐
                    │    Caddy    │  (reverse proxy, TLS auto)
                    └──┬──────┬───┘
                       │      │
              /api/*   │      │  /*
              strip    │      │
              /api     │      │
                ┌──────▼──┐ ┌─▼───────────┐
                │  Backend│ │   Frontend   │
                │  Go/Gin │ │ Next.js SPA  │
                └──┬──────┘ └─────────────┘
                   │
              ┌────▼────┐
              │  Postgres│
              └─────────┘
```

### Backend (`backend/`)

```
backend/
├── config/          # DB, Casbin, seeder
├── controllers/     # Handlers Gin
├── dto/             # Structs for binding/validation
├── middleware/      # Auth, CSRF, rate limit, resource access
├── models/          # GORM models
├── parsers/         # Scanner output parsers (registry pattern)
├── routes/          # Route registration
├── utils/           # JWT, tokens, blacklist, secrets, response helpers
├── main.go          # Entrypoint, router setup
└── air.toml         # Hot-reload config
```

### Frontend (`frontend/`)

```
frontend/
├── pages/                  # Next.js Pages Router
│   ├── index.tsx           # Dashboard
│   ├── login.tsx           # Auth
│   ├── applications/       # CRUD + détails + versions + ingest
│   ├── findings/           # Liste + détail + assign/review/comments
│   ├── scans.tsx           # Liste filtrée
│   ├── groups.tsx          # CRUD groups
│   ├── teams/              # CRUD teams + membres
│   └── admin/              # Users management, permissions
├── components/ui/          # shadcn-ui components
├── lib/api.ts              # Axios instance (CSRF + refresh interceptors)
├── styles/globals.css      # Tailwind + CSS variables
└── tailwind.config.ts
```

---

## 3. Authentification & Sécurité

### JWT (access + refresh tokens)

| Token | Stockage | Expiration | Usage |
|-------|----------|------------|-------|
| `access_token` | Cookie HTTP-only | 15 min | Authentifie chaque requête API |
| `refresh_token` | Cookie HTTP-only | 7 jours | Renouvelle l'access token |

Les deux tokens sont signés avec des secrets distincts (`JWT_SECRET`, `REFRESH_SECRET`) validés au démarrage. Aucun fallback `"secret"` n'existe.

### CSRF Protection

- Double submit cookie : un cookie `_gorilla_csrf` + header `X-CSRF-Token` (ou `X-Csrf-Token`)
- Le middleware CSRF compare les deux valeurs — lecture du cookie, pas de session serveur
- Toutes les routes mutantes (POST/PUT/PATCH/DELETE) sont protégées sauf `/login`, `/register`, `/refresh`, `/csrf-token` (anonymous)

### Rate Limiting

- Per-IP avec un token bucket (`golang.org/x/time/rate`)
- Configurable via `SSC_RATE_LIMIT_RPS` (défaut : 10 req/s) et `SSC_RATE_LIMIT_BURST` (défaut : 20)

### Blacklist de Tokens

PostgreSQL table `blacklisted_tokens` (pas de map mémoire) :
- Les refresh tokens révoqués sont stockés avec leur expiration
- Un goroutine de cleanup supprime les entrées expirées toutes les 5 minutes

### Sécurité des Cookies

- `HttpOnly` : oui (empêche JS de lire le token)
- `Secure` : oui (HTTPS only)
- `SameSite` : `Strict` (pas envoyé en cross-site)
- `Path` : `/api` (scope limité)

### Initialisation

`ValidateSecrets()` dans `main.go` vérifie au démarrage :
- `JWT_SECRET`, `REFRESH_SECRET`, `CSRF_SECRET` sont définis et ≠ `"secret"`
- `SSC_ADMIN_PASSWORD` ≥ 8 caractères (ou fatal)

---

## 4. Modèles de données

### Relations

```
User (1) ──< TeamMember >── (1) Team
  │
  ├── (1) ──< Finding (assignedTo)
  ├── (1) ──< Finding (reviewedBy)
  └── (1) ──< Comment

Group (1) ──< Application (1) ──< ApplicationVersion (1) ──< Scan (1) ──< Finding
                                   ApplicationVersion (1) ──< Finding (direct shortcut)
                                   Application (1) ──< Webhook

ScannerType (1) ──< Scan
ScannerType (1) ──< Finding
```

### Modèles principaux

#### `User`

| Champ | Type | Notes |
|-------|------|-------|
| `ID` | uint | PK auto |
| `Username` | string | Unique, 3-50 chars |
| `Email` | string | Unique |
| `Password` | string | bcrypt hash |
| `Role` | string | `member` ou `admin` |
| `Banned` | bool | |

#### `Group`

| Champ | Type |
|-------|------|
| `ID` | uint |
| `Name` | string |
| `Description` | string |
| `Path` | string (unique) |

Flat (pas de `ParentID`). La hiérarchie peut être ajoutée plus tard via une table séparée.

#### `Application`

| Champ | Type | Notes |
|-------|------|-------|
| `ID` | uint | PK |
| `Name` | string | max 200 |
| `Slug` | string | unique, utilisé dans CI/CD |
| `GroupID` | uint | FK → Group |
| `RepositoryURL` | string | optionnel |
| `ApiToken` | string | 64 hex chars, regenable |
| `Versions` | []ApplicationVersion | has-many, preloaded |

#### `ApplicationVersion`

| Champ | Type | Notes |
|-------|------|-------|
| `ID` | uint | PK |
| `ApplicationID` | uint | FK, unique avec Name |
| `Name` | string | ex: "v1.0.0", "develop" |
| `Branch` | string | optionnel |
| `Tag` | string | semver tag, optionnel |
| `IsDefault` | bool | version par défaut |

Unicité `(ApplicationID, Name)`. Upserté automatiquement pendant l'ingestion.

#### `Scan`

| Champ | Type | Notes |
|-------|------|-------|
| `ID` | uint | |
| `ApplicationVersionID` | uint | FK → ApplicationVersion |
| `ScannerTypeID` | uint | FK → ScannerType |
| `Status` | string | `completed`, `failed` |
| `ScanDate` | time.Time | |
| `Findings` | []Finding | has-many |

#### `Finding`

| Champ | Type | Notes |
|-------|------|-------|
| `ID` | uint | |
| `ScanID` | uint | FK → Scan |
| `ApplicationVersionID` | uint | FK → ApplicationVersion |
| `ScannerTypeID` | uint | FK → ScannerType |
| `RuleID` | string | |
| `Title` | string | |
| `Severity` | string | `critical`, `high`, `medium`, `low` |
| `Status` | string | `open`, `confirmed`, `fixed`, `false_positive`, `closed` |
| `FilePath` | string | |
| `LineStart` / `LineEnd` | int | |
| `CweID` | string | ex: `"CWE-79: ..."` |
| `Remediation` | text | |
| `Description` | text | |
| `AssignedTo` | *uint | FK → User (nullable) |
| `DueDate` | *time.Time | nullable |
| `ReviewedBy` | *uint | FK → User (nullable) |
| `FixedAt` | *time.Time | set automatiquement quand status → `fixed` |

Index sur `(severity, status)` pour les filtres dashboard.

#### `ScannerType`

| Champ | Type |
|-------|------|
| `ID` | uint |
| `Name` | string |
| `Description` | string |
| `Parser` | string (identifiant du handler Go) |

6 types seedés : `semgrep`, `trivy`, `gitleaks`, `grype`, `snyk`, `checkov`.

#### `Team`

| Champ | Type |
|-------|------|
| `ID` | uint |
| `Name` | string, max 100 |
| `Description` | string, max 500 |
| `Members` | []TeamMember | has-many, preloaded dans GetTeam |

#### `TeamMember`

| Champ | Type | Notes |
|-------|------|-------|
| `TeamID` | uint | FK, unique avec UserID |
| `UserID` | uint | FK |
| `Role` | string | `admin` ou `member` |

#### `Webhook`

| Champ | Type |
|-------|------|
| `ID` | uint |
| `ApplicationID` | uint | FK |
| `URL` | string |
| `Secret` | string | optionnel, HMAC |
| `Events` | string | ex: `"finding.critical"` |
| `IsActive` | bool |

#### `Comment`

| Champ | Type |
|-------|------|
| `ID` | uint |
| `FindingID` | uint | FK |
| `UserID` | uint | FK |
| `Body` | text |
| `CreatedAt` | time.Time |

---

## 5. RBAC & Permissions (Casbin)

### Modèle (`casbin_model.conf`)

```
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && (keyMatch2(r.obj, p.obj) || p.obj == "/*") && (r.act == p.act || p.act == "*")
```

- `keyMatch2` : `/applications/*` match `/applications/1`, `/applications/1/versions`, etc.
- `p.obj == "/*"` wildcard tout
- `p.act == "*"` wildcard toute action
- `g(r.sub, p.sub)` : sans g-rules explicites, l'identité est utilisée (r.sub == p.sub)

### Policies CSV

| Paths | Subjects |
|-------|----------|
| `/login`, `/register`, `/refresh`, `/csrf-token` | `anonymous` |
| `/logout`, `/me`, `/dashboard`, `/users/search` | `member` |
| `/groups`, `/groups/*` | `member` |
| `/applications`, `/applications/*` | `member` |
| `/teams`, `/teams/*` | `member` |
| `/scans`, `/scans/*` | `member` |
| `/findings`, `/findings/*` | `member` |
| `/scanner-types`, `/scanner-types/*` | `member` |
| `/admin/permissions`, `/admin/permissions/*` | `admin` |
| `/*` | `admin` |

### Middleware : `RequireResourceAccess(resourcePath, action)`

1. **Admin bypass** : si `user.Role == "admin"`, tout est autorisé
2. **Check direct** : `Enforce("user:{uid}", resourcePath, action)`
3. **Check teams** : pour chaque `TeamMember` de l'user, `Enforce("team:{tid}", resourcePath, action)`
4. **Group inheritance** (applications seulement) : si le path est `/applications/{id}`, récupère le `GroupID` de l'app et vérifie `user:{uid}` ou `team:{tid}` contre `/groups/{groupID}`

### Auto-grant à la création

Quand un utilisateur crée une ressource, une policy lui est automatiquement octroyée :

| Ressource | Policy ajoutée |
|-----------|----------------|
| Group | `user:{creatorID} → /groups/{id} → *` |
| Application | `user:{creatorID} → /applications/{id} → *` |
| Team | `user:{creatorID} → /teams/{id} → *` + `team:{id} → /teams/{id} → read` |

### Permissions applicatives

L'utilisateur qui a `write` ou `*` sur une application peut gérer les accès via :

| Endpoint | Action |
|----------|--------|
| `GET    /applications/:id/permissions` | Lister les policies de l'app |
| `POST   /applications/:id/permissions` | Ajouter `user:{id}` ou `team:{id}` avec `read`/`write` |
| `DELETE /applications/:id/permissions` | Supprimer une policy |

Le subject est validé (user/team doit exister). L'action est restreinte à `read`/`write` (pas de `*`) pour éviter l'escalade de privilèges.

### Permissions teams

| User | Lecture (`GET /teams/:id`) | Écriture (`POST /teams/:id/members`, etc.) |
|------|--------------------------|---------------------------------------------|
| Membre standard | ✅ via `team:{id} → read` | ❌ |
| Admin de la team | ✅ via `user:{id} → *` | ✅ via `user:{id} → *` |
| Non-membre | ❌ | ❌ |

### Permissions admin globales

Les endpoints `/admin/permissions` (GET/POST/DELETE) sont réservés aux admins. Ils acceptent `subject` (`user:{id}`, `team:{id}`) ou `userId` (rétrocompatibilité) et tout type de ressource/action.

---

## 6. Endpoints API

### Auth

| Méthode | Path | Auth | Description |
|---------|------|------|-------------|
| POST | `/api/register` | Anonymous | Crée un compte (username, email, password) |
| POST | `/api/login` | Anonymous | Retourne access + refresh tokens (cookies) |
| POST | `/api/logout` | Member | Blackliste le refresh token |
| POST | `/api/refresh` | Anonymous | Échange refresh token → nouveau access token |
| GET | `/api/csrf-token` | Anonymous | Retourne `{ csrfToken }` |

### Users

| Méthode | Path | Auth | Description |
|---------|------|------|-------------|
| GET | `/api/me` | Member | Profil de l'utilisateur connecté |
| GET | `/api/users` | Admin | Liste tous les utilisateurs |
| GET | `/api/users/search?q=` | Member | Recherche utilisateurs par username/email (min 2 chars, max 20 résultats) — retourne `id` + `username` uniquement |
| GET | `/api/users/:id` | Admin | Détail d'un utilisateur |
| PUT | `/api/users/:id` | Admin | Modifier un utilisateur (bannir, changer rôle) |
| DELETE | `/api/users/:id` | Admin | Supprimer un utilisateur |

### Groups

| Méthode | Path | Auth | Description |
|---------|------|------|-------------|
| GET | `/api/groups` | Member | Liste tous les groups |
| POST | `/api/groups` | Member | Créer un group (auto-grant `*` au créateur) |
| GET | `/api/groups/:id` | Member | Détail d'un group (resource-level) |
| PUT | `/api/groups/:id` | Member | Modifier (resource-level) |
| DELETE | `/api/groups/:id` | Member | Supprimer (resource-level) |

### Applications

| Méthode | Path | Auth | Description |
|---------|------|------|-------------|
| GET | `/api/applications` | Member | Liste |
| POST | `/api/applications` | Member | Créer (auto-grant `*` au créateur) |
| GET | `/api/applications/:id` | Member | Détail + versions + defaultVersion (resource-level) |
| GET | `/api/applications/by-slug/:slug` | Member | Idem par slug |
| PUT | `/api/applications/:id` | Member | Modifier (resource-level) |
| DELETE | `/api/applications/:id` | Member | Supprimer (resource-level) |
| POST | `/api/applications/:id/regenerate-token` | Member | Nouvel API token (resource-level) |

#### Versions

| Méthode | Path | Auth |
|---------|------|------|
| GET | `/api/applications/:id/versions` | Member |
| POST | `/api/applications/:id/versions` | Member |
| GET | `/api/applications/:id/versions/:versionId` | Member |
| PATCH | `/api/applications/:id/versions/:versionId` | Member |
| DELETE | `/api/applications/:id/versions/:versionId` | Member |
| GET | `/api/applications/:id/versions/:versionId/findings` | Member |
| GET | `/api/applications/:id/versions/compare?from=&to=` | Member |

La route `compare` est déclarée **avant** `:versionId` dans Gin pour éviter le conflit. `from` et `to` peuvent être des ID numériques ou des noms de version.

La réponse de comparaison :
```json
{
  "from": { "id": 1, "name": "v1.0.0", ... },
  "to":   { "id": 2, "name": "v1.0.1", ... },
  "fixed":   [ /* findings présents dans from, absents de to */ ],
  "new":     [ /* findings présents dans to, absents de from */ ],
  "stillPresent": [ /* findings dans les deux */ ]
}
```

Déduplication par `ruleId + filePath`.

#### Permissions applicatives

| Méthode | Path | Auth | Description |
|---------|------|------|-------------|
| GET | `/api/applications/:id/permissions` | read sur l'app | Liste les policies |
| POST | `/api/applications/:id/permissions` | write sur l'app | Ajouter `user:{id}` / `team:{id}` avec `read`/`write` |
| DELETE | `/api/applications/:id/permissions` | write sur l'app | Supprimer une policy |

#### Webhooks

| Méthode | Path | Auth |
|---------|------|------|
| GET | `/api/applications/:id/webhooks` | Member (resource-level) |
| POST | `/api/applications/:id/webhooks` | Member (resource-level) |
| DELETE | `/api/applications/:id/webhooks/:webhookId` | Member (resource-level) |

#### Ingest

| Méthode | Path | Auth |
|---------|------|------|
| POST | `/api/applications/:id/ingest` | Session JWT + CSRF |
| POST | `/api/applications/by-slug/:slug/ingest` | Session JWT + CSRF |
| POST | `/api/ingest` | API token (`X-Api-Token` header) |

Multipart form-data :
| Field | Type | Description |
|-------|------|-------------|
| `file` | file | Fichier JSON/SARIF du scanner |
| `scannerType` | string | Optionnel, auto-détecté sinon |
| `version` | string | Nom de version (upsert) |
| `branch` | string | Optionnel |

### Teams

| Méthode | Path | Auth | Description |
|---------|------|------|-------------|
| GET | `/api/teams` | Member | Liste filtrée par membership (admin voit tout) |
| POST | `/api/teams` | Member | Créer (auto-grant `*` au créateur + `team:{id} → read`) |
| GET | `/api/teams/:id` | Member | Détail + members (resource-level) |
| PUT | `/api/teams/:id` | Member | Modifier (resource-level, write) |
| DELETE | `/api/teams/:id` | Member | Supprimer (resource-level, write) |
| POST | `/api/teams/:id/members` | Member | Ajouter membre (resource-level, write) |
| DELETE | `/api/teams/:id/members/:userId` | Member | Supprimer membre (resource-level, write) |

`GET /api/teams/:id` retourne :
```json
{
  "id": 1,
  "name": "SecTeam",
  "description": "...",
  "createdAt": "...",
  "updatedAt": "...",
  "members": [
    { "userId": 1, "role": "admin", "userName": "Alice" }
  ]
}
```

Les membres sont préchargés via GORM (pas d'appel séparé). L'email n'est pas exposé.

### Scans

| Méthode | Path | Auth | Description |
|---------|------|------|-------------|
| GET | `/api/scans` | Member | Liste filtrée par accès applicatif. Query params : `applicationId`, `applicationVersionId`, `scannerTypeId` |
| GET | `/api/scans/:id` | Member | Détail |

### Findings

| Méthode | Path | Auth | Description |
|---------|------|------|-------------|
| GET | `/api/findings` | Member | Liste filtrée par accès applicatif. Query params : `applicationId`, `applicationVersionId`, `severity`, `status`, `scannerTypeId`, `scanId`, `assignedTo=me` |
| GET | `/api/findings/:id` | Member | Détail |
| PATCH | `/api/findings/:id/status` | Member | Changer le status |
| PATCH | `/api/findings/:id/assign` | Member | Assigner à un user + dueDate |
| PATCH | `/api/findings/:id/review` | Member | Review (status optionnel) |
| POST | `/api/findings/:id/comments` | Member | Ajouter un commentaire |
| GET | `/api/findings/:id/comments` | Member | Liste des commentaires |

### Scanner Types

| Méthode | Path | Auth |
|---------|------|------|
| GET | `/api/scanner-types` | Member |

### Dashboard

| Méthode | Path | Auth | Description |
|---------|------|------|-------------|
| GET | `/api/dashboard/stats` | Member | Statistiques enrichies |

Réponse :
```json
{
  "totalUsers": 2,
  "adminUsers": 1,
  "memberUsers": 1,
  "bannedUsers": 0,
  "registeredAt": "2026-06-18",
  "totalFindings": 142,
  "bySeverity": [ { "severity": "critical", "count": 3 }, ... ],
  "byScanner": [ { "scannerType": "semgrep", "count": 98 } ],
  "byStatus": [ { "status": "open", "count": 80 }, ... ],
  "recentScans": 5,
  "topFindings": [ { "ruleId": "go.lang.security...", "title": "...", "count": 8 } ],
  "myOpenFindings": 3,
  "overdueFindings": 1
}
```

### Admin Permissions

| Méthode | Path | Auth | Description |
|---------|------|------|-------------|
| GET | `/api/admin/permissions?subject=` | Admin | Liste toutes les policies, optionnellement filtrées par subject |
| POST | `/api/admin/permissions` | Admin | Ajouter une policy (`{ subject, resource, action }` ou `{ userId, resource, action }`) |
| DELETE | `/api/admin/permissions` | Admin | Supprimer une policy |

---

## 7. Scan Ingestion

### Flux

1. Le CI/CD push un fichier de résultats via `POST /api/ingest` avec `X-Api-Token`
2. Le token identifie l'application (et la version si le champ `version` est fourni)
3. Le fichier est parsé par le handler correspondant au scanner type
4. Les findings sont créés et liés à la version, au scan, et au scanner type
5. Les webhooks sont déclenchés pour les findings critical/high
6. La version est upsertée (créée si elle n'existe pas, mise à jour sinon)

### Parsers

Package `backend/parsers/` avec registry pattern :

| Parser | Status |
|--------|--------|
| `semgrep` | ✅ Fonctionnel (JSON array) |
| `trivy` | ❌ Stub |
| `gitleaks` | ❌ Stub |
| `grype` | ❌ Stub |
| `snyk` | ❌ Stub |
| `checkov` | ❌ Stub |

L'auto-détection du scanner type se fait par heuristique sur les champs JSON racine du fichier uploadé.

### Exemple CI/CD

```yaml
curl -X POST https://servasec.local/api/applications/by-slug/${CI_PROJECT_PATH_SLUG}/ingest \
  -F "file=@results.sarif" \
  -F "scannerType=semgrep" \
  -F "version=${CI_COMMIT_BRANCH}"
```

---

## 8. Cycle de Vie des Findings

### Workflow des statuts

```
open ──→ confirmed ──→ fixed ──→ closed
  │                       │
  └──→ false_positive ←───┘
```

- `open` → tout statut
- `confirmed` → `fixed` ou `false_positive`
- `fixed` → `closed` ou `false_positive`
- `false_positive` → terminal
- `closed` → terminal

Quand le status passe à `fixed`, le champ `fixedAt` est automatiquement timestampé.

### Assignation

- `PATCH /api/findings/:id/assign` : body `{ userId, dueDate? }`
- `dueDate` est une date ISO optionnelle
- Findings avec `dueDate` dépassée et status non fermé sont comptés comme "overdue"

### Review

- `PATCH /api/findings/:id/review` : body `{ status? }`
- Le `reviewedBy` est automatiquement set à l'utilisateur connecté
- Si `status` est fourni, il est appliqué au finding

### Commentaires

- `POST /api/findings/:id/comments` : body `{ body }`
- `GET /api/findings/:id/comments` : liste chronologique
- Réponse : `{ id, findingId, userId, user: { id, username }, body, createdAt }`

---

## 9. Webhooks

### Configuration

Un webhook est lié à une application :
```json
{
  "id": 1,
  "applicationId": 1,
  "url": "https://hooks.slack.com/...",
  "secret": "optional-hmac-secret",
  "events": "finding.critical",
  "isActive": true
}
```

Les events sont stockés comme une string (ex: `"finding.critical"`).

### Déclenchement

Après un ingest, si des findings `critical` ou `high` sont créés, un POST JSON est envoyé en arrière-plan vers chaque webhook actif :

```json
{
  "event": "finding.critical",
  "applicationId": 1,
  "scanId": 5,
  "findings": [ /* seuls les critical/high */ ],
  "timestamp": "2026-06-21T19:50:35Z"
}
```

Header `X-Servasec-Signature` : HMAC-SHA256 hex du body (si `secret` configuré).

---

## 10. Dev Setup

### Docker Compose (`docker-compose.dev.yml`)

| Service | Image | Port | Notes |
|---------|-------|------|-------|
| `backend` | `golang:1.24-alpine` | 8080 | Air hot-reload, montage `./backend` |
| `frontend` | `node:20-alpine` | 3000 | Next.js dev, montage `./frontend` |
| `postgres` | `postgres:16-alpine` | 5432 | Volume `pgdata` |
| `caddy` | `caddy:2-alpine` | 443 (443) | Auto TLS, reverse proxy |

### Makefile

| Commande | Description |
|----------|-------------|
| `make dev` | Lance tous les services |
| `make build` | Build les binaires |
| `make down` | Arrête les services |
| `make logs` | Suit les logs |

### Air (hot-reload backend)

`backend/air.toml` surveille `backend/` et recompile au changement. Pas de build manuel nécessaire en dev.

### Caddy

- TLS automatique via `caddy/servasec.local.crt` + `caddy/servasec.local.key` (self-signed)
- `/api/*` → backend (strip `/api`)
- `/*` → frontend Next.js
- Les cookies sont set sur le même domaine, pas de CORS

### Volumes

`USER_UID`/`USER_GID` dans le `.env` pour que les fichiers créés dans les conteneurs aient le bon owner.

---

## 11. CI/CD

### GitHub Release Workflow (`.github/workflows/release.yml`)

Déclenché sur `push` vers `main`.

Étapes :
1. **Auto-bump SemVer** via `anothrNick/github-tag-action` — détecte le prochain tag basé sur les commits depuis le dernier tag
2. **Release notes** générées par `ncipollo/release-action` avec un groupeur custom :
   - `feat` → 🚀 Features
   - `fix` → 🐛 Bug Fixes
   - `docs` → 📚 Documentation
   - `chore` → 🔧 Maintenance
   - `refactor` → ♻️ Refactoring
3. Pas de build Docker dans le CI (optionnel, à ajouter)

---

## 12. Audit de Sécurité

Réalisé le 2026-06-18 avec Opengrep (15 règles custom, scope 22 fichiers Go).

### Résultats

| ID | Titre | Sévérité | Correctif |
|----|-------|----------|-----------|
| CWE-330 | JWT secret fallback à `"secret"` | Critical | Validation au démarrage, fatal si manquant |
| CWE-798 | Mot de passe admin en dur | Critical | Fatal si `SSC_ADMIN_PASSWORD` < 8 chars |
| CWE-613 | Blacklist en mémoire seulement | Critical | Migration vers PostgreSQL + cleanup goroutine |
| P1 | Casbin `LoadPolicy()` par requête | High | `LoadPolicy()` unique à l'init |
| P1 | Erreurs DB non checkées | High | `.Error` checké partout |
| P2 | Détection doublons par string match | Medium | Pre-checks avant création |
| P2 | `binding:"required"` manquant | Medium | Ajouté aux DTO |
| P2 | Trusted proxies hardcodés | Medium | Variable d'env `TRUSTED_PROXIES_CIDR` |
| P2 | `gin.Default()` en prod | Medium | `gin.New()` + logger conditionnel |
| P3 | Dashboard sans error checks | Low | `.Count()` errors checkés |
| P3 | Casbin stale après seed CSV | Low | `LoadPolicy()` après seed |

Community rulesets : 0 findings (pas de SQLi, command injection, open redirect, JWT none, TLS faible, NoSQL injection).

---

## Pages Frontend

| Page | Fichier | Description |
|------|---------|-------------|
| Login | `pages/login.tsx` | Authentification |
| Dashboard | `pages/index.tsx` | KPIs, sévérité, top findings |
| Applications | `pages/applications/index.tsx` | CRUD liste |
| Application detail | `pages/applications/[id].tsx` | Fiche, versions, ingest, webhooks, permissions |
| Comparaison | `pages/applications/[id]/compare.tsx` | Comparer deux versions |
| Groups | `pages/groups.tsx` | CRUD groups |
| Teams | `pages/teams/index.tsx` | CRUD liste |
| Team detail | `pages/teams/[id].tsx` | Détail + membres |
| Findings | `pages/findings.tsx` | Liste filtrée |
| Finding detail | `pages/findings/[id].tsx` | Détail + assign/review/comments |
| Scans | `pages/scans.tsx` | Liste filtrée |
| Admin Users | `pages/admin/users.tsx` | Gestion utilisateurs |
| Admin Permissions | `pages/admin/permissions.tsx` | Gestion policies Casbin |

---

## Références

- `docs/tasks/01-architecture-test-plan.md` — Architecture détaillée + plan de test curl
- `docs/tasks/02-security-audit.md` — Rapport d'audit détaillé (Opengrep)
- `docs/tasks/03-backend-versions-ingest-parsers.md` — Versions, Scanner Types, Ingest, Parsers
- `docs/tasks/04-frontend-adaptations.md` — Adaptations frontend versions/scanner types
- `docs/tasks/05-dashboard-comparison-lifecycle-webhooks.md` — Dashboard, Comparaison, Cycle de vie, Permissions, Webhooks
- `docs/tasks/06-scope-fixes-teams-permissions.md` — Scope teams/findings/scans, Permissions teams
- `docs/tasks/07-teams-read-fix.md` — Correctif lecture teams, user search
- `docs/tasks/08-app-permissions.md` — Permissions applicatives, correctif lecture teams
