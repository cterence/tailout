package xit

import (
	"encoding/json"
	"fmt"

	"github.com/cterence/xit/internal"
)

func (app *App) Init() error {
	tsApiKey := app.Config.Tailscale.APIKey
	tailnet := app.Config.Tailscale.Tailnet
	dryRun := app.Config.DryRun
	nonInteractive := app.Config.NonInteractive

	// Get the policy configuration
	policy, err := internal.GetPolicy(tsApiKey, tailnet)
	if err != nil {
		return fmt.Errorf("failed to get policy: %w", err)
	}

	allowXitSSH := internal.SSHConfiguration{
		Action: "check",
		Src:    []string{"autogroup:members"},
		Dst:    []string{"tag:xit"},
		Users:  []string{"autogroup:nonroot", "root"},
	}

	xitSSHConfigExists, xitTagExists, xitAutoApproversExists := false, false, false

	for _, sshConfig := range policy.SSH {
		if sshConfig.Action == "check" && sshConfig.Src[0] == "autogroup:members" && sshConfig.Dst[0] == "tag:xit" && sshConfig.Users[0] == "autogroup:nonroot" && sshConfig.Users[1] == "root" {
			xitSSHConfigExists = true
		}
	}

	if policy.TagOwners["tag:xit"] != nil {
		fmt.Println("Tag 'tag:xit' already exists.")
		xitTagExists = true
	} else {
		policy.TagOwners["tag:xit"] = []string{}
	}

	if policy.AutoApprovers.ExitNode != nil {
		fmt.Println("Auto approvers for tag:xit nodes already exists.")
		xitAutoApproversExists = true
	} else {
		policy.AutoApprovers.ExitNode = []string{"tag:xit"}
	}

	if xitSSHConfigExists {
		fmt.Println("SSH configuration for xit already exists.")
	} else {
		policy.SSH = append(policy.SSH, allowXitSSH)
	}

	if xitTagExists && xitAutoApproversExists && xitSSHConfigExists && !dryRun {
		fmt.Println("Nothing to do.")
		return nil
	}

	// Validate the updated policy configuration
	err = internal.ValidatePolicy(tsApiKey, tailnet, policy)
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
- Add tag:xit to tagOwners
- Update auto approvers to allow exit nodes tagged with tag:xit
- Add a SSH configuration allowing users to SSH into tagged xit nodes

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

		err = internal.UpdatePolicy(tsApiKey, tailnet, policy)
		if err != nil {
			return fmt.Errorf("failed to update policy: %w", err)
		}

		fmt.Println("Policy updated.")
	} else {
		fmt.Println("Dry run, not updating policy.")
	}
	return nil
}
