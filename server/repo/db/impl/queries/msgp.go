package queries

import (
	"net"
)

//go:generate msgp -tests=false -unexported

//msgp:tuple msgpIPNet
//msgp:replace net.IP with:[]byte
//msgp:replace net.IPMask with:[]byte

type msgpIPNet struct {
	IP   net.IP
	Mask net.IPMask
}
