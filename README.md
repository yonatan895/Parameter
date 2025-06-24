# Twitter Clone

This repository contains a minimal Twitter-like application consisting of a Go backend and a React frontend. It is intended as a learning project demonstrating how to wire common infrastructure components together and deploy them to Kubernetes.

## Features

- User registration and login
- Post messages and view a personal feed
- Simple traffic generator that inserts random posts
- Deployment via Helm with Postgres, Redis, Kafka (KRaft) and Minio

## Project layout

```
backend/   Go API server and Dockerfile
frontend/  React client and Dockerfile
helm-chart/ Kubernetes manifests packaged as a chart
.github/    CI workflow definition
scripts/    Helper scripts including `setup.sh`
```

A more detailed explanation of the architecture is available in [docs/architecture.md](docs/architecture.md).

## Requirements

- [Go](https://golang.org/) 1.20+
- [Node.js](https://nodejs.org/) 18+
- [Docker](https://www.docker.com/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/) and [minikube](https://minikube.sigs.k8s.io/) or another Kubernetes cluster
- [Helm](https://helm.sh/)
- [ArgoCD](https://argo-cd.readthedocs.io/en/stable/)

## Quickstart

Run the provided script to build the Docker images, start Minikube and deploy everything:

```bash
chmod +x setup.sh
./setup.sh
```

Once the services are running, access the frontend via the URL printed by `minikube service frontend --url`.

## Configuration

The backend reads several environment variables:

- `DATABASE_URL` – Postgres connection string
- `REDIS_ADDR` – Redis address (default `localhost:6379`)
- `KAFKA_ADDR` – Kafka broker address (default `localhost:9092`)

These can be customized in `helm-chart/values.yaml` when deploying to Kubernetes.

## Testing

Backend tests and linters:

```bash
cd backend
go vet ./...
go test ./...
```

Frontend linting and build checks:

```bash
cd frontend
npm ci
npm run lint
```

## Continuous Integration

Pull requests trigger the [GitHub Actions workflow](.github/workflows/ci.yml). The pipeline caches Go modules and Docker layers, runs `golangci-lint`, executes Go tests with coverage, and builds backend and frontend Docker images. Details can be found in [docs/ci.md](docs/ci.md).

## Running manually with Minikube

If you prefer to run the commands manually instead of using `setup.sh`:

1. Start minikube: `minikube start`
2. Use the minikube Docker daemon: `eval $(minikube docker-env)`
3. Build images:
   ```bash
   docker build -t backend:latest ./backend
   docker build -t frontend:latest ./frontend
   ```
4. Deploy the chart:
   ```bash
   helm install twitter-clone ./helm-chart
   ```
5. Load the database schema using `kubectl exec` as shown in `setup.sh`.



