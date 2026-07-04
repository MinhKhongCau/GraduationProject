#!/bin/bash
set -e

# Sử dụng psql trực tiếp với block EOF để chạy chuỗi lệnh
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    CREATE DATABASE "${PROFILE_DB_NAME:-profile_db}";
    CREATE DATABASE "${PAYMENT_DB_NAME:-payment_db}";
    CREATE DATABASE "${BOOKING_DB_NAME:-booking_db}";
    CREATE DATABASE "${ASSESSMENT_DB_NAME:-assessment_db}";
EOSQL