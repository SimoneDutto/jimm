# MicroJAAS demo

Run from the repository root with PostgreSQL configured, LXD reachable at
`192.168.1.14:8443`, and `192.168.1.14 microjaas.test` in `/etc/hosts`.
The local `localhost` LXD credential and `./jaas` binary must already exist.

## Build

```sh
make microjaas-snap
```

## Install

```sh
sudo snap install --dangerous dist/microjaas_poc_amd64.snap
sudo snap set microjaas postgres.host=127.0.0.1 postgres.port=5432 postgres.user=microjaas_provisioner postgres.password=microjaas-development-only
```

## Bootstrap MicroJAAS

```sh
sudo microjaas bootstrap
```

## Trust CA

```sh
microjaas certificate | sudo tee /usr/local/share/ca-certificates/microjaas-local.crt >/dev/null
sudo update-ca-certificates
```

## Login

```sh
sudo microjaas credentials
juju login microjaas.test:18443 -c microjaas --trust --no-prompt
```

## Dashboard

```sh
xdg-open https://microjaas.test:18600
```

## Add cloud

```sh
juju add-cloud --client host-lxd local/microjaas-demo/host-lxd.yaml
```

## Copy LXD credential

```sh
(
  set -euo pipefail
  umask 077
  credentials=$(mktemp)
  trap 'rm -f "$credentials"' EXIT
  juju credentials localhost --client --show-secrets --format json |
    jq '{credentials: {"host-lxd": (.["client-credentials"].localhost["cloud-credentials"] |
      with_entries(.value = ({"auth-type": .value["auth-type"]} + .value.details)))}}' > "$credentials"
  juju update-credential --client host-lxd -f "$credentials"
)
```

## Generate cloud-init

```sh
bash local/microjaas-demo/cloudinit.sh
```

## Bootstrap controller

```sh
juju switch microjaas
./jaas bootstrap host-lxd/default my-controller 3.6.29 \
  --credential localhost --config local/microjaas-demo/cloudinit.yaml
```

## Upload credential

```sh
juju update-credential host-lxd localhost -c microjaas
```

## Create model

```sh
./jaas add-model my-model host-lxd/default \
  -c microjaas --target-controller my-controller --credential localhost
```

## Status

```sh
juju status
```
