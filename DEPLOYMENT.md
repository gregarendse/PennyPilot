# PennyPilot Deployment Guide (Homelab)

This guide explains how to deploy PennyPilot to your homelab using the infrastructure patterns found in [gregarendse/homelab](https://github.com/gregarendse/homelab).

## Prerequisites

1.  A Kubernetes cluster managed by Argo CD (as per the `homelab` repo).
2.  GitHub Container Registry (GHCR) access (or your own registry).
3.  A PostgreSQL database.

## Automated Image Builds

The `.github/workflows/ci.yml` in this repository is configured to build and push the combined app Docker image to GHCR on every push to `master`.

Image:
- `ghcr.io/gregarendse/pennypilot:latest`

## Homelab Integration Steps

### 1. Database Setup

PennyPilot requires a PostgreSQL database. Ensure the database is accessible from the PennyPilot app at its configured `DATABASE_URL`.

### 2. Secret Management

Create a Kubernetes secret named `pennypilot-backend-secrets` in the namespace where you plan to deploy the app. This secret should contain the following environment variables from `.env.example`:

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
  - name: pennypilot
    type: helm
    namespace: pennypilot
    helm:
      kind: path
      path: server
      releaseName: pennypilot
      valueFiles:
        - applications/pennypilot/values.yaml
```

### 4. Configuration Example

Create the following values file in your `homelab` repository under `applications/pennypilot/values.yaml`.

```yaml
image:
  repository: ghcr.io/gregarendse/pennypilot
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
    - host: pennypilot.arendse.nom.za
      path: /
      tls: true
      service:
        name: http

secrets:
  environment:
    - pennypilot-backend-secrets
```
