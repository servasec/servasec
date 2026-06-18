# Security Analysis Report — servasec

**Date:** 2026-06-18  
**Scope:** Backend Go codebase (`backend/`)

---

## 1. Methodology

Custom Opengrep rules targeting common web backend vulnerabilities in Go + Gin + GORM + Casbin projects. Scans were run manually and findings triaged.

### Rules applied

| Rule file | Target |
|-----------|--------|
| `rules/audit/audit-hardcoded-credentials` | Hardcoded passwords, API keys, secrets |
| `rules/audit/audit-crypto-weak-md5` | Weak hashing (MD5, SHA1 for passwords) |
| `rules/audit/audit-weak-cookie` | Cookies without Secure/HttpOnly/SameSite |
| `rules/csharp/lang/security/hardcoded-secret` | Generic secret-in-code detection |
| `rules/command-line-injection` | Unsafe `os/exec` with user input |
| `rules/go/lang/security/audit/open-redirect` | User-controlled redirect targets |
| `rules/go/lang/security/hardcoded-credentials` | Hardcoded JWT secrets, DB passwords |
| `rules/go/lang/security/insufficient-session-expiration` | Missing token expiry |
| `rules/go/lang/security/jwt-none-algorithm` | JWT alg `none` acceptance |
| `rules/go/lang/security/no-sql-injection` | NoSQL injection patterns (N/A) |
| `rules/go/lang/security/sql-injection` | Unsafe SQL string building |
| `rules/go/lang/security/ssl-verify` | Disabled TLS verification |
| `rules/go/lang/security/use-of-weak-crypto` | Weak PRNG usage (math/rand for secrets) |
| `rules/request-forgery` | CSRF weaknesses |
| `rules/spring/security/audit/cors-all-origin` | Permissive CORS `*` |

### Scope: 22 Go files

`controllers/auth.go`, `controllers/user.go`, `config/db.go`, `config/casbin.go`, `config/seeder.go`, `dto/auth.go`, `main.go`, `middleware/auth.go`, `middleware/csrf.go`, `middleware/rate_limiter.go`, `utils/jwt.go`, `utils/token.go`, `utils/blacklist.go`, `utils/secrets.go`, `utils/response.go`, `utils/validation.go`, `debug/log.go`, `docker-compose.dev.yml`, `backend/.env.example`, `routes/auth.go`, `routes/user.go`, `routes/dashboard.go`

---

## 2. Findings Summary

| ID | Title | Severity | Status |
|----|-------|----------|--------|
| CWE-330 | JWT secret fallback to `"secret"` | **Critical** | ✅ Fixed |
| CWE-798 | Hardcoded admin password in plaintext | **Critical** | ✅ Fixed |
| CWE-613 | Token blacklist in memory only | **Critical** | ✅ Fixed |
| P1 | Casbin `LoadPolicy()` called per-request | High | ✅ Fixed |
| P1 | Missing error checks on DB.Save/Delete | High | ✅ Fixed |
| P2 | Error-string matching instead of pre-checks | Medium | ✅ Fixed |
| P2 | `LoginInput` missing `binding:"required"` | Medium | ✅ Fixed |
| P2 | Trusted proxies hardcoded to single IP | Medium | ✅ Fixed |
| P2 | `gin.Default()` used (logs+panic recovery in prod) | Medium | ✅ Fixed |
| P3 | Dashboard queries missing error checks | Low | ✅ Fixed |
| P3 | Casbin seeding (CEF in-memory stale after CSV seed) | Low | ✅ Fixed |

**Community rulesets:** 0 findings matched (no SQLi, no command injection, no open redirect, no JWT `none` algorithm, no weak TLS, no NoSQL injection).

---

## 3. Details

### CWE-330 — JWT secret fallback to `"secret"` (Critical) ✅ Fixed

**File:** `utils/secrets.go:27`  
**Before:** `JWTSecret` fell back to `[]byte("secret")` when env var was unset.  
**After:** `ValidateSecrets()` in `main.go` fatals early if `JWT_SECRET`, `REFRESH_SECRET`, or `CSRF_SECRET` are missing or equal to `"secret"`.

### CWE-798 — Hardcoded admin password (Critical) ✅ Fixed

**File:** `config/seeder.go`  
**Before:** Admin password was `"changeme"` if `SSC_ADMIN_PASSWORD` unset.  
**After:** `utils/secrets.go:67` — if `SSC_ADMIN_PASSWORD` is unset or `< 8 chars`, the server fatals on startup.

### CWE-613 — In-memory blacklist (Critical) ✅ Fixed

**File:** `utils/blacklist.go:12`  
**Before:** `var blacklistedTokens = make(map[string]bool)` — lost on restart.  
**After:** Blacklist moved to `blacklisted_tokens` PostgreSQL table; 5-min cleanup goroutine deletes expired entries.

### P1 — Per-request `LoadPolicy()` (High) ✅ Fixed

**File:** `middleware/auth.go` (previously called `CEF.LoadPolicy()` inside `CheckPolicy`).  
**After:** `LoadPolicy()` called once in `config.InitCasbin()`. Policy changes propagate via `SavePolicy()`.

### P1 — Missing error checks on DB.Save/Delete (High) ✅ Fixed

**Files:** `controllers/user.go:84`, `controllers/auth.go:238,280`, `config/seeder.go:53`  
**Before:** `config.DB.Save(&user)` without `.Error` check.  
**After:** All calls now check and return `InternalServerError` on failure.

### P2 — Error-string matching (Medium) ✅ Fixed

**Before:** `strings.Contains(err.Error(), "users_username_key")` to detect duplicates.  
**After:** Pre-checks with `DB.Where(...).First(&existing)` before creating.

### P2 — Missing `binding:"required"` (Medium) ✅ Fixed

**File:** `dto/auth.go:4` — added `binding:"required"` to `Username` and `Password` fields.

### P2 — Hardcoded trusted proxies (Medium) ✅ Fixed

**File:** `main.go` — `router.SetTrustedProxies([]string{"172.70.1.0/24"})` replaced with `getTrustedProxies()` reading `TRUSTED_PROXIES_CIDR` env var (defaults to `172.70.1.0/24`).

### P2 — `gin.Default()` (Medium) ✅ Fixed

**Before:** `router := gin.Default()` — includes Logger and Recovery middleware (logs in prod).  
**After:** `gin.New()` + explicit `router.Use(gin.Logger())` and `router.Use(gin.Recovery())` only when `SSC_DEBUG_ENABLED=true`.

### P3 — Dashboard missing error checks (Low) ✅ Fixed

**Files:** `controllers/dashboard.go` — added error checks to `DB.Model(...).Count(...)` calls.

### P3 — Casbin seeding stale in-memory (Low) ✅ Fixed

**File:** `config/casbin.go:InitCasbin()` — added `CEF.LoadPolicy()` call after `SeedCasbinFromCsv()` to reload in-memory policies from DB.

---

## 4. Opengrep rules

Custom rules are in `backend/rules/`. Run:

```bash
docker run --rm -v "$PWD:/src" opengrep/opengrep \
  opengrep scan --config /src/backend/rules --severity WARNING --error /src/backend
```

Note: The `.Error` suffix convention used across the codebase (`utils.NotFoundError(c, "message")`) causes false positives in rules looking for raw `err != nil` patterns. All `P1`–`P3` findings have been manually verified and fixed where real.

---

## 5. Files modified

| File | Change |
|------|--------|
| `config/db.go` | AutoMigrate BlacklistedToken, error check |
| `config/seeder.go` | bcrypt admin password, error checks |
| `config/casbin.go` | LoadPolicy after seed |
| `controllers/auth.go` | Error checks on Save, pre-checks |
| `controllers/user.go` | Error checks on Save/Delete |
| `controllers/dashboard.go` | Error checks on Count queries |
| `dto/auth.go` | binding:"required" |
| `middleware/auth.go` | Remove LoadPolicy from CheckPolicy |
| `middleware/csrf.go` | Secure/HttpOnly/SameSite cookies |
| `utils/blacklist.go` | PostgreSQL-backed blacklist |
| `utils/secrets.go` | Validation, admin password check |
| `utils/jwt.go` | Remove fallback secrets |
| `main.go` | ValidateSecrets(), cleanup cleanup |
