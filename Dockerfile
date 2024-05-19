FROM golang:1.18 as builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o name ./namenode

FROM debian:buster-slim

COPY --from=builder /app/name /app/name

EXPOSE 50052

ENTRYPOINT ["/app/name"]
