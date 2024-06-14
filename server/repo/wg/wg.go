package wg

import (
	"context"

	"github.com/infastin/wg-wish/pkg/wgtypes"
	"github.com/infastin/wg-wish/server/entity"
)

type AtomicCallback func(repo Repo) error

type Repo interface {
	LoadServerConfig(ctx context.Context, config *wgtypes.ServerConfig) (err error)
	WriteServerConfig(ctx context.Context) (err error)
	AddServerPeer(ctx context.Context, peer *wgtypes.ServerPeer) (err error)
	RemoveServerPeer(ctx context.Context, name string) (err error)
	GetPeerStats(ctx context.Context) (stats map[wgtypes.Key]entity.WireGuardPeerStats, err error)
	StartServer(ctx context.Context) (err error)
	StopServer(ctx context.Context) (err error)
	ReloadServer(ctx context.Context) (err error)
}
