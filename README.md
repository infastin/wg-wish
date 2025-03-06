# wg-wish

Server application that simplifies creation and management of WireGuard VPN server.

Somewhat similar to [wg-eazy](https://github.com/wg-easy/wg-easy),
but uses [wish](https://github.com/charmbracelet/wish) library
to provide SSH interface for managing WireGuard server and peers.

Use the following Docker Compose file with your own public SSH keys and server host:
```yaml
services:
  wg-wish:
    image: ghcr.io/infastin/wg-wish
    environment:
      WG_HOST: ADDRESS OF YOUR SERVER (e.g. 23.192.228.84 or example.com)
      SSH_ADMIN_KEYS: >-
        YOUR AUTHORIZED KEYS GO HERE
    ports:
      - 51820:51820/udp
      - 51822:51822/tcp
    restart: unless-stopped
    volumes:
      - wg-wish-data:/var/lib/wg-wish
    cap_add:
      - NET_ADMIN
      - SYS_MODULE
    sysctls:
      - net.ipv4.ip_forward=1
      - net.ipv4.conf.all.src_valid_mark=1
volumes:
  wg-wish-data:
```

Explore the CLI with just an SSH client:
```console
$ ssh localhost -p 51822 -- --help
Usage: wg-wish <command> [flags]

Manage WireGuard.

Flags:
  -h, --help    Show context-sensitive help.

Commands:
  publickey <command> [flags]
    Manage public keys.

  wireguard <command> [flags]
    Manage WireGuard.

Run "wg-wish <command> --help" for more information on a command.
Error: expected one of "publickey",  "wireguard"
```

Add a new peer:
```console
$ ssh localhost -p 51822 -- wireguard add NAME
# NAME
[Interface]
Address    = 10.9.8.2/32
PrivateKey = uO9Ei4kF0DMVfApyY8CCDiR0+YZjuB+SJV0HZHYog3M=
DNS        = 1.1.1.1,8.8.8.8

# Server
[Peer]
Endpoint            = 23.192.228.84:51820
PublicKey           = 21x/13xfAzf1qMxrQAabVh5lM3dbRzriE59ohieriiU=
AllowedIPs          = 0.0.0.0/0
PersistentKeepalive = 25
```

Reload WireGuard itself to make the new peer work:
```console
$ ssh localhost -p 51822 -- wireguard reload
```
