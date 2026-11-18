#!/bin/bash
set -euo pipefail

cd "$(dirname "$0")/../../.."
root=$PWD
work="$root/dist/microjaas"
project="$work/package"
stage="$project/payload"
arch=$(go env GOARCH)
case "$arch" in
    amd64|arm64) ;;
    *) echo "Only Linux amd64 and arm64 are supported" >&2; exit 1 ;;
esac
if [ "$(go env GOOS)" != linux ]; then
    echo "Build this PoC on Linux" >&2
    exit 1
fi

dex_version=v2.46.1
openfga_version=1.10.0
fga_version=0.9.0
dashboard_version=0.16.0
fixture="$work/dev-fixture/config"
for file in state.json signing.pem jwks.json; do
    if [ ! -f "$fixture/$file" ]; then
        echo "Missing local development fixture: $fixture/$file (see snaps/microjaas/README.md)" >&2
        exit 1
    fi
done
mkdir -p "$stage/bin" "$stage/share/dex" "$stage/licenses" "$stage/setup" "$stage/runtime" "$stage/templates" "$stage/meta/hooks" "$work/downloads"

echo "Building JIMM (no tests)..."
CGO_ENABLED=0 go build -mod=readonly -trimpath -ldflags='-s -w' -o "$stage/bin/" ./cmd/jimmsrv
rm -f "$stage/bin/microjaas" "$stage/templates/dex.json"
mkdir -p "$stage/dev-fixture"
install -m 600 "$fixture/state.json" "$fixture/signing.pem" "$fixture/jwks.json" "$stage/dev-fixture/"

echo "Building Dex $dex_version..."
if [ ! -f "$work/downloads/dex-$dex_version.tar.gz" ]; then
    curl -fL --retry 3 "https://github.com/dexidp/dex/archive/refs/tags/$dex_version.tar.gz" \
        -o "$work/downloads/dex-$dex_version.tar.gz.tmp"
    mv "$work/downloads/dex-$dex_version.tar.gz.tmp" "$work/downloads/dex-$dex_version.tar.gz"
fi
mkdir -p "$work/dex-$dex_version"
tar -xzf "$work/downloads/dex-$dex_version.tar.gz" -C "$work/dex-$dex_version" --strip-components=1
(
    cd "$work/dex-$dex_version"
    CGO_ENABLED=0 go build -mod=readonly -trimpath -ldflags="-s -w -X main.version=$dex_version" -o "$stage/bin/dex" ./cmd/dex
)
cp -R "$work/dex-$dex_version/web" "$stage/share/dex/"
cp "$work/dex-$dex_version/LICENSE" "$stage/licenses/dex-LICENSE"
cp LICENSE "$stage/licenses/jimm-LICENSE"

echo "Fetching Juju Dashboard $dashboard_version..."
asset="juju-dashboard-v$dashboard_version.charm"
if [ ! -f "$work/downloads/$asset" ]; then
    curl -fL --retry 3 "https://github.com/canonical/juju-dashboard/releases/download/v$dashboard_version/juju-dashboard_ubuntu%4024.04-amd64.charm" \
        -o "$work/downloads/$asset.tmp"
    mv "$work/downloads/$asset.tmp" "$work/downloads/$asset"
fi
echo "716fc1903c820b17fd3c387c8bb7693ef9914ba1ef8bb0e577e1735d22527135  $work/downloads/$asset" | sha256sum --check -
mkdir -p "$work/dashboard" "$stage/share/dashboard"
unzip -oq "$work/downloads/$asset" 'src/dist/*' -d "$work/dashboard"
cp -R "$work/dashboard/src/dist/." "$stage/share/dashboard/"
cp snaps/microjaas/templates/dashboard-config.js "$stage/share/dashboard/config.js"
curl -fL --retry 3 "https://raw.githubusercontent.com/canonical/juju-dashboard/v$dashboard_version/LICENSE.md" \
    -o "$stage/licenses/juju-dashboard-LICENSE"

echo "Fetching OpenFGA $openfga_version..."
asset="openfga_${openfga_version}_linux_${arch}.tar.gz"
release="https://github.com/openfga/openfga/releases/download/v$openfga_version"
if [ ! -f "$work/downloads/$asset" ]; then
    curl -fL --retry 3 "$release/$asset" -o "$work/downloads/$asset.tmp"
    mv "$work/downloads/$asset.tmp" "$work/downloads/$asset"
fi
curl -fL --retry 3 "$release/checksums.txt" -o "$work/downloads/openfga-checksums.txt"
(
    cd "$work/downloads"
    grep " $asset\$" openfga-checksums.txt | sha256sum --check -
)
tar -xzf "$work/downloads/$asset" -C "$stage/bin" openfga
curl -fL --retry 3 "https://raw.githubusercontent.com/openfga/openfga/v$openfga_version/LICENSE" \
    -o "$stage/licenses/openfga-LICENSE"
rm -f "$stage/bin/load-env" "$stage/setup/microjaas.sh" "$stage/setup/trust-ca.sh" "$stage/runtime/trust-ca.sh"
install -m 755 snaps/microjaas/runtime/* "$stage/runtime/"
install -m 755 snaps/microjaas/setup/{bootstrap,configure,postgresql,postgres-env,openfga,certificates}.sh "$stage/setup/"
cp snaps/microjaas/setup/postgresql.sql "$stage/setup/"
install -m 755 snaps/microjaas/setup/hooks/configure "$stage/meta/hooks/configure"
cp snaps/microjaas/templates/* "$stage/templates/"
chmod 755 "$stage/bin/"*

echo "Converting the authorization model with FGA CLI $fga_version..."
asset="fga_${fga_version}_linux_${arch}.tar.gz"
release="https://github.com/openfga/cli/releases/download/v$fga_version"
if [ ! -f "$work/downloads/$asset" ]; then
    curl -fL --retry 3 "$release/$asset" -o "$work/downloads/$asset.tmp"
    mv "$work/downloads/$asset.tmp" "$work/downloads/$asset"
fi
curl -fL --retry 3 "$release/checksums.txt" -o "$work/downloads/fga-checksums.txt"
(
    cd "$work/downloads"
    grep " $asset\$" fga-checksums.txt | sha256sum --check -
)
mkdir -p "$work/fga"
tar -xzf "$work/downloads/$asset" -C "$work/fga" fga
"$work/fga/fga" model transform --file openfga/authorisation_model.fga --output-format json \
    > "$stage/setup/authorisation-model.json"

cp snaps/microjaas/snapcraft.yaml "$project/snapcraft.yaml"
(
    cd "$project"
    snapcraft pack --use-lxd
)
mv "$project/microjaas_poc_${arch}.snap" "$root/dist/"
echo "Install: sudo snap install --dangerous dist/microjaas_poc_${arch}.snap"
echo "Setup:   sudo microjaas bootstrap"
