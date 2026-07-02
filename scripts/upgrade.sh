#!/bin/sh
# Servasec Upgrade Script
# Usage:
#   ./scripts/upgrade.sh              # build latest
#   ./scripts/upgrade.sh v1.1.0       # build and tag v1.1.0
#
# Automatically handles PostgreSQL major version upgrades
# (dump old PG, remove volume, start new PG, restore).
#
# Requires: docker compose v2, running servasec stack

set -e

SSC_VERSION="${1:-latest}"
SSC_COMPOSE="docker compose -f docker-compose.prod.yml"
SSC_ENV_FILE=".env"

TS=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="servasec_backup_${TS}.sql"

# Track PG version to detect major upgrades
PG_VERSION_FILE="scripts/.pg_version"
PG_TARGET="${POSTGRES_VERSION:-17}"
PG_CURRENT=""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

info()  { printf "${BLUE}ℹ${NC}  %s\n" "$1"; }
ok()    { printf "${GREEN}✓${NC}  %s\n" "$1"; }
warn()  { printf "${YELLOW}⚠${NC}  %s\n" "$1"; }
fail()  { printf "${RED}✗${NC}  %s\n" "$1"; exit 1; }

if [ -f "$PG_VERSION_FILE" ]; then
	PG_CURRENT=$(cat "$PG_VERSION_FILE")
fi

# ──────────────────────────────────────────────
#  Check prerequisites
# ──────────────────────────────────────────────

if [ ! -f "$SSC_ENV_FILE" ]; then
	warn "No .env file found — using default env vars"
fi

info "Upgrading Servasec to ${SSC_VERSION}"
echo ""

# ──────────────────────────────────────────────
#  1. Backup
# ──────────────────────────────────────────────

if [ -f "$BACKUP_FILE" ]; then
	warn "Backup already exists: ${BACKUP_FILE} — skipping"
else
	info "[1/5] Backing up database → ${BACKUP_FILE}"
	if $SSC_COMPOSE exec -T db pg_dump -U "${POSTGRES_USER:-servasec}" "${POSTGRES_DB:-servasec}" > "$BACKUP_FILE" 2>/dev/null; then
		ok "Backup saved (${BACKUP_FILE})"
	else
		warn "Backup skipped (is the stack running?)"
		rm -f "$BACKUP_FILE"
	fi
fi

# ──────────────────────────────────────────────
#  2. Detect PostgreSQL major version change
# ──────────────────────────────────────────────

echo ""

if [ "$PG_CURRENT" != "" ] && [ "$PG_CURRENT" != "$PG_TARGET" ]; then
	warn "PostgreSQL version changed: ${PG_CURRENT} → ${PG_TARGET}"
	warn "A full dump/restore is required — the old volume will be destroyed."
	echo ""
	printf "  Proceed? [y/N] "
	read -r CONFIRM
	if [ "$CONFIRM" != "y" ] && [ "$CONFIRM" != "Y" ]; then
		fail "Aborted by user"
	fi

	info "[2/5] Removing old PostgreSQL ${PG_CURRENT} volume..."
	$SSC_COMPOSE down
	if docker volume rm servasec_postgres_data 2>/dev/null; then
		ok "Old volume removed"
	else
		warn "Volume not found or already removed — continuing"
	fi

	info "Starting new PostgreSQL ${PG_TARGET}..."
	POSTGRES_VERSION="$PG_TARGET" $SSC_COMPOSE up -d db
	sleep 5
	ok "PostgreSQL ${PG_TARGET} ready"

	info "Restoring data..."
	POSTGRES_VERSION="$PG_TARGET" $SSC_COMPOSE exec -T db psql -U "${POSTGRES_USER:-servasec}" "${POSTGRES_DB:-servasec}" < "$BACKUP_FILE"
	ok "Data restored"

	info "Stopping DB for full service restart..."
	$SSC_COMPOSE down

	echo "$PG_TARGET" > "$PG_VERSION_FILE"
	ok "PG version recorded: ${PG_TARGET}"
else
	info "PostgreSQL version unchanged (${PG_TARGET}) — skipping"

	if [ ! -f "$PG_VERSION_FILE" ]; then
		echo "$PG_TARGET" > "$PG_VERSION_FILE"
	fi
fi

# ──────────────────────────────────────────────
#  3. Pull base images
# ──────────────────────────────────────────────

echo ""
info "[3/5] Pulling base images..."
$SSC_COMPOSE pull --quiet 2>/dev/null || true
ok "Images pulled"

# ──────────────────────────────────────────────
#  4. Build & restart
# ──────────────────────────────────────────────

echo ""
info "[4/5] Building servasec:${SSC_VERSION} and restarting..."
SSC_VERSION="$SSC_VERSION" $SSC_COMPOSE up --build -d
ok "Services restarted"

# ──────────────────────────────────────────────
#  5. Verify migrations
# ──────────────────────────────────────────────

echo ""
info "[5/5] Waiting for migrations..."
sleep 4

if $SSC_COMPOSE logs backend --tail=30 2>&1 | grep -qE "migration|applied|version|Server starting"; then
	$SSC_COMPOSE logs backend --tail=10 2>/dev/null | while IFS= read -r line; do
		case "$line" in
			*"migration"*|*"applied"*) ok "$(echo "$line" | sed 's/.*\s//')" ;;
		esac
	done
	ok "Upgrade complete (${SSC_VERSION})"
else
	warn "Could not verify migration status — check manually:"
	echo "  docker compose logs backend --tail=30 | grep -i migration"
fi

echo ""
echo "────────────────────────────────────────────────"
echo "  ${GREEN}Servasec ${SSC_VERSION} is running${NC}"
echo ""
echo "  Check full logs:"
echo "    ${SSC_COMPOSE} logs backend --tail=50"
echo ""
echo "  Rollback if needed:"
echo "    cat ${BACKUP_FILE} | ${SSC_COMPOSE} exec -T db psql -U servasec servasec"
echo "────────────────────────────────────────────────"
