package ssh

import (
	"github.com/infastin/wg-wish/server/errors"
	"runtime"

	"github.com/charmbracelet/ssh"
)

func PanicHandler(handler ssh.Handler) ssh.Handler {
	return func(session ssh.Session) {
		defer func() {
			if p := recover(); p != nil {
				stack := make([]byte, 1<<16)
				stack = stack[:runtime.Stack(stack, false)]
				AbortError(nil, session, errors.NewPanicError(p, stack))
			}
		}()
		handler(session)
	}
}
