#!/bin/bash
set -e

exec /bin/main \
 --migrations=$MIGRATIONS_DIR \
 --x-migrations-table=$MIGRATIONS_TABLE_NAME
