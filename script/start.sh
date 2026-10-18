#!/bin/bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BINARY_PATH="${ROOT_DIR}/bin/self_media_script_tool"

#bash "${ROOT_DIR}/build.sh"
exec "${BINARY_PATH}"
