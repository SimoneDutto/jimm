#!/bin/bash
set -euo pipefail

state="$SNAP_COMMON/config/state.json"
generation=$(jq -er '.uuid | gsub("-"; "")' "$state")
for service in jimm openfga dex; do
    MICROJAAS_DB_NAME="microjaas_${service}_$generation"
    MICROJAAS_DB_PASSWORD=$(jq -er --arg service "$service" '.database_passwords[$service]' "$state")
    export MICROJAAS_DB_NAME MICROJAAS_DB_PASSWORD
    "$SNAP/usr/lib/postgresql/16/bin/psql" -X -q -w -v ON_ERROR_STOP=1 \
        < "$SNAP/setup/postgresql.sql"
done
