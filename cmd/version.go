package cmd

import (
	"errors"
	"fmt"
	"github.com/cterence/tailout/tailout"
	"github.com/spf13/cobra"
	"runtime/debug"
)

func buildVersionString(buildInfo *debug.BuildInfo) string {
	var revision string

	for _, kv := range buildInfo.Settings {
		switch kv.Key {
		case "vcs.revision":
			revision = kv.Value[max(0, len(kv.Value)-7):]
		}
	}

	if revision == "" {
		revision = "unknown"
	}

	return fmt.Sprintf("%s (%s)", buildInfo.Main.Version, revision)
}

func buildVersionCommand(app *tailout.App) *cobra.Command {
	cmd := &cobra.Command{
		Args:  cobra.ArbitraryArgs,
		Use:   "version",
		Short: "Print the Tailout version",
		RunE: func(cmd *cobra.Command, args []string) error {
			buildInfo, ok := debug.ReadBuildInfo()
			if !ok {
				return errors.New("Unable to ReadBuildInfo(), which shouldn't happen, as Tailout should be built with module support")
			}
			_, err := fmt.Printf("Tailout version %s\n", buildVersionString(buildInfo))

			if err != nil {
				return err
			}
			return nil
		},
	}

	return cmd
}
