package wgrepo

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"sync"
	"time"

	"github.com/guregu/null/v5"
	"github.com/infastin/gorack/fastconv"
	"github.com/infastin/wg-wish/pkg/wgtypes"
	"github.com/infastin/wg-wish/server/entity"
	"github.com/infastin/wg-wish/server/errors"
	"github.com/rs/zerolog"
)

type WireGuardRepoParams struct {
	Logger zerolog.Logger

	Path string
}

type WireGuardRepo struct {
	lg zerolog.Logger

	path   string
	config wgtypes.ServerConfig
	mu     *sync.RWMutex
}

func New(params *WireGuardRepoParams) *WireGuardRepo {
	return &WireGuardRepo{
		lg:     params.Logger,
		path:   params.Path,
		config: wgtypes.ServerConfig{},
		mu:     &sync.RWMutex{},
	}
}

func (wg *WireGuardRepo) LoadServerConfig(ctx context.Context, cfg *wgtypes.ServerConfig) (err error) {
	wg.config = *cfg
	return nil
}

func (wg *WireGuardRepo) WriteServerConfig(ctx context.Context) (err error) {
	file, err := os.Create(wg.path)
	if err != nil {
		return err
	}
	defer file.Close()

	return wg.config.Encode(file)
}

func (wg *WireGuardRepo) AddServerPeer(ctx context.Context, client *wgtypes.ServerPeer) (err error) {
	wg.mu.Lock()
	defer wg.mu.Unlock()

	for i := range wg.config.Peers {
		if wg.config.Peers[i].Name == client.Name {
			return errors.ErrWireGuardServerPeerExists
		}
	}

	wg.config.Peers = append(wg.config.Peers, *client)
	return nil
}

func (wg *WireGuardRepo) RemoveServerPeer(ctx context.Context, name string) (err error) {
	wg.mu.Lock()
	defer wg.mu.Unlock()

	for i := range wg.config.Peers {
		if wg.config.Peers[i].Name == name {
			wg.config.Peers = slices.Delete(wg.config.Peers, i, i+1)
			return nil
		}
	}

	return errors.ErrWireGuardServerPeerNotFound
}

func (*WireGuardRepo) GetPeerStats(ctx context.Context) (stats map[wgtypes.Key]entity.WireGuardPeerStats, err error) {
	cmd := exec.CommandContext(ctx, "wg", "show", "wg0", "dump")

	out, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			err = errors.NewCommandError(cmd, err, fastconv.String(ee.Stderr))
		}
		return nil, err
	}

	nl := bytes.IndexByte(out, '\n')
	if nl == -1 {
		return stats, nil
	}

	out = out[nl+1:]
	stats = make(map[wgtypes.Key]entity.WireGuardPeerStats)
	publicKey := wgtypes.Key{}
	stat := entity.WireGuardPeerStats{}
	fieldNum := 0

	for len(out) != 0 {
		end := len(out)
		for i, c := range out {
			if c == ' ' || c == '\t' || c == '\n' {
				end = i
				break
			}
		}

		field := fastconv.String(out[:end])
		out = out[end:]

		switch fieldNum {
		case 0:
			publicKey, err = wgtypes.ParseKey(field)
			if err != nil {
				return nil, err
			}
		case 4:
			timestamp, err := strconv.ParseInt(field, 10, 64)
			if err != nil {
				return nil, err
			}
			if timestamp != 0 {
				stat.LatestHandshake = null.TimeFrom(time.Unix(timestamp, 0))
			}
		case 5:
			stat.Received, err = strconv.ParseUint(field, 10, 64)
			if err != nil {
				return nil, err
			}
		case 6:
			stat.Sent, err = strconv.ParseUint(field, 10, 64)
			if err != nil {
				return nil, err
			}
		}

		fieldNum = (fieldNum + 1) % 8
		if fieldNum == 0 {
			stats[publicKey] = stat
			stat = entity.WireGuardPeerStats{}
		}

		start := len(out)
		for i, c := range out {
			if c != ' ' && c != '\t' && c != '\n' {
				start = i
				break
			}
		}

		out = out[start:]
	}

	return stats, nil
}

func (wg *WireGuardRepo) StartServer(ctx context.Context) (err error) {
	cmd := exec.CommandContext(ctx, "wg-quick", "up", wg.path) //nolint:gosec
	return cmd.Run()
}

func (wg *WireGuardRepo) StopServer(ctx context.Context) (err error) {
	cmd := exec.CommandContext(ctx, "wg-quick", "down", wg.path) //nolint:gosec
	return cmd.Run()
}

func (wg *WireGuardRepo) ReloadServer(ctx context.Context) (err error) {
	f, err := os.CreateTemp("", "wg0-*.conf")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())

	var stderr bytes.Buffer

	stripCmd := exec.CommandContext(ctx, "wg-quick", "strip", wg.path) //nolint:gosec
	stripCmd.Stdout = f
	stripCmd.Stderr = &stderr

	if err := stripCmd.Run(); err != nil {
		return errors.NewCommandError(stripCmd, err, stderr.String())
	}

	wgCmd := exec.CommandContext(ctx, "wg", "syncconf", "wg0", f.Name()) //nolint:gosec
	wgCmd.Stderr = &stderr

	if err := wgCmd.Run(); err != nil {
		return errors.NewCommandError(wgCmd, err, stderr.String())
	}

	return nil
}
