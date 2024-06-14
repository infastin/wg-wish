package db

import (
	"context"
)

type AtomicCallback func(repo Repo) error

type Repo interface {
	Update(ctx context.Context, cb AtomicCallback) (err error)
	View(ctx context.Context, cb AtomicCallback) (err error)
	Batch(ctx context.Context, cb AtomicCallback) (err error)
	PublicKeyRepo() PublicKeyRepo
	WireGuardClientRepo() WireGuardClientRepo
	WireGuardServerRepo() WireGuardServerRepo
}
