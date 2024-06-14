package ssh

import (
	"bytes"
	"fmt"
	"net"

	"github.com/guregu/null/v5"
	"github.com/infastin/wg-wish/pkg/netutils"
	"github.com/infastin/wg-wish/server/service"
	"github.com/mdp/qrterminal/v3"
)

type WireGuardCmd struct {
	Add struct {
		Name                string      `arg:"" help:"Client's name."`
		Address             null.String `optional:"" short:"a" placeholder:"ADDR" help:"Client's address."`
		DNS                 []string    `optional:"" short:"d" help:"Client's DNS list."`
		AllowedIPs          []string    `optional:"" short:"i" name:"ips" placeholder:"IP" help:"Client's allowed IPs."`
		PersistentKeepalive null.Int    `optional:"" short:"k" name:"keepalive" placeholder:"SECONDS" help:"Client's persistent keepalive."`
		QR                  bool        `optional:"" name:"qr" help:"Print QR code."`
	} `cmd:"" help:"Add client."`

	Get struct {
		Name string `arg:"" help:"Client's name."`
		QR   bool   `optional:"" name:"qr" help:"Print QR code."`
	} `cmd:"" help:"Get client config."`

	Rm struct {
		Name string `arg:"" help:"Client's name."`
	} `cmd:"" help:"Remove client."`

	Reload struct{} `cmd:"" help:"Reload server."`

	Ls struct{} `cmd:"" help:"List clients."`
}

func (cmd *WireGuardCmd) Run(ctx *Context) (err error) {
	switch ctx.kctx.Command() {
	case "wireguard add <name>":
		err = cmd.HandleAdd(ctx)
	case "wireguard rm <name>":
		err = cmd.HandleRm(ctx)
	case "wireguard get <name>":
		err = cmd.HandleGet(ctx)
	case "wireguard reload":
		err = cmd.HandleReload(ctx)
	case "wireguard ls":
		err = cmd.HandleLs(ctx)
	}
	return err
}

func (cmd *WireGuardCmd) HandleAdd(ctx *Context) (err error) {
	var address null.Value[net.IPNet]
	if cmd.Add.Address.Valid {
		addr, err := netutils.ParseAddress(cmd.Add.Address.String)
		if err != nil {
			return err
		}
		address = null.ValueFrom(addr)
	}

	var dns []net.IP
	if cmd.Add.DNS != nil {
		dns, err = netutils.ParseIPs(cmd.Add.DNS)
		if err != nil {
			return err
		}
	}

	var ips []net.IPNet
	if cmd.Add.AllowedIPs != nil {
		ips, err = netutils.ParseAddresses(cmd.Add.AllowedIPs)
		if err != nil {
			return err
		}
	}

	cfg, err := ctx.wireguardService.AddClient(ctx, cmd.Add.Name,
		&service.AddClientOptions{
			Address:             address,
			DNS:                 dns,
			AllowedIPs:          ips,
			PersistentKeepalive: cmd.Add.PersistentKeepalive,
		})
	if err != nil {
		return err
	}

	var conf bytes.Buffer
	if err := cfg.Encode(&conf); err != nil {
		return err
	}

	if cmd.Add.QR {
		qrterminal.GenerateHalfBlock(conf.String(), qrterminal.L, ctx.session)
	} else {
		_, _ = ctx.session.Write(conf.Bytes())
	}

	return nil
}

func (cmd *WireGuardCmd) HandleRm(ctx *Context) (err error) {
	return ctx.wireguardService.RemoveClient(ctx, cmd.Rm.Name)
}

func (cmd *WireGuardCmd) HandleGet(ctx *Context) (err error) {
	cfg, err := ctx.wireguardService.GetClient(ctx, cmd.Get.Name)
	if err != nil {
		return err
	}

	var conf bytes.Buffer
	if err := cfg.Encode(&conf); err != nil {
		return err
	}

	if cmd.Get.QR {
		qrterminal.GenerateHalfBlock(conf.String(), qrterminal.L, ctx.session)
	} else {
		_, _ = ctx.session.Write(conf.Bytes())
	}

	return nil
}

func (*WireGuardCmd) HandleReload(ctx *Context) (err error) {
	return ctx.wireguardService.ReloadServer(ctx)
}

func (*WireGuardCmd) HandleLs(ctx *Context) (err error) {
	infos, err := ctx.wireguardService.GetClientInfos(ctx)
	if err != nil {
		return err
	}

	var b bytes.Buffer
	for i := range infos {
		info := &infos[i]
		fmt.Fprintf(&b, "%d. %s\n", i+1, infos[i].Config.Interface.Name)
		fmt.Fprintf(&b, "Address: %s\n", infos[i].Config.Interface.Address.String())
		if info.Stats.Valid {
			fmt.Fprintf(&b, "Received: %s\n", humanReadableByteCount(info.Stats.V.Received))
			fmt.Fprintf(&b, "Sent: %s\n", humanReadableByteCount(info.Stats.V.Sent))
			if info.Stats.V.LatestHandshake.Valid {
				fmt.Fprintf(&b, "Latest handshake: %v\n", info.Stats.V.LatestHandshake.Time.Format("_2 Jan 2006 15:04:05 MST"))
			}
		}
	}
	_, _ = ctx.session.Write(b.Bytes())

	return nil
}

// Borrowed from here: https://yourbasic.org/golang/formatting-byte-size-to-human-readable-format.
func humanReadableByteCount(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB",
		float64(b)/float64(div), "KMGTPE"[exp])
}
