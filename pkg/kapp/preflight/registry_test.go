// Copyright 2024 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0
package preflight

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	ctlconf "github.com/vmware-tanzu/carvel-kapp/pkg/kapp/config"
	"github.com/vmware-tanzu/carvel-kapp/pkg/kapp/diffgraph"
	ctlres "github.com/vmware-tanzu/carvel-kapp/pkg/kapp/resources"
)

func TestRegistrySet(t *testing.T) {
	testCases := []struct {
		name       string
		preflights string
		registry   *Registry
		shouldErr  bool
	}{
		{
			name:       "no preflight checks registered, parsing skipped, any value can be provided",
			preflights: "someCheck",
			registry:   &Registry{},
		},
		{
			name:       "preflight checks registered, invalid check format in flag, error returned",
			preflights: ",",
			registry: &Registry{
				known: map[string]Check{
					"some": NewCheck(func(_ context.Context, _ *diffgraph.ChangeGraph, _ CheckConfig) error { return nil }, true),
				},
				enabledFlag: map[string]bool{},
			},
			shouldErr: true,
		},
		{
			name:       "preflight checks registered, unknown preflight check specified, error returned",
			preflights: "nonexistent",
			registry: &Registry{
				known: map[string]Check{
					"exists": NewCheck(func(_ context.Context, _ *diffgraph.ChangeGraph, _ CheckConfig) error { return nil }, true),
				},
				enabledFlag: map[string]bool{},
			},
			shouldErr: true,
		},
		{
			name:       "preflight checks registered, valid input, no error returned",
			preflights: "someCheck",
			registry: &Registry{
				known: map[string]Check{
					"someCheck": NewCheck(func(_ context.Context, _ *diffgraph.ChangeGraph, _ CheckConfig) error { return nil }, true),
				},
				enabledFlag: map[string]bool{},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.registry.Set(tc.preflights)
			if tc.shouldErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRegistryRun(t *testing.T) {
	testCases := []struct {
		name      string
		registry  *Registry
		shouldErr bool
	}{
		{
			name:     "no preflight checks registered, no error returned",
			registry: &Registry{},
		},
		{
			name: "preflight checks registered, disabled checks don't run",
			registry: &Registry{
				known: map[string]Check{
					"disabledCheck": NewCheck(func(_ context.Context, _ *diffgraph.ChangeGraph, _ CheckConfig) error {
						return errors.New("should be disabled")
					}, false),
				},
			},
		},
		{
			name: "preflight checks registered, enabled check returns an error, error returned",
			registry: &Registry{
				known: map[string]Check{
					"errorCheck": NewCheck(func(_ context.Context, _ *diffgraph.ChangeGraph, _ CheckConfig) error { return errors.New("error") }, true),
				},
			},
			shouldErr: true,
		},
		{
			name: "preflight checks registered, enabled checks successful, no error returned",
			registry: &Registry{
				known: map[string]Check{
					"someCheck": NewCheck(func(_ context.Context, _ *diffgraph.ChangeGraph, _ CheckConfig) error { return nil }, true),
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.registry.Run(nil, nil)
			require.Equal(t, tc.shouldErr, err != nil)
		})
	}
}

func TestRegistryDirectConfig(t *testing.T) {
	testCases := []struct {
		name      string
		registry  *Registry
		config    map[string]any
		shouldErr bool
	}{
		{
			name: "preflight checks registered, no config",
			registry: &Registry{
				known: map[string]Check{
					"disabledCheck": NewCheck(func(_ context.Context, _ *diffgraph.ChangeGraph, cfg CheckConfig) error {
						if cfg != nil {
							return errors.New("config should not be present")
						}
						return nil
					}, true),
				},
			},
		},
		{
			name: "preflight checks registered, config set, no error",
			registry: &Registry{
				known: map[string]Check{
					"someCheck": NewCheck(func(_ context.Context, _ *diffgraph.ChangeGraph, cfg CheckConfig) error {
						if cfg == nil {
							return errors.New("config should be present")
						}
						v, ok := cfg["foo"]
						if !ok {
							return errors.New("foo config not present")
						}
						if v != "bar" {
							return errors.New("foo should equal 'bar'")
						}
						return nil
					}, true),
				},
			},
			config: map[string]any{
				"foo": "bar",
			},
		},
		{
			name: "preflight checks registered, unexpected config set, no error",
			registry: &Registry{
				known: map[string]Check{
					"someCheck": NewCheck(func(_ context.Context, _ *diffgraph.ChangeGraph, cfg CheckConfig) error {
						if cfg == nil {
							return errors.New("config should be present")
						}
						_, ok := cfg["foobar"]
						if ok {
							return errors.New("foo config should not present")
						}
						return nil
					}, true),
				},
			},
			config: map[string]any{
				"foo": "bar",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for _, check := range tc.registry.known {
				_ = check.SetConfig(tc.config)
			}
			err := tc.registry.Run(nil, nil)
			if tc.shouldErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRegistryConfig(t *testing.T) {
	testCases := []struct {
		name      string
		registry  *Registry
		configErr bool
		shouldErr bool
	}{
		{
			name: "preflight checks registered, config set, no error",
			registry: &Registry{
				known: map[string]Check{
					"someCheck": NewCheck(func(_ context.Context, _ *diffgraph.ChangeGraph, cfg CheckConfig) error {
						if cfg == nil {
							return errors.New("config should be present")
						}
						v, ok := cfg["foo"]
						if !ok {
							return errors.New("foo config not present")
						}
						if v != "bar" {
							return errors.New("foo should equal 'bar'")
						}
						return nil
					}, true),
				},
			},
		},
		{
			name: "preflight checks registered, unexpected config set, error",
			registry: &Registry{
				known: map[string]Check{
					"someCheck": NewCheck(func(_ context.Context, _ *diffgraph.ChangeGraph, cfg CheckConfig) error {
						if cfg == nil {
							return errors.New("config should be present")
						}
						_, ok := cfg["foobar"]
						if ok {
							return errors.New("foobar config should not present")
						}
						return nil
					}, true),
				},
			},
		},
		{
			name: "preflight checks registered, unexpected preflight check set, error",
			registry: &Registry{
				known: map[string]Check{
					"otherCheck": NewCheck(func(_ context.Context, _ *diffgraph.ChangeGraph, cfg CheckConfig) error {
						if cfg != nil {
							return errors.New("config should not be present")
						}
						return nil
					}, true),
				},
			},
			configErr: true,
		},
	}
	configYAML := `---
apiVersion: kapp.k14s.io/v1alpha1
kind: Config
preflightRules:
- name: someCheck
  config:
    foo: bar
`

	configRs, err := ctlres.NewFileResource(ctlres.NewBytesSource([]byte(configYAML))).Resources()
	require.NoErrorf(t, err, "Parsing resources")
	_, conf, err := ctlconf.NewConfFromResources(configRs)
	require.NoErrorf(t, err, "Parsing config")

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.registry.SetConfig(conf)
			if tc.configErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			err = tc.registry.Run(nil, nil)
			if tc.shouldErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
