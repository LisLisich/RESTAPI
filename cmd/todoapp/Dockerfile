FROM node:22-alpine AS web-builder
WORKDIR /app/web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web ./
RUN npm run build

FROM golang:1.25.12-bookworm AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=web-builder /app/public /app/public
RUN CGO_ENABLED=0 GOOS=linux go build -o /todoapp ./cmd/todoapp

FROM alpine:3.23
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /todoapp /app/todoapp
COPY --from=builder /app/public /app/public
ENV PROJECT_ROOT=/app
CMD ["/app/todoapp"]
