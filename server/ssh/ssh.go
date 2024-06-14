package ssh

import (
	"context"
	"fmt"
	"strconv"

	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/infastin/wg-wish/server/service"
	"github.com/rs/zerolog"
)

type Server struct {
	lg     zerolog.Logger
	server *ssh.Server
}

type ServerParams struct {
	Logger           zerolog.Logger
	Port             int
	HostKeyPath      string
	PublicKeyService service.PublicKeyService
	WireGuardService service.WireGuardService
}

func New(params *ServerParams) (srv *Server, err error) {
	addr := "0.0.0.0:" + strconv.Itoa(params.Port)
	srv = &Server{
		lg:     params.Logger,
		server: nil,
	}

	if srv.server, err = wish.NewServer(
		wish.WithAddress(addr),
		wish.WithHostKeyPath(params.HostKeyPath),
		wish.WithPublicKeyAuth(func(ctx ssh.Context, key ssh.PublicKey) bool {
			exists, _ := params.PublicKeyService.PublicKeyExists(ctx, key)
			return exists
		}),
		wish.WithMiddleware(
			NewCommandsHandler(&CommandsHandlerParams{
				Logger:           params.Logger,
				PublicKeyService: params.PublicKeyService,
				WireGuardService: params.WireGuardService,
			}),
			PanicHandler,
			NewLoggerMiddleware(params.Logger),
			ErrorHandler,
		),
	); err != nil {
		return nil, err
	}

	return srv, nil
}

func (s *Server) Run() error {
	err := s.server.ListenAndServe()
	if err != nil {
		return fmt.Errorf("ssh: failed to serve: %w", err)
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("ssh: failed to shutdown server: %w", err)
	}
	return nil
}
