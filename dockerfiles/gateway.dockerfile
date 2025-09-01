FROM golang:1.23.6-alpine AS builder

WORKDIR /app

COPY go.* /app/

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o gateway ./cmd/main.go

FROM alpine:latest

RUN apk add --no-cache bash curl

WORKDIR /app

COPY --from=builder /app/gateway .

EXPOSE 8080

CMD ["./gateway"]