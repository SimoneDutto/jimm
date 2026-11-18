# MicroJAAS PoC

## Build and bootstrap

Install Go, curl, tar, unzip and Snapcraft on Linux amd64/arm64; initialize LXD.
Supply the Git-ignored development fixture before building:
`dist/microjaas/dev-fixture/config/{state.json,signing.pem,jwks.json}`.
Do not commit these files or use real data.

Run from the repository root. PostgreSQL must have the `microjaas_provisioner`
account; use its password below. Add `192.168.1.14 microjaas.test` to `/etc/hosts`.

Build and install (use `arm64` instead of `amd64` where appropriate):

```sh
make microjaas-snap
sudo snap install --dangerous dist/microjaas_poc_amd64.snap
sudo snap set microjaas postgres.host=127.0.0.1 postgres.port=5432 postgres.user=microjaas_provisioner postgres.password=microjaas-development-only
sudo microjaas bootstrap
microjaas certificate | sudo tee /usr/local/share/ca-certificates/microjaas-local.crt >/dev/null
sudo update-ca-certificates
```

Bootstrap creates fresh databases and a new JIMM identity. Do not rerun it to update.

## Login and operations

To update the snap without creating a fresh deployment:

```sh
make microjaas-snap
sudo snap install --dangerous dist/microjaas_poc_amd64.snap
sudo snap start --enable microjaas.dashboard
sudo snap restart microjaas
```

Open <https://microjaas.test:18600>.
If your browser uses a separate trust store, import `microjaas certificate` there too.

```sh
sudo microjaas credentials             # human and service-account credentials
juju login microjaas.test:18443 -c microjaas --trust --no-prompt
sudo microjaas status
sudo snap logs microjaas -f
sudo snap restart microjaas            # preserves current data
sudo snap stop microjaas               # leaves PostgreSQL running
sudo snap start microjaas
```

To grant the `automation` service account administrator access, run as the human admin:

```sh
SERVICE_ACCOUNT_ID=automation bash local/jimm/setup-service-account.sh
```

Allow TCP ports 18443 and 18556 through the firewall for controllers.
Allow TCP port 18600 for dashboard access from other machines.

## Demo

See the [command walkthrough](../../local/microjaas-demo/README.md) for LXD
controller bootstrap and model creation.
