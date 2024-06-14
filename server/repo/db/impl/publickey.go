package dbrepo

import (
	"context"

	"github.com/charmbracelet/ssh"
	"github.com/infastin/wg-wish/server/entity"
	"github.com/infastin/wg-wish/server/errors"
	"github.com/infastin/wg-wish/server/repo/db/impl/queries"
)

func (db *DatabaseRepo) AddPublicKey(ctx context.Context, pkey *entity.PublicKey) (err error) {
	if db.queries.PublicKeyExists(pkey.Key) {
		return errors.ErrPublicKeyExists
	}

	return db.queries.SetPublicKey(&queries.PublicKey{
		Key:     pkey.Key,
		Comment: pkey.Comment,
	})
}

func (db *DatabaseRepo) SetPublicKey(ctx context.Context, pkey *entity.PublicKey) (err error) {
	return db.queries.SetPublicKey(&queries.PublicKey{
		Key:     pkey.Key,
		Comment: pkey.Comment,
	})
}

func (db *DatabaseRepo) PublicKeyExists(ctx context.Context, pkey ssh.PublicKey) (exists bool, err error) {
	return db.queries.PublicKeyExists(pkey), nil
}

func (db *DatabaseRepo) RemovePublicKey(ctx context.Context, pkey ssh.PublicKey) (err error) {
	return db.queries.RemovePublicKey(pkey)
}

func (db *DatabaseRepo) GetPublicKeys(ctx context.Context) (pkeys []entity.PublicKey, err error) {
	keys, err := db.queries.GetPublicKeys()
	if err != nil {
		return nil, err
	}

	pkeys = make([]entity.PublicKey, 0, len(keys))
	for i := range keys {
		pkeys = append(pkeys, entity.PublicKey{
			Key:     keys[i].Key,
			Comment: keys[i].Comment,
		})
	}

	return pkeys, nil
}

func (db *DatabaseRepo) SetPublicKeys(ctx context.Context, pkeys []entity.PublicKey) (err error) {
	err = db.queries.ClearPublicKeys()
	if err != nil {
		return err
	}

	for _, pkey := range pkeys {
		if err := db.queries.SetPublicKey(&queries.PublicKey{
			Key:     pkey.Key,
			Comment: pkey.Comment,
		}); err != nil {
			return err
		}
	}

	return nil
}
