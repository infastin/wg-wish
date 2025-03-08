# Build
FROM golang:1.23-alpine3.21 AS build
RUN apk add git

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o build/server ./server

# Run
FROM alpine:3.21

HEALTHCHECK CMD /usr/bin/timeout 5s /bin/sh -c "/usr/bin/wg show | /bin/grep -q interface || exit 1" --interval=1m --timeout=5s --retries=3

RUN apk add --no-cache iptables wireguard-tools

COPY --from=build /app/build/server /usr/bin/wg-wish
WORKDIR /var/lib/wg-wish

ENTRYPOINT ["/usr/bin/wg-wish"]
