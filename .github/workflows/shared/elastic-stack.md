---
services:
  es-proxy:
    image: backplane/socat-forward
    ports:
      - 9201:9201
    options: >-
      --add-host host.docker.internal:host-gateway
      -e LISTEN_PORT=9201
      -e DEST_PORT=9200
      -e DEST_HOST=host.docker.internal
  kb-proxy:
    image: backplane/socat-forward
    ports:
      - 5602:5602
    options: >-
      --add-host host.docker.internal:host-gateway
      -e LISTEN_PORT=5602
      -e DEST_PORT=5601
      -e DEST_HOST=host.docker.internal
network:
  allowed:
    - terraform
steps:
  - name: Setup Elastic Stack
    run: make docker-fleet
    env:
      ELASTICSEARCH_BIND: "0.0.0.0"
      KIBANA_BIND: "0.0.0.0"
  - name: Setup Kibana user
    run: make set-kibana-password
  - id: get-api-key
    name: Get ES API key
    run: |-
      echo "apikey=$(make create-es-api-key | jq -r .encoded)" >> "$GITHUB_OUTPUT"
  - id: setup-fleet
    name: Setup Fleet
    run: |-
      make setup-kibana-fleet
    env:
      FLEET_NAME: "fleet"
  - name: Docker compose logs
    if: failure()
    run: docker compose logs --no-color
---

# Elastic Stack environment setup

Starts Elasticsearch and Kibana (with Fleet) via Docker Compose, then configures
credentials and an API key. Exposes the stack to the AWF agentic sandbox through
socat proxy services so the agent can reach the stack via
`host.docker.internal:9201` (Elasticsearch) and `host.docker.internal:5602`
(Kibana). The proxies listen on AWF-allowed ports and forward to the actual
stack ports (9200/5601) on the Docker host..
