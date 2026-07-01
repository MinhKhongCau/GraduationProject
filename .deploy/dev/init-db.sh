#!/bin/bash
# Runs once when the postgres-db container first initializes its data volume.
# POSTGRES_DB is already created by the base image; this creates the
# remaining per-service databases so each microservice gets its own schema.
set -e

for db in "$PROFILE_DB_NAME" "$PAYMENT_DB_NAME" "$BOOKING_DB_NAME" "$ASSESSMENT_DB_NAME"; do
  exists=$(psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" -tAc "SELECT 1 FROM pg_database WHERE datname = '$db'")
  if [ "$exists" != "1" ]; then
    psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" -c "CREATE DATABASE \"$db\""
  fi
done
