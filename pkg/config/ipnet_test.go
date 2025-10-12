package config

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/goccy/go-yaml"
)

func TestMaskAsIP(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	var config Config
	err := yaml.Unmarshal([]byte(`network: { ipnet: 192.168.0.0/24 }`), &config)
	g.Expect(err).NotTo(HaveOccurred())

	b, err := config.Template("{{ .Network.IPNet.MaskAsIP }}")
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(b).To(BeEquivalentTo("255.255.255.0"))
}
