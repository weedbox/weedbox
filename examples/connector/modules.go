package main

import (
	"github.com/weedbox/common-modules/daemon"
	"github.com/weedbox/common-modules/logger"
	"github.com/weedbox/weedbox/examples/connector/pkg/consumer"
	"github.com/weedbox/weedbox/examples/connector/pkg/english"
	"github.com/weedbox/weedbox/examples/connector/pkg/french"
	"go.uber.org/fx"
)

func preloadModules() ([]fx.Option, error) {
	return []fx.Option{
		fx.Supply(config),
		logger.Module(),
	}, nil
}

func loadModules() ([]fx.Option, error) {
	// Load order matters for the *unnamed default*: the first Greeter
	// loaded across the process wins ClaimDefault[greeter.Greeter] and
	// becomes the value injected when no `name:"..."` tag is used.
	//
	// Swap these two lines to make the French greeter the default.
	return []fx.Option{
		english.Module("english"),
		french.Module("french"),
		consumer.Module("consumer"),
	}, nil
}

func afterModules() ([]fx.Option, error) {
	return []fx.Option{
		daemon.Module("daemon"),
	}, nil
}
