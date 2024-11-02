FROM golang:1.23.1

RUN mkdir /app

ADD . /app

WORKDIR /app

RUN go get notice-me-server
RUN go build ./cmd/server/main.go

CMD ["/app/main"]
