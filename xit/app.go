package xit

import (
	"io"
	"os"

	"github.com/cterence/xit/xit/config"
)

type App struct {
	Config *config.Config

	Out io.Writer
	Err io.Writer
}

type Tailscale struct {
	AuthKey string
	APIKey  string
	Tailnet string
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
