package ssh

import (
	"github.com/alecthomas/kong"
	"github.com/infastin/wg-wish/server/errors"

	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
)

func ErrorHandler(handler ssh.Handler) ssh.Handler {
	return func(session ssh.Session) {
		handler(session)

		switch v := session.Context().Value("error").(type) {
		case *kong.ParseError, errors.DomainError:
			wish.Fatalf(session, "Error: %s\n", v)
		case error:
			wish.Fatalln(session, "Internal Error")
		}
	}
}
