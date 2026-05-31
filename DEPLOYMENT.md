# PennyPilot Deployment Guide (Homelab)

This guide explains how to deploy PennyPilot to your homelab using the infrastructure patterns found in [gregarendse/homelab](https://github.com/gregarendse/homelab).

## Prerequisites

1.  A Kubernetes cluster managed by Argo CD (as per the `homelab` repo).
2.  GitHub Container Registry (GHCR) access (or your own registry).
3.  A PostgreSQL database.

## Automated Image Builds

The `.github/workflows/ci.yml` in this repository is configured to build and push Docker images for both the backend and frontend to GHCR on every push to `master`.

Images:
- `ghcr.io/gregarendse/pennypilot-backend:latest`
- `ghcr.io/gregarendse/pennypilot-frontend:latest`

## Homelab Integration Steps

### 1. Database Setup

PennyPilot requires a PostgreSQL database. You can use the standard Bitnami PostgreSQL Helm chart or a dedicated instance.

Ensure the database is accessible from the PennyPilot backend at:
`postgres://pennypilot:pennypilot@pennypilot-db:5432/pennypilot?sslmode=disable`

### 2. Secret Management

Create a Kubernetes secret named `pennypilot-backend-secrets` in the namespace where you plan to deploy the backend. This secret should contain the following environment variables from `.env.example`:

- `ENCRYPTION_KEY_HEX`
- `MONZO_CLIENT_ID`
- `MONZO_CLIENT_SECRET`
- `TRUELAYER_CLIENT_ID`
- `TRUELAYER_CLIENT_SECRET`
- `GOCARDLESS_SECRET_ID`
- `GOCARDLESS_SECRET_KEY`

### 3. Argo CD Configuration

Add PennyPilot to your `clusters/<cluster>/apps.yaml` in the `homelab` repository:

```yaml
apps:
  - name: pennypilot-backend
    type: helm
    namespace: pennypilot
    helm:
      kind: path
      path: server
      releaseName: pennypilot-backend
      valueFiles:
        - applications/pennypilot/backend-values.yaml

  - name: pennypilot-frontend
    type: helm
    namespace: pennypilot
    helm:
      kind: path
      path: server
      releaseName: pennypilot-frontend
      valueFiles:
        - applications/pennypilot/frontend-values.yaml
```

### 4. Configuration Values

Copy the provided values files from this repository to your `homelab` repository:

1.  `deploy/homelab/backend-values.yaml` -> `applications/pennypilot/backend-values.yaml`
2.  `deploy/homelab/frontend-values.yaml` -> `applications/pennypilot/frontend-values.yaml`

Adjust the `hosts` and `environment` variables in these files to match your homelab's domain and network setup.
