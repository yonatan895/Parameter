# Architecture Overview

This project is a minimal Twitter-like clone composed of a Go backend and a React frontend.
It demonstrates how to combine several popular infrastructure components in a Kubernetes
cluster.

## Components

- **Backend (`backend/`)** – REST API built with [Gin](https://github.com/gin-gonic/gin).
  It stores data in Postgres, caches messages in Redis and publishes events to Kafka.
- **Frontend (`frontend/`)** – Simple React application bundled with webpack.
- **Database** – Postgres is used as the primary data store.
- **Cache** – Redis is used to cache posted messages.
- **Message queue** – Kafka runs in KRaft mode for simplicity and receives message events.
- **Object storage** – Minio is included in the Helm chart but not used by the demo code.
- **Kubernetes** – A Helm chart under `helm-chart/` deploys all services. An optional
  ArgoCD application manifest keeps the cluster in sync with the repository.

## Data flow

1. Users register and log in via the backend API.
2. Posting a message writes the record to Postgres, caches the content in Redis and
   publishes the event to Kafka.
3. The frontend polls the `/feed` endpoint to display messages for the authenticated user.
4. A small traffic generator inserts random messages periodically so the feed is never empty.

The `setup.sh` script provisions a local Minikube cluster, builds Docker images and
installs the Helm chart to get everything running with minimal manual work.
