package db

import (
	"context"

	"github.com/infastin/wg-wish/server/entity"
)

type WireGuardClientRepo interface {
	AddWireGuardClient(ctx context.Context, client *entity.WireGuardClient) (err error)
	RemoveWireGuardClient(ctx context.Context, name string) (err error)
	WireGuardClientExists(ctx context.Context, name string) (exists bool, err error)
	GetWireGuardClient(ctx context.Context, name string) (client entity.WireGuardClient, err error)
	GetWireGuardClients(ctx context.Context) (clients []entity.WireGuardClient, err error)
}
