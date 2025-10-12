package main

import (
	"fmt"
	"io"
	"os"

	"provisioner/pkg/config"
)

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

	var template string
	if len(os.Args) == 1 {
		b, err := io.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
		template = string(b)

	} else {
		template = os.Args[1]
	}

	b, err := config.Template(template)
	if err != nil {
		return err
	}

	_, err = os.Stdout.Write(b)
	return err
}
