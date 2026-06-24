#!/usr/bin/env bash
# Pre-Release Validation Script for gymfony/di
# Runs all quality checks before creating a release.
# Mirrors CI checks plus additional local validations.
#
# Usage:
#   bash scripts/pre-release-check.sh          # full check
#   bash scripts/pre-release-check.sh --quick  # skip slow steps (tests, lint)

set -e

QUICK=false
[[ "${1:-}" == "--quick" ]] && QUICK=true

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info()    { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
log_error()   { echo -e "${RED}[ERROR]${NC} $1"; }

echo ""
echo "================================================"
echo "  gymfony/di — Pre-Release Check"
echo "================================================"
echo ""

ERRORS=0
WARNINGS=0

# 1. Go version
log_info "Checking Go version..."
GO_VERSION=$(go version | awk '{print $3}')
REQUIRED="go1.26"
if [[ "$GO_VERSION" < "$REQUIRED" ]]; then
    log_error "Go $REQUIRED+ required, found $GO_VERSION"
    ERRORS=$((ERRORS + 1))
else
    log_success "Go version: $GO_VERSION"
fi
echo ""

# 2. Git status
log_info "Checking git status..."
if git diff-index --quiet HEAD --; then
    log_success "Working directory is clean"
else
    log_warning "Uncommitted changes detected"
    git status --short
    WARNINGS=$((WARNINGS + 1))
fi
echo ""

# 3. Formatting
log_info "Checking formatting (gofmt)..."
UNFORMATTED=$(gofmt -l .)
if [ -n "$UNFORMATTED" ]; then
    log_error "Files need formatting:"
    echo "$UNFORMATTED"
    log_info "Run: go fmt ./..."
    ERRORS=$((ERRORS + 1))
else
    log_success "All files are properly formatted"
fi
echo ""

# 4. go vet
log_info "Running go vet..."
if go vet ./... 2>&1; then
    log_success "go vet passed"
else
    log_error "go vet failed"
    ERRORS=$((ERRORS + 1))
fi
echo ""

# 5. go.mod / go.sum
log_info "Validating go.mod..."
go mod verify
if [ $? -eq 0 ]; then
    log_success "go.mod verified"
else
    log_error "go.mod verification failed"
    ERRORS=$((ERRORS + 1))
fi

go mod tidy
if git diff --quiet go.mod go.sum 2>/dev/null; then
    log_success "go.mod is tidy"
else
    log_warning "go.mod needs tidying (run 'go mod tidy')"
    git diff go.mod go.sum 2>/dev/null || true
    WARNINGS=$((WARNINGS + 1))
fi
echo ""

if [ "$QUICK" = false ]; then
    # 6. golangci-lint
    log_info "Verifying golangci-lint config..."
    if command -v golangci-lint &>/dev/null; then
        LINT_OUT=$(golangci-lint config verify 2>&1 || true)
        if echo "$LINT_OUT" | grep -qE "deadline exceeded|connection refused|no such host"; then
            log_warning "golangci-lint config verify skipped (no network)"
            WARNINGS=$((WARNINGS + 1))
        elif [ -n "$LINT_OUT" ]; then
            log_error "golangci-lint config invalid: $LINT_OUT"
            ERRORS=$((ERRORS + 1))
        else
            log_success "golangci-lint config valid"
        fi

        log_info "Running golangci-lint..."
        if golangci-lint run --timeout=5m ./... 2>&1; then
            log_success "golangci-lint passed"
        else
            log_error "golangci-lint found issues"
            ERRORS=$((ERRORS + 1))
        fi
    else
        log_warning "golangci-lint not installed — https://golangci-lint.run/welcome/install/"
        WARNINGS=$((WARNINGS + 1))
    fi
    echo ""

    # 7. Tests with race detector
    log_info "Running tests with race detector..."
    TEST_TMP=$(mktemp)
    go test -v -race ./... 2>&1 | tee "$TEST_TMP" || true
    TEST_OUT=$(cat "$TEST_TMP"); rm -f "$TEST_TMP"
    if echo "$TEST_OUT" | grep -q "FAIL"; then
        log_error "Tests failed"
        ERRORS=$((ERRORS + 1))
    else
        log_success "All tests passed"
    fi
    echo ""

    # 8. Coverage
    log_info "Checking test coverage (target ≥85%)..."
    go test -count=1 -coverprofile=/tmp/di-coverage.txt -covermode=atomic ./... 2>/dev/null || true

    COVERAGE=""
    if [ -f /tmp/di-coverage.txt ]; then
        COVERAGE=$(go tool cover -func=/tmp/di-coverage.txt \
            | grep "^total:" | awk '{print $3}' | sed 's/%//')
    fi

    if [ -n "$COVERAGE" ]; then
        echo "  total coverage: ${COVERAGE}%"
        if awk -v cov="$COVERAGE" 'BEGIN {exit !(cov >= 85.0)}'; then
            log_success "Coverage meets requirement (≥85%)"
        else
            log_warning "Coverage below 85% (${COVERAGE}%)"
            WARNINGS=$((WARNINGS + 1))
        fi
    else
        log_warning "Could not determine coverage (no tests yet?)"
        WARNINGS=$((WARNINGS + 1))
    fi
    rm -f /tmp/di-coverage.txt
    echo ""
fi

# 9. Required files
log_info "Checking required files..."
MISSING=0
for f in README.md LICENSE CONTRIBUTING.md go.mod; do
    if [ ! -f "$f" ]; then
        log_error "Missing: $f"
        MISSING=1
        ERRORS=$((ERRORS + 1))
    fi
done
[ $MISSING -eq 0 ] && log_success "All required files present"
echo ""

# Summary
echo "========================================"
echo "  Summary"
echo "========================================"
echo ""

if [ $ERRORS -eq 0 ] && [ $WARNINGS -eq 0 ]; then
    log_success "All checks passed — ready for release."
    exit 0
elif [ $ERRORS -eq 0 ]; then
    log_warning "Completed with $WARNINGS warning(s) — review before releasing."
    exit 0
else
    log_error "Failed with $ERRORS error(s) and $WARNINGS warning(s) — fix before releasing."
    exit 1
fi
