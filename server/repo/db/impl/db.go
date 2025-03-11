package dbrepo

import (
	"context"

	"github.com/charmbracelet/ssh"
	"github.com/infastin/gorack/errdefer"
	"github.com/infastin/gorack/fastconv"
	"github.com/infastin/wg-wish/server/entity"
	"github.com/infastin/wg-wish/server/errors"
	database "github.com/infastin/wg-wish/server/repo/db"
	"github.com/infastin/wg-wish/server/repo/db/impl/queries"
	"github.com/rs/zerolog"
	"go.etcd.io/bbolt"
)

var ErrTxNotStarted = errors.New("transaction not started")

type DatabaseRepoParams struct {
	Logger zerolog.Logger

	Path      string
	AdminKeys []string
}

type DatabaseRepo struct {
	lg      zerolog.Logger
	db      *bbolt.DB
	queries *queries.Queries
}

func New(params *DatabaseRepoParams) (dbrepo *DatabaseRepo, err error) {
	db, err := bbolt.Open(params.Path, 0600, nil)
	if err != nil {
		return nil, err
	}
	defer errdefer.Close(&err, db.Close)

	err = queries.Prepare(db)
	if err != nil {
		return nil, err
	}

	repo := &DatabaseRepo{
		lg:      params.Logger,
		db:      db,
		queries: nil,
	}

	ctx := context.Background()

	if err := repo.Update(ctx, func(repo database.Repo) error {
		for _, adminKey := range params.AdminKeys {
			pkey, comment, _, _, err := ssh.ParseAuthorizedKey(fastconv.Bytes(adminKey))
			if err != nil {
				return err
			}

			if err := repo.PublicKeyRepo().SetPublicKey(ctx, &entity.PublicKey{
				Key:     pkey,
				Comment: comment,
			}); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return repo, nil
}

func (db *DatabaseRepo) Close() error {
	return db.db.Close()
}

func (db *DatabaseRepo) atomic(callback database.AtomicCallback, tx *bbolt.Tx) (err error) {
	return callback(&DatabaseRepo{
		lg:      db.lg,
		db:      db.db,
		queries: queries.New(tx),
	})
}

func (db *DatabaseRepo) Update(ctx context.Context, callback database.AtomicCallback) (err error) {
	return db.db.Update(func(tx *bbolt.Tx) error {
		return db.atomic(callback, tx)
	})
}

func (db *DatabaseRepo) View(ctx context.Context, callback database.AtomicCallback) (err error) {
	return db.db.View(func(tx *bbolt.Tx) error {
		return db.atomic(callback, tx)
	})
}

func (db *DatabaseRepo) Batch(ctx context.Context, callback database.AtomicCallback) (err error) {
	return db.db.Batch(func(tx *bbolt.Tx) error {
		return db.atomic(callback, tx)
	})
}

func (db *DatabaseRepo) PublicKeyRepo() database.PublicKeyRepo {
	if db.queries == nil {
		panic(ErrTxNotStarted)
	}
	return db
}

func (db *DatabaseRepo) WireGuardClientRepo() database.WireGuardClientRepo {
	if db.queries == nil {
		panic(ErrTxNotStarted)
	}
	return db
}

func (db *DatabaseRepo) WireGuardServerRepo() database.WireGuardServerRepo {
	if db.queries == nil {
		panic(ErrTxNotStarted)
	}
	return db
}
