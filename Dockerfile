FROM golang:1.18 as builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o director ./director

FROM debian:buster-slim

COPY --from=builder /app/director /app/director

EXPOSE 50051

ENTRYPOINT ["/app/director"]