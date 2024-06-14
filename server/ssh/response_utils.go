package ssh

import (
	"fmt"

	"github.com/charmbracelet/ssh"
)

func Abortf(handler ssh.Handler, session ssh.Session, f string, args ...any) {
	err := fmt.Errorf(f, args...)
	session.Context().SetValue("error", err)
	if handler != nil {
		handler(session)
	}
}

func AbortError(handler ssh.Handler, session ssh.Session, err error) {
	session.Context().SetValue("error", err)
	if handler != nil {
		handler(session)
	}
}
