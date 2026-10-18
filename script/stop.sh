#!/bin/sh
set -eu

ROOT_DIR=$(CDPATH= cd "$(dirname "$0")/.." && pwd)
BINARY_PATH="${ROOT_DIR}/bin/self_media_script_tool"
OUTPUT_PATH="${ROOT_DIR}/logs/service.out.log"
LOG_DIR=$(dirname "${OUTPUT_PATH}")

if ! mkdir -p "${LOG_DIR}"; then
  printf 'Failed to create log directory: %s\n' "${LOG_DIR}" >&2
  exit 1
fi

log() {
  printf '[%s] %s\n' "$(date '+%Y-%m-%d %H:%M:%S')" "$*" | tee -a "${OUTPUT_PATH}"
}

log_error() {
  printf '[%s] %s\n' "$(date '+%Y-%m-%d %H:%M:%S')" "$*" | tee -a "${OUTPUT_PATH}" >&2
}

if pids=$(pgrep -f -x "${BINARY_PATH}"); then
  :
else
  pgrep_status=$?
  if [ "${pgrep_status}" -eq 1 ]; then
    log "Service is not running."
    exit 0
  fi
  log_error "Failed to find the service process (pgrep exit code: ${pgrep_status})."
  exit "${pgrep_status}"
fi

for pid in ${pids}; do
  if ! kill -TERM "${pid}" 2>/dev/null && kill -0 "${pid}" 2>/dev/null; then
    log_error "Failed to stop service process ${pid}."
    exit 1
  fi
done

for attempt in 1 2 3 4 5 6 7 8 9 10; do
  still_running=0
  for pid in ${pids}; do
    if kill -0 "${pid}" 2>/dev/null; then
      still_running=1
    fi
  done

  if [ "${still_running}" -eq 0 ]; then
    log "Service stopped."
    exit 0
  fi

  sleep 1
done

for pid in ${pids}; do
  if kill -0 "${pid}" 2>/dev/null; then
    if ! kill -KILL "${pid}" 2>/dev/null && kill -0 "${pid}" 2>/dev/null; then
      log_error "Failed to force-stop service process ${pid}."
      exit 1
    fi
  fi
done

log "Service stopped (forced after timeout)."
