package queries

import (
	"net"

	"github.com/guregu/null/v5"
	"github.com/infastin/wg-wish/pkg/wgtypes"
)

//go:generate msgp -tests=false -unexported

var wgClientBucketName = []byte("wgclient")

func wgClientMarshalKey(b []byte, name string) []byte {
	return append(b, name...)
}

func wgClientUnmarshalKey(b []byte) (name string, err error) {
	return string(b), nil
}

//msgp:tuple wgClientValueV1
//msgp:replace wgtypes.Key with:[32]byte
//msgp:replace net.IP with:[]byte
//msgp:replace net.IPNet with:msgpIPNet
//msgp:replace null.Int with:msgpNullInt

type wgClientValueV1 struct {
	Address             net.IPNet
	PrivateKey          wgtypes.Key
	PublicKey           wgtypes.Key
	DNS                 []net.IP
	AllowedIPs          []net.IPNet
	PersistentKeepalive *int64
}

func wgClientMarshalValueV1(b []byte, value *wgClientValueV1) []byte {
	b, _ = value.MarshalMsg(b)
	return b
}

func wgClientUnmarshalValueV1(b []byte) (val wgClientValueV1, err error) {
	_, err = val.UnmarshalMsg(b)
	return val, err
}

//msgp:ignore WireGuardClient

type WireGuardClient struct {
	Name                string
	Address             net.IPNet
	PrivateKey          wgtypes.Key
	PublicKey           wgtypes.Key
	DNS                 []net.IP
	AllowedIPs          []net.IPNet
	PersistentKeepalive null.Int
}

func (queries *Queries) SetWireGuardClient(client *WireGuardClient) (err error) {
	b := queries.tx.Bucket(wgClientBucketName)

	keyb := wgClientMarshalKey(nil, client.Name)

	valb := Meta(0).Append(nil)
	valb = wgClientMarshalValueV1(valb, &wgClientValueV1{
		Address:             client.Address,
		PrivateKey:          client.PrivateKey,
		PublicKey:           client.PublicKey,
		DNS:                 client.DNS,
		AllowedIPs:          client.AllowedIPs,
		PersistentKeepalive: client.PersistentKeepalive.Ptr(),
	})

	return b.Put(keyb, valb)
}

func (queries *Queries) GetWireGuardClient(name string) (client WireGuardClient, err error) {
	b := queries.tx.Bucket(wgClientBucketName)

	keyb := wgClientMarshalKey(nil, name)

	valb := b.Get(keyb)
	if valb == nil {
		return WireGuardClient{}, ErrKeyNotFound
	}

	val, err := wgClientUnmarshalValueV1(valb[1:])
	if err != nil {
		return WireGuardClient{}, err
	}

	return WireGuardClient{
		Name:                name,
		Address:             val.Address,
		PrivateKey:          val.PrivateKey,
		PublicKey:           val.PublicKey,
		DNS:                 val.DNS,
		AllowedIPs:          val.AllowedIPs,
		PersistentKeepalive: null.IntFromPtr(val.PersistentKeepalive),
	}, nil
}

func (queries *Queries) GetWireGuardClients() (clients []WireGuardClient, err error) {
	b := queries.tx.Bucket(wgClientBucketName)

	c := b.Cursor()
	for keyb, valb := c.First(); keyb != nil; keyb, valb = c.Next() {
		key, err := wgClientUnmarshalKey(keyb)
		if err != nil {
			return nil, err
		}

		val, err := wgClientUnmarshalValueV1(valb[1:])
		if err != nil {
			return nil, err
		}

		clients = append(clients, WireGuardClient{
			Address:             val.Address,
			Name:                key,
			PrivateKey:          val.PrivateKey,
			PublicKey:           val.PublicKey,
			DNS:                 val.DNS,
			AllowedIPs:          val.AllowedIPs,
			PersistentKeepalive: null.IntFromPtr(val.PersistentKeepalive),
		})
	}

	return clients, nil
}

func (queries *Queries) RemoveWireGuardClient(name string) (err error) {
	b := queries.tx.Bucket(wgClientBucketName)
	keyb := wgClientMarshalKey(nil, name)
	return b.Delete(keyb)
}

func (queries *Queries) WireGuardClientExists(name string) (exists bool) {
	b := queries.tx.Bucket(wgClientBucketName)
	keyb := wgClientMarshalKey(nil, name)
	return b.Get(keyb) != nil
}
