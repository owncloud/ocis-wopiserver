package main

import (
	"os"

	"github.com/owncloud/ocis-wopiserver/pkg/command"
	"github.com/owncloud/ocis-wopiserver/pkg/config"
)

func main() {
	if err := command.Execute(config.New()); err != nil {
		os.Exit(1)
	}
}
