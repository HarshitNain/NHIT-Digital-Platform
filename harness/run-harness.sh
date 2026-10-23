#!/usr/bin/env bash
# ==============================================================================
#   NATIONAL HIGHWAYS INFRA TRUST (NHIT) — ENTERPRISE TEST HARNESS RUNNER (POSIX)
# ==============================================================================
#   Automated multi-phase invariant verification and health enforcement suite.
#
#   Usage:
#     ./harness/run-harness.sh [--fast] [--bail] [--phase <1-7>] [--json] [--html <file>] [--markdown <file>] [--verbose]
# ==============================================================================

set -eo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
HARNESS_PY="${REPO_ROOT}/NHIT Backend/scripts/harness.py"
ENV_FILE="${SCRIPT_DIR}/configs/test.env"
REPORTS_DIR="${SCRIPT_DIR}/reports"

mkdir -p "${REPORTS_DIR}"

DEFAULT_HTML="${REPORTS_DIR}/harness_report.html"
DEFAULT_MD="${REPORTS_DIR}/harness_report.md"

# Load environment configuration if available
if [ -f "${ENV_FILE}" ]; then
  while IFS='=' read -r key val || [ -n "$key" ]; do
    key=$(echo "$key" | xargs)
    if [[ -n "$key" && ! "$key" =~ ^# ]]; then
      val=$(echo "$val" | xargs | sed -e 's/^"//' -e 's/"$//' -e "s/^'//" -e "s/'$//")
      if [ -z "${!key}" ]; then
        export "${key}=${val}"
      fi
    fi
  done < "${ENV_FILE}"
fi

# Resolve Python executable
if command -v python3 >/dev/null 2>&1; then
  PYTHON_EXE="python3"
elif command -v python >/dev/null 2>&1; then
  PYTHON_EXE="python"
else
  echo "[ERROR] Python 3.10+ runtime not found in PATH."
  exit 1
fi

ARGS=()
HAS_HTML=0
HAS_MD=0

while [[ $# -gt 0 ]]; do
  case "$1" in
    --fast)
      ARGS+=("--fast")
      shift
      ;;
    --bail)
      ARGS+=("--bail")
      shift
      ;;
    --phase)
      ARGS+=("--phase" "$2")
      shift 2
      ;;
    --json)
      ARGS+=("--json")
      shift
      ;;
    --html)
      ARGS+=("--html" "$2")
      HAS_HTML=1
      shift 2
      ;;
    --markdown)
      ARGS+=("--markdown" "$2")
      HAS_MD=1
      shift 2
      ;;
    --verbose)
      ARGS+=("--verbose")
      shift
      ;;
    *)
      ARGS+=("$1")
      shift
      ;;
  esac
done

if [ $HAS_HTML -eq 0 ]; then
  ARGS+=("--html" "${DEFAULT_HTML}")
fi
if [ $HAS_MD -eq 0 ]; then
  ARGS+=("--markdown" "${DEFAULT_MD}")
fi

"${PYTHON_EXE}" "${HARNESS_PY}" "${ARGS[@]}"
