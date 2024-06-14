package publickeyservice

import (
	"context"

	"github.com/charmbracelet/ssh"
	"github.com/infastin/wg-wish/server/entity"
	"github.com/infastin/wg-wish/server/repo/db"
	"github.com/rs/zerolog"
)

type PublicKeyServiceParams struct {
	Logger zerolog.Logger
	Repo   db.Repo
}

type PublicKeyService struct {
	lg   zerolog.Logger
	repo db.Repo
}

func New(params *PublicKeyServiceParams) *PublicKeyService {
	return &PublicKeyService{
		lg:   params.Logger,
		repo: params.Repo,
	}
}

func (s *PublicKeyService) AddPublicKey(ctx context.Context, pkey *entity.PublicKey) (err error) {
	return s.repo.Update(ctx, func(repo db.Repo) error {
		return repo.PublicKeyRepo().AddPublicKey(ctx, pkey)
	})
}

func (s *PublicKeyService) PublicKeyExists(ctx context.Context, pkey ssh.PublicKey) (exists bool, err error) {
	if err := s.repo.View(ctx, func(repo db.Repo) error {
		exists, err = repo.PublicKeyRepo().PublicKeyExists(ctx, pkey)
		return err
	}); err != nil {
		return false, err
	}
	return exists, nil
}

func (s *PublicKeyService) RemovePublicKey(ctx context.Context, pkey ssh.PublicKey) (err error) {
	return s.repo.Update(ctx, func(repo db.Repo) error {
		return repo.PublicKeyRepo().RemovePublicKey(ctx, pkey)
	})
}

func (s *PublicKeyService) GetPublicKeys(ctx context.Context) (pkeys []entity.PublicKey, err error) {
	err = s.repo.View(ctx, func(repo db.Repo) error {
		pkeys, err = repo.PublicKeyRepo().GetPublicKeys(ctx)
		return err
	})
	return pkeys, err
}
