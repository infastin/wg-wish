package errors

import (
	"fmt"
	"os/exec"
)

var (
	ErrPublicKeyExists                = NewDomainError("pubkey", "public key already exists")
	ErrPublicKeyNotFound              = NewDomainError("pubkey", "public key not found")
	ErrWireGuardClientExists          = NewDomainError("wg", "wireguard client already exists")
	ErrWireGuardClientNotFound        = NewDomainError("wg", "wireguard client not found")
	ErrWireGuardClientAddressOverlaps = NewDomainError("wg", "wireguard client address overlaps with wireguard server address")
	ErrWireGuardServerPeerExists      = NewInternalError(NewDomainError("wg", "wireguard server peer already exists"))
	ErrWireGuardServerPeerNotFound    = NewInternalError(NewDomainError("wg", "wireguard server peer not found"))
	ErrWireGuardServerConfigNotFound  = NewInternalError(NewDomainError("wg", "wireguard server config not found"))
)

type PanicError struct {
	Panic any
	Stack []byte
}

func NewPanicError(p any, stack []byte) error {
	return &PanicError{
		Panic: p,
		Stack: stack,
	}
}

func (*PanicError) Error() string {
	return "internal error"
}

type CommandError struct {
	internal error
}

func NewCommandError(cmd *exec.Cmd, err error, msg string) error {
	return &CommandError{
		internal: fmt.Errorf("failed to run %q: err=%w, msg=%s", cmd, err, msg),
	}
}

func (*CommandError) Error() string {
	return "internal error"
}

func (e *CommandError) Internal() error {
	return e.internal
}

func (e *CommandError) Unwrap() error {
	return e.internal
}
