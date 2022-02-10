FROM golang:1.17-alpine as builder
WORKDIR /go/src/github.com/target/pull-request-code-coverage
COPY . .
ENV GO111MODULE=on
ENV CGO_ENABLED=0
ENV GOOS=linux

RUN apk add --no-cache git && \
    go build -a -installsuffix cgo -o bin/plugin


FROM alpine:latest
COPY --from=builder /go/src/github.com/target/pull-request-code-coverage/bin/plugin /
RUN apk --no-cache add ca-certificates git bash openssh-client
WORKDIR /root/
COPY scripts/start.sh /
CMD ["/start.sh"]