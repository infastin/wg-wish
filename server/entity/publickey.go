package entity

import "github.com/charmbracelet/ssh"

type PublicKey struct {
	Key     ssh.PublicKey
	Comment string
}
