# fxmodule

Small helpers on top of [`go.uber.org/fx`](https://github.com/uber-go/fx) for
the "one interface, several swappable implementations" pattern — database
connectors, cache backends, message brokers, and other plug points where
several modules want to register the same interface side-by-side.

## Why this exists

Vanilla `fx.Provide` rejects two providers for the same type. The usual
workaround is to tag every provider with `fx.ResultTags(name:"...")` and tag
every consumer to match. That works, but it breaks the common backwards-
compatibility case: callers that already inject the interface *without* a
name tag suddenly fail to wire when a second implementation shows up.

`fxmodule` papers over that with a tiny convention: every implementation
registers itself by scope (named), and the **first** implementation loaded
in the process also claims the unnamed default slot. Single-load consumers
keep working; multi-load consumers address each implementation by name.

## `InterfaceModule[I](scope, ctor)`

The headline helper. Use it from a concrete module's `Module(scope)`
factory:

```go
package sqlite_connector

import (
    "github.com/weedbox/common-modules/database"
    "github.com/weedbox/weedbox/fxmodule"
    "go.uber.org/fx"
)

func Module(scope string) fx.Option {
    return fxmodule.InterfaceModule[database.DatabaseConnector](
        scope,
        func(p Params) database.DatabaseConnector {
            c := &SQLiteConnector{ /* ... */ }
            p.Lifecycle.Append(fx.Hook{
                OnStart: c.onStart,
                OnStop:  c.onStop,
            })
            return c
        },
    )
}
```

What `InterfaceModule[I]` does, in order:

1. Registers `ctor` with `name:"<scope>"` returning `I`.
2. Adds an internal `fx.Invoke` that depends on the named result. This
   forces materialization, so lifecycle hooks wired inside `ctor` run even
   if no other consumer references the instance.
3. The **first** call across the process also aliases the same instance to
   the unnamed default of `I` (via `Alias`). Subsequent calls only
   contribute their named instance — no duplicate-provider conflict.

## Consuming implementations

```go
fx.New(
    sqlite_connector.Module("cache"),
    postgres_connector.Module("main"),
    fx.Invoke(func(p struct {
        fx.In
        Default database.DatabaseConnector                       // == cache (loaded first)
        Cache   database.DatabaseConnector `name:"cache"`
        Main    database.DatabaseConnector `name:"main"`
    }) {
        // use p.Default / p.Cache / p.Main
    }),
)
```

If load order is brittle in your app, always inject by named tag and ignore
the unnamed default. The default exists for the single-load /
backwards-compat case, not as a routing mechanism.

## Test caveat: `ResetClaim`

The "first call wins" claim on the unnamed default uses process-level
state. Tests that build more than one `fx.App` in the same process must
reset that claim between apps, otherwise the second app cannot register an
unnamed default and any consumer asking for `I` without a name tag will
fail to wire:

```go
import "github.com/weedbox/weedbox/fxmodule"

func TestSomething(t *testing.T) {
    fxmodule.ResetClaim[database.DatabaseConnector]()
    t.Cleanup(func() {
        fxmodule.ResetClaim[database.DatabaseConnector]()
    })
    // ... build fx.App ...
}
```

## Lower-level primitives

`InterfaceModule` is composed from smaller helpers, exported for custom
wiring:

| Helper              | What it does                                                                                                        |
|---------------------|---------------------------------------------------------------------------------------------------------------------|
| `Provide(name, c)`  | Like `fx.Provide`, but tags the result `name:"<name>"` when `name` is non-empty. Empty `name` falls back to plain provide. |
| `Invoke(name, fn)`  | Like `fx.Invoke`, but tags the function's first parameter `name:"<name>"` when `name` is non-empty.                 |
| `Alias[T](name)`    | Re-exports a `name:"<name>"`-tagged `T` as the unnamed default.                                                     |
| `ClaimDefault[T]()` | Atomically claims the unnamed default slot for `T`. Returns `true` on the first call, `false` afterwards.           |
| `ResetClaim[T]()`   | Clears any prior `ClaimDefault[T]`. Test-only.                                                                      |

Reach for these directly when you need behavior `InterfaceModule` doesn't
cover — e.g. claiming the default conditionally, or registering a
concrete type rather than an interface.

## Runnable example

See [`examples/connector`](../examples/connector) for a self-contained app
that uses `InterfaceModule` to load two implementations of a `Greeter`
interface side-by-side and inject them three ways (named, named, default).
