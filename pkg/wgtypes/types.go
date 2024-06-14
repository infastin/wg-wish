package wgtypes

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"

	"github.com/guregu/null/v5"
	"github.com/infastin/wg-wish/pkg/netutils"
	"golang.org/x/crypto/curve25519"
	"gopkg.in/ini.v1"
)

const KeyLen = 32

type Key [KeyLen]byte

func GeneratePrivateKey() (key Key, err error) {
	if _, err := rand.Read(key[:]); err != nil {
		return Key{}, fmt.Errorf("couldn't generate private key: %w", err)
	}

	key[0] &= 248
	key[31] &= 127
	key[31] |= 64

	return key, nil
}

func ParseKey(s string) (key Key, err error) {
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return Key{}, fmt.Errorf("failed to parse base64-encoded key: %w", err)
	}

	if len(b) != KeyLen {
		return Key{}, fmt.Errorf("incorrect key size: %d", len(b))
	}

	return Key(b), nil
}

func (key Key) PublicKey() (pub Key) {
	curve25519.ScalarBaseMult((*[KeyLen]byte)(&pub), (*[KeyLen]byte)(&key))
	return pub
}

func (key Key) String() string {
	return base64.StdEncoding.EncodeToString(key[:])
}

type ServerConfig struct {
	Interface ServerInterface
	Peers     []ServerPeer
}

type ServerConfigParams struct {
	PrivateKey Key
	Address    string
	Device     string
	ListenPort null.Int
}

func NewServerConfig(params *ServerConfigParams) (cfg ServerConfig, err error) {
	cfg.Interface.Address, err = netutils.ParseAddress(params.Address)
	if err != nil {
		return ServerConfig{}, err
	}

	cfg.Interface.ListenPort = params.ListenPort
	if !cfg.Interface.ListenPort.Valid {
		cfg.Interface.ListenPort = null.IntFrom(51820)
	}

	cfg.Interface.PrivateKey = params.PrivateKey

	device := params.Device
	if device == "" {
		device = "eth0"
	}

	subnet := net.IPNet{
		IP:   cfg.Interface.Address.IP.Mask(cfg.Interface.Address.Mask),
		Mask: cfg.Interface.Address.Mask,
	}

	cfg.Interface.PostUp = []string{
		fmt.Sprintf("iptables -t nat -A POSTROUTING -s %s -o %s -j MASQUERADE", subnet.String(), device),
		fmt.Sprintf("iptables -A INPUT -i %s -p udp -m udp --dport %d -j ACCEPT", device, cfg.Interface.ListenPort.Int64),
		fmt.Sprintf("iptables -A FORWARD -i wg0 -o %s -j ACCEPT", device),
		fmt.Sprintf("iptables -A FORWARD -i %s -o wg0 -j ACCEPT", device),
	}

	cfg.Interface.PostDown = []string{
		fmt.Sprintf("iptables -t nat -D POSTROUTING -s %s -o %s -j MASQUERADE", subnet.String(), device),
		fmt.Sprintf("iptables -D INPUT -i %s -p udp -m udp --dport %d -j ACCEPT", device, cfg.Interface.ListenPort.Int64),
		fmt.Sprintf("iptables -D FORWARD -i wg0 -o %s -j ACCEPT", device),
		fmt.Sprintf("iptables -D FORWARD -i %s -o wg0 -j ACCEPT", device),
	}

	return cfg, nil
}

func (cfg *ServerConfig) Encode(writer io.Writer) (err error) {
	file := ini.Empty(ini.LoadOptions{
		AllowShadows:           true,
		AllowNonUniqueSections: true,
	})

	ifaceSection, err := file.NewSection("Interface")
	if err != nil {
		return err
	}

	err = cfg.Interface.store(ifaceSection)
	if err != nil {
		return err
	}

	for i := range cfg.Peers {
		peerSection, err := file.NewSection("Peer")
		if err != nil {
			return err
		}

		err = cfg.Peers[i].store(peerSection)
		if err != nil {
			return err
		}
	}

	_, err = file.WriteTo(writer)
	if err != nil {
		return err
	}

	return nil
}

func (cfg *ServerConfig) Decode(reader io.Reader) (err error) {
	file, err := ini.LoadSources(ini.LoadOptions{
		AllowShadows:           true,
		AllowNonUniqueSections: true,
	}, reader)

	if err != nil {
		return err
	}

	ifaceSection, err := file.GetSection("Interface")
	if err != nil {
		return err
	}

	err = cfg.Interface.load(ifaceSection)
	if err != nil {
		return err
	}

	if file.HasSection("Peer") {
		peerSections, err := file.SectionsByName("Peer")
		if err != nil {
			return err
		}

		cfg.Peers = make([]ServerPeer, len(peerSections))
		for i := range cfg.Peers {
			err = cfg.Peers[i].load(peerSections[i])
			if err != nil {
				return err
			}
		}
	}

	return nil
}

type ServerInterface struct {
	Name       string
	Address    net.IPNet
	ListenPort null.Int
	PrivateKey Key
	PostUp     []string
	PostDown   []string
}

func (si *ServerInterface) store(section *ini.Section) (err error) {
	if si.Name != "" {
		section.Comment = "# " + si.Name
	}

	_, err = section.NewKey("Address", si.Address.String())
	if err != nil {
		return err
	}

	if si.ListenPort.Valid {
		_, err = section.NewKey("ListenPort", strconv.FormatInt(si.ListenPort.Int64, 10))
		if err != nil {
			return err
		}
	}

	_, err = section.NewKey("PrivateKey", si.PrivateKey.String())
	if err != nil {
		return err
	}

	if len(si.PostUp) != 0 {
		for _, cmd := range si.PostUp {
			_, err = section.NewKey("PostUp", cmd)
			if err != nil {
				return err
			}
		}
	}

	if len(si.PostDown) != 0 {
		for _, cmd := range si.PostDown {
			_, err = section.NewKey("PostDown", cmd)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (si *ServerInterface) load(section *ini.Section) (err error) {
	si.Name = strings.TrimLeft(section.Comment, "# ")

	addressKey, err := section.GetKey("Address")
	if err != nil {
		return err
	}

	si.Address, err = netutils.ParseAddress(addressKey.Value())
	if err != nil {
		return err
	}

	if section.HasKey("ListenPort") {
		portKey, err := section.GetKey("ListenPort")
		if err != nil {
			return err
		}

		port, err := portKey.Int64()
		if err != nil {
			return err
		}

		si.ListenPort = null.IntFrom(port)
	}

	pkKey, err := section.GetKey("PrivateKey")
	if err != nil {
		return err
	}

	si.PrivateKey, err = ParseKey(pkKey.Value())
	if err != nil {
		return err
	}

	if section.HasKey("PostUp") {
		postUpKey, err := section.GetKey("PostUp")
		if err != nil {
			return err
		}
		si.PostUp = postUpKey.ValueWithShadows()
	}

	if section.HasKey("PostDown") {
		postDownKey, err := section.GetKey("PostDown")
		if err != nil {
			return err
		}
		si.PostDown = postDownKey.ValueWithShadows()
	}

	return nil
}

type ServerPeer struct {
	Name       string
	PublicKey  Key
	AllowedIPs []net.IPNet
}

func (sp *ServerPeer) store(section *ini.Section) (err error) {
	if sp.Name != "" {
		section.Comment = "# " + sp.Name
	}

	_, err = section.NewKey("PublicKey", sp.PublicKey.String())
	if err != nil {
		return err
	}

	_, err = section.NewKey("AllowedIPs", netutils.FormatAddresses(sp.AllowedIPs, ","))
	if err != nil {
		return err
	}

	return nil
}

func (sp *ServerPeer) load(section *ini.Section) (err error) {
	sp.Name = strings.TrimLeft(section.Comment, "# ")

	pkKey, err := section.GetKey("PublicKey")
	if err != nil {
		return err
	}

	sp.PublicKey, err = ParseKey(pkKey.Value())
	if err != nil {
		return err
	}

	ipsKey, err := section.GetKey("AllowedIPs")
	if err != nil {
		return err
	}

	sp.AllowedIPs, err = netutils.ParseAddresses(ipsKey.Strings(","))
	if err != nil {
		return err
	}

	return nil
}

type ClientConfig struct {
	Interface ClientInterface
	Peer      ClientPeer
}

func (cfg *ClientConfig) Encode(writer io.Writer) (err error) {
	file := ini.Empty(ini.LoadOptions{
		AllowShadows: true,
	})

	ifaceSection, err := file.NewSection("Interface")
	if err != nil {
		return err
	}

	err = cfg.Interface.store(ifaceSection)
	if err != nil {
		return err
	}

	peerSection, err := file.NewSection("Peer")
	if err != nil {
		return err
	}

	err = cfg.Peer.store(peerSection)
	if err != nil {
		return err
	}

	_, err = file.WriteTo(writer)
	if err != nil {
		return err
	}

	return nil
}

func (cfg *ClientConfig) Decode(reader io.Reader) (err error) {
	file, err := ini.LoadSources(ini.LoadOptions{
		AllowShadows: true,
	}, reader)

	if err != nil {
		return err
	}

	ifaceSection, err := file.GetSection("Interface")
	if err != nil {
		return err
	}

	err = cfg.Interface.load(ifaceSection)
	if err != nil {
		return err
	}

	peerSection, err := file.GetSection("Peer")
	if err != nil {
		return err
	}

	err = cfg.Peer.load(peerSection)
	if err != nil {
		return err
	}

	return nil
}

type ClientInterface struct {
	Name       string
	Address    net.IPNet
	PrivateKey Key
	DNS        []net.IP
}

func (ci *ClientInterface) store(section *ini.Section) (err error) {
	if ci.Name != "" {
		section.Comment = "# " + ci.Name
	}

	_, err = section.NewKey("Address", ci.Address.String())
	if err != nil {
		return err
	}

	_, err = section.NewKey("PrivateKey", ci.PrivateKey.String())
	if err != nil {
		return err
	}

	if len(ci.DNS) != 0 {
		_, err = section.NewKey("DNS", netutils.FormatIPs(ci.DNS, ","))
		if err != nil {
			return err
		}
	}

	return nil
}

func (ci *ClientInterface) load(section *ini.Section) (err error) {
	ci.Name = strings.TrimLeft(section.Comment, "# ")

	addressKey, err := section.GetKey("Address")
	if err != nil {
		return err
	}

	ci.Address, err = netutils.ParseAddress(addressKey.Value())
	if err != nil {
		return err
	}

	pkKey, err := section.GetKey("PrivateKey")
	if err != nil {
		return err
	}

	ci.PrivateKey, err = ParseKey(pkKey.Value())
	if err != nil {
		return err
	}

	if section.HasKey("DNS") {
		dnsKey, err := section.GetKey("DNS")
		if err != nil {
			return err
		}

		ci.DNS, err = netutils.ParseIPs(dnsKey.Strings(","))
		if err != nil {
			return err
		}
	}

	return nil
}

type ClientPeer struct {
	Name                string
	EndpointHost        string
	EndpointPort        int
	PublicKey           Key
	AllowedIPs          []net.IPNet
	PersistentKeepalive null.Int
}

func (cp *ClientPeer) store(section *ini.Section) (err error) {
	if cp.Name != "" {
		section.Comment = "# " + cp.Name
	}

	_, err = section.NewKey("Endpoint", cp.EndpointHost+":"+strconv.Itoa(cp.EndpointPort))
	if err != nil {
		return err
	}

	_, err = section.NewKey("PublicKey", cp.PublicKey.String())
	if err != nil {
		return err
	}

	_, err = section.NewKey("AllowedIPs", netutils.FormatAddresses(cp.AllowedIPs, ","))
	if err != nil {
		return err
	}

	if cp.PersistentKeepalive.Valid {
		_, err = section.NewKey("PersistentKeepalive", strconv.FormatInt(cp.PersistentKeepalive.Int64, 10))
		if err != nil {
			return err
		}
	}

	return nil
}

func (cp *ClientPeer) load(section *ini.Section) (err error) {
	cp.Name = strings.TrimLeft(section.Comment, "# ")

	endpointKey, err := section.GetKey("Endpoint")
	if err != nil {
		return err
	}

	endpointHost, endpointPort, err := net.SplitHostPort(endpointKey.Value())
	if err != nil {
		return err
	}

	cp.EndpointHost = endpointHost

	cp.EndpointPort, err = strconv.Atoi(endpointPort)
	if err != nil {
		return err
	}

	pkKey, err := section.GetKey("PublicKey")
	if err != nil {
		return err
	}

	cp.PublicKey, err = ParseKey(pkKey.Value())
	if err != nil {
		return err
	}

	ipsKey, err := section.GetKey("AllowedIPs")
	if err != nil {
		return err
	}

	cp.AllowedIPs, err = netutils.ParseAddresses(ipsKey.Strings(","))
	if err != nil {
		return err
	}

	keepaliveKey, err := section.GetKey("PersistentKeepalive")
	if err != nil {
		return err
	}

	keepalive, err := keepaliveKey.Int64()
	if err != nil {
		return err
	}

	cp.PersistentKeepalive = null.IntFrom(keepalive)

	return nil
}
