#!/bin/bash
set -euo pipefail
umask 077
config="$SNAP_COMMON/config"

if [ "$1" = prepare ]; then
    mkdir -p "$config"
    id=$(openssl rand -hex 16)
    uuid="${id:0:8}-${id:8:4}-4${id:13:3}-a${id:17:3}-${id:20:12}"
    jq --arg uuid "$uuid" '.uuid = $uuid | .store_id = "" | .model_id = "" | del(.model_hash)' \
        "$SNAP/dev-fixture/state.json" > "$config/state.json"
    cp "$SNAP/dev-fixture/signing.pem" "$SNAP/dev-fixture/jwks.json" "$config/"
    cp "$SNAP/setup/authorisation-model.json" "$config/"
fi

jq -r --arg host "$PGHOST" --arg port "$PGPORT" '
    (.uuid | gsub("-"; "")) as $id |
    . + {
        jimm_dsn: ("postgres://microjaas_jimm_" + $id + ":" + (.database_passwords.jimm | @uri)
            + "@" + (if $host | contains(":") then "[" + $host + "]" else $host end)
            + ":" + $port + "/microjaas_jimm_" + $id + "?sslmode=disable"),
        openfga_dsn: ("postgres://microjaas_openfga_" + $id + ":" + (.database_passwords.openfga | @uri)
            + "@" + (if $host | contains(":") then "[" + $host + "]" else $host end)
            + ":" + $port + "/microjaas_openfga_" + $id + "?sslmode=disable")
    } | to_entries[] | select(.value | type == "string") |
    (.key | ascii_upcase) + "=" + (.value | @sh)
' "$config/state.json" > "$config/credentials.env"
jq --arg common "$SNAP_COMMON" --arg root "${SNAP%/*}/current" --arg host "$PGHOST" --argjson port "$PGPORT" \
    -f "$SNAP/templates/dex.json.jq" "$config/state.json" > "$config/dex.json"
