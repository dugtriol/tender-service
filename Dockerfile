FROM golang:alpine
WORKDIR /app

RUN go version
ENV GOPATH=/

COPY ./ ./

RUN go mod download
RUN go build -o app ./cmd/app/main.go

CMD ["./app"]