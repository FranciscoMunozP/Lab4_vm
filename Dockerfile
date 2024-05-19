FROM golang:1.18 as builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o banco ./Bank

FROM debian:buster-slim

COPY --from=builder /app/banco /app/banco

EXPOSE 50054

ENTRYPOINT ["/app/banco"]