package main

import (
	"fmt"
	"net"
	"os"
	"slices"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"provisioner/api/v1alpha1"
	"provisioner/pkg/dhcp"
	"provisioner/pkg/http"
	"provisioner/pkg/tftp"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(clientcmd.NewDefaultClientConfigLoadingRules(), nil).ClientConfig()
	if err != nil {
		return err
	}

	var scheme = runtime.NewScheme()
	if err = v1alpha1.AddToScheme(scheme); err != nil {
		return err
	}

	cli, err := client.New(config, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		return err
	}

	ipNet, err := interfaceIPNetv4("eth0")
	if err != nil {
		return err
	}

	dhcps, err := dhcp.New(cli, ipNet)
	if err != nil {
		return err
	}

	tftps := tftp.New()
	https := http.New()

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

func interfaceIPNetv4(name string) (*net.IPNet, error) {
	iface, err := net.InterfaceByName(name)
	if err != nil {
		return nil, err
	}

	addrs, err := iface.Addrs()
	if err != nil {
		return nil, err
	}

	i := slices.IndexFunc(addrs, func(addr net.Addr) bool {
		ipNet, ok := addr.(*net.IPNet)
		return ok && ipNet.IP.To4() != nil
	})
	if i == -1 {
		return nil, fmt.Errorf("no IPv4 address found on interface %q", name)
	}

	return addrs[i].(*net.IPNet), nil
}
