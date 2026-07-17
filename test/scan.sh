#!/usr/bin/env bash
set -uo pipefail

REPORTS_DIR="/reports"
TARGET_DIR="/target/files"
TARGET_IMAGE="/target/image.tar"

mkdir -p "$REPORTS_DIR"

echo "=== servasec scanner fixture generator ==="
echo ""

# --- Container scanners ---
if [ -f "$TARGET_IMAGE" ]; then
    echo "[scanning] grype (container image)..."
    grype docker-archive:"$TARGET_IMAGE" -o json > "$REPORTS_DIR/grype.json" 2>/dev/null \
        || echo '{"matches":[]}' > "$REPORTS_DIR/grype.json"
    echo "  → OK"

    echo "[scanning] trivy (container image)..."
    trivy image --input "$TARGET_IMAGE" --format json > "$REPORTS_DIR/trivy.json" 2>/dev/null \
        || echo '{"Results":[]}' > "$REPORTS_DIR/trivy.json"
    echo "  → OK"

    echo "[scanning] trivy-iac..."
    trivy config --format json "$TARGET_DIR" > "$REPORTS_DIR/trivy-iac.json" 2>/dev/null \
        || echo '{"Results":[]}' > "$REPORTS_DIR/trivy-iac.json"
    echo "  → OK"
else
    echo "[skip] grype, trivy (no image.tar)"
    echo '{"matches":[]}' > "$REPORTS_DIR/grype.json"
    echo '{"Results":[]}' > "$REPORTS_DIR/trivy.json"

    echo "[scanning] trivy-iac..."
    trivy config --format json "$TARGET_DIR" > "$REPORTS_DIR/trivy-iac.json" 2>/dev/null \
        || echo '{"Results":[]}' > "$REPORTS_DIR/trivy-iac.json"
    echo "  → OK"
fi

# --- Python SAST ---
# bandit exits 1 when findings exist — only treat >1 as failure
echo "[scanning] bandit..."
bandit_exit=0
bandit -r "$TARGET_DIR" -f json > "$REPORTS_DIR/bandit.json" 2>/dev/null || bandit_exit=$?
if [ "$bandit_exit" -le 1 ]; then
    echo "  → OK (exit $bandit_exit)"
else
    echo '{"results":[]}' > "$REPORTS_DIR/bandit.json"
    echo "  → fallback (exit $bandit_exit)"
fi

# --- Go SAST ---
echo "[scanning] gosec..."
if find "$TARGET_DIR" -name '*.go' -print -quit 2>/dev/null | grep -q .; then
    gosec -fmt=json -out="$REPORTS_DIR/gosec.json" "$TARGET_DIR" 2>/dev/null \
        || echo '{"Issues":[]}' > "$REPORTS_DIR/gosec.json"
else
    echo '{"Issues":[]}' > "$REPORTS_DIR/gosec.json"
fi
echo "  → OK"

# --- IaC scanners ---
# checkov: use --output json (not -o), --compact to reduce verbosity
echo "[scanning] checkov..."
checkov -d "$TARGET_DIR" --output json --compact 2>/dev/null \
    | jq '{results:{failed_checks:(.results.failed_checks // [])}}' \
    > "$REPORTS_DIR/checkov.json" \
    || echo '{"results":{"failed_checks":[]}}' > "$REPORTS_DIR/checkov.json"
echo "  → OK"

# tfsec: --format json may not exist in all versions; capture stdout
echo "[scanning] tfsec..."
tfsec "$TARGET_DIR" --format json 2>/dev/null \
    > "$REPORTS_DIR/tfsec.json" \
    || echo '{"results":{"passed":[],"failed":[]}}' > "$REPORTS_DIR/tfsec.json"
echo "  → OK"

# --- Secret scanners ---
# gitleaks: dir subcommand needs --report-path for file output
echo "[scanning] gitleaks..."
gitleaks dir "$TARGET_DIR" --report-format json --report-path "$REPORTS_DIR/gitleaks.json" --no-banner 2>/dev/null \
    || [ -f "$REPORTS_DIR/gitleaks.json" ] \
    || echo '[]' > "$REPORTS_DIR/gitleaks.json"
echo "  → OK"

# --- Dependency scanners ---
echo "[scanning] osv-scanner..."
OSVArgs=()
if [ -f "$TARGET_DIR/package-lock.json" ]; then
    OSVArgs+=(--lockfile "$TARGET_DIR/package-lock.json")
fi
if [ -f "$TARGET_DIR/requirements.txt" ]; then
    OSVArgs+=(--lockfile "$TARGET_DIR/requirements.txt")
fi
if [ ${#OSVArgs[@]} -gt 0 ]; then
    osv-scanner --json "${OSVArgs[@]}" > "$REPORTS_DIR/osv-scanner.json" 2>/dev/null \
        || echo '{"results":[]}' > "$REPORTS_DIR/osv-scanner.json"
else
    echo '{"results":[]}' > "$REPORTS_DIR/osv-scanner.json"
fi
echo "  → OK"

# --- SAST (optional, heavy) ---
echo "[scanning] semgrep..."
semgrep --config auto --json --output "$REPORTS_DIR/semgrep.json" "$TARGET_DIR" 2>/dev/null \
    || echo '{"results":[]}' > "$REPORTS_DIR/semgrep.json"
echo "  → OK"

# --- DAST (nuclei needs running targets, generate empty fixture) ---
echo "[skip] nuclei (DAST needs running targets; generate fixture manually)"
echo '[]' > "$REPORTS_DIR/nuclei.json"

echo ""
echo "=== Done ==="
ls -1 "$REPORTS_DIR"/*.json 2>/dev/null | while read f; do
    size=$(wc -c < "$f")
    echo "  $(basename "$f")  (${size} bytes)"
done
