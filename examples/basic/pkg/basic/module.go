package basic

import (
	"context"

	"github.com/weedbox/weedbox"
	"github.com/weedbox/weedbox/examples/basic/pkg/example"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const (
	ModuleName = "Basic"
)

type Params struct {
	weedbox.Params
	Example *example.Example
}

type Basic struct {
	weedbox.Module[*Params]
}

func Module(scope string) fx.Option {
	m := new(Basic)

	return fx.Module(
		scope,
		fx.Supply(
			fx.Annotated{
				Name:   scope,
				Target: m,
			},
		),
		fx.Invoke(func(p Params) {
			weedbox.InitModule(scope, &p, m)
		}),
	)
}

func (m *Basic) OnStart(ctx context.Context) error {
	m.Logger().Info("Starting " + ModuleName)

	m.Logger().Info("config", zap.String("path", m.GetConfigPath("XX")))

	m.Params().Example.Hello()

	return nil
}

func (m *Basic) OnStop(ctx context.Context) error {
	m.Logger().Info("Stopped " + ModuleName)
	return nil
}

func (m *Basic) InitDefaultConfigs() {
	// Initialize default configurations here
	m.Logger().Info("Initializing default configs for " + ModuleName)
}
