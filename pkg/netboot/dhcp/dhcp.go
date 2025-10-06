package dhcp

import (
	"bytes"
	"log"
	"net"
	"slices"
	"time"

	"github.com/insomniacslk/dhcp/dhcpv4"
	"github.com/insomniacslk/dhcp/dhcpv4/server4"
)

type DHCPServer struct {
	iface string
	mac   net.HardwareAddr
	ipNet *net.IPNet

	server *server4.Server
}

func New(iface string, mac net.HardwareAddr, ipNet *net.IPNet) (*DHCPServer, error) {
	ds := &DHCPServer{
		iface: iface,
		mac:   mac,
		ipNet: ipNet,
	}

	var err error
	ds.server, err = server4.NewServer(iface, nil, ds.handle)
	if err != nil {
		return nil, err
	}

	return ds, nil
}

func (ds *DHCPServer) Serve() error {
	return ds.server.Serve()
}

func (ds *DHCPServer) handle(conn net.PacketConn, peer net.Addr, m *dhcpv4.DHCPv4) {
	if !bytes.Equal(m.ClientHWAddr, ds.mac) {
		return
	}

	yourIP := incrementIP(ds.ipNet.IP)

	resp, err := dhcpv4.NewReplyFromRequest(m,
		dhcpv4.WithYourIP(yourIP),
		dhcpv4.WithServerIP(ds.ipNet.IP),
		dhcpv4.WithOption(dhcpv4.OptServerIdentifier(ds.ipNet.IP)),
		dhcpv4.WithOption(dhcpv4.OptIPAddressLeaseTime(time.Hour)),
		dhcpv4.WithOption(dhcpv4.OptSubnetMask(ds.ipNet.Mask)),
		dhcpv4.WithOption(dhcpv4.OptRouter(ds.ipNet.IP)), // https://bugs.launchpad.net/subiquity/+bug/2079222
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
		resp.UpdateOption(dhcpv4.OptBootFileName("/amd64/pxelinux.0"))
	}

	_, err = conn.WriteTo(resp.ToBytes(), peer)
	if err != nil {
		log.Print(err)
		return
	}
}

func incrementIP(ip net.IP) net.IP {
	ip = slices.Clone(ip.To4())

	for i := len(ip) - 1; i >= 0; i-- {
		ip[i]++
		if ip[i] != 0 {
			break
		}
	}

	return ip
}
