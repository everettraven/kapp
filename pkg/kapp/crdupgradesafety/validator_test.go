// Copyright 2024 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package crdupgradesafety

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

func TestValidator(t *testing.T) {
	for _, tc := range []struct {
		name        string
		validations []Validation
		shouldErr   bool
	}{
		{
			name:        "no validators, no error",
			validations: []Validation{},
		},
		{
			name: "passing validator, no error",
			validations: []Validation{
				NewValidationFunc("pass", func(_, _ v1.CustomResourceDefinition) error {
					return nil
				}),
			},
		},
		{
			name: "failing validator, error",
			validations: []Validation{
				NewValidationFunc("fail", func(_, _ v1.CustomResourceDefinition) error {
					return errors.New("boom")
				}),
			},
			shouldErr: true,
		},
		{
			name: "passing+failing validator, error",
			validations: []Validation{
				NewValidationFunc("pass", func(_, _ v1.CustomResourceDefinition) error {
					return nil
				}),
				NewValidationFunc("fail", func(_, _ v1.CustomResourceDefinition) error {
					return errors.New("boom")
				}),
			},
			shouldErr: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			v := Validator{
				Validations: tc.validations,
			}
			var o, n v1.CustomResourceDefinition

			err := v.Validate(o, n)
			require.Equal(t, tc.shouldErr, err != nil)
		})
	}
}
