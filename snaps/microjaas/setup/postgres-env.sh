#!/bin/bash
export PGHOST PGPORT PGUSER PGPASSWORD
PGHOST=$(/usr/bin/snapctl get postgres.host)
PGPORT=$(/usr/bin/snapctl get postgres.port)
PGUSER=$(/usr/bin/snapctl get postgres.user)
PGPASSWORD=$(/usr/bin/snapctl get postgres.password)
: "${PGHOST:?Set postgres.host with snap set microjaas}"
: "${PGPORT:?Set postgres.port with snap set microjaas}"
: "${PGUSER:?Set postgres.user with snap set microjaas}"
: "${PGPASSWORD:?Set postgres.password with snap set microjaas}"
export PGDATABASE=postgres PGSSLMODE=disable PGCONNECT_TIMEOUT=10
