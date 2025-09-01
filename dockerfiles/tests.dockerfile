# dockerfiles/tests.dockerfile
FROM golang:1.22

WORKDIR /app

COPY gateway/go.mod gateway/go.sum .

RUN go mod download

COPY ./tests ./tests

CMD ["go", "test", "./tests", "-tags=integration", "-v", "-count=1"]