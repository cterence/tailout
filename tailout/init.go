package tailout

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/cterence/tailout/internal"
	tsapi "tailscale.com/client/tailscale/v2"
)

func (app *App) Init(ctx context.Context) error {
	dryRun := app.Config.DryRun
	nonInteractive := app.Config.NonInteractive

	baseURL, err := url.Parse(app.Config.Tailscale.BaseURL)
	if err != nil {
		return fmt.Errorf("failed to parse base URL: %w", err)
	}

	apiClient := &tsapi.Client{
		APIKey:  app.Config.Tailscale.APIKey,
		Tailnet: app.Config.Tailscale.Tailnet,
		BaseURL: baseURL,
	}

	// Get the ACL configuration
	acl, err := apiClient.PolicyFile().Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get acl: %w", err)
	}

	allowTailoutSSH := tsapi.ACLSSH{
		Action:      "check",
		Source:      []string{"autogroup:member"},
		Destination: []string{"tag:tailout"},
		Users:       []string{"autogroup:nonroot", "root"},
	}

	tailoutSSHConfigExists, tailoutTagExists, tailoutAutoApproversExists := false, false, false

	for _, sshConfig := range acl.SSH {
		if sshConfig.Action == "check" && sshConfig.Source[0] == "autogroup:member" && sshConfig.Destination[0] == "tag:tailout" && sshConfig.Users[0] == "autogroup:nonroot" && sshConfig.Users[1] == "root" {
			tailoutSSHConfigExists = true
		}
	}

	if acl.TagOwners["tag:tailout"] != nil {
		fmt.Println("Tag 'tag:tailout' already exists.")
		tailoutTagExists = true
	} else {
		acl.TagOwners["tag:tailout"] = []string{}
	}

	if acl.AutoApprovers == nil {
		fmt.Println("Auto approvers configuration does not exist.")
		acl.AutoApprovers = &tsapi.ACLAutoApprovers{}
	}

	for _, exitNode := range acl.AutoApprovers.ExitNode {
		if exitNode == "tag:tailout" {
			fmt.Println("Auto approvers for tag:tailout nodes already exists.")
			tailoutAutoApproversExists = true
		}
	}

	if !tailoutAutoApproversExists {
		acl.AutoApprovers.ExitNode = append(acl.AutoApprovers.ExitNode, "tag:tailout")
	}

	if tailoutSSHConfigExists {
		fmt.Println("SSH configuration for tailout already exists.")
	} else {
		acl.SSH = append(acl.SSH, allowTailoutSSH)
	}

	if tailoutTagExists && tailoutAutoApproversExists && tailoutSSHConfigExists && !dryRun {
		fmt.Println("Nothing to do.")
		return nil
	}

	// Validate the updated acl configuration
	err = apiClient.PolicyFile().Validate(ctx, *acl)
	if err != nil {
		return fmt.Errorf("failed to validate acl: %w", err)
	}

	// Update the acl configuration
	aclJSON, err := json.MarshalIndent(acl, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal acl: %w", err)
	}

	// Make a prompt to show the update that will be done
	fmt.Printf(`
The following update to the acl will be done:
- Add tag:tailout to tagOwners
- Update auto approvers to allow exit nodes tagged with tag:tailout
- Add a SSH configuration allowing users to SSH into tagged tailout nodes

Your new acl document will look like this:
%s
`, aclJSON)

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

		err = apiClient.PolicyFile().Set(ctx, *acl, "")
		if err != nil {
			return fmt.Errorf("failed to update acl: %w", err)
		}

		fmt.Println("ACL updated.")
	} else {
		fmt.Println("Dry run, not updating acl.")
	}
	return nil
}
