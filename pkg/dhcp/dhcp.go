package dhcp

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/insomniacslk/dhcp/dhcpv4"
	"github.com/insomniacslk/dhcp/dhcpv4/server4"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"provisioner/api/v1alpha1"
)

type DHCPServer struct {
	cache cache.Cache
	ipNet *net.IPNet

	server *server4.Server
}

func New(cache cache.Cache, iface string, ipNet *net.IPNet) (_ *DHCPServer, err error) {
	ds := &DHCPServer{
		cache: cache,
		ipNet: ipNet,
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

	machines := &v1alpha1.MachineList{}
	err := ds.cache.List(ctx, machines, client.MatchingFields{"spec.macAddress": m.ClientHWAddr.String()}, client.Limit(2))
	if err != nil {
		log.Print(err)
		return
	}

	if len(machines.Items) != 1 {
		log.Printf("%d items found for mac %q", len(machines.Items), m.ClientHWAddr)
		return
	}

	yourIP := net.ParseIP(machines.Items[0].Spec.IPAddress)
	if yourIP == nil {
		log.Printf("invalid IP address %q", machines.Items[0].Spec.IPAddress)
		return
	}

	resp, err := dhcpv4.NewReplyFromRequest(m,
		dhcpv4.WithYourIP(yourIP),
		dhcpv4.WithServerIP(ds.ipNet.IP),
		dhcpv4.WithOption(dhcpv4.OptServerIdentifier(ds.ipNet.IP)),
		dhcpv4.WithOption(dhcpv4.OptIPAddressLeaseTime(time.Hour)),
		dhcpv4.WithOption(dhcpv4.OptSubnetMask(ds.ipNet.Mask)),
		dhcpv4.WithOption(dhcpv4.OptRouter(net.IPv4(192, 168, 123, 1))), // TODO: hard-coded IP; should remove in prod (disconnected) mode
		dhcpv4.WithOption(dhcpv4.OptDNS(net.IPv4(8, 8, 8, 8))),          // TODO: hard-coded IP; should remove in prod (disconnected) mode
	)
	if err != nil {
		log.Print(err)
		return
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
