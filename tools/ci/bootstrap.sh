#!/bin/sh
set -eu

root=$(CDPATH='' cd -- "$(dirname -- "$0")/../.." && pwd)
docker compose --project-directory "$root" -f "$root/compose.yaml" up -d --no-recreate runner
exec docker compose --project-directory "$root" -f "$root/compose.yaml" run --no-deps --build --rm bootstrap "$@"
