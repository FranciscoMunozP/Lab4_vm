FROM golang:1.18 as builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o datanode ./datanode

FROM debian:buster-slim

COPY --from=builder /app/datanode /app/datanode

EXPOSE 50053

ENTRYPOINT ["/app/datanode"]