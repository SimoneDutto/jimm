#!/bin/bash
set -euo pipefail
umask 077
trap 'echo "Bootstrap failed at line $LINENO; fix the error and rerun microjaas bootstrap." >&2' ERR

exec 9>"$SNAP_COMMON/setup.lock"
flock -n 9

source "$SNAP/setup/postgres-env.sh"

/usr/bin/snapctl stop microjaas
echo "Creating a fresh deployment; previous databases are left untouched."
if [ -f "$SNAP_COMMON/config/state.json" ]; then
    previous=$(jq -er .uuid "$SNAP_COMMON/config/state.json")
    mkdir -p "$SNAP_COMMON/history/$previous"
    cp -a "$SNAP_COMMON/config/." "$SNAP_COMMON/history/$previous/"
fi
bash "$SNAP/setup/configure.sh" prepare
bash "$SNAP/setup/certificates.sh"
bash "$SNAP/setup/postgresql.sh"
bash "$SNAP/setup/openfga.sh"
bash "$SNAP/setup/configure.sh" render

/usr/bin/snapctl start --enable microjaas.dex
curl --fail --silent --show-error --noproxy '*' --cacert "$SNAP_COMMON/ca.crt" \
    --retry 30 --retry-connrefused --retry-delay 1 --retry-max-time 90 --max-time 5 \
    https://microjaas.test:18556/dex/.well-known/openid-configuration > /dev/null

/usr/bin/snapctl start --enable microjaas.jimm
curl --fail --silent --show-error --noproxy '*' --cacert "$SNAP_COMMON/ca.crt" \
    --retry 30 --retry-connrefused --retry-delay 1 --retry-max-time 90 --max-time 5 \
    https://microjaas.test:18443/.well-known/jwks.json > /dev/null

/usr/bin/snapctl start --enable microjaas.dashboard
curl --fail --silent --show-error --noproxy '*' --cacert "$SNAP_COMMON/ca.crt" \
    --retry 30 --retry-connrefused --retry-delay 1 --retry-max-time 90 --max-time 5 \
    https://microjaas.test:18600/ > /dev/null

bash "$SNAP/runtime/microjaas.sh" credentials
