#!/usr/bin/env bash
# Full-stack E2E orchestration: start Go back-end + React Router front-end,
# run Probitas scenarios, then tear down both servers.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Environment variables for E2E testing
export VIBRA_LISTEN_ADDR=127.0.0.1:3001
export VIBRA_CORS_ORIGIN=http://127.0.0.1:3000
export VIBRA_DEV_USER=e2e-test-user
export VIBRA_ALLOWED_DIRS=/tmp/vibra-e2e
export VIBRA_DEFAULT_WORKDIR=/tmp/vibra-e2e
export PORT=3000

mkdir -p /tmp/vibra-e2e

BACK_PID=""
FRONT_PID=""

cleanup() {
  echo "Tearing down servers..."
  if [[ -n "$BACK_PID" ]]; then
    kill "$BACK_PID" 2>/dev/null || true
    # Kill child processes spawned by go run (compiled binary)
    pkill -P "$BACK_PID" 2>/dev/null || true
  fi
  if [[ -n "$FRONT_PID" ]]; then
    kill "$FRONT_PID" 2>/dev/null || true
    pkill -P "$FRONT_PID" 2>/dev/null || true
  fi
}
trap cleanup EXIT

# wait_for_server polls url every 500ms until it responds or 30s elapses.
# On process failure (pid exits early), prints an error and returns 1.
wait_for_server() {
  local name="$1"
  local url="$2"
  local pid="$3"
  local max_attempts=60 # 60 * 0.5s = 30s
  local attempt=0

  echo "Waiting for $name at $url..."
  while ! curl -s -o /dev/null "$url" 2>/dev/null; do
    if ! kill -0 "$pid" 2>/dev/null; then
      echo "ERROR: $name process (PID $pid) exited unexpectedly" >&2
      return 1
    fi
    sleep 0.5
    attempt=$((attempt + 1))
    if [[ $attempt -ge $max_attempts ]]; then
      echo "ERROR: $name at $url did not become ready within 30s" >&2
      return 1
    fi
  done
  echo "$name is ready."
}

# Build the React Router front-end before starting any servers.
echo "Building React Router front-end..."
cd "$REPO_ROOT/front"
if ! pnpm build 2>&1; then
  echo "ERROR: Front-end build failed (127.0.0.1:3000)" >&2
  exit 1
fi

# Start Go back-end (go run compiles and runs in one step).
echo "Starting Go back-end..."
cd "$REPO_ROOT/back"
go run ./cmd/vibra/ &
BACK_PID=$!

# Start React Router front-end (build artifact already created above).
echo "Starting React Router front-end..."
cd "$REPO_ROOT/front"
pnpm start &
FRONT_PID=$!

if ! wait_for_server "back-end" "http://127.0.0.1:3001" "$BACK_PID"; then
  echo "ERROR: Back-end (127.0.0.1:3001) failed to start" >&2
  exit 1
fi

if ! wait_for_server "front-end" "http://127.0.0.1:3000" "$FRONT_PID"; then
  echo "ERROR: Front-end (127.0.0.1:3000) failed to start" >&2
  exit 1
fi

# Run Probitas E2E scenarios and capture exit code.
echo "Running E2E tests (excluding docker-tagged scenarios)..."
cd "$SCRIPT_DIR"
set +e
probitas run -s '!tag:docker' --verbose --reload 2>&1
E2E_EXIT=$?
set -e

exit $E2E_EXIT
