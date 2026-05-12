FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o cabinet-rest ./cmd/server

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/cabinet-rest .
COPY --from=builder /app/config ./config

EXPOSE 8082

CMD ["./cabinet-rest"]