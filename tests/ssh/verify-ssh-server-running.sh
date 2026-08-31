#!/bin/bash

# This script verifies SSH access to a machine through JIMM's jump server.

set -euo pipefail

key_path="$(mktemp -u "${TMPDIR:-/tmp}/jimm-ssh-test-key.XXXXXX")"
model_name="ssh-test-$RANDOM"

cleanup() {
	rm -f "$key_path" "$key_path.pub"
	juju destroy-model "$model_name" --force --no-prompt --destroy-all-models 2>/dev/null || true
}
trap cleanup EXIT

ssh-keygen -q -t rsa -N "" -f "$key_path"
juju add-ssh-key "$(cat "$key_path.pub")"

juju add-model "$model_name"
juju add-machine
until [ "$(juju status 0 --format json | jq -r '.machines["0"]["juju-status"].current')" = "started" ]; do
	sleep 5
done
juju ssh --jump 0 true
