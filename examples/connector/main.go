package main

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/weedbox/common-modules/configs"
	"go.uber.org/fx"
)

const (
	appName        = "connector"
	appDescription = "connector demonstrates fxmodule.InterfaceModule with multiple swappable implementations of one interface."
)

var (
	config  *configs.Config
	pc      bool
	verbose bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:  appName,
		Long: appDescription,
		RunE: func(cmd *cobra.Command, args []string) error {
			return run()
		},
	}

	config = configs.NewConfig("SERVICE")
	rootCmd.Flags().BoolVar(&pc, "print_configs", false, "Print all available configs")
	rootCmd.Flags().BoolVar(&verbose, "verbose", false, "Display detailed logs")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func initModules() ([]fx.Option, error) {
	var modules []fx.Option

	for _, loader := range []func() ([]fx.Option, error){
		preloadModules,
		loadModules,
		afterModules,
	} {
		m, err := loader()
		if err != nil {
			return modules, err
		}
		modules = append(modules, m...)
	}

	if !verbose {
		modules = append(modules, fx.NopLogger)
	}

	return modules, nil
}

func run() error {
	modules, err := initModules()
	if err != nil {
		return err
	}

	app := fx.New(modules...)

	if pc {
		config.PrintAllSettings()
	}

	app.Run()
	return nil
}
