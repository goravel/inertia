// Command setup wires goravel-inertia into a Goravel application when installed
// via `./artisan package:install github.com/goravel/inertia`.
//
// It registers the Inertia ServiceProvider in bootstrap/providers.go and installs
// the Inertia facade into app/facades/inertia.go. Frontend scaffolding (Vue 3 /
// Vite, demo pages, config) is handled separately by `./artisan inertia:install`.
package main

import (
	"os"

	"github.com/goravel/framework/packages"
	"github.com/goravel/framework/packages/modify"
	"github.com/goravel/framework/support/path"
)

func main() {
	setup := packages.Setup(os.Args)
	stubs := Stubs{}
	moduleImport := setup.Paths().Module().Import()
	// The ServiceProvider lives in the package root (package inertia), so it
	// is referenced by the module's package name — matching the convention of other
	// official packages (e.g. &gin.ServiceProvider{}).
	provider := "&inertia.ServiceProvider{}"
	inertiaFacadePath := path.Facade("inertia.go")
	facadesPackage := setup.Paths().Facades().Package()

	setup.Install(
		// Register the provider in bootstrap/providers.go.
		modify.RegisterProvider(moduleImport, provider),

		// Install the Inertia facade into app/facades/inertia.go.
		modify.File(inertiaFacadePath).Overwrite(stubs.InertiaFacade(facadesPackage)),
	).Uninstall(
		// Remove the Inertia facade.
		modify.File(inertiaFacadePath).Remove(),

		// Remove the provider from bootstrap/providers.go.
		modify.UnregisterProvider(moduleImport, provider),
	).Execute()
}
