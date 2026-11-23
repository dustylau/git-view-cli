#!/bin/sh
set -eu

export GIT_CONFIG_NOSYSTEM=1 GIT_CONFIG_GLOBAL=/dev/null
# A host checkout may be owned by a different UID than the container user.
export GIT_CONFIG_COUNT=1 GIT_CONFIG_KEY_0=safe.directory GIT_CONFIG_VALUE_0=/repo

git --version
go version
status=$(git-view --repo /repo status)
changes=$(git-view --repo /repo changes)
printf '%s\n' "$status" "$changes"

if [ "$1" = seeded ]; then
    printf '%s\n' "$status" | grep -Fx 'Branch: "main"' >/dev/null
    printf '%s\n' "$status" | grep -Fx 'Changes: staged 1, unstaged 1, untracked 1, conflicts 0' >/dev/null
    printf '%s\n' "$changes" | grep -Fx 'M M "tracked.txt"' >/dev/null
    printf '%s\n' "$changes" | grep -Fx '? ? "new note.txt"' >/dev/null
fi

exec go test -race -mod=readonly ./... -count=1
