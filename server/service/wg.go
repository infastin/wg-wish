package service

import (
	"context"
	"net"

	"github.com/guregu/null/v5"
	"github.com/infastin/wg-wish/pkg/wgtypes"
	"github.com/infastin/wg-wish/server/entity"
)

type AddClientOptions struct {
	Address             null.Value[net.IPNet]
	DNS                 []net.IP
	AllowedIPs          []net.IPNet
	PersistentKeepalive null.Int
}

type WireGuardService interface {
	AddClient(ctx context.Context, name string, opts *AddClientOptions) (client wgtypes.ClientConfig, err error)
	RemoveClient(ctx context.Context, name string) (err error)
	GetClient(ctx context.Context, name string) (client wgtypes.ClientConfig, err error)
	GetClientInfos(ctx context.Context) (clients []entity.WireGuardClientInfo, err error)
	ReloadServer(ctx context.Context) (err error)
}
