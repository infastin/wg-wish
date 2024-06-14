package errdefer

type Closer interface {
	Close() error
}

type Releaser interface {
	Release() error
}

func Close(err *error, closer Closer) error {
	if *err != nil {
		return closer.Close()
	}
	return nil
}

func Release(err *error, releaser Releaser) error {
	if *err != nil {
		return releaser.Release()
	}
	return nil
}
