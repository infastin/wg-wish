package queries

import "go.etcd.io/bbolt"

type Queries struct {
	tx *bbolt.Tx
}

func New(tx *bbolt.Tx) *Queries {
	return &Queries{
		tx: tx,
	}
}

func Prepare(db *bbolt.DB) (err error) {
	return db.Update(func(tx *bbolt.Tx) error {
		_, err = tx.CreateBucketIfNotExists(publicKeyBucketName)
		if err != nil {
			return err
		}

		_, err = tx.CreateBucketIfNotExists(wgClientBucketName)
		if err != nil {
			return err
		}

		_, err = tx.CreateBucketIfNotExists(wgServerBucketName)
		if err != nil {
			return err
		}

		return nil
	})
}
