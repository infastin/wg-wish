package ssh

import (
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/infastin/wg-wish/server/errors"
	"github.com/rs/zerolog"
)

func NewLoggerMiddleware(lg zerolog.Logger) wish.Middleware {
	return func(handler ssh.Handler) ssh.Handler {
		return func(session ssh.Session) {
			ct := time.Now()
			hpk := session.PublicKey() != nil
			command := getCommandString(session.Command())
			user := session.User()
			addr := session.RemoteAddr()

			handler(session)

			elapsed := time.Since(ct)

			lgCtx := lg.With().
				Str("command", command).
				Str("user", user).
				Str("addr", addr.String()).
				Bool("hpk", hpk).
				Dur("elapsed", elapsed)

			switch e := session.Context().Value("error").(type) {
			case *errors.PanicError:
				lg := lgCtx.
					Any("panic", e.Panic).
					Bytes("stack", e.Stack).
					Logger()
				lg.Error().Msg("command panic")
			case errors.InternalError:
				lg := lgCtx.Logger()
				lg.Err(e.Internal()).Msg("command error")
			case errors.DomainError:
				lg := lgCtx.
					Str("domain", e.Domain()).
					Logger()
				lg.Err(e).Msg("command error")
			case error:
				lg := lgCtx.Logger()
				lg.Err(e).Msg("command error")
			case nil:
				lg := lgCtx.Logger()
				lg.Info().Msg("command ok")
			}
		}
	}
}

func getCommandString(args []string) string {
	var b strings.Builder

	for i := 0; i < len(args); i++ {
		if i != 0 {
			switch {
			case strings.HasPrefix(args[i-1], "--"):
				b.WriteByte('=')
			default:
				b.WriteByte(' ')
			}
		}

		if strings.ContainsAny(args[i], " \t\n") {
			b.WriteString(strconv.Quote(args[i]))
		} else {
			b.WriteString(args[i])
		}
	}

	return b.String()
}
