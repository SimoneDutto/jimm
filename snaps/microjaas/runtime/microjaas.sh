#!/bin/bash
set -euo pipefail

case "${1:-}" in
    bootstrap)
        exec bash "$SNAP/setup/bootstrap.sh"
        ;;
    status)
        exec /usr/bin/snapctl services
        ;;
    credentials)
        echo 'DEVELOPMENT ONLY: credentials and private keys are bundled in this snap.'
        echo 'Open: https://microjaas.test:18600'
        echo 'Login: admin@microjaas.local'
        jq -r '"Password: " + .admin_password' "$SNAP_COMMON/config/state.json"
        jq -er '.staticClients[] | select(.id == "automation") |
            "Service account client ID: " + .id,
            "Service account secret: " + .secret' "$SNAP_COMMON/config/dex.json"
        echo 'Juju: juju login microjaas.test:18443 -c microjaas --trust --no-prompt'
        ;;
    certificate)
        exec cat "$SNAP_COMMON/ca.crt"
        ;;
    *)
        echo 'Usage: microjaas bootstrap|status|credentials|certificate' >&2
        exit 1
        ;;
esac
