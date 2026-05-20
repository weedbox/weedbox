# basic

A minimal Weedbox application that ties together the two module-authoring
styles supported by the framework.

## What it shows

- The canonical app skeleton: cobra CLI, `configs`, `logger`, `daemon`
- A module written directly against `go.uber.org/fx` (`pkg/example`)
- A module written using the `weedbox.Module[*Params]` generic base class
  (`pkg/basic`) that depends on the FX-native module via `Params`

Use this as the starting point for a new service. Both styles can coexist
in the same app — pick whichever fits the module at hand.

## Layout

```
basic/
├── main.go                 # cobra root, fx.New, --print_configs / --verbose flags
├── modules.go              # preloadModules / loadModules / afterModules
└── pkg/
    ├── example/            # FX-native style
    │   └── module.go
    └── basic/              # weedbox.Module[*Params] style
        └── module.go
```

## Run

```bash
cd examples/basic
go run . --verbose
```

Flags:

- `--verbose` — enable the FX startup log (otherwise `fx.NopLogger` is used)
- `--print_configs` — print every config key registered by loaded modules

Stop with `Ctrl-C` — the `daemon` module installs the signal handler that
unwinds FX lifecycle hooks in reverse.

## Module-authoring styles

### FX-native (`pkg/example`)

Plain `fx.Module` + `fx.Provide` + `fx.Invoke`, with lifecycle hooks
appended to `fx.Lifecycle`. Use this when you want maximum control or when
the module doesn't need anything the base class offers.

### Weedbox base class (`pkg/basic`)

Embeds `weedbox.Module[*Params]` and lets `weedbox.InitModule` wire the
lifecycle, scoped logger, and config-path helper. The module file becomes
mostly business logic — `OnStart`, `OnStop`, `InitDefaultConfigs` — instead
of FX plumbing.

`pkg/basic` also demonstrates module-to-module dependency: its `Params`
declares `Example *example.Example`, so FX injects the instance built by
`pkg/example` and the base class exposes it through `m.Params().Example`.

## Loading order

`modules.go` splits modules into three phases so dependencies resolve in
the order you'd expect:

1. `preloadModules` — infrastructure (`configs`, `logger`)
2. `loadModules` — business modules (`example`, `basic`)
3. `afterModules` — post-processing (`daemon`)

Order within each phase matters when a later module's `Params` injects an
earlier one.
