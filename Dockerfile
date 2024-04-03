# Build stage
FROM golang:1.22-alpine3.19 AS builder

RUN mkdir app

COPY . /app

WORKDIR /app

RUN CGO_ENABLED=0 go build -o pcbookApp ./cmd/server/main.go

RUN chmod +x /app/pcbookApp

# Run stage
FROM alpine:3.19

RUN mkdir /app

COPY --from=builder /app/pcbookApp /app

COPY .env .

RUN mkdir /cert
COPY cert/server-cert.pem /cert
COPY cert/server-key.pem /cert
COPY cert/ca-cert.pem /cert

EXPOSE 8080 

CMD ["/app/pcbookApp", "-port", "8080", "-tls", "true"]
