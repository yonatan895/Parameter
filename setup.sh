
#!/usr/bin/env bash
set -euo pipefail

# Ensure user is in docker group
USER_NAME=${USER:-$(whoami)}
if ! groups $USER_NAME | grep -q '\bdocker\b'; then
  echo "Adding $USER_NAME to docker group. Please log out and back in, then re-run this script."
  sudo usermod -aG docker $USER_NAME && newgrp docker
fi

# Ensure Go modules are downloaded
(cd backend && go mod tidy && cd ..)

# Start minikube with docker driver
minikube start --driver=docker

# Use minikube docker daemon
eval $(minikube docker-env --shell bash)

# Build images
docker build -t backend:latest ./backend
docker build -t frontend:latest ./frontend

# Deploy services via Helm
helm upgrade --install twitter-clone ./helm-chart

# Wait for postgres pod
kubectl wait --for=condition=ready pod -l app=postgres --timeout=120s

# Load database schema
POSTGRES_POD=$(kubectl get pods -l app=postgres -o jsonpath='{.items[0].metadata.name}')
kubectl cp backend/schema.sql "$POSTGRES_POD":/schema.sql
kubectl exec "$POSTGRES_POD" -- psql -U user -d twitter -f /schema.sql
