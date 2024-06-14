package netutils

import (
	"bytes"
	"encoding/binary"
	"errors"
	"math/bits"
	"net"
	"strings"
)

var ErrIPAddressOverflow = errors.New("ip address overflow")

var v4InV6Prefix = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0xff, 0xff}

func allFF(b []byte) bool {
	for _, c := range b {
		if c != 0xff {
			return false
		}
	}
	return true
}

func LastIP(addr net.IPNet) net.IP {
	ip := addr.IP
	mask := addr.Mask

	if len(mask) == net.IPv6len && len(ip) == net.IPv4len && allFF(mask[:12]) {
		mask = mask[12:]
	}

	if len(mask) == net.IPv4len && len(ip) == net.IPv6len && bytes.Equal(ip[:12], v4InV6Prefix) {
		ip = ip[12:]
	}

	n := len(ip)
	if n != len(mask) {
		return nil
	}

	out := make(net.IP, n)
	for i := 0; i < n; i++ {
		out[i] = ip[i] | ^mask[i]
	}

	return out
}

func incrementIP(ip net.IP, inc uint8) (result net.IP, overflow bool) {
	switch len(ip) {
	case net.IPv4len:
		return incrementIPv4(ip, inc)
	case net.IPv6len:
		return incrementIPv6(ip, inc)
	}
	return nil, false
}

func incrementIPv4(ip net.IP, inc uint8) (result net.IP, overflow bool) {
	num := binary.BigEndian.Uint32(ip)

	num, carry := bits.Add32(num, uint32(inc), 0)
	if carry != 0 {
		return nil, true
	}

	result = make(net.IP, 0, net.IPv4len)
	result = binary.BigEndian.AppendUint32(result, num)

	return result, false
}

func incrementIPv6(ip net.IP, inc uint8) (result net.IP, overflow bool) {
	num1 := binary.BigEndian.Uint64(ip[:8])
	num2 := binary.BigEndian.Uint64(ip[8:])

	num2, carry := bits.Add64(num2, uint64(inc), 0)
	num1, carry = bits.Add64(num1, carry, 0)

	if carry != 0 {
		return nil, true
	}

	result = make(net.IP, 0, net.IPv6len)
	result = binary.BigEndian.AppendUint64(result, num1)
	result = binary.BigEndian.AppendUint64(result, num2)

	return result, false
}

func NextAddress(addr net.IPNet) (next net.IPNet, err error) {
	lastIP := LastIP(addr)

	inc := uint8(1)
	if !allFF(addr.Mask) {
		inc += 1
	}

	nextIP, overflow := incrementIP(lastIP, inc)
	if overflow {
		return net.IPNet{}, ErrIPAddressOverflow
	}

	return net.IPNet{
		IP:   nextIP,
		Mask: addr.Mask,
	}, nil
}

func FormatAddresses(s []net.IPNet, sep string) string {
	var builder strings.Builder

	for i := range s {
		if i != 0 {
			builder.WriteString(sep)
		}
		builder.WriteString(s[i].String())
	}

	return builder.String()
}

func ParseAddress(in string) (out net.IPNet, err error) {
	ip, ipNet, err := net.ParseCIDR(in)
	if err != nil {
		return net.IPNet{}, err
	}

	return net.IPNet{
		IP:   ip,
		Mask: ipNet.Mask,
	}, nil
}

func ParseAddresses(in []string) (out []net.IPNet, err error) {
	out = make([]net.IPNet, len(in))

	for i := range in {
		out[i], err = ParseAddress(in[i])
		if err != nil {
			return nil, err
		}
	}

	return out, nil
}

func FormatIPs(s []net.IP, sep string) string {
	var builder strings.Builder

	for i := range s {
		if i != 0 {
			builder.WriteString(sep)
		}
		builder.WriteString(s[i].String())
	}

	return builder.String()
}

func ParseIPs(in []string) (out []net.IP, err error) {
	out = make([]net.IP, len(in))

	for i := range in {
		ip := net.ParseIP(in[i])
		if ip == nil {
			return nil, &net.ParseError{
				Type: "IP address",
				Text: in[i],
			}
		}

		out[i] = ip
	}

	return out, nil
}
