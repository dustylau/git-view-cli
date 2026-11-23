#!/bin/sh
set -eu

export GIT_CONFIG_NOSYSTEM=1 GIT_CONFIG_GLOBAL=/dev/null
cd /repo
if [ -n "$(ls -A)" ]; then
    printf '%s\n' 'Refusing to seed a nonempty fixture volume.' >&2
    exit 1
fi

git init --quiet --template= --initial-branch=main
printf '%s\n' original >tracked.txt
git add -- tracked.txt
GIT_AUTHOR_DATE=2000-01-01T00:00:00Z \
    GIT_COMMITTER_DATE=2000-01-01T00:00:00Z \
    git -c user.name=Fixture -c user.email=fixture@example.invalid \
    commit --quiet -m 'Seed container fixture'

printf '%s\n' staged >tracked.txt
git add -- tracked.txt
printf '%s\n' unstaged >>tracked.txt
printf '%s\n' untracked >'new note.txt'
