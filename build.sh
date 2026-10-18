#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUTPUT_DIR="${ROOT_DIR}/bin"
BINARY_NAME="self_media_script_tool"
OUTPUT_PATH="${OUTPUT_DIR}/${BINARY_NAME}"

mkdir -p "${OUTPUT_DIR}"

echo "[1/2] Building Go project: ${BINARY_NAME}"
go build -o "${OUTPUT_PATH}" "${ROOT_DIR}"

echo "[2/2] Build succeeded: ${OUTPUT_PATH}"
