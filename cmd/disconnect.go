package cmd

import (
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/cterence/xit/common"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// disconnectCmd represents the disconnect command
var disconnectCmd = &cobra.Command{
	Use:   "disconnect",
	Short: "Disconnect from the current exit node",
	Long:  ``,
	PreRun: func(cmd *cobra.Command, args []string) {
		viper.BindPFlag("non_interactive", cmd.PersistentFlags().Lookup("non-interactive"))
	},
	Run: func(cmd *cobra.Command, args []string) {
		nonInteractive := viper.GetBool("non_interactive")

		var status common.TailscaleStatus

		out, err := exec.Command("tailscale", "debug", "prefs").CombinedOutput()
		if err != nil {
			fmt.Println("Failed to get tailscale preferences:", err)
			return
		}

		json.Unmarshal(out, &status)

		if status.ExitNodeID == "" {
			fmt.Println("Not connected to an exit node.")
			return
		}

		err = common.RunTailscaleUpCommand("tailscale up --exit-node=", nonInteractive)
		if err != nil {
			fmt.Println("Failed to run tailscale up command:", err)
			return
		}

		fmt.Println("Disconnected.")
	},
}

func init() {
	rootCmd.AddCommand(disconnectCmd)

	disconnectCmd.PersistentFlags().BoolP("non-interactive", "n", false, "Do not prompt for confirmation")
}
