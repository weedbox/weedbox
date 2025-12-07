# Weedbox

A Go application development framework built on [Uber FX](https://uber-go.github.io/fx/), providing modular architecture and dependency injection.

## Features

- Dependency injection based on Uber FX
- Modular application architecture
- Generic-based module foundation
- Unified lifecycle management (OnStart/OnStop)
- Built-in logging support (using Zap)
- Configuration management integration

## Installation

```bash
go get github.com/weedbox/weedbox
```

## Quick Start

### Basic Application Structure

See `examples/basic/` for a complete example.

```go
package main

import (
    "github.com/spf13/cobra"
    "github.com/weedbox/common-modules/configs"
    "go.uber.org/fx"
)

func main() {
    rootCmd := &cobra.Command{
        Use:  "myapp",
        Long: "My application",
        RunE: func(cmd *cobra.Command, args []string) error {
            return run()
        },
    }

    rootCmd.Execute()
}

func run() error {
    modules, err := initModules()
    if err != nil {
        return err
    }

    app := fx.New(modules...)
    app.Run()

    return nil
}
```

## Creating Modules

Weedbox supports two approaches for creating modules:

### Approach 1: Using Weedbox Base Class (Recommended)

This approach uses the generic base class `weedbox.Module`, providing a cleaner implementation.

```go
package mymodule

import (
    "context"
    "github.com/weedbox/weedbox"
    "go.uber.org/fx"
)

const ModuleName = "MyModule"

// Define parameter structure
type Params struct {
    weedbox.Params
    // Inject other module dependencies here
    // OtherModule *other.Module
}

// Define module structure, embedding weedbox.Module
type MyModule struct {
    weedbox.Module[*Params]
}

// Factory function to create the module
func Module(scope string) fx.Option {
    m := new(MyModule)

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

// Implement OnStart lifecycle
func (m *MyModule) OnStart(ctx context.Context) error {
    m.Logger().Info("Starting " + ModuleName)

    // Use config path
    configPath := m.GetConfigPath("some_key")
    m.Logger().Info("Config path: " + configPath)

    // Access other modules
    // m.Params().OtherModule.DoSomething()

    return nil
}

// Implement OnStop lifecycle
func (m *MyModule) OnStop(ctx context.Context) error {
    m.Logger().Info("Stopped " + ModuleName)
    return nil
}

// Initialize default configurations
func (m *MyModule) InitDefaultConfigs() {
    m.Logger().Info("Initializing default configs for " + ModuleName)
    // Set default configurations here
}
```

### Approach 2: Using FX Directly

This approach uses Uber FX's native API directly, providing more control and flexibility.

```go
package example

import (
    "context"
    "go.uber.org/fx"
    "go.uber.org/zap"
)

const ModuleName = "Example"

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
```

## Module Lifecycle

All modules support the following lifecycle methods:

- `OnStart(context.Context) error`: Called when the module starts
- `OnStop(context.Context) error`: Called when the module stops
- `InitDefaultConfigs()`: Initialize default configurations

## Built-in Features

### Logging

Use the `Logger()` method to get a named Zap logger instance:

```go
m.Logger().Info("message", zap.String("key", "value"))
```

### Configuration Management

Use `GetConfigPath(key string)` to get module-scoped configuration paths:

```go
configPath := m.GetConfigPath("database_url")
// Returns: "mymodule.database_url"
```

### Module Dependency Injection

Declare dependencies in Params:

```go
type Params struct {
    weedbox.Params
    Database *database.Module
    Cache    *cache.Module
}

func (m *MyModule) OnStart(ctx context.Context) error {
    // Use injected dependencies
    m.Params().Database.Connect()
    m.Params().Cache.Set("key", "value")
    return nil
}
```

## Module Loading Order

Define module loading order in `modules.go`:

```go
func preloadModules() ([]fx.Option, error) {
    // Infrastructure modules (config, logger, etc.)
    modules := []fx.Option{
        fx.Supply(config),
        logger.Module(),
    }
    return modules, nil
}

func loadModules() ([]fx.Option, error) {
    // Business modules
    modules := []fx.Option{
        database.Module("database"),
        cache.Module("cache"),
        api.Module("api"),
    }
    return modules, nil
}

func afterModules() ([]fx.Option, error) {
    // Post-processing modules (daemon, etc.)
    modules := []fx.Option{
        daemon.Module("daemon"),
    }
    return modules, nil
}
```

## Complete Example

See the `examples/basic/` directory for a complete working example:

- `main.go`: Application entry point and CLI setup
- `modules.go`: Module loading and organization
- `pkg/example/`: Module example using native FX
- `pkg/basic/`: Module example using Weedbox base class

Run the example:

```bash
cd examples/basic
go run . --verbose
```

## Core Interfaces

### ModuleInterface

```go
type ModuleInterface interface {
    OnStart(context.Context) error
    OnStop(context.Context) error
    GetConfigPath(key string) string
    Logger() *zap.Logger
    InitDefaultConfigs()
}
```

### ModuleParams

```go
type ModuleParams interface {
    GetLifecycle() fx.Lifecycle
    GetLogger() *zap.Logger
}
```

## Dependencies

- [Uber FX](https://github.com/uber-go/fx) - Dependency injection framework
- [Uber Zap](https://github.com/uber-go/zap) - High-performance logging
- [Cobra](https://github.com/spf13/cobra) - CLI framework (used in examples)

## License

See the LICENSE file for details.
