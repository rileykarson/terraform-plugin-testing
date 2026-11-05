// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package resource

import (
	"context"
	"fmt"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/mitchellh/go-testing-interface"

	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/hashicorp/terraform-plugin-testing/internal/logging"
	"github.com/hashicorp/terraform-plugin-testing/internal/plugintest"
)

func testStepNewRefreshState(ctx context.Context, t testing.T, wd *plugintest.WorkingDir, step TestStep, providers *providerFactories) error {
	t.Helper()

	var err error
	// Explicitly ensure prior state exists before refresh.
	err = runProviderCommand(ctx, t, wd, providers, func() error {
		_, _, err = getState(ctx, t, wd)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Error getting state: %s", err)
	}

	err = runProviderCommand(ctx, t, wd, providers, func() error {
		return wd.Refresh(ctx)
	})
	if err != nil {
		return err
	}

	var refreshState *terraform.State
	err = runProviderCommand(ctx, t, wd, providers, func() error {
		_, refreshState, err = getState(ctx, t, wd)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Error getting state: %s", err)
	}

	// Go through the refreshed state and verify
	if step.Check != nil {
		logging.HelperResourceDebug(ctx, "Calling TestStep Check for RefreshState")

		if err := step.Check(refreshState); err != nil {
			t.Fatal(err)
		}

		logging.HelperResourceDebug(ctx, "Called TestStep Check for RefreshState")
	}

	// do a plan
	var hasChanges bool
	err = runProviderCommand(ctx, t, wd, providers, func() error {
		var err error
		hasChanges, err = wd.CreatePlan(ctx)
		return err
	})
	if err != nil {
		return fmt.Errorf("Error running post-refresh plan: %w", err)
	}

	planEmpty := !hasChanges
	if hasChanges || len(step.RefreshPlanChecks.PostRefresh) > 0 {
		var plan *tfjson.Plan
		err = runProviderCommand(ctx, t, wd, providers, func() error {
			var err error
			plan, err = wd.SavedPlan(ctx)
			return err
		})
		if err != nil {
			return fmt.Errorf("Error retrieving post-refresh plan: %w", err)
		}

		// Run post-refresh plan checks
		if len(step.RefreshPlanChecks.PostRefresh) > 0 {
			err = runPlanChecks(ctx, t, plan, step.RefreshPlanChecks.PostRefresh)
			if err != nil {
				return fmt.Errorf("Post-refresh plan check(s) failed:\n%w", err)
			}
		}

		planEmpty = planIsEmpty(plan, wd.GetHelper().TerraformVersion())
	}

	if !planEmpty && !step.ExpectNonEmptyPlan {
		var stdout string
		err = runProviderCommand(ctx, t, wd, providers, func() error {
			var err error
			stdout, err = wd.SavedPlanRawStdout(ctx)
			return err
		})
		if err != nil {
			return fmt.Errorf("Error retrieving formatted plan output: %w", err)
		}
		return fmt.Errorf("After refreshing state during this test step, a followup plan was not empty.\nstdout:\n\n%s", stdout)
	}

	return nil
}
