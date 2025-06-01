#!/bin/bash
set -e

# start minikube if not running
if ! minikube status >/dev/null 2>&1; then
  echo "Starting minikube..."
  minikube start
fi

# use minikube docker env
eval $(minikube docker-env)

# build backend image
pushd backend >/dev/null
if command -v go >/dev/null 2>&1; then
  go mod tidy
fi
docker build -t backend:latest .
popd >/dev/null

# build frontend image
pushd frontend >/dev/null
if command -v npm >/dev/null 2>&1; then
  npm install
  npm run build
fi
docker build -t frontend:latest .
popd >/dev/null

# deploy via helm
helm upgrade --install twitter-clone ./helm-chart

# apply database schema
kubectl rollout status deployment/postgres --timeout=120s
kubectl exec -i deployment/postgres -- psql -U user -d twitter < backend/schema.sql

echo "Deployment complete. Access the frontend with: $(minikube service frontend --url)"
