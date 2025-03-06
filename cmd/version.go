package cmd

import (
	"fmt"
	"runtime/debug"

	"github.com/spf13/cobra"
)

func buildVersionString() string {
	revision := "unknown"
	commitTime := "unknown"

	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}

	for _, kv := range buildInfo.Settings {
		switch kv.Key {
		case "vcs.revision":
			revision = kv.Value
		case "vcs.time":
			commitTime = kv.Value
		}
	}

	return fmt.Sprintf("tailout version: %s\ncommit hash: %s\ncommit time: %s\ngo version: %s\n", buildInfo.Main.Version, revision, commitTime, buildInfo.GoVersion)
}

func buildVersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Args:  cobra.ArbitraryArgs,
		Use:   "version",
		Short: "Print the Tailout version",
		RunE: func(cmd *cobra.Command, args []string) error {
			version := buildVersionString()
			_, err := fmt.Print(version)
			if err != nil {
				return fmt.Errorf("failed to print version: %w", err)
			}
			return nil
		},
	}

	return cmd
}
