FROM golang:1.27

WORKDIR /usr/src/app

COPY . .
RUN go mod download
RUN go build -v -o /usr/local/bin/app ./...

CMD ["app"]
