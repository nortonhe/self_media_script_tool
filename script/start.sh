#!/bin/bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BINARY_PATH="${ROOT_DIR}/bin/self_media_script_tool"
OUTPUT_PATH="${ROOT_DIR}/logs/service.out.log"

log() {
  printf '[%s] %s\n' "$(date '+%Y-%m-%d %H:%M:%S')" "$*" | tee -a "${OUTPUT_PATH}"
}

log_error() {
  printf '[%s] %s\n' "$(date '+%Y-%m-%d %H:%M:%S')" "$*" | tee -a "${OUTPUT_PATH}" >&2
}

if ! mkdir -p "$(dirname "${OUTPUT_PATH}")"; then
  printf '[%s] Failed to create log directory: %s\n' "$(date '+%Y-%m-%d %H:%M:%S')" "$(dirname "${OUTPUT_PATH}")" >&2
  exit 1
fi

set +e
pids="$(pgrep -f -x "${BINARY_PATH}")"
pgrep_status=$?
set -e

if [[ ${pgrep_status} -eq 0 ]]; then
  log "Service is already running (PID: ${pids//$'\n'/, })."
  exit 0
elif [[ ${pgrep_status} -ne 1 ]]; then
  log_error "Failed to check service status (pgrep exit code: ${pgrep_status})."
  exit "${pgrep_status}"
fi

if [[ ! -x "${BINARY_PATH}" ]]; then
  log_error "Service binary not found or not executable: ${BINARY_PATH}"
  exit 1
fi

cd "${ROOT_DIR}"
nohup "${BINARY_PATH}" 2>&1 | tee -a "${OUTPUT_PATH}" &

for ((attempt = 0; attempt < 10; attempt++)); do
  sleep 0.2

  set +e
  pids="$(pgrep -f -x "${BINARY_PATH}")"
  pgrep_status=$?
  set -e

  if [[ ${pgrep_status} -eq 0 ]]; then
    log "Service started successfully (PID: ${pids//$'\n'/, })."
    exit 0
  elif [[ ${pgrep_status} -ne 1 ]]; then
    log_error "Failed to check service status (pgrep exit code: ${pgrep_status})."
    exit "${pgrep_status}"
  fi
done

log_error "Failed to start service. Check the log: ${OUTPUT_PATH}"
exit 1
