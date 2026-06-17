# Build SPA
FROM node:22 AS frontend-builder
WORKDIR /web
COPY web/package*.json ./
RUN npm install
COPY web/ ./
RUN npm run build

# Build Go binary
FROM golang:1.24 AS go-builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
RUN go build -o /out/ultratrader ./cmd/ultratrader

# Final image
FROM debian:bookworm-slim
WORKDIR /app
RUN apt-get update && apt-get install -y ca-certificates curl && rm -rf /var/lib/apt/lists/*
COPY --from=go-builder /out/ultratrader /usr/local/bin/ultratrader
COPY --from=frontend-builder /web/dist ./web/dist
COPY config ./config
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/ultratrader"]
CMD ["--config", "/app/config/development-timer.json"]
