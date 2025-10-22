#!/usr/bin/env bash
set -euo pipefail

# Installs the nqcli binary by downloading the appropriate release artifact
# from GitHub. Intended for usage like:
#   curl -fsSL https://raw.githubusercontent.com/ankit-lilly/nqcli/main/scripts/install.sh | bash

REPO="${REPO:-ankit-lilly/nqcli}"
BINARY_NAME="${BINARY_NAME:-nq}"
export BINARY_NAME
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
VERSION="${VERSION:-latest}"
BINARY_FALLBACK="${BINARY_FALLBACK:-${BINARY_NAME}cli}"
export BINARY_FALLBACK

log() {
  printf '==> %s\n' "$*"
}

err() {
  printf 'error: %s\n' "$*" >&2
  exit 1
}

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    err "required command '$1' not found"
  fi
}

detect_python() {
  if command -v python3 >/dev/null 2>&1; then
    PYTHON_BIN=python3
  elif command -v python >/dev/null 2>&1; then
    PYTHON_BIN=python
  else
    err "python3 (or python) is required to parse the GitHub API response"
  fi
}

detect_platform() {
  case "$(uname -s)" in
    Darwin) TARGET_OS=darwin ;;
    Linux) TARGET_OS=linux ;;
    *)
      err "unsupported OS: $(uname -s)"
      ;;
  esac

  case "$(uname -m)" in
    x86_64 | amd64) TARGET_ARCH=amd64 ;;
    arm64 | aarch64) TARGET_ARCH=arm64 ;;
    *)
      err "unsupported architecture: $(uname -m)"
      ;;
  esac
  export TARGET_OS TARGET_ARCH
}

fetch_release_json() {
  local api_url curl_args
  if [[ "${VERSION}" == "latest" ]]; then
    api_url="https://api.github.com/repos/${REPO}/releases/latest"
  else
    api_url="https://api.github.com/repos/${REPO}/releases/tags/${VERSION}"
  fi

  curl_args=(-fsSL -H "Accept: application/vnd.github+json" -H "User-Agent: ${USER_AGENT:-nqcli-installer}")
  if [[ -n "${GITHUB_TOKEN:-}" ]]; then
    curl_args+=(-H "Authorization: Bearer ${GITHUB_TOKEN}")
  fi

  release_json=$(curl "${curl_args[@]}" "${api_url}") || err "failed to fetch release metadata from ${api_url}"
  if [[ -z "${release_json}" ]]; then
    err "received empty response from GitHub API; check network access or ensure releases exist for ${REPO}"
  fi
}

select_asset() {
  local asset_info
  asset_info=$(printf '%s' "${release_json}" | "${PYTHON_BIN}" <<'PY'
import json
import os
import sys

data = json.load(sys.stdin)
assets = data.get("assets", [])
target_os = os.environ["TARGET_OS"].lower()
target_arch = os.environ["TARGET_ARCH"].lower()
binary_name = os.environ["BINARY_NAME"].lower()
fallback_name = os.environ.get("BINARY_FALLBACK", (binary_name + "cli"))

def match_asset(asset, *extra_terms):
    name = asset.get("name", "").lower()
    terms = [target_os, target_arch] + [term for term in extra_terms if term]
    return all(term in name for term in terms)

for candidate in assets:
    if match_asset(candidate, binary_name):
        print(candidate["name"])
        print(candidate["browser_download_url"])
        sys.exit(0)

for candidate in assets:
    if match_asset(candidate, fallback_name):
        print(candidate["name"])
        print(candidate["browser_download_url"])
        sys.exit(0)

sys.exit("no matching asset found for os='{}' arch='{}'".format(target_os, target_arch))
PY
) || {
    err "$(printf 'failed to locate a release asset for %s/%s\n%s' "${TARGET_OS}" "${TARGET_ARCH}" "${asset_info}")"
  }

  IFS=$'\n' read -r ASSET_NAME ASSET_URL <<<"${asset_info}"
  if [[ -z "${ASSET_URL}" ]]; then
    err "release asset URL was empty"
  fi
}

download_asset() {
  TMP_DIR=$(mktemp -d)
  trap 'rm -rf "$TMP_DIR"' EXIT

  DOWNLOAD_PATH="${TMP_DIR}/${ASSET_NAME:-download}"
  curl -fL --retry 3 --retry-delay 2 -o "${DOWNLOAD_PATH}" "${ASSET_URL}" || err "failed to download ${ASSET_URL}"
}

extract_binary() {
  case "${DOWNLOAD_PATH}" in
    *.tar.gz | *.tgz)
      require_cmd tar
      tar -xzf "${DOWNLOAD_PATH}" -C "${TMP_DIR}"
      ;;
    *.zip)
      require_cmd unzip
      unzip -qo "${DOWNLOAD_PATH}" -d "${TMP_DIR}"
      ;;
  esac

  local found
  found=$(find "${TMP_DIR}" -maxdepth 3 -type f \
    \( -name "${BINARY_NAME}" -o -name "${BINARY_NAME}*" -o -name "${BINARY_FALLBACK}" -o -name "${BINARY_FALLBACK}*" \) \
    ! -name "*.tar" ! -name "*.gz" ! -name "*.zip" ! -name "*.tgz" \
    -print -quit)
  if [[ -z "${found}" ]]; then
    err "could not locate ${BINARY_NAME} inside downloaded artifact"
  fi

  BINARY_PATH="${found}"
  chmod +x "${BINARY_PATH}"
}

maybe_clear_quarantine() {
  if [[ "${TARGET_OS}" != "darwin" ]]; then
    return
  fi

  if command -v xattr >/dev/null 2>&1; then
    if xattr -p com.apple.quarantine "${BINARY_PATH}" >/dev/null 2>&1; then
      log "Removing macOS quarantine attribute"
      xattr -d com.apple.quarantine "${BINARY_PATH}" || log "warning: failed to clear com.apple.quarantine; continuing"
    fi
  else
    log "warning: xattr command not available; skipping quarantine removal"
  fi
}

install_binary() {
  log "Installing ${BINARY_NAME} to ${INSTALL_DIR}"
  mkdir -p "${INSTALL_DIR}"

  local target="${INSTALL_DIR}/${BINARY_NAME}"
  if [[ -w "${INSTALL_DIR}" ]]; then
    install -m 0755 "${BINARY_PATH}" "${target}"
  else
    if command -v sudo >/dev/null 2>&1; then
      sudo install -m 0755 "${BINARY_PATH}" "${target}"
    else
      err "${INSTALL_DIR} is not writable; re-run with sudo or set INSTALL_DIR to a user-writable path"
    fi
  fi

  log "Installed ${BINARY_NAME} -> ${target}"
  log "Ensure ${INSTALL_DIR} is on your PATH."
}

main() {
  require_cmd curl
  require_cmd install
  detect_python
  detect_platform
  fetch_release_json
  select_asset
  download_asset
  extract_binary
  maybe_clear_quarantine
  install_binary
}

main "$@"
