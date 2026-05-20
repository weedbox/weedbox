package english

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

type english struct {
	logger *zap.Logger
	scope  string
}

func (e *english) Greet(name string) string {
	return "Hello, " + name + "!"
}

func (e *english) onStart(ctx context.Context) error {
	e.logger.Info("english greeter started", zap.String("scope", e.scope))
	return nil
}

func (e *english) onStop(ctx context.Context) error {
	e.logger.Info("english greeter stopped", zap.String("scope", e.scope))
	return nil
}

// Module registers an English Greeter under the given scope. Because it goes
// through fxmodule.InterfaceModule[greeter.Greeter], it shows up as both:
//
//   - a named provider tagged `name:"<scope>"`, addressable by consumers that
//     want to pick this specific implementation
//   - if this is the first Greeter loaded in the process, the unnamed default
//     of greeter.Greeter — so single-load consumers that inject Greeter without
//     a name tag keep working
func Module(scope string) fx.Option {
	return fxmodule.InterfaceModule[greeter.Greeter](
		scope,
		func(p Params) greeter.Greeter {
			e := &english{
				logger: p.Logger.Named(scope),
				scope:  scope,
			}
			p.Lifecycle.Append(fx.Hook{
				OnStart: e.onStart,
				OnStop:  e.onStop,
			})
			return e
		},
	)
}
