package tailout

import (
	"io"
	"os"

	"github.com/cterence/tailout/tailout/config"
)

type App struct {
	Config *config.Config

	Out io.Writer
	Err io.Writer
}

func New() (*App, error) {
	c := &config.Config{}
	app := &App{
		Config: c,
		Out:    os.Stdout,
		Err:    os.Stderr,
	}
	return app, nil
}
