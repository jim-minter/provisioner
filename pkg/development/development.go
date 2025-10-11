package development

import (
	"log"
	"net"
	"os"
)

var Config struct {
	DefaultGateway net.IP
	Nameserver     net.IP

	SshAuthorizedKey string
}

func init() {
	if defaultGateway := os.Getenv("DEVELOPMENT_DEFAULT_GATEWAY"); defaultGateway != "" {
		if Config.DefaultGateway = net.ParseIP(defaultGateway); Config.DefaultGateway == nil {
			panic("invalid DEVELOPMENT_DEFAULT_GATEWAY")
		}
		log.Printf("DEVELOPMENT_DEFAULT_GATEWAY: %q", Config.DefaultGateway)
	}

	if nameserver := os.Getenv("DEVELOPMENT_NAMESERVER"); nameserver != "" {
		if Config.Nameserver = net.ParseIP(nameserver); Config.Nameserver == nil {
			panic("invalid DEVELOPMENT_NAMESERVER")
		}
		log.Printf("DEVELOPMENT_NAMESERVER: %q", Config.Nameserver)
	}

	if sshAuthorizedKey := os.Getenv("DEVELOPMENT_SSH_AUTHORIZED_KEY"); sshAuthorizedKey != "" {
		Config.SshAuthorizedKey = sshAuthorizedKey
		log.Printf("DEVELOPMENT_SSH_AUTHORIZED_KEY: %q", Config.SshAuthorizedKey)
	}
}
