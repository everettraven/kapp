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
					"some": NewCheck(func(_ context.Context, _ *diffgraph.ChangeGraph, _ CheckConfig) error {
						return nil
					}, nil, true),
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
					"exists": NewCheck(func(_ context.Context, _ *diffgraph.ChangeGraph, _ CheckConfig) error {
						return nil
					}, nil, true),
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
					"someCheck": NewCheck(func(_ context.Context, _ *diffgraph.ChangeGraph, _ CheckConfig) error {
						return nil
					}, nil, true),
				},
				enabledFlag: map[string]bool{},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.registry.Set(tc.preflights)
			require.Equalf(t, tc.shouldErr, err != nil, "Unexpected error: %v", err)
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
					}, nil, false),
				},
			},
		},
		{
			name: "preflight checks registered, enabled check returns an error, error returned",
			registry: &Registry{
				known: map[string]Check{
					"errorCheck": NewCheck(func(_ context.Context, _ *diffgraph.ChangeGraph, _ CheckConfig) error {
						return errors.New("error")
					}, nil, true),
				},
			},
			shouldErr: true,
		},
		{
			name: "preflight checks registered, enabled checks successful, no error returned",
			registry: &Registry{
				known: map[string]Check{
					"someCheck": NewCheck(func(_ context.Context, _ *diffgraph.ChangeGraph, _ CheckConfig) error {
						return nil
					}, nil, true),
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.registry.Run(nil, nil)
			require.Equalf(t, tc.shouldErr, err != nil, "Unexpected error: %v", err)
		})
	}
}

func TestRegistryConfig(t *testing.T) {
	testCases := []struct {
		name       string
		registry   *Registry
		configYaml string
		shouldErr  bool
	}{
		{
			name: "preflight checks registered, config set, no error",
			registry: &Registry{
				known: map[string]Check{
					"someCheck": NewCheck(
						nil,
						func(cfg CheckConfig) error {
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
							_, ok = cfg["foobar"]
							if ok {
								return errors.New("foobar config should not present")
							}
							return nil
						},
						true,
					),
				},
			},
			configYaml: `---
apiVersion: kapp.k14s.io/v1alpha1
kind: Config
preflightRules:
- name: someCheck
  config:
    foo: bar
`,
		},
		{
			name: "preflight checks registered, unexpected preflight check set, error",
			registry: &Registry{
				known: map[string]Check{
					"someCheck": NewCheck(
						nil,
						func(cfg CheckConfig) error {
							if cfg != nil {
								return errors.New("config should not be present")
							}
							return nil
						},
						true,
					),
				},
			},
			configYaml: `---
apiVersion: kapp.k14s.io/v1alpha1
kind: Config
preflightRules:
- name: otherCheck
  config:
    foo: bar
`,
			shouldErr: true,
		},
		{
			name: "preflight checks registered, duplicate config entry, error",
			registry: &Registry{
				known: map[string]Check{
					"someCheck": NewCheck(nil, nil, true),
				},
			},
			configYaml: `---
apiVersion: kapp.k14s.io/v1alpha1
kind: Config
preflightRules:
- name: someCheck
  config:
    foo: bar
- name: someCheck
  config:
    bar: foo
`,
			shouldErr: true,
		},
	}

	for _, tc := range testCases {
		configRs, err := ctlres.NewFileResource(ctlres.NewBytesSource([]byte(tc.configYaml))).Resources()
		require.NoErrorf(t, err, "Parsing resources")
		_, conf, err := ctlconf.NewConfFromResources(configRs)
		require.NoErrorf(t, err, "Parsing config")

		t.Run(tc.name, func(t *testing.T) {
			err := tc.registry.SetConfig(conf)
			require.Equalf(t, tc.shouldErr, err != nil, "Unexpected error: %v", err)
		})
	}
}
