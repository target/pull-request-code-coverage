FROM alpine:latest
COPY /go/src/git.target.com/searchoss/pull-request-code-coverage/plugin /
RUN apk --no-cache add ca-certificates git bash openssh-client
WORKDIR /root/
COPY scripts/start.sh /
CMD ["/start.sh"]