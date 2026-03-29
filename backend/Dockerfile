FROM golang:1.26-alpine AS builder

WORKDIR /src/backend

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend ./

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/tradelab-api ./cmd/api && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/tradelab-migrate ./cmd/migrate

FROM alpine:3.22

RUN addgroup -S tradelab && adduser -S tradelab -G tradelab

WORKDIR /app

COPY --from=builder /out/tradelab-api /app/tradelab-api
COPY --from=builder /out/tradelab-migrate /app/tradelab-migrate
COPY backend/migrations /app/migrations

USER tradelab

EXPOSE 8080

ENTRYPOINT ["/app/tradelab-api"]
