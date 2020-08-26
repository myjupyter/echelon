package main

import (
	"errors"
	"flag"
	"net/http"

	app "github.com/myjupyter/echelon/internal/application"
	log "github.com/sirupsen/logrus"
)

func main() {
	var configPath, roleConfig string
	flag.StringVar(&configPath, "config-path", "./configs/config.json", "Path to configure file")
	flag.StringVar(&roleConfig, "role-config", "./configs/roles.json", "Path to roles configure file")

	flag.Parse()

	application := app.NewApplication(configPath, roleConfig)
	if err := application.Start(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}
}
