FROM golang:1.23 AS builder

ENV GOPROXY=https://goproxy.cn

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o skadi_bot

FROM debian:bookworm-slim

WORKDIR /app
COPY --from=builder /app/skadi_bot .
COPY entrypoint.sh .
RUN chmod +x entrypoint.sh
CMD ["./entrypoint.sh"]
