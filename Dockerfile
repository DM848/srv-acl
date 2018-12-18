FROM golang:1.11.2 as builder

WORKDIR /app
COPY . /app

RUN go test ./...
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o webserver cmd/webserver/server.go

FROM dm848/consul-service:v3
WORKDIR /server

COPY . /server
COPY --from=builder /app/webserver .
RUN chmod +x /server/webserver

ENV WEB_SERVER_PORT 8888
EXPOSE 8888
CMD ["/server/start-acl.sh"]
