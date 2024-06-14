package ssh

import (
	"bytes"
	"fmt"

	"github.com/charmbracelet/ssh"
	"github.com/infastin/wg-wish/pkg/fastconv"
	"github.com/infastin/wg-wish/server/entity"
	gossh "golang.org/x/crypto/ssh"
)

type PublicKeyCmd struct {
	Add struct {
		Key string `arg:"" help:"Public key to be added."`
	} `cmd:"" help:"Add public key."`

	Rm struct {
		Key string `arg:"" help:"Public key to be removed."`
	} `cmd:"" help:"Remove public key."`

	Ls struct{} `cmd:"" help:"List public keys."`
}

func (cmd *PublicKeyCmd) Run(ctx *Context) (err error) {
	switch ctx.kctx.Command() {
	case "publickey add <key>":
		err = cmd.HandleAdd(ctx)
	case "publickey rm <key>":
		err = cmd.HandleRm(ctx)
	case "publickey ls":
		err = cmd.HandleLs(ctx)
	}
	return err
}

func (cmd *PublicKeyCmd) HandleAdd(ctx *Context) (err error) {
	pkey, comment, _, _, err := ssh.ParseAuthorizedKey(fastconv.Bytes(cmd.Add.Key))
	if err != nil {
		return err
	}

	return ctx.publicKeyService.AddPublicKey(ctx, &entity.PublicKey{
		Key:     pkey,
		Comment: comment,
	})
}

func (cmd *PublicKeyCmd) HandleRm(ctx *Context) (err error) {
	pkey, _, _, _, err := ssh.ParseAuthorizedKey(fastconv.Bytes(cmd.Rm.Key))
	if err != nil {
		return err
	}
	return ctx.publicKeyService.RemovePublicKey(ctx, pkey)
}

func (*PublicKeyCmd) HandleLs(ctx *Context) (err error) {
	pkeys, err := ctx.publicKeyService.GetPublicKeys(ctx)
	if err != nil {
		return err
	}

	var b bytes.Buffer
	for i := range pkeys {
		name := pkeys[i].Comment
		if name == "" {
			name = "<empty>"
		}

		fmt.Fprintf(&b, "%d. %s\n", i+1, name)
		b.Write(gossh.MarshalAuthorizedKey(pkeys[i].Key))
	}
	_, _ = ctx.session.Write(b.Bytes())

	return nil
}
