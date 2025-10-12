package http

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"log"
	"math/big"
	"net"
	"net/http"
	"os"
	"time"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"provisioner/pkg/cache"
	"provisioner/pkg/config"
	"provisioner/pkg/httputil"
)

type HTTPServer struct {
	config *config.Config
	client client.Client
	cache  *cache.Cache
	ip     net.IP

	mux *http.ServeMux
}

func New(config *config.Config, client client.Client, cache *cache.Cache, ip net.IP) *HTTPServer {
	hs := &HTTPServer{
		config: config,
		client: client,
		cache:  cache,
		ip:     ip,

		mux: &http.ServeMux{},
	}

	hs.mux.Handle("GET /meta-data", httputil.File(nil))
	hs.mux.Handle("GET /vendor-data", httputil.File(nil))
	hs.mux.HandleFunc("GET /user-data", hs.userData)

	return hs
}

func (hs *HTTPServer) Serve() error {
	return http.ListenAndServe(":80", &httputil.Logger{Handler: hs.mux})
}

func (hs *HTTPServer) userData(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	remoteIP, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		log.Print(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	machine, err := hs.cache.Get(ctx, cache.ByIP, remoteIP)
	if err != nil {
		log.Print(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	token, err := hs.bootstrapToken(ctx)
	if err != nil {
		log.Print(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	userData := map[string]any{
		"runcmd": []any{
			"kubeadm join " + hs.config.Laptop.IP.String() + ":6443 --token " + token + " --discovery-token-ca-cert-hash " + caCertHash,
		},
		"fqdn":                      machine.Name,
		"prefer_fqdn_over_hostname": true,
	}

	if len(hs.config.AuthorizedKeys) > 0 {
		userData["ssh_authorized_keys"] = hs.config.AuthorizedKeys
	}

	b, err := json.MarshalIndent(userData, "", "  ")
	if err != nil {
		log.Print(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Write(append([]byte("#cloud-config\n"), b...))
}

// +kubebuilder:rbac:namespace=kube-system,groups="",resources=secrets,verbs=create

func (hs *HTTPServer) bootstrapToken(ctx context.Context) (string, error) {
	tokenID, err := randomString(6)
	if err != nil {
		return "", err
	}

	tokenSecret, err := randomString(16)
	if err != nil {
		return "", err
	}

	secret := &corev1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Namespace: "kube-system",
			Name:      "bootstrap-token-" + tokenID,
		},
		Data: map[string][]byte{
			"auth-extra-groups":              []byte("system:bootstrappers:kubeadm:default-node-token"),
			"expiration":                     []byte(time.Now().UTC().Add(time.Hour).Format(time.RFC3339)),
			"token-id":                       []byte(tokenID),
			"token-secret":                   []byte(tokenSecret),
			"usage-bootstrap-authentication": []byte("true"),
			"usage-bootstrap-signing":        []byte("true"),
		},
		Type: corev1.SecretTypeBootstrapToken,
	}

	if err = hs.client.Create(ctx, secret); err != nil {
		return "", err
	}

	return tokenID + "." + tokenSecret, err
}

func randomString(n int) (string, error) {
	const letters = "0123456789abcdefghijklmnopqrstuvwxyz"

	b := make([]byte, 0, n)

	for range n {
		r, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}

		b = append(b, letters[r.Int64()])
	}

	return string(b), nil
}

var caCertHash = must(getCACertHash())

func getCACertHash() (string, error) {
	b, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/ca.crt")
	if err != nil {
		return "", err
	}

	var p *pem.Block
	for {
		p, b = pem.Decode(b)
		if p == nil {
			return "", errors.New("certificate not found")
		}
		if p.Type == "CERTIFICATE" {
			break
		}
	}

	c, err := x509.ParseCertificate(p.Bytes)
	if err != nil {
		return "", err
	}

	b, err = x509.MarshalPKIXPublicKey(c.PublicKey.(*rsa.PublicKey))
	if err != nil {
		return "", err
	}

	h := sha256.New()
	h.Write(b)

	return "sha256:" + hex.EncodeToString(h.Sum(nil)), nil
}

func must[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}

	return t
}
