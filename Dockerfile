FROM golang:1.22.4-alpine3.20 as build
RUN apk add make
WORKDIR /app
COPY . .
RUN make build

FROM alpine:3.20
HEALTHCHECK CMD /usr/bin/timeout 5s /bin/sh -c "/usr/bin/wg show | /bin/grep -q interface || exit 1" --interval=1m --timeout=5s --retries=3
COPY --from=build /app/build/server /usr/bin/wg-wish
RUN apk add --no-cache iptables wireguard-tools
WORKDIR /var/lib/wg-wish
ENTRYPOINT ["/usr/bin/wg-wish"]
