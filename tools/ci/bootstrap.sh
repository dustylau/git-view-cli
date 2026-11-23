#!/bin/sh
set -eu

root=$(CDPATH='' cd -- "$(dirname -- "$0")/../.." && pwd)
exec docker compose --project-directory "$root" -f "$root/compose.yaml" run --build --rm bootstrap "$@"
