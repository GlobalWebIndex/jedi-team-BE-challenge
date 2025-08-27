FROM ollama/ollama:latest

# Copy our entrypoint script
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/bin/sh", "-c", "/entrypoint.sh"]
