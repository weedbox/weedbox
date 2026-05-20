// Package greeter defines a tiny interface with multiple swappable
// implementations. It exists to illustrate the fxmodule connector pattern:
// one interface, several modules, each addressable by scope.
package greeter

// Greeter returns a localized greeting for the given name.
type Greeter interface {
	Greet(name string) string
}
