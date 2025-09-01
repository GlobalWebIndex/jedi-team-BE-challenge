#!/usr/bin/env sh
set -eu

MODEL_NAME="${OLLAMA_MODEL:-granite3-dense:8b}"

ollama serve &
SERVE_PID=$!

cleanup() {
  echo "Shutting down Ollama server..."
  kill $SERVE_PID 2>/dev/null || true
  wait $SERVE_PID 2>/dev/null || true
}
trap cleanup EXIT INT TERM

until curl -sf http://localhost:11434/api/tags >/dev/null 2>&1; do
  sleep 0.5
done

if ! ollama list | grep -qE "^${MODEL_NAME}(\s|$)"; then
  echo "Pulling model ${MODEL_NAME}…"
  ollama pull "${MODEL_NAME}" || echo "Warning: initial pull failed; model can still be pulled later."
fi

wait "${SERVE_PID}"