#!/bin/bash
set -euo pipefail
umask 077

set -a
source "$SNAP_COMMON/config/openfga.env"
set +a
"$SNAP/bin/openfga" migrate
/usr/bin/snapctl start --enable microjaas.openfga

endpoint="http://$OPENFGA_HTTP_ADDR"
curl --fail --silent --show-error --noproxy '*' \
    --retry 30 --retry-connrefused --retry-delay 1 --retry-max-time 90 --max-time 5 \
    "$endpoint/healthz" > /dev/null

state="$SNAP_COMMON/config/state.json"
model="$SNAP_COMMON/config/authorisation-model.json"
curl_args=(--fail-with-body --silent --show-error --noproxy '*' --max-time 10
    -H "Authorization: Bearer $OPENFGA_AUTHN_PRESHARED_KEYS" -H 'Content-Type: application/json')

store_id=$(curl "${curl_args[@]}" -d '{"name":"microjaas"}' "$endpoint/stores" |
    jq -er '.id // error("OpenFGA did not return a store ID")')
model_id=$(curl "${curl_args[@]}" --data-binary "@$model" "$endpoint/stores/$store_id/authorization-models" |
    jq -er '.authorization_model_id // error("OpenFGA did not return a model ID")')
jq --arg store "$store_id" --arg model "$model_id" \
    '.store_id = $store | .model_id = $model' "$state" > "$state.tmp"
mv "$state.tmp" "$state"
