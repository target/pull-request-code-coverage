FROM golang:1.11.2
RUN go get -u github.com/golang/dep/cmd/dep
RUN go get -u github.com/alecthomas/gometalinter
RUN gometalinter --install
COPY . /go/src/git.target.com/searchoss/pull-request-code-coverage
WORKDIR /go/src/git.target.com/searchoss/pull-request-code-coverage
RUN dep ensure
RUN go test -v -coverpkg=./... -coverprofile=coverage.txt ./...
RUN go tool cover -func=coverage.txt
RUN ./bin/lint.sh
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o plugin .

FROM alpine:latest
RUN apk --no-cache add ca-certificates git bash openssh-client
WORKDIR /root/
COPY --from=0 /go/src/git.target.com/searchoss/pull-request-code-coverage/plugin /
COPY scripts/start.sh /
CMD ["/start.sh"]