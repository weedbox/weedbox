package consumer

import (
	"context"

	"github.com/weedbox/weedbox/examples/connector/pkg/greeter"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Params shows the three ways to consume a connector-style interface:
//
//   - Default: injected without a name tag. Resolves to whichever Greeter was
//     loaded first in modules.go (the one that won ClaimDefault).
//   - English / French: injected by `name:"<scope>"`, addressing a specific
//     implementation regardless of load order.
type Params struct {
	fx.In

	Lifecycle fx.Lifecycle
	Logger    *zap.Logger

	Default greeter.Greeter
	English greeter.Greeter `name:"english"`
	French  greeter.Greeter `name:"french"`
}

// Module wires a small OnStart hook that exercises all three injections so a
// reader running `go run .` can see which Greeter resolved where.
func Module(scope string) fx.Option {
	return fx.Module(
		scope,
		fx.Invoke(func(p Params) {
			logger := p.Logger.Named(scope)
			p.Lifecycle.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					logger.Info("default greeting",
						zap.String("msg", p.Default.Greet("world")))
					logger.Info("english greeting",
						zap.String("msg", p.English.Greet("Alice")))
					logger.Info("french greeting",
						zap.String("msg", p.French.Greet("Bob")))
					return nil
				},
			})
		}),
	)
}
