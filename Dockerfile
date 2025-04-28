FROM golang:1.24

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod tidy

COPY . .

RUN mkdir -p /app/build \
    && go build -o /app/build/balancer ./cmd/load-balancer \
    && go clean -cache -modcache

EXPOSE 8080

CMD ["/app/build/balancer"]
