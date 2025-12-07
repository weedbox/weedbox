package weedbox

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type ModuleParams interface {
	GetLifecycle() fx.Lifecycle
	GetLogger() *zap.Logger
}

type Params struct {
	fx.In

	Lifecycle fx.Lifecycle
	Logger    *zap.Logger
}

func (p *Params) GetLifecycle() fx.Lifecycle {
	return p.Lifecycle
}

func (p *Params) GetLogger() *zap.Logger {
	return p.Logger
}
