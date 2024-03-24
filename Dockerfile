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

COPY .env /app

EXPOSE 8080 

CMD ["/app/pcbookApp"]