package tailout

import (
	"github.com/cterence/tailout/tailout/config"
)

type App struct {
	Config *config.Config
}

func New() (*App, error) {
	c := &config.Config{}
	app := &App{
		Config: c,
	}
	return app, nil
}
