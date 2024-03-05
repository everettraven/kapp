// Copyright 2024 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package preflight

import (
	"context"

	ctldgraph "github.com/vmware-tanzu/carvel-kapp/pkg/kapp/diffgraph"
)

// The following is the interface for Preflight checks
type CheckConfig map[string]any
type CheckFunc func(context.Context, *ctldgraph.ChangeGraph, CheckConfig) error

type Check interface {
	Enabled() bool
	SetEnabled(bool)
	SetConfig(CheckConfig) error
	Run(context.Context, *ctldgraph.ChangeGraph) error
}

// The following is an example/test Preflight check
type checkImpl struct {
	enabled   bool
	checkFunc CheckFunc
	config    map[string]any
}

func NewCheck(cf CheckFunc, enabled bool) Check {
	return &checkImpl{
		enabled:   enabled,
		checkFunc: cf,
	}
}

func (cf *checkImpl) Enabled() bool {
	return cf.enabled
}

func (cf *checkImpl) SetEnabled(enabled bool) {
	cf.enabled = enabled
}

func (cf *checkImpl) SetConfig(config CheckConfig) error {
	cf.config = config
	return nil
}

func (cf *checkImpl) Run(ctx context.Context, changeGraph *ctldgraph.ChangeGraph) error {
	return cf.checkFunc(ctx, changeGraph, cf.config)
}
