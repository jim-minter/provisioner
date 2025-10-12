package dhcp

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/insomniacslk/dhcp/dhcpv4"
	"github.com/insomniacslk/dhcp/dhcpv4/server4"

	"provisioner/pkg/cache"
	"provisioner/pkg/config"
)

type DHCPServer struct {
	config *config.Config
	cache  *cache.Cache
	ipNet  *net.IPNet

	server *server4.Server
}

func New(config *config.Config, cache *cache.Cache, iface string, ipNet *net.IPNet) (_ *DHCPServer, err error) {
	ds := &DHCPServer{
		config: config,
		cache:  cache,
		ipNet:  ipNet,
	}

	ds.server, err = server4.NewServer(iface, nil, ds.handle)
	if err != nil {
		return nil, err
	}

	return ds, nil
}

func (ds *DHCPServer) Serve() error {
	return ds.server.Serve()
}

// +kubebuilder:rbac:groups=dummy.group,resources=machines,verbs=list;watch

func (ds *DHCPServer) handle(conn net.PacketConn, peer net.Addr, m *dhcpv4.DHCPv4) {
	ctx := context.Background()

	machine, err := ds.cache.Get(ctx, cache.ByMAC, m.ClientHWAddr.String())
	if err != nil {
		log.Print(err)
		return
	}

	yourIP := net.ParseIP(machine.Spec.IPAddress)
	if yourIP == nil {
		log.Printf("invalid IP address %q", machine.Spec.IPAddress)
		return
	}

	resp, err := dhcpv4.NewReplyFromRequest(m,
		dhcpv4.WithYourIP(yourIP),
		dhcpv4.WithServerIP(ds.ipNet.IP),
		dhcpv4.WithOption(dhcpv4.OptServerIdentifier(ds.ipNet.IP)),
		dhcpv4.WithOption(dhcpv4.OptIPAddressLeaseTime(time.Hour)),
		dhcpv4.WithOption(dhcpv4.OptSubnetMask(ds.ipNet.Mask)),
	)
	if err != nil {
		log.Print(err)
		return
	}

	if len(ds.config.Network.Gateway) > 0 {
		resp.UpdateOption(dhcpv4.OptRouter(ds.config.Network.Gateway))
	}

	if len(ds.config.Network.Nameserver) > 0 {
		resp.UpdateOption(dhcpv4.OptDNS(ds.config.Network.Nameserver))
	}

	switch mt := m.MessageType(); mt {
	case dhcpv4.MessageTypeDiscover:
		resp.UpdateOption(dhcpv4.OptMessageType(dhcpv4.MessageTypeOffer))
	case dhcpv4.MessageTypeRequest:
		resp.UpdateOption(dhcpv4.OptMessageType(dhcpv4.MessageTypeAck))
	}

	if m.IsOptionRequested(dhcpv4.OptionBootfileName) {
		resp.UpdateOption(dhcpv4.OptBootFileName("amd64/pxelinux.0")) // TODO: based on custom resource
	}

	_, err = conn.WriteTo(resp.ToBytes(), peer)
	if err != nil {
		log.Print(err)
		return
	}
}
