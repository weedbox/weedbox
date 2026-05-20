package french

import (
	"context"

	"github.com/weedbox/weedbox/examples/connector/pkg/greeter"
	"github.com/weedbox/weedbox/fxmodule"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Params struct {
	fx.In

	Lifecycle fx.Lifecycle
	Logger    *zap.Logger
}

type french struct {
	logger *zap.Logger
	scope  string
}

func (f *french) Greet(name string) string {
	return "Bonjour, " + name + " !"
}

func (f *french) onStart(ctx context.Context) error {
	f.logger.Info("french greeter started", zap.String("scope", f.scope))
	return nil
}

func (f *french) onStop(ctx context.Context) error {
	f.logger.Info("french greeter stopped", zap.String("scope", f.scope))
	return nil
}

// Module registers a French Greeter under the given scope. See the english
// package for the full story on how InterfaceModule wires the named and
// unnamed-default registrations.
func Module(scope string) fx.Option {
	return fxmodule.InterfaceModule[greeter.Greeter](
		scope,
		func(p Params) greeter.Greeter {
			f := &french{
				logger: p.Logger.Named(scope),
				scope:  scope,
			}
			p.Lifecycle.Append(fx.Hook{
				OnStart: f.onStart,
				OnStop:  f.onStop,
			})
			return f
		},
	)
}
