package tailout

import (
	"encoding/json"
	"fmt"

	"github.com/cterence/tailout/internal"
	"github.com/cterence/tailout/tailout/config"
	"github.com/cterence/tailout/tailout/tailscale"
)

func (app *App) Init() error {
	dryRun := app.Config.DryRun
	nonInteractive := app.Config.NonInteractive

	c := tailscale.NewClient(&app.Config.Tailscale)

	// Get the policy configuration
	policy, err := c.GetPolicy()
	if err != nil {
		return err
	}

	allowXitSSH := config.SSHConfiguration{
		Action: "check",
		Src:    []string{"autogroup:members"},
		Dst:    []string{"tag:tailout"},
		Users:  []string{"autogroup:nonroot", "root"},
	}

	xitSSHConfigExists, xitTagExists, xitAutoApproversExists := false, false, false

	for _, sshConfig := range policy.SSH {
		if sshConfig.Action == "check" && sshConfig.Src[0] == "autogroup:members" && sshConfig.Dst[0] == "tag:tailout" && sshConfig.Users[0] == "autogroup:nonroot" && sshConfig.Users[1] == "root" {
			xitSSHConfigExists = true
		}
	}

	if policy.TagOwners["tag:tailout"] != nil {
		fmt.Println("Tag 'tag:tailout' already exists.")
		xitTagExists = true
	} else {
		policy.TagOwners["tag:tailout"] = []string{}
	}

	if policy.AutoApprovers.ExitNode != nil {
		fmt.Println("Auto approvers for tag:tailout nodes already exists.")
		xitAutoApproversExists = true
	} else {
		policy.AutoApprovers.ExitNode = []string{"tag:tailout"}
	}

	if xitSSHConfigExists {
		fmt.Println("SSH configuration for tailout already exists.")
	} else {
		policy.SSH = append(policy.SSH, allowXitSSH)
	}

	if xitTagExists && xitAutoApproversExists && xitSSHConfigExists && !dryRun {
		fmt.Println("Nothing to do.")
		return nil
	}

	// Validate the updated policy configuration
	err = c.ValidatePolicy(policy)
	if err != nil {
		return fmt.Errorf("failed to validate policy: %w", err)
	}

	// Update the policy configuration
	policyJSON, err := json.MarshalIndent(policy, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal policy: %w", err)
	}

	// Make a prompt to show the update that will be done
	fmt.Printf(`
The following update to the policy will be done:
- Add tag:tailout to tagOwners
- Update auto approvers to allow exit nodes tagged with tag:tailout
- Add a SSH configuration allowing users to SSH into tagged tailout nodes

Your new policy document will look like this:
%s
`, policyJSON)

	if !dryRun {
		if !nonInteractive {
			result, err := internal.PromptYesNo("Do you want to continue?")
			if err != nil {
				return fmt.Errorf("failed to prompt for confirmation: %w", err)
			}

			if !result {
				fmt.Println("Aborting...")
				return nil
			}
		}

		err = c.UpdatePolicy(policy)
		if err != nil {
			return fmt.Errorf("failed to update policy: %w", err)
		}

		fmt.Println("Policy updated.")
	} else {
		fmt.Println("Dry run, not updating policy.")
	}
	return nil
}
