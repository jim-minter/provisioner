package netboot

import (
	"fmt"
	"net"
	"slices"

	"provisioner/pkg/netboot/dhcp"
	"provisioner/pkg/netboot/http"
	"provisioner/pkg/netboot/tftp"
)

type NetBoot struct {
	Interface string
	MAC       net.HardwareAddr
	Password  string
}

func (nb *NetBoot) Run() error {
	ipNet, err := interfaceIPNet(nb.Interface)
	if err != nil {
		return err
	}

	dhcps, err := dhcp.New(nb.Interface, nb.MAC, ipNet)
	if err != nil {
		return err
	}

	tftps, err := tftp.New(nb.MAC, ipNet.IP)
	if err != nil {
		return err
	}

	https, err := http.New(ipNet.IP, nb.Password)
	if err != nil {
		return err
	}

	errch := make(chan error, 3)

	go func() {
		errch <- dhcps.Serve()
	}()

	go func() {
		errch <- tftps.Serve()
	}()

	go func() {
		errch <- https.Serve()
	}()

	return <-errch
}

func interfaceIPNet(name string) (*net.IPNet, error) {
	iface, err := net.InterfaceByName(name)
	if err != nil {
		return nil, err
	}

	addrs, err := iface.Addrs()
	if err != nil {
		return nil, err
	}

	i := slices.IndexFunc(addrs, func(addr net.Addr) bool {
		_, ok := addr.(*net.IPNet)
		return ok
	})
	if i == -1 {
		return nil, fmt.Errorf("no IP address found on interface %q", name)
	}

	return addrs[i].(*net.IPNet), nil
}
