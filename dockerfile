FROM golang:1.26-bookworm AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 go build -o cabinet-rest ./cmd/server


FROM debian:bookworm-slim

WORKDIR /app

COPY --from=builder /app/cabinet-rest .
COPY --from=builder /app/config ./config

EXPOSE 8082

CMD ["./cabinet-rest"]