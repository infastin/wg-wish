package dbrepo

import (
	"github.com/infastin/wg-wish/server/entity"
	"github.com/infastin/wg-wish/server/errors"
	"github.com/infastin/wg-wish/server/repo/db/impl/queries"
)

func (db *DatabaseRepo) SetWireGuardServerConfig(config *entity.WireGuardServerConfig) (err error) {
	return db.queries.SetWireGuardServerConfig(&queries.WireGuardServerConfig{
		PrivateKey: config.PrivateKey,
	})
}

func (db *DatabaseRepo) GetWireGuardServerConfig() (config entity.WireGuardServerConfig, err error) {
	cfg, err := db.queries.GetWireGuardServerConfig()
	if err != nil {
		if err == queries.ErrKeyNotFound {
			err = errors.ErrWireGuardServerConfigNotFound
		}
		return entity.WireGuardServerConfig{}, err
	}

	return entity.WireGuardServerConfig{
		PrivateKey: cfg.PrivateKey,
	}, nil
}

func (db *DatabaseRepo) WireGuardServerConfigExists() (exists bool, err error) {
	return db.queries.WireGuardServerConfigExists(), nil
}
