#!/usr/bin/env bash

sudo apt update && sudo apt upgrade -y
sudo apt install -y zstd

# install-firecracker.sh
# Downloads the latest Firecracker release and installs `firecracker` + `jailer`
# into /usr/local/bin. Works on Linux x86_64 and aarch64.
#
# Usage:
#   chmod +x install-firecracker.sh
#   ./install-firecracker.sh                 # latest stable
#   ./install-firecracker.sh v1.15.1         # pin a version

set -euo pipefail

# --- resolve version ----------------------------------------------------------
if [[ $# -ge 1 ]]; then
  VERSION="$1"
else
  # follow the /releases/latest redirect to learn the tag, no jq needed
  VERSION=$(basename "$(curl -fsSLI -o /dev/null -w '%{url_effective}' \
    https://github.com/firecracker-microvm/firecracker/releases/latest)")
fi

ARCH="$(uname -m)"
case "$ARCH" in
  x86_64|aarch64) ;;
  *) echo "unsupported arch: $ARCH (need x86_64 or aarch64)" >&2; exit 1 ;;
esac

TARBALL="firecracker-${VERSION}-${ARCH}.tgz"
URL="https://github.com/firecracker-microvm/firecracker/releases/download/${VERSION}/${TARBALL}"

echo ">> installing Firecracker ${VERSION} for ${ARCH}"
echo ">> source: ${URL}"

# --- download into a tempdir --------------------------------------------------
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

curl -fsSL -o "${TMP}/${TARBALL}" "${URL}"
tar -xzf "${TMP}/${TARBALL}" -C "${TMP}"

EXTRACTED="${TMP}/release-${VERSION}-${ARCH}"
[[ -d "$EXTRACTED" ]] || { echo "extraction failed: ${EXTRACTED} not found" >&2; exit 1; }

# --- install binaries ---------------------------------------------------------
SUDO=""
[[ $EUID -ne 0 ]] && SUDO="sudo"

$SUDO install -m 0755 "${EXTRACTED}/firecracker-${VERSION}-${ARCH}" /usr/local/bin/firecracker
$SUDO install -m 0755 "${EXTRACTED}/jailer-${VERSION}-${ARCH}"      /usr/local/bin/jailer

echo
echo ">> installed:"
firecracker --version
jailer --version

# --- KVM access setup ---------------------------------------------------------
echo
if [[ ! -e /dev/kvm ]]; then
  echo ">> warning: /dev/kvm does not exist — KVM module not loaded or no virt support"
  echo "   check: lsmod | grep kvm  &&  egrep -c '(vmx|svm)' /proc/cpuinfo"
elif [[ -r /dev/kvm && -w /dev/kvm ]]; then
  echo ">> /dev/kvm is accessible — ready to boot microVMs"
else
  TARGET_USER="${SUDO_USER:-$(whoami)}"
  echo ">> /dev/kvm exists but ${TARGET_USER} can't access it — fixing group membership"

  # ensure the kvm group exists
  if ! getent group kvm >/dev/null; then
    $SUDO groupadd -r kvm
  fi

  # make sure the device is owned by group kvm with 0660 (it usually is on Ubuntu, but be safe)
  $SUDO chgrp kvm /dev/kvm
  $SUDO chmod 0660 /dev/kvm

  # add user to the group
  if id -nG "$TARGET_USER" | tr ' ' '\n' | grep -qx kvm; then
    echo "   ${TARGET_USER} is already in the kvm group — current shell just hasn't picked it up"
  else
    $SUDO usermod -aG kvm "$TARGET_USER"
    echo "   added ${TARGET_USER} to the kvm group"
  fi

  echo
  echo ">> NOTE: group changes don't apply to your current shell."
  echo "   run one of:"
  echo "     newgrp kvm        # activates the group in this shell"
  echo "     exit + re-login   # cleaner, applies everywhere"
  echo "   then verify: [ -r /dev/kvm ] && [ -w /dev/kvm ] && echo OK"
fi
