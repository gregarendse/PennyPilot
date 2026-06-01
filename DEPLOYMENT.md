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

PennyPilot requires a PostgreSQL database. Ensure the database is accessible from the PennyPilot backend at its configured `DATABASE_URL`.

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

### 4. Configuration Examples

Create the following values files in your `homelab` repository under `applications/pennypilot/`.

#### `backend-values.yaml`

```yaml
image:
  repository: ghcr.io/gregarendse/pennypilot-backend
  tag: latest

keel:
  policy: force
  trigger: poll
  pollSchedule: "@midnight"
  approvals: "0"

environment:
  - name: HTTP_ADDR
    value: ":8080"
  - name: DATABASE_URL
    value: "postgres://pennypilot:pennypilot@pennypilot-db:5432/pennypilot?sslmode=disable"
  - name: FRONTEND_URL
    value: "https://pennypilot.arendse.nom.za"

service:
  type: ClusterIP

ports:
  http:
    target: 8080
    protocol: TCP

ingress:
  annotations:
    kubernetes.io/tls-acme: "true"
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
  hosts:
    - host: pennypilot-api.arendse.nom.za
      path: /
      tls: true
      service:
        name: http

secrets:
  environment:
    - pennypilot-backend-secrets
```

#### `frontend-values.yaml`

```yaml
image:
  repository: ghcr.io/gregarendse/pennypilot-frontend
  tag: latest

keel:
  policy: force
  trigger: poll
  pollSchedule: "@midnight"
  approvals: "0"

environment:
  - name: NEXT_PUBLIC_API_BASE_URL
    value: "https://pennypilot-api.arendse.nom.za"

service:
  type: ClusterIP

ports:
  http:
    target: 3000
    protocol: TCP

ingress:
  annotations:
    kubernetes.io/tls-acme: "true"
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
  hosts:
    - host: pennypilot.arendse.nom.za
      path: /
      tls: true
      service:
        name: http
```
