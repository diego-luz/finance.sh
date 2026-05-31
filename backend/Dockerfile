# syntax=docker/dockerfile:1
#
# finance.sh — single application image.
#
# The frontend SPA is built and then EMBEDDED into the Go binary (internal/web,
# go:embed), so one container serves both the static app and the JSON API on a
# single HTTP port. There is no bundled nginx/TLS: self-hosters put their own
# reverse proxy (Traefik, Caddy, Nginx Proxy Manager, ...) in front to terminate
# TLS.
#
# Build context MUST be the project root (parent of backend/ and frontend/):
#   docker build -t finance-sh -f backend/Dockerfile .

# ---- stage 1: build the SPA bundle -----------------------------------------
FROM node:22-alpine AS spa
WORKDIR /spa

# Install deps first for better layer caching; fall back to install if no lock.
COPY frontend/package.json frontend/package-lock.json* ./
RUN npm ci || npm install

# Source tree (respects frontend/.dockerignore).
COPY frontend/ ./

# Vite bakes these in at build time. Defaults match a same-origin deploy: the
# SPA and the API are served by the very same binary, so /api/v1 is relative.
ARG VITE_API_URL=/api/v1
ENV VITE_API_URL=${VITE_API_URL}
ARG VITE_LANDING_URL=
ENV VITE_LANDING_URL=${VITE_LANDING_URL}

RUN npm run build

# ---- stage 2: build the Go binary (with the SPA embedded) ------------------
FROM golang:1.22-alpine AS build
WORKDIR /src

# Cache module downloads separately from the source tree.
COPY backend/go.mod backend/go.sum* ./
RUN go mod download

COPY backend/ .

# Replace the committed placeholder with the real SPA bundle so go:embed bakes
# the production app into the binary.
RUN rm -rf internal/web/dist
COPY --from=spa /spa/dist ./internal/web/dist

# Static binary so the final image needs no libc.
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

# ---- stage 3: runtime ------------------------------------------------------
# PROD: pin by digest, e.g. FROM alpine:3.20@sha256:...
FROM alpine:3.20
# wget (busybox) is used by the compose healthcheck; ca-certs/tzdata for TLS+time.
RUN apk add --no-cache ca-certificates tzdata && adduser -D -u 10001 finance_sh
WORKDIR /app

COPY --from=build /out/api /app/api
COPY backend/docs /app/docs

ENV APP_PORT=8080 DOCS_DIR=/app/docs
EXPOSE 8080
# Run unprivileged. The compose runtime also sets read_only rootfs + no-new-privileges.
USER finance_sh

ENTRYPOINT ["/app/api"]
