# Deployment and continuous delivery

This repository owns application source, container images, and deployment examples only. Keep cluster-specific Kubernetes manifests, secrets, hostnames, storage classes, and GitOps automation in a separate deployment repository.

## What this repo now publishes

The `Container images` GitHub Actions workflow builds two multi-architecture images:

| Service | GHCR image | Optional Docker Hub image |
| --- | --- | --- |
| Backend API | `ghcr.io/<owner>/pennypilot-backend` | `<dockerhub-user>/pennypilot-backend` |
| Frontend UI | `ghcr.io/<owner>/pennypilot-frontend` | `<dockerhub-user>/pennypilot-frontend` |

The workflow runs for pull requests, pushes to `main`, version tags such as `v0.1.0`, and manual `workflow_dispatch` runs. Pull requests build the images without pushing them. Pushes and tags publish images.

Image tags include:

- `latest` for the default branch.
- The branch name for branch builds.
- The Git tag for releases, for example `v0.1.0`.
- An immutable commit tag such as `sha-<git-sha>`.

## Registry setup

### GitHub Container Registry (recommended default)

GHCR publishing uses the built-in `GITHUB_TOKEN`; no extra secrets are required. Make sure repository Actions are allowed to write packages:

1. Open **Settings → Actions → General**.
2. Under **Workflow permissions**, select **Read and write permissions**.
3. Save the change.

If the package is private, your deployment cluster needs an image pull secret with a GitHub token that can read packages.

### Docker Hub (optional mirror)

Add these repository secrets if you also want Docker Hub images:

| Secret | Value |
| --- | --- |
| `DOCKERHUB_USERNAME` | Docker Hub username or organization account |
| `DOCKERHUB_TOKEN` | Docker Hub access token with write access |

If these secrets are absent, the workflow still publishes to GHCR.

## Publishing release builds

Create and push a semantic version Git tag:

```bash
git tag v0.1.0
git push origin v0.1.0
```

That publishes both images with the `v0.1.0` tag as well as the immutable `sha-<git-sha>` tag. Use the version tag for human-readable releases and the SHA tag when you want a deployment repo to pin an exact build.

## Continuous deployment model

Use the separate deployment repository as the source of truth for your cluster. This application repo should only publish images and examples.

A typical loop is:

1. Merge code to `main` in this repository.
2. GitHub Actions builds and publishes `pennypilot-backend` and `pennypilot-frontend` images.
3. The deployment repository updates Kubernetes image tags, either manually or with automation.
4. A GitOps controller such as Flux or Argo CD reconciles the cluster.
5. You test the running app and iterate.

Recommended deployment-repo options:

- **Flux Image Automation** watches GHCR or Docker Hub and writes new tags back to the deployment repo.
- **Argo CD Image Updater** updates image tags consumed by Argo CD applications.
- **Renovate** watches container image tags and opens pull requests against the deployment repo.
- **Manual promotion** updates tags after you validate a build.

For fast iteration in a test namespace, track `main` or `sha-*` tags. For stable environments, promote `v*` release tags.

## Kubernetes example

The sample manifests in `examples/kubernetes/base` are intentionally generic. Copy them into your deployment repo and customize:

- Image registry, owner, and tags.
- Hostnames and TLS configuration.
- Secret management, preferably External Secrets, Sealed Secrets, SOPS, or your cloud provider's secret manager.
- PostgreSQL persistence and backups.
- Resource requests and limits.
- Ingress class and annotations.

Apply the sample directly only for a throwaway namespace:

```bash
kubectl apply -k examples/kubernetes/base
```

Before applying, replace the placeholder images and secrets in the copied manifests.

## Runtime configuration checklist

Backend:

- `DATABASE_URL` pointing at PostgreSQL.
- OAuth credentials for Monzo and TrueLayer.
- `ENCRYPTION_KEY_HEX` generated with `openssl rand -hex 32`.
- Public callback URLs that match the OAuth applications.

Frontend:

- `NEXT_PUBLIC_API_BASE_URL` set to the browser-reachable API URL.

Infrastructure:

- PostgreSQL database with durable storage and backups.
- Ingress or gateway routing for frontend and backend.
- Image pull credentials if the registry packages are private.
- TLS certificates, usually via cert-manager.
