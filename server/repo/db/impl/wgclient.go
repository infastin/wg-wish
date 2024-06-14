package dbrepo

import (
	"context"

	"github.com/infastin/wg-wish/server/entity"
	"github.com/infastin/wg-wish/server/errors"
	"github.com/infastin/wg-wish/server/repo/db/impl/queries"
)

func (db *DatabaseRepo) AddWireGuardClient(ctx context.Context, client *entity.WireGuardClient) (err error) {
	if db.queries.WireGuardClientExists(client.Name) {
		return errors.ErrWireGuardClientExists
	}

	return db.queries.SetWireGuardClient(&queries.WireGuardClient{
		Name:                client.Name,
		Address:             client.Address,
		PrivateKey:          client.PrivateKey,
		PublicKey:           client.PublicKey,
		DNS:                 client.DNS,
		AllowedIPs:          client.AllowedIPs,
		PersistentKeepalive: client.PersistentKeepalive,
	})
}

func (db *DatabaseRepo) RemoveWireGuardClient(ctx context.Context, name string) (err error) {
	return db.queries.RemoveWireGuardClient(name)
}

func (db *DatabaseRepo) WireGuardClientExists(ctx context.Context, name string) (exists bool, err error) {
	return db.queries.WireGuardClientExists(name), nil
}

func (db *DatabaseRepo) GetWireGuardClient(ctx context.Context, name string) (client entity.WireGuardClient, err error) {
	dbClient, err := db.queries.GetWireGuardClient(name)
	if err != nil {
		return entity.WireGuardClient{}, err
	}
	return mapToWireGuardClient(&dbClient), nil
}

func (db *DatabaseRepo) GetWireGuardClients(ctx context.Context) (clients []entity.WireGuardClient, err error) {
	dbClients, err := db.queries.GetWireGuardClients()
	if err != nil {
		return nil, err
	}

	clients = make([]entity.WireGuardClient, len(dbClients))
	for i := range dbClients {
		clients[i] = mapToWireGuardClient(&dbClients[i])
	}

	return clients, nil
}

func mapToWireGuardClient(client *queries.WireGuardClient) entity.WireGuardClient {
	return entity.WireGuardClient{
		Name:                client.Name,
		Address:             client.Address,
		PrivateKey:          client.PrivateKey,
		PublicKey:           client.PublicKey,
		DNS:                 client.DNS,
		AllowedIPs:          client.AllowedIPs,
		PersistentKeepalive: client.PersistentKeepalive,
	}
}
