FROM golang:1.26.3-alpine AS builder
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
# Run the plugin directly. With PARAMETER_DIFF_SOURCE=github it fetches the PR
# diff from the GitHub API; for the stdin path, pipe a `git diff` into the
# container (docker run -i ... | git diff ...).
ENTRYPOINT ["/plugin"]