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
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type HTTPServer struct {
	client client.Client

	mux *http.ServeMux
}

func New(client client.Client) *HTTPServer {
	hs := &HTTPServer{
		client: client,

		mux: &http.ServeMux{},
	}

	hs.mux.Handle("GET /meta-data", file(nil))
	hs.mux.Handle("GET /vendor-data", file(nil))
	hs.mux.HandleFunc("GET /user-data", hs.userData)

	return hs
}

func (hs *HTTPServer) Serve() error {
	return http.ListenAndServe(":80", &logger{hs.mux})
}

func (hs *HTTPServer) userData(w http.ResponseWriter, r *http.Request) {
	token, err := hs.bootstrapToken(r.Context())
	if err != nil {
		log.Print(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		log.Print(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	userData := map[string]any{
		// TODO: hard-coded public key
		"ssh_authorized_keys": []any{
			"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCyJkNEKCvLD66gesir0+Z2l6Nrq1j5id3L+Cttrgs7Xv9FU6Gaare6yAUkBq5gkKNeS0PJHZDDX9HwCX4Yghy/cbnBeXuzPppDkR6wIHKzUOKuYZfFW76Gl4he8ZNJGpn6QEoY4uKAxRpRg/elwvKfyuJ6Iw5F/fimSI7LyqRsj/CtnGLZ7PPpxdVPSSEEoXPaE2aWn8L67t6kT7iKqexRRAcyl6271/U3yIp4I0ZbpJnDSgqtwLZ/L93AEa/w4L7kykvuxtEd6vIbEULUy4BWCGiQi2ENJOdPkbkQuq6ugbGxJ5Dwvkxa4Nvz3c4VVKLdXZIByC/RU5PjF/KDXc5G6i1CbhNNTLa8pCFDWA1xOWzdjTK0UkGcRsbT+GL3Is1/ZQlqvl1sLKqsYpN8HQxOrHO4flIsMb+KTe722P/Nhj+i65dBb546GNuDx4JNNCmOjNc9XQwZm1llou7CXNuhvmCJB7ee6pJ4VTGM+1xTSqU6x6qQ730k/Z7AtPy2Na8= t540p@t540p",
		},
		// TODO: hard-coded IP
		"runcmd": []any{
			"kubeadm join 192.168.123.2:6443 --token " + token + " --discovery-token-ca-cert-hash " + caCertHash,
		},
		// TODO: based on custom resource
		"hostname": "node-" + strings.ReplaceAll(host, ".", "-"),
	}

	b, err := json.Marshal(userData)
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
