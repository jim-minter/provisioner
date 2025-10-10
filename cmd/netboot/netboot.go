package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"provisioner/api/v1alpha1"
	"provisioner/pkg/cache"
	"provisioner/pkg/dhcp"
	"provisioner/pkg/http"
	"provisioner/pkg/tftp"
)

func main() {
	if err := run(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	// TODO: remove this hack
	for {
		l, err := net.Listen("tcp", ":80")
		if err == nil {
			l.Close()
			break
		}

		log.Print("waiting")
		time.Sleep(time.Second)
	}

	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(clientcmd.NewDefaultClientConfigLoadingRules(), nil).ClientConfig()
	if err != nil {
		return err
	}

	if err = v1alpha1.AddToScheme(scheme.Scheme); err != nil {
		return err
	}

	client, err := client.New(config, client.Options{})
	if err != nil {
		return err
	}

	cache, err := cache.New(ctx, config)
	if err != nil {
		return err
	}

	iface, ipNet, err := getIPv4() // TODO: IPv6
	if err != nil {
		return err
	}

	dhcps, err := dhcp.New(cache, iface, ipNet)
	if err != nil {
		return err
	}

	tftps := tftp.New(cache, ipNet.IP)
	https := http.New(client, cache, ipNet.IP)

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

func getIPv4() (string, *net.IPNet, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", nil, err
	}

	for _, iface := range ifaces {
		if !(strings.HasPrefix(iface.Name, "en") || strings.HasPrefix(iface.Name, "eth")) {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			return "", nil, err
		}

		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if ok && ipNet.IP.To4() != nil {
				return iface.Name, ipNet, nil
			}

		}
	}

	return "", nil, fmt.Errorf("no IPv4 address found")
}
