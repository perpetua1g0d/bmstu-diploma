#!/bin/sh
export POSTGRES_USER=$(cat /etc/postgres-env/POSTGRES_USER)
export POSTGRES_PASSWORD=$(cat /etc/postgres-env/POSTGRES_PASSWORD)
export POSTGRES_DB=$(cat /etc/postgres-env/POSTGRES_DB)
export POSTGRES_HOST=localhost
export POSTGRES_PORT=5432
# export POSTGRES_HOST=$(cat /etc/postgres-env/POSTGRES_HOST)
# export POSTGRES_PORT=$(cat /etc/postgres-env/POSTGRES_PORT)
./postgres-sidecar
