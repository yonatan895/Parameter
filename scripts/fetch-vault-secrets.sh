#!/usr/bin/env bash
set -euo pipefail

OUT_FILE="helm-chart/values.secrets.yaml"

POSTGRES_USER=$(vault kv get -field=user secret/twitter/postgres)
POSTGRES_PASSWORD=$(vault kv get -field=password secret/twitter/postgres)
MINIO_ACCESS=$(vault kv get -field=accessKey secret/twitter/minio)
MINIO_SECRET=$(vault kv get -field=secretKey secret/twitter/minio)

cat > "$OUT_FILE" <<EOF2
postgres:
  user: "$POSTGRES_USER"
  password: "$POSTGRES_PASSWORD"
minio:
  accessKey: "$MINIO_ACCESS"
  secretKey: "$MINIO_SECRET"
EOF2
echo "Secrets written to $OUT_FILE"
