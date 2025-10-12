package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"provisioner/pkg/config"
	"provisioner/pkg/httputil"
)

type server struct {
	config *config.Config
	mux    *http.ServeMux
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	config, err := config.Load()
	if err != nil {
		return err
	}

	s := &server{
		config: config,
		mux:    &http.ServeMux{},
	}

	s.mux.Handle("GET /meta-data", httputil.File(nil))
	s.mux.Handle("GET /vendor-data", httputil.File(nil))
	s.mux.HandleFunc("GET /user-data", s.userData)

	s.mux.Handle("GET /", http.FileServer(http.Dir("hack/image-builder/images/capi/output")))

	return http.ListenAndServe(":8000", &httputil.Logger{Handler: s.mux})
}

func (s *server) userData(w http.ResponseWriter, r *http.Request) {
	userData := map[string]any{
		"runcmd": []any{
			// TODO: should go into a script baked into the image
			"kubeadm init --kubernetes-version 1.32.4 --pod-network-cidr=10.244.0.0/16",
			"KUBECONFIG=/etc/kubernetes/admin.conf kubectl apply -f https://github.com/flannel-io/flannel/releases/latest/download/kube-flannel.yml",
		},
		"fqdn": "laptop",
	}

	if len(s.config.AuthorizedKeys) > 0 {
		userData["ssh_authorized_keys"] = s.config.AuthorizedKeys
	}

	b, err := json.MarshalIndent(userData, "", "  ")
	if err != nil {
		log.Print(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Write(append([]byte("#cloud-config\n"), b...))
}
