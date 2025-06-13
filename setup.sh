#!/usr/bin/env bash
set -euo pipefail

# --- CONFIGURABLE VARIABLES ---
BACKEND_IMAGE_TAG=${BACKEND_IMAGE_TAG:-latest}
FRONTEND_IMAGE_TAG=${FRONTEND_IMAGE_TAG:-latest}

# --- TOOL CHECKS ---
for tool in minikube kubectl docker go; do
  if ! command -v $tool &>/dev/null; then
    echo "[ERROR] Required tool '$tool' is not installed. Please install it and re-run the script."
    exit 1
  fi
done

# --- CLEANUP OPTION ---
if [[ "${1:-}" == "cleanup" ]]; then
  echo "[INFO] Cleaning up Kubernetes resources and minikube..."
  kubectl delete -f helm-chart/argocd-app.yaml || true
  kubectl delete namespace argocd || true
  minikube delete || true
  echo "[INFO] Cleanup complete."
  exit 0
fi

# --- DOCKER GROUP CHECK ---
USER_NAME=${USER:-$(whoami)}
if ! groups $USER_NAME | grep -q '\bdocker\b'; then
  echo "[INFO] Adding $USER_NAME to docker group. Please log out and back in, then re-run this script."
  sudo usermod -aG docker $USER_NAME && newgrp docker
fi

# --- GO MODULES ---
echo "[INFO] Downloading Go modules..."
(cd backend && go mod tidy && cd ..)

# --- MINIKUBE START ---
echo "[INFO] Starting minikube with docker driver..."
minikube start --driver=docker

echo "[INFO] Using minikube docker daemon..."
eval $(minikube docker-env --shell bash)

# --- BUILD IMAGES ---
echo "[INFO] Building backend image (tag: $BACKEND_IMAGE_TAG)..."
docker build -t backend:$BACKEND_IMAGE_TAG ./backend

echo "[INFO] Building frontend image (tag: $FRONTEND_IMAGE_TAG)..."
docker build -t frontend:$FRONTEND_IMAGE_TAG ./frontend

# --- ARGO CD INSTALL ---
echo "[INFO] Creating argocd namespace (if not exists)..."
kubectl create namespace argocd || true

echo "[INFO] Installing Argo CD..."
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml

# --- WAIT FOR ARGO CD PODS TO APPEAR ---
echo "[INFO] Waiting for Argo CD pods to be created..."
for i in {1..30}; do
  if kubectl get pods -n argocd | grep -q argocd; then
    break
  fi
  sleep 5
done

# --- WAIT FOR ARGO CD TO BE READY ---
echo "[INFO] Waiting for Argo CD server pod to be ready..."
kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=argocd-server -n argocd --timeout=180s
 

# --- EXPOSE ARGO CD UI LOCALLY ---
echo "[INFO] Forwarding Argo CD UI to http://localhost:8080 ... (backgrounded)"
kubectl port-forward svc/argocd-server -n argocd 8080:80 &

# --- PRINT ARGO CD ADMIN PASSWORD ---
echo "[INFO] Fetching Argo CD admin password..."
kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d; echo

# --- APPLY ARGO CD APPLICATION MANIFEST ---
echo "[INFO] Apply Argo CD application manifest..."
kubectl apply -f helm-chart/argocd-app.yaml

# --- WAIT FOR ARGO CD APPLICATION TO BE SYNCED AND HEALTHY ---
echo "[INFO] Waiting for Argo CD application to be Synced and Healthy..."
for i in {1..30}; do
  STATUS=$(kubectl get application twitter-clone -n argocd -o jsonpath='{.status.sync.status}' 2>/dev/null || echo "")
  HEALTH=$(kubectl get application twitter-clone -n argocd -o jsonpath='{.status.health.status}' 2>/dev/null || echo "")
  if [[ "$STATUS" == "Synced" && "$HEALTH" == "Healthy" ]]; then
    echo "[INFO] Argo CD application is Synced and Healthy."
    break
  fi
  sleep 5
done

# Check if app is healthy, else print status and exit
STATUS=$(kubectl get application twitter-clone -n argocd -o jsonpath='{.status.sync.status}' 2>/dev/null || echo "")
HEALTH=$(kubectl get application twitter-clone -n argocd -o jsonpath='{.status.health.status}' 2>/dev/null || echo "")
if [[ "$STATUS" != "Synced" || "$HEALTH" != "Healthy" ]]; then
  echo "[ERROR] Argo CD application did not become Synced and Healthy."
  echo "[ERROR] Status: $STATUS, Health: $HEALTH"
  echo "[ERROR] Application conditions:"
  kubectl get application twitter-clone -n argocd -o jsonpath='{.status.conditions}'
  exit 1
fi

# --- WAIT FOR POSTGRES POD TO APPEAR ---
echo "[INFO] Waiting for postgres pod to be created..."
for i in {1..30}; do
  if kubectl get pods -l app=postgres | grep -q postgres; then
    break
  fi
  sleep 5
done

# --- WAIT FOR POSTGRES POD TO BE READY ---
echo "[INFO] Waiting for postgres pod to be ready..."
kubectl wait --for=condition=ready pod -l app=postgres --timeout=120s

# --- LOAD DATABASE SCHEMA ---
POSTGRES_POD=$(kubectl get pods -l app=postgres -o jsonpath='{.items[0].metadata.name}')
if [[ -z "$POSTGRES_POD" ]]; then
  echo "[ERROR] Could not find a running Postgres pod."
  exit 1
fi

echo "[INFO] Copying schema.sql to Postgres pod..."
kubectl cp backend/schema.sql "$POSTGRES_POD":/schema.sql

echo "[INFO] Loading schema into Postgres..."
kubectl exec "$POSTGRES_POD" -- psql -U user -d twitter -f /schema.sql

echo "[INFO] Setup complete!"
