package errdefer

type Closer interface {
	Close() error
}

type Releaser interface {
	Release() error
}

func Close[C Closer](err *error, closer C) error {
	if *err != nil {
		return closer.Close()
	}
	return nil
}

func Release[R Releaser](err *error, releaser R) error {
	if *err != nil {
		return releaser.Release()
	}
	return nil
}
