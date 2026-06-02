# syntax=docker/dockerfile:1

FROM node:22-alpine AS frontend-deps
WORKDIR /src/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci

FROM node:22-alpine AS frontend-build
WORKDIR /src/frontend
ENV NEXT_PUBLIC_API_BASE_URL=""
COPY --from=frontend-deps /src/frontend/node_modules ./node_modules
COPY frontend/ ./
RUN npm run build

FROM golang:1.24-alpine AS backend-build
RUN apk add --no-cache build-base
WORKDIR /src/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=1 GOOS=linux go build \
    -tags 'osusergo,netgo,sqlite_omit_load_extension' \
    -ldflags='-linkmode external -extldflags "-static" -s -w' \
    -o /out/pennypilot ./cmd/server

FROM alpine:3.20 AS certs
RUN apk add --no-cache ca-certificates

FROM scratch
WORKDIR /app
ENV HTTP_ADDR=:8080
ENV STATIC_PATH=/app/public
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=backend-build /out/pennypilot /app/pennypilot
COPY --from=frontend-build /src/frontend/out /app/public
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/app/pennypilot"]
