FROM golang:1.25.0-bookworm AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o todoapp ./cmd/todoapp

FROM alpine:3.23
WORKDIR /app
COPY --from=builder /app/cmd/todoapp/exe /app
CMD ["/app/exe"]