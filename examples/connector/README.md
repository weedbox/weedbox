# connector

A runnable demo of the [`fxmodule`](../../fxmodule) connector pattern: one
interface, several swappable implementations, all loaded side-by-side.

## What it shows

- Two `Greeter` implementations (`english`, `french`) each registered via
  `fxmodule.InterfaceModule[greeter.Greeter]`
- A consumer module that injects:
  - the unnamed default (resolves to whichever Greeter was loaded first)
  - the English greeter by `name:"english"`
  - the French greeter by `name:"french"`
- Lifecycle hooks wired inside each connector's ctor running on app
  start/stop, even though no consumer explicitly references the named
  instance — `InterfaceModule`'s internal `Invoke` forces materialization

## Layout

```
connector/
├── main.go                 # cobra root, fx.New, --print_configs / --verbose flags
├── modules.go              # preloadModules / loadModules / afterModules
└── pkg/
    ├── greeter/            # interface definition
    │   └── greeter.go
    ├── english/            # implementation #1 — uses fxmodule.InterfaceModule
    │   └── module.go
    ├── french/             # implementation #2 — uses fxmodule.InterfaceModule
    │   └── module.go
    └── consumer/           # injects all three (default + 2 named)
        └── module.go
```

## Run

```bash
cd examples/connector
go run . --verbose
```

You should see, among the FX startup logs:

```
english   english greeter started   {"scope": "english"}
french    french greeter started    {"scope": "french"}
consumer  default greeting          {"msg": "Hello, world!"}
consumer  english greeting          {"msg": "Hello, Alice!"}
consumer  french greeting           {"msg": "Bonjour, Bob !"}
```

Stop with `Ctrl-C` — the `daemon` module installs the signal handler.

## Things to try

**Swap the load order in `modules.go`.** Put `french.Module("french")`
before `english.Module("english")` and re-run. The `default greeting`
line should switch to `Bonjour, world !` — load order is what claims the
unnamed default slot via `ClaimDefault[greeter.Greeter]`.

**Drop one implementation.** Remove `french.Module(...)` from
`loadModules` — the consumer will fail to wire because nothing provides
`greeter.Greeter` tagged `name:"french"`. This is how `fx.In` keeps
multi-implementation wiring honest.

**Reference the named instance only.** Take the `Default` field out of
the consumer's `Params`. The app still runs; the named injections are
independent of the unnamed-default claim.

## How this maps back to `fxmodule`

Each connector module's `Module(scope)` returns
`fxmodule.InterfaceModule[greeter.Greeter](scope, ctor)`, which:

1. provides `ctor` tagged `name:"<scope>"`
2. forces materialization via an internal `fx.Invoke`
3. on the **first** call across the process, aliases the same instance to
   the unnamed default of `greeter.Greeter`

The application composition root (`modules.go`) doesn't need to know any
of that — it just lists `english.Module("english")` and
`french.Module("french")` like any other module.

See [`fxmodule/README.md`](../../fxmodule/README.md) for the full design
notes and the `ResetClaim` test caveat.
