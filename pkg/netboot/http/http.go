package http

import (
	"encoding/json"
	"net"
	"net/http"
	"os"
	"provisioner/pkg/cache"
	"provisioner/pkg/repo/crio"
	"provisioner/pkg/repo/kubernetes"

	"github.com/go-crypt/crypt/algorithm/shacrypt"
)

type HTTPServer struct {
	ip       string
	password string

	mux *http.ServeMux
}

func New(ip net.IP, password string) (*HTTPServer, error) {
	hs := &HTTPServer{
		ip:       ip.String(),
		password: password,

		mux: &http.ServeMux{},
	}

	userData, err := hs.cloudConfig()
	if err != nil {
		return nil, err
	}

	hs.mux.Handle("/", http.FileServer(http.Dir("cache")))

	hs.mux.Handle("/meta-data", file(nil))
	hs.mux.Handle("/vendor-data", file(nil))
	hs.mux.Handle("/user-data", file(userData))

	return hs, nil
}

func (hs *HTTPServer) Serve() error {
	return http.ListenAndServe(net.JoinHostPort(hs.ip, "80"), &logger{hs.mux})
}

func (hs *HTTPServer) cloudConfig() ([]byte, error) {
	password, err := hash(hs.password)
	if err != nil {
		return nil, err
	}

	crioGPGPublicKeyLocalPath, err := cache.LocalPath(crio.GPGPublicKeyURL)
	if err != nil {
		return nil, err
	}

	crioGPGPublicKey, err := os.ReadFile(crioGPGPublicKeyLocalPath)
	if err != nil {
		return nil, err
	}

	kubernetesGPGPublicKeyLocalPath, err := cache.LocalPath(kubernetes.GPGPublicKeyURL)
	if err != nil {
		return nil, err
	}

	kubernetesGPGPublicKey, err := os.ReadFile(kubernetesGPGPublicKeyLocalPath)
	if err != nil {
		return nil, err
	}

	b, err := json.Marshal(map[string]any{
		"autoinstall": map[string]any{
			"version": 1,
			"apt": map[string]any{
				"preserve_sources_list": false,
				"disable_suites":        []any{"backports", "proposed"},
				"disable_components":    []any{"restricted", "universe", "multiverse"},
				"primary": []any{
					map[string]any{
						"uri":    "http://" + hs.ip + "/archive.ubuntu.com/ubuntu",
						"arches": []any{"default"},
					},
				},
				"security": []any{
					map[string]any{
						"uri":    "http://" + hs.ip + "/archive.ubuntu.com/ubuntu",
						"arches": []any{"default"},
					},
				},
				"sources": map[string]any{
					"crio.list": map[string]any{
						"source": "deb http://" + hs.ip + "/download.opensuse.org/repositories/isv:/cri-o:/stable:/v1.34/deb/ /",
						"key":    string(crioGPGPublicKey),
					},
					"kubernetes.list": map[string]any{
						"source": "deb http://" + hs.ip + "/pkgs.k8s.io/core:/stable:/v1.34/deb/ /",
						"key":    string(kubernetesGPGPublicKey),
					},
				},
			},
			"identity": map[string]any{
				"hostname": "ubuntu-server",
				"username": "ubuntu",
				"password": password,
			},
			"kernel": map[string]any{
				"flavor": "generic",
			},
			"packages": []any{
				"cri-o",
				"kubelet",
				"kubeadm",
				"kubectl",
			},
			"ssh": map[string]any{
				"install-server": true,
			},
			"updates": "all",
		},
	})
	if err != nil {
		return nil, err
	}

	return append([]byte("#cloud-config\n"), b...), nil
}

func hash(password string) (string, error) {
	hasher, err := shacrypt.NewSHA256()
	if err != nil {
		return "", err
	}

	digest, err := hasher.Hash(password)
	if err != nil {
		return "", err
	}

	return digest.Encode(), nil
}
