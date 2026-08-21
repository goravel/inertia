package main

import (
	"strings"
)

// Stubs holds the text templates written into the target application by setup.
type Stubs struct{}

// InertiaFacade returns the app/facades/inertia.go source, with the package name
// replaced by the target application's facades package name.
func (s Stubs) InertiaFacade(pkg string) string {
	content := `package DummyPackage

import (
	"github.com/goravel/inertia/contracts"
)

// Inertia returns the registered Inertia manager. The manager is bound as the
// "goravel.inertia" singleton by the Inertia ServiceProvider.
func Inertia() contracts.Inertia {
	instance, err := App().Make("goravel.inertia")
	if err != nil {
		panic(err)
	}

	return instance.(contracts.Inertia)
}
`

	return strings.ReplaceAll(content, "DummyPackage", pkg)
}
