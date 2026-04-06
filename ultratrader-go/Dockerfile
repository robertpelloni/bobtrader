FROM golang:1.26 AS builder
WORKDIR /src
COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal
RUN go build -o /out/ultratrader ./cmd/ultratrader

FROM debian:bookworm-slim
WORKDIR /app
COPY --from=builder /out/ultratrader /usr/local/bin/ultratrader
COPY config ./config
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/ultratrader"]
CMD ["--config", "/app/config/development-timer.json"]
