package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"provisioner/pkg/development"
	"provisioner/pkg/httputil"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	mux := &http.ServeMux{}

	mux.Handle("GET /meta-data", httputil.File(nil))
	mux.Handle("GET /vendor-data", httputil.File(nil))
	mux.HandleFunc("GET /user-data", userData)

	mux.Handle("GET /", http.FileServer(http.Dir("hack/image-builder/images/capi/output")))

	return http.ListenAndServe(":8000", &httputil.Logger{Handler: mux})
}

func userData(w http.ResponseWriter, r *http.Request) {
	userData := map[string]any{
		"runcmd": []any{
			"kubeadm init --kubernetes-version 1.32.4 --pod-network-cidr=10.244.0.0/16",
			"KUBECONFIG=/etc/kubernetes/admin.conf kubectl apply -f https://github.com/flannel-io/flannel/releases/latest/download/kube-flannel.yml",
		},
		"fqdn": "laptop",
	}

	if development.Config.SshAuthorizedKey != "" {
		userData["ssh_authorized_keys"] = []any{
			development.Config.SshAuthorizedKey,
		}
	}

	b, err := json.Marshal(userData)
	if err != nil {
		log.Print(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Write(append([]byte("#cloud-config\n"), b...))
}
