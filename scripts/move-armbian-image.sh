#!/bin/bash
#
# Move the armbian .img to build/.
#
set -eu

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# TODO(hkjn): Consider defining these config values somewhere higher-level than inside this script.
KERNEL_VERSION="${KERNEL_VERSION:-4.4.182}"
ARMBIAN_VERSION="${ARMBIAN_VERSION:-5.91}"
IMG_FILE="${SCRIPT_DIR}/../armbian/armbian-build/output/images/Armbian_${ARMBIAN_VERSION}_Rockpro64_Debian_buster_default_${KERNEL_VERSION}.img"
TARGET_FILE="BitBoxBase_Armbian_RockPro64.img"

if [[ ! -e "${IMG_FILE}" ]]; then
  echo "fatal: no Armbian image exists at expected path:" >&2
  echo "  ${IMG_FILE}" >&2
  exit 1
fi
mv -v "${IMG_FILE}" "${SCRIPT_DIR}/../provisioning/${TARGET_FILE}"
# Image moved successfully.
exit 0
