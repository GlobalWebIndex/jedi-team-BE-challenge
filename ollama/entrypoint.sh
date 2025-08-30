#!/usr/bin/env sh
set -eu

# Start the Ollama server
ollama serve &
SERVE_PID=$!

# Wait until the API is responsive
until curl -sf http://localhost:11434/api/tags >/dev/null 2>&1; do
  sleep 0.5
done

# Ensure granite3-dense:8b is present (pull if missing)
if ! ollama list | grep -qE '^granite3-dense:8b(\s|$)'; then
  echo "Pulling model granite3-dense:8b…"
  # If the pull fails due to transient network errors, don't crash the container
  ollama pull granite3-dense:8b || echo "Warning: initial pull failed; model can still be pulled later."
fi

# Keep the server in the foreground
wait "${SERVE_PID}"
