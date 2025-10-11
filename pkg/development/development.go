package development

import (
	"log"
	"net"
	"os"
	"strings"
)

var Config struct {
	DefaultGateway net.IP
	Nameservers    []net.IP

	SshAuthorizedKey string
}

func init() {
	if defaultGateway, ok := os.LookupEnv("DEVELOPMENT_DEFAULT_GATEWAY"); ok {
		if Config.DefaultGateway = net.ParseIP(defaultGateway); Config.DefaultGateway == nil {
			panic("invalid DEVELOPMENT_DEFAULT_GATEWAY")
		}
		log.Printf("DEVELOPMENT_DEFAULT_GATEWAY: %q", Config.DefaultGateway)
	}

	if nameservers, ok := os.LookupEnv("DEVELOPMENT_NAMESERVERS"); ok {
		for nameserver := range strings.SplitSeq(nameservers, ":") {
			ns := net.ParseIP(nameserver)
			if ns == nil {
				panic("invalid DEVELOPMENT_NAMESERVERS")
			}
			Config.Nameservers = append(Config.Nameservers, ns)
		}
		log.Printf("DEVELOPMENT_NAMESERVERS: %q", Config.Nameservers)
	}

	if sshAuthorizedKey, ok := os.LookupEnv("DEVELOPMENT_SSH_AUTHORIZED_KEY"); ok {
		Config.SshAuthorizedKey = sshAuthorizedKey
		log.Printf("DEVELOPMENT_SSH_AUTHORIZED_KEY: %q", Config.SshAuthorizedKey)
	}
}
