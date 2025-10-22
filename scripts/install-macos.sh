#!/usr/bin/env bash
set -euo pipefail

SCRIPT_NAME=$(basename "$0")
INSTALL_DIR=${INSTALL_DIR:-/usr/local/bin}
BINARY_TARGET="${INSTALL_DIR}/nq"

print_usage() {
  cat <<EOF
Usage: ${SCRIPT_NAME} /path/to/nq

Downloads from the GitHub Releases page are marked as quarantined by macOS.
This helper script clears the quarantine attribute and installs the nq binary
into ${INSTALL_DIR} (override with INSTALL_DIR env var).

Example:
  chmod +x ${SCRIPT_NAME}
  ./${SCRIPT_NAME} ~/Downloads/nq
EOF
}

if [[ "${1:-}" =~ ^(-h|--help)$ ]]; then
  print_usage
  exit 0
fi

if [[ $# -ne 1 ]]; then
  echo "error: missing path to nq binary" >&2
  print_usage >&2
  exit 1
fi

SOURCE_BINARY=$1

if [[ ! -f "${SOURCE_BINARY}" ]]; then
  echo "error: ${SOURCE_BINARY} does not exist or is not a regular file" >&2
  exit 1
fi

if [[ "${OSTYPE:-}" != darwin* ]]; then
  echo "warning: script intended for macOS; detected OSTYPE='${OSTYPE:-unknown}'" >&2
fi

if command -v xattr >/dev/null 2>&1; then
  if xattr -p com.apple.quarantine "${SOURCE_BINARY}" >/dev/null 2>&1; then
    echo "Removing com.apple.quarantine attribute..."
    xattr -d com.apple.quarantine "${SOURCE_BINARY}" || {
      echo "warning: failed to remove com.apple.quarantine (continuing)" >&2
    }
  fi
else
  echo "warning: xattr command not found; skipping quarantine removal" >&2
fi

chmod +x "${SOURCE_BINARY}"

mkdir -p "${INSTALL_DIR}"

if [[ -w "${INSTALL_DIR}" ]]; then
  install -m 0755 "${SOURCE_BINARY}" "${BINARY_TARGET}"
else
  if command -v sudo >/dev/null 2>&1; then
    echo "Installing nq to ${BINARY_TARGET} (using sudo)..."
    sudo install -m 0755 "${SOURCE_BINARY}" "${BINARY_TARGET}"
  else
    echo "error: ${INSTALL_DIR} not writable and sudo not available; set INSTALL_DIR to a writable directory" >&2
    exit 1
  fi
fi

echo "nq installed to ${BINARY_TARGET}"
echo "Ensure ${INSTALL_DIR} is on your PATH."
