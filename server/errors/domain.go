package errors

type InternalError interface {
	error
	Internal() error
}

type internalError struct {
	err error
}

func NewInternalError(err error) InternalError {
	return &internalError{
		err: err,
	}
}

func (*internalError) Error() string {
	return "internal error"
}

func (e *internalError) Internal() error {
	return e.err
}

type DomainError interface {
	error
	Domain() string
}

type domainError struct {
	domain string
	msg    string
}

func NewDomainError(domain, msg string) DomainError {
	return &domainError{
		domain: domain,
		msg:    msg,
	}
}

func (e *domainError) Error() string {
	return e.msg
}

func (e *domainError) Domain() string {
	return e.domain
}
