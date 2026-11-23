#!/bin/sh
set -eu
umask 077

gitea_cli() {
    su-exec git gitea --config /data/gitea/conf/app.ini "$@"
}

if [ ! -s /credentials/password ]; then
    od -An -N24 -tx1 /dev/urandom | tr -d ' \n' >/credentials/password.tmp
    mv /credentials/password.tmp /credentials/password
fi

password=$(cat /credentials/password)
auth=$(printf 'local-ci:%s' "$password" | base64 -w0)
if ! wget -q -O /dev/null --header="Authorization: Basic $auth" http://git:3000/api/v1/user 2>/dev/null; then
    gitea_cli admin user create --admin --username local-ci --password "$password" \
        --email local-ci@example.invalid --must-change-password=false
fi

if [ ! -s /runner-token/token ]; then
    token=$(gitea_cli actions generate-runner-token)
    test -n "$token"
    printf '%s\n' "$token" >/runner-token/token.tmp
    mv /runner-token/token.tmp /runner-token/token
fi
