package repo

type AvailablePackages map[string]*Package

func (availablePackages AvailablePackages) AddPackages(pkgs []*Package) {
	for _, pkg := range pkgs {
		if pkg.Architecture != "all" && pkg.Architecture != pkg.Repo.Arch {
			continue
		}

		if availablePackages[pkg.Package] == nil ||
			availablePackages[pkg.Package].Version.Compare(pkg.Version) == -1 {
			availablePackages[pkg.Package] = pkg
		}
	}
}
