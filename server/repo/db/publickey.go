package db

import (
	"context"

	"github.com/charmbracelet/ssh"
	"github.com/infastin/wg-wish/server/entity"
)

type PublicKeyRepo interface {
	AddPublicKey(ctx context.Context, pkey *entity.PublicKey) (err error)
	SetPublicKey(ctx context.Context, pkey *entity.PublicKey) (err error)
	PublicKeyExists(ctx context.Context, pkey ssh.PublicKey) (exists bool, err error)
	RemovePublicKey(ctx context.Context, pkey ssh.PublicKey) (err error)
	GetPublicKeys(ctx context.Context) (pkeys []entity.PublicKey, err error)
	SetPublicKeys(ctx context.Context, pkeys []entity.PublicKey) (err error)
}
