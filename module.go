package weedbox

import (
	"context"
	"fmt"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

type ModuleInterface interface {
	OnStart(context.Context) error
	OnStop(context.Context) error
	GetConfigPath(key string) string
	Logger() *zap.Logger
	InitDefaultConfigs()
}

type ModuleContainer[P ModuleParams] interface {
	ModuleInterface
	SetBaseModule(Module[P])
}

func InitModule[P ModuleParams](scope string, p P, target ModuleContainer[P]) {
	p.GetLifecycle().Append(
		fx.Hook{
			OnStart: target.OnStart,
			OnStop:  target.OnStop,
		},
	)

	target.SetBaseModule(Module[P]{
		scope:  scope,
		target: target,
		logger: p.GetLogger().Named(scope),
		params: p,
	})

	target.InitDefaultConfigs()
}

type Module[P ModuleParams] struct {
	scope  string
	target ModuleInterface
	logger *zap.Logger
	params P
}

func (m *Module[P]) SetBaseModule(base Module[P]) {
	*m = base
}

func (m *Module[P]) InitDefaultConfigs() {
	// Initialize default configurations here
}

func (m *Module[P]) GetConfigPath(key string) string {
	return fmt.Sprintf("%s.%s", m.scope, key)
}

func (m *Module[P]) Logger() *zap.Logger {
	return m.logger
}

func (m *Module[P]) OnStart(ctx context.Context) error {
	m.Logger().Info("Starting module: " + m.scope)
	return nil
}

func (m *Module[P]) OnStop(ctx context.Context) error {
	m.Logger().Info("Stopped module: " + m.scope)
	return nil
}

func (m *Module[P]) Params() P {
	return m.params
}
