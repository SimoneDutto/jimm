#!/bin/bash

# This script verifies SSH access to a machine through JIMM's jump server.

set -euo pipefail

JIMM_CONTROLLER_NAME="${JIMM_CONTROLLER_NAME:-jimm-dev}"
BACKING_CONTROLLER_NAME="${BACKING_CONTROLLER_NAME:-qa-lxd}"
key_path="$(mktemp -u "${TMPDIR:-/tmp}/jimm-ssh-test-key.XXXXXX")"
model_name="ssh-test-$RANDOM"

# Source the `JAAS` variable for executing jaas commands.
source "local/jimm/detect-jaas.sh"

cleanup() {
	rm -f "$key_path" "$key_path.pub"
	juju destroy-model "$model_name" --force --no-prompt --destroy-all-models 2>/dev/null || true
}
trap cleanup EXIT

# Tests run sequentially and share Juju's current controller/model. Start on
# JIMM's controller, then create and select this test's model before executing
# model-scoped commands.
juju switch "$JIMM_CONTROLLER_NAME"
$JAAS add-model "$model_name" localhost --target-controller "$BACKING_CONTROLLER_NAME"

ssh-keygen -q -t rsa -N "" -f "$key_path"
juju add-ssh-key "$(cat "$key_path.pub")"
juju add-machine
until [ "$(juju status 0 --format json | jq -r '.machines["0"]["juju-status"].current')" = "started" ]; do
	sleep 5
done
juju ssh 0 true
