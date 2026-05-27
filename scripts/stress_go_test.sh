#!/usr/bin/env bash
set -euo pipefail

usage() {
    cat <<'EOF'
Stress test a single Go test with sensible defaults.

Usage:
  scripts/stress_go_test.sh <go-package> <test-name-or-regex>

Examples:
  scripts/stress_go_test.sh ./internal/jimm/juju TestWatcherClearsControllerUnavailable
  scripts/stress_go_test.sh ./internal/jimm/jobs '^TestListJobs_ErrorState$'

Environment variables:
  STRESS_COUNT    Number of runs before stopping (default: 200)
  STRESS_P        Parallel workers (default: 8)
  STRESS_TIMEOUT  Per-run timeout passed to stress (default: 10m)
  STRESS_OUT      Failure log output prefix (default: /tmp/go-stress-<test>-)
  STRESS_EXACT    If "1", wraps test arg as ^<name>$. If "0", uses argument as regex (default: 1)
  STRESS_BIN      Compiled test binary path (default: /tmp/<pkg>.test)

Requirements:
  - stress must be available in PATH.
  - go must be installed.
EOF
}

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
    usage
    exit 0
fi

if [[ $# -lt 2 ]]; then
    usage
    exit 2
fi

pkg="$1"
test_arg="$2"

if ! command -v go >/dev/null 2>&1; then
    echo "error: go is not available in PATH" >&2
    exit 1
fi

if ! command -v stress >/dev/null 2>&1; then
    echo "error: stress is not available in PATH" >&2
    exit 1
fi

count="${STRESS_COUNT:-200}"
parallel="${STRESS_P:-8}"
timeout="${STRESS_TIMEOUT:-10m}"
exact="${STRESS_EXACT:-1}"

safe_pkg_name="$(echo "$pkg" | tr '/.' '__')"
default_bin="/tmp/${safe_pkg_name}.test"
bin_path="${STRESS_BIN:-$default_bin}"

if [[ "$exact" == "1" ]]; then
    test_regex="^${test_arg}$"
else
    test_regex="$test_arg"
fi

safe_test_name="$(echo "$test_arg" | sed 's/[^a-zA-Z0-9._-]/_/g')"
out_prefix_default="/tmp/go-stress-${safe_test_name}-"
out_prefix="${STRESS_OUT:-$out_prefix_default}"

echo "Compiling test binary: pkg=$pkg -> $bin_path"
go test -c "$pkg" -o "$bin_path"

echo "Running stress: test.run=$test_regex, count=$count, p=$parallel, timeout=$timeout"
stress -p "$parallel" -count "$count" -timeout "$timeout" -o "$out_prefix" "$bin_path" -test.run "$test_regex" -test.count=1
