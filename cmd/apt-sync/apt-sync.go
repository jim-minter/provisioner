package main

import (
	"fmt"
	"os"

	"provisioner/pkg/cache"
	"provisioner/pkg/repo"
	"provisioner/pkg/repo/docker"
	"provisioner/pkg/repo/ubuntu"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	// our upstream apt repos
	repos := []*repo.Repo{
		{
			Base:      "https://archive.ubuntu.com/ubuntu",
			Suite:     "noble",
			Component: "main",
			Arch:      "amd64",
		},
		{
			Base:      "https://archive.ubuntu.com/ubuntu",
			Suite:     "noble-updates",
			Component: "main",
			Arch:      "amd64",
		},
		{
			Base:      "https://archive.ubuntu.com/ubuntu",
			Suite:     "noble-security",
			Component: "main",
			Arch:      "amd64",
		},
		{
			Base:      "https://download.docker.com/linux/ubuntu",
			Suite:     "noble",
			Component: "stable",
			Arch:      "amd64",
		},
		{
			Base:  "https://pkgs.k8s.io/core:/stable:/v1.34/deb",
			Suite: "/",
			Arch:  "amd64",
		},
	}

	requiredPackages := repo.RequiredPackages{}
	availablePackages := repo.AvailablePackages{}

	for _, repo := range repos {
		// cache the repos' metadata
		if err := repo.SyncMetadata(); err != nil {
			return err
		}

		// read the repos' packages and add/update them in availablePackages
		pkgs, err := repo.Packages()
		if err != nil {
			return err
		}

		availablePackages.AddPackages(pkgs)
	}

	// include the default installed packages in our required set
	if err := requiredPackages.Add(availablePackages, ubuntu.NobleServerDefaultInstalledPackages()...); err != nil {
		return err
	}

	// include any additional packages wanted in our requried set
	if err := requiredPackages.Add(availablePackages, "docker-ce"); err != nil {
		return err
	}

	// calculate the closure
	if err := requiredPackages.AddDependencies(availablePackages); err != nil {
		return err
	}

	// download needed packages
	for _, pkg := range requiredPackages {
		if err := cache.Get(pkg.Repo.Base + "/" + pkg.Filename); err != nil {
			return err
		}
	}

	// also download the server ISO and netboot tar.gz, etc.
	for _, url := range []string{
		ubuntu.NobleServerISOURL,
		ubuntu.NobleServerNetbootURL,
		docker.GPGPublicKeyURL,
	} {
		if err := cache.Get(url); err != nil {
			return err
		}
	}

	return cache.Clean()
}
