# Continuous Integration

The project uses GitHub Actions to lint, test and build Docker images for both the
backend and the frontend. The workflow definition can be found in
`.github/workflows/ci.yml`.

## Backend job

1. **Checkout** sources and set up Go.
2. **Cache** the Go build and module directories for faster runs.
3. **Download** dependencies and run `golangci-lint`.
4. **Execute** the unit tests with coverage reporting.
5. **Build** a Docker image using Buildx. The builder stage is executed first to
   run tests inside the container and the final image is built afterwards.

## Frontend job

1. Waits for the backend job to finish.
2. **Checkout** sources and set up Node.js.
3. **Cache** the npm directory.
4. **Install** dependencies and run ESLint and any frontend tests.
5. **Build** static assets and a Docker image using Buildx.

Caching of Buildx layers and Go modules ensures subsequent CI runs are much
faster. The resulting images are suitable for running integration tests or
deployment in Kubernetes.
