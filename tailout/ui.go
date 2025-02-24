package tailout

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/cterence/tailout/internal"
	"github.com/cterence/tailout/internal/views"
	"github.com/tailscale/tailscale-client-go/tailscale"

	"github.com/a-h/templ"
)

func (app *App) UI(args []string) error {
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
				slog.Error("failed to create node", "error", err)
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
				slog.Error("failed to stop nodes", "error", err)
			}
		}()
		w.WriteHeader(http.StatusNoContent)
	})

	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		nodes, err := internal.GetActiveNodes(client)
		if err != nil {
			slog.Error("failed to get active nodes", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		table := ""
		for _, node := range nodes {
			table += fmt.Sprintf("<tr class=\"bg-white border-b\"><td class=\"px-4 py-2\">%s</td><td class=\"px-4 py-2\">%s</td><td class=\"px-4 py-2\">%s</td></tr>", node.Hostname, node.Addresses[0], node.LastSeen)
		}
		if _, err := w.Write([]byte(table)); err != nil {
			slog.Error("failed to write response", "error", err)
		}
	})

	http.Handle("/health", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte(`{"status": {"server": "OK"}}`)); err != nil {
			slog.Error("failed to write health check response", "error", err)
		}
	}))

	// Serve assets files
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("internal/assets"))))

	slog.Info("Server starting", "address", app.Config.UI.Address, "port", app.Config.UI.Port)
	srv := &http.Server{
		Addr:         app.Config.UI.Address + ":" + app.Config.UI.Port,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil {
		slog.Error("Failed to start server", "error", err)
		panic(err)
	}
	return nil
}
