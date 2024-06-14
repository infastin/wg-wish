package entity

import (
	"net"

	"github.com/guregu/null/v5"
	"github.com/infastin/wg-wish/pkg/wgtypes"
)

type WireGuardClient struct {
	Name                string
	Address             net.IPNet
	PrivateKey          wgtypes.Key
	PublicKey           wgtypes.Key
	DNS                 []net.IP
	AllowedIPs          []net.IPNet
	PersistentKeepalive null.Int
}

type WireGuardPeerStats struct {
	Received        uint64
	Sent            uint64
	LatestHandshake null.Time
}

type WireGuardClientInfo struct {
	Config wgtypes.ClientConfig
	Stats  null.Value[WireGuardPeerStats]
}
