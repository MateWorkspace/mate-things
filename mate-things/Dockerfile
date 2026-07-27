# syntax=docker/dockerfile:1

################################################################################
# Backend build stage
################################################################################
FROM golang:1.26-alpine AS backend-builder

WORKDIR /src/backend

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o /out/backend ./cmd/main
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o /out/seeder ./cmd/seeder

################################################################################
# Frontend build stage
################################################################################
FROM node:22-alpine AS frontend-builder

WORKDIR /src/frontend

COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci

COPY frontend/ ./
RUN npm run build

################################################################################
# Runtime stage - backend + frontend behind a single nginx port, /api routed
# to the backend and everything else routed to the frontend. Postgres
# migrations run once on container start, before the app starts serving.
################################################################################
FROM node:22-alpine AS runtime

RUN apk add --no-cache nginx bash jq

# Migration runner (static binary) + migration files
COPY --from=migrate/migrate:v4.18.1 /usr/local/bin/migrate /usr/local/bin/migrate
COPY backend/database/migrations /app/migrations

# Backend
COPY --from=backend-builder /out/backend /app/backend/backend
COPY --from=backend-builder /out/seeder /app/backend/seeder

# Frontend (Next.js standalone output)
COPY --from=frontend-builder /src/frontend/.next/standalone /app/frontend
COPY --from=frontend-builder /src/frontend/.next/static /app/frontend/.next/static
COPY --from=frontend-builder /src/frontend/public /app/frontend/public

# Proxy
COPY docker/nginx.conf /etc/nginx/nginx.conf
COPY docker/entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

EXPOSE 80

ENTRYPOINT ["entrypoint.sh"]
