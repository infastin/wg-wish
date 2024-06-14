package queries

import (
	"github.com/charmbracelet/ssh"
)

//go:generate msgp -tests=false -unexported

var publicKeyBucketName = []byte("publickey")

func publicKeyMarshalKey(b []byte, pkey ssh.PublicKey) []byte {
	return append(b, pkey.Marshal()...)
}

func publicKeyUnmarshalKey(b []byte) (pkey ssh.PublicKey, err error) {
	return ssh.ParsePublicKey(b)
}

//msgp:tuple publicKeyValueV1

type publicKeyValueV1 struct {
	Comment string
}

func publicKeyMarshalValueV1(b []byte, value *publicKeyValueV1) []byte {
	b, _ = value.MarshalMsg(b)
	return b
}

func publicKeyUnmarshalValueV1(b []byte) (val publicKeyValueV1, err error) {
	_, err = val.UnmarshalMsg(b)
	return val, err
}

//msgp:ignore PublicKey

type PublicKey struct {
	Key     ssh.PublicKey
	Comment string
}

func (queries *Queries) SetPublicKey(pkey *PublicKey) (err error) {
	b := queries.tx.Bucket(publicKeyBucketName)

	keyb := publicKeyMarshalKey(nil, pkey.Key)

	valb := Meta(0).Append(nil)
	valb = publicKeyMarshalValueV1(valb, &publicKeyValueV1{Comment: pkey.Comment})

	return b.Put(keyb, valb)
}

func (queries *Queries) GetPublicKey(pkey ssh.PublicKey) (key PublicKey, err error) {
	b := queries.tx.Bucket(publicKeyBucketName)

	keyb := publicKeyMarshalKey(nil, pkey)

	valb := b.Get(keyb)
	if valb == nil {
		return PublicKey{}, ErrKeyNotFound
	}

	val, err := publicKeyUnmarshalValueV1(valb[1:])
	if err != nil {
		return PublicKey{}, err
	}

	return PublicKey{
		Key:     pkey,
		Comment: val.Comment,
	}, nil
}

func (queries *Queries) GetPublicKeys() (keys []PublicKey, err error) {
	b := queries.tx.Bucket(publicKeyBucketName)

	c := b.Cursor()
	for keyb, valb := c.First(); keyb != nil; keyb, valb = c.Next() {
		key, err := publicKeyUnmarshalKey(keyb)
		if err != nil {
			return nil, err
		}

		val, err := publicKeyUnmarshalValueV1(valb[1:])
		if err != nil {
			return nil, err
		}

		keys = append(keys, PublicKey{
			Key:     key,
			Comment: val.Comment,
		})
	}

	return keys, nil
}

func (queries *Queries) ClearPublicKeys() (err error) {
	b := queries.tx.Bucket(publicKeyBucketName)

	c := b.Cursor()
	for keyb, _ := c.First(); keyb != nil; keyb, _ = c.Next() {
		err = c.Delete()
		if err != nil {
			return err
		}
	}

	return nil
}

func (queries *Queries) RemovePublicKey(pkey ssh.PublicKey) (err error) {
	b := queries.tx.Bucket(publicKeyBucketName)
	keyb := publicKeyMarshalKey(nil, pkey)
	return b.Delete(keyb)
}

func (queries *Queries) PublicKeyExists(pkey ssh.PublicKey) (exists bool) {
	b := queries.tx.Bucket(publicKeyBucketName)
	keyb := publicKeyMarshalKey(nil, pkey)
	return b.Get(keyb) != nil
}
