package db

import "github.com/infastin/wg-wish/server/entity"

type WireGuardServerRepo interface {
	SetWireGuardServerConfig(config *entity.WireGuardServerConfig) (err error)
	GetWireGuardServerConfig() (config entity.WireGuardServerConfig, err error)
	WireGuardServerConfigExists() (has bool, err error)
}
