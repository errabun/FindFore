# Stage 1: Build frontend
FROM node:22-alpine AS frontend
WORKDIR /app/frontend
COPY frontend/package.json frontend/package-lock.json* ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# Stage 2: Build Go binary
FROM golang:1.25-alpine AS backend
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /app/frontend/dist ./frontend/dist
RUN CGO_ENABLED=0 GOOS=linux go build -o /findfore .
RUN CGO_ENABLED=0 GOOS=linux go build -o /findfore-migrate ./cmd/migrate

# Stage 3: Minimal runtime
FROM gcr.io/distroless/static-debian12
COPY --from=backend /findfore /findfore
COPY --from=backend /findfore-migrate /findfore-migrate
COPY --from=backend /app/frontend/dist /frontend/dist
COPY --from=backend /app/migrations /migrations
ENV PORT=8080
EXPOSE 8080
CMD ["/findfore"]
