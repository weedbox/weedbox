package fxmodule

import (
	"testing"

	"go.uber.org/fx"
)

type greeter interface {
	Hello() string
}

type impl struct{ msg string }

func (g *impl) Hello() string { return g.msg }

func newImpl(msg string) func() greeter {
	return func() greeter { return &impl{msg: msg} }
}

func TestProvide_Unnamed(t *testing.T) {
	var got greeter
	app := fx.New(
		Provide("", newImpl("hi")),
		fx.Populate(&got),
		fx.NopLogger,
	)
	if err := app.Err(); err != nil {
		t.Fatalf("fx err: %v", err)
	}
	if got.Hello() != "hi" {
		t.Errorf("want hi, got %s", got.Hello())
	}
}

func TestProvide_Named(t *testing.T) {
	type params struct {
		fx.In
		G greeter `name:"foo"`
	}
	var got greeter
	app := fx.New(
		Provide("foo", newImpl("named")),
		fx.Invoke(func(p params) { got = p.G }),
		fx.NopLogger,
	)
	if err := app.Err(); err != nil {
		t.Fatalf("fx err: %v", err)
	}
	if got.Hello() != "named" {
		t.Errorf("want named, got %s", got.Hello())
	}
}

func TestInvoke_Named(t *testing.T) {
	var got greeter
	app := fx.New(
		Provide("bar", newImpl("bar-value")),
		Invoke("bar", func(g greeter) { got = g }),
		fx.NopLogger,
	)
	if err := app.Err(); err != nil {
		t.Fatalf("fx err: %v", err)
	}
	if got.Hello() != "bar-value" {
		t.Errorf("want bar-value, got %s", got.Hello())
	}
}

func TestInvoke_Unnamed(t *testing.T) {
	var got greeter
	app := fx.New(
		Provide("", newImpl("unnamed-invoke")),
		Invoke("", func(g greeter) { got = g }),
		fx.NopLogger,
	)
	if err := app.Err(); err != nil {
		t.Fatalf("fx err: %v", err)
	}
	if got.Hello() != "unnamed-invoke" {
		t.Errorf("want unnamed-invoke, got %s", got.Hello())
	}
}

func TestAlias_ExposesNamedAsUnnamed(t *testing.T) {
	var got greeter
	app := fx.New(
		Provide("baz", newImpl("baz-value")),
		Alias[greeter]("baz"),
		fx.Populate(&got),
		fx.NopLogger,
	)
	if err := app.Err(); err != nil {
		t.Fatalf("fx err: %v", err)
	}
	if got.Hello() != "baz-value" {
		t.Errorf("want baz-value, got %s", got.Hello())
	}
}

func TestAlias_CoexistsWithMultipleNamed(t *testing.T) {
	type params struct {
		fx.In
		Default greeter
		A       greeter `name:"a"`
		B       greeter `name:"b"`
	}
	var got params
	app := fx.New(
		Provide("a", newImpl("alpha")),
		Provide("b", newImpl("beta")),
		Alias[greeter]("a"),
		fx.Invoke(func(p params) { got = p }),
		fx.NopLogger,
	)
	if err := app.Err(); err != nil {
		t.Fatalf("fx err: %v", err)
	}
	if got.Default.Hello() != "alpha" {
		t.Errorf("default want alpha, got %s", got.Default.Hello())
	}
	if got.A.Hello() != "alpha" {
		t.Errorf("name=a want alpha, got %s", got.A.Hello())
	}
	if got.B.Hello() != "beta" {
		t.Errorf("name=b want beta, got %s", got.B.Hello())
	}
}

func TestClaimDefault_FirstCallWins(t *testing.T) {
	ResetClaim[greeter]()
	t.Cleanup(func() { ResetClaim[greeter]() })

	if !ClaimDefault[greeter]() {
		t.Fatal("first claim should succeed")
	}
	if ClaimDefault[greeter]() {
		t.Fatal("second claim should fail")
	}
	if ClaimDefault[greeter]() {
		t.Fatal("third claim should also fail")
	}
}

func TestClaimDefault_DifferentTypesIndependent(t *testing.T) {
	ResetClaim[greeter]()
	ResetClaim[*impl]()
	t.Cleanup(func() {
		ResetClaim[greeter]()
		ResetClaim[*impl]()
	})

	if !ClaimDefault[greeter]() {
		t.Fatal("greeter claim should succeed")
	}
	if !ClaimDefault[*impl]() {
		t.Fatal("*impl claim should succeed independently")
	}
}

func TestResetClaim_AllowsReclaim(t *testing.T) {
	ResetClaim[greeter]()
	t.Cleanup(func() { ResetClaim[greeter]() })

	if !ClaimDefault[greeter]() {
		t.Fatal("first claim should succeed")
	}
	ResetClaim[greeter]()
	if !ClaimDefault[greeter]() {
		t.Fatal("after reset, claim should succeed again")
	}
}

// fooModule is a connector-style module that exposes itself via the
// `greeter` interface and uses InterfaceModule for registration.
func fooModule(scope, msg string) fx.Option {
	return InterfaceModule[greeter](scope, func() greeter { return &impl{msg: msg} })
}

func TestInterfaceModule_SingleLoadHasUnnamedDefault(t *testing.T) {
	ResetClaim[greeter]()
	t.Cleanup(func() { ResetClaim[greeter]() })

	type params struct {
		fx.In
		Default greeter
		Named   greeter `name:"only"`
	}
	var got params
	app := fx.New(
		fooModule("only", "single"),
		fx.Invoke(func(p params) { got = p }),
		fx.NopLogger,
	)
	if err := app.Err(); err != nil {
		t.Fatalf("fx err: %v", err)
	}
	if got.Default == nil || got.Named == nil {
		t.Fatal("both default and named must resolve")
	}
	if got.Default != got.Named {
		t.Error("single-load default should be the same instance as the named one")
	}
}

func TestInterfaceModule_MultiLoadNamedDistinctFirstIsDefault(t *testing.T) {
	ResetClaim[greeter]()
	t.Cleanup(func() { ResetClaim[greeter]() })

	type params struct {
		fx.In
		Default greeter
		A       greeter `name:"a"`
		B       greeter `name:"b"`
	}
	var got params
	app := fx.New(
		fooModule("a", "alpha"),
		fooModule("b", "beta"),
		fx.Invoke(func(p params) { got = p }),
		fx.NopLogger,
	)
	if err := app.Err(); err != nil {
		t.Fatalf("fx err: %v", err)
	}
	if got.A.Hello() != "alpha" {
		t.Errorf("name=a want alpha, got %s", got.A.Hello())
	}
	if got.B.Hello() != "beta" {
		t.Errorf("name=b want beta, got %s", got.B.Hello())
	}
	if got.Default != got.A {
		t.Error("unnamed default should be the first-loaded module (a)")
	}
	if got.A == got.B {
		t.Error("named instances should be distinct objects")
	}
}

func TestInterfaceModule_LoadOrderControlsDefault(t *testing.T) {
	ResetClaim[greeter]()
	t.Cleanup(func() { ResetClaim[greeter]() })

	type params struct {
		fx.In
		Default greeter
		B       greeter `name:"b"`
	}
	var got params
	app := fx.New(
		fooModule("b", "beta"), // loaded first → wins the default slot
		fooModule("a", "alpha"),
		fx.Invoke(func(p params) { got = p }),
		fx.NopLogger,
	)
	if err := app.Err(); err != nil {
		t.Fatalf("fx err: %v", err)
	}
	if got.Default != got.B {
		t.Errorf("default should be the first-loaded (b/beta), got %s", got.Default.Hello())
	}
}

func TestInterfaceModule_MaterializesEvenWithoutConsumer(t *testing.T) {
	ResetClaim[greeter]()
	t.Cleanup(func() { ResetClaim[greeter]() })

	// If the connector relies on InterfaceModule's internal Invoke to force
	// instantiation, the ctor must run even when no fx.Invoke / fx.Populate
	// references the interface.
	var ran bool
	app := fx.New(
		InterfaceModule[greeter]("noref", func() greeter {
			ran = true
			return &impl{msg: "noref"}
		}),
		fx.NopLogger,
	)
	if err := app.Err(); err != nil {
		t.Fatalf("fx err: %v", err)
	}
	if !ran {
		t.Error("ctor must run because InterfaceModule's Invoke forces materialization")
	}
}
