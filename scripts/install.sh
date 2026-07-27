#!/bin/sh
# Servasec Install Script
# Usage:
#   curl -fsSL https://servasec.com/install.sh | sh          # one-liner (clones repo)
#   curl -fsSL https://servasec.com/install.sh | sh -s -- -i  # interactive
#   ./scripts/install.sh                                      # from repo root
#   ./scripts/install.sh -i                                   # interactive
#   ./scripts/install.sh --no-check                           # skip git check
#
# Builds the stack using docker-compose.prod.yml.

set -e

# ──────────────────────────────────────────────
#  Resolve script directory
# ──────────────────────────────────────────────

REPO_URL="https://github.com/servasec/servasec.git"
CLONE_DIR="$HOME/.servasec-install"

resolve_script_dir() {
    case "$0" in
        */*/*|*/*)
            if [ -f "$0" ]; then
                SCRIPT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
                if [ -f "$SCRIPT_DIR/.env.example" ]; then
                    echo "$SCRIPT_DIR"
                    return
                fi
            fi
            ;;
    esac
    echo ""
}

REPO_ROOT="$(resolve_script_dir)"

# ──────────────────────────────────────────────
#  If not in a repo, clone it
# ──────────────────────────────────────────────

if [ -z "$REPO_ROOT" ]; then
    printf "\n"
    printf "\033[0;34mi\033[0m  Not inside Servasec repo — cloning latest release...\n"

    command -v git >/dev/null 2>&1 || {
        printf "\033[0;31m✗\033[0m  git is required for remote install. Install git first.\n"
        exit 1
    }

    if [ -d "$CLONE_DIR/.git" ]; then
        printf "\033[0;34mi\033[0m  Updating existing clone at %s...\n" "$CLONE_DIR"
        git -C "$CLONE_DIR" fetch --tags --quiet 2>/dev/null || true
    else
        rm -rf "$CLONE_DIR"
        git clone --quiet --no-checkout "$REPO_URL" "$CLONE_DIR"
    fi

    # Always checkout latest release tag, never default branch
    LATEST_TAG=$(git -C "$CLONE_DIR" tag --list --sort=-version:refname | head -1)
    if [ -n "$LATEST_TAG" ]; then
        git -C "$CLONE_DIR" checkout --quiet "$LATEST_TAG"
        printf "\033[0;32m✓\033[0m  On release %s\n" "$LATEST_TAG"
    else
        fail "No release tags found. Cannot install from remote."
    fi

    REPO_ROOT="$CLONE_DIR"
fi

cd "$REPO_ROOT" || { printf "\033[0;31m✗\033[0m  Cannot cd to %s\n" "$REPO_ROOT"; exit 1; }

# ──────────────────────────────────────────────
#  Flags
# ──────────────────────────────────────────────

INTERACTIVE=false
CHECK_GIT=true
PRO_ENABLED=false
PRO_REPO_DIR=""

while [ $# -gt 0 ]; do
    case "$1" in
        -i|--interactive) INTERACTIVE=true ;;
        --no-check)       CHECK_GIT=false ;;
        --pro)            PRO_ENABLED=true ;;
        --pro-repo)
            PRO_ENABLED=true
            PRO_REPO_DIR="$2"
            shift
            ;;
        -h|--help)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  -i, --interactive        Ask questions before install"
            echo "  --no-check               Skip git branch/tag verification"
            echo "  --pro                    Build with pro features (requires servasec-pro repo)"
            echo "  --pro-repo <path>        Path to servasec-pro repo (default: ../servasec-pro)"
            echo "  -h, --help               Show this help"
            exit 0
            ;;
        *)
            printf "\033[0;31m✗\033[0m  Unknown option: %s (use -h for help)\n" "$1"
            exit 1
            ;;
    esac
    shift
done

# ──────────────────────────────────────────────
#  Config
# ──────────────────────────────────────────────

COMPOSE_FILE="docker-compose.prod.yml"
SSC_COMPOSE="docker compose -f $COMPOSE_FILE"
SSC_ENV_FILE=".env"
SSC_EXAMPLE=".env.example"

DEFAULT_DOMAIN="servasec.local"
DEFAULT_SSO_ENABLED=false
DEFAULT_PRO_REPO="../servasec-pro"

# ──────────────────────────────────────────────
#  Colors & helpers
# ──────────────────────────────────────────────

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
BOLD='\033[1m'
DIM='\033[2m'
NC='\033[0m'

info()  { printf "${BLUE}i${NC}  %s\n" "$1"; }
ok()    { printf "${GREEN}✓${NC}  %s\n" "$1"; }
warn()  { printf "${YELLOW}⚠${NC}  %s\n" "$1"; }
fail()  { printf "${RED}✗${NC}  %s\n" "$1"; exit 1; }
header() { printf "\n${BOLD}${CYAN}── %s ──${NC}\n" "$1"; }
prompt() { printf "  ${BOLD}%s${NC} " "$1"; }
dim()   { printf "${DIM}%s${NC}" "$1"; }

# ──────────────────────────────────────────────
#  1. Check prerequisites
# ──────────────────────────────────────────────

printf "\n"
info "Checking prerequisites..."

command -v docker >/dev/null 2>&1 || fail "Docker is not installed. Install it first: https://docs.docker.com/get-docker/"
docker info >/dev/null 2>&1 || fail "Docker daemon is not running. Start Docker and try again."
docker compose version --short 2>/dev/null  || fail "Docker Compose v2 is required. Install it: https://docs.docker.com/compose/install/"

HAS_OPENSSL=true
command -v openssl >/dev/null 2>&1 || HAS_OPENSSL=false

ok "Docker found"
ok "Docker Compose v2 found"
if [ "$HAS_OPENSSL" = true ]; then
    ok "OpenSSL found"
else
    warn "OpenSSL not found — using /dev/urandom as fallback"
fi

# ──────────────────────────────────────────────
#  2. Check git state
# ──────────────────────────────────────────────

if [ "$CHECK_GIT" = true ] && command -v git >/dev/null 2>&1 && git rev-parse --git-dir >/dev/null 2>&1; then
    printf "\n"
    info "Checking git state..."

    CURRENT_REF=$(git symbolic-ref --short HEAD 2>/dev/null || echo "")
    IS_TAG=false
    if git describe --tags --exact HEAD >/dev/null 2>&1; then
        CURRENT_REF=$(git describe --tags --exact HEAD 2>/dev/null)
        IS_TAG=true
    fi

    LATEST_TAG=$(git tag --list --sort=-version:refname | head -1)

    if [ "$IS_TAG" = true ]; then
        ok "On release tag: $CURRENT_REF"
    elif [ "$CURRENT_REF" = "main" ]; then
        if [ -n "$LATEST_TAG" ]; then
            warn "On 'main' branch — switching to latest release: $LATEST_TAG"
            git fetch --tags --quiet 2>/dev/null || true
            git checkout "$LATEST_TAG" --quiet
            ok "Now on $LATEST_TAG"
        else
            fail "On 'main' branch and no release tags found. Checkout a release tag first."
        fi
    elif [ "$CURRENT_REF" = "develop" ] || [ "$CURRENT_REF" != "" ]; then
        if [ -n "$LATEST_TAG" ]; then
            warn "Not on a release tag (current: $CURRENT_REF, latest tag: $LATEST_TAG)"
            if [ "$INTERACTIVE" = true ]; then
                printf "  Switch to latest tag ($LATEST_TAG)? [Y/n] "
                read -r CONFIRM
                case "$CONFIRM" in
                    ""|[yY]|yes|YES)
                        git fetch --tags --quiet 2>/dev/null || true
                        git checkout "$LATEST_TAG" --quiet
                        ok "Now on $LATEST_TAG"
                        ;;
                    *)
                        warn "Continuing on $CURRENT_REF — not recommended for production"
                        ;;
                esac
            else
                warn "Continuing on $CURRENT_REF (use -i to switch to a release tag)"
            fi
        fi
    fi
fi

# ──────────────────────────────────────────────
#  3. Check .env.example exists
# ──────────────────────────────────────────────

[ -f "$SSC_EXAMPLE" ] || fail "File $SSC_EXAMPLE not found. Are you in the Servasec root directory?"

# ──────────────────────────────────────────────
#  4. Generate secrets
# ──────────────────────────────────────────────

printf "\n"
info "Generating secrets..."

generate_secret() {
    if [ "$HAS_OPENSSL" = true ]; then
        openssl rand -base64 32 | tr -d '\n'
    else
        head -c 32 /dev/urandom | base64 | tr -d '\n'
    fi
}

JWT_SECRET=$(generate_secret)
REFRESH_SECRET=$(generate_secret)
CSRF_SECRET=$(generate_secret)
ENCRYPTION_KEY=$(generate_secret)

ok "Secrets generated"

# ──────────────────────────────────────────────
#  5. Interactive prompts (-i mode)
# ──────────────────────────────────────────────

DOMAIN="$DEFAULT_DOMAIN"
ADMIN_PASSWORD=""
SSO_ENABLED=false
SSO_PROVIDER=""
REGISTRATION_ENABLED=true

if [ "$INTERACTIVE" = true ]; then
    printf "\n"
    header "Configuration"
    echo ""

    prompt "Admin password"
    dim "(leave empty for random):"
    printf " "
    read -r INPUT_PASSWORD
    if [ -n "$INPUT_PASSWORD" ]; then
        ADMIN_PASSWORD="$INPUT_PASSWORD"
    else
        ADMIN_PASSWORD=$(generate_secret | head -c 16)
    fi

    printf "\n"
    prompt "Domain"
    dim "[$DOMAIN]:"
    printf " "
    read -r INPUT_DOMAIN
    if [ -n "$INPUT_DOMAIN" ]; then
        DOMAIN="$INPUT_DOMAIN"
    fi

    printf "\n"
    prompt "Allow public registration? [Y/n]: "
    read -r INPUT_REG
    case "$INPUT_REG" in
        [nN]|no|NO) REGISTRATION_ENABLED=false ;;
    esac

    printf "\n"
    header "SSO Configuration"
    echo ""
    printf "  ${DIM}1)${NC} None (default)\n"
    printf "  ${DIM}2)${NC} GitHub\n"
    printf "  ${DIM}3)${NC} GitLab\n"
    printf "  ${DIM}4)${NC} OIDC (Keycloak, Auth0, etc.)\n"
    printf "  ${DIM}5)${NC} Multiple providers\n"
    echo ""
    prompt "SSO provider"
    dim "[1]:"
    printf " "
    read -r INPUT_SSO

    case "$INPUT_SSO" in
        2) SSO_ENABLED=true; SSO_PROVIDER="github" ;;
        3) SSO_ENABLED=true; SSO_PROVIDER="gitlab" ;;
        4) SSO_ENABLED=true; SSO_PROVIDER="oidc" ;;
        5) SSO_ENABLED=true; SSO_PROVIDER="multiple" ;;
        "") SSO_ENABLED=false ;;
        *)
            SSO_ENABLED=true
            case "$INPUT_SSO" in
                [gG][iI][tT][hH][uU][bB]) SSO_PROVIDER="github" ;;
                [gG][iI][tT][lL][aA][bB]) SSO_PROVIDER="gitlab" ;;
                [oO][iI][dD][cC]) SSO_PROVIDER="oidc" ;;
                *) SSO_ENABLED=false ;;
            esac
            ;;
    esac

    if [ "$SSO_ENABLED" = true ]; then
        printf "\n"
        case "$SSO_PROVIDER" in
            github)
                prompt "GitHub Client ID: "
                read -r SSO_GITHUB_CLIENT_ID
                prompt "GitHub Client Secret: "
                read -r SSO_GITHUB_CLIENT_SECRET
                ;;
            gitlab)
                prompt "GitLab Client ID: "
                read -r SSO_GITLAB_CLIENT_ID
                prompt "GitLab Client Secret: "
                read -r SSO_GITLAB_CLIENT_SECRET
                prompt "GitLab URL"
                dim " [https://gitlab.com]:"
                printf " "
                read -r SSO_GITLAB_BASE_URL
                if [ -z "$SSO_GITLAB_BASE_URL" ]; then
                    SSO_GITLAB_BASE_URL="https://gitlab.com"
                fi
                ;;
            oidc)
                prompt "OIDC Client ID: "
                read -r SSO_OIDC_CLIENT_ID
                prompt "OIDC Client Secret: "
                read -r SSO_OIDC_CLIENT_SECRET
                prompt "OIDC Issuer URL: "
                read -r SSO_OIDC_ISSUER_URL
                prompt "OIDC Scopes"
                dim " [openid profile email]:"
                printf " "
                read -r SSO_OIDC_SCOPES
                if [ -z "$SSO_OIDC_SCOPES" ]; then
                    SSO_OIDC_SCOPES="openid profile email"
                fi
                ;;
            multiple)
                printf "\n"
                info "Configure each provider below. Press Enter to skip."
                echo ""

                prompt "GitHub Client ID: "
                read -r SSO_GITHUB_CLIENT_ID
                prompt "GitHub Client Secret: "
                read -r SSO_GITHUB_CLIENT_SECRET

                printf "\n"
                prompt "GitLab Client ID: "
                read -r SSO_GITLAB_CLIENT_ID
                prompt "GitLab Client Secret: "
                read -r SSO_GITLAB_CLIENT_SECRET
                prompt "GitLab URL"
                dim " [https://gitlab.com]:"
                printf " "
                read -r SSO_GITLAB_BASE_URL
                if [ -z "$SSO_GITLAB_BASE_URL" ]; then
                    SSO_GITLAB_BASE_URL="https://gitlab.com"
                fi

                printf "\n"
                prompt "OIDC Client ID: "
                read -r SSO_OIDC_CLIENT_ID
                prompt "OIDC Client Secret: "
                read -r SSO_OIDC_CLIENT_SECRET
                prompt "OIDC Issuer URL: "
                read -r SSO_OIDC_ISSUER_URL
                prompt "OIDC Scopes"
                dim " [openid profile email]:"
                printf " "
                read -r SSO_OIDC_SCOPES
                if [ -z "$SSO_OIDC_SCOPES" ]; then
                    SSO_OIDC_SCOPES="openid profile email"
                fi
                ;;
        esac
    fi

    # ── Summary ──
    printf "\n"
    header "Install Summary"
    echo ""
    printf "  Domain:        ${BOLD}%s${NC}\n" "$DOMAIN"
    printf "  Admin pass:    ${BOLD}%s${NC}\n" "$ADMIN_PASSWORD"
    printf "  Registration:  ${BOLD}%s${NC}\n" "$([ "$REGISTRATION_ENABLED" = true ] && echo 'enabled' || echo 'disabled')"
    if [ "$SSO_ENABLED" = true ]; then
        printf "  SSO:           ${BOLD}%s${NC}\n" "$SSO_PROVIDER"
    else
        printf "  SSO:           ${BOLD}disabled${NC}\n"
    fi
    printf "  Pro features:  ${BOLD}%s${NC}\n" "$([ "$PRO_ENABLED" = true ] && echo 'yes' || echo 'no')"
    echo ""
    prompt "Proceed with install? [Y/n]: "
    read -r CONFIRM
    case "$CONFIRM" in
        [nN]|no|NO)
            printf "\n"
            info "Install cancelled."
            exit 0
            ;;
    esac
fi

# ──────────────────────────────────────────────
#  6. Copy pro files (if --pro)
# ──────────────────────────────────────────────

if [ "$PRO_ENABLED" = true ]; then
    printf "\n"
    info "Setting up pro features..."

    if [ -z "$PRO_REPO_DIR" ]; then
        PRO_REPO_DIR="$DEFAULT_PRO_REPO"
    fi

    if [ ! -d "$PRO_REPO_DIR" ]; then
        fail "Pro repo not found at: $PRO_REPO_DIR"
    fi

    if [ ! -d "$PRO_REPO_DIR/backend/pro" ]; then
        fail "Pro backend not found at: $PRO_REPO_DIR/backend/pro/"
    fi

    mkdir -p backend/pro
    cp "$PRO_REPO_DIR"/backend/pro/*.go backend/pro/
    ok "Pro files copied from $PRO_REPO_DIR"

    if [ -f "$SSC_ENV_FILE" ]; then
        if grep -q "^BUILD_TAGS=" "$SSC_ENV_FILE" 2>/dev/null; then
            sed -i.bak 's|^BUILD_TAGS=.*|BUILD_TAGS=pro|' "$SSC_ENV_FILE" && rm -f "$SSC_ENV_FILE.bak"
        else
            echo "BUILD_TAGS=pro" >> "$SSC_ENV_FILE"
        fi
    fi

    ok "BUILD_TAGS=pro configured"
fi

# ──────────────────────────────────────────────
#  7. Build .env
# ──────────────────────────────────────────────

printf "\n"
info "Creating .env..."

# Fill defaults for non-interactive mode
if [ "$INTERACTIVE" = false ]; then
    ADMIN_PASSWORD=$(generate_secret | head -c 16)
fi

if [ -f "$SSC_ENV_FILE" ]; then
    if [ "$INTERACTIVE" = true ]; then
        printf "  .env already exists. Overwrite? [y/N] "
        read -r CONFIRM
        case "$CONFIRM" in
            ""|[nN]|no|NO)
                warn "Keeping existing .env"
                ;;
            *)
                cp "$SSC_EXAMPLE" "$SSC_ENV_FILE"
                ok ".env overwritten from $SSC_EXAMPLE"
                ;;
        esac
    else
        warn ".env already exists — keeping it (use -i to overwrite)"
    fi
else
    cp "$SSC_EXAMPLE" "$SSC_ENV_FILE"
    ok ".env created from $SSC_EXAMPLE"
fi

if [ -f "$SSC_ENV_FILE" ]; then
    update_env() {
        if grep -q "^$1=" "$SSC_ENV_FILE" 2>/dev/null; then
            sed -i.bak "s|^$1=.*|$1=$2|" "$SSC_ENV_FILE" && rm -f "$SSC_ENV_FILE.bak"
        else
            echo "$1=$2" >> "$SSC_ENV_FILE"
        fi
    }

    update_env "JWT_SECRET" "$JWT_SECRET"
    update_env "REFRESH_SECRET" "$REFRESH_SECRET"
    update_env "CSRF_SECRET" "$CSRF_SECRET"
    update_env "SSC_FEATURES_ENCRYPTION_KEY" "$ENCRYPTION_KEY"
    update_env "SSC_ADMIN_PASSWORD" "$ADMIN_PASSWORD"
    update_env "SSC_PUBLIC_DOMAIN" "$DOMAIN"
    update_env "SSC_PUBLIC_URL" "https://$DOMAIN"
    update_env "SSC_SITE_NAME" "servasec"
    update_env "SSC_REGISTRATION_ENABLED" "$REGISTRATION_ENABLED"
    update_env "SSC_SEED_DATABASE" "true"

    # SSO
    if [ "$SSO_ENABLED" = true ]; then
        [ -n "$SSO_GITHUB_CLIENT_ID" ]     && update_env "SSO_GITHUB_CLIENT_ID" "$SSO_GITHUB_CLIENT_ID"
        [ -n "$SSO_GITHUB_CLIENT_SECRET" ]  && update_env "SSO_GITHUB_CLIENT_SECRET" "$SSO_GITHUB_CLIENT_SECRET"
        [ -n "$SSO_GITLAB_CLIENT_ID" ]      && update_env "SSO_GITLAB_CLIENT_ID" "$SSO_GITLAB_CLIENT_ID"
        [ -n "$SSO_GITLAB_CLIENT_SECRET" ]  && update_env "SSO_GITLAB_CLIENT_SECRET" "$SSO_GITLAB_CLIENT_SECRET"
        [ -n "$SSO_GITLAB_BASE_URL" ]       && update_env "SSO_GITLAB_BASE_URL" "$SSO_GITLAB_BASE_URL"
        [ -n "$SSO_OIDC_CLIENT_ID" ]        && update_env "SSO_OIDC_CLIENT_ID" "$SSO_OIDC_CLIENT_ID"
        [ -n "$SSO_OIDC_CLIENT_SECRET" ]    && update_env "SSO_OIDC_CLIENT_SECRET" "$SSO_OIDC_CLIENT_SECRET"
        [ -n "$SSO_OIDC_ISSUER_URL" ]       && update_env "SSO_OIDC_ISSUER_URL" "$SSO_OIDC_ISSUER_URL"
        [ -n "$SSO_OIDC_SCOPES" ]           && update_env "SSO_OIDC_SCOPES" "$SSO_OIDC_SCOPES"
    else
        sed -i.bak \
            -e 's|^SSO_GITHUB_CLIENT_ID=.*|# SSO_GITHUB_CLIENT_ID=|' \
            -e 's|^SSO_GITHUB_CLIENT_SECRET=.*|# SSO_GITHUB_CLIENT_SECRET=|' \
            -e 's|^SSO_GITLAB_CLIENT_ID=.*|# SSO_GITLAB_CLIENT_ID=|' \
            -e 's|^SSO_GITLAB_CLIENT_SECRET=.*|# SSO_GITLAB_CLIENT_SECRET=|' \
            -e 's|^SSO_GITLAB_BASE_URL=.*|# SSO_GITLAB_BASE_URL=|' \
            -e 's|^SSO_OIDC_CLIENT_ID=.*|# SSO_OIDC_CLIENT_ID=|' \
            -e 's|^SSO_OIDC_CLIENT_SECRET=.*|# SSO_OIDC_CLIENT_SECRET=|' \
            -e 's|^SSO_OIDC_ISSUER_URL=.*|# SSO_OIDC_ISSUER_URL=|' \
            -e 's|^SSO_OIDC_SCOPES=.*|# SSO_OIDC_SCOPES=|' \
            "$SSC_ENV_FILE" 2>/dev/null && rm -f "$SSC_ENV_FILE.bak"
    fi

    ok "Secrets written to .env"
fi

# ──────────────────────────────────────────────
#  8. Start stack
# ──────────────────────────────────────────────

printf "\n"
info "Building and starting Servasec..."
info "This may take a few minutes on first run..."
echo ""

if [ "$PRO_ENABLED" = true ]; then
    BUILD_TAGS=pro $SSC_COMPOSE up --build -d
else
    $SSC_COMPOSE up --build -d
fi

ok "Services started"

# ──────────────────────────────────────────────
#  9. Verify
# ──────────────────────────────────────────────

printf "\n"
info "Waiting for services to be ready..."

sleep 5

RUNNING=$($SSC_COMPOSE ps --services --filter status=running 2>/dev/null | wc -l)
if [ "$RUNNING" -ge 2 ]; then
    ok " services running"
else
    warn "Some services may not be running. Check logs:"
    warn "  $SSC_COMPOSE logs --tail=20"
fi

# ──────────────────────────────────────────────
#  10. Summary
# ──────────────────────────────────────────────

echo ""
echo "────────────────────────────────────────────────"
echo ""
if [ "$PRO_ENABLED" = true ]; then
    printf "  ${GREEN}${BOLD}servasec (Pro) is running!${NC}\n"
else
    printf "  ${GREEN}${BOLD}servasec is running!${NC}\n"
fi
echo ""
echo "  URL:      https://${DOMAIN}"
echo "  Login:    admin"
echo "  Password: ${ADMIN_PASSWORD}"
echo ""
echo "  Useful commands:"
echo "    View logs:    $SSC_COMPOSE logs -f"
echo "    Stop:         $SSC_COMPOSE down"
echo "    Upgrade:      bash ./scripts/upgrade.sh"
echo ""
echo "  Note: Change the admin password after first login."
echo ""
echo "────────────────────────────────────────────────"
