#!/bin/bash

set -euo pipefail

source /etc/profile.d/go.sh

echo "--- Regenerating the Kibana client"
make -C generated/kbapi all

echo "--- Building the provider"
make build

echo "--- Starting Stack containers"
make docker-fleet
docker ps 
sleep 30

echo "--- Collecting docker info"
docker ps 
docker logs terraform-elasticstack-kb 2>&1 > kibana.log
docker logs terraform-elasticstack-es 2>&1 > es.log
docker logs terraform-elasticstack-fleet 2>&1 > fleet.log

buildkite-agent artifact upload kibana.log
buildkite-agent artifact upload es.log
buildkite-agent artifact upload fleet.log

echo "--- Running acceptance tests"
ELASTICSEARCH_ENDPOINTS=http://localhost:9200 KIBANA_ENDPOINT=http://localhost:5601 ELASTICSEARCH_USERNAME=elastic ELASTICSEARCH_PASSWORD=password make testacc
