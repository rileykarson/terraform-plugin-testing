// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package logging_test

import (
	"bytes"
	"context"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-log/tfsdklog"
	"github.com/hashicorp/terraform-plugin-log/tfsdklogtest"

	"github.com/hashicorp/terraform-plugin-testing/internal/logging"
)

func TestHelperResourceDebug(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer

	ctx := tfsdklogtest.RootLogger(context.Background(), &output)

	// InitTestContext messes with the standard library log package, which
	// we want to avoid in this unit testing. Instead, just create the
	// helper_resource subsystem and avoid the other InitTestContext logic.
	ctx = tfsdklog.NewSubsystem(ctx, logging.SubsystemHelperResource)

	logging.HelperResourceDebug(ctx, "test message")

	entries, err := tfsdklogtest.MultilineJSONDecode(&output)

	if err != nil {
		t.Fatalf("unable to read multiple line JSON: %s", err)
	}

	expectedEntries := []map[string]interface{}{
		{
			"@level":   "debug",
			"@message": "test message",
			"@module":  "sdk.helper_resource",
		},
	}

	if diff := cmp.Diff(entries, expectedEntries); diff != "" {
		t.Errorf("unexpected difference: %s", diff)
	}
}

func TestHelperResourceError(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer

	ctx := tfsdklogtest.RootLogger(context.Background(), &output)

	// InitTestContext messes with the standard library log package, which
	// we want to avoid in this unit testing. Instead, just create the
	// helper_resource subsystem and avoid the other InitTestContext logic.
	ctx = tfsdklog.NewSubsystem(ctx, logging.SubsystemHelperResource)

	logging.HelperResourceError(ctx, "test message")

	entries, err := tfsdklogtest.MultilineJSONDecode(&output)

	if err != nil {
		t.Fatalf("unable to read multiple line JSON: %s", err)
	}

	expectedEntries := []map[string]interface{}{
		{
			"@level":   "error",
			"@message": "test message",
			"@module":  "sdk.helper_resource",
		},
	}

	if diff := cmp.Diff(entries, expectedEntries); diff != "" {
		t.Errorf("unexpected difference: %s", diff)
	}
}

func TestHelperResourceTrace(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer

	ctx := tfsdklogtest.RootLogger(context.Background(), &output)

	// InitTestContext messes with the standard library log package, which
	// we want to avoid in this unit testing. Instead, just create the
	// helper_resource subsystem and avoid the other InitTestContext logic.
	ctx = tfsdklog.NewSubsystem(ctx, logging.SubsystemHelperResource)

	logging.HelperResourceTrace(ctx, "test message")

	entries, err := tfsdklogtest.MultilineJSONDecode(&output)

	if err != nil {
		t.Fatalf("unable to read multiple line JSON: %s", err)
	}

	expectedEntries := []map[string]interface{}{
		{
			"@level":   "trace",
			"@message": "test message",
			"@module":  "sdk.helper_resource",
		},
	}

	if diff := cmp.Diff(entries, expectedEntries); diff != "" {
		t.Errorf("unexpected difference: %s", diff)
	}
}

func TestHelperResourceWarn(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer

	ctx := tfsdklogtest.RootLogger(context.Background(), &output)

	// InitTestContext messes with the standard library log package, which
	// we want to avoid in this unit testing. Instead, just create the
	// helper_resource subsystem and avoid the other InitTestContext logic.
	ctx = tfsdklog.NewSubsystem(ctx, logging.SubsystemHelperResource)

	logging.HelperResourceWarn(ctx, "test message")

	entries, err := tfsdklogtest.MultilineJSONDecode(&output)

	if err != nil {
		t.Fatalf("unable to read multiple line JSON: %s", err)
	}

	expectedEntries := []map[string]interface{}{
		{
			"@level":   "warn",
			"@message": "test message",
			"@module":  "sdk.helper_resource",
		},
	}

	if diff := cmp.Diff(entries, expectedEntries); diff != "" {
		t.Errorf("unexpected difference: %s", diff)
	}
}

func TestIsHelperResourceTrace(t *testing.T) {
	t.Cleanup(func() {
		var restoreBuf bytes.Buffer
		restoreCtx := tfsdklogtest.RootLogger(context.Background(), &restoreBuf)
		_ = tfsdklog.NewSubsystem(restoreCtx, logging.SubsystemHelperResource)
	})

	t.Setenv("TF_LOG", "")
	t.Setenv("TF_ACC_LOG_PATH", "")
	t.Setenv("TF_LOG_PROVIDER", "")
	t.Setenv(logging.EnvTfLogSdk, "")
	t.Setenv(logging.EnvTfLogSdkHelperResource, "")
	t.Setenv("TF_LOG_PATH", filepath.Join(t.TempDir(), "test.log"))

	testCases := map[string]struct {
		tfLog                string
		useAccLogPath        bool
		tfLogProvider        string
		tfLogSdk             string
		tfLogHelperResource  string
		expectedTraceEnabled bool
	}{
		"unset": {
			expectedTraceEnabled: false,
		},
		"tf-log-debug": {
			tfLog:                "DEBUG",
			expectedTraceEnabled: false,
		},
		"tf-log-trace": {
			tfLog:                "TRACE",
			expectedTraceEnabled: true,
		},
		"tf-log-json": {
			tfLog:                "JSON",
			expectedTraceEnabled: true,
		},
		"tf-acc-log-path": {
			useAccLogPath:        true,
			expectedTraceEnabled: true,
		},
		"tf-log-provider-trace": {
			tfLogProvider:        "TRACE",
			expectedTraceEnabled: true,
		},
		"tf-log-sdk-trace": {
			tfLogSdk:             "TRACE",
			expectedTraceEnabled: true,
		},
		"tf-log-trace-overridden-by-sdk-warn": {
			tfLog:                "TRACE",
			tfLogSdk:             "WARN",
			expectedTraceEnabled: false,
		},
		"tf-log-sdk-warn-overridden-by-helper-resource-trace": {
			tfLog:                "TRACE",
			tfLogSdk:             "WARN",
			tfLogHelperResource:  "TRACE",
			expectedTraceEnabled: true,
		},
		"tf-log-trace-overridden-by-helper-resource-debug": {
			tfLog:                "TRACE",
			tfLogHelperResource:  "DEBUG",
			expectedTraceEnabled: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Setenv("TF_LOG", tc.tfLog)
			if tc.useAccLogPath {
				t.Setenv("TF_ACC_LOG_PATH", filepath.Join(t.TempDir(), "acc.log"))
			} else {
				t.Setenv("TF_ACC_LOG_PATH", "")
			}
			t.Setenv("TF_LOG_PROVIDER", tc.tfLogProvider)
			t.Setenv(logging.EnvTfLogSdk, tc.tfLogSdk)
			t.Setenv(logging.EnvTfLogSdkHelperResource, tc.tfLogHelperResource)

			ctx := logging.InitTestContext(context.Background(), t)

			if got := logging.IsHelperResourceTrace(ctx); got != tc.expectedTraceEnabled {
				t.Fatalf("expected IsHelperResourceTrace=%t, got %t", tc.expectedTraceEnabled, got)
			}
		})
	}
}
