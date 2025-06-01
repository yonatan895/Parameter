#!/usr/bin/env bash
set -euo pipefail

# Ensure current user can run docker
sudo usermod -aG docker $USER && newgrp docker

# Start minikube with docker driver
minikube start --driver=docker

# Use docker from minikube
eval $(minikube docker-env)

# Build Docker images
docker build -t backend:latest ./backend
docker build -t frontend:latest ./frontend

# Deploy services using Helm
helm upgrade --install twitter-clone ./helm-chart

# Wait for postgres pod to be ready
kubectl wait --for=condition=ready pod -l app=postgres --timeout=120s

# Load database schema
POSTGRES_POD=$(kubectl get pods -l app=postgres -o jsonpath='{.items[0].metadata.name}')
kubectl cp backend/schema.sql "$POSTGRES_POD":/schema.sql
kubectl exec "$POSTGRES_POD" -- psql -U user -d twitter -f /schema.sql

