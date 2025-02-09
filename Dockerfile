FROM golang:1.23 AS builder

ENV GOPROXY=https://goproxy.cn

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o skadi_bot

FROM debian:bookworm-slim

WORKDIR /app
RUN apt-get update && apt-get install -y ca-certificates=20230311 --no-install-recommends && rm -rf /var/lib/apt/lists/*
COPY --from=builder /app/skadi_bot .
COPY entrypoint.sh .
RUN chmod +x entrypoint.sh
CMD ["./entrypoint.sh"]
