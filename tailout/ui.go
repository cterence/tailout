package tailout

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/cterence/tailout/internal"
	"github.com/cterence/tailout/internal/views"
	"github.com/tailscale/tailscale-client-go/tailscale"

	"github.com/a-h/templ"
)

func (app *App) Ui(args []string) error {
	indexComponent := views.Index()
	app.Config.NonInteractive = true

	client, err := tailscale.NewClient(app.Config.Tailscale.APIKey, app.Config.Tailscale.Tailnet)
	if err != nil {
		return fmt.Errorf("failed to create tailscale client: %w", err)
	}

	http.Handle("/", templ.Handler(indexComponent))

	http.HandleFunc("/create", func(w http.ResponseWriter, r *http.Request) {
		slog.Info("Creating tailout node")
		go func() {
			err := app.Create()
			if err != nil {
				slog.Error("failed to create node: " + err.Error())
			}
		}()
		w.WriteHeader(http.StatusCreated)
	})

	http.HandleFunc("/stop", func(w http.ResponseWriter, r *http.Request) {
		slog.Info("Stopping tailout nodes")
		app.Config.Stop.All = true
		go func() {
			err := app.Stop(nil)
			if err != nil {
				slog.Error("failed to create node: " + err.Error())
			}
		}()
		w.WriteHeader(http.StatusNoContent)
	})

	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		nodes, err := internal.GetActiveNodes(client)
		if err != nil {
			slog.Error("failed to get active nodes: " + err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		table := ""
		for _, node := range nodes {
			table += fmt.Sprintf("<tr class=\"bg-white border-b\"><td class=\"px-4 py-2\">%s</td><td class=\"px-4 py-2\">%s</td><td class=\"px-4 py-2\">%s</td></tr>", node.Hostname, node.Addresses[0], node.LastSeen)
		}
		w.Write([]byte(table))
	})

	http.Handle("/health", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status": {"server": "OK"}}`))
	}))

	// Serve assets files
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("internal/assets"))))

	slog.Info("Listening on " + app.Config.Ui.Address + ":" + app.Config.Ui.Port)
	err = http.ListenAndServe(app.Config.Ui.Address+":"+app.Config.Ui.Port, nil)
	if err != nil {
		slog.Error("Failed to start server")
		panic(err)
	}
	return nil
}
