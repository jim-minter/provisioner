package config

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

type IPNet struct {
	net.IPNet
}

func (ipNet *IPNet) UnmarshalYAML(b []byte) error {
	s := strings.TrimSpace(string(b))

	parts := strings.Split(s, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid IPNet %q", s)
	}

	ones, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return err
	}

	ipNet.IP = net.ParseIP(parts[0])
	ipNet.Mask = net.CIDRMask(int(ones), 32)
	return nil
}

func (ipNet IPNet) MaskAsIP() net.IP {
	return net.IP(ipNet.Mask)
}
