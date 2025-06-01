# Twitter Clone

This repository contains a simple twitter-like clone demonstrating a full stack application with Go backend and TypeScript frontend. The project is designed to run locally on Kubernetes using Helm charts and ArgoCD.

## Features
- User registration and login
- Post messages and view a personal feed
- Artificial traffic generator on the backend
- Uses Postgres, Redis, Kafka and Minio

## Requirements
- [Go](https://golang.org/) 1.20+
- [Node.js](https://nodejs.org/) 18+
- [Docker](https://www.docker.com/) and [Docker Compose](https://docs.docker.com/compose/) (for building images)
- [kubectl](https://kubernetes.io/docs/tasks/tools/) and [minikube](https://minikube.sigs.k8s.io/docs/) or any Kubernetes cluster
- [Helm](https://helm.sh/)
- [ArgoCD](https://argo-cd.readthedocs.io/)


## Quickstart
Run the provided script to build the Docker images, start Minikube with the
Docker driver and deploy everything:

```bash
./setup.sh
```
The script will ensure you are in the `docker` group, generate `go.sum` with
`go mod tidy` if needed, build the images and load the database schema.


## Backend
The backend lives in `backend/` and exposes a small REST API using Gin. Configuration is done via environment variables. The schema is defined in `backend/schema.sql`.

### Build image
```bash
cd backend
go mod tidy # ensure dependencies
docker build -t backend:latest .
```

## Frontend
The frontend is a minimal React + TypeScript application found in `frontend/`.

### Build image
```bash
cd frontend
npm install # fetch packages
npm run build
docker build -t frontend:latest .
```

## Running locally with Minikube
The easiest way to run everything is with the `setup.sh` script described above.
The steps below outline what the script performs manually.
1. Start minikube:
   ```bash
   minikube start
   ```
2. Load images into the cluster (or push them to a registry accessible by the cluster):
   ```bash
   eval $(minikube docker-env)

   (cd backend && go mod tidy)
   docker build -t backend:latest ./backend
   docker build -t frontend:latest ./frontend
   ```
3. Deploy the stack using Helm:
   ```bash
   helm install twitter-clone ./helm-chart
   ```
4. Access the frontend:
   ```bash
   minikube service frontend --url
   ```

## Using ArgoCD
1. Install ArgoCD in your cluster (see the [official docs](https://argo-cd.readthedocs.io/)).
2. Apply the ArgoCD application manifest:
   ```bash
   kubectl apply -f helm-chart/argocd-app.yaml
   ```
   ArgoCD will then deploy the chart and keep it in sync with the repository.

## Database setup
After Postgres is running you can create the tables using the provided schema:
```bash
kubectl exec -it deployment/postgres -- psql -U user -d twitter -f /schema.sql
```
Adjust credentials if you changed them in `values.yaml`.

## Notes
This project is intentionally simple and aims to provide a starting point. Feel free to extend authentication, add more APIs, or integrate Kafka consumers and producers for real-time updates.
