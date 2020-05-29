FROM alpine:latest
COPY bin/plugin /
RUN apk --no-cache add ca-certificates git bash openssh-client
WORKDIR /root/
COPY scripts/start.sh /
CMD ["/start.sh"]