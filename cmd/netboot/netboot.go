package main

import (
	"flag"
	"fmt"
	"net"
	"os"

	"provisioner/pkg/netboot"
)

var (
	iface    string
	mac      string
	password string
)

func init() {
	flag.StringVar(&iface, "interface", "", "interface to bind to, e.g. eth0, virbr0, etc.")
	flag.StringVar(&mac, "mac", "", "client MAC, e.g. 11:22:33:44:55:66")
	flag.StringVar(&password, "password", "", "password of ubuntu user")
}

func main() {
	flag.Parse()

	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	if password == "" {
		return fmt.Errorf("password must be set")
	}

	mac, err := net.ParseMAC(mac)
	if err != nil {
		return err
	}

	nb := &netboot.NetBoot{
		Interface: iface,
		MAC:       mac,
		Password:  password,
	}

	return nb.Run()
}
