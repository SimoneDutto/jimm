#!/bin/bash
set -euo pipefail
umask 077

config="$SNAP_COMMON/config"
trap 'rm -f "$config/tls.csr"' EXIT

if [ ! -f "$SNAP_COMMON/ca.crt" ]; then
    openssl req -x509 -newkey rsa:2048 -noenc -days 365 \
        -subj '/CN=MicroJAAS local development CA' \
        -addext 'basicConstraints=critical,CA:TRUE' \
        -addext 'keyUsage=critical,keyCertSign,cRLSign' \
        -keyout "$config/ca.key" -out "$SNAP_COMMON/ca.crt"
    chmod 644 "$SNAP_COMMON/ca.crt"
fi

openssl x509 -in "$SNAP_COMMON/ca.crt" -checkend 0 -noout
openssl req -new -newkey rsa:2048 -noenc -subj '/CN=microjaas.test' \
    -addext 'subjectAltName=DNS:microjaas.test,DNS:localhost,DNS:juju-apiserver,IP:127.0.0.1,IP:::1' \
    -addext 'basicConstraints=critical,CA:FALSE' \
    -addext 'keyUsage=digitalSignature,keyEncipherment' \
    -addext 'extendedKeyUsage=serverAuth' \
    -keyout "$config/tls.key" -out "$config/tls.csr"
openssl x509 -req -in "$config/tls.csr" -copy_extensions copy -days 365 \
    -CA "$SNAP_COMMON/ca.crt" -CAkey "$config/ca.key" \
    -set_serial "0x$(openssl rand -hex 16)" -out "$config/tls.crt"
cat "$SNAP_COMMON/ca.crt" >> "$config/tls.crt"
