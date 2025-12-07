package example

import (
	"context"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

const (
	ModuleName = "Example"
)

type Example struct {
	params Params
	logger *zap.Logger
	scope  string
}

type Params struct {
	fx.In

	Lifecycle fx.Lifecycle
	Logger    *zap.Logger
}

func Module(scope string) fx.Option {

	var m *Example

	return fx.Module(
		scope,
		fx.Provide(func(p Params) *Example {
			return &Example{
				params: p,
				logger: p.Logger.Named(scope),
				scope:  scope,
			}
		}),
		fx.Populate(&m),
		fx.Invoke(func(p Params) {

			p.Lifecycle.Append(
				fx.Hook{
					OnStart: m.onStart,
					OnStop:  m.onStop,
				},
			)
		}),
	)

}

func (m *Example) onStart(ctx context.Context) error {
	m.logger.Info("Starting " + ModuleName)
	return nil
}

func (m *Example) onStop(ctx context.Context) error {
	m.logger.Info("Stopped " + ModuleName)
	return nil
}

func (m *Example) Hello() {
	m.logger.Info("Hello from Example module!")
}
