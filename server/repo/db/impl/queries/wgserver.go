package queries

import "github.com/infastin/wg-wish/pkg/wgtypes"

//go:generate msgp -tests=false -unexported

var (
	wgServerBucketName = []byte("wgserver")
	wgServerConfigKey  = []byte("config")
)

//msgp:tuple wgServerConfigValueV1
//msgp:replace wgtypes.Key with:[32]byte

type wgServerConfigValueV1 struct {
	PrivateKey wgtypes.Key
}

func wgServerConfigMarshalValueV1(b []byte, value *wgServerConfigValueV1) []byte {
	b, _ = value.MarshalMsg(b)
	return b
}

func wgServerConfigUnmarshalValueV1(b []byte) (value wgServerConfigValueV1, err error) {
	_, err = value.UnmarshalMsg(b)
	return value, err
}

//msgp:ignore WireGuardServerConfig

type WireGuardServerConfig struct {
	PrivateKey wgtypes.Key
}

func (queries *Queries) SetWireGuardServerConfig(config *WireGuardServerConfig) (err error) {
	b := queries.tx.Bucket(wgServerBucketName)

	valb := Meta(0).Append(nil)
	valb = wgServerConfigMarshalValueV1(valb, &wgServerConfigValueV1{
		PrivateKey: config.PrivateKey,
	})

	return b.Put(wgServerConfigKey, valb)
}

func (queries *Queries) GetWireGuardServerConfig() (config WireGuardServerConfig, err error) {
	b := queries.tx.Bucket(wgServerBucketName)

	valb := b.Get(wgServerConfigKey)
	if valb == nil {
		return WireGuardServerConfig{}, ErrKeyNotFound
	}

	val, err := wgServerConfigUnmarshalValueV1(valb[1:])
	if err != nil {
		return WireGuardServerConfig{}, err
	}

	return WireGuardServerConfig(val), nil
}

func (queries *Queries) WireGuardServerConfigExists() (exists bool) {
	b := queries.tx.Bucket(wgServerBucketName)
	return b.Get(wgServerConfigKey) != nil
}
