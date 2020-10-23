#!/bin/bash
docker run \
  --rm --interactive --tty \
  --name foxtrot \
  --expose 8080 \
  --network web \
  --label 'traefik.enable=true' \
  --label 'traefik.http.routers.foxtrot.rule=Host(`foxtrot.camh.dev`)' \
  --label 'traefik.http.routers.foxtrot.entrypoints=https' \
  --label 'traefik.http.services.foxtrot.loadbalancer.server.port=8080' \
  foxtrot
