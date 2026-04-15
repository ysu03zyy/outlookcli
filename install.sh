#!/usr/bin/env sh
set -eu

OWNER="${OUTLOOKCLI_OWNER:-ysu03zyy}"
REPO="${OUTLOOKCLI_REPO:-outlookcli}"
INSTALL_DIR="${OUTLOOKCLI_INSTALL_DIR:-/usr/local/bin}"
VERSION="${OUTLOOKCLI_VERSION:-}"

fail() {
  echo "Error: $*" >&2
  exit 1
}

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || fail "missing required command: $1"
}

detect_os() {
  os="$(uname -s | tr '[:upper:]' '[:lower:]')"
  case "$os" in
    darwin) echo "darwin" ;;
    linux) echo "linux" ;;
    *) fail "unsupported OS: $os" ;;
  esac
}

detect_arch() {
  arch="$(uname -m)"
  case "$arch" in
    x86_64|amd64) echo "amd64" ;;
    arm64|aarch64) echo "arm64" ;;
    *) fail "unsupported architecture: $arch" ;;
  esac
}

latest_version() {
  url="https://api.github.com/repos/${OWNER}/${REPO}/releases/latest"
  tag="$(curl -fsSL "$url" | sed -n 's/.*"tag_name":[[:space:]]*"\([^"]*\)".*/\1/p' | head -n 1)"
  [ -n "$tag" ] || fail "cannot determine latest release tag"
  echo "${tag#v}"
}

main() {
  need_cmd curl
  need_cmd tar
  need_cmd shasum

  os="$(detect_os)"
  arch="$(detect_arch)"
  version="$VERSION"
  if [ -z "$version" ]; then
    version="$(latest_version)"
  fi

  archive="outlookcli_${version}_${os}_${arch}.tar.gz"
  checksums="checksums.txt"
  base_url="https://github.com/${OWNER}/${REPO}/releases/download/v${version}"
  tmpdir="$(mktemp -d)"
  trap 'rm -rf "$tmpdir"' EXIT INT TERM

  echo "Installing outlookcli v${version} for ${os}/${arch}..."
  curl -fsSL -o "${tmpdir}/${archive}" "${base_url}/${archive}"
  curl -fsSL -o "${tmpdir}/${checksums}" "${base_url}/${checksums}"

  (
    cd "$tmpdir"
    expected="$(awk -v file="$archive" '$2 == file { print $1 }' "$checksums")"
    [ -n "$expected" ] || fail "checksum entry not found for ${archive}"
    actual="$(shasum -a 256 "$archive" | awk '{print $1}')"
    [ "$expected" = "$actual" ] || fail "checksum mismatch for ${archive}"
  )

  tar -xzf "${tmpdir}/${archive}" -C "$tmpdir"
  [ -f "${tmpdir}/outlookcli" ] || fail "binary not found in archive: ${archive}"

  if [ ! -w "$INSTALL_DIR" ]; then
    echo "Need elevated permissions to write to ${INSTALL_DIR}, using sudo..."
    sudo mkdir -p "$INSTALL_DIR"
    sudo install -m 0755 "${tmpdir}/outlookcli" "${INSTALL_DIR}/outlookcli"
  else
    mkdir -p "$INSTALL_DIR"
    install -m 0755 "${tmpdir}/outlookcli" "${INSTALL_DIR}/outlookcli"
  fi

  echo "Installed: ${INSTALL_DIR}/outlookcli"
  "${INSTALL_DIR}/outlookcli" --version
}

main "$@"
