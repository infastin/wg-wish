package ssh

import (
	"context"

	"github.com/alecthomas/kong"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/infastin/wg-wish/server/service"
	"github.com/rs/zerolog"
)

type Context struct {
	context.Context

	kctx *kong.Context

	lg      zerolog.Logger
	session ssh.Session

	publicKeyService service.PublicKeyService
	wireguardService service.WireGuardService
}

type CommandsHandlerParams struct {
	Logger           zerolog.Logger
	PublicKeyService service.PublicKeyService
	WireGuardService service.WireGuardService
}

func NewCommandsHandler(params *CommandsHandlerParams) wish.Middleware {
	return func(handler ssh.Handler) ssh.Handler {
		return func(session ssh.Session) {
			var cli struct {
				PublicKey PublicKeyCmd `cmd:"" name:"publickey" help:"Manage public keys."`
				WireGuard WireGuardCmd `cmd:"" name:"wireguard" help:"Manage WireGuard."`
			}

			k, err := kong.New(&cli,
				kong.Writers(session, session.Stderr()),
				kong.Name("wg-wish"),
				kong.Description("Manage WireGuard."),
				kong.ConfigureHelp(kong.HelpOptions{ //nolint:exhaustruct
					NoExpandSubcommands: true,
				}),
				kong.Exit(func(i int) {}),
			)

			if err != nil {
				Abortf(handler, session, "could not initialize session: %v", err)
				return
			}

			kctx, err := k.Parse(session.Command())
			if err != nil {
				AbortError(handler, session, err)
				return
			}

			err = kctx.Run(&Context{
				Context:          context.Background(),
				kctx:             kctx,
				lg:               params.Logger,
				session:          session,
				publicKeyService: params.PublicKeyService,
				wireguardService: params.WireGuardService,
			})

			if err != nil {
				AbortError(handler, session, err)
				return
			}

			handler(session)
		}
	}
}
