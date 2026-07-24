#!/usr/bin/env sh
set -eu

if [ "$#" -ne 1 ]; then
  printf 'Usage: %s <image-tag>\n' "$0" >&2
  exit 2
fi

export IMAGE_TAG="$1"
docker compose pull todoapp
docker compose up -d --no-deps todoapp
docker compose ps todoapp
