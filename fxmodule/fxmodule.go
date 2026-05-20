// Package fxmodule provides small helpers on top of go.uber.org/fx that make
// it ergonomic to build modules supporting both an unnamed default registration
// and additional named registrations of the same type.
//
// The intended pattern: a connector-style module always registers a named
// instance keyed by its scope, and the first such module to load also exposes
// itself as the unnamed default via Alias. Subsequent modules of the same type
// only contribute their named instance, avoiding fx's duplicate-provider error.
//
// Single-process state caveat: ClaimDefault uses package-level state so the
// "first call wins" semantic spans the whole process. Tests that build more
// than one fx.App in the same process must call ResetClaim between apps, or
// the later apps will not be able to claim the default slot.
package fxmodule

import (
	"reflect"
	"sync"

	"go.uber.org/fx"
)

// Provide wraps fx.Provide. If name is the empty string the constructor is
// registered as an unnamed provider; otherwise the result is tagged with
// `name:"<name>"`.
func Provide(name string, ctor any) fx.Option {
	if name == "" {
		return fx.Provide(ctor)
	}
	return fx.Provide(fx.Annotate(ctor, fx.ResultTags(`name:"`+name+`"`)))
}

// Invoke wraps fx.Invoke for a function whose single dependency may be named.
// If name is the empty string the invoke receives the unnamed instance;
// otherwise its first parameter is tagged with `name:"<name>"`.
func Invoke(name string, fn any) fx.Option {
	if name == "" {
		return fx.Invoke(fn)
	}
	return fx.Invoke(fx.Annotate(fn, fx.ParamTags(`name:"`+name+`"`)))
}

// Alias exposes a named instance of T as the unnamed default by registering
// a pass-through constructor that depends on the named value.
func Alias[T any](name string) fx.Option {
	return fx.Provide(fx.Annotate(
		func(t T) T { return t },
		fx.ParamTags(`name:"`+name+`"`),
	))
}

var claimedDefaults sync.Map // typeKey -> struct{}

// ClaimDefault attempts to atomically claim the unnamed default slot for T.
// Returns true if the caller is the first to claim within this process,
// false if some earlier call already claimed T.
//
// Callers that win the claim are expected to follow up with Alias[T](scope)
// so the unnamed slot is actually populated. Callers that lose only register
// their named instance.
func ClaimDefault[T any]() bool {
	_, loaded := claimedDefaults.LoadOrStore(typeKey[T](), struct{}{})
	return !loaded
}

// ResetClaim clears any prior ClaimDefault for T. Intended for test setup so
// successive fx.App constructions in the same process can each claim the
// default slot independently.
func ResetClaim[T any]() {
	claimedDefaults.Delete(typeKey[T]())
}

// InterfaceModule registers ctor as a named implementation of interface I
// keyed by scope, suitable for "one interface, multiple swappable
// implementations" scenarios (database connectors, cache backends, message
// brokers, etc.).
//
// Behavior:
//   - ctor is registered with `name:"<scope>"` and must return I.
//   - The result is forced to materialize via an Invoke, so lifecycle hooks
//     wired inside ctor run even if no other consumer references the named
//     instance.
//   - The first call to InterfaceModule[I] across the process also exposes
//     itself as the unnamed default of I (via Alias), so existing single-load
//     consumers that inject I without a tag keep working with zero changes.
//
// Use this from a concrete module's Module(scope) factory. The application
// composition root continues to write `<pkg>.Module("<scope>")` exactly as
// before — InterfaceModule is connector-author scaffolding, not a
// coordinator the application needs to wire.
//
// Tests that build multiple fx.Apps in the same process must call
// ResetClaim[I] between apps; see the package-level doc for the single-
// process state caveat.
func InterfaceModule[I any](scope string, ctor any) fx.Option {
	opts := []fx.Option{
		Provide(scope, ctor),
		Invoke(scope, func(I) {}),
	}
	if ClaimDefault[I]() {
		opts = append(opts, Alias[I](scope))
	}
	return fx.Module(scope, opts...)
}

func typeKey[T any]() string {
	return reflect.TypeFor[T]().String()
}
