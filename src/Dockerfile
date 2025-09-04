# Use official Golang image
FROM golang:1.25

WORKDIR /app

# Copy go.mod and go.sum before copying the rest
COPY go.mod ./
COPY go.sum ./
RUN go mod download

# Now copy the source files
COPY . .

# Build the Go binary
RUN go build -o main .

CMD ["./main"]