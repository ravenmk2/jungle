// Package main is the jungle binary entrypoint.
package main

import (
	"flag"

	"github.com/ravenmk2/jungle/internal/app"
	"github.com/sirupsen/logrus"
)

func main() {
	addr := flag.String("addr", "", "listen address (overrides config)")
	configDir := flag.String("config-dir", "./config", "config directory")
	dataDir := flag.String("data-dir", "", "data directory (overrides config)")
	flag.Parse()

	if err := app.Run(*configDir, *dataDir, *addr); err != nil {
		logrus.Fatal(err)
	}
}
