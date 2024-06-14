package wgservice

import (
	"bytes"
	"context"
	"net"

	"github.com/guregu/null/v5"
	"github.com/infastin/wg-wish/pkg/netutils"
	"github.com/infastin/wg-wish/pkg/wgtypes"
	"github.com/infastin/wg-wish/server/entity"
	"github.com/infastin/wg-wish/server/errors"
	"github.com/infastin/wg-wish/server/repo/db"
	wireguard "github.com/infastin/wg-wish/server/repo/wg"
	"github.com/infastin/wg-wish/server/service"
	"github.com/rs/zerolog"
)

type WireGuardServiceParams struct {
	Logger        zerolog.Logger
	DatabaseRepo  db.Repo
	WireGuardRepo wireguard.Repo

	Host                string
	Address             string
	Port                int
	Device              string
	DNS                 []net.IP
	AllowedIPs          []net.IPNet
	PersistentKeepalive null.Int
}

type WireGuardService struct {
	lg     zerolog.Logger
	dbRepo db.Repo
	wgRepo wireguard.Repo

	publicKey           wgtypes.Key
	address             net.IPNet
	port                int
	host                string
	dns                 []net.IP
	allowedIPs          []net.IPNet
	persistentKeepalive null.Int
	lastAddress         net.IPNet
}

func New(params *WireGuardServiceParams) (wgservice *WireGuardService, err error) {
	var address net.IPNet
	var lastAddress net.IPNet
	var publicKey wgtypes.Key

	ctx := context.Background()

	if err := params.DatabaseRepo.Update(ctx, func(repo db.Repo) error {
		config, err := repo.WireGuardServerRepo().GetWireGuardServerConfig()
		if err != nil && err != errors.ErrWireGuardServerConfigNotFound {
			return err
		}

		if err == errors.ErrWireGuardServerConfigNotFound {
			config.PrivateKey, err = wgtypes.GeneratePrivateKey()
			if err != nil {
				return err
			}

			err = repo.WireGuardServerRepo().SetWireGuardServerConfig(&config)
			if err != nil {
				return err
			}
		}

		cfg, err := wgtypes.NewServerConfig(
			&wgtypes.ServerConfigParams{
				PrivateKey: config.PrivateKey,
				Address:    params.Address,
				Device:     params.Device,
				ListenPort: null.IntFrom(int64(params.Port)),
			})
		if err != nil {
			return err
		}

		clients, err := repo.WireGuardClientRepo().GetWireGuardClients(ctx)
		if err != nil {
			return err
		}

		address = cfg.Interface.Address
		publicKey = config.PrivateKey.PublicKey()

		lastAddress = address
		lastAddress.Mask = net.CIDRMask(32, 32)
		lastIP := lastAddress.IP

		cfg.Peers = make([]wgtypes.ServerPeer, len(clients))
		for i := range clients {
			clientAddress := clients[i].Address
			if clientLastIP := netutils.LastIP(clientAddress); bytes.Compare(lastIP, clientLastIP) < 0 {
				lastAddress, lastIP = clientAddress, clientLastIP
			}

			cfg.Peers[i] = wgtypes.ServerPeer{
				Name:       clients[i].Name,
				PublicKey:  clients[i].PublicKey,
				AllowedIPs: []net.IPNet{clients[i].Address},
			}
		}

		return params.WireGuardRepo.LoadServerConfig(ctx, &cfg)
	}); err != nil {
		return nil, err
	}

	return &WireGuardService{
		lg:                  params.Logger,
		dbRepo:              params.DatabaseRepo,
		wgRepo:              params.WireGuardRepo,
		publicKey:           publicKey,
		address:             address,
		port:                params.Port,
		host:                params.Host,
		dns:                 params.DNS,
		allowedIPs:          params.AllowedIPs,
		persistentKeepalive: params.PersistentKeepalive,
		lastAddress:         lastAddress,
	}, nil
}

func (wg *WireGuardService) AddClient(ctx context.Context, name string, opts *service.AddClientOptions,
) (clientConfig wgtypes.ClientConfig, err error) {
	var client entity.WireGuardClient

	if err := wg.dbRepo.Update(ctx, func(repo db.Repo) error {
		exists, err := repo.WireGuardClientRepo().WireGuardClientExists(ctx, name)
		if err != nil {
			return err
		}

		if exists {
			return errors.ErrWireGuardClientExists
		}

		clientParams, err := wg.mapToAddClientParams(opts)
		if err != nil {
			return err
		}

		client = entity.WireGuardClient{
			Name:                name,
			Address:             clientParams.Address,
			PrivateKey:          clientParams.PrivateKey,
			PublicKey:           clientParams.PublicKey,
			DNS:                 clientParams.DNS,
			AllowedIPs:          clientParams.AllowedIPs,
			PersistentKeepalive: clientParams.PersistentKeepalive,
		}

		err = repo.WireGuardClientRepo().AddWireGuardClient(ctx, &client)
		if err != nil {
			return err
		}

		serverPeer := wgtypes.ServerPeer{
			Name:       name,
			PublicKey:  client.PublicKey,
			AllowedIPs: []net.IPNet{client.Address},
		}

		err = wg.wgRepo.AddServerPeer(ctx, &serverPeer)
		if err != nil {
			return err
		}

		wg.lastAddress = client.Address
		return nil
	}); err != nil {
		return wgtypes.ClientConfig{}, err
	}

	return wg.mapToClientConfig(&client), nil
}

type addClientParams struct {
	PrivateKey          wgtypes.Key
	PublicKey           wgtypes.Key
	Address             net.IPNet
	DNS                 []net.IP
	AllowedIPs          []net.IPNet
	PersistentKeepalive null.Int
}

func (wg *WireGuardService) mapToAddClientParams(opts *service.AddClientOptions) (params addClientParams, err error) {
	params.PrivateKey, err = wgtypes.GeneratePrivateKey()
	if err != nil {
		return addClientParams{}, err
	}

	params.PublicKey = params.PrivateKey.PublicKey()

	if opts != nil && opts.Address.Valid {
		params.Address = opts.Address.V

		peerSubnet := net.IPNet{
			IP:   params.Address.IP.Mask(params.Address.Mask),
			Mask: params.Address.Mask,
		}

		if peerSubnet.Contains(wg.address.IP) {
			return addClientParams{}, errors.ErrWireGuardClientAddressOverlaps
		}
	} else {
		params.Address, err = netutils.NextAddress(wg.lastAddress)
		if err != nil {
			return addClientParams{}, err
		}
	}

	if opts != nil && opts.DNS != nil {
		params.DNS = opts.DNS
	} else {
		params.DNS = wg.dns
	}

	if opts != nil && len(opts.AllowedIPs) != 0 {
		params.AllowedIPs = opts.AllowedIPs
	} else {
		params.AllowedIPs = wg.allowedIPs
	}

	if opts != nil && opts.PersistentKeepalive.Valid {
		params.PersistentKeepalive = opts.PersistentKeepalive
	} else {
		params.PersistentKeepalive = wg.persistentKeepalive
	}

	return params, nil
}

func (wg *WireGuardService) RemoveClient(ctx context.Context, name string) (err error) {
	return wg.dbRepo.Update(ctx, func(repo db.Repo) error {
		exists, err := repo.WireGuardClientRepo().WireGuardClientExists(ctx, name)
		if err != nil {
			return err
		}

		if !exists {
			return errors.ErrWireGuardClientNotFound
		}

		err = repo.WireGuardClientRepo().RemoveWireGuardClient(ctx, name)
		if err != nil {
			return err
		}

		return wg.wgRepo.RemoveServerPeer(ctx, name)
	})
}

func (wg *WireGuardService) GetClient(ctx context.Context, name string) (client wgtypes.ClientConfig, err error) {
	var dbClient entity.WireGuardClient

	if err := wg.dbRepo.View(ctx, func(repo db.Repo) error {
		dbClient, err = repo.WireGuardClientRepo().GetWireGuardClient(ctx, name)
		return err
	}); err != nil {
		return wgtypes.ClientConfig{}, err
	}

	return wg.mapToClientConfig(&dbClient), nil
}

func (wg *WireGuardService) GetClientInfos(ctx context.Context) (clients []entity.WireGuardClientInfo, err error) {
	var dbClients []entity.WireGuardClient

	if err := wg.dbRepo.View(ctx, func(repo db.Repo) error {
		dbClients, err = repo.WireGuardClientRepo().GetWireGuardClients(ctx)
		return err
	}); err != nil {
		return nil, err
	}

	peerStats, err := wg.wgRepo.GetPeerStats(ctx)
	if err != nil {
		if ie, ok := err.(errors.InternalError); ok {
			err = ie.Internal()
		}
		wg.lg.Err(err).Msg("failed to get peer stats")
	}

	clients = make([]entity.WireGuardClientInfo, len(dbClients))
	for i := range dbClients {
		clients[i].Config = wg.mapToClientConfig(&dbClients[i])
		if stats, ok := peerStats[dbClients[i].PublicKey]; ok {
			clients[i].Stats = null.ValueFrom(stats)
		}
	}

	return clients, nil
}

func (wg *WireGuardService) mapToClientConfig(client *entity.WireGuardClient) wgtypes.ClientConfig {
	return wgtypes.ClientConfig{
		Interface: wgtypes.ClientInterface{
			Name:       client.Name,
			Address:    client.Address,
			PrivateKey: client.PrivateKey,
			DNS:        client.DNS,
		},
		Peer: wgtypes.ClientPeer{
			Name:                "Server",
			EndpointHost:        wg.host,
			EndpointPort:        wg.port,
			PublicKey:           wg.publicKey,
			AllowedIPs:          client.AllowedIPs,
			PersistentKeepalive: client.PersistentKeepalive,
		},
	}
}

func (wg *WireGuardService) StartServer(ctx context.Context) (err error) {
	if err := wg.wgRepo.WriteServerConfig(ctx); err != nil {
		return err
	}
	return wg.wgRepo.StartServer(ctx)
}

func (wg *WireGuardService) StopServer(ctx context.Context) (err error) {
	return wg.wgRepo.StopServer(ctx)
}

func (wg *WireGuardService) ReloadServer(ctx context.Context) (err error) {
	if err := wg.wgRepo.WriteServerConfig(ctx); err != nil {
		return err
	}
	return wg.wgRepo.ReloadServer(ctx)
}
