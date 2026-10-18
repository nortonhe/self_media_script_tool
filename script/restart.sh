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
  log "Service is running (PID: ${pids//$'\n'/, }); stopping it before restart."
  if ! bash "${ROOT_DIR}/script/stop.sh"; then
    log_error "Failed to stop service; restart aborted."
    exit 1
  fi
elif [[ ${pgrep_status} -eq 1 ]]; then
  log "Service is not running; starting it."
elif [[ ${pgrep_status} -ne 0 ]]; then
  log_error "Failed to check service status (pgrep exit code: ${pgrep_status})."
  exit "${pgrep_status}"
fi

if ! bash "${ROOT_DIR}/script/start.sh"; then
  log_error "Failed to start service during restart."
  exit 1
fi

log "Service restart completed."