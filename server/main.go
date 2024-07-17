package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"syscall"

	charmssh "github.com/charmbracelet/ssh"
	"github.com/guregu/null/v5"
	"github.com/infastin/wg-wish/pkg/netutils"
	"github.com/infastin/wg-wish/server/app"
	"github.com/infastin/wg-wish/server/errors"
	dbrepo "github.com/infastin/wg-wish/server/repo/db/impl"
	wgrepo "github.com/infastin/wg-wish/server/repo/wg/impl"
	publickeyservice "github.com/infastin/wg-wish/server/service/impl/publickey"
	wgservice "github.com/infastin/wg-wish/server/service/impl/wg"
	"github.com/infastin/wg-wish/server/ssh"
	"github.com/oklog/run"
	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

func runApp(args []string) (err error) {
	cli, err := app.NewCLI(args)
	if err != nil {
		return fmt.Errorf("failed to parse command line arguments: %w", err)
	}

	config, err := app.NewConfig(cli.Config)
	if err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	var logWriters []io.Writer

	if config.Logger.Directory != "" {
		logs := &lumberjack.Logger{
			Filename:   path.Join(config.Logger.Directory, "server.log"),
			MaxSize:    config.Logger.MaxSize,
			MaxAge:     config.Logger.MaxAge,
			MaxBackups: config.Logger.MaxBackups,
			LocalTime:  false,
			Compress:   false,
		}
		defer logs.Close()

		logWriters = append(logWriters, logs)
	}

	loggerLevel, err := zerolog.ParseLevel(config.Logger.Level)
	if err != nil {
		return err
	}

	logger, err := app.NewLogger(
		&app.LoggerParams{
			Level:             loggerLevel,
			AdditionalWriters: logWriters,
		})
	if err != nil {
		return err
	}

	dbRepo, err := dbrepo.New(
		&dbrepo.DatabaseRepoParams{
			Logger:    logger.With().Str("tag", "db_repo").Logger(),
			Path:      config.Database.Path,
			AdminKeys: config.SSH.AdminKeys,
		})
	if err != nil {
		return err
	}
	defer func() {
		if err := dbRepo.Close(); err != nil {
			logger.Err(err).Msg("failed to close database repo")
		}
	}()

	wgRepo := wgrepo.New(
		&wgrepo.WireGuardRepoParams{
			Logger: logger.With().Str("tag", "wg_repo").Logger(),
			Path:   config.WireGuard.Path,
		})

	pubKeyService := publickeyservice.New(
		&publickeyservice.PublicKeyServiceParams{
			Logger: logger.With().Str("tag", "pubkey_service").Logger(),
			Repo:   dbRepo,
		})

	dns, err := netutils.ParseIPs(config.WireGuard.DNS)
	if err != nil {
		return err
	}

	ips, err := netutils.ParseAddresses(config.WireGuard.AllowedIPs)
	if err != nil {
		return err
	}

	wireguardService, err := wgservice.New(
		&wgservice.WireGuardServiceParams{
			Logger:              logger.With().Str("tag", "wg_service").Logger(),
			DatabaseRepo:        dbRepo,
			WireGuardRepo:       wgRepo,
			Host:                config.WireGuard.Host,
			Address:             config.WireGuard.Address,
			Port:                config.WireGuard.Port,
			Device:              config.WireGuard.Device,
			DNS:                 dns,
			AllowedIPs:          ips,
			PersistentKeepalive: null.IntFrom(int64(config.WireGuard.PersistentKeepalive)),
		})
	if err != nil {
		return err
	}

	ctx := context.Background()

	err = wireguardService.StartServer(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err := wireguardService.StopServer(ctx); err != nil {
			logger.Err(err).Msg("failed to stop wireguard server")
		}
	}()

	sshSrv, err := ssh.New(
		&ssh.ServerParams{
			Logger:           logger.With().Str("tag", "ssh").Logger(),
			Port:             config.SSH.Port,
			HostKeyPath:      config.SSH.HostKeyPath,
			PublicKeyService: pubKeyService,
			WireGuardService: wireguardService,
		})
	if err != nil {
		return err
	}

	g := new(run.Group)

	g.Add(run.SignalHandler(ctx, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM))

	g.Add(func() error {
		logger.Info().Msg("starting ssh server")
		if err := sshSrv.Run(); err != nil && !errors.Is(err, charmssh.ErrServerClosed) {
			logger.Err(err).Msg("failed to start ssh server")
			return err
		}
		return nil
	}, func(err error) {
		logger.Info().Msg("shutting down ssh server")
		if err := sshSrv.Shutdown(ctx); err != nil {
			logger.Err(err).Msg("failed to shutdown ssh server")
		}
	})

	if err := g.Run(); err != nil {
		if _, ok := err.(run.SignalError); !ok {
			return err
		}
	}

	return nil
}

func main() {
	if err := runApp(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "failed to run: %s\n", err)
		os.Exit(1)
	}
}
