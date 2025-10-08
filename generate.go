package provisioner

//go:generate go run sigs.k8s.io/controller-tools/cmd/controller-gen@v0.19.0 paths=./... object
//go:generate go run sigs.k8s.io/controller-tools/cmd/controller-gen@v0.19.0 paths=./... crd rbac:roleName=netboot output:dir=manifests
